package athena_abi

import (
	"fmt"
	"math/big"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeInvalidEnumsraises(t *testing.T) {

	testEnum := StarknetEnum{
		Name: "TestEnum",
		Variants: []struct {
			Name string
			Type StarknetType
		}{
			{"a", StarknetCoreType(U256)},
			{"b", StarknetCoreType(U64)},
			{"c", StarknetCoreType(Bool)},
		},
	}

	errorPattern := regexp.MustCompile("enum value .* must have exactly one key-value pair")
	// Test case 1: More than one key-value pair
	t.Run("MoreThanOneKeyValuePair", func(t *testing.T) {
		_, err := EncodeFromTypes(
			[]StarknetType{testEnum},
			[]interface{}{
				map[string]interface{}{"a": 100, "b": 200},
			},
		)

		assert.Error(t, err, "Expected an error for multiple key-value pairs")

		if err != nil {
			assert.Regexp(t, errorPattern, err.Error(), "Error should match the expected pattern for multiple key-value pairs")
		}
	})

	// Test case 2: Zero key-value pairs
	t.Run("ZeroKeyValuePair", func(t *testing.T) {
		_, err := EncodeFromTypes(
			[]StarknetType{testEnum},
			[]interface{}{
				map[string]interface{}{},
			},
		)

		assert.Error(t, err, "Expected an error for zero key-value pairs")
		if err != nil {
			assert.Regexp(t, errorPattern, err.Error(), "Error should match the expected pattern for zero key-value pairs")
		}
	})
}

func TestEncodeInvalidFelts(t *testing.T) {
	// Test 1: Exceeding max Felt value
	errorPattern := regexp.MustCompile(`\d+ does not fit into (Felt|ContractAddress|EthAddress)`)

	maxFeltValue, _ := StarknetCoreType(Felt).maxValue()
	tooLargeFelt := new(big.Int).Add(maxFeltValue, big.NewInt(1))

	_, err := EncodeFromTypes(
		[]StarknetType{StarknetCoreType(Felt)},
		[]interface{}{tooLargeFelt},
	)
	assert.Error(t, err)
	assert.Regexp(t, errorPattern, err.Error(), "Error should match the expected error pattern")

	// Test 2: Exceeding max ContractAddress value
	maxContractAddrValue, _ := StarknetCoreType(ContractAddress).maxValue()
	tooLargeContractAddr := new(big.Int).Add(maxContractAddrValue, big.NewInt(1))

	_, err = EncodeFromTypes(
		[]StarknetType{StarknetCoreType(ContractAddress)},
		[]interface{}{tooLargeContractAddr},
	)
	assert.Error(t, err)
	assert.Regexp(t, errorPattern, err.Error(), "Error should match the expected error pattern")

	// Test 3: Exceeding max EthAddress value
	tooLargeEthAddress := new(big.Int)
	tooLargeEthAddress.SetString("123456789012345678901234567890123456789012", 16)

	_, err = EncodeFromTypes(
		[]StarknetType{StarknetCoreType(EthAddress)},
		[]interface{}{tooLargeEthAddress},
	)
	assert.Error(t, err)
	assert.Regexp(t, errorPattern, err.Error(), "Error should match the expected error pattern")
}

func TestEncodeInvalidIntValue(t *testing.T) {
	// Test 1: Value exceeding max U256
	tooLargeU256 := new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)

	_, err := EncodeFromTypes(
		[]StarknetType{StarknetCoreType(U256)},
		[]interface{}{tooLargeU256},
	)
	assert.Error(t, err)
	regexU256 := regexp.MustCompile(`value \d+ is out of range for U256`)
	assert.Regexp(t, regexU256, err.Error())

	// Test 2: Negative value for U128
	negativeValue := big.NewInt(-1)

	_, err = EncodeFromTypes(
		[]StarknetType{StarknetCoreType(U128)},
		[]interface{}{negativeValue},
	)
	assert.Error(t, err)
	regexU128 := regexp.MustCompile(`value -?\d+ is out of range for U128`)
	assert.Regexp(t, regexU128, err.Error())

	// Test 3: Value exceeding max U64
	tooLargeU64 := new(big.Int).Exp(big.NewInt(2), big.NewInt(64), nil)

	_, err = EncodeFromTypes(
		[]StarknetType{StarknetCoreType(U64)},
		[]interface{}{tooLargeU64},
	)
	assert.Error(t, err)
	regexU64 := regexp.MustCompile(`value \d+ is out of range for U64`)
	assert.Regexp(t, regexU64, err.Error())

}

