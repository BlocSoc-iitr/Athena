package decoder

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type InterfaceType struct {
	Type  string     `json:"type"`
	Name  string     `json:"name"`
	Items []Function `json:"items"`
}

type EventType struct {
	Kind    string   `json:"kind"`
	Name    string   `json:"name"`
	Members []Member `json:"members"`
}

type Function struct {
	Type            string   `json:"type"`
	Name            string   `json:"name"`
	Inputs          []Member `json:"inputs"`
	Outputs         []Output `json:"outputs"`
	StateMutability string   `json:"state_mutability"`
}

type Member struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Kind string `json:"kind"`
}

type Output struct {
	Type string `json:"type"`
}

func TypeToReadableName(typ string) string {
	typeMappings := map[string]string{
		"felt252":       "Felt252",
		"address":       "ContractAddress",
		"u256":          "U256",
		"core::felt252": "Felt252",
		"core::address": "ContractAddress",
		"core::u256":    "U256",
	}

	parts := strings.Split(typ, "::")
	if len(parts) > 0 {
		if name, exists := typeMappings[parts[len(parts)-1]]; exists {
			return name
		}
	}

	return typ
}

func GetParsedAbi(abi_to_decode json.RawMessage) {
	var abi []map[string]interface{}
	var abiString string
	if err := json.Unmarshal(abi_to_decode, &abiString); err == nil {
		err = json.Unmarshal([]byte(abiString), &abi)
		if err != nil {
			log.Fatalf("Error parsing ABI string: %v", err)
		}
	} else {
		log.Fatalf("Error unmarshalling ABI: %v", err)
	}

	// var events []string
	// var functions []string

	// Process ABI items
	for _, item := range abi {
		itemBytes, err := json.Marshal(item)
		if err != nil {
			log.Printf("Error marshaling item: %v", err)
			continue
		}

		switch item["type"] {
		case "interface":
			var i InterfaceType
			if err := json.Unmarshal(itemBytes, &i); err == nil {
				for _, funcItem := range i.Items {
					inputs := []string{}
					for _, input := range funcItem.Inputs {
						inputType := TypeToReadableName(input.Type)
						inputs = append(inputs, fmt.Sprintf("%s: %s", input.Name, inputType))
					}
					outputs := []string{}
					for _, output := range funcItem.Outputs {
						outputType := TypeToReadableName(output.Type)
						outputs = append(outputs, outputType)
					}
					inputSignature := strings.Join(inputs, ", ")
					outputSignature := strings.Join(outputs, ", ")
					fmt.Printf("Function: %s(%s) -> (%s) [State Mutability: %s]\n", funcItem.Name, inputSignature, outputSignature, funcItem.StateMutability)
				}
			}
		case "event":
			var event EventType
			if err := json.Unmarshal(itemBytes, &event); err == nil {
				parts := strings.Split(event.Name, "::")
				eventName := parts[len(parts)-1]
				members := []string{}
				for _, member := range event.Members {
					memberType := TypeToReadableName(member.Type)
					members = append(members, fmt.Sprintf("%s: %s", member.Name, memberType))
				}
				membersSignature := strings.Join(members, ", ")
				fmt.Printf("Event: %s(%s)\n", eventName, membersSignature)
			}
		default:
			continue
		}
	}
}

// map[
// 	kind:struct
// 	members:[map[kind:key name:owner type:core::felt252] map[kind:data name:guardian
//           type:core::felt252]]
// 	name:account::argent_account::ArgentAccount::AccountCreated
// 	type:event
// 	]
