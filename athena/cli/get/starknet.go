package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// Structs for JSON-RPC requests and responses
type JsonRPCRequest struct {
	Jsonrpc string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	ID      int                    `json:"id"`
}

type JsonRPCResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		Abi json.RawMessage `json:"abi"`
	} `json:"result"`
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    string `json:"data"`
	} `json:"error"`
}

type AbiItem struct {
	Type            string  `json:"type"`
	Name            string  `json:"name"`
	Inputs          []Param `json:"inputs,omitempty"`
	Outputs         []Param `json:"outputs,omitempty"`
	StateMutability string  `json:"state_mutability,omitempty"`
}

type Param struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Fetch ABI from StarkNet
func getStarknetClassABI(classHash, jsonRpcUrl string) (json.RawMessage, error) {
	request := JsonRPCRequest{
		Jsonrpc: "2.0",
		Method:  "starknet_getClass",
		Params: map[string]interface{}{
			"class_hash": classHash,
			"block_id":   "latest",
		},
		ID: 1,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}

	resp, err := http.Post(jsonRpcUrl, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var jsonResponse JsonRPCResponse
	if err := json.Unmarshal(responseBody, &jsonResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if jsonResponse.Error.Code != 0 {
		return nil, fmt.Errorf("JSON-RPC error: %d, message: %s, data: %s", jsonResponse.Error.Code, jsonResponse.Error.Message, jsonResponse.Error.Data)
	}

	return jsonResponse.Result.Abi, nil
}

// Save data to file
func saveToFile(filename string, data []byte) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data to file: %w", err)
	}

	return nil
}

// Parse ABI JSON
func parseABI(abiJson json.RawMessage) ([]AbiItem, error) {
	// Try parsing directly as JSON array
	var abiItems []AbiItem
	if err := json.Unmarshal(abiJson, &abiItems); err == nil {
		return abiItems, nil
	}

	// If parsing as array fails, try parsing as a string
	var abiString string
	if err := json.Unmarshal(abiJson, &abiString); err != nil {
		return nil, fmt.Errorf("failed to parse ABI JSON: %w", err)
	}

	// Parse the string as JSON array
	if err := json.Unmarshal([]byte(abiString), &abiItems); err != nil {
		return nil, fmt.Errorf("failed to parse ABI JSON string: %w", err)
	}

	return abiItems, nil
}


// Main function with CLI integration
func main() {
	// Define CLI flags
	classHash := flag.String("classHash", "", "The class hash of the StarkNet contract.")
	jsonRpcUrl := flag.String("jsonRpcUrl", "", "The JSON-RPC URL for the StarkNet node.")
	outputFile := flag.String("output", "abi.json", "The file to save the ABI JSON.")
	flag.Parse()

	if *classHash == "" || *jsonRpcUrl == "" {
		flag.Usage()
		return
	}

	// Get ABI as json.RawMessage
	abi, err := getStarknetClassABI(*classHash, *jsonRpcUrl)
	if err != nil {
		log.Fatalf("Error getting ABI: %v", err)
	}

	// Convert ABI to JSON with indentation
	abiJson, err := json.MarshalIndent(abi, "", "  ")
	if err != nil {
		log.Fatalf("Error serializing ABI: %v", err)
	}

	// Save ABI JSON to file
	if err := saveToFile(*outputFile, abiJson); err != nil {
		log.Fatalf("Error saving ABI to file: %v", err)
	}

	fmt.Printf("ABI for class %s has been saved to %s\n", *classHash, *outputFile)

	// Parse and print events and functions from the ABI
	abiItems, err := parseABI(abi)
	if err != nil {
		log.Fatalf("Error parsing ABI: %v", err)
	}

	for _, item := range abiItems {
		switch item.Type {
		case "function":
			fmt.Printf("Function: %s\n", item.Name)
			for _, input := range item.Inputs {
				fmt.Printf("  Input Name: %s, Type: %s\n", input.Name, input.Type)
			}
			for _, output := range item.Outputs {
				fmt.Printf("  Output Name: %s, Type: %s\n", output.Name, output.Type)
			}
		case "event":
			fmt.Printf("Event: %s\n", item.Name)
			for _, input := range item.Inputs {
				fmt.Printf("  Member Name: %s, Type: %s\n", input.Name, input.Type)
			}
		}
	}
}
