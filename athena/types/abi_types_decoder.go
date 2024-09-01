package types

import (
	"fmt"
	"math/big"
)

func DecodeCoreType(decodeType StarknetCoreType, calldata *[]int) (interface{}, string) {
	if len(*calldata) == 0 {
		return nil, "Not enough calldata to decode"
	}

	switch decodeType {
	case U8, U16, U32, U64, U128:
		value := (*calldata)[0]
		*calldata = (*calldata)[1:]
		maxValue, _ := decodeType.MaxValue()
		if value < 0 || big.NewInt(int64(value)).Cmp(maxValue) > 0 {
			return nil, fmt.Sprintf("%d exceeds %v max range", value, decodeType.String())
		}
		return value, ""

	case U256:
		if len(*calldata) < 2 {
			return nil, "Not enough calldata to decode U256"
		}
		low := (*calldata)[0]
		high := (*calldata)[1]
		*calldata = (*calldata)[2:]

		lowBig := big.NewInt(int64(low))
		highBig := big.NewInt(int64(high))
		maxValue, _ := decodeType.MaxValue()

		uint256 := new(big.Int).Add(lowBig, new(big.Int).Lsh(highBig, 128))
		if uint256.Cmp(maxValue) > 0 {
			return nil, "U256 value exceeds max range"
		}
		return uint256, ""

	case Bool:
		value := (*calldata)[0]
		*calldata = (*calldata)[1:]
		if value != 0 && value != 1 {
			return nil, "Bool value must be 0 or 1"
		}
		return value == 1, ""

	case Felt:
		value := (*calldata)[0]
		*calldata = (*calldata)[1:]
		maxValue, _ := decodeType.MaxValue()
		if value < 0 || big.NewInt(int64(value)).Cmp(maxValue) > 0 {
			return nil, fmt.Sprintf("%d exceeds Felt max range", value)
		}
		return fmt.Sprintf("0x%x", value), ""

	case ContractAddress, ClassHash, StorageAddress:
		value := (*calldata)[0]
		*calldata = (*calldata)[1:]
		maxValue, _ := decodeType.MaxValue()
		if value < 0 || big.NewInt(int64(value)).Cmp(maxValue) > 0 {
			return nil, fmt.Sprintf("%d exceeds %v max range", value, decodeType.String())
		}
		return fmt.Sprintf("0x%064x", value), ""

	case EthAddress:
		value := (*calldata)[0]
		*calldata = (*calldata)[1:]
		maxValue, _ := decodeType.MaxValue()
		if value < 0 || big.NewInt(int64(value)).Cmp(maxValue) > 0 {
			return nil, fmt.Sprintf("%d exceeds EthAddress max range", value)
		}
		return fmt.Sprintf("0x%040x", value), ""

	case Bytes31:
		value := (*calldata)[0]
		*calldata = (*calldata)[1:]
		maxValue, _ := decodeType.MaxValue()
		if value < 0 || big.NewInt(int64(value)).Cmp(maxValue) > 0 {
			return nil, fmt.Sprintf("%d exceeds Bytes31 max range", value)
		}
		return fmt.Sprintf("0x%062x", value), ""

	case NoneType:
		return "", ""

	default:
		return nil, fmt.Sprintf("Unable to decode Starknet Core type: %v", decodeType.String())
	}
}

func DecodeFromTypes(Types []StarknetType, calldata *[]int) ([]interface{}, error) {
	var outputData []interface{}

	for _, starknetType := range Types {
		if t, ok := starknetType.(StarknetCoreType); ok {
			decodedValue, _ := DecodeCoreType(t, calldata)
			// if err != nil {
			// 	return nil, fmt.Errorf("failed to decode core type: %w", err)
			// }
			outputData = append(outputData, decodedValue)

		} else if t, ok := starknetType.(StarknetArray); ok {
			if len(*calldata) < 1 {
				return nil, fmt.Errorf("insufficient calldata to decode StarknetArray")
			}
			arrayLen := int((*calldata)[0])
			*calldata = (*calldata)[1:]

			var arrayValues []interface{}
			for i := 0; i < arrayLen; i++ {
				decodedValues, err := DecodeFromTypes([]StarknetType{t.InnerType}, calldata)
				if err != nil {
					return nil, err
				}
				arrayValues = append(arrayValues, decodedValues[0])
			}
			outputData = append(outputData, arrayValues)

		} else if t, ok := starknetType.(StarknetOption); ok {
			if len(*calldata) < 1 {
				return nil, fmt.Errorf("insufficient calldata to decode StarknetOption")
			}
			optionFlag := (*calldata)[0]
			*calldata = (*calldata)[1:]

			if optionFlag == 1 {
				outputData = append(outputData, nil)
			} else {
				decodedValues, err := DecodeFromTypes([]StarknetType{t.InnerType}, calldata)
				if err != nil {
					return nil, err
				}
				outputData = append(outputData, decodedValues[0])
			}

		} else if t, ok := starknetType.(StarknetStruct); ok {
			decodedStruct, err := DecodeFromParams(t.Members, calldata)
			if err != nil {
				return nil, err
			}
			outputData = append(outputData, decodedStruct)

		} else if t, ok := starknetType.(StarknetEnum); ok {
			if len(*calldata) < 1 {
				return nil, fmt.Errorf("insufficient calldata to decode StarknetEnum")
			}
			enumIndex := int((*calldata)[0])
			*calldata = (*calldata)[1:]

			if enumIndex < 0 || enumIndex >= len(t.Variants) {
				return nil, fmt.Errorf("invalid enum index %d for StarknetEnum", enumIndex)
			}

			variant := t.Variants[enumIndex]
			decodedValues, err := DecodeFromTypes([]StarknetType{variant.VariantType}, calldata)
			if err != nil {
				return nil, err
			}
			outputData = append(outputData, map[string]interface{}{variant.VariantName: decodedValues[0]})

		} else if t, ok := starknetType.(StarknetTuple); ok {
			var tupleValues []interface{}
			for _, tupleMember := range t.Members {
				decodedValues, err := DecodeFromTypes([]StarknetType{tupleMember}, calldata)
				if err != nil {
					return nil, err
				}
				tupleValues = append(tupleValues, decodedValues[0])
			}
			outputData = append(outputData, tupleValues)

		} else {
			return nil, fmt.Errorf("cannot decode calldata for type: %T", starknetType)
		}
	}

	return outputData, nil
}

func DecodeFromParams(params []AbiParameter, calldata *[]int) (map[string]interface{}, error) {
	parameterNames := make([]string, len(params))
	parameterTypes := make([]StarknetType, len(params))

	for i, param := range params {
		parameterNames[i] = param.Name
		parameterTypes[i] = param.Type
	}

	decodedValues, err := DecodeFromTypes(parameterTypes, calldata)
	if err != nil {
		return nil, fmt.Errorf("failed to decode from types: %w", err)
	}

	result := make(map[string]interface{}, len(params))
	for i, name := range parameterNames {
		result[name] = decodedValues[i]
	}

	return result, nil
}
