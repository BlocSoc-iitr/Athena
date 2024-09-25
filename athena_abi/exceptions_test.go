package athena_abi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidAbiError(t *testing.T) {
	err := InvalidAbiError{Msg: "Invalid ABI format"}
	expectedMsg := "Invalid ABI Error: Invalid ABI format"

	// Assert that the error message is as expected
	assert.Equal(t, expectedMsg, err.Error(), "The error message for InvalidAbiError should match the expected message")
}

func TestInvalidCalldataError(t *testing.T) {
	err := InvalidCalldataError{Msg: "Not enough calldata to decode"}
	expectedMsg := "Invalid Calldata Error: Not enough calldata to decode"

	// Assert that the error message is as expected
	assert.Equal(t, expectedMsg, err.Error(), "The error message for InvalidCalldataError should match the expected message")
}

func TestTypeDecodeError(t *testing.T) {
	err := TypeDecodeError{Msg: "Failed to decode type"}
	expectedMsg := "Type Decode Error: Failed to decode type"

	// Assert that the error message is as expected
	assert.Equal(t, expectedMsg, err.Error(), "The error message for TypeDecodeError should match the expected message")
}

func TestTypeEncodeError(t *testing.T) {
	err := TypeEncodeError{Msg: "Failed to encode type"}
	expectedMsg := "Type Encode Error: Failed to encode type"

	// Assert that the error message is as expected
	assert.Equal(t, expectedMsg, err.Error(), "The error message for TypeEncodeError should match the expected message")
}

func TestDispatcherDecodeError(t *testing.T) {
	err := DispatcherDecodeError{Msg: "Failed to decode dispatcher"}
	expectedMsg := "Dispatcher Decode Error: Failed to decode dispatcher"

	// Assert that the error message is as expected
	assert.Equal(t, expectedMsg, err.Error(), "The error message for DispatcherDecodeError should match the expected message")
}
