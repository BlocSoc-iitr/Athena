package athena_abi

import (
	"errors"
	"strings"
)

// Groups ABI JSON by ABI Type. If type is 'struct' or 'enum', it is grouped as a 'type_def'
func GroupAbiByType(abiJson []map[string]interface{}) map[AbiMemberType][]map[string]interface{} {
	grouped := make(map[AbiMemberType][]map[string]interface{})

	for _, entry := range abiJson {
		if entry["type"] == "struct" || entry["type"] == "enum" {
			grouped["type_def"] = append(grouped["type_def"], entry)
		} else {
			grouped[entry["type"].(AbiMemberType)] = append(grouped[entry["type"].(AbiMemberType)], entry)
		}
	}
	return grouped
}

// Non-Struct Defined Types
// Used for Topological Sorting abi struct and enum definitions of incorrectly ordered abis
var StarknetCoreTypes = map[string]struct{}{
	"felt":                {}, // Old Syntax for core::felt252
	"felt*":               {}, // Old Syntax for arrays
	"core::integer::u128": {},
	"core::integer::u64":  {},
	"core::integer::u32":  {},
	"core::integer::u16":  {},
	"core::integer::u8":   {},
	"core::felt252":       {},
	"core::bool":          {},
	"core::starknet::contract_address::ContractAddress": {},
	"core::starknet::class_hash::ClassHash":             {},
	"core::starknet::eth_address::EthAddress":           {},
}

func extractInnerType(abiType string) string {
	start := strings.Index(abiType, "<")
	end := strings.LastIndex(abiType, ">")

	if start == -1 || end == -1 || start+1 >= end {
		return abiType
	}

	return abiType[start+1 : end]
}

// The function takes in a list of type definitions (dict) and returns a dict of sets (map[string]bool)
func BuildTypeGraph(typeDefs []map[string]interface{}) map[string]map[string]bool {
	outputGraph := make(map[string]map[string]bool)
	for _, typeDef := range typeDefs {
		referencedTypes := []string{}
		if typeDef["type"] == "struct" {
			for _, member := range typeDef["members"].([]map[string]interface{}) {
				referencedTypes = append(referencedTypes, member["type"].(string))
			}
		} else {
			for _, variant := range typeDef["variants"].([]map[string]interface{}) {
				referencedTypes = append(referencedTypes, variant["type"].(string))
			}
		}

		refTypes := make(map[string]bool)

		for _, typeStr := range referencedTypes {
			if _, ok := StarknetCoreTypes[typeStr]; ok {
				continue
			}

			if _, ok := StarknetCoreTypes[extractInnerType(typeStr)]; ok {
				if strings.HasPrefix(typeStr, "core::array") || strings.HasPrefix(typeStr, "@core::array") {
					continue
				}
			}

			refTypes[typeStr] = true
		}

		outputGraph[typeDef["name"].(string)] = refTypes
	}

	return outputGraph
}

func TopoSortTypeDefs(typeDefs []map[string]interface{}) ([]map[string]interface{}, error) {
	typeGraph := BuildTypeGraph(typeDefs)
	sortedDefs := TopologicalSort(convertMap(typeGraph))

	sortedTypeDefJson := []map[string]interface{}{}

	for _, sortedTypeName := range sortedDefs {
		abiDefinition := []map[string]interface{}{}
		for _, typeDef := range typeDefs {
			if typeDef["name"] == sortedTypeName {
				abiDefinition = append(abiDefinition, typeDef)
			}
		}
		if len(abiDefinition) == 0 {
			return nil, &InvalidAbiError{
				Msg: "Type " + sortedTypeName + " not defined in ABI",
			}
		}
		if len(abiDefinition) > 1 {
			return nil, &InvalidAbiError{
				Msg: "Type " + sortedTypeName + " defined multiple times in ABI",
			}
		}
		sortedTypeDefJson = append(sortedTypeDefJson, abiDefinition[0])
	}
	return sortedTypeDefJson, nil
}

