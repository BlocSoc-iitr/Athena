package athena_abi

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFunctionSignatures(t *testing.T) {

	transfer := NewAbiFunction("transfer", []AbiParameter{
		{Name: "recipient", Type: ContractAddress},
		{Name: "amount", Type: U256},
	}, []StarknetType{Bool}, "")

	assert.Equal(t, "Function(recipient:ContractAddress,amount:U256) -> (Bool)", transfer.idStr())

	expectedSignature := "0083afd3f4caedc6eebf44246fe54e38c95e3179a5ec9ea81740eca5b482d12e"
	assert.Equal(t, expectedSignature, hex.EncodeToString(transfer.signature))
}

func TestEventSignatures(t *testing.T) {
	transfer := NewAbiEvent(
		"Transfer",
		[]string{"from", "to", "amount"},
		map[string]StarknetType{
			"from":   ContractAddress,
			"to":     ContractAddress,
			"amount": U256,
		},
		make(map[string]StarknetType),
		"",
	)
	idStr, err := transfer.idStr()
	assert.NoError(t, err, "Error getting id string")

	assert.Equal(t, "Event(from:ContractAddress,to:ContractAddress,amount:U256)", idStr, "Unexpected id string")

	expectedSignature := "0099cd8bde557814842a3121e8ddfd433a539b8c9f14bf31ebf108d12e6196e9"
	assert.Equal(t, expectedSignature, hex.EncodeToString(transfer.signature), "Unexpected signature")
}

func TestKeyEventSignature(t *testing.T) {
	transfer := NewAbiEvent(
		"Transfer",
		[]string{"from", "to", "amount"},
		map[string]StarknetType{
			"amount": U256,
		},
		map[string]StarknetType{
			"from": ContractAddress,
			"to":   ContractAddress,
		},
		"",
	)

	expectedIDStr := "Event(<from>:ContractAddress,<to>:ContractAddress,amount:U256)"
	actualIDStr, _ := transfer.idStr()

	assert.Equal(t, expectedIDStr, actualIDStr, "ID strings should match")

	expectedSignatureHex := "0099cd8bde557814842a3121e8ddfd433a539b8c9f14bf31ebf108d12e6196e9"
	actualSignatureHex := hex.EncodeToString(transfer.signature)

	assert.Equal(t, expectedSignatureHex, actualSignatureHex, "Signatures should match")
}

func TestLoadEthAbi(t *testing.T) {
	// Load the ABI using the load_abi function
	ethAbi, err := loadAbi("starknet_eth", 2)
	assert.NoError(t, err, "Loading ABI should not return an error")

	// Convert the hexadecimal string to a byte slice
	classHash, err := hex.DecodeString("05ffbcfeb50d200a0677c48a129a11245a3fc519d1d98d76882d1c9a1b19c6ed")
	assert.NoError(t, err, "Decoding class hash from hex should not return an error")

	// Call the StarknetAbi from JSON method
	ethDecoder, err := StarknetAbiFromJSON(ethAbi, "starknet_eth", classHash)
	assert.NoError(t, err, "Decoding ABI from JSON should not return an error")

	// Additional assertions can be made here based on the expected structure of ethDecoder
	// For example, check the number of functions or events parsed, etc.
	assert.NotNil(t, ethDecoder, "EthDecoder should not be nil")
	assert.Equal(t, "starknet_eth", *ethDecoder.ABIName, "ABI name should match")
}

