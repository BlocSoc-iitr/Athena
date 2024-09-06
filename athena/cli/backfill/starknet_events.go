package main

import (
	
	"flag"
	"fmt"
	"log"
	"github.com/BlocSoc-iitr/Athena/athena/backfill/importers"
	"github.com/NethermindEth/starknet.go/rpc"
)

func main() {
	// Define CLI flags
	fromBlockNumber := flag.Uint64("from", 0, "Starting block number")
	toBlockNumber := flag.Uint64("to", 0, "Ending block number")
	rpcURL := flag.String("rpc-url", "", "RPC URL of the blockchain node")
	outputFile := flag.String("output", "events.csv", "Output CSV file")
	ChunkSize := flag.Int("chunk-size", 100, "Number of events per request")

	flag.Parse()

	// Validate required flags
	if *fromBlockNumber == 0 || *toBlockNumber == 0 || *rpcURL == "" {
		fmt.Println("Missing required flags. Use --help for usage.")
		return
	}

	// Initialize RPC provider
	provider, err := rpc.NewProvider(*rpcURL)
	if err != nil {
		log.Fatalf("Failed to create RPC provider: %v", err)
	}

	// Define the block IDs
	fromBlockID := rpc.BlockID{Number: fromBlockNumber}
	toBlockID := rpc.BlockID{Number: toBlockNumber}

	// Set up the filter and pagination
	filter := rpc.EventFilter{
		FromBlock: fromBlockID,
		ToBlock:   toBlockID,
	}

	resultPage := rpc.ResultPageRequest{
		ChunkSize: *ChunkSize,
	}

	// Fetch events
	events, err := importers.FetchEvents(provider, filter, resultPage)
	if err != nil {
		log.Fatalf("Error fetching events: %v", err)
	}

	// Export events to CSV
	err = importers.ExportEventsToCSV(events, *outputFile)
	if err != nil {
		log.Fatalf("Error exporting events to CSV: %v", err)
	}

	fmt.Printf("Events exported to %s\n", *outputFile)
}
