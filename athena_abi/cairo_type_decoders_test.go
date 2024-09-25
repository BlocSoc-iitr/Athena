package athena_abi

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayDecodingAndEncoding(t *testing.T) {
	tests := []struct {
		name         string
		starknetType StarknetType
		calldata     []*big.Int
		decoded      interface{}
	}{
		{
			name:         "Empty U256 Array",
			starknetType: StarknetArray{InnerType: U256},
			calldata:     []*big.Int{big.NewInt(0)},
			decoded:      []interface{}(nil),
		},
		{
			name:         "U256 Array with Two Elements",
			starknetType: StarknetArray{InnerType: U256},
			calldata:     []*big.Int{big.NewInt(2), big.NewInt(16), big.NewInt(0), big.NewInt(48), big.NewInt(0)},
			decoded:      []interface{}{big.NewInt(16), big.NewInt(48)},
		},
		{
			name: "Nested Arrays with U32",
			starknetType: StarknetArray{
				InnerType: StarknetArray{
					InnerType: StarknetArray{
						InnerType: U32,
					},
				},
			},
			calldata: []*big.Int{big.NewInt(1), big.NewInt(1), big.NewInt(2), big.NewInt(22), big.NewInt(38)},
			decoded: []interface{}{
				[]interface{}{
					[]interface{}{
						big.NewInt(22),
						big.NewInt(38),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Copy calldata
			calldata := make([]*big.Int, len(tt.calldata))
			copy(calldata, tt.calldata)

			// Test decoding
			decodedValues, err := DecodeFromTypes([]StarknetType{tt.starknetType}, &calldata)
			assert.NoError(t, err, "DecodeFromTypes should not return an error")
			assert.Equal(t, tt.decoded, decodedValues[0], "Decoded values should match expected")
			//here we check deep equality

			// Test encoding
			encodedCalldata, err := EncodeFromTypes([]StarknetType{tt.starknetType}, []interface{}{tt.decoded})
			assert.NoError(t, err, "EncodeFromTypes should not return an error")
			assert.Equal(t, tt.calldata, encodedCalldata, "Encoded calldata should match original")

			// Check if calldata is empty after decoding
			assert.Empty(t, calldata, "Calldata should be empty after decoding")
		})
	}
}

func TestEnumTypeSerialization(t *testing.T) {
	variedTypeEnum := StarknetEnum{
		Name: "Enum A",
		Variants: []struct {
			Name string
			Type StarknetType
		}{
			{Name: "a", Type: U256},
			{Name: "b", Type: U128},
			{Name: "c", Type: StarknetStruct{
				Name: "Struct A",
				Members: []AbiParameter{
					{Name: "my_option", Type: StarknetOption{InnerType: U128}},
					{Name: "my_uint", Type: U256},
				},
			}},
		},
	}

	testCases := []struct {
		name     string
		calldata []*big.Int
		decoded  map[string]interface{}
		//encodingInput map[string]interface{}
	}{
		{
			name:     "Case 1",
			calldata: []*big.Int{big.NewInt(0), big.NewInt(100), big.NewInt(0)},
			//decoded:  map[string]interface{}{"a": []interface{}{big.NewInt(100)}},
			decoded: map[string]interface{}{"a": big.NewInt(100)}, // Same as previous decoded value
		},
		{
			name:     "Case 2",
			calldata: []*big.Int{big.NewInt(1), big.NewInt(200)},
			//decoded:  map[string]interface{}{"b": []interface{}{big.NewInt(200)}},
			decoded: map[string]interface{}{"b": big.NewInt(200)}, // Same as previous decoded value
		},
		{
			name:     "Case 3",
			calldata: []*big.Int{big.NewInt(2), big.NewInt(0), big.NewInt(300), big.NewInt(300), big.NewInt(0)},
			/*decoded: map[string]interface{}{
				"c": []interface{}{
					map[string]interface{}{
						"my_option": big.NewInt(300),
						"my_uint":   big.NewInt(300),
					},
				},
			},*/
			decoded: map[string]interface{}{
				"c": map[string]interface{}{ // Same as previous decoded value
					"my_option": big.NewInt(300),
					"my_uint":   big.NewInt(300),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			calldata := make([]*big.Int, len(tc.calldata))
			copy(calldata, tc.calldata)

			// Decode the calldata using the DecodeFromTypes function
			decodedValues, err := DecodeFromTypes([]StarknetType{variedTypeEnum}, &calldata)
			assert.NoError(t, err, "DecodeFromTypes should not return an error")

			// Assert that all calldata is consumed
			assert.Empty(t, calldata, "All calldata should be consumed during decoding")

			// Encode the values back into calldata using encodingInput
			encodedCalldata, err := EncodeFromTypes([]StarknetType{variedTypeEnum}, []interface{}{tc.decoded})
			assert.NoError(t, err, "EncodeFromTypes should not return an error")

			// Assert that the encoded calldata matches the original calldata
			assert.Equal(t, tc.calldata, encodedCalldata, "Encoded calldata should match original")

			// Assert that the decoded values match the expected decoded data
			assert.Equal(t, tc.decoded, decodedValues[0], "Decoded values should match expected")
		})
	}
}

func TestValidBoolValues(t *testing.T) {
	testCases := []struct {
		calldata []int64
		decoded  bool
	}{
		{[]int64{0}, false},
		{[]int64{1}, true},
	}

	for _, tc := range testCases {
		calldata := make([]*big.Int, len(tc.calldata))
		for i, v := range tc.calldata {
			calldata[i] = big.NewInt(v)
		}

		_calldata := make([]*big.Int, len(calldata))
		copy(_calldata, calldata)

		decodedValues, err := DecodeFromTypes([]StarknetType{Bool}, &_calldata)
		assert.NoError(t, err)

		encodedCalldata, err := EncodeFromTypes([]StarknetType{Bool}, []interface{}{tc.decoded})
		assert.NoError(t, err)

		assert.Equal(t, tc.decoded, decodedValues[0])
		assert.Equal(t, 1, len(decodedValues))
		assert.Equal(t, calldata, encodedCalldata)
		assert.Equal(t, 0, len(_calldata))
	}
}

func TestLiteralEnum(t *testing.T) {
	literalEnum := StarknetEnum{
		Name: "TxStatus",
		Variants: []struct {
			Name string
			Type StarknetType
		}{
			{Name: "Submitted", Type: NoneType},
			{Name: "Executed", Type: NoneType},
			{Name: "Finalized", Type: NoneType},
		},
	}

	testCases := []struct {
		calldata []int64
		decoded  map[string]interface{}
	}{
		{[]int64{0}, map[string]interface{}{"Submitted": ""}},
		{[]int64{1}, map[string]interface{}{"Executed": ""}},
		{[]int64{2}, map[string]interface{}{"Finalized": ""}},
	}

	for _, tc := range testCases {
		calldata := make([]*big.Int, len(tc.calldata))
		for i, v := range tc.calldata {
			calldata[i] = big.NewInt(v)
		}

		_calldata := make([]*big.Int, len(calldata))
		copy(_calldata, calldata)

		decodedValues, err := DecodeFromTypes([]StarknetType{literalEnum}, &_calldata)
		assert.NoError(t, err)

		encodedCalldata, err := EncodeFromTypes([]StarknetType{literalEnum}, []interface{}{tc.decoded})
		assert.NoError(t, err)

		assert.Equal(t, tc.decoded, decodedValues[0])
		assert.Equal(t, calldata, encodedCalldata)
		assert.Equal(t, 0, len(_calldata))
	}
}
func TestHexTypes(t *testing.T) {
	testCases := []struct {
		starknetType StarknetCoreType
		calldata     []*big.Int
		decoded      string
	}{
		{
			starknetType: Felt,
			calldata:     []*big.Int{big.NewInt(0x0123456789ABCDEF)},
			decoded:      "0x0123456789abcdef",
		},
		{
			starknetType: ContractAddress,
			calldata: func() []*big.Int {
				val, _ := new(big.Int).SetString("049D36570D4E46F48E99674BD3FCC84644DDD6B96F7C741B1562B82F9E004DC7", 16)
				return []*big.Int{val}
			}(),
			decoded: "0x049d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7",
		},
		{
			starknetType: ClassHash,
			calldata: func() []*big.Int {
				val, _ := new(big.Int).SetString("05FFBCFEB50D200A0677C48A129A11245A3FC519D1D98D76882D1C9A1B19C6ED", 16)
				return []*big.Int{val}
			}(),
			decoded: "0x05ffbcfeb50d200a0677c48a129a11245a3fc519d1d98d76882d1c9a1b19c6ed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.starknetType.String(), func(t *testing.T) {
			calldata := make([]*big.Int, len(tc.calldata))
			copy(calldata, tc.calldata)

			decodedValues, err := DecodeFromTypes([]StarknetType{tc.starknetType}, &calldata)
			assert.NoError(t, err)

			encodedCalldata, err := EncodeFromTypes([]StarknetType{tc.starknetType}, []interface{}{tc.decoded})
			assert.NoError(t, err)

			assert.Equal(t, tc.calldata, encodedCalldata)
			assert.Equal(t, tc.decoded, decodedValues[0])
			assert.Len(t, calldata, 0)
		})
	}
}

func TestOptionSerializer(t *testing.T) {
	testCases := []struct {
		starknetType StarknetType
		decoded      interface{}
		calldata     []*big.Int
	}{
		{
			starknetType: StarknetOption{InnerType: U128},
			decoded:      big.NewInt(123),
			calldata:     []*big.Int{big.NewInt(0), big.NewInt(123)},
		},
		{
			starknetType: StarknetOption{InnerType: U256},
			decoded:      big.NewInt(1),
			calldata:     []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(0)},
		},
		{
			starknetType: StarknetOption{InnerType: U128},
			decoded:      nil,
			calldata:     []*big.Int{big.NewInt(1)},
		},
		{
			starknetType: StarknetOption{InnerType: U256},
			decoded:      nil,
			calldata:     []*big.Int{big.NewInt(1)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.starknetType.idStr(), func(t *testing.T) {
			calldata := make([]*big.Int, len(tc.calldata))
			copy(calldata, tc.calldata)

			decodedValues, err := DecodeFromTypes([]StarknetType{tc.starknetType}, &calldata)
			assert.NoError(t, err)

			encodedCalldata, err := EncodeFromTypes([]StarknetType{tc.starknetType}, []interface{}{tc.decoded})
			assert.NoError(t, err)

			assert.Equal(t, tc.calldata, encodedCalldata)
			assert.Equal(t, tc.decoded, decodedValues[0])
			assert.Empty(t, calldata)
		})
	}
}

func TestStructSerializerValidValues(t *testing.T) {
	testCases := []struct {
		name       string
		structType StarknetStruct
		calldata   []*big.Int
		decoded    map[string]interface{}
	}{
		{
			name: "CartesianPoint",
			structType: StarknetStruct{
				Name: "CartesianPoint",
				Members: []AbiParameter{
					{Name: "x", Type: U128},
					{Name: "y", Type: U128},
				},
			},
			calldata: []*big.Int{big.NewInt(1), big.NewInt(2)},
			decoded: map[string]interface{}{
				"x": big.NewInt(1),
				"y": big.NewInt(2),
			},
		},
		{
			name: "Queue",
			structType: StarknetStruct{
				Name: "Queue",
				Members: []AbiParameter{
					{Name: "head", Type: U8},
					{Name: "items", Type: StarknetArray{InnerType: U128}},
					{Name: "metadata", Type: StarknetStruct{
						Name: "MetaData",
						Members: []AbiParameter{
							{Name: "version", Type: U8},
							{Name: "init_timestamp", Type: U64},
						},
					}},
				},
			},
			calldata: []*big.Int{
				big.NewInt(22), big.NewInt(2), big.NewInt(38), big.NewInt(334),
				big.NewInt(5), big.NewInt(123456),
			},
			decoded: map[string]interface{}{
				"head":  big.NewInt(22),
				"items": []interface{}{big.NewInt(38), big.NewInt(334)},
				"metadata": map[string]interface{}{
					"version":        big.NewInt(5),
					"init_timestamp": big.NewInt(123456),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			calldata := make([]*big.Int, len(tc.calldata))
			copy(calldata, tc.calldata)

			decodedValues, err := DecodeFromTypes([]StarknetType{tc.structType}, &calldata)
			assert.NoError(t, err, "DecodeFromTypes should not return an error")
			assert.Equal(t, tc.decoded, decodedValues[0], "Decoded values should match expected")

			encodedCalldata, err := EncodeFromTypes([]StarknetType{tc.structType}, []interface{}{tc.decoded})
			assert.NoError(t, err, "EncodeFromTypes should not return an error")
			assert.Equal(t, tc.calldata, encodedCalldata, "Encoded calldata should match original")

			assert.Empty(t, calldata, "Calldata should be empty after decoding")
		})
	}
}

func TestTupleValidValues(t *testing.T) {
	tests := []struct {
		name         string
		starknetType StarknetType
		calldata     []*big.Int
		decoded      interface{}
	}{
		{
			name:         "Simple U32 Tuple",
			starknetType: StarknetTuple{Members: []StarknetType{U32, U32}},
			calldata:     []*big.Int{big.NewInt(1), big.NewInt(2)},
			decoded:      []interface{}{big.NewInt(1), big.NewInt(2)},
		},
		{
			name: "U32 and Array Tuple",
			starknetType: StarknetTuple{Members: []StarknetType{
				U32,
				StarknetArray{InnerType: U32},
			}},
			calldata: []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(22), big.NewInt(38)},
			decoded: []interface{}{
				big.NewInt(1),
				[]interface{}{big.NewInt(22), big.NewInt(38)},
			},
		},
		{
			name: "Three Nested Tuples",
			starknetType: StarknetTuple{Members: []StarknetType{
				StarknetTuple{Members: []StarknetType{
					StarknetTuple{Members: []StarknetType{U32}},
				}},
				Bool,
			}},
			calldata: []*big.Int{big.NewInt(1), big.NewInt(0)},
			decoded: []interface{}{
				[]interface{}{
					[]interface{}{big.NewInt(1)},
				},
				false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Copy calldata
			calldata := make([]*big.Int, len(tt.calldata))
			copy(calldata, tt.calldata)

			// Test decoding
			decodedValues, err := DecodeFromTypes([]StarknetType{tt.starknetType}, &calldata)
			assert.NoError(t, err, "DecodeFromTypes should not return an error")
			assert.Equal(t, tt.decoded, decodedValues[0], "Decoded values should match expected")

			// Test encoding
			encodedCalldata, err := EncodeFromTypes([]StarknetType{tt.starknetType}, []interface{}{tt.decoded})
			assert.NoError(t, err, "EncodeFromTypes should not return an error")
			assert.Equal(t, tt.calldata, encodedCalldata, "Encoded calldata should match original")

			// Check if calldata is empty after decoding
			assert.Empty(t, calldata, "Calldata should be empty after decoding")
		})
	}
}

func TestBytes31(t *testing.T) {
	tests := []struct {
		name     string
		calldata []*big.Int
		decoded  string
	}{
		{
			name: "Bytes31 Decoding and Encoding",
			calldata: func() []*big.Int {
				val, _ := new(big.Int).SetString("3DC782D803B8A574D29E3383A4885EBDDDA9D8D7E15CD5A5F1FB1651EE052E", 16)
				return []*big.Int{val}
			}(),
			decoded: "0x3dc782d803b8a574d29e3383a4885ebddda9d8d7e15cd5a5f1fb1651ee052e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Copy calldata
			calldata := make([]*big.Int, len(tt.calldata))
			copy(calldata, tt.calldata)

			// Test decoding
			decodedValues, err := DecodeFromTypes([]StarknetType{Bytes31}, &calldata)
			assert.NoError(t, err, "DecodeFromTypes should not return an error")

			assert.Equal(t, 1, len(decodedValues), "Should have one decoded value")
			decodedStr, ok := decodedValues[0].(string)
			assert.True(t, ok, "Decoded value should be a string")
			assert.Equal(t, 62, len(decodedStr[2:]), "Decoded string should have 62 characters (excluding '0x')")
			assert.Equal(t, tt.decoded, decodedStr, "Decoded value should match expected")

			// Test encoding
			encodedCalldata, err := EncodeFromTypes([]StarknetType{Bytes31}, []interface{}{tt.decoded})
			assert.NoError(t, err, "EncodeFromTypes should not return an error")
			assert.Equal(t, tt.calldata, encodedCalldata, "Encoded calldata should match original")

			// Check if calldata is empty after decoding
			assert.Empty(t, calldata, "Calldata should be empty after decoding")
		})
	}
}
