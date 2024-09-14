package types

import (
	// "encoding/hex"
)

// TODO : Check if this file is useful. 

// Dataclass is a map from string key to any value
// type Dataclass map[string]interface{}

// ToBytes converts a string or byte array to a byte array with optional padding.
// func ToBytes(data interface{}, pad int) []byte {
// 	var b []byte
// 	switch v := data.(type) {
// 		case string:
// 			b, _ = hex.DecodeString(v)
// 		case []byte:
// 			b = v
// 		default:
// 			return nil
// 	}
// 	if pad > 0 && len(b) < pad {
// 		padded := make([]byte, pad)
// 		copy(padded[pad-len(b):], b)
// 		return padded
// 	}
// 	return b
// }

// DataclassToJson converts a Dataclass to JSON with custom byte-to-hex encoding.
// func DataclassToJson(obj Dataclass) (string, error) {
// 	jsonData, err := json.MarshalIndent(obj, "", "    ")
// 	if err != nil {
// 		return "", err
// 	}
// 	return string(jsonData), nil
// }

// JsonToDataclass converts JSON string back to a Dataclass.
// func JsonToDataclass(jsonStr string, cls *Dataclass) error {
// 	return json.Unmarshal([]byte(jsonStr), cls)
// }

// GetTransactionHashForDataclass returns the transaction hash for a dataclass as a byte array.
// func GetTransactionHashForDataclass(dataclass Dataclass) []byte {
// 	if v, ok := dataclass["transaction_hash"]; ok {
// 		return ToBytes(v, 32)
// 	}
// 	if v, ok := dataclass["tx_hash"]; ok {
// 		return ToBytes(v, 32)
// 	}
// 	if v, ok := dataclass["hash"]; ok {
// 		return ToBytes(v, 32)
// 	}
// 	return nil
// }

// GetBlockNumberForDataclass returns the block number for a dataclass as an integer.
// func GetBlockNumberForDataclass(dataclass Dataclass) *int {
// 	if v, ok := dataclass["block_number"]; ok {
// 		if number, ok := v.(float64); ok { // JSON numbers are float64 in Go
// 			num := int(number)
// 			return &num
// 		}
// 	}
// 	if v, ok := dataclass["block"]; ok {
// 		if number, ok := v.(float64); ok {
// 			num := int(number)
// 			return &num
// 		}
// 	}
// 	if v, ok := dataclass["number"]; ok {
// 		if number, ok := v.(float64); ok {
// 			num := int(number)
// 			return &num
// 		}
// 	}
// 	return nil
// }
