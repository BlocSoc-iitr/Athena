package decoder

import (
	"errors"
	"fmt"
	"github.com/BlocSoc-iitr/Athena"
	
	"github.com/olekukonko/tablewriter"
	"log"
	"strings"
)

type DecodingDispatcher struct {
	//removed decoder_os argument since we are only implementing starknet network
	loadedABIs       []string //stores all abis loaded into dispatcher
	all_abis         bool
	functionDecoders map[string]AbiFunctionDecoder//map of function decoders with signature as key and function decoder as value
	eventDecoders    map[string]map[int]AbiEventDecoder//map of event decoders with signature as key and map of indexed parameters and event decoder as value
	dbSession        *sqlx.DB //type? to be implemented in database
}

func NewDecodingDispatcher(bind *sqlx.DB) *DecodingDispatcher {//constructor function
	return &DecodingDispatcher{
		dbSession:        bind,
		loadedABIs:       []string{},
		functionDecoders: make(map[string]AbiFunctionDecoder),
		eventDecoders:    make(map[string]map[int]AbiEventDecoder),
	}
}
func (d *DecodingDispatcher) AddABI(abiName string, abiData interface{}, priority int, kwargs ...interface{}) error {
	//interface allows to hold any data type
	for _, loadedAbi := range d.loadedABIs {
		if loadedAbi == abiName {//already in dispatcher
			return athena.NewDecodingError(abiName + " ABI already loaded into dispatcher")
		}
	}
	log.Printf("Adding ABI %s to dispatcher with priority %d", abiName, priority)
	abiBytes, ok := abiData.([]byte)//asserts that abi data is of list of bytes
	if !ok {
		return errors.New("abiData must be of type []byte")
	}
	abi, err := FromJSON(abiBytes, abiName, kwargs[0].([]byte))
	if err != nil {
		return err
	}
	var functions []AbiFunctionDecoder
	var events []AbiEventDecoder
	for name, f := range abi.Functions {
		functions = append(functions, NewCairoFunctionDecoder(name, f.Inputs, f.Outputs, abiName, priority))
	}
	for name, e := range abi.Events {
		events = append(events, NewCairoEventDecoder(name, e.Parameters, e.Data, e.Keys, abiName, priority))
	}
	d.AddFunctionDecoders(functions) //Adds function decoders from given ABI to the dispatcher.
	d.AddEventDecoders(events)  //Adds event decoders from given ABI to the dispatcher.
	d.loadedABIs = append(d.loadedABIs, abiName)
	log.Printf("Successfully Added %s ABI to DecodingDispatcher", abiName)
	return nil
}
func (d *DecodingDispatcher) AddFunctionDecoders(functions []AbiFunctionDecoder) {
	if len(functions) > 0 {
		funcNames := []string{}
		for _, f := range functions {
			funcNames = append(funcNames, f.Name())
		}
		log.Printf("Adding %s Functions: %s", functions[0].ABIName(), strings.Join(funcNames, ", "))//lists abi name along with its functions that are being loaded
	}
	for _, funcDecoder := range functions {	// Iterate through each function decoder
		existingDecoder, exists := d.functionDecoders[string(funcDecoder.Signature())]
		if !exists || existingDecoder.Priority() < funcDecoder.Priority() {
			// Logging debug information
			log.Printf("Adding function %s from ABI %s to dispatcher with selector 0x%x",
				funcDecoder.Name(), funcDecoder.ABIName(), funcDecoder.Signature())
			// Update the function decoder in the map
			d.functionDecoders[string(funcDecoder.Signature())] = funcDecoder
		} else if existingDecoder.Priority() > funcDecoder.Priority() {
			// Logging debug information
			log.Printf("Function %s with Signature 0x%x already defined in ABI %s with Priority: %d",
				funcDecoder.Name(), funcDecoder.Signature(), existingDecoder.ABIName(), existingDecoder.Priority())
		} else {
			// Handle priority conflict
			log.Printf("ABI %s and %s share the decoder for the function 0x%x, and both are set to priority %d. Increase or decrease the priority of an ABI to resolve this conflict.",
				funcDecoder.ABIName(), existingDecoder.ABIName(), funcDecoder.Signature(), funcDecoder.Priority())
		}
	}
}
func (d *DecodingDispatcher) AddEventDecoders(events []AbiEventDecoder) {
	if len(events) > 0 {
		eventNames := []string{}
		for _, e := range events {
			eventNames = append(eventNames, e.Name())
		}
		log.Printf("Adding %s Events: %s", events[0].ABIName(), strings.Join(eventNames, ", "))
	}
	for _, newEvent := range events {
		existingEvent, exists := d.eventDecoders[string(newEvent.Signature())]
		if !exists {//case 1
			log.Printf("Adding event %s from ABI %s to dispatcher", newEvent.Name(), newEvent.ABIName())
			d.setEventDecoder(newEvent, false)
		} else if existingIndexDecoder, indexExists := existingEvent[newEvent.IndexedParams()]; indexExists {//case2,3
			if existingIndexDecoder.Priority() < newEvent.Priority() {//case 2
				d.setEventDecoder(newEvent, true)
			}
		} else {//case 4
			d.setEventDecoder(newEvent, true)
		}
	}
}
func (d *DecodingDispatcher) setEventDecoder(event AbiEventDecoder, setIndex bool) {
	if !setIndex {//adds the event decoder directly to the eventDecoders map (case 1)
		d.eventDecoders[string(event.Signature())] = map[int]AbiEventDecoder{
			event.IndexedParams(): event,
		}
		return
	}
	if existingDecoder, exists := d.eventDecoders[string(event.Signature())]; exists {
		if existingIndexDecoder, indexExists := existingDecoder[event.IndexedParams()]; indexExists {
			if existingIndexDecoder.Priority() < event.Priority() {
				d.eventDecoders[string(event.Signature())][event.IndexedParams()] = event
			}
		} else {
			d.eventDecoders[string(event.Signature())][event.IndexedParams()] = event
		}
	} else {
		d.eventDecoders[string(event.Signature())] = map[int]AbiEventDecoder{
			event.IndexedParams(): event,
		}
	}
}
// DecodeTransaction decodes a transaction using the appropriate function decoder.
func (d *DecodingDispatcher) DecodeTransaction(txData []byte) (*DecodedFuncDataclass, error) {
	// Extract function signature from the transaction data
	signature := extractFunctionSignature(txData) //?check this function
	funcDecoder, exists := d.functionDecoders[string(signature)]
	if !exists {
		return nil, fmt.Errorf("function decoder not found for signature: %x", signature)
	}
	calldata, result := splitTxData(txData) // ?check this function
	decodedFunc, err := funcDecoder.Decode(calldata, result)
	if err != nil {
		return nil, err
	}
	return decodedFunc, nil
}
func (d *DecodingDispatcher) DecodeEvent(eventData []byte) (*DecodedEventDataclass, error) {
	// Extract event signature from the event data
	signature := extractEventSignature(eventData) // Implement this function to extract the event signature
	// Lookup the event decoders for the extracted signature
	eventDecodersForSignature, exists := d.eventDecoders[string(signature)]
	if !exists {
		return nil, fmt.Errorf("event decoders not found for signature: %x", signature)
	}

	// Extract indexed parameters from the event data
	indexedParams := extractIndexedParams(eventData) // Implement this function to extract indexed parameters

	// Lookup the event decoder for the extracted indexed parameters
	eventDecoder, exists := eventDecodersForSignature[indexedParams]
	if !exists {
		return nil, fmt.Errorf("event decoder not found for indexed params: %d", indexedParams)
	}

	// Split eventData into data and keys; implement this if needed
	data, keys := splitEventData(eventData) // You need to implement this function

	// Decode the event using the found decoder
	decodedEvent, err := eventDecoder.Decode(data, keys)
	if err != nil {
		return nil, err
	}

	return decodedEvent, nil
}

