package backfill

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"strings"
	"time"
)

// ExportMode defines the mode of export: db_models or csv.
type ExportMode string

const (
	DBModels ExportMode = "db_models"
	CSV      ExportMode = "csv"
)

// BackfillError is a custom error type for backfill-related errors.
var BackfillError = errors.New("backfill error")

// BackfillDataType represents different types of blockchain data that can be backfilled.
type BackfillDataType string

const (
	FullBlocks   BackfillDataType = "full_blocks"
	Blocks       BackfillDataType = "blocks"
	Transactions BackfillDataType = "transactions"
	Transfers    BackfillDataType = "transfers"
	Events       BackfillDataType = "events"
	Traces       BackfillDataType = "traces"
)

//here the value stored inside the enum becomes the string

// AbstractResourceExporter is the base class for resource exporters.
type AbstractResourceExporter struct {
	exportMode     ExportMode
	initTime       time.Time
	resourcesSaved int
}

// use pointer as allow the actual value modification
func (e *AbstractResourceExporter) Init() {
	e.initTime = time.Now()
	e.resourcesSaved = 0
}

func (e *AbstractResourceExporter) Close() {
	duration := time.Since(e.initTime)

	// Convert duration to minutes and seconds
	minutes := int(duration / time.Minute)
	seconds := int(duration % time.Minute / time.Second)

	fmt.Printf("Exported %d rows in %d minutes and %d seconds\n", e.resourcesSaved, minutes, seconds)
}

func (e *AbstractResourceExporter) EncodeDataclassAsDict(dataclass map[string]interface{}) map[string]interface{} {
	return dataclass
}

// FileResourceExporter is used to export resources to a CSV file.
type FileResourceExporter struct {
	AbstractResourceExporter
	fileName     string
	writeHeaders bool
	csvSeparator string
	fileHandle   *os.File
	writer       *csv.Writer
}

func NewFileResourceExporter(fileName string, append bool) (*FileResourceExporter, error) { //constructor function or init funct
	if !strings.HasSuffix(fileName, ".csv") {
		return nil, errors.New("export file name must be a .csv file")
	}

	// Determine file mode
	// var mode string
	// if append {
	// 	mode = "a"
	// } else {
	// 	mode = "w"
	// }

	// Open file
	fileHandle, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	// Initialize writer
	writer := csv.NewWriter(fileHandle)

	exporter := &FileResourceExporter{
		fileName:     fileName,
		writeHeaders: false,
		csvSeparator: "|", //unless specified explicitly this will remain the default value same as in python where there is no way to explicitly define this
		fileHandle:   fileHandle,
		writer:       writer,
	}
	exporter.Init()

	// Check if file is empty to determine if headers should be written
	fileInfo, err := fileHandle.Stat() //has variuos methods defined on it like name , size , mode permissions and modification time etc
	if err != nil {
		return nil, err
	}
	if fileInfo.Size() == 0 {
		exporter.writeHeaders = true
	}

	return exporter, nil
}

func (e *FileResourceExporter) CSVEncodeValue(val interface{}) (string, error) { //[]interface{} holds value of any type just like a list in python -->  this is only in a slice mde by using []
	switch v := val.(type) {
	case nil:
		return "", nil
	case string:
		return v, nil
	case int, float64:
		return fmt.Sprint(v), nil
	case []interface{}:
		encodedList := make([]string, len(v))
		for i, item := range v {
			encodedItem, err := e.CSVEncodeValue(item)
			if err != nil {
				return "", err
			}
			encodedList[i] = encodedItem
		}
		return "[" + strings.Join(encodedList, ",") + "]", nil
	case []byte:
		return fmt.Sprintf("%x", v), nil
	case map[string]interface{}: //doesnt matter if key is string as not used in implementation
		//interface{} can store any value in it
		encodedMap, err := json.Marshal(v) //This function call converts the map v into a JSON-encoded byte slice. The json.Marshal function serializes Go values into JSON format
		if err != nil {
			return "", err
		}
		return string(encodedMap), nil
	default:
		return "", fmt.Errorf("cannot encode %v to CSV", v)
	}
}

