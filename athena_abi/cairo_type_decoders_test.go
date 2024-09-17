package athena_abi

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayDecoding(t *testing.T) {
	tests := []struct {
		starknetType StarknetType
		calldata     []*big.Int
		decoded      []*big.Int
	}{
		{StarknetArray{U256}, []*big.Int{big.NewInt(0)}, []*big.Int{}},
		{StarknetArray{U256}, []*big.Int{big.NewInt(2), big.NewInt(16), big.NewInt(0), big.NewInt(48), big.NewInt(0)}, []*big.Int{big.NewInt(16), big.NewInt(48)}},
	}

	for _, test := range tests {
		_calldata := make([]*big.Int, len(test.calldata))
		copy(_calldata, test.calldata)
		decodedValues, err := DecodeFromTypes([]StarknetType{test.starknetType}, &_calldata)
		assert.Equal(t, nil, err)
		for i := 0; i < len(test.decoded); i++ {
			assert.Equal(t, test.decoded[i], decodedValues[i])
		}
		// needed to convert []*big.Int into []interface{}
		interfaceSlice := make([]interface{}, len(test.decoded))
		for i, v := range test.decoded {
			interfaceSlice[i] = v
		}
		encodedCalldata, err := EncodeFromTypes([]StarknetType{test.starknetType}, []interface{}{interfaceSlice})
		assert.Equal(t, nil, err)
		assert.Equal(t, test.calldata, encodedCalldata)
		assert.Equal(t, 0, len(_calldata))
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
