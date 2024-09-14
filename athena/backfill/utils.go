package backfill

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type GraceFullkiller struct {
	killNow bool
}

type Network string

const (
	Ethereum Network = "ethereum"
	Starknet Network = "starknet"
)

type BlockIdentifier string

const (
	latest    BlockIdentifier = "latest"
	earliest  BlockIdentifier = "earliest"
	safe      BlockIdentifier = "safe"
	finalized BlockIdentifier = "finalized"
	pending   BlockIdentifier = "pending"
)

func Default_rpc(network Network) (string, error) {
	switch network {
	case Ethereum:
		return "https://eth.public-rpc.com/", nil
	case Starknet:
		return "https://free-rpc.nethermind.io/mainnet-juno/", nil
	default:
		return "", fmt.Errorf("Network not supported")
	}
}

func Etherscan_base_url(network Network) (string, error) {
	switch network {
	case Ethereum:
		return "https://api.etherscan.io/api", nil
	default:
		return "", fmt.Errorf("Network not available from etherscan")
	}
}

func Get_current_block_number(network Network) (int, error) {
	switch network {
	case Starknet:
		rpc := os.Getenv("JSON_RPC")
		if rpc == "" {
			rpc, _ = Default_rpc(network)
		}

		client := &http.Client{Timeout: 30 * time.Second}
		reqBody := map[string]interface{}{
			"id":      1,
			"jsonrpc": "2.0",
			"method":  "starknet_blockNumber",
		}
		body, err := json.Marshal(reqBody)
		if err != nil {
			return 0, err
		}

		resp, err := client.Post(rpc, "application/json", io.NopCloser(bytes.NewReader(body)))
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()

		var response map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return 0, err
		}

		result, ok := response["result"].(float64)
		if !ok {
			return 0, fmt.Errorf("error fetching current block number for Starknet: %v", response)
		}

		return int(result), nil
	case Ethereum:
		rpc := os.Getenv("JSON_RPC")
		if rpc == "" {
			rpc, _ = Default_rpc(network)
		}

		client := &http.Client{Timeout: 30 * time.Second}
		reqBody := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      0,
			"method":  "eth_blockNumber",
		}
		body, err := json.Marshal(reqBody)
		if err != nil {
			return 0, err
		}

		resp, err := client.Post(rpc, "application/json", bytes.NewReader(body))
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()

		var response map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return 0, err
		}

		resultHex, ok := response["result"].(string)
		if !ok {
			return 0, fmt.Errorf("error fetching current block number for Ethereum: %v", response)
		}

		blockNumber, err := strconv.ParseInt(resultHex, 16, 64)
		if err != nil {
			return 0, fmt.Errorf("error converting block number: %v", err)
		}

		return int(blockNumber), nil

	default:
		return 0, fmt.Errorf("Network not supported")
	}

}

func Block_Identifier_To_Block(identifier BlockIdentifier, network Network) (int, error) {
	switch identifier {
	case latest:
		return Get_current_block_number(network)
	case earliest:
		return 0, nil
	case safe:
		return 0, fmt.Errorf("not implemented")
	case finalized:
		return 0, fmt.Errorf("not implemented")
	case pending:
		blockNumber, err := Get_current_block_number(network)
		if err != nil {
			return 0, fmt.Errorf("error in getting a block number")
		}
		return blockNumber + 1, nil
	default:
		return 0, fmt.Errorf("block Identifier not supported")
	}
}

func New_Gracfull_Killer() *GraceFullkiller {
	killer := &GraceFullkiller{killNow: false}
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signalChannel
		log.Printf("Received signal: %v", sig)

		// Gracefully kill the process
		killer.killNow = true
	}()
	return killer
}

func (g *GraceFullkiller) KillNow() bool {
	return g.killNow
}
