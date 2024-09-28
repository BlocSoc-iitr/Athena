package athena_abi

import (
	"crypto/md5"
	"errors"
	"fmt"
	"math/big"
)

type ClassNotFoundError struct {
	Msg string
}

func (e *ClassNotFoundError) Error() string {
	return fmt.Sprintf("Class with hash %s not found", e.Msg)
}

type FunctionDispatchInfo struct {
	DecoderReference [8]byte
	FunctionName     string
}

type EventDispatchInfo struct {
	DecoderReference [8]byte
	EventName        string
}

type ClassDispatcher struct {
	FunctionIDs map[[8]byte]FunctionDispatchInfo
	EventIDs    map[[8]byte]EventDispatchInfo
	AbiName     *string
	ClassHash   []byte
}

type DecodingDispatcher struct {
	ClassIDs      map[[8]byte]ClassDispatcher
	FunctionTypes map[[8]byte]FunctionType
	EventTypes    map[[8]byte]EventType
}

type FunctionType struct {
	InputParams  []AbiParameter
	OutputParams []StarknetType
}

type EventType struct {
	Parameters []string
	Keys       map[string]StarknetType
	Data       map[string]StarknetType
}

// Helper to get consistent 8-byte hash
func idHash(idStr string) [8]byte {
	hash := md5.Sum([]byte(idStr))
	var shortHash [8]byte
	copy(shortHash[:], hash[8:])
	return shortHash
}

func NewDecodingDispatcher() *DecodingDispatcher {
	return &DecodingDispatcher{
		ClassIDs:      make(map[[8]byte]ClassDispatcher),
		FunctionTypes: make(map[[8]byte]FunctionType),
		EventTypes:    make(map[[8]byte]EventType),
	}
}

func (d *DecodingDispatcher) GetClass(classHash [32]byte) (*ClassDispatcher, bool) {
	classID := classHash[24:]
	var key [8]byte

	copy(key[:], classID)

	classDispatcher, exists := d.ClassIDs[key]
	return &classDispatcher, exists
}

func (d *DecodingDispatcher) AddAbiFunctions(abi StarknetABI) map[[8]byte]FunctionDispatchInfo {

	functionIDs := make(map[[8]byte]FunctionDispatchInfo)

	for _, function := range abi.Functions {
		functionTypeID := idHash(function.name)

		if _, exists := d.FunctionTypes[functionTypeID]; !exists {
			d.FunctionTypes[functionTypeID] = FunctionType{
				InputParams:  function.inputs,
				OutputParams: function.outputs,
			}

		}

		if len(function.signature) < 24 {
			continue // Skip this function
		}

		var key [8]byte
		copy(key[:], function.signature[24:])

		functionIDs[key] = FunctionDispatchInfo{
			DecoderReference: functionTypeID,
			FunctionName:     function.name,
		}
	}
	return functionIDs
}

func (d *DecodingDispatcher) AddAbiEvents(abi StarknetABI) map[[8]byte]EventDispatchInfo {
	eventIDs := make(map[[8]byte]EventDispatchInfo)

	for _, event := range abi.Events {
		eventTypeID := idHash(event.name)

		if _, exists := d.EventTypes[eventTypeID]; !exists {
			d.EventTypes[eventTypeID] = EventType{
				Parameters: event.parameters,
				Keys:       event.keys,
				Data:       event.data,
			}
		}

		if len(event.signature) < 24 {
			continue // Skip this function
		}

		var key [8]byte
		copy(key[:], event.signature[24:])

		eventIDs[key] = EventDispatchInfo{
			DecoderReference: eventTypeID,
			EventName:        event.name,
		}
	}
	return eventIDs
}

func (d *DecodingDispatcher) AddAbi(abi StarknetABI) {

	if len(abi.ClassHash) < 32 {
		fmt.Printf("ClassHash is too short: %d bytes", len(abi.ClassHash))
		return
	}
	classID := abi.ClassHash[24:]
	var key [8]byte

	copy(key[:], classID)

	d.ClassIDs[key] = ClassDispatcher{
		AbiName:     abi.ABIName,
		ClassHash:   abi.ClassHash,
		FunctionIDs: d.AddAbiFunctions(abi),
		EventIDs:    d.AddAbiEvents(abi),
	}
}

