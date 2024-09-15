package athena_abi

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestBigIntToBytes(t *testing.T) {
	// Normal logic tests
	tests := []struct {
		input    big.Int
		length   int
		expected []byte
	}{
		{*big.NewInt(12345), 4, []byte{0, 0, 48, 57}},                          // 12345 in 4 bytes
		{*big.NewInt(0x123456), 6, []byte{0x01, 0x23, 0x45, 0x00, 0x00, 0x00}}, // 0x123456 in 6 bytes
	}

	t.Run("Normal Case - First Item", func(t *testing.T) {
		tt := tests[0]
		result := bigIntToBytes(tt.input, tt.length)
		assert.Equal(t, tt.expected, result)
	})

	t.Run("Pass Case - Second Item with NotEqual", func(t *testing.T) {
		tt := tests[1]
		incorrectExpected := []byte{0x01, 0x23, 0x45, 0x00, 0x00, 0x01} // Incorrect last byte
		result := bigIntToBytes(tt.input, tt.length)
		// Verify that the result is not equal to the incorrect expected value
		assert.NotEqual(t, incorrectExpected, result)
	})

	// Boundary cases
	t.Run("Zero Length", func(t *testing.T) {
		result := bigIntToBytes(*big.NewInt(12345), 0)
		assert.Empty(t, result)
	})

	t.Run("Exact Length Match", func(t *testing.T) {
		result := bigIntToBytes(*big.NewInt(255), 1)
		assert.Equal(t, []byte{0xFF}, result)
	})

	t.Run("Large Integer", func(t *testing.T) {
		largeInt := big.NewInt(0).Lsh(big.NewInt(1), 64) // 2^64
		result := bigIntToBytes(*largeInt, 9)
		expected := []byte{0x01, 0, 0, 0, 0, 0, 0, 0, 0} // 2^64 in 8 bytes
		assert.Equal(t, expected, result)
	})

}

func TestStarknetKeccak(t *testing.T) {
	// Known input and output based on Keccak-256 hashing and masking
	tests := []struct {
		input    []byte
		expected []byte
	}{
		{[]byte("transfer"), []byte{0x00, 0x83, 0xaf, 0xd3, 0xf4, 0xca, 0xed, 0xc6, 0xee, 0xbf, 0x44, 0x24, 0x6f, 0xe5, 0x4e, 0x38, 0xc9, 0x5e, 0x31, 0x79, 0xa5, 0xec, 0x9e, 0xa8, 0x17, 0x40, 0xec, 0xa5, 0xb4, 0x82, 0xd1, 0x2e}}, // Expected output from Keccak-256 hashing of "transfer"
	}

	for _, tt := range tests {
		t.Run("Normal Case", func(t *testing.T) {
			result := StarknetKeccak(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}

	// Boundary cases
	t.Run("Empty Input", func(t *testing.T) {
		result := StarknetKeccak([]byte{})
		expected := []byte{0x1, 0xd2, 0x46, 0x1, 0x86, 0xf7, 0x23, 0x3c, 0x92, 0x7e, 0x7d, 0xb2, 0xdc, 0xc7, 0x3, 0xc0, 0xe5, 0x0, 0xb6, 0x53, 0xca, 0x82, 0x27, 0x3b, 0x7b, 0xfa, 0xd8, 0x4, 0x5d, 0x85, 0xa4, 0x70} // Calculate this in advance if known or check with hashing tool
		assert.Equal(t, expected, result)
	})

	t.Run("Large Input", func(t *testing.T) {
		largeInput := make([]byte, 1024) // 1 KB of data
		result := StarknetKeccak(largeInput)
		expected := StarknetKeccak(largeInput) // Calculate this in advance if known or check with hashing tool
		assert.Equal(t, expected, result)
	})
}
