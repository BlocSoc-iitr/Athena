package athena_abi

import (
	"errors"
	"fmt"
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
	fmt.Println("hello")
	fmt.Println("grouped abi", groupedAbi["type_def"])
	definedTypes, err := ParseEnumsAndStructs(groupedAbi["type_def"])
	fmt.Println("defined types Map contents:", definedTypes)
	fmt.Println("is there error", err)
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
	fmt.Println("now parsing interfaces")
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
	fmt.Println("now parsing functions")
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
	fmt.Println("now parsing events")
	fmt.Println("the grouped abi events is ", groupedAbi["event"])
	for _, eventData := range groupedAbi["event"] {
		fmt.Println("eventdata is ", eventData)
		parsedEvent, errParsingEvent := ParseAbiEvent(eventData, definedTypes)
		fmt.Println("parsed event is ", parsedEvent)
		fmt.Println("the err is ", errParsingEvent)

		if errParsingEvent != nil {
			return nil, errParseEvents
		}
		//parsedAbiEvents = append(parsedAbiEvents, *parsedEvent)
		if parsedEvent != nil {
			parsedAbiEvents = append(parsedAbiEvents, *parsedEvent)
		} else {
			// Handle the nil case if necessary
			//return nil, errors.New("parsed event is nil")
			continue
		}
	}

	events := make(map[string]AbiEvent)
	fmt.Println("parsedabievents are ", parsedAbiEvents)
	for _, event := range parsedAbiEvents {
		if event.name != "" {
			events[event.name] = event
		}
	}

	// Parse constructor
	var constructor []AbiParameter
	fmt.Println("now parsing constructor")
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
	fmt.Println("now parsing l1handler")
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
	fmt.Println("now parsing implemented interfaces")
	implArray, ok := groupedAbi["impl"]
	if !ok {
		//return nil, errParseImplementedInterfaces
		//this if block is not required
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
	fmt.Println("abi name : ", abiName)
	fmt.Println("classhash is  : ", classHash)
	fmt.Println()
	fmt.Println("functions is  : ", functions)
	fmt.Println()
	fmt.Println("events is  : ", events)
	fmt.Println()
	fmt.Println("constructor is  : ", constructor)
	fmt.Println()
	fmt.Println("l1handles is  ", l1Handler)
	fmt.Println()
	fmt.Println("implemented interfaces is  : ", implementedInterfaces)
	fmt.Println()
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
