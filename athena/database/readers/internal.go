package readers

import (
	"github.com/BlocSoc-iitr/Athena/athena/types"

	"gorm.io/gorm"
	"github.com/BlocSoc-iitr/Athena/athena/datbase/models"
	
	"log"
	"os"
	"encoding/json"
	"path/filepath"
	"sort"
	"time"
	"fmt"
)

func FetchBackfillsByDatatype(db *gorm.DB, dataType types.BackfillDataType, network types.SupportedNetwork) []models.BackfilledRange{
	var backFilledRanges []models.BackfilledRange

	err := db.Where("data_type = ? AND network = ?", dataType, network).Order("start_block").Find(&backFilledRanges).Error
	if err != nil {
		log.Fatalf("Error in fetching backfill ranges: %v", err)
	}

	return backFilledRanges
}

func FetchBackfillsByID(db *gorm.DB, backfillID interface{}) ([]models.BackfilledRange) {
    var backfilledRanges []models.BackfilledRange

    if id, ok := backfillID.(string); ok {
        backfillID = []string{id}
    }

    err := db.Where("backfill_id IN ?", backfillID).
        Find(&backfilledRanges).Error
    if err != nil {
		log.Fatalf("Error in fetching backfill ranges: %v", err)
    }

    return backfilledRanges
}

func GetAbis(db *gorm.DB, abiNames []string, decoderOS string) ([]models.ContractABI) {
    if db != nil {
        var contractABIs []models.ContractABI
        query := db.Where("decoder_os = ?", decoderOS)

        if len(abiNames) > 0 {
            query = query.Where("abi_name IN ?", abiNames)
        }

        query.Order("priority DESC, abi_name").Find(&contractABIs)
		if query.Error != nil {
			log.Fatalf("Error in fetching abis: %v", query.Error)
		}
        return contractABIs
    }

    appDir, err := os.UserConfigDir()
    if err != nil {
		log.Fatalf("Error in fetching abis: %v", err)
    }

    appDir = filepath.Join(appDir, "athena")

    if _, err := os.Stat(appDir); os.IsNotExist(err) {
        if err := os.Mkdir(appDir, os.ModePerm); err != nil {
			log.Fatalf("Error in fetching abis: %v", err)
        }
    }

    contractPath := filepath.Join(appDir, "contract-abis.json")
    if _, err := os.Stat(contractPath); os.IsNotExist(err) {
        return []models.ContractABI{}
    }

    fileData, err := os.ReadFile(contractPath)
    if err != nil {
        log.Fatalf("Error in fetching abis: %v", err)
    }

    if len(fileData) == 0 {
        return []models.ContractABI{}
    }

    var abiJSON []models.ContractABI
    if err := json.Unmarshal(fileData, &abiJSON); err != nil {
		log.Fatalf("Error in fetching abis: %v", err)
	}

	contains := func(slice []string, item string) bool{
		for _, s := range slice {
			if s == item {
				return true
			}
		}
		return false
	}

    var contractABIs []models.ContractABI
    for _, abi := range abiJSON {
        if (len(abiNames) == 0 || contains(abiNames, abi.AbiName)) && abi.DecoderOS == decoderOS {
            contractABIs = append(contractABIs, abi)
        }
    }

    return contractABIs
}

func FirstBlockTimestamp(network types.SupportedNetwork) (time.Time) {
	switch network {
	case types.StarkNet:
		return time.Date(2021, 11, 16, 13, 24, 8, 0, time.UTC)
	default:
		log.Fatalf("Cannot fetch Initial Block Time for Network: %s", network)
		return time.Time{}
	}
}

func GetBlockTimestamps(db *gorm.DB, network types.SupportedNetwork, resolution int, fromBlock int64) ([]types.BlockTimestamp) {
	var blockTimestamps []types.BlockTimestamp

	if db != nil {
		var results []models.Block
		err := db.Table("blocks").
			Where("block_number % ? = 0 AND block_number >= ?", resolution, fromBlock).
			Order("block_number").
			Select("block_number, timestamp").
			Scan(&results).Error
		if err != nil {
			log.Fatalf("Error in getting block timestamps: %v", err)
		}

		for _, row := range results {
			timestamp := time.Unix(row.Timestamp, 0).UTC()
			if row.BlockNumber == 0 {
				t := FirstBlockTimestamp(network)
				timestamp = t
			}
			blockTimestamps = append(blockTimestamps, types.BlockTimestamp{
				BlockNumber: int(row.BlockNumber),
				Timestamp:   timestamp,
			})
		}
		return blockTimestamps	
	}

	appDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Error in getting block timestamps: %v", err)
	}
	appDir = filepath.Join(appDir, "entro")

	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		err = os.MkdirAll(appDir, os.ModePerm)
		if err != nil {
			log.Fatalf("Error in getting block timestamps: %v", err)
		}
	}

	filePath := filepath.Join(appDir, fmt.Sprintf("%s-timestamps.json", network.String()))

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []types.BlockTimestamp{}
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error in getting block timestamps: %v", err)
	}
	defer file.Close()

	var timestampJson []types.BlockTimestamp
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&timestampJson); err != nil {
		log.Fatalf("Error in getting block timestamps: %v", err)
	}

	var filteredTimestamps []types.BlockTimestamp
	for _, t := range timestampJson {
		if t.BlockNumber%int(resolution) == 0 {
			filteredTimestamps = append(filteredTimestamps, t)
		}
	}

	sort.Slice(filteredTimestamps, func(i, j int) bool {
		return filteredTimestamps[i].BlockNumber < filteredTimestamps[j].BlockNumber
	})

	return filteredTimestamps
}