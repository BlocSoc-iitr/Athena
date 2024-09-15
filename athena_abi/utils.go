package athena_abi

import (
	"math/big"

	"golang.org/x/crypto/sha3"
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
