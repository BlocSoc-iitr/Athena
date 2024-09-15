package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/BlocSoc-iitr/Athena/athena/decoder"
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

// Main function with CLI integration
func main() {
	// Define CLI flags
	classHash := flag.String("classHash", "", "The class hash of the StarkNet contract.")
	jsonRpcUrl := flag.String("jsonRpcUrl", "", "The JSON-RPC URL for the StarkNet node.")
	outputFile := flag.String("output", "abi.json", "The file to save the ABI JSON.")
	decode := flag.Bool("decode", false, "Decode the ABI and display readable names.")
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

	if *decode {
		decoder.GetParsedAbi(abi)
	} else {
		abiJson, err := json.MarshalIndent(abi, "", "  ")
		if err != nil {
			log.Fatalf("Error serializing ABI: %v", err)
		}

		// Save ABI JSON to file
		if err := saveToFile(*outputFile, abiJson); err != nil {
			log.Fatalf("Error saving ABI to file: %v", err)
		}

		fmt.Printf("ABI for class %s has been saved to %s\n", *classHash, *outputFile)
	}
}
