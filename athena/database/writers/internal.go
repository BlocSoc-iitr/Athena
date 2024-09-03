package writers

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/jinzhu/gorm"
	"github.com/spf13/afero"
	"github.com/DarkLord017/athena/athena/database/models"
)

type Config struct {
	AppName string
}

var config = Config{
	AppName: "entro",
}

var db *gorm.DB

func writeABI(abi *models.ContractABI, db *gorm.DB) error {
	if db != nil {
		if err := db.Save(abi).Error; err != nil {
			return err
		}
	} else {
		appFS := afero.NewOsFs()
		appDir, err := afero.GetAppDir(appFS, config.AppName)
		if err != nil {
			return err
		}

		contractPath := filepath.Join(appDir, "contract-abis.json")

		var abiJson []map[string]interface{}

		if exists, _ := afero.Exists(appFS, contractPath); exists {
			file, err := appFS.OpenFile(contractPath, os.O_RDONLY, 0644)
			if err != nil {
				return err
			}
			defer file.Close()

			if stat, _ := file.Stat(); stat.Size() != 0 {
				if err := json.NewDecoder(file).Decode(&abiJson); err != nil {
					return err
				}
			}
		}

		abiMap, err := StructToMap(abi)
		if err != nil {
			return err
		}

		abiJson = append(abiJson, abiMap)

		file, err := appFS.OpenFile(contractPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer file.Close()

		if err := json.NewEncoder(file).Encode(abiJson); err != nil {
			return err
		}
	}
	return nil
}

// StructToMap converts a struct to a map[string]interface{} using reflection
func StructToMap(data interface{}) (map[string]interface{}, error) {
	// Implementation to convert struct to map
	// You can use libraries like "github.com/fatih/structs"
	return nil, nil
}

func main() {
	// Database initialization
	var err error
	db, err = gorm.Open("mysql", "user:password@tcp(localhost:3306)/dbname?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Example usage
	abi := &ContractABI{
		AbiName:   "example",
		DecoderOS: "EVM",
	}
	if err := writeABI(abi, db); err != nil {
		log.Fatal(err)
	}
}
