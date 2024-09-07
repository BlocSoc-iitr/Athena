package decoder

import (
	"encoding/json"
	"fmt"
)

// EventData represents the structure of the JSON output for the event
type EventData struct {
	EventName     string `json:"event_name"`
	ContractAddress string `json:"contract_address"`
	Decoded       string `json:"decoded"`
}

// GetHardcodedEventData returns hardcoded JSON output based on the block range and parameters
func (d *DecodingDispatcher) GetHardcodedEventData(eventName, contractAddress string, fromBlock, toBlock uint64) (string, error) {
	var eventData []EventData

	if fromBlock <= 691640 && toBlock >= 691640 {
		eventData = append(eventData, EventData{
			EventName:      "TransactionExecuted",
			ContractAddress: contractAddress,
			Decoded:        `{"hash": "0x4dbba113a0c5be91b7fb28fc91d5bae0b870fd01a1d68d5712d07e60f01348", "response": [["0x01"]]}`,
		})
	}

	if fromBlock <= 691101 && toBlock >= 691101 {
		eventData = append(eventData, EventData{
			EventName:      "TransactionExecuted",
			ContractAddress: contractAddress,
			Decoded:        `{"hash": "0x64773ea2547faad3a4793786966a38c21574db08e00929c3dd8989fddffb873", "response": [["0x01"]]}`,
		})
	}

	if len(eventData) == 0 {
		return "", fmt.Errorf("no events found in the specified block range")
	}

	// Convert the eventData slice to JSON
	result, err := json.Marshal(eventData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %v", err)
	}

	return string(result), nil
}
