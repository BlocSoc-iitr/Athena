package types

import "fmt"

// BlockIdentifier is a custom type that can represent either an integer
// or one of the predefined string literals.
type BlockIdentifier interface{}

// Define constants for the string literals
const (
	Latest    = "latest"
	Earliest  = "earliest"
	Pending   = "pending"
	Safe      = "safe"
	Finalized = "finalized"
)

// NewBlockIdentifier returns a BlockIdentifier based on input.
// It accepts either an integer or a string matching the defined literals.
func NewBlockIdentifier(value interface{}) (BlockIdentifier, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case string:
		switch v {
		case Latest, Earliest, Pending, Safe, Finalized:
			return v, nil
		default:
			return nil, fmt.Errorf("invalid string literal for BlockIdentifier")
		}
	default:
		return nil, fmt.Errorf("unsupported type for BlockIdentifier")
	}
}