func (d *DecodingDispatcher) DecodeFunction(calldata *[]*big.Int, result *[]*big.Int, functionSelector, classHash [32]byte) (*DecodedFunction, error) {
	classDispatcher, exists := d.GetClass(classHash)
	if !exists {
		return nil, errors.New("class not found")
	}

	functionID := functionSelector[24:]
	var key [8]byte

	copy(key[:], functionID)

	functionDispatchInfo, exists := classDispatcher.FunctionIDs[key]
	if !exists {
		return nil, errors.New("function not found")
	}

	inputOutputTypes := d.FunctionTypes[functionDispatchInfo.DecoderReference]
	inputTypes := inputOutputTypes.InputParams
	outputTypes := inputOutputTypes.OutputParams

	decodedInputs, err := DecodeFromParams(inputTypes, calldata)
	if err != nil {
		return nil, err
	}

	if len(*calldata) != 0 {
		return nil, errors.New("calldata remaining after decoding inputs")
	}

	decodedOutputs, errWhileDecoding := DecodeFromTypes(outputTypes, result)
	if errWhileDecoding != nil {
		return nil, errWhileDecoding
	}

	if len(*result) > 0 && len(outputTypes) > 0 {
		_retry := *result
		decodedOutput, err := DecodeFromTypes(outputTypes, &_retry)
		if err != nil {
			return nil, fmt.Errorf("calldata remaining after decoding function result %v from %v: %w", result, outputTypes, err)
		}

		decodedOutputs[0] = append(decodedOutputs[0].([]interface{}), decodedOutput[0])
	}

	// If there's only one output and it's a StarknetArray, return the single decoded output.
	if len(decodedOutputs) == 1 {
		if _, ok := outputTypes[0].(StarknetArray); ok {
			decodedOutputs = decodedOutputs[0].([]interface{})
		}
	}

	return &DecodedFunction{
		abiName: *classDispatcher.AbiName,
		name:    functionDispatchInfo.FunctionName,
		inputs:  decodedInputs,
		outputs: decodedOutputs,
	}, nil
}

func (d *DecodingDispatcher) DecodeEvent(data *[]*big.Int, keys *[]*big.Int, classHash [32]byte) (*DecodedEvent, error) {
	classDispatcher, _ := d.GetClass(classHash)
	if classDispatcher == nil {
		return nil, &ClassNotFoundError{
			Msg: fmt.Sprintf("Class 0x%x not found", classHash),
		}
	}

	if len(*keys) == 0 {
		return nil, &InvalidCalldataError{
			Msg: "Events require at least 1 key parameter as the selector",
		}
	}

	// Convert keys[0] to 32 bytes for the event selector
	eventSelector := make([]byte, 32)
	var eventKey [8]uint8

	copy(eventKey[:], eventSelector[len(eventSelector)-8:])

	eventDispatcher := classDispatcher.EventIDs[eventKey]
	eventParams := d.EventTypes[eventDispatcher.DecoderReference].Parameters
	eventKeys := d.EventTypes[eventDispatcher.DecoderReference].Keys
	eventData := d.EventTypes[eventDispatcher.DecoderReference].Data

	_data := append([]*big.Int{}, *data...)
	_keys := append([]*big.Int{}, *keys...) // Skip the first key (event signature)

	decodedData := make(map[string]interface{})

	for _, param := range eventParams {
		if eventData[param] != nil {
			decodedValue, err := DecodeFromTypes([]StarknetType{eventData[param]}, &_data)
			if err != nil {
				return nil, err
			}
			decodedData[param] = decodedValue[0]
		} else if eventKeys[param] != nil {
			decodedValue, err := DecodeFromTypes([]StarknetType{eventKeys[param]}, &_keys)
			if err != nil {
				return nil, err
			}
			decodedData[param] = decodedValue[0]
		} else {
			return nil, &TypeDecodeError{
				Msg: fmt.Sprintf("Event Parameter %v not present in Keys or Data for Event 0x%x for class 0x%x", param, eventSelector, classHash),
			}
		}
	}

	// Check if all data and keys were consumed
	if len(_data) != 0 || len(_keys) != 0 {
		return nil, &InvalidCalldataError{
			Msg: fmt.Sprintf("Calldata not completely consumed decoding Event 0x%x for class 0x%x. Keys: %v Data: %v", eventSelector, classHash, keys, data),
		}
	}

	// Return decoded event
	return &DecodedEvent{
		abiName: *classDispatcher.AbiName,
		name:    eventDispatcher.EventName,
		data:    decodedData,
	}, nil
}
