package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

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

func getStarknetClassABI(classHash, jsonRpcUrl string) (json.RawMessage, error) {
	// Create the JSON-RPC request payload
	request := JsonRPCRequest{
		Jsonrpc: "2.0",
		Method:  "starknet_getClass",
		Params: map[string]interface{}{
			"class_hash": classHash,
			"block_id":   "latest",
		},
		ID: 1,
	}

	// Serialize the request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}

	// Make the HTTP POST request
	resp, err := http.Post(jsonRpcUrl, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read and deserialize the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var jsonResponse JsonRPCResponse
	if err := json.Unmarshal(responseBody, &jsonResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors in the JSON-RPC response
	if jsonResponse.Error.Code != 0 {
		return nil, fmt.Errorf("JSON-RPC error: %d, message: %s, data: %s", jsonResponse.Error.Code, jsonResponse.Error.Message, jsonResponse.Error.Data)
	}

	// Return the ABI as json.RawMessage
	return jsonResponse.Result.Abi, nil
}

func saveToFile(filename string, data []byte) error {
	// Create or open the file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write data to the file
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data to file: %w", err)
	}

	return nil
}

func parseAndFormatABI(abiJson json.RawMessage) ([]byte, error) {
	var result interface{}
	if err := json.Unmarshal(abiJson, &result); err != nil {
		return nil, fmt.Errorf("failed to parse ABI JSON: %w", err)
	}

	// Format ABI JSON
	formattedABI, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to format ABI JSON: %w", err)
	}
	return formattedABI, nil
}

func main() {
	classHash := "0x029927c8af6bccf3f6fda035981e765a7bdbf18a2dc0d630494f8758aa908e2b"
	jsonRpcUrl := "https://free-rpc.nethermind.io/mainnet-juno/"

	// Get ABI as json.RawMessage
	abi, err := getStarknetClassABI(classHash, jsonRpcUrl)
	if err != nil {
		log.Fatalf("Error getting ABI: %v", err)
	}

	// Parse and format ABI JSON
	formattedABI, err := parseAndFormatABI(abi)
	if err != nil {
		log.Fatalf("Error parsing and formatting ABI: %v", err)
	}

	// Save formatted ABI JSON to file
	filename := "abi.json"
	if err := saveToFile(filename, formattedABI); err != nil {
		log.Fatalf("Error saving ABI to file: %v", err)
	}

	fmt.Printf("ABI for class %s has been saved to %s\n", classHash, filename)
}
