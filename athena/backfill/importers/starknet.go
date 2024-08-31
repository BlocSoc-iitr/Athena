package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
)

func main() {
	// Create a provider
	provider, err := rpc.NewProvider("https://free-rpc.nethermind.io/mainnet-juno/")
	if err != nil {
		log.Fatalf("Error creating provider: %v", err)
	}

	// Initialize blockNumber
	blockNumber := uint64(678000)
	blockId := rpc.BlockID{Number: &blockNumber} // Pass the address of blockNumber

	// Fetch transactions by block number
	transactions, error := provider.BlockWithTxHashes(context.Background(), blockId)
	if error != nil {
		log.Fatalf("Error fetching block data: %v", err)
	}

	// Handle the response based on its type
	switch transactionsType := transactions.(type) {
	case *rpc.BlockTxHashes:
		block := transactions.(*rpc.BlockTxHashes)
		fmt.Println("Block Hash:", block.BlockHash.String())
		fmt.Println("Number of Transactions:", len(block.Transactions))
		for _, txHash := range block.Transactions {
			fmt.Println("Transaction Hash:", txHash.String())
		}
	case *rpc.PendingBlockTxHashes:
		pBlock := transactions.(*rpc.PendingBlockTxHashes)
		fmt.Println("Pending Block Parent Hash:", pBlock.ParentHash.String())
		fmt.Println("Sequencer Address:", pBlock.SequencerAddress.String())
		for _, txHash := range pBlock.Transactions {
			fmt.Println("Pending Transaction Hash:", txHash.String())
		}
	default:
		log.Fatalf("Unexpected block type, found: %T\n", transactionsType)
	}

	hash_in_felt, err := utils.HexToFelt("0x11d0af90c13f9e9457fe2b9a9e76ec4750bdc542525ec644a1ccb747e139e74")
	if err != nil {
		log.Fatalf("Error converting hex to felt: %v", err)
	}

	txn_trace, err2 := provider.TraceTransaction(context.Background(), hash_in_felt)
	if err2 != nil {
		log.Fatalf("Error fetching transaction trace: %v", err2)
	}
	fmt.Println("Transaction Trace:", txn_trace)

	// Pretty-print the transaction trace
	txn_trace_json, err := json.MarshalIndent(txn_trace, "", "  ")
	if err != nil {
		log.Fatalf("Error formatting transaction trace: %v", err)
	}

	fmt.Println("Transaction Trace:\n", string(txn_trace_json))
	ToBlockNumber := uint64(678000)

	/// Fetching the transaction events
	blockIdTill := rpc.BlockID{Number: &ToBlockNumber}
	filter := rpc.EventFilter{
		FromBlock: blockId,
		ToBlock:   blockIdTill,
	}
	resultPage := rpc.ResultPageRequest{
		ChunkSize: 100,
	}

	Input := rpc.EventsInput{
		EventFilter:       filter,
		ResultPageRequest: resultPage,
	}

	events, error2 := provider.Events(context.Background(), Input)

	if error2 != nil {
		log.Fatalf("Error fetching events: %v", err)
	}
	fmt.Println("Events:", events)

	events_json, error3 := json.MarshalIndent(events, "", "  ")

	if error3 != nil {
		log.Fatalf("Error formatting events: %v", error3)
	}

	fmt.Println("Events:\n", string(events_json))

}
