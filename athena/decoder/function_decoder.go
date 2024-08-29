package decoder

import (
	"encoding/binary"
	"starknet-athena/athena/types"
)

type CairoFunctionDecoder struct {
	*types.AbiFunction
	Priority int
	AbiName  string
}

func NewCairoFunctionDecoder(name string, inputs []types.AbiParameter, outputs []types.StarknetType, abiName string, priority int) *CairoFunctionDecoder {
	return &CairoFunctionDecoder{
		AbiFunction: types.NewAbiFunction(name, inputs, outputs, abiName),
		Priority:    priority,
		AbiName:     abiName,
	}
}

// Decode converts byte arrays to integers and calls the parent Decode method.
func (cfd *CairoFunctionDecoder) Decode(calldata [][]byte, result [][]byte) (*types.DecodedFunction, error) {
	intCalldata := make([]int, len(calldata))
	for i, data := range calldata {
		intCalldata[i] = int(binary.BigEndian.Uint64(data))
	}

	var intResult []int
	if result != nil {
		intResult = make([]int, len(result))
		for i, res := range result {
			intResult[i] = int(binary.BigEndian.Uint64(res))
		}
	}

	decodedFunction := cfd.AbiFunction.Decode(intCalldata, intResult)
	return &decodedFunction, nil
}

// IDStr returns the string representation of the function, either with a full signature or just the name.
func (cfd *CairoFunctionDecoder) IDStr(fullSignature bool) string {
	if fullSignature {
		return cfd.AbiFunction.IdStr()
	}
	return cfd.AbiFunction.Name
}

// func decodeFromParams(params []types.AbiParameter, calldata []int) map[string]interface{} {

// 	return make(map[string]interface{})
// }

// func decodeFromTypes(types []types.StarknetType, result []int) []interface{} {

// 	return make([]interface{}, len(types))
// }
