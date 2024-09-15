package athena_abi

import (
	"fmt"
	"math/big"
)

func GetEnumIndex(enumType StarknetEnum, enumKey string) (int, StarknetType, error) {
	for idx, variant := range enumType.Variants {
		if variant.Name == enumKey {
			return idx, variant.Type, nil
		}
	}
	return 0, nil, fmt.Errorf("enum key %s not found in enum %s", enumKey, enumType.Name)
}

func EncodeCoreType(encodeType StarknetCoreType, value interface{}) ([]*big.Int, error) {
	switch encodeType {
	case U8, U16, U32, U64, U128, U256:
		var bigIntValue *big.Int
		switch v := value.(type) {
		case *big.Int:
			bigIntValue = v
		default:
			return nil, &TypeEncodeError{Msg: fmt.Sprintf("cannot encode value of type %T to %s", value, encodeType)}
		}
		maxValue, _ := encodeType.maxValue()
		if bigIntValue.Sign() < 0 || bigIntValue.Cmp(maxValue) > 0 {
			return nil, &TypeEncodeError{Msg: fmt.Sprintf("value %s is out of range for %s", bigIntValue.String(), encodeType)}
		}
		if encodeType == U256 {
			high := new(big.Int).Rsh(bigIntValue, 128)
			low := new(big.Int).And(bigIntValue, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)))
			return []*big.Int{low, high}, nil
		}
		return []*big.Int{bigIntValue}, nil

	case Bool:
		boolValue, ok := value.(bool)
		if !ok {
			return nil, &TypeEncodeError{Msg: fmt.Sprintf("cannot encode non-boolean value '%v' to %s", value, encodeType)}
		}
		if boolValue {
			return []*big.Int{big.NewInt(1)}, nil
		}
		return []*big.Int{big.NewInt(0)}, nil

	case Felt, ClassHash, ContractAddress, EthAddress, StorageAddress, Bytes31:
		var intEncoded *big.Int
		switch v := value.(type) {
		case string:
			if len(v) > 2 && v[:2] == "0x" {
				var ok bool
				intEncoded, ok = new(big.Int).SetString(v[2:], 16)
				if !ok {
					return nil, &TypeEncodeError{Msg: fmt.Sprintf("invalid hex string: %s", v)}
				}
			} else {
				return nil, &TypeEncodeError{Msg: fmt.Sprintf("hex strings must be 0x prefixed: %s", v)}
			}
		case *big.Int:
			intEncoded = v
		case []byte:
			intEncoded = new(big.Int).SetBytes(v)
		default:
			return nil, &TypeEncodeError{Msg: fmt.Sprintf("cannot encode type %T to %s", value, encodeType)}
		}

		maxValue, _ := encodeType.maxValue()
		if intEncoded.Sign() < 0 || intEncoded.Cmp(maxValue) > 0 {
			return nil, &TypeEncodeError{Msg: fmt.Sprintf("%s does not fit into %s", intEncoded.String(), encodeType)}
		}

		return []*big.Int{intEncoded}, nil

	case NoneType:
		return []*big.Int{}, nil

	default:
		return nil, &TypeEncodeError{Msg: fmt.Sprintf("unable to encode type %s", encodeType)}
	}
}

