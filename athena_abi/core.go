package athena_abi

import (
	"errors"
	"log"
)

type StarknetABI struct {
	ABIName               *string
	ClassHash             []byte
	Functions             map[string]AbiFunction
	Events                map[string]AbiEvent
	Constructor           []AbiParameter
	L1Handler             *AbiFunction
	ImplementedInterfaces map[string]AbiInterface
}

// Declare errors
var (
	errParseDefinedTypes          = errors.New("unable to parse defined types")
	errParseInterfaces            = errors.New("unable to parse interfaces")
	errParseFunctions             = errors.New("unable to parse functions")
	errParseEvents                = errors.New("unable to parse events")
	errParseConstructor           = errors.New("unable to parse constructor")
	errParseL1Handler             = errors.New("unable to parse L1 handler")
	errParseImplementedInterfaces = errors.New("unable to parse implemented interfaces")
)

// Parse Starknet ABI from JSON
// @param abiJSON
// @param abiname
// @param classHash
func StarknetAbiFromJSON(abiJson []map[string]interface{}, abiName string, classHash []byte) (*StarknetABI, error) {
	groupedAbi := GroupAbiByType(abiJson)

	// Parse defined types (structs and enums)
	definedTypes, err := ParseEnumsAndStructs(groupedAbi["type_def"])
	if err != nil {
		sortedDefs, errDef := TopoSortTypeDefs(groupedAbi["type_def"])
		if errDef == nil {
			defineTypes, errDtypes := ParseEnumsAndStructs(sortedDefs)
			definedTypes = defineTypes
			errDef = errDtypes
		}
		if errDef != nil {
			return nil, errParseDefinedTypes
		}
		log.Println("ABI Struct and Enum definitions out of order & required topological sorting")
	}

	// Parse interfaces
	var definedInterfaces []AbiInterface
	for _, iface := range groupedAbi["interface"] {
		functions := []AbiFunction{}
		for _, funcData := range iface["items"].([]interface{}) {
			parsedAbi, errWhileParsing := ParseAbiFunction(funcData.(map[string]interface{}), definedTypes)
			if errWhileParsing != nil {
				return nil, errParseInterfaces
			}
			functions = append(functions, *parsedAbi)
		}
		definedInterfaces = append(definedInterfaces, AbiInterface{
			name:      iface["name"].(string),
			functions: functions,
		})
	}

	// Parse functions
	functions := make(map[string]AbiFunction)
	for _, functionData := range groupedAbi["function"] {
		funcName := functionData["name"].(string)
		abiFunc, errParsingFunctions := ParseAbiFunction(functionData, definedTypes)
		if errParsingFunctions != nil {
			return nil, errParseFunctions
		}
		functions[funcName] = *abiFunc
	}

	// Add functions from interfaces
	for _, iface := range definedInterfaces {
		for _, function := range iface.functions {
			functions[function.name] = function
		}
	}

	// Parse events
	parsedAbiEvents := []AbiEvent{}
	for _, eventData := range groupedAbi["event"] {
		parsedEvent, errParsingEvent := ParseAbiEvent(eventData, definedTypes)
		if errParsingEvent != nil {
			return nil, errParseEvents
		}
		parsedAbiEvents = append(parsedAbiEvents, *parsedEvent)
	}

	events := make(map[string]AbiEvent)
	for _, event := range parsedAbiEvents {
		if event.name != "" {
			events[event.name] = event
		}
	}

	// Parse constructor
	var constructor []AbiParameter
	if len(groupedAbi["constructor"]) == 1 {
		for _, paramData := range groupedAbi["constructor"][0]["inputs"].([]interface{}) {
			param := paramData.(map[string]interface{})
			typed, errorParsingType := parseType(param["type"].(string), definedTypes)
			if errorParsingType != nil {
				return nil, errParseConstructor
			}
			constructor = append(constructor, AbiParameter{
				Name: param["name"].(string),
				Type: typed,
			})
		}
	} else {
		constructor = nil
	}

	// Parse L1 handler
	var l1Handler *AbiFunction
	if len(groupedAbi["l1_handler"]) == 1 {
		handler, errorParsingFunction := ParseAbiFunction(groupedAbi["l1_handler"][0], definedTypes)
		if errorParsingFunction != nil {
			return nil, errParseL1Handler
		}
		l1Handler = handler
	} else {
		l1Handler = nil
	}

	// Parse implemented interfaces
	implementedInterfaces := make(map[string]AbiInterface)
	implArray, ok := groupedAbi["impl"]
	if !ok {
		return nil, errParseImplementedInterfaces
	}
	for _, implData := range implArray {
		implMap := implData
		if ifaceName, ok := implMap["interface_name"].(string); ok {
			for _, iface := range definedInterfaces {
				if iface.name == ifaceName {
					implementedInterfaces[iface.name] = iface
				}
			}
		}
	}

	// Return the populated StarknetAbi struct
	return &StarknetABI{
		ABIName:               &abiName,
		ClassHash:             classHash,
		Functions:             functions,
		Events:                events,
		Constructor:           constructor,
		L1Handler:             l1Handler,
		ImplementedInterfaces: implementedInterfaces,
	}, nil
}
