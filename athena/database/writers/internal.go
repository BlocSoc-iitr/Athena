package writers

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"time"
	"gorm.io/gorm"
	"github.com/spf13/afero"
	"github.com/sirupsen/logrus"
	"github.com/BlocSoc-iitr/Athena/athena/database/models"
	"github.com/BlocSoc-iitr/Athena/athena/types"
)

type Config struct {
	AppName string
}

var config = Config{
	AppName: "athena",
}

var db *gorm.DB
var logger = logrus.New()

func writeABI(abi *models.ContractABI, db *gorm.DB) {
	if db != nil {
		if err := db.Save(abi).Error; err != nil {
			log.Fatalf("Error in writing abi: %v", err)
		}
	} else {
		appFS := afero.NewOsFs()
		appDir, err := exec.LookPath(config.AppName)
		if err != nil {
			log.Fatalf("Error in writing abi: %v", err)
		}
		appDir = filepath.Dir(appDir)

		contractPath := filepath.Join(appDir, "contract-abis.json")

		var abiJson []map[string]interface{}

		if exists, _ := afero.Exists(appFS, contractPath); exists {
			file, err := appFS.OpenFile(contractPath, os.O_RDONLY, 0644)
			if err != nil {
				log.Fatalf("Error in writing abi: %v", err)
			}
			defer file.Close()

			if stat, _ := file.Stat(); stat.Size() != 0 {
				if err := json.NewDecoder(file).Decode(&abiJson); err != nil {
					log.Fatalf("Error in writing abi: %v", err)
				}
			}
		}

		abiMap := StructToMap(abi)

		abiJson = append(abiJson, abiMap)

		file, err := appFS.OpenFile(contractPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatalf("Error in writing abi: %v", err)
		}
		defer file.Close()

		if err := json.NewEncoder(file).Encode(abiJson); err != nil {
			log.Fatalf("Error in writing abi: %v", err)
		}
	}
}

// StructToMap converts a struct to a map[string]interface{} using reflection
func StructToMap(obj interface{}) (map[string]interface{}) {
	result := make(map[string]interface{})
	value := reflect.ValueOf(obj)
	typ := value.Type()

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldName := typ.Field(i).Name

		if field.CanInterface() {
			result[fieldName] = field.Interface()
		}
	}

	return result
}

func deleteABI(abiName string, db *gorm.DB, decoderOS types.SupportedNetwork) {
	if db != nil {
		logger.Info("Running DB Query to Delete ABI")
		result := db.Where("abi_name = ? AND decoder_os = ?", abiName, decoderOS).Delete(&models.ContractABI{})
		if result.Error != nil {
			logger.Errorf("Error deleting ABI from database: %v", result.Error)
			return
		}
		logger.Infof("Deleted %d records", result.RowsAffected)
	} else {
		appDir, err := os.UserConfigDir()
		if err != nil {
			logger.Errorf("Error getting config directory: %v", err)
			return
		}
		appDir = filepath.Join(appDir, "athena")

		if err := os.MkdirAll(appDir, 0755); err != nil {
			logger.Errorf("Error creating application directory: %v", err)
			return
		}

		contractPath := filepath.Join(appDir, "contract-abis.json")
		if _, err := os.Stat(contractPath); os.IsNotExist(err) {
			logger.Info("ABI File does not exist... No ABIs to delete")
			return
		}

		file, err := os.ReadFile(contractPath)
		if err != nil {
			logger.Errorf("Error reading ABI file: %v", err)
			return
		}

		if len(file) == 0 {
			logger.Info("ABI File is empty... No ABIs to delete")
			return
		}

		var contractABIs []models.ContractABI
		if err := json.Unmarshal(file, &contractABIs); err != nil {
			logger.Errorf("Error unmarshaling ABI JSON: %v", err)
			return
		}

		found := false
		updatedABIs := make([]models.ContractABI, 0, len(contractABIs))
		for _, abi := range contractABIs {
			if abi.AbiName != abiName || abi.DecoderOS != decoderOS.String() {
				updatedABIs = append(updatedABIs, abi)
			} else {
				found = true
			}
		}

		if found {
			logger.Info("ABI Found... Deleting from File Cache")
			updatedJSON, err := json.Marshal(updatedABIs)
			if err != nil {
				logger.Errorf("Error marshaling updated ABIs: %v", err)
				return
			}

			if err := os.WriteFile(contractPath, updatedJSON, 0644); err != nil {
				logger.Errorf("Error writing updated ABI file: %v", err)
				return
			}

			logger.Info("ABI Cache Updated")
		} else {
			logger.Info("ABI Not Found in Cache... No Deletion Necessary")
		}
	}
}


func writeBlockTimestamps(timestamps []types.BlockTimestamp, network types.SupportedNetwork) {
	appDir, err := os.UserConfigDir()
	if err != nil {
		logger.Errorf("Error getting config directory: %v", err)
		return
	}
	appDir = filepath.Join(appDir, "athena")

	if err := os.MkdirAll(appDir, 0755); err != nil {
		logger.Errorf("Error creating application directory: %v", err)
		return
	}

	filePath := filepath.Join(appDir, network.String()+"-timestamps.json")

	var existingTimestamps []types.BlockTimestamp

	if _, err := os.Stat(filePath); err == nil {
		file, err := os.ReadFile(filePath)
		if err != nil {
			logger.Errorf("Error reading timestamp file: %v", err)
			return
		}

		if len(file) > 0 {
			var timestampJSON []map[string]interface{}
			if err := json.Unmarshal(file, &timestampJSON); err != nil {
				logger.Errorf("Error unmarshaling timestamp JSON: %v", err)
				return
			}

			for _, t := range timestampJSON {
				blockNumber := int64(t["block_number"].(float64))
				timestamp, err := time.Parse(time.RFC3339, t["timestamp"].(string))
				if err != nil {
					logger.Errorf("Error parsing timestamp: %v", err)
					continue
				}
				existingTimestamps = append(existingTimestamps, types.BlockTimestamp{
					BlockNumber: int(blockNumber),
					Timestamp:   timestamp,
				})
			}
		}
	}

	existingBlocks := make(map[int64]bool)
	for _, t := range existingTimestamps {
		existingBlocks[int64(t.BlockNumber)] = true
	}

	for _, t := range timestamps {
		if !existingBlocks[int64(t.BlockNumber)] {
			existingTimestamps = append(existingTimestamps, t)
		}
	}

	dataclassDicts := make([]map[string]interface{}, len(existingTimestamps))
	for i, t := range existingTimestamps {
		dataclassDicts[i] = map[string]interface{}{
			"block_number": t.BlockNumber,
			"timestamp":    t.Timestamp.Format(time.RFC3339),
		}
	}

	jsonData, err := json.Marshal(dataclassDicts)
	if err != nil {
		logger.Errorf("Error marshaling timestamp data: %v", err)
		return
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		logger.Errorf("Error writing timestamp file: %v", err)
		return
	}

	logger.Info("Block timestamps written successfully")
}
