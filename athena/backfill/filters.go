package ma

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/DarkLord017/athena/athena/types"
	"github.com/DarkLord017/athena/athena/decoder" // Replace with actual module path
)

const rpcURL = "https://starknet-mainnet.public.blastapi.io/rpc/v0_7" // Replace with actual StarkNet RPC URL

type EventFilter struct {
	FromBlock string   `json:"from_block,omitempty"`
	ToBlock   string   `json:"to_block,omitempty"`
	Address   string   `json:"address,omitempty"`
	Keys      []string `json:"keys,omitempty"`
	ChunkSize int      `json:"chunk_size,omitempty"`
	PageToken string   `json:"page_token,omitempty"`
}

type GetEventsRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  EventFilter `json:"params"`
	ID      int         `json:"id"`
}

type GetEventsResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		Events            []Event `json:"events"`
		ContinuationToken string  `json:"continuation_token,omitempty"`
	} `json:"result"`
}

type Event struct {
	BlockHash       string   `json:"block_hash"`
	BlockNumber     int      `json:"block_number"`
	TransactionHash string   `json:"transaction_hash"`
	FromAddress     string   `json:"from_address"`
	Keys            []string `json:"keys"`
	Data            []string `json:"data"`
}

func loadABI(filename string, abiName string, additionalData []byte) (decoder.ABI, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return decoder.ABI{}, fmt.Errorf("failed to open ABI file: %v", err)
	}

	abi, err := decoder.FromJSON(file, abiName, additionalData)
	if err != nil {
		return decoder.ABI{}, fmt.Errorf("failed to decode ABI: %v", err)
	}

	return abi, nil
}

func fetchEvents(filter EventFilter) (*GetEventsResponse, error) {
	requestBody := GetEventsRequest{
		Jsonrpc: "2.0",
		Method:  "starknet_getEvents",
		Params:  filter,
		ID:      1,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(rpcURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var eventsResponse GetEventsResponse
	if err := json.Unmarshal(responseBody, &eventsResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &eventsResponse, nil
}

func writeEventsToCSV(events []Event, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header
	header := []string{
		"Block Hash", "Block Number", "Transaction Hash", "From Address", "Event Keys", "Event Data",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	// Write the data
	for _, event := range events {
		record := []string{
			event.BlockHash,
			fmt.Sprintf("%d", event.BlockNumber),
			event.TransactionHash,
			event.FromAddress,
			strings.Join(event.Keys, ", "), // Convert keys slice to string
			strings.Join(event.Data, ", "), // Convert data slice to string
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %v", err)
		}
	}

	return nil
}

// GenerateTopics generates event topics from ABI and event names
func GenerateTopics(abi decoder.ABI, eventNames []string) ([]string, [][]string, error) {
	var selectors []string

	for _, eventName := range eventNames {
		if event, exists := abi.Events[eventName]; exists {
			// Assuming the event name is used directly for selector generation
			selector := fmt.Sprintf("0x%s", strings.ToLower(hex.EncodeToString([]byte(eventName))))
			selectors = append(selectors, selector)
		}
	}

	topics := [][]string{selectors}

	return eventNames, topics, nil
}

func main() {
	// Load ABI from file
	abiName := "YourABIName" // Adjust as needed
	abi, err := loadABI("contract_abi.json", abiName, nil)
	if err != nil {
		fmt.Printf("Error loading ABI: %v\n", err)
		return
	}

	// Define the filter
	filter := EventFilter{
		FromBlock: "0xA6040",                                                            // Start from block number 256
		ToBlock:   "0xA788F",                                                            // End at block number 512
		Address:   "0x004e05ea122a7a06a287d3679a2282c95819448a5a1d55778bcbd666ec9081dc", // Replace with the contract address you're interested in
		ChunkSize: 100,                                                                  // Fetch 100 events per request
	}

	// Generate topics from ABI
	eventNames := []string{"IAccount", "IUpgradableCallback"} // Replace with actual event names
	_, topics, err := GenerateTopics(abi, eventNames)
	if err != nil {
		fmt.Printf("Error generating topics: %v\n", err)
		return
	}

	// Add generated topics to the filter
	filter.Keys = topics[0]

	var allEvents []Event
	continuationToken := ""

	for {
		filter.PageToken = continuationToken
		eventsResponse, err := fetchEvents(filter)
		if err != nil {
			fmt.Printf("Error fetching events: %v\n", err)
			return
		}

		allEvents = append(allEvents, eventsResponse.Result.Events...)

		if eventsResponse.Result.ContinuationToken == "" {
			break
		}

		continuationToken = eventsResponse.Result.ContinuationToken
	}

	// Export all events to a CSV file
	if err := writeEventsToCSV(allEvents, "starknet_events.csv"); err != nil {
		fmt.Printf("Error writing events to CSV: %v\n", err)
		return
	}

	fmt.Printf("Exported %d events to starknet_events.csv\n", len(allEvents))
}
