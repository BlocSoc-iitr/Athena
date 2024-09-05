package main
import (
	
	"context"
	
	"encoding/json"
	"flag"
	"fmt"
	"sync"
	"github.com/BlocSoc-iitr/Athena/athena/backfill/importers"
	
)

type rpcRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

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
	BlockHash        *string    `json:"block_hash"`
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
func main() {
	
	fromBlockNumber := flag.Uint64("from", 0, "Starting block number")
	toBlockNumber := flag.Uint64("to", 0, "Ending block number")
	rpcURL := flag.String("rpc-url", "", "RPC URL of the blockchain node")
	outputFile := flag.String("output", "block_details.csv", "Output CSV file")
	transactionHashFlag := flag.Bool("transactionhash", false, "Fetch transaction hashes as well")

	flag.Parse()

	if *fromBlockNumber == 0 || *toBlockNumber == 0 || *rpcURL == "" {
		fmt.Println("Missing required flags. Use --help for usage.")
		return
	}

	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		blockDetails, err := importers.GetBlockHashDetails(ctx, *rpcURL, *fromBlockNumber, *toBlockNumber)
		if err != nil {
			fmt.Printf("Failed to fetch block details: %v\n", err)
			return
		}

		if err := importers.WriteBlockHashesToCSV(blockDetails, *outputFile); err != nil {
			fmt.Printf("Failed to write block details to CSV: %v\n", err)
			return
		}

		fmt.Printf("Block details written to %s\n", *outputFile)
	}()

	if *transactionHashFlag {
		wg.Add(1)

		go func() {
			defer wg.Done()

			blockTxHashesDetails, err := importers.GetBlockDetails(ctx, *rpcURL, *fromBlockNumber, *toBlockNumber)
			if err != nil {
				fmt.Printf("Failed to fetch block transaction hashes details: %v\n", err)
				return
			}

			transactionOutputFile := "transaction_hashes_" + *outputFile
			if err := importers.WriteBlockDetailsToCSV(*blockTxHashesDetails, transactionOutputFile); err != nil {
				fmt.Printf("Failed to write block transaction hashes details to CSV: %v\n", err)
				return
			}

			fmt.Printf("Block transaction hashes details written to %s\n", transactionOutputFile)
		}()
	}

	wg.Wait()
}
