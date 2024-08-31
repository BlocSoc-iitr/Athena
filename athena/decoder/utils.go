package decoder

import (
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

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
