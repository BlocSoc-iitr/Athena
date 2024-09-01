package types

import (
	"encoding/hex"
	"encoding/json"
)

// Dataclass is a map that holds key-value pairs for any type.
// In Go, it's usually represented as a map[string]interface{}.
type Dataclass map[string]interface{}

// HexEnabledJsonEncoder is a JSON encoder that converts bytes to hex strings.
func HexEnabledJsonEncoder(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "    ")
}

// JSONMarshal with custom encoding for bytes to hex strings.
func JSONMarshal(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "    ")
}

// ToBytes converts a string or byte array to a byte array with optional padding.
func ToBytes(data interface{}, pad int) []byte {
	var b []byte
	switch v := data.(type) {
	case string:
		b, _ = hex.DecodeString(v)
	case []byte:
		b = v
	default:
		return nil
	}
	if pad > 0 && len(b) < pad {
		padded := make([]byte, pad)
		copy(padded[pad-len(b):], b)
		return padded
	}
	return b
}

// DataclassToJson converts a Dataclass to JSON with custom byte-to-hex encoding.
func DataclassToJson(obj Dataclass) (string, error) {
	jsonData, err := JSONMarshal(obj)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// JsonToDataclass converts JSON string back to a Dataclass.
func JsonToDataclass(jsonStr string, cls *Dataclass) error {
	return json.Unmarshal([]byte(jsonStr), cls)
}

// GetTransactionHashForDataclass returns the transaction hash for a dataclass as a byte array.
func GetTransactionHashForDataclass(dataclass Dataclass) []byte {
	if v, ok := dataclass["transaction_hash"]; ok {
		return ToBytes(v, 32)
	}
	if v, ok := dataclass["tx_hash"]; ok {
		return ToBytes(v, 32)
	}
	if v, ok := dataclass["hash"]; ok {
		return ToBytes(v, 32)
	}
	return nil
}

// GetBlockNumberForDataclass returns the block number for a dataclass as an integer.
func GetBlockNumberForDataclass(dataclass Dataclass) *int {
	if v, ok := dataclass["block_number"]; ok {
		if number, ok := v.(float64); ok { // JSON numbers are float64 in Go
			num := int(number)
			return &num
		}
	}
	if v, ok := dataclass["block"]; ok {
		if number, ok := v.(float64); ok {
			num := int(number)
			return &num
		}
	}
	if v, ok := dataclass["number"]; ok {
		if number, ok := v.(float64); ok {
			num := int(number)
			return &num
		}
	}
	return nil
}
