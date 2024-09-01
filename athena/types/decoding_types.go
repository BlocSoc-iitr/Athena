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
	decodedInputs, _ := DecodeFromParams(af.Inputs, &calldataCopy)

	var decodedOutputs []interface{}
	if result != nil {
		resultCopy := make([]int, len(result))
		copy(resultCopy, result)
		decodedOutputs, _ = DecodeFromTypes(af.Outputs, &resultCopy)
	}

	return DecodedFunction{
		AbiName:           af.AbiName,
		Name:              af.Name,
		FunctionSignature: fmt.Sprintf("%x", af.Signature),
		Input:             decodedInputs,
		Output:            decodedOutputs,
	}

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

func (e *AbiEvent) Decode(data []int, keys []int) (DecodedEvent, error) {
	_data := make([]int, len(data))
	copy(_data, data)

	_keys := make([]int, len(keys)-1)
	copy(_keys, keys[1:]) // Skip the first key (event signature)

	decodedData := make(map[string]interface{})

	for _, param := range e.Parameters {
		var value interface{}
		//var err error

		if _, ok := e.Data[param]; ok {
			// Decode from data
			decoded, err := DecodeFromTypes([]StarknetType{e.Data[param]}, &_data)
			if err != nil {
				return DecodedEvent{}, err
			}
			value = decoded[0]
		} else if _, ok := e.Keys[param]; ok {
			// Decode from keys
			decoded, err := DecodeFromTypes([]StarknetType{e.Keys[param]}, &_keys)
			if err != nil {
				return DecodedEvent{}, err
			}
			value = decoded[0]
		} else {
			return DecodedEvent{}, fmt.Errorf("event Parameter %s not present in Keys or Data for Event %s", param, e.Name)
		}

		decodedData[param] = value
	}

	if len(_data) != 0 || len(_keys) != 0 {
		return DecodedEvent{}, fmt.Errorf(" calldata Not Completely Consumed decoding Event")
	}

	return DecodedEvent{
		AbiName: *e.AbiName,
		Name:    e.Name,
		Data:    decodedData,
	}, nil
}

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
