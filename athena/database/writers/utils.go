package writers

import (
	"encoding/hex"
	// "errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	// "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Logger setup
var rootLogger = log.New(log.Writer(), "nethermind: ", log.LstdFlags)
var logger = log.New(rootLogger.Writer(), "entro.db.utils: ", log.LstdFlags)

// DatabaseError for handling database-specific errors
type DatabaseError struct {
	Message string
	Err     error
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

// ModelToDict converts a GORM model to a map (similar to Python dict)
func ModelToDict(model interface{}) (map[string]interface{}, error) {
	modelMap := make(map[string]interface{})
	// Use reflection to iterate over fields of the struct
	err := schema.Load(modelMap, model)
	if err != nil {
		return nil, err
	}
	return modelMap, nil
}

// DBEncodeHex encodes data to a hex string or bytes depending on the database dialect
func DBEncodeHex(data interface{}, dbDialect string) (interface{}, error) {
	switch dbDialect {
	case "postgresql":
		switch v := data.(type) {
		case string:
			return hex.DecodeString(v[2:])
		case []byte:
			return v, nil
		default:
			return nil, fmt.Errorf("Invalid data type: %T", v)
		}
	default:
		switch v := data.(type) {
		case string:
			if strings.HasPrefix(v, "0x") {
				return v, nil
			}
			return "0x" + v, nil
		case []byte:
			return "0x" + hex.EncodeToString(v), nil
		default:
			return nil, fmt.Errorf("Invalid data type: %T", v)
		}
	}
}

// TraceAddressToString converts a trace address to a string representation that can be used in a primary key
func TraceAddressToString(traceAddress []int) string {
	stringParts := make([]string, len(traceAddress))
	for i, val := range traceAddress {
		stringParts[i] = strconv.Itoa(val)
	}
	return "[" + strings.Join(stringParts, ",") + "]"
}

// StringToTraceAddress converts a trace address string to a list of integers
func StringToTraceAddress(traceAddressString string) ([]int, error) {
	trimmed := traceAddressString[1 : len(traceAddressString)-1]
	stringParts := strings.Split(trimmed, ",")
	traceAddress := make([]int, len(stringParts))
	for i, strVal := range stringParts {
		intVal, err := strconv.Atoi(strVal)
		if err != nil {
			return nil, err
		}
		traceAddress[i] = intVal
	}
	return traceAddress, nil
}

// AutomapGORMModel automaps a list of tables from a database schema and returns a map of table names to GORM models
func AutomapGORMModel(db *gorm.DB, tableNames []string, schemaName string) (map[string]interface{}, error) {
	logger.Printf("Automapping tables %v from schema %s\n", tableNames, schemaName)

	var tables []string
	err := db.Table("information_schema.tables").Select("table_name").
		Where("table_schema = ?", schemaName).
		Scan(&tables).Error
	if err != nil {
		return nil, &DatabaseError{"Could not load tables", err}
	}

	tableMap := make(map[string]interface{})

	for _, tableName := range tableNames {
		if !contains(tables, tableName) {
			return nil, &DatabaseError{fmt.Sprintf("Table %s not found in database", tableName), nil}
		}

		tableMap[tableName] = db.Table(tableName)
	}

	return tableMap, nil
}

// Utility function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}
