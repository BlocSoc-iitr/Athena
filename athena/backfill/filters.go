package backfill

import (
	"log"

	"github.com/DarkLord017/athena/athena/backfill/importers"
	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
)

func FilterEventsByContractAddress(providername string, contractAddress string, FromBlockNumber uint64, ToBlockNumber uint64, filename string) {
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
	events, err := importers.FetchEvents(provider, filter, resultPage)
	if err != nil {
		log.Fatalf("Error fetching events: %v", err)
	}

	// Export filtered events to CSV
	err = importers.ExportEventsToCSV(events, filename)
	if err != nil {
		log.Fatalf("Error exporting events to CSV: %v", err)
	}

}

func FilterEventsByHexKeyString(providername string, hexKeyString []string, FromBlockNumber uint64, ToBlockNumber uint64, filename string) {
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
	events, err := importers.FetchEvents(provider, filter, resultPage)
	if err != nil {
		log.Fatalf("Error fetching events: %v", err)
	}

	// Export filtered events to CSV
	err = importers.ExportEventsToCSV(events, filename)
	if err != nil {
		log.Fatalf("Error exporting events to CSV: %v", err)
	}
}