/*func TestLoadWildcardArraySyntax(t *testing.T) {
    // Load the ABI
    wildcardAbi, err := loadAbi("complex_array", 1)
    assert.NoError(t, err, "Failed to load ABI")

    // Convert the hex string to bytes
    classHash, err := hex.DecodeString("0031da92cf5f54bcb81b447e219e2b791b23f3052d12b6c9abd04ff2e5626576")
    assert.NoError(t, err, "Failed to decode hex string")

    // Create the decoder using the from JSON function
    decoder, err := StarknetAbiFromJSON(wildcardAbi, "complex_array", classHash)
    assert.NoError(t, err, "Failed to create Starknet ABI from JSON")

    // Add assertions to verify the properties of `decoder`
    // Example assertions (adjust according to expected values):
    assert.NotNil(t, decoder, "Decoder should not be nil")
    //assert.Equal(t, expectedValue, decoder.SomeProperty, "Unexpected value for SomeProperty")
}


func TestLoadWildcardArraySyntax(t *testing.T) {
    wildcardAbi,err := loadAbi("complex_array", 1)
    assert.NoError(t, err, "Loading ABI should not return an error")
    // Decode the hex string directly
    decodedClassHash, err := hex.DecodeString("0031da92cf5f54bcb81b447e219e2b791b23f3052d12b6c9abd04ff2e5626576")
    assert.NoError(t, err, "Failed to decode hex string")

    decoder, err := StarknetAbiFromJSON(
        wildcardAbi,
        "complex_array",
        decodedClassHash,
    )
    assert.NoError(t, err, "Failed to decode ABI")

    parsedEvent := decoder.Events["log_storage_cells"]

    // Assert the length of parsed event data
    assert.Len(t, parsedEvent.data, 1, "Expected parsed event data length to be 1")

    // Assert the storage_cells data matches the expected structure
    assert.Equal(t, parsedEvent.data["storage_cells"], StarknetArray(
        StarknetStruct{
            Name: "StorageCell",
            Members: []AbiParameter{
                {Name: "key", Type: StarknetCoreType.Felt},
                {Name: "value", Type: StarknetCoreType.Felt},
            },
        },
    ), "Expected storage_cells data structure to match")

    // Assert the event name
    assert.Equal(t, parsedEvent.Name, "log_storage_cells", "Expected event name to be 'log_storage_cells'")
}
*/

func TestLoadWildcardArraySyntax(t *testing.T) {
	// Load the ABI (you'll need to implement this function)
	wildcardAbi, err := loadAbi("complex_array", 1)
	assert.NoError(t, err, "Loading ABI should not return an error")

	classHash, _ := hex.DecodeString("0031da92cf5f54bcb81b447e219e2b791b23f3052d12b6c9abd04ff2e5626576")
	//fmt.Println("wildcard",wildcardAbi)
	decoder, err := StarknetAbiFromJSON(wildcardAbi, "complex_array", classHash)
	assert.NoError(t, err, "there should not be error")
	fmt.Println("decoder is ", decoder)
	fmt.Println("the err is ", err)
	parsedEvent, ok := decoder.Events["log_storage_cells"]
	fmt.Println("parsedevent is ", parsedEvent)
	assert.True(t, ok, "Event 'log_storage_cells' should exist")

	assert.Equal(t, 1, len(parsedEvent.data), "Parsed event should have 1 data field")

	storageCellsType, ok := parsedEvent.data["storage_cells"]
	assert.True(t, ok, "Event should have 'storage_cells' field")

	arrayType, ok := storageCellsType.(StarknetArray)
	assert.True(t, ok, "storage_cells should be a StarknetArray")

	structType, ok := arrayType.InnerType.(StarknetStruct)
	assert.True(t, ok, "Array element should be a StarknetStruct")
	assert.Equal(t, "StorageCell", structType.Name)

	assert.Equal(t, 2, len(structType.Members), "StorageCell struct should have 2 members")
	fmt.Println("hello hello the val is ", structType.Members)
	assert.Equal(t, "key", structType.Members[0].Name)
	assert.Equal(t, StarknetCoreType(Felt), structType.Members[0].Type)

	assert.Equal(t, "value", structType.Members[1].Name)
	assert.Equal(t, StarknetCoreType(Felt), structType.Members[1].Type)

	assert.Equal(t, "log_storage_cells", parsedEvent.name)

	// Test the idStr() methods
	expectedIdStr := "[{key:Felt,value:Felt}]"
	assert.Equal(t, expectedIdStr, arrayType.idStr(), "Array idStr should match expected")

	expectedStructIdStr := "{key:Felt,value:Felt}"
	assert.Equal(t, expectedStructIdStr, structType.idStr(), "Struct idStr should match expected")
}

