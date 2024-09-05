package decoder

import (
	"encoding/binary"
	"fmt"

	"github.com/BlocSoc-iitr/Athena/athena/types"
	
)

type CairoEventDecoder struct {
	types.AbiEvent
	Priority      int
	AbiName       string
	IndexedParams int
	Name          string
}

func NewCairoEventDecoder(
	name string,
	parameters []string,
	data map[string]types.StarknetType,
	keys map[string]types.StarknetType,
	abiName string,
	priority int,
) *CairoEventDecoder {
	return &CairoEventDecoder{
		AbiEvent: types.AbiEvent{
			Name:       name,
			Parameters: parameters,
			Data:       data,
			Keys:       keys,
		},
		Priority:      priority,
		AbiName:       abiName,
		IndexedParams: len(keys),
	}
}

// func (d *CairoEventDecoder) Decode(data [][]byte, keys [][]byte) (*types.DecodedEvent, error) {
// 	decodedData := make(map[string]interface{})
// 	for paramName, starknetType := range d.AbiEvent.Data {
// 		decodedValue, err := starknetType.Decode(data[0])
// 		if err != nil {
// 			return nil, err
// 		}
// 		decodedData[paramName] = decodedValue
// 		data = data[1:]
// 	}

//	for paramName, starknetType := range d.AbiEvent.Keys {
//		decodedValue, err := starknetType.Decode(keys[0])
//		if err != nil {
//			return nil, err
//		}
//		decodedData[paramName] = decodedValue
//		keys = keys[1:]
//	}
func (e *CairoEventDecoder) Decode(data [][]byte, keys [][]byte) (types.DecodedEvent, error) {
	_data := make([]int, len(data))
	for i, d := range data {
		if len(d) > 8 {
			return types.DecodedEvent{}, fmt.Errorf("Invalid data length at index %d", i)
		}

		_data[i] = int(int64(binary.BigEndian.Uint64(append(make([]byte, 8-len(d)), d...))))
	}

	_keys := make([]int, len(keys)-1)
	for i := 1; i < len(keys); i++ {
		if len(keys[i]) > 8 {
			return types.DecodedEvent{}, fmt.Errorf("Invalid key length at index %d", i)
		}

		_keys[i-1] = int(int64(binary.BigEndian.Uint64(append(make([]byte, 8-len(keys[i])), keys[i]...))))
	}

	decodedData := make(map[string]interface{})

	for _, param := range e.Parameters {
		var value interface{}
		var err error

		if _, ok := e.Data[param]; ok {
			// Decode from data
			value, err = types.DecodeFromTypes([]types.StarknetType{e.Data[param]}, &_data)
		} else if _, ok := e.Keys[param]; ok {
			// Decode from keys
			value, err = types.DecodeFromTypes([]types.StarknetType{e.Keys[param]}, &_keys)
		} else {
			return types.DecodedEvent{}, fmt.Errorf("event Parameter %s not present in Keys or Data for Event %s", param, e.Name)
		}

		if err != nil {
			return types.DecodedEvent{}, err
		}
		decodedData[param] = value
	}

	if len(_data) != 0 || len(_keys) != 0 {
		return types.DecodedEvent{}, fmt.Errorf("calldata Not Completely Consumed decoding Event: %s", e.idStr())
	}

	return types.DecodedEvent{
		AbiName: e.AbiName,
		Name:    e.Name,
		Data:    decodedData,
	}, nil

}

func (d *CairoEventDecoder) idStr() string {
	return d.Name
}
