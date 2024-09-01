package importers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Define the RPC request structure
type rpcRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

// Define the RPC response structure
type rpcResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error"`
	ID      int             `json:"id"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type L1GasPrice struct {
	PriceInWei string `json:"price_in_wei"`
	PriceInFri string `json:"price_in_fri"`
}

type BlockTxHashes struct {
	ParentHash       string     `json:"parent_hash"`
	Timestamp        int64      `json:"timestamp"`
	SequencerAddress string     `json:"sequencer_address"`
	L1GasPrice       L1GasPrice `json:"l1_gas_price"`
	StarknetVersion  string     `json:"starknet_version"`
	L1DataGasPrice   string     `json:"l1_data_gas_price"`
	L1DAMode         string     `json:"l1_da_mode"`
	BlockHash        *string    `json:"block_hash"` // Pointer to detect nil (pending blocks)
	Transactions     []struct {
		Hash    string `json:"hash"`
		Receipt struct {
			Type            string `json:"type"`
			TransactionHash string `json:"transaction_hash"`
			ActualFee       struct {
				Amount string `json:"amount"`
				Unit   string `json:"unit"`
			} `json:"actual_fee"`
			ExecutionStatus string `json:"execution_status"`
			FinalityStatus  string `json:"finality_status"`
			Events          []struct {
				FromAddress string   `json:"from_address"`
				Keys        []string `json:"keys"`
				Data        []string `json:"data"`
			} `json:"events"`
			ExecutionResources struct {
				Steps                         int `json:"steps"`
				PedersenBuiltinApplications   int `json:"pedersen_builtin_applications"`
				RangeCheckBuiltinApplications int `json:"range_check_builtin_applications"`
				EcdsaBuiltinApplications      int `json:"ecdsa_builtin_applications"`
			} `json:"execution_resources"`
		} `json:"receipt"`
	} `json:"transactions"`
}

func makeRPCCall(ctx context.Context, url string, method string, params interface{}) (*rpcResponse, error) {
	reqBody := rpcRequest{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	// Serialize the request to JSON
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create a new HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var rpcResp rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Check for RPC errors
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error: %s", rpcResp.Error.Message)
	}

	return &rpcResp, nil
}

func GetBlockDetails(ctx context.Context, url string, fromBlockNumber uint64, toBlockNumber uint64) ([]BlockTxHashes, error) {
	var blockDetails []BlockTxHashes
	var wg sync.WaitGroup
	blockChan := make(chan BlockTxHashes, toBlockNumber-fromBlockNumber+1)
	errChan := make(chan error, toBlockNumber-fromBlockNumber+1)

	// Loop through each block number
	for blockNumber := fromBlockNumber; blockNumber <= toBlockNumber; blockNumber++ {
		wg.Add(1)
		go func(blockNumber uint64) {
			defer wg.Done()

			params := map[string]interface{}{
				"block_id": map[string]interface{}{
					"block_number": int(blockNumber),
				},
			}

			resp, err := makeRPCCall(ctx, url, "starknet_getBlockWithReceipts", params)
			if err != nil {
				errChan <- fmt.Errorf("failed to get block details for block %d: %v", blockNumber, err)
				return
			}

			var block BlockTxHashes
			if err := json.Unmarshal(resp.Result, &block); err != nil {
				errChan <- fmt.Errorf("failed to unmarshal block details for block %d: %v", blockNumber, err)
				return
			}

			blockChan <- block
		}(blockNumber)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(blockChan)
		close(errChan)
	}()

	// Collect results and handle errors
	for block := range blockChan {
		blockDetails = append(blockDetails, block)
	}

	// If any errors occurred, return the first one
	if len(errChan) > 0 {
		return nil, <-errChan
	}

	return blockDetails, nil
}

//example usage use this to implement the logic in cli
//IGNORE PRINT STATEMENTS
// unc main() {
// 	url := "https://starknet-mainnet.public.blastapi.io/rpc/v0_7" // replace with your API endpoint
// 	fromBlockNumber := uint64(67800)
// 	toBlockNumber := uint64(67810)

// 	ctx := context.Background()
// 	blockDetails, err := GetBlockDetails(ctx, url, fromBlockNumber, toBlockNumber)
// 	if err != nil {
// 		log.Fatalf("Error fetching block details: %v", err)
// 	}

// 	// Print block details
// 	for _, block := range blockDetails {
// 		fmt.Printf("Block Hash: %v\n", block.BlockHash)
// 		fmt.Printf("Timestamp: %d\n", block.Timestamp)
// 		fmt.Printf("Sequencer Address: %s\n", block.SequencerAddress)
// 		fmt.Printf("L1 Gas Price: %s %s\n", block.L1GasPrice.PriceInWei, block.L1GasPrice.PriceInFri)
// 		fmt.Printf("Starknet Version: %s\n", block.StarknetVersion)
// 		fmt.Printf("L1 Data Gas Price: %s\n", block.L1DataGasPrice)
// 		fmt.Printf("L1 DA Mode: %s\n", block.L1DAMode)

// 		fmt.Println("Transactions:")
// 		for _, tx := range block.Transactions {
// 			fmt.Printf("  Transaction Hash: %s\n", tx.Hash)
// 			fmt.Printf("  Receipt Type: %s\n", tx.Receipt.Type)
// 			fmt.Printf("  Transaction Hash: %s\n", tx.Receipt.TransactionHash)
// 			fmt.Printf("  Actual Fee: %s %s\n", tx.Receipt.ActualFee.Amount, tx.Receipt.ActualFee.Unit)
// 			fmt.Printf("  Execution Status: %s\n", tx.Receipt.ExecutionStatus)
// 			fmt.Printf("  Finality Status: %s\n", tx.Receipt.FinalityStatus)

// 			fmt.Println("  Events:")
// 			for _, event := range tx.Receipt.Events {
// 				fmt.Printf("    From Address: %s\n", event.FromAddress)
// 				fmt.Printf("    Keys: %v\n", event.Keys)
// 				fmt.Printf("    Data: %v\n", event.Data)
// 			}

// 			fmt.Printf("  Execution Resources:\n")
// 			fmt.Printf("    Steps: %d\n", tx.Receipt.ExecutionResources.Steps)
// 			fmt.Printf("    Pedersen Built-in Applications: %d\n", tx.Receipt.ExecutionResources.PedersenBuiltinApplications)
// 			fmt.Printf("    Range Check Built-in Applications: %d\n", tx.Receipt.ExecutionResources.RangeCheckBuiltinApplications)
// 			fmt.Printf("    ECDSA Built-in Applications: %d\n", tx.Receipt.ExecutionResources.EcdsaBuiltinApplications)
// 		}
// 	}
// }
