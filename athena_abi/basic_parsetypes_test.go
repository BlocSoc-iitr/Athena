package athena_abi

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test parsing integer types
func TestParseIntTypes(t *testing.T) {
	tests := []struct {
		input        string
		expectedType StarknetType
	}{
		{"core::integer::u256", U256},
		{"core::integer::u128", U128},
		{"core::integer::u64", U64},
		{"core::integer::u32", U32},
		{"core::integer::u16", U16},
		{"core::integer::u8", U8},
		{"Uint256", U256},
	}

	for _, tt := range tests {
		res, err := parseType(tt.input, nil)
		assert.NoError(t, err, "parseType(%v) failed with error", tt.input)
		assert.Equal(t, tt.expectedType, res, "Expected %v, got %v", tt.expectedType, res)
	}
}

// Test parsing address types
func TestParseAddressTypes(t *testing.T) {
	tests := []struct {
		input        string
		expectedType StarknetType
	}{
		{"core::starknet::contract_address::ContractAddress", ContractAddress},
		{"core::starknet::class_hash::ClassHash", ClassHash},
		{"core::starknet::eth_address::EthAddress", EthAddress},
	}

	for _, tt := range tests {
		res, err := parseType(tt.input, nil)
		assert.NoError(t, err, "parseType(%v) failed with error", tt.input)
		assert.Equal(t, tt.expectedType, res, "Expected %v, got %v", tt.expectedType, res)
	}
}

// Test parsing felt types
func TestParseFelts(t *testing.T) {
	res, err := parseType("core::felt252", nil)
	assert.NoError(t, err, "parseType(core::felt252) failed with error")
	assert.Equal(t, Felt, res, "Expected %v, got %v", Felt, res)
}

// Test parsing entry point felt types
func TestParseEntryPointFelts(t *testing.T) {
	tests := []struct {
		input        string
		expectedType StarknetType
	}{
		{"felt", Felt},
		{"felt*", StarknetArray{Felt}},
	}

	for _, tt := range tests {
		res, err := parseType(tt.input, nil)
		assert.NoError(t, err, "parseType(%v) failed with error", tt.input)
		assert.Equal(t, tt.expectedType, res, "Expected %v, got %v", tt.expectedType, res)
	}
}

// Test parsing boolean types
func TestParseBool(t *testing.T) {
	res, err := parseType("core::bool", nil)
	assert.NoError(t, err, "parseType(core::bool) failed with error")
	assert.Equal(t, Bool, res, "Expected %v, got %v", Bool, res)
}

// Test parsing array types
func TestParseArray(t *testing.T) {
	tests := []struct {
		input        string
		expectedType StarknetType
	}{
		{"core::array::Array::<core::integer::u256>", StarknetArray{U256}},
		{"core::array::Array::<core::bool>", StarknetArray{Bool}},
	}

	for _, tt := range tests {
		res, err := parseType(tt.input, nil)
		assert.NoError(t, err, "parseType(%v) failed with error", tt.input)
		assert.Equal(t, tt.expectedType, res, "Expected %v, got %v", tt.expectedType, res)
	}
}

// Test parsing option types
func TestParseOption(t *testing.T) {
	tests := []struct {
		input        string
		expectedType StarknetType
	}{
		{"core::option::Option::<core::integer::u256>", StarknetOption{U256}},
		{"core::option::Option::<core::bool>", StarknetOption{Bool}},
	}

	for _, tt := range tests {
		res, err := parseType(tt.input, nil)
		assert.NoError(t, err, "parseType(%v) failed with error", tt.input)
		assert.Equal(t, tt.expectedType, res, "Expected %v, got %v", tt.expectedType, res)
	}
}

// Test parsing legacy types
func TestLegacyTypes(t *testing.T) {
	res, err := parseType("(x: felt, y: felt)", nil)
	assert.NoError(t, err, "parseType((x: felt, y: felt)) failed with error")
	expected := StarknetTuple{[]StarknetType{Felt, Felt}}
	assert.Equal(t, expected, res, "Expected %v, got %v", expected, res)
}

// Test parsing storage address types
func TestParseStorageAddress(t *testing.T) {
	res, err := parseType("core::starknet::storage_access::StorageAddress", nil)
	assert.NoError(t, err, "parseType(core::starknet::storage_access::StorageAddress) failed with error")
	assert.Equal(t, StorageAddress, res, "Expected %v, got %v", StorageAddress, res)
}

// Test parsing bytes types
func TestParseBytes(t *testing.T) {
	res, err := parseType("core::bytes_31::bytes31", nil)
	assert.NoError(t, err, "parseType(core::bytes_31::bytes31) failed with error")
	assert.Equal(t, Bytes31, res, "Expected %v, got %v", Bytes31, res)
}
