package importers

import (
	"context"
	"fmt"
	"log"

	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
)

func Get_transactions_by_block(blockNumber *uint64) error {

	provider, error := rpc.NewProvider("https://free-rpc.nethermind.io/mainnet-juno/")
	if error != nil {
		return error
	}

	blockId := rpc.BlockID{Number: blockNumber}
	transactions, err := provider.BlockWithTxHashes(context.Background(), blockId)

	if err != nil {
		log.Fatalf("Error fetching block data: %v", err)
	}

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

	utils.BigNumberTo

	return nil

}