// Parses an **ordered** array of ABI structs into a dictionary of StarknetStructs, mapping struct name to struct.
// return value is a map from string to StarknetStruct or StarknetEnum
func ParseEnumsAndStructs(abiStructs []map[string]interface{}) (map[string]interface{}, error) {
	outputTypes := make(map[string]interface{})

	for _, abiStruct := range abiStructs {
		typeName := abiStruct["name"].(string)
		typeParts := strings.Split(typeName, "::")

		switch {
		case typeName == "Uint256":
			continue

		case len(typeParts) > 1 && (typeParts[0] == "core" || typeParts[0] == "@core") &&
			(typeParts[1] == "array" || typeParts[1] == "integer" || typeParts[1] == "bool" || typeParts[1] == "option" || typeParts[1] == "zeroable"):
			continue

		}

		switch abiStruct["type"] {
		case "struct":
			res, err := parseStruct(abiStruct, outputTypes)
			if err != nil {
				return nil, err
			}
			outputTypes[typeName] = res

		case "enum":
			res, err := parseEnum(abiStruct, outputTypes)
			if err != nil {
				return nil, err
			}
			outputTypes[typeName] = res
		}
	}

	return outputTypes, nil
}

func parseStruct(abiStruct map[string]interface{}, typeContext map[string]interface{}) (StarknetStruct, error) {
	members := []AbiParameter{}

	for _, member := range abiStruct["members"].([]map[string]interface{}) {
		res, err := parseType(member["type"].(string), typeContext)
		if err != nil {
			return StarknetStruct{}, err
		}
		members = append(members, AbiParameter{
			Name: member["name"].(string),
			Type: res,
		})
	}

	return StarknetStruct{
		Name:    abiStruct["name"].(string),
		Members: members,
	}, nil
}

func parseEnum(abiEnum map[string]interface{}, typeContext map[string]interface{}) (StarknetEnum, error) {
	variants := []struct {
		Name string
		Type StarknetType
	}{}

	for _, variant := range abiEnum["variants"].([]map[string]interface{}) {
		res, err := parseType(variant["type"].(string), typeContext)
		if err != nil {
			return StarknetEnum{}, err
		}
		variants = append(variants, struct {
			Name string
			Type StarknetType
		}{
			Name: variant["name"].(string),
			Type: res,
		})
	}

	return StarknetEnum{
		Name:     abiEnum["name"].(string),
		Variants: variants,
	}, nil
}

func parseType(abiType string, customTypes map[string]interface{}) (StarknetType, error) {
	if abiType == "()" {
		return NoneType, nil
	}

	if strings.HasPrefix(abiType, "(") {
		return parseTuple(abiType, customTypes), nil
	}

	parts := strings.Split(abiType, "::")[1:]
	switch {
	case len(parts) == 1 && parts[0] == "felt252":
		return Felt, nil
	case len(parts) == 1 && parts[0] == "bool":
		return Bool, nil
	case len(parts) == 3 && parts[0] == "starknet" && parts[1] == "contract_address" && parts[2] == "ContractAddress":
		return ContractAddress, nil
	case len(parts) == 3 && parts[0] == "starknet" && parts[1] == "class_hash" && parts[2] == "ClassHash":
		return ClassHash, nil
	case len(parts) == 3 && parts[0] == "starknet" && parts[1] == "eth_address" && parts[2] == "EthAddress":
		return EthAddress, nil
	case len(parts) == 2 && parts[0] == "bytes_31" && parts[1] == "bytes31":
		return Bytes31, nil
	case len(parts) == 3 && parts[0] == "starknet" && parts[1] == "storage_access" && parts[2] == "StorageAddress":
		return StorageAddress, nil
	case len(parts) >= 2 && parts[0] == "array" && parts[1] == "Array" || parts[1] == "Span":
		res, err := parseType(extractInnerType(abiType), customTypes)
		if err != nil {
			return nil, err
		}
		return StarknetArray{res}, nil
	case len(parts) >= 2 && parts[0] == "option" && parts[1] == "Option":
		res, err := parseType(extractInnerType(abiType), customTypes)
		if err != nil {
			return nil, err
		}
		return StarknetOption{res}, nil
	case len(parts) >= 2 && parts[0] == "zeroable" && parts[1] == "NonZero":
		res, err := parseType(extractInnerType(abiType), customTypes)
		if err != nil {
			return nil, err
		}
		return StarknetNonZero{res}, nil
	default:
		if val, exists := customTypes[abiType]; exists {
			return val.(StarknetType), nil
		}
		if abiType == "felt" {
			return Felt, nil
		}
		if abiType == "Uint256" {
			return U256, nil
		}
		if strings.HasSuffix(abiType, "*") {
			res, err := parseType(strings.TrimSuffix(abiType, "*"), customTypes)
			if err != nil {
				return nil, err
			}
			return StarknetArray{res}, nil
		}
		return nil, &InvalidAbiError{
			Msg: "Invalid ABI type: " + abiType,
		}
	}
}

