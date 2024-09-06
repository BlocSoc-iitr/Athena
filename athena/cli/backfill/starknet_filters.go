package main

import (
	"flag"
	"fmt"
	"strings"
	"github.com/BlocSoc-iitr/Athena/athena/backfill"
	
)

func main() {
	// Define CLI flags
	providerName := flag.String("rpc-url", "", "RPC URL of the blockchain node")
	contractAddress := flag.String("contract-address", "", "Contract address to filter events by")
	hexKeys := flag.String("keys", "", "Comma-separated list of hex keys to filter events by")
	fromBlockNumber := flag.Uint64("from", 0, "Starting block number")
	toBlockNumber := flag.Uint64("to", 0, "Ending block number")
	outputFile := flag.String("output", "events.csv", "Output CSV file")
	
	flag.Parse()

	// Validate required flags
	if *providerName == "" || (*contractAddress == "" && *hexKeys == "") || *fromBlockNumber == 0 || *toBlockNumber == 0 {
		fmt.Println("Missing required flags. Use --help for usage.")
		return
	}

	if *contractAddress != "" {
		// Filter events by contract address
		backfill.FilterEventsByContractAddress(*providerName, *contractAddress, *fromBlockNumber, *toBlockNumber, *outputFile)
	} else if *hexKeys != "" {
		// Filter events by hex keys
		hexKeyStrings := strings.Split(*hexKeys, ",")
		backfill.FilterEventsByHexKeyString(*providerName, hexKeyStrings, *fromBlockNumber, *toBlockNumber, *outputFile)
	}
	fmt.Printf("Events filtered exported to %s\n", *outputFile)
}
