package main
import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
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

func filterEventsByContractAddress(providername string, contractAddress string, FromBlockNumber uint64, ToBlockNumber uint64, filename string) {
	provider, err := rpc.NewProvider(providername)
	if err != nil {
		log.Fatalf("Error creating provider: %v", err)
	}

	blockId := rpc.BlockID{Number: &FromBlockNumber}
	blockIdTill := rpc.BlockID{Number: &ToBlockNumber}

	contract, err := utils.HexToFelt(contractAddress)

	filter := rpc.EventFilter{
		FromBlock: blockId,
		ToBlock:   blockIdTill,
		Address:   contract,
	}

	resultPage := rpc.ResultPageRequest{
		ChunkSize: 100, // Fetch 100 events per request
	}

	// Fetch all events
	events, err := FetchEvents(provider, filter, resultPage)
	if err != nil {
		log.Fatalf("Error fetching events: %v", err)
	}

	// Export filtered events to CSV
	err = ExportEventsToCSV(events, filename)
	if err != nil {
		log.Fatalf("Error exporting events to CSV: %v", err)
	}

}

func filterEventsByHexKeyString(providername string, hexKeyString []string, FromBlockNumber uint64, ToBlockNumber uint64, filename string) {
	provider, err := rpc.NewProvider(providername)
	if err != nil {
		log.Fatalf("Error creating provider: %v", err)
	}

	blockId := rpc.BlockID{Number: &FromBlockNumber}
	blockIdTill := rpc.BlockID{Number: &ToBlockNumber}

	var events_Key []*felt.Felt
	for i := 0; i < len(hexKeyString); i++ {
		hexKey, err := utils.HexToFelt(hexKeyString[i])
		if err != nil {
			log.Fatalf("Error converting hex to felt: %v", err)
		}
		events_Key = append(events_Key, hexKey)
	}
	var final_events [][]*felt.Felt
	final_events = append(final_events, events_Key)

	filter := rpc.EventFilter{
		FromBlock: blockId,
		ToBlock:   blockIdTill,
		Keys:      final_events,
	}

	resultPage := rpc.ResultPageRequest{
		ChunkSize: 100, // Fetch 100 events per request
	}

	// Fetch all events
	events, err := FetchEvents(provider, filter, resultPage)
	if err != nil {
		log.Fatalf("Error fetching events: %v", err)
	}

	// Export filtered events to CSV
	err = ExportEventsToCSV(events, filename)
	if err != nil {
		log.Fatalf("Error exporting events to CSV: %v", err)
	}
}

func main() {
	provider := "https://rpc.nethermind.io/mainnet-juno?x-apikey=MIkLH4AOTdTH9uqu8PqvSHUBNnAnMU1fXdROa3qc1DsSVxvOcGRrwr6kSj1zsNjT"
	// Define the block numbers
	FromBlockNumber := uint64(691575) // Starting block number
	ToBlockNumber := uint64(691600)

	hexstring := []string{"0x99cd8bde557814842a3121e8ddfd433a539b8c9f14bf31ebf108d12e6196e9"}
	filterEventsByHexKeyString(provider, hexstring, FromBlockNumber, ToBlockNumber, "events_filtered_1.csv")

}
