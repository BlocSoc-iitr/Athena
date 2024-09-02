package decoder

import (
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func extractFunctionSignature(txData []byte) []byte {
	if len(txData) < 4 {
		return nil // Not enough data to extract signature
	}
	return txData[:4] // Assuming the first 4 bytes are the function signature
}
func splitTxData(txData []byte) (calldata [][]byte, result [][]byte) {
	if len(txData) <= 4 {
		return [][]byte{}, [][]byte{} // No calldata if data length is less than or equal to 4
	}
	
	// Extract the function signature
	signature := txData[:4]
	
	// The remaining bytes are considered calldata
	calldata = [][]byte{signature, txData[4:]} // Splitting into signature and the rest as calldata

	// In this example, we don't have a separate result part, so result is empty
	return calldata, [][]byte{}
}
func extractEventSignature(eventData []byte) []byte {
	if len(eventData) < 4 {
		return nil // Not enough data to extract signature
	}
	return eventData[:4] // Assuming the first 4 bytes are the event signature
}
func extractIndexedParams(eventData []byte) int {
	// This is a placeholder. You need to implement this based on your data format
	// Typically, this could be extracted from event metadata or ABI information.
	return 0 // Default value; replace with actual logic as needed
}
func splitEventData(eventData []byte) (data [][]byte, keys [][]byte) {
	if len(eventData) <= 4 {
		return [][]byte{}, [][]byte{} // Not enough data to split
	}
	
	// Extract the event signature
	signature := eventData[:4]
	
	// The rest of the data
	remainingData := eventData[4:]
	
	// Assuming you need to split remainingData into data and keys
	// For simplicity, this example treats all remaining data as a single part.
	data = [][]byte{signature, remainingData}
	keys = [][]byte{} // You would split into keys based on your specific data format

	return data, keys
}
func AbiToSignature(abiFunc abi.Method) string {
	collapsed := make([]string, len(abiFunc.Inputs))
	for i, input := range abiFunc.Inputs {
		collapsed[i] = CollapseIfTuple(input)
	}
	return fmt.Sprintf("%s(%s)", abiFunc.Name, strings.Join(collapsed, ","))
}

func CollapseIfTuple(input abi.Argument) string {
	if !strings.HasPrefix(input.Type.String(), "tuple") {
		return input.Type.String()
	}

	components := input.Type.TupleElems
	delimited := make([]string, len(components))
	for i, component := range components {
		delimited[i] = component.String()
	}

	return fmt.Sprintf("(%s)%s", strings.Join(delimited, ","), input.Type.String()[5:])
}

func AbiSignatureToName(signature string) string {
	index := strings.Index(signature, "(")
	if index != -1 {
		return signature[:index]
	}
	return signature
}

func FilterFunctions(contractABI abi.ABI) []abi.Method {
	var methods []abi.Method
	for _, method := range contractABI.Methods {
		methods = append(methods, method)
	}
	return methods
}

func FilterEvents(contractABI abi.ABI) []abi.Event {
	var events []abi.Event
	for _, event := range contractABI.Events {
		events = append(events, event)
	}
	return events
}

func DecodeEvmAbiFromTypes(types []string, data []byte) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Unknown error while decoding %s for types %v: %v", hex.EncodeToString(data), types, r)
		}
	}()

	arguments := make(abi.Arguments, len(types))
	for i, typ := range types {
		argType, err := abi.NewType(typ, "", nil)
		if err != nil {
			log.Printf("Error creating new type for %s: %v", typ, err)
			return nil, err
		}
		arguments[i] = abi.Argument{Type: argType}
	}

	result, err := arguments.Unpack(data)
	if err != nil {
		log.Printf("Error while decoding %s for types %v: %v", hex.EncodeToString(data), types, err)
		return nil, err
	}
	return result, nil
}
