package importers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/NethermindEth/starknet.go/rpc"
)

func FetchEvents(provider *rpc.Provider, filter rpc.EventFilter, resultPage rpc.ResultPageRequest) ([]rpc.EventChunk, error) {
	var allEvents []rpc.EventChunk

	for {
		// Create the input for the events request
		Input := rpc.EventsInput{
			EventFilter:       filter,
			ResultPageRequest: resultPage,
		}

		// Fetch events
		eventsResponse, err := provider.Events(context.Background(), Input)
		if err != nil {
			return nil, fmt.Errorf("error fetching events: %w", err)
		}

		// Dereference eventsResponse and append the fetched events to the list
		if eventsResponse != nil {
			allEvents = append(allEvents, *eventsResponse)

		} else {
			return nil, fmt.Errorf("received nil eventsResponse")
		}

		// Check if there's a continuation token for the next pages
		if eventsResponse.ContinuationToken == "" {
			break // No more pages, exit the loop
		}

		// Update the result page with the continuation token for the next request
		resultPage.ContinuationToken = eventsResponse.ContinuationToken
	}

	return allEvents, nil
}

// ExportEventsToCSV exports a slice of events to a CSV file.
func ExportEventsToCSV(events []rpc.EventChunk, filename string) error {
	// Create a CSV file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header to the CSV file
	header := []string{"Block Number", "Transaction Hash", "Event Index", "Data", "Keys"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header to CSV file: %w", err)
	}

	// Write the event data to the CSV file
	for _, event := range events {
		data := event.Events
		for _, d := range data {
			var keystore []string
			for _, key := range d.Event.Keys {
				keys := fmt.Sprintf("%s", key)
				keystore = append(keystore, keys)
			}
			keustore := strings.Join(keystore, ",")
			var datastore []string
			for _, value := range d.Event.Data {
				data := fmt.Sprintf("%s", value)
				datastore = append(datastore, data)
			}
			dayastore := strings.Join(datastore, ",")

			record := []string{
				fmt.Sprintf("%d", d.BlockNumber),
				fmt.Sprintf("%d", d.TransactionHash),
				fmt.Sprintf("%d", d.BlockHash),
				fmt.Sprintf("%d", d.Event.FromAddress),
				string(keustore),
				string(dayastore),
			}

			if err := writer.Write(record); err != nil {
				return fmt.Errorf("failed to write record to CSV file: %w", err)
			}
		}

	}
	return nil
}

// example usage remove this after implementing it in filters and in cli
func main() {

	provider, _ := rpc.NewProvider("https://starknet-mainnet.public.blastapi.io/rpc/v0_7")
	// Define the block numbers
	FromBlockNumber := uint64(67800) // Starting block number
	ToBlockNumber := uint64(67801)   // Ending block number

	// Initialize the block IDs
	blockId := rpc.BlockID{Number: &FromBlockNumber}
	blockIdTill := rpc.BlockID{Number: &ToBlockNumber}

	// Define the filter and result page request
	filter := rpc.EventFilter{
		FromBlock: blockId,
		ToBlock:   blockIdTill,
	}

	resultPage := rpc.ResultPageRequest{
		ChunkSize: 100, // Fetch 100 events per request
	}

	// Fetch all events
	events, err := FetchEvents(provider, filter, resultPage)
	if err != nil {
		log.Fatalf("Error fetching events: %v", err)
	}

	// Convert events to JSON
	eventsJSON, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		log.Fatalf("Error formatting events: %v", err)
	}

	// Print events
	fmt.Println("Events:\n", string(eventsJSON))

	// Export to CSV
	err = ExportEventsToCSV(events, "events_1.csv")
	if err != nil {
		log.Fatalf("Error exporting events to CSV: %v", err)
	}
}