func TestEncodeInvalidDictValues(t *testing.T) {

	rawUintStruct := StarknetStruct{
		Name: "CustomUint256",
		Members: []AbiParameter{
			{Name: "low", Type: StarknetCoreType(U128)},
			{Name: "high", Type: StarknetCoreType(U128)},
		},
	}

	// Test 1: Negative value for "low"
	negativeLow := map[string]interface{}{
		"low":  big.NewInt(-1),
		"high": big.NewInt(12324),
	}

	_, err := EncodeFromTypes([]StarknetType{rawUintStruct}, []interface{}{negativeLow})
	assert.Error(t, err)
	errorPattern := regexp.MustCompile(`Failed to Encode (.*?) to \w+`)
	assert.Regexp(t, errorPattern, err.Error())

	// Test 2: "low" exceeds MAX_U128
	MAX_U128, _ := StarknetCoreType(U128).maxValue()
	tooLargeLow := map[string]interface{}{
		"low":  new(big.Int).Add(MAX_U128, big.NewInt(1)),
		"high": big.NewInt(4543535),
	}

	_, err = EncodeFromTypes([]StarknetType{rawUintStruct}, []interface{}{tooLargeLow})
	assert.Error(t, err)
	assert.Regexp(t, errorPattern, err.Error())

	// Test 3: Negative value for "high"
	negativeHigh := map[string]interface{}{
		"low":  big.NewInt(652432),
		"high": big.NewInt(-1),
	}

	_, err = EncodeFromTypes([]StarknetType{rawUintStruct}, []interface{}{negativeHigh})
	assert.Error(t, err)
	assert.Regexp(t, errorPattern, err.Error())

	// Test 4: "high" exceeds MAX_U128
	tooLargeHigh := map[string]interface{}{
		"low":  big.NewInt(0),
		"high": new(big.Int).Add(MAX_U128, big.NewInt(1)),
	}

	_, err = EncodeFromTypes([]StarknetType{rawUintStruct}, []interface{}{tooLargeHigh})
	assert.Error(t, err)
	assert.Regexp(t, errorPattern, err.Error())
}

func TestEncodeInvalidType(t *testing.T) {
	tests := []struct {
		encodeType   StarknetType
		encodeValues []interface{}
		errorMessage string
	}{
		{
			encodeType: StarknetCoreType(U64),
			encodeValues: []interface{}{
				"wololoo",
				nil,
				map[string]interface{}{"low": 12},
				"0xaabbccddff001122334455",
			},
			errorMessage: `cannot encode value of type .* to U64`,
		},
		{
			encodeType: StarknetCoreType(Bool),
			encodeValues: []interface{}{
				nil,
				[]interface{}{nil, true, false},
				123,
				map[string]interface{}{"low": 1234, "high": 0},
			},
			errorMessage: `cannot encode non-boolean value .* to Bool`,
		},
		{
			encodeType: StarknetArray{InnerType: StarknetCoreType(U64)},
			encodeValues: []interface{}{
				nil,
				"0x12334455",
				map[string]interface{}{"low": 1234, "high": 0},
				false,
			},
			errorMessage: `.* cannot be encoded into a StarknetArray`,
		},
	}

	for _, tt := range tests {
		for _, encodeValue := range tt.encodeValues {
			t.Run(fmt.Sprintf("%v-%v", tt.encodeType, encodeValue), func(t *testing.T) {
				_, err := EncodeFromTypes([]StarknetType{tt.encodeType}, []interface{}{encodeValue})
				if err == nil {
					t.Errorf("expected error matching %q but got none", tt.errorMessage)
					return
				}

				if !errorMatches(err.Error(), tt.errorMessage) {
					t.Errorf("expected error matching %q but got %q", tt.errorMessage, err.Error())
				}
			})
		}
	}
}

func TestDecodeExtraCalldata(t *testing.T) {
	types := []StarknetType{StarknetCoreType(U8), StarknetCoreType(U256)}
	callData := []*big.Int{big.NewInt(123), big.NewInt(0)}
	errorPattern := "not enough calldata to decode U256"
	t.Run(fmt.Sprintf("%v", types), func(t *testing.T) {
		_, err := DecodeFromTypes(types, &callData)
		if err == nil {
			t.Errorf("expected error matching %q but got none", errorPattern)
			return
		}

		if !errorMatches(err.Error(), errorPattern) {
			t.Errorf("expected error matching %q but got %q", errorPattern, err.Error())
		}
	})
}

func TestDecodeInvalidUIntValues(t *testing.T) {

	MAX_U128, _ := StarknetCoreType(U128).maxValue()
	tests := []struct {
		types         []StarknetType
		callData      []*big.Int
		expectedError string
	}{
		{
			types:         []StarknetType{StarknetCoreType(U256)},
			callData:      []*big.Int{new(big.Int).Add(MAX_U128, big.NewInt(1)), big.NewInt(0)},
			expectedError: "low Exceeds U128 range",
		},
		{
			types:         []StarknetType{StarknetCoreType(U256)},
			callData:      []*big.Int{new(big.Int).Add(MAX_U128, big.NewInt(1)), new(big.Int).Add(MAX_U128, big.NewInt(1))},
			expectedError: "low Exceeds U128 range",
		},
		{
			types:         []StarknetType{StarknetCoreType(U256)},
			callData:      []*big.Int{new(big.Int).Sub(big.NewInt(0), big.NewInt(1)), big.NewInt(0)},
			expectedError: "low Exceeds U128 range",
		},
		{
			types:         []StarknetType{StarknetCoreType(U256)},
			callData:      []*big.Int{big.NewInt(0), new(big.Int).Add(MAX_U128, big.NewInt(1))},
			expectedError: "high Exceeds U128 range",
		},
		{
			types:         []StarknetType{StarknetCoreType(U256)},
			callData:      []*big.Int{big.NewInt(0), new(big.Int).Sub(big.NewInt(0), big.NewInt(1))},
			expectedError: "high Exceeds U128 range",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.callData), func(t *testing.T) {
			_, err := DecodeFromTypes(tt.types, &tt.callData)
			if err == nil {
				t.Errorf("expected error matching %q but got none", tt.expectedError)
				return
			}

			if !errorMatches(err.Error(), tt.expectedError) {
				t.Errorf("expected error matching %q but got %q", tt.expectedError, err.Error())
			}
		})
	}
}

func errorMatches(errorMessage, pattern string) bool {

	matched, err := regexp.MatchString(pattern, errorMessage)
	if err != nil {
		return false
	}
	return matched
}
