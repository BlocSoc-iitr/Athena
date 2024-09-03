package writers 

import (
	"testing"
	"fmt"
	"database/sql"
	// "log"
	// "gorm.io/driver/mysql"
	// "gorm.io/gorm"
	// "github.com/DarkLord017/athena/athena/database/models"

)

//**************************/
//****** Utils.go **********/
//**************************/
func TestModelToDict(t *testing.T) {
	db, err := sql.Open("mysql", "user:password@/dbname?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Example struct
	type User struct {
		ID   uint
		Name string
	}

	user := User{ID: 1, Name: "John Doe"}
	userDict := ModelToDict(user)
	fmt.Printf("User as dict: %v\n", userDict)

	hexData := DBEncodeHex("1234", "mysql")
	fmt.Printf("Encoded hex: %v\n", hexData)

	traceAddr := []int{0, 1, 2}
	traceStr := TraceAddressToString(traceAddr)
	fmt.Printf("Trace address as string: %s\n", traceStr)

	backToAddr := StringToTraceAddress(traceStr)
	fmt.Printf("Back to trace address: %v\n", backToAddr)

	tableInfo := AutomapSqlalchemyModel(db, []string{"users", "posts"}, "mydatabase")
	for tableName, info := range tableInfo {
		fmt.Printf("Table: %s\n", tableName)
		for _, column := range info.Columns {
			fmt.Printf("  Column: %s, Type: %s\n", column.Name, column.Type)
		}
	}
}