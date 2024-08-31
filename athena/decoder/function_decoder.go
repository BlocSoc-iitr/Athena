package decoder

import (
	"encoding/binary"
	"github.com/DarkLord017/athena/athena/types"
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

func (cfd *CairoFunctionDecoder) IDStr(fullSignature bool) string {
	if fullSignature {
		return cfd.AbiFunction.IdStr()
	}
	return cfd.AbiFunction.Name
}