func TestWildcardSizeSyntax(t *testing.T) {
	// felt* syntax length parameter can be calldata_len or calldata_size
	abiFunction := map[string]interface{}{
		"inputs": []interface{}{
			map[string]interface{}{"name": "selector", "type": "felt"},
			map[string]interface{}{"name": "calldata_size", "type": "felt"},
			map[string]interface{}{"name": "calldata", "type": "felt*"},
		},
		"name": "__default__",
		"outputs": []interface{}{
			map[string]interface{}{"name": "retdata_size", "type": "felt"},
			map[string]interface{}{"name": "retdata", "type": "felt*"},
		},
		"type": "function",
	}

	customTypes := make(map[string]interface{})
	parsedAbiFunc, err := ParseAbiFunction(abiFunction, customTypes)

	assert.NoError(t, err, "ParseAbiFunction should not return an error")
	assert.NotNil(t, parsedAbiFunc, "ParseAbiFunction should return a non-nil result")

	assert.Len(t, parsedAbiFunc.inputs, 2, "There should be 2 inputs")

	assert.Equal(t, "selector", parsedAbiFunc.inputs[0].Name, "First input name should be 'selector'")
	assert.Equal(t, Felt, parsedAbiFunc.inputs[0].Type, "First input type should be Felt")

	assert.Equal(t, "calldata", parsedAbiFunc.inputs[1].Name, "Second input name should be 'calldata'")
	assert.IsType(t, StarknetArray{}, parsedAbiFunc.inputs[1].Type, "Second input type should be StarknetArray")

	calldataType, ok := parsedAbiFunc.inputs[1].Type.(StarknetArray)
	assert.True(t, ok, "Second input type should be castable to StarknetArray")
	assert.Equal(t, Felt, calldataType.InnerType, "Inner type of calldata should be Felt")
}
func TestNoStructDefinition(t *testing.T) {
	// Assuming NO_STRUCT_ABI_DEFINITION is a JSON string
	var abiJson []map[string]interface{}
	err := json.Unmarshal([]byte(NO_STRUCT_ABI_DEFINITION), &abiJson) // Unmarshal JSON string into abiJson
	assert.NoError(t, err, "Error unmarshalling ABI for no_struct")

	classHash, err := hex.DecodeString(NO_STRUCT_CLASS_HASH[2:])
	assert.NoError(t, err, "Error decoding class hash for no_struct")

	// Decode ABI using StarknetAbiFromJSON
	decoder, err := StarknetAbiFromJSON(abiJson, "no_struct", classHash)
	assert.NoError(t, err, "Error decoding ABI for no_struct")
	assert.NotNil(t, decoder, "Expected decoder to be non-nil for no_struct")
}
func TestFeltTypes(t *testing.T) {
	// Assuming VERSION_0_ABI_DEFINITION is a JSON string
	var abiJson []map[string]interface{}
	err := json.Unmarshal([]byte(VERSION_0_ABI_DEFINITION), &abiJson) // Unmarshal JSON string into abiJson
	assert.NoError(t, err, "Error unmarshalling ABI for felt_types")

	classHash, err := hex.DecodeString(VERSION_0_CLASS_HASH[2:])
	assert.NoError(t, err, "Error decoding class hash for felt_types")

	// Decode ABI using StarknetAbiFromJSON
	decoder, err := StarknetAbiFromJSON(abiJson, "felt_types", classHash)
	assert.NoError(t, err, "Error decoding ABI for felt_types")
	assert.NotNil(t, decoder, "Expected decoder to be non-nil for felt_types")
}

// Test for parsing event keys from the ERC20 ABI definition
func TestParseEventKeys(t *testing.T) {
	// Load the ABI for erc20_key_events with version 2
	abiJson, err := loadAbi("erc20_key_events", 2)
	assert.NoError(t, err, "Error loading ABI for erc20_key_events")

	// Decode the class hash from a hex string
	classHash, err := hex.DecodeString("0261ad90e1901833f794ee3d69816846f68ddb4fb7bb9ffec2d8f0c8608e298d")
	assert.NoError(t, err, "Error decoding class hash for erc20_key_events")

	// Decode the ABI using StarknetAbiFromJSON
	parsedAbi, err := StarknetAbiFromJSON(abiJson, "erc20_key_events", classHash)
	fmt.Println("parsedabi is helc  ", parsedAbi)
	fmt.Println("the err is helc ", err)
	assert.NoError(t, err, "Error parsing ABI for erc20_key_events")
	assert.NotNil(t, parsedAbi, "Parsed ABI should not be nil for erc20_key_events")

	// Access the "Approval" event from the parsed ABI
	approveEvent, ok := parsedAbi.Events["Approval"]
	assert.True(t, ok, "Approval event should be found in the parsed ABI")
	fmt.Println("approve event ", approveEvent)
	// Validate the event's parameters
	expectedParameters := []string{"owner", "spender", "value"}
	assert.Equal(t, expectedParameters, approveEvent.parameters, "Expected parameters do not match")

	// Validate the event's keys
	expectedKeys := map[string]StarknetType{ //confirm this change
		"owner":   ContractAddress,
		"spender": ContractAddress,
	}
	assert.Equal(t, expectedKeys, approveEvent.keys, "Expected keys do not match") //

	// Validate the event's data
	expectedData := map[string]StarknetType{ //confirm this change
		"value": U256,
	}
	assert.Equal(t, expectedData, approveEvent.data, "Expected data do not match") //

	// Validate the event's name
	assert.Equal(t, "Approval", approveEvent.name, "Expected event name does not match")
}
