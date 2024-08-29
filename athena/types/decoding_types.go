package types

import (
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/sha3"
)

type DecodedFunction struct {
	AbiName           string
	Name              string
	FunctionSignature string
	Input             map[string]interface{}
	Output            []interface{}
}

func (df DecodedFunction) String() string {
	return fmt.Sprintf(
		"DecodedFunction(abi_name=%s, name=%s, function_signature=%s, input=%v, output=%v)",
		df.AbiName,
		df.Name,
		df.FunctionSignature,
		df.Input,
		df.Output,
	)
}

type DecodedEvent struct {
	AbiName        string
	Name           string
	EventSignature string
	Data           map[string]interface{}
}

func (de DecodedEvent) String() string {
	return fmt.Sprintf(
		"DecodedEvent(abi_name=%s, name=%s, event_signature=%s, data=%v)",
		de.AbiName,
		de.Name,
		de.EventSignature,
		de.Data,
	)
}

type DecodedTrace struct {
	AbiName       string
	Name          string
	Signature     string
	DecodedInput  map[string]interface{}
	DecodedOutput map[string]interface{}
}

// String returns a string representation of the DecodedTrace.
func (dt DecodedTrace) String() string {
	return fmt.Sprintf(
		"DecodedTrace(abi_name=%s, name=%s, signature=%s, decoded_input=%v, decoded_output=%v)",
		dt.AbiName,
		dt.Name,
		dt.Signature,
		dt.DecodedInput,
		dt.DecodedOutput,
	)
}

type AbiFunction struct {
	Name      string
	AbiName   string
	Signature []byte
	Inputs    []AbiParameter
	Outputs   []StarknetType
}

func NewAbiFunction(name string, inputs []AbiParameter, outputs []StarknetType, abiName string) *AbiFunction {
	return &AbiFunction{
		Name:      name,
		AbiName:   abiName,
		Inputs:    inputs,
		Outputs:   outputs,
		Signature: calculateSignature(name),
	}
}

func calculateSignature(name string) []byte {
	hash := sha256.New()
	hash.Write([]byte(name))
	return hash.Sum(nil)
}

func (af *AbiFunction) IdStr() string {
	var inputsStr, outputsStr string
	for i, param := range af.Inputs {
		if i > 0 {
			inputsStr += ","
		}
		inputsStr += param.IDStr()
	}
	for i, output := range af.Outputs {
		if i > 0 {
			outputsStr += ","
		}
		outputsStr += output.IDStr()
	}
	return fmt.Sprintf("Function(%s) -> (%s)", inputsStr, outputsStr)
}

// Decode decodes the calldata and result into a DecodedFunction.
func (af *AbiFunction) Decode(calldata []int, result []int) DecodedFunction {
	calldataCopy := make([]int, len(calldata))
	copy(calldataCopy, calldata)
	decodedInputs := decodeFromParams(af.Inputs, calldataCopy)

	var decodedOutputs []interface{}
	if result != nil {
		resultCopy := make([]int, len(result))
		copy(resultCopy, result)
		decodedOutputs = decodeFromTypes(af.Outputs, resultCopy)
	}

	return DecodedFunction{
		AbiName:           af.AbiName,
		Name:              af.Name,
		FunctionSignature: fmt.Sprintf("%x", af.Signature),
		Input:             decodedInputs,
		Output:            decodedOutputs,
	}

}

// Encode encodes the inputs of the function into calldata.
func (af *AbiFunction) Encode(inputs map[string]interface{}) []int {
	return encodeFromParams(af.Inputs, inputs)
}

func encodeFromParams(inputs []AbiParameter, params map[string]interface{}) []int {

	return []int{}
}

// Helper function to decode from parameters (dummy implementation)
func decodeFromParams(inputs []AbiParameter, calldata []int) map[string]interface{} {

	return map[string]interface{}{}
}

// Helper function to decode from types (dummy implementation)
func decodeFromTypes(types []StarknetType, result []int) []interface{} {

	return []interface{}{}
}

type AbiEvent struct {
	Name       string
	AbiName    *string
	Signature  int
	Parameters []string
	Keys       map[string]StarknetType
	Data       map[string]StarknetType
}

// NewAbiEvent is a constructor function for creating a new AbiEvent instance.
func NewAbiEvent(
	name string,
	parameters []string,
	data map[string]StarknetType,
	keys map[string]StarknetType, //starknetKeccak([]byte(name))
	abiName *string,
) *AbiEvent {
	signature, _ := sha3.NewLegacyKeccak256().Write([]byte(name))
	if keys == nil {
		keys = make(map[string]StarknetType)
	}

	return &AbiEvent{
		Name:       name,
		AbiName:    abiName,
		Signature:  signature,
		Parameters: parameters,
		Keys:       keys,
		Data:       data,
	}
}

// // IdStr returns a string representation of the ABI Event.
// func (ae *AbiEvent) IdStr() (string, error) {
// 	eventParams := []string{}

// 	for _, param := range ae.Parameters {
// 		if typ, ok := ae.Data[param]; ok {
// 			eventParams = append(eventParams, fmt.Sprintf("%s:%s", param, typ.IdStr()))
// 		} else if typ, ok := ae.Keys[param]; ok {
// 			eventParams = append(eventParams, fmt.Sprintf("<%s>:%s", param, typ.IdStr()))
// 		} else {
// 			return "", fmt.Errorf("TypeDecodeError: Event Parameter %s not part of event keys or Data", param)
// 		}
// 	}

// 	return fmt.Sprintf("Event(%s)", strings.Join(eventParams, ",")), nil
// }

// Decode decodes the keys and data of an event.
// func (ae *AbiEvent) Decode(data, keys []int) (*DecodedEvent, error) {
// 	_data := make([]int, len(data))
// 	copy(_data, data)
// 	_keys := make([]int, len(keys[1:]))
// 	copy(_keys, keys[1:])

// 	decodedData := map[string]interface{}{}

// 	for _, param := range ae.Parameters {
// 		if typ, ok := ae.Data[param]; ok {
// 			decodedValue, err := DecodeFromTypes([]StarknetType{typ}, _data)
// 			if err != nil {
// 				return nil, err
// 			}
// 			decodedData[param] = decodedValue[0]
// 		} else if typ, ok := ae.Keys[param]; ok {
// 			decodedValue, err := DecodeFromTypes([]StarknetType{typ}, _keys)
// 			if err != nil {
// 				return nil, err
// 			}
// 			decodedData[param] = decodedValue[0]
// 		} else {
// 			return nil, fmt.Errorf("TypeDecodeError: Event Parameter %s not present in Keys or Data for Event %s", param, ae.Name)
// 		}
// 	}

// 	if len(_data) != 0 || len(_keys) != 0 {
// 		return nil, fmt.Errorf("InvalidCalldataError: Calldata Not Completely Consumed decoding Event: %s", ae.IdStr())
// 	}

// 	return &DecodedEvent{AbiName: ae.AbiName, Name: ae.Name, Data: decodedData}, nil
// }

// AbiInterface represents an ABI Interface, including a name and a list of functions.
type AbiInterface struct {
	Name      string
	Functions []AbiFunction
}

// NewAbiInterface is a constructor function for creating a new AbiInterface instance.
func NewAbiInterface(name string, functions []AbiFunction) *AbiInterface {
	return &AbiInterface{
		Name:      name,
		Functions: functions,
	}
}
