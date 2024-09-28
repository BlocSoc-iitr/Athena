package athena_abi

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodingDispatcher(t *testing.T) {
	// Create a new decoding dispatcher
	dispatcher := NewDecodingDispatcher()

	abis := []struct {
		abiName    string
		abiVersion int
		classHash  string
	}{
		{"contract_abi", 1, "010455c752b86932ce552f2b0fe81a880746649b9aee7e0d842bf3f52378f9f8"},
		{"starknet_eth", 2, "05ffbcfeb50d200a0677c48a129a11245a3fc519d1d98d76882d1c9a1b19c6ed"},
	}

	for _, abi := range abis {
		abiJSON, _ := loadAbi(abi.abiName, abi.abiVersion)
		hex := parseHex(abi.classHash)

		parsedAbi, _ := StarknetAbiFromJSON(abiJSON, abi.abiName, hex[:])
		dispatcher.AddAbi(*parsedAbi)
	}

	// Verify that the classes have been added correctly
	for _, abi := range abis {
		classHashBytes := parseHex(abi.classHash)
		classDispatcher, exists := dispatcher.GetClass(classHashBytes)

		assert.True(t, exists, "Class should exist in dispatcher")
		assert.Equal(t, abi.abiName, *classDispatcher.AbiName, "ABI names should match")

		// Check that the number of class IDs is correct
		assert.Len(t, dispatcher.ClassIDs, 2, "Expected 2 class IDs in the dispatcher")
	}
}

func parseHex(hexStr string) [32]byte {
	bytes, _ := hex.DecodeString(hexStr)
	var arr [32]byte
	copy(arr[:], bytes)
	return arr
}
