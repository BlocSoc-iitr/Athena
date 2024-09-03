package writers

import (
	"fmt"
	"testing"

	// "database/sql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	// "log"
	// "gorm.io/driver/mysql"
	// "gorm.io/gorm"
	"github.com/DarkLord017/athena/athena/database"
	"github.com/DarkLord017/athena/athena/database/models"
	"github.com/DarkLord017/athena/athena/database/readers"
)

//**************************/
//****** Utils.go **********/
//**************************/
func TestModelToDict(t *testing.T) {
	dsn := "root:MySQLDatabase$24@tcp(127.0.0.1:3306)/athena?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}

	database.MigrateUp(db)

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

	tableInfo := AutomapSqlalchemyModel(db, []string{"transactions","blocks","contract_abis","blocks"}, "athena")
	for tableName, info := range tableInfo {
		fmt.Printf("Table: %s\n", tableName)
		for _, column := range info.Columns {
			fmt.Printf("  Column: %s, Type: %s\n", column.Name, column.Type)
		}
	}
}

func TestWriteAbi(t *testing.T) {
	dsn := "root:MySQLDatabase$24@tcp(127.0.0.1:3306)/athena?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}

	database.MigrateUp(db)

	contractAbi1 := models.ContractABI{
		AbiName: "transaction1",
		AbiJson: []map[string]interface{}{},
		Priority: 100,
		DecoderOS: "cairo",
	}

	contractAbi2 := models.ContractABI{
		AbiName: "transaction2",
		AbiJson: []map[string]interface{}{},
		Priority: 120,
		DecoderOS: "cairo",
	}

	writeABI(&contractAbi1, db)
	writeABI(&contractAbi2, db)

	fetchedAbi := readers.GetAbis(db, []string{"transaction1", "transaction2"}, "cairo")

	fmt.Printf("Decoded ABI: %v", fetchedAbi)
}