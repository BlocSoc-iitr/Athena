package decoder

import "starknet-athena/athena/types"

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

func (d *CairoEventDecoder) Decode(data [][]byte, keys [][]byte) (*types.DecodedEvent, error) {
	decodedData := make(map[string]interface{})
	for paramName, starknetType := range d.AbiEvent.Data {
		decodedValue, err := starknetType.Decode(data[0])
		if err != nil {
			return nil, err
		}
		decodedData[paramName] = decodedValue
		data = data[1:]
	}

	for paramName, starknetType := range d.AbiEvent.Keys {
		decodedValue, err := starknetType.Decode(keys[0])
		if err != nil {
			return nil, err
		}
		decodedData[paramName] = decodedValue
		keys = keys[1:]
	}

	return &types.DecodedEvent{
		AbiName:        d.AbiName,
		Name:           d.AbiEvent.Name,
		EventSignature: d.idStr(),
		Data:           decodedData,
	}, nil
}

func (d *CairoEventDecoder) idStr() string {
	return d.Name
}
