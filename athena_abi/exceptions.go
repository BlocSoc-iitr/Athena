package athenaabi

import (
	"fmt"
)

// InvalidAbiError is raised when malformed ABI JSON is supplied to the parser.
type InvalidAbiError struct {
	Msg string
}

func (e *InvalidAbiError) Error() string {
	return fmt.Sprintf("Invalid ABI Error: %s", e.Msg)
}

// InvalidCalldataError is raised when there is not enough calldata to decode the type.
type InvalidCalldataError struct {
	Msg string
}

func (e *InvalidCalldataError) Error() string {
	return fmt.Sprintf("Invalid Calldata Error: %s", e.Msg)
}

// TypeDecodeError is raised when a type cannot be decoded from the calldata.
type TypeDecodeError struct {
	Msg string
}

func (e *TypeDecodeError) Error() string {
	return fmt.Sprintf("Type Decode Error: %s", e.Msg)
}

// TypeEncodeError is raised when a type cannot be encoded from the calldata.
type TypeEncodeError struct {
	Msg string
}

func (e *TypeEncodeError) Error() string {
	return fmt.Sprintf("Type Encode Error: %s", e.Msg)
}

// DispatcherDecodeError is raised when there is an error decoding Functions, Events, or User Operations using the decoding dispatcher.
type DispatcherDecodeError struct {
	Msg string
}

func (e *DispatcherDecodeError) Error() string {
	return fmt.Sprintf("Dispatcher Decode Error: %s", e.Msg)
}
