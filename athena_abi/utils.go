package athena_abi

import (
	"io/fs"
	"math/big"

	"golang.org/x/crypto/sha3"

	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func bigIntToBytes(value big.Int, length int) []byte {
	b := make([]byte, length)
	for i := length - 1; i >= 0; i-- {
		b[i] = byte(new(big.Int).And(&value, big.NewInt(0xff)).Int64())
		value = *value.Rsh(&value, 8)
	}
	return b
}

func StarknetKeccak(data []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)

	var masked big.Int
	masked.SetBytes(hash.Sum(nil))
	masked = *masked.And(&masked, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 250), big.NewInt(1)))
	return bigIntToBytes(masked, 32)
}

func TopologicalSort(graph map[string][]string) []string {
	inDegree := make(map[string]int)
	order := []string{}
	queue := []string{}

	for node := range graph {
		inDegree[node] = 0
	}

	for _, neighbours := range graph {
		for _, neighbour := range neighbours {
			inDegree[neighbour]++
		}
	}

	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		order = append(order, current)

		for _, neighbour := range graph[current] {
			inDegree[neighbour]--
			if inDegree[neighbour] == 0 {
				queue = append(queue, neighbour)
			}
		}
	}
	return order
}

func convertMap(input map[string]map[string]bool) map[string][]string {
	result := make(map[string][]string)

	for outerKey, innerMap := range input {
		var trueKeys []string

		for innerKey, value := range innerMap {
			if value {
				trueKeys = append(trueKeys, innerKey)
			}
		}

		result[outerKey] = trueKeys
	}

	return result
}

func loadABI(abiName string, abiVersion int) (map[string]interface{}, error) {
	abiFilePath := filepath.Join("abis", fmt.Sprintf("v%d", abiVersion), abiName+".json")
	abiFile, err := os.Open(abiFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ABI file: %w", err)
	}
	defer abiFile.Close()

	var abiData map[string]interface{}
	decoder := json.NewDecoder(abiFile)
	if err := decoder.Decode(&abiData); err != nil {
		return nil, fmt.Errorf("failed to decode ABI JSON: %w", err)
	}

	return abiData, nil
}

func GetAbisForVersion(abiVersion string) (map[string][]map[string]interface{}, error) {
	abiDir := filepath.Join("abis", abiVersion)
	abis := make(map[string][]map[string]interface{})

	err := filepath.Walk(abiDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check for JSON files only
		if filepath.Ext(path) == ".json" {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			var rawData interface{}
			if err := json.NewDecoder(file).Decode(&rawData); err != nil {
				return err
			}

			// Check if rawData is of type []interface{}
			abiList, ok := rawData.([]interface{})
			if !ok {
				return fmt.Errorf("expected ABI data to be []interface{}, got %T", rawData)
			}

			// Convert []interface{} to []map[string]interface{}
			var abiData []map[string]interface{}
			for _, item := range abiList {
				abiMap, ok := item.(map[string]interface{})
				if !ok {
					return fmt.Errorf("expected item to be map[string]interface{}, got %T", item)
				}
				abiData = append(abiData, abiMap)
			}

			abis[info.Name()] = abiData
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load ABIs: %w", err)
	}

	return abis, nil
}
