package main

import (
	"context"
	"fmt"
	"log"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
)

func GetImplementedClass(contractAddress string, BlockID uint64) (*felt.Felt, error) {
	contractAddressInFelt, _ := utils.HexToFelt(contractAddress)
	fmt.Printf("Contract Address in Felt: %v\n", contractAddressInFelt)
	provider, err := rpc.NewProvider("https://free-rpc.nethermind.io/mainnet-juno/")
	if err != nil {
		return nil, fmt.Errorf("error creating provider: %v", err)
	}

	blockID := rpc.BlockID{Number: &BlockID}

	classHash, err := provider.ClassHashAt(context.Background(), blockID, contractAddressInFelt)
	if err != nil {
		return nil, fmt.Errorf("error fetching class hash: %v", err)
	}

	return classHash, nil
}

func isProxy(classHash *felt.Felt) (bool, string) {
	// Implement the logic to determine if the class is a proxy
	// For now, we’ll mock this function
	// This would involve decoding the class and checking if it has a proxy method
	// Return true if it’s a proxy and the selector of the proxy method
	return true, "proxyMethodSelector" // mock proxy method selector
}

func getProxyImplHistory(contractAddress string, fromBlock, toBlock int64, proxyMethod string) map[string]string {
	// Implement the logic to fetch proxy implementation history
	// This function would involve calling the appropriate RPC method and returning a map
	// For now, we’ll mock this function
	return map[string]string{
		"proxy_from_block": fmt.Sprintf("fromBlock: %d", fromBlock),
		"proxy_to_block":   fmt.Sprintf("toBlock: %d", toBlock),
		"proxy_method":     proxyMethod,
	}
}

func fetchImplementationHistory(fromBlock int64, toBlock int64) {
	contractImplHistory := make(map[string]interface{})
	var oldClass *felt.Felt

	for currentBlock := fromBlock; currentBlock >= toBlock; currentBlock-- {
		implClass, err := GetImplementedClass("0x03b207d9237a3b6354078a3b4ba3c41e925913dd83f9deb30c94a80c1bf619ba", uint64(currentBlock))
		if err != nil {
			fmt.Printf("Error getting implemented class: %v at block: %d\n", err, currentBlock)
			continue
		}

		isProxy, proxyMethod := isProxy(implClass)

		if !isProxy {
			contractImplHistory[fmt.Sprintf("%d", currentBlock)] = implClass
			oldClass = implClass
			continue
		}

		var proxyToBlock int64
		if currentBlock == toBlock {
			proxyToBlock = toBlock
		} else {
			proxyToBlock = currentBlock - 1 // 1 block before the next root class is deployed
		}

		proxyImplHistory := getProxyImplHistory("0x03b207d9237a3b6354078a3b4ba3c41e925913dd83f9deb30c94a80c1bf619ba", currentBlock, proxyToBlock, proxyMethod)

		contractImplHistory[fmt.Sprintf("%d", currentBlock)] = map[string]interface{}{
			"proxy_class": implClass,
			"proxy_info":  proxyImplHistory,
		}

		oldClass = implClass
	}

	log.Println("Contract Implementation History:", contractImplHistory)
}

func main() {
	fromBlock := int64(151001)
	toBlock := int64(131000)
	fetchImplementationHistory(fromBlock, toBlock)
}
