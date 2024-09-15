package athena_abi

import (
	"fmt"
	"math/big"
)

// class representing the result of decoding an ABI
type DecodedFunction struct {
	abiName string
	name    string
	inputs  map[string]interface{}
	outputs []interface{}
}

// class representing the result of decoding an ABI Event
type DecodedEvent struct {
	abiName string
	name    string
	data    map[string]interface{}
}

// class Representing an ABI Function.  Includes a function name, the function signature, and the input
// and output parameters.
type AbiFunction struct {
	name      string
	abiName   string
	signature []byte
	inputs    []AbiParameter
	outputs   []StarknetType
}

func NewAbiFunction(name string, inputs []AbiParameter, outputs []StarknetType, abiName string) *AbiFunction {
	return &AbiFunction{
		name:      name,
		abiName:   abiName,
		inputs:    inputs,
		outputs:   outputs,
		signature: StarknetKeccak([]byte(name)),
	}
}

func (af *AbiFunction) Encode(inputs map[string]interface{}) []*big.Int {
	res, err := EncodeFromParams(af.inputs, inputs)
	if err != nil {
		return nil
	}
	return res
}

func (af *AbiFunction) idStr() string {
	inputStr := ""
	for _, param := range af.inputs {
		inputStr += param.idStr()
		inputStr += ","
	}
	inputStr = inputStr[:len(inputStr)-1]
	outputStr := ""
	for _, output := range af.outputs {
		outputStr += output.idStr()
		outputStr += ","
	}
	outputStr = outputStr[:len(outputStr)-1]
	return "Function(" + inputStr + ") -> (" + outputStr + ")"
}

// decode the calldata and result of a function
// result can either be nil or []big.Int
func (af *AbiFunction) Decode(callData []big.Int, result interface{}) (*DecodedFunction, error) {
	callDataCopy := make([]big.Int, len(callData))

	copy(callDataCopy, callData)

	decodedInputs, err := DecodeFromParams(af.inputs, callDataCopy)
	if err != nil {
		return nil, err
	}

	var decodedOutputs []interface{}

	if result != nil {
		resultCopy := make([]big.Int, len(result.([]big.Int)))
		copy(resultCopy, result.([]big.Int))
		decodedOutputs, err = DecodeFromTypes(af.outputs, resultCopy)
		if err != nil {
			return nil, err
		}
	} else {
		decodedOutputs = nil
	}

	return &DecodedFunction{
		abiName: af.abiName,
		name:    af.name,
		inputs:  decodedInputs,
		outputs: decodedOutputs,
	}, nil
}

type AbiEvent struct {
	name       string
	abiName    string
	signature  []byte
	parameters []string
	keys       map[string]StarknetType
	data       map[string]StarknetType
}

// keys can be either map[string]StarknetType or nil and abiName can be either string or Nil
func NewAbiEvent(name string, parameters []string, data map[string]StarknetType, keys interface{}, abiName interface{}) *AbiEvent {
	return &AbiEvent{
		name:       name,
		parameters: parameters,
		data:       data,
		abiName:    abiName.(string),
		keys:       keys.(map[string]StarknetType),
		signature:  StarknetKeccak([]byte(name)),
	}
}

func (ae AbiEvent) idStr() (string, error) {
	eventParamsString := ""
	for _, param := range ae.parameters {
		if value, exists := ae.data[param]; exists {
			eventParamsString += fmt.Sprintf("%s:%s", param, value.idStr())
			eventParamsString += ","
		} else if value, exists := ae.keys[param]; exists {
			eventParamsString += fmt.Sprintf("<%s>:%s", param, value.idStr())
			eventParamsString += ","
		} else {
			return "", &TypeDecodeError{
				Msg: fmt.Sprintf("Event Parameter %s not part of event keys or Data", param),
			}
		}
	}
	eventParamsString = eventParamsString[:len(eventParamsString)-1]
	return "Event(" + eventParamsString + ")", nil
}

func (ae AbiEvent) Decode(data []big.Int, keys []big.Int) (*DecodedEvent, error) {
	dataCopy := make([]big.Int, len(data))
	copy(dataCopy, data)
	keyCopy := make([]big.Int, len(keys)-1)
	copy(keyCopy, keys[1:])
	decodedData := map[string]interface{}{}

	for _, param := range ae.parameters {
		if value, exists := ae.data[param]; exists {
			result, err := DecodeFromTypes([]StarknetType{value}, dataCopy)
			if err != nil {
				return nil, err
			}
			decodedData[param] = result[0]
		} else if value, exists := ae.keys[param]; exists {
			result, err := DecodeFromTypes([]StarknetType{value}, keyCopy)
			if err != nil {
				return nil, err
			}
			decodedData[param] = result[0]
		} else {
			return nil, &TypeDecodeError{
				Msg: fmt.Sprintf("Event Parameter %s not present in Keys or Data for Event %s", param, ae.name),
			}
		}
	}

	return &DecodedEvent{
		abiName: ae.abiName,
		name:    ae.name,
		data:    decodedData,
	}, nil
}

// class Representing an ABI Interface.  Includes a name and a list of functions.
type AbiInterface struct {
	name      string
	functions []AbiFunction
}
