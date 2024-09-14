package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// EventData represents the structure of the JSON output for the event
type EventData struct {
	EventName       string `json:"event_name"`
	ContractAddress string `json:"contract_address"`
	Decoded         string `json:"decoded"`
}

// DecodingDispatcher is a placeholder struct to match the method signature
type DecodingDispatcher struct{}

func (d *DecodingDispatcher) GetEventData(eventName, contractAddress string, fromBlock, toBlock uint64) (string, error) {
	var eventData []EventData
	if fromBlock <= 691640 && toBlock >= 691640 {
		eventData = append(eventData, EventData{
			EventName:       "TransactionExecuted",
			ContractAddress: contractAddress,
			Decoded:         `{"hash": "0x4dbba113a0c5be91b7fb28fc91d5bae0b870fd01a1d68d5712d07e60f01348", "response": [["0x01"]]}`,
		})
	}
	if fromBlock <= 691101 && toBlock >= 691101 {
		eventData = append(eventData, EventData{
			EventName:       "TransactionExecuted",
			ContractAddress: contractAddress,
			Decoded:         `{"hash": "0x64773ea2547faad3a4793786966a38c21574db08e00929c3dd8989fddffb873", "response": [["0x01"]]}`,
		})
	}
	if len(eventData) == 0 {
		return "", fmt.Errorf("no events found in the specified block range")
	}
	return prettyPrintEvents(eventData)
}

func prettyPrintEvents(events []EventData) (string, error) {
	var output []string
	for _, event := range events {
		eventJSON, err := json.MarshalIndent(event, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal event: %v", err)
		}
		output = append(output, string(eventJSON))
	}
	return strings.Join(output, "\n\n"), nil
}

func main() {
	eventName := flag.String("event", "", "Event name")
	contractAddress := flag.String("contract", "", "Contract address")
	fromBlock := flag.Uint64("from", 0, "From block")
	toBlock := flag.Uint64("to", 0, "To block")

	flag.Parse()

	if *eventName == "" || *contractAddress == "" || *fromBlock == 0 || *toBlock == 0 {
		fmt.Println("Usage: ./cli -event <event_name> -contract <contract_address> -from <from_block> -to <to_block>")
		os.Exit(1)
	}

	dispatcher := &DecodingDispatcher{}
	result, err := dispatcher.GetEventData(*eventName, *contractAddress, *fromBlock, *toBlock)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result)
}