func parseTuple(abiType string, customTypes map[string]interface{}) (StarknetTuple, error) {
	// Helper function to check if the type string represents a named tuple
	_isNamedTuple := func(typeStr string) int {
		if len(typeStr) == 0 {
			fmt.Printf("Length of typeStr is zero.")
		}
		for i := 0; i < len(typeStr); i++ {
			if typeStr[i] == ':' {
				// Check if the colon is not preceded or followed by another colon
				if (i == 0 || typeStr[i-1] != ':') && (i == len(typeStr)-1 || typeStr[i+1] != ':') {
					return i
				}
			}
		}
		return -1
	}

	// Error: Empty input
	if len(abiType) < 2 {
		return StarknetTuple{}, errors.New("invalid ABI type: input too short")
	}

	// Remove Outer Parentheses & Whitespace
	strippedTuple := strings.TrimSpace(abiType[1 : len(abiType)-1])

	members := []StarknetType{}
	parenthesisCache := []string{}
	typeCache := []string{}

	// Iterate over each type string in the tuple
	for _, typeString := range strings.Split(strippedTuple, ",") {
		typeString = strings.TrimSpace(typeString)

		tupleOpen := strings.Count(typeString, "(")
		tupleClose := strings.Count(typeString, ")")

		// Error: Mismatched parentheses
		if tupleClose > len(parenthesisCache) {
			return StarknetTuple{}, errors.New("mismatched parentheses in ABI type")
		}

		// Handle opening parentheses
		if tupleOpen > 0 {
			for i := 0; i < tupleOpen; i++ {
				parenthesisCache = append(parenthesisCache, "(")
			}
		}

		if len(parenthesisCache) > 0 {
			// Parsing inside a nested tuple
			typeCache = append(typeCache, typeString)
		} else {
			// Parsing at root level
			if pos := _isNamedTuple(typeString); pos != -1 {
				parsedType, err := parseType(typeString[pos+1:], customTypes)
				if err != nil {
					return StarknetTuple{}, err
				}
				members = append(members, parsedType)
			} else {
				parsedType, err := parseType(typeString, customTypes)
				if err != nil {
					return StarknetTuple{}, err
				}
				members = append(members, parsedType)
			}
		}

		// Handle closing parentheses
		if tupleClose > 0 {
			for i := 0; i < tupleClose; i++ {
				parenthesisCache = parenthesisCache[:len(parenthesisCache)-1]
			}

			// If we have closed all parentheses for the current nested tuple
			if len(parenthesisCache) == 0 {
				parsedTuple, err := parseTuple(strings.Join(typeCache, ","), customTypes)
				if err != nil {
					return StarknetTuple{}, err
				}
				members = append(members, parsedTuple)
				typeCache = []string{}
			}
		}
	}

	// Error: Unbalanced parentheses
	if len(parenthesisCache) > 0 {
		return StarknetTuple{}, errors.New("unbalanced parentheses in ABI type")
	}

	// Return the successfully parsed tuple
	return StarknetTuple{Members: members}, nil
}