func (d *DecodingDispatcher) DecodeDataclasses(dataKind string, dataclasses []interface{}) {
	switch dataKind {
	case "transactions":
		for _, tx := range dataclasses {
			txData, ok := tx.([]byte)
			if !ok {
				log.Println("Invalid data type for transaction decoding")
				continue
			}
			_, err := d.DecodeTransaction(txData)
			if err != nil {
				log.Println("Error decoding transaction:", err)
			}
		}
	case "events":
		for _, event := range dataclasses {
			eventData, ok := event.([]byte)
			if !ok {
				log.Println("Invalid data type for event decoding")
				continue
			}
			_, err := d.DecodeEvent(eventData)
			if err != nil {
				log.Println("Error decoding event:", err)
			}
		}
	default:
		log.Println("Trace decoding not yet implemented for dataKind:", dataKind)
	}
}

//FROMABI NOT IMPLEMENTED YET
func (d *DecodingDispatcher) GroupABIs() map[string][3]interface{} {
	output := make(map[string][3]interface{})

	for _, name := range d.loadedABIs {
	  /*•	0 for priority (placeholder value)
		•	An empty slice of AbiFunctionDecoder
		•	An empty slice of AbiEventDecoder*/
		output[name] = [3]interface{}{0, []AbiFunctionDecoder{}, []AbiEventDecoder{}}
	}
	for _, funcDecoder := range d.functionDecoders {
		//
		group := output[funcDecoder.ABIName()]
		group[0] = funcDecoder.Priority()
		group[1] = append(group[1].([]AbiFunctionDecoder), funcDecoder)
		output[funcDecoder.ABIName()] = group//Updates the output map with the modified group.
	}
	for _, eventDecoders := range d.eventDecoders {
		for _, eventDecoder := range eventDecoders {
			group := output[eventDecoder.ABIName()]
			group[0] = eventDecoder.Priority()
			group[2] = append(group[2].([]AbiEventDecoder), eventDecoder)
			output[eventDecoder.ABIName()] = group
		}
	}
	return output
}
//CHECK THIS FUNCTION
func (d *DecodingDispatcher) DecoderTable(printFunctions, printEvents, fullSignatures bool) string {
	var result strings.Builder
	// Use strings.Builder as output for tablewriter
	table := tablewriter.NewWriter(&result)
	table.SetHeader([]string{"Name", "Priority", "Functions", "Events"})
	groupedABIs := d.GroupABIs()
	for abiName, data := range groupedABIs {
		priority := data[0].(int)
		functions := data[1].([]AbiFunctionDecoder)
		events := data[2].([]AbiEventDecoder)
		functionsList := []string{}
		if printFunctions {
			for _, f := range functions {
				functionsList = append(functionsList, f.IDStr(fullSignatures))
			}
		}
		eventsList := []string{}
		if printEvents {
			for _, e := range events {
				eventsList = append(eventsList, e.IDStr(fullSignatures))
			}
		}

		table.Append([]string{
			abiName,
			fmt.Sprintf("%d", priority),
			strings.Join(functionsList, "\n"),
			strings.Join(eventsList, "\n"),
		})
	}

	// Render the table to the strings.Builder
	table.Render()
	return result.String()
}
