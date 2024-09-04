package importers

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
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

type BlockData struct {
	ParentHash       string     `json:"parent_hash"`
	Timestamp        int64      `json:"timestamp"`
	SequencerAddress string     `json:"sequencer_address"`
	L1GasPrice       L1GasPrice `json:"l1_gas_price"`
	StarknetVersion  string     `json:"starknet_version"`
	L1DataGasPrice   L1GasPrice `json:"l1_data_gas_price"`
	L1DAMode         string     `json:"l1_da_mode"`
	BlockHash        string     `json:"block_hash"`
}

type BlockTxHashes struct {
	ParentHash       string     `json:"parent_hash"`
	Timestamp        int64      `json:"timestamp"`
	SequencerAddress string     `json:"sequencer_address"`
	L1GasPrice       L1GasPrice `json:"l1_gas_price"`
	StarknetVersion  string     `json:"starknet_version"`
	L1DataGasPrice   L1GasPrice `json:"l1_data_gas_price"`
	L1DAMode         string     `json:"l1_da_mode"`
	BlockHash        *string    `json:"block_hash"` // Pointer to detect nil (pending blocks)
	Transactions     []struct {
		Transaction struct {
			Hash      string   `json:"transaction_hash"`
			Version   string   `json:"version"`
			Nonce     string   `json:"nonce"`
			Calldata  []string `json:"calldata"`
			Signature []string `json:"signature"`
		} `json:"transaction"`
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

func MakeRPCCall(ctx context.Context, url string, method string, params interface{}) (*rpcResponse, error) {
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

func GetBlockHashDetails(ctx context.Context, url string, fromBlockNumber uint64, toBlockNumber uint64) ([]BlockData, error) {
	var blockDetails []BlockData
	var wg sync.WaitGroup
	blockChan := make(chan BlockData, toBlockNumber-fromBlockNumber+1)
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

			resp, err := MakeRPCCall(ctx, url, "starknet_getBlockWithTxHashes", params)
			if err != nil {
				errChan <- fmt.Errorf("failed to get block details for block %d: %v", blockNumber, err)
				return
			}

			var block BlockData
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

func GetBlockDetails(ctx context.Context, url string, fromBlockNumber uint64, toBlock uint64) (*[]BlockTxHashes, error) {
	var blockDetails []BlockTxHashes
	var wg sync.WaitGroup
	blockChan := make(chan BlockTxHashes, toBlock-fromBlockNumber+1)
	errChan := make(chan error, toBlock-fromBlockNumber+1)

	// Loop through each block number
	for blockNumber := fromBlockNumber; blockNumber <= toBlock; blockNumber++ {
		wg.Add(1)
		go func(blockNumber uint64) {
			defer wg.Done()

			params := map[string]interface{}{
				"block_id": map[string]interface{}{
					"block_number": int(blockNumber),
				},
			}
			resp, err := MakeRPCCall(ctx, url, "starknet_getBlockWithReceipts", params)
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

	// Close channels after all goroutines have finished
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

	return &blockDetails, nil
}

func WriteBlockHashesToCSV(blockDetails []BlockData, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header
	header := []string{
		"Block Hash", "Parent Hash", "Timestamp", "Sequencer Address", "L1 Gas Price (Wei)", "L1 Gas Price (Fri)",
		"Starknet Version", "L1 Data Gas Price (Wei)", "L1 Data Gas Price (Fri)", "L1 DA Mode",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	// Write the data
	for _, block := range blockDetails {
		record := []string{
			block.BlockHash,
			block.ParentHash,
			strconv.FormatInt(block.Timestamp, 10),
			block.SequencerAddress,
			block.L1GasPrice.PriceInWei,
			block.L1GasPrice.PriceInFri,
			block.StarknetVersion,
			block.L1DataGasPrice.PriceInWei,
			block.L1DataGasPrice.PriceInFri,
			block.L1DAMode,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %v", err)
		}
	}

	return nil
}

func WriteBlockDetailsToCSV(blockDetails []BlockTxHashes, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header with all relevant fields
	header := []string{
		"Transaction Hash", "Version", "Nonce", "Calldata", "Signature",
		"Receipt Type", "Transaction Hash", "Actual Fee Amount", "Actual Fee Unit", "Execution Status",
		"Finality Status", "Event From Address", "Event Keys", "Event Data",
		"Steps", "Pedersen Builtin Applications", "Range Check Builtin Applications", "ECDSA Builtin Applications",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header to CSV file: %w", err)
	}

	// Write block and transaction details
	for _, block := range blockDetails {
		for _, tx := range block.Transactions {
			// Flatten the events into strings
			events := flattenEvents(tx.Receipt.Events)

			record := []string{
				tx.Transaction.Hash,
				tx.Transaction.Version,
				tx.Transaction.Nonce,
				fmt.Sprintf("%v", tx.Transaction.Calldata),
				fmt.Sprintf("%v", tx.Transaction.Signature),
				tx.Receipt.Type,
				tx.Receipt.TransactionHash,
				tx.Receipt.ActualFee.Amount,
				tx.Receipt.ActualFee.Unit,
				tx.Receipt.ExecutionStatus,
				tx.Receipt.FinalityStatus,
				events.FromAddress,
				events.Keys,
				events.Data,
				fmt.Sprintf("%d", tx.Receipt.ExecutionResources.Steps),
				fmt.Sprintf("%d", tx.Receipt.ExecutionResources.PedersenBuiltinApplications),
				fmt.Sprintf("%d", tx.Receipt.ExecutionResources.RangeCheckBuiltinApplications),
				fmt.Sprintf("%d", tx.Receipt.ExecutionResources.EcdsaBuiltinApplications),
			}

			if err := writer.Write(record); err != nil {
				return fmt.Errorf("failed to write record to CSV file: %w", err)
			}
		}
	}

	return nil
}

func flattenEvents(events []struct {
	FromAddress string   `json:"from_address"`
	Keys        []string `json:"keys"`
	Data        []string `json:"data"`
}) (flattened struct {
	FromAddress string
	Keys        string
	Data        string
}) {
	for _, event := range events {
		flattened.FromAddress += event.FromAddress + ";"
		flattened.Keys += fmt.Sprintf("%v", event.Keys) + ";"
		flattened.Data += fmt.Sprintf("%v", event.Data) + ";"
	}

	// Remove the last semicolon
	if len(flattened.FromAddress) > 0 {
		flattened.FromAddress = flattened.FromAddress[:len(flattened.FromAddress)-1]
	}
	if len(flattened.Keys) > 0 {
		flattened.Keys = flattened.Keys[:len(flattened.Keys)-1]
	}
	if len(flattened.Data) > 0 {
		flattened.Data = flattened.Data[:len(flattened.Data)-1]
	}

	return
}