func (e *FileResourceExporter) EncodeDataclass(dataclass map[string]interface{}) ([]string, string, error) {
	encodedDict := e.EncodeDataclassAsDict(dataclass)
	csvEncoded := make([]string, len(encodedDict))
	headers := make([]string, 0, len(encodedDict))

	i := 0
	for key, val := range encodedDict {
		headers = append(headers, key)
		encodedVal, err := e.CSVEncodeValue(val)
		if err != nil {
			return nil, "", err
		}
		csvEncoded[i] = encodedVal
		i++
	}

	if e.exportMode == CSV {
		return headers, strings.Join(csvEncoded, e.csvSeparator), nil
	}

	return nil, "", fmt.Errorf("export mode %s is not implemented", e.exportMode)
}

func (e *FileResourceExporter) Write(resources []map[string]interface{}) error {
	for _, resource := range resources {
		headers, csvRow, err := e.EncodeDataclass(resource)
		if err != nil {
			return err
		}
		if e.writeHeaders {
			e.writer.Write(headers)
			e.writeHeaders = false
		}
		e.writer.Write(strings.Split(csvRow, e.csvSeparator))
	}
	e.resourcesSaved += len(resources)
	e.writer.Flush()
	return nil
}

// Backfill logic
func GetFileExportersForBackfill(backfillType BackfillDataType, kwargs map[string]interface{}) (map[string]*FileResourceExporter, error) {
	switch backfillType {
	case FullBlocks:
		if blockFile, ok1 := kwargs["block_file"]; ok1 {
			if txFile, ok2 := kwargs["transaction_file"]; ok2 {
				if eventFile, ok3 := kwargs["event_file"]; ok3 {
					return map[string]*FileResourceExporter{
						"blocks":       blockFile.(*FileResourceExporter),
						"transactions": txFile.(*FileResourceExporter),
						"events":       eventFile.(*FileResourceExporter),
					}, nil
				}
			}
		}
		return nil, BackfillError
	case Blocks:
		if blockFile, ok := kwargs["block_file"]; ok {
			return map[string]*FileResourceExporter{
				"blocks": blockFile.(*FileResourceExporter),
			}, nil
		}
		return nil, BackfillError
	case Events:
		if eventFile, ok := kwargs["event_file"]; ok {
			return map[string]*FileResourceExporter{
				"events": eventFile.(*FileResourceExporter),
			}, nil
		}
		return nil, BackfillError
	case Transactions:
		if txFile, ok1 := kwargs["transaction_file"]; ok1 {
			if blockFile, ok2 := kwargs["block_file"]; ok2 {
				return map[string]*FileResourceExporter{
					"blocks":       blockFile.(*FileResourceExporter),
					"transactions": txFile.(*FileResourceExporter),
				}, nil
			}
		}
		return nil, BackfillError
	case Transfers:
		if transferFile, ok := kwargs["transfer_file"]; ok {
			return map[string]*FileResourceExporter{
				"transfers": transferFile.(*FileResourceExporter),
			}, nil
		}
		return nil, BackfillError
	case Traces:
		if traceFile, ok := kwargs["trace_file"]; ok {
			return map[string]*FileResourceExporter{
				"traces": traceFile.(*FileResourceExporter),
			}, nil
		}
		return nil, BackfillError
	default:
		return nil, fmt.Errorf("backfill type %s cannot be exported to CSV", backfillType)
	}
}

func main() {
	// Example usage
	exporter, err := NewFileResourceExporter("output.csv", true)
	if err != nil {
		fmt.Println(err)
		return
	}

	data := []map[string]interface{}{
		{"block_number": 123, "tx_count": 10},
		{"block_number": 124, "tx_count": 12},
	}

	err = exporter.Write(data)
	if err != nil {
		fmt.Println(err)
	}
	exporter.Close()
}