func EncodeFromTypes(types []StarknetType, values []interface{}) ([]*big.Int, error) {
	var encodedCalldata []*big.Int

	for i, encodeType := range types {
		encodeValue := values[i]

		switch t := encodeType.(type) {
		case StarknetCoreType:
			encoded, err := EncodeCoreType(t, encodeValue)
			if err != nil {
				return nil, err
			}
			encodedCalldata = append(encodedCalldata, encoded...)

		case StarknetArray:
			arrayValue, ok := encodeValue.([]interface{})
			if !ok {
				return nil, &TypeEncodeError{Msg: fmt.Sprintf("%v cannot be encoded into a StarknetArray", encodeValue)}
			}
			encodedCalldata = append(encodedCalldata, big.NewInt(int64(len(arrayValue))))
			for _, arrayElement := range arrayValue {
				encoded, err := EncodeFromTypes([]StarknetType{t.InnerType}, []interface{}{arrayElement})
				if err != nil {
					return nil, err
				}
				encodedCalldata = append(encodedCalldata, encoded...)
			}

		case StarknetOption:
			if encodeValue == nil {
				encodedCalldata = append(encodedCalldata, big.NewInt(1))
			} else {
				encodedCalldata = append(encodedCalldata, big.NewInt(0))
				encoded, err := EncodeFromTypes([]StarknetType{t.InnerType}, []interface{}{encodeValue})
				if err != nil {
					return nil, err
				}
				encodedCalldata = append(encodedCalldata, encoded...)
			}

		case StarknetStruct:
			structValue, ok := encodeValue.(map[string]interface{})
			if !ok {
				return nil, &TypeEncodeError{Msg: fmt.Sprintf("%v cannot be encoded into a StarknetStruct", encodeValue)}
			}
			encoded, err := EncodeFromParams(t.Members, structValue)
			if err != nil {
				return nil, err
			}
			encodedCalldata = append(encodedCalldata, encoded...)

		case StarknetEnum:
			enumValue, ok := encodeValue.(map[string]interface{})
			if !ok {
				return nil, &TypeEncodeError{Msg: fmt.Sprintf("%v cannot be encoded into a StarknetEnum", encodeValue)}
			}
			if len(enumValue) != 1 {
				return nil, &TypeEncodeError{Msg: fmt.Sprintf("enum value %v must have exactly one key-value pair", enumValue)}
			}
			var enumKey string
			var enumValueInner interface{}
			for k, v := range enumValue {
				enumKey = k
				enumValueInner = v
				break
			}
			enumIndex, enumType, err := GetEnumIndex(t, enumKey)
			if err != nil {
				return nil, err
			}
			encodedCalldata = append(encodedCalldata, big.NewInt(int64(enumIndex)))
			encoded, err := EncodeFromTypes([]StarknetType{enumType}, []interface{}{enumValueInner})
			if err != nil {
				return nil, err
			}
			encodedCalldata = append(encodedCalldata, encoded...)

		case StarknetTuple:
			tupleValue, ok := encodeValue.([]interface{})
			if !ok {
				return nil, &TypeEncodeError{Msg: fmt.Sprintf("%v cannot be encoded into a StarknetTuple", encodeValue)}
			}
			for i, tupleType := range t.Members {
				encoded, err := EncodeFromTypes([]StarknetType{tupleType}, []interface{}{tupleValue[i]})
				if err != nil {
					return nil, err
				}
				encodedCalldata = append(encodedCalldata, encoded...)
			}

		case StarknetNonZero:
			encoded, err := EncodeFromTypes([]StarknetType{t.InnerType}, []interface{}{encodeValue})
			if err != nil {
				return nil, err
			}
			if len(encoded) == 0 || encoded[0].Sign() == 0 {
				return nil, &TypeEncodeError{Msg: fmt.Sprintf("zero value %v cannot encode to StarknetNonZero", encodeValue)}
			}
			encodedCalldata = append(encodedCalldata, encoded...)

		default:
			return nil, &TypeEncodeError{Msg: fmt.Sprintf("cannot encode %v for type %T", encodeValue, encodeType)}
		}
	}

	return encodedCalldata, nil
}

func EncodeFromParams(params []AbiParameter, encodeValues map[string]interface{}) ([]*big.Int, error) {
	if len(encodeValues) != len(params) {
		return nil, &InvalidCalldataError{Msg: fmt.Sprintf("number of encode values '%d' does not match number of ABI params '%d'", len(encodeValues), len(params))}
	}

	var parameterTypes []StarknetType
	var encodeValuesList []interface{}

	for _, param := range params {
		value, ok := encodeValues[param.Name]
		if !ok {
			return nil, &InvalidCalldataError{Msg: fmt.Sprintf("missing encode value for param: %s", param.Name)}
		}
		parameterTypes = append(parameterTypes, param.Type)
		encodeValuesList = append(encodeValuesList, value)
	}

	return EncodeFromTypes(parameterTypes, encodeValuesList)
}