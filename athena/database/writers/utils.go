package writers

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"gorm.io/gorm"
	_ "github.com/go-sql-driver/mysql"
)

func ModelToDict(model interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	modelValue := reflect.ValueOf(model)
	modelType := modelValue.Type()

	for i := 0; i < modelValue.NumField(); i++ {
		field := modelValue.Field(i)
		fieldName := modelType.Field(i).Name

		if field.CanInterface() {
			result[fieldName] = field.Interface()
		}
	}

	return result
}

// DBEncodeHex encodes data to a hex string or bytes depending on the database dialect
func DBEncodeHex(data interface{}, dbDialect string) (interface{}) {
	switch dbDialect {
	case "mysql":
		switch v := data.(type) {
		case string:
			if strings.HasPrefix(v, "0x") {
				return v
			}
			return "0x" + v
		case []byte:
			return "0x" + hex.EncodeToString(v)
		default:
			logger.Errorf("invalid data type: %T", data)
		}
	default:
		logger.Errorf("unsupported database dialect: %s", dbDialect)
		return nil
	}
	return nil
}

// TraceAddressToString converts a trace address to a string representation
func TraceAddressToString(traceAddress []int) string {
	strInts := make([]string, len(traceAddress))
	for i, v := range traceAddress {
		strInts[i] = strconv.Itoa(v)
	}
	return "[" + strings.Join(strInts, ",") + "]"
}

// StringToTraceAddress converts a trace address string to a slice of integers
func StringToTraceAddress(traceAddressString string) ([]int) {
	trimmed := strings.Trim(traceAddressString, "[]")
	strInts := strings.Split(trimmed, ",")
	result := make([]int, len(strInts))
	for i, s := range strInts {
		v, err := strconv.Atoi(s)
		if err != nil {
			logger.Errorf("Error in String to Trace Address: %v", err)
		}
		result[i] = v
	}
	return result
}

// TableInfo represents information about a database table
type TableInfo struct {
	Name    string
	Columns []ColumnInfo
}

// ColumnInfo represents information about a database column
type ColumnInfo struct {
	Name string
	Type string
	DefaultValue string
}

// GetTableInfo retrieves table information from the MySQL database
func AutomapSqlalchemyModel(db *gorm.DB, tableNames []string, schema string) map[string]TableInfo {
	logger.Infof("Getting table info for %v from schema %s", tableNames, schema)

	result := make(map[string]TableInfo)

	for _, tableName := range tableNames {
		query := fmt.Sprintf("SHOW COLUMNS FROM %s.%s", schema, tableName)
		rows, err := db.Raw(query).Rows()
		if err != nil {
			logger.Errorf("could not get columns for table %s: %v", tableName, err)
			continue
		}
		defer rows.Close()

		var columns []ColumnInfo
		for rows.Next() {
			var column ColumnInfo
			var null, key, extra string
			var defaultValue sql.NullString
			err := rows.Scan(&column.Name, &column.Type, &null, &key, &defaultValue, &extra)
			if err != nil {
				logger.Errorf("error scanning column info: %v", err)
				continue
			}
			column.DefaultValue = defaultValue.String
			columns = append(columns, column)
		}

		result[tableName] = TableInfo{
			Name:    tableName,
			Columns: columns,
		}
	}

	return result
}

// func main() {
// 	// Example usage
// 	db, err := sql.Open("mysql", "user:password@/dbname?charset=utf8&parseTime=True&loc=Local")
// 	if err != nil {
// 		logger.Fatalf("Failed to connect to database: %v", err)
// 	}
// 	defer db.Close()

// 	// Example struct
// 	type User struct {
// 		ID   uint
// 		Name string
// 	}

// 	user := User{ID: 1, Name: "John Doe"}
// 	userDict := ModelToDict(user)
// 	fmt.Printf("User as dict: %v\n", userDict)

// 	hexData, err := DBEncodeHex("1234", "mysql")
// 	if err != nil {
// 		logger.Fatalf("Error encoding hex: %v", err)
// 	}
// 	fmt.Printf("Encoded hex: %v\n", hexData)

// 	traceAddr := []int{0, 1, 2}
// 	traceStr := TraceAddressToString(traceAddr)
// 	fmt.Printf("Trace address as string: %s\n", traceStr)

// 	backToAddr, err := StringToTraceAddress(traceStr)
// 	if err != nil {
// 		logger.Fatalf("Error converting trace address string: %v", err)
// 	}
// 	fmt.Printf("Back to trace address: %v\n", backToAddr)

// 	tableInfo, err := GetTableInfo(db, []string{"users", "posts"}, "mydatabase")
// 	if err != nil {
// 		logger.Fatalf("Error getting table info: %v", err)
// 	}
// 	for tableName, info := range tableInfo {
// 		fmt.Printf("Table: %s\n", tableName)
// 		for _, column := range info.Columns {
// 			fmt.Printf("  Column: %s, Type: %s\n", column.Name, column.Type)
// 		}
// 	}
// }