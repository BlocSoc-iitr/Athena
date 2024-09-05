package decoder

import (
	"encoding/json"
	"github.com/BlocSoc-iitr/Athena/athena/types"
	
)
type Function1 struct {
	Name    string   `json:"name"`
	Inputs  []types.AbiParameter `json:"inputs"`
	Outputs []types.StarknetType `json:"outputs"`
}

type Event struct {
	Name       string   `json:"name"`
	Parameters []string `json:"parameters"`
	Data       map[string]types.StarknetType `json:"data"`
	Keys       map[string]types.StarknetType `json:"keys"`
}

type ABI struct {
	Functions map[string]Function `json:"functions"`
	Events    map[string]Event    `json:"events"`
}

func FromJSON(abiData []byte, abiName string, additionalData []byte) (ABI, error) {
	var abi ABI
	err := json.Unmarshal(abiData, &abi)
	if err != nil {
		return ABI{}, err
	}
	// Handle additionalData if needed
	return abi, nil
}