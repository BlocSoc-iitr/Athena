package athenaabi

import (
	"fmt"
	"math/big"
	"strings"
)

func pop(nums *[]*big.Int) (big.Int, error) {
	if len(*nums) == 0 {
		return *big.NewInt(-1), fmt.Errorf("cannot pop from an empty slice")
	}

	lastElement := *(*nums)[0]
	*nums = (*nums)[1:]

	return lastElement, nil
}

// Decodes Calldata using Starknet Core Type. The function Takes in two parameters, a StarknetCoreType, and a mutable reference
// to a calldata array and returns either a string, an int or a bool(hence we have taken return type as an interface).
// When decoding, calldata is popped off the top of the calldata array. This reference to the calldata array is
// recursively passed between type decoders, so this array is modified during decoding.
func DecodeCoreTypes(decodeType StarknetCoreType, callData []*big.Int) (interface{}, error) {
	switch decodeType {
	case U8, U16, U32, U64, U128:
		decoded, err := pop(&callData)
		if err != nil {
			return nil, &InvalidCalldataError{
				Msg: fmt.Sprintf("not enough calldata to decode %s", decodeType.idStr()),
			}
		}
		decodeTypeMaxVal, _ := decodeType.maxValue()
		if decoded.Cmp(big.NewInt(0)) >= 0 && decoded.Cmp(decodeTypeMaxVal) <= 0 {
			return decoded, nil
		} else {
			return nil, fmt.Errorf("%s exceeds %s Max Range", &decoded, decodeType.idStr())
		}
	case U256:
		decodedLow, err := pop(&callData)
		if err != nil {
			return nil, &InvalidCalldataError{
				Msg: fmt.Sprintf("not enough calldata to decode %s", decodeType.idStr()),
			}
		}
		decodedHigh, err := pop(&callData)
		if err != nil {
			return nil, &InvalidCalldataError{
				Msg: fmt.Sprintf("not enough calldata to decode %s", decodeType.idStr()),
			}
		}
		decodeTypeMaxVal, _ := decodeType.maxValue()
		if decodedLow.Cmp(big.NewInt(0)) < 0 || decodedLow.Cmp(decodeTypeMaxVal) > 0 {
			return nil, fmt.Errorf("low Exceeds U128 range")
		}
		if decodedHigh.Cmp(big.NewInt(0)) < 0 || decodedHigh.Cmp(decodeTypeMaxVal) > 0 {
			return nil, fmt.Errorf("high Exceeds U128 range")
		}
		return new(big.Int).Add(&decodedLow, new(big.Int).Lsh(&decodedHigh, 128)), nil
	case Bool:
		decoded, err := pop(&callData)
		if err != nil {
			return nil, &InvalidCalldataError{
				Msg: fmt.Sprintf("not enough calldata to decode %s", decodeType.idStr()),
			}
		}
		if decoded.Cmp(big.NewInt(0)) == 0 {
			return false, nil
		} else if decoded.Cmp(big.NewInt(1)) == 0 {
			return true, nil
		} else {
			return nil, fmt.Errorf("invalid Bool Value")
		}
	case Felt:
		decoded, err := pop(&callData)
		if err != nil {
			return nil, &InvalidCalldataError{
				Msg: fmt.Sprintf("not enough calldata to decode %s", decodeType.idStr()),
			}
		}
		decodeTypeMaxVal, _ := decodeType.maxValue()
		if decoded.Cmp(big.NewInt(0)) < 0 || decoded.Cmp(decodeTypeMaxVal) > 0 {
			return nil, fmt.Errorf("%s larger than Felt", &decoded)
		}
		decodedHexStr := decoded.Text(16)
		if len(decodedHexStr)%2 != 0 {
			return "0x0" + decodedHexStr, nil
		}
		return "0x" + decodedHexStr, nil
	case ContractAddress, ClassHash, StorageAddress:
		decoded, err := pop(&callData)
		if err != nil {
			return nil, &InvalidCalldataError{
				Msg: fmt.Sprintf("not enough calldata to decode %s", decodeType.idStr()),
			}
		}
		decodeTypeMaxVal, _ := decodeType.maxValue()
		if decoded.Cmp(big.NewInt(0)) < 0 || decoded.Cmp(decodeTypeMaxVal) > 0 {
			return nil, fmt.Errorf("%s larger than Felt Address", &decoded)
		}
		decodedHexStr := decoded.Text(16)
		if len(decodedHexStr) < 64 {
			decodedHexStr = strings.Repeat("0", 64-len(decodedHexStr)) + decodedHexStr
		}
		return "0x" + decodedHexStr, nil
	case EthAddress:
		decoded, err := pop(&callData)
		if err != nil {
			return nil, &InvalidCalldataError{
				Msg: fmt.Sprintf("not enough calldata to decode %s", decodeType.idStr()),
			}
		}
		decodeTypeMaxVal, _ := decodeType.maxValue()
		if decoded.Cmp(big.NewInt(0)) < 0 || decoded.Cmp(decodeTypeMaxVal) > 0 {
			return nil, fmt.Errorf("%s larger than Felt Address", &decoded)
		}
		decodedHexStr := decoded.Text(16)
		if len(decodedHexStr) < 40 {
			decodedHexStr = strings.Repeat("0", 40-len(decodedHexStr)) + decodedHexStr
		}
		return "0x" + decodedHexStr, nil
	case Bytes31:
		decoded, err := pop(&callData)
		if err != nil {
			return nil, &InvalidCalldataError{
				Msg: fmt.Sprintf("not enough calldata to decode %s", decodeType.idStr()),
			}
		}
		decodeTypeMaxVal, _ := decodeType.maxValue()
		if decoded.Cmp(big.NewInt(0)) < 0 || decoded.Cmp(decodeTypeMaxVal) > 0 {
			return nil, fmt.Errorf("%s larger than Felt Address", &decoded)
		}
		decodedHexStr := decoded.Text(16)
		if len(decodedHexStr) < 62 {
			decodedHexStr = strings.Repeat("0", 62-len(decodedHexStr)) + decodedHexStr
		}
		return "0x" + decodedHexStr, nil
	case NoneType:
		return "", nil
	default:
		return nil, &TypeDecodeError{
			Msg: fmt.Sprintf("unable to decode Starknet Core type: %s", decodeType),
		}
	}
}

// Decodes calldata array using a list of StarknetTypes.
func DecodeFromTypes(types []StarknetType, callData []*big.Int) ([]interface{}, error) {
	var outputData []interface{}

	for _, starknet_type := range types {
		switch t := starknet_type.(type) {
		case StarknetCoreType:
			decoded, err := DecodeCoreTypes(t, callData)
			if err != nil {
				return nil, err
			}
			outputData = append(outputData, decoded)
		case StarknetArray:
			arrayLen, err := pop(&callData)
			if err != nil {
				return nil, &InvalidCalldataError{
					Msg: fmt.Sprintf("not enough calldata to decode %s", starknet_type.idStr()),
				}
			}
			for i := 0; i < int(arrayLen.Int64()); i++ {
				decoded, err := DecodeFromTypes([]StarknetType{t.InnerType}, callData)
				if err != nil {
					return nil, err
				}
				outputData = append(outputData, decoded...)
			}
		case StarknetOption:
			optionPresent, err := pop(&callData)
			if err != nil {
				return nil, &InvalidCalldataError{
					Msg: fmt.Sprintf("not enough calldata to decode %s", starknet_type.idStr()),
				}
			}
			if optionPresent.Cmp(big.NewInt(0)) == 0 {
				outputData = append(outputData, nil)
			} else {
				decoded, err := DecodeFromTypes([]StarknetType{t.InnerType}, callData)
				if err != nil {
					return nil, err
				}
				outputData = append(outputData, decoded...)
			}
		case StarknetEnum:
			enumIndex, err := pop(&callData)
			if err != nil {
				return nil, &InvalidCalldataError{
					Msg: fmt.Sprintf("not enough calldata to decode %s", starknet_type.idStr()),
				}
			}
			variantName, variantType := t.Variants[int(enumIndex.Int64())].Name, t.Variants[int(enumIndex.Int64())].Type
			decoded, err := DecodeFromTypes([]StarknetType{variantType}, callData)
			if err != nil {
				return nil, err
			}
			outputData = append(outputData, map[string]interface{}{variantName: decoded})
		case StarknetTuple:
			for _, tupleMembers := range t.Members {
				decoded, err := DecodeFromTypes([]StarknetType{tupleMembers}, callData)
				if err != nil {
					return nil, err
				}
				outputData = append(outputData, decoded...)
			}
		case StarknetNonZero:
			decoded, err := DecodeFromTypes([]StarknetType{t.InnerType}, callData)
			if err != nil {
				return nil, err
			}

			if decoded[0].(int) == 0 {
				return nil, fmt.Errorf("zero Value Encoded in StarknetNonZero")
			}
			outputData = append(outputData, decoded...)
		default:
			return nil, &TypeDecodeError{
				Msg: fmt.Sprintf("unable to decode Starknet type: %s", starknet_type),
			}
		}
	}

	return outputData, nil
}

// Decodes Calldata using AbiParameters, which have names and types
func DecodeFromParams(params []AbiParameter, callData []*big.Int) (map[string]interface{}, error) {
	var parameterNames []string
	var parameterTypes []StarknetType

	for _, param := range params {
		parameterNames = append(parameterNames, param.Name)
		parameterTypes = append(parameterTypes, param.Type)
	}

	decodedData, err := DecodeFromTypes(parameterTypes, callData)
	if err != nil {
		return nil, err
	}

	outputData := map[string]interface{}{}

	for i, name := range parameterNames {
		outputData[name] = decodedData[i]
	}

	return outputData, nil
}
