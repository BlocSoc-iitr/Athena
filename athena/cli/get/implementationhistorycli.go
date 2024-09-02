package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
)

// BlockResult holds the block number and the corresponding class hash result.
type BlockResult struct {
	BlockNumber int64
	ClassHash   *felt.Felt
}

// AbiItem represents an item in an ABI.
type AbiItem struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// GetImplementedClass retrieves the class hash for a given contract address at a specific block.
func GetImplementedClass(provider *rpc.Provider, contractAddress string, BlockID uint64) (*felt.Felt, error) {
	contractAddressInFelt, _ := utils.HexToFelt(contractAddress)
	blockID := rpc.BlockID{Number: &BlockID}
	classHash, err := provider.ClassHashAt(context.Background(), blockID, contractAddressInFelt)
	if err != nil {
		return nil, fmt.Errorf("error fetching class hash: %v", err)
	}

	return classHash, nil
}

// FetchBlockData is a worker function that fetches data for blocks from the channel.
func FetchBlockData(provider *rpc.Provider, blocks <-chan int64, results chan<- BlockResult, errs chan<- error, wg *sync.WaitGroup, contractAddress string) {
	defer wg.Done()
	for block := range blocks {
		implClass, err := GetImplementedClass(provider, contractAddress, uint64(block))
		if err != nil {
			errs <- err
			continue
		}
		results <- BlockResult{BlockNumber: block, ClassHash: implClass}
	}
}

// FetchClassAbi fetches and saves the ABI for a given class hash.
func FetchClassAbi(provider *rpc.Provider, classHash *felt.Felt) ([]AbiItem, error) {
	// Placeholder for ABI fetching logic.
	return nil, fmt.Errorf("ClassDefinition method is not available in the provided library")
}

func main() {
	// Command-line flags
	contractAddress := flag.String("contract", "", "Contract address to query.")
	fromBlock := flag.Int64("from", 0, "The block number to start from.")
	toBlock := flag.Int64("to", 0, "The block number to end at.")
	rpcUrl := flag.String("rpc", "https://starknet-mainnet.public.blastapi.io", "RPC provider URL.")
	fetchAbi := flag.Bool("fetchabi", false, "Fetch and save ABI for class hashes found.")
	flag.Parse()

	// Input validation
	if *contractAddress == "" || *fromBlock == 0 || *toBlock == 0 {
		log.Fatalf("Please provide valid contract address, fromBlock, and toBlock.")
	}

	// Provider setup
	provider, err := rpc.NewProvider(*rpcUrl)
	if err != nil {
		log.Fatalf("Error creating provider: %v", err)
	}

	blockRange := *fromBlock - *toBlock + 1
	if blockRange <= 0 {
		log.Fatalf("Invalid block range: fromBlock must be greater than toBlock.")
	}

	blocks := make(chan int64, blockRange)
	results := make(chan BlockResult, blockRange)
	errs := make(chan error, blockRange)

	const numWorkers = 50 // Number of worker goroutines

	var wg sync.WaitGroup

	// Create worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go FetchBlockData(provider, blocks, results, errs, &wg, *contractAddress)
	}

	// Send blocks to be processed
	for block := *fromBlock; block >= *toBlock; block-- {
		blocks <- block
	}
	close(blocks)

	// Wait for all goroutines to finish
	wg.Wait()
	close(results)
	close(errs)

	// To store the latest block number for each unique class hash
	classHashMap := make(map[string]int64)

	for result := range results {
		hashStr := result.ClassHash.String()
		// If the class hash is not in the map or if the current block is later, update the map
		if block, exists := classHashMap[hashStr]; !exists || result.BlockNumber > block {
			classHashMap[hashStr] = result.BlockNumber
		}
	}

	for err := range errs {
		fmt.Printf("Error fetching block data: %v\n", err)
	}

	// Process and optionally fetch ABI
	for hash, block := range classHashMap {
		fmt.Printf("Class hash: %s has the latest implementation at block %d\n", hash, block)
		if *fetchAbi {
			classHash, _ := felt.NewFelt(hash) // Use felt.NewFelt with single return value and remove the dereference
			abi, err := FetchClassAbi(provider, classHash)
			if err != nil {
				fmt.Printf("Error fetching ABI for class hash %s: %v\n", hash, err)
			} else {
				fmt.Printf("Fetched ABI for class hash %s: %+v\n", hash, abi)
			}
		}
	}
}
