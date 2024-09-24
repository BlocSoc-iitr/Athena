package athena_abi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ethAbiJson []map[string]interface{}
var argentAccountAbi []map[string]interface{}

// loadAbi loads the ABI JSON file and returns its content
func loadAbi(abiName string, abiVersion int) ([]map[string]interface{}, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get current file path")
	}

	parentDir := filepath.Dir(filename)
	abiPath := filepath.Join(parentDir, "abis", fmt.Sprintf("v%d", abiVersion), fmt.Sprintf("%s.json", abiName))

	file, err := os.Open(abiPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ABI file: %w", err)
	}
	defer file.Close()

	var abiJson []map[string]interface{}
	err = json.NewDecoder(file).Decode(&abiJson)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ABI JSON: %w", err)
	}
	return abiJson, nil
}
func init() {
	var err error

	// Load starknet_eth ABI
	ethAbiJson, err = loadAbi("starknet_eth", 2)
	if err != nil {
		fmt.Println("Failed to load starknet_eth ABI:", err)
	}

	// Load argent_account ABI
	argentAccountAbi, err = loadAbi("argent_account", 2)
	if err != nil {
		fmt.Println("Failed to load argent_account ABI:", err)
	}
}

func TestStructOrdering(t *testing.T) {
	// Load ABI JSONs
	require.NotNil(t, ethAbiJson, "starknet_eth ABI should be loaded")
	require.NotNil(t, argentAccountAbi, "argent_account ABI should be loaded")

	// Group ABI by type
	groupedAbi := GroupAbiByType(ethAbiJson)

	// Get structs from grouped ABI
	structs, ok := groupedAbi["type_def"]
	require.True(t, ok, "type_def should exist in grouped ABI")

	// Assert the order of structs
	assert.Equal(t, "core::integer::u256", structs[0]["name"], "First struct should be core::integer::u256")
	assert.Equal(t, "struct", structs[0]["type"], "First item should be a struct")

	assert.Equal(t, "core::array::Span::<core::felt252>", structs[1]["name"], "Second struct should be core::array::Span::<core::felt252>")
	assert.Equal(t, "struct", structs[1]["type"], "Second item should be a struct")

	assert.Equal(t, "src::replaceability_interface::EICData", structs[2]["name"], "Third struct should be src::replaceability_interface::EICData")
	assert.Equal(t, "struct", structs[2]["type"], "Third item should be a struct")

	assert.Equal(t, "core::option::Option::<src::replaceability_interface::EICData>", structs[3]["name"], "Fourth item should be core::option::Option::<src::replaceability_interface::EICData>")
	assert.Equal(t, "enum", structs[3]["type"], "Fourth item should be an enum")

	assert.Equal(t, "core::bool", structs[4]["name"], "Fifth item should be core::bool")
	assert.Equal(t, "enum", structs[4]["type"], "Fifth item should be an enum")

	assert.Equal(t, "src::replaceability_interface::ImplementationData", structs[5]["name"], "Sixth struct should be src::replaceability_interface::ImplementationData")
	assert.Equal(t, "struct", structs[5]["type"], "Sixth item should be a struct")
}

func TestExcludeCommonStructsAndEnums(t *testing.T) {
	// Assuming ETH_ABI_JSON is defined somewhere in your test setup
	require.NotNil(t, ethAbiJson, "starknet_eth ABI should be loaded")
	require.NotNil(t, argentAccountAbi, "argent_account ABI should be loaded")
	groupedAbi := GroupAbiByType(ethAbiJson)
	structDict, err := ParseEnumsAndStructs(groupedAbi["type_def"])
	assert.NoError(t, err)
	assert.Equal(t, 2, len(structDict))

	expectedEicDataStruct := StarknetStruct{
		Name: "src::replaceability_interface::EICData",
		Members: []AbiParameter{
			{Name: "eic_hash", Type: ClassHash},
			{Name: "eic_init_data", Type: StarknetArray{InnerType: Felt}},
		},
	}

	assert.Equal(t, expectedEicDataStruct, structDict["src::replaceability_interface::EICData"])

	expectedImplementationDataStruct := StarknetStruct{
		Name: "src::replaceability_interface::ImplementationData",
		Members: []AbiParameter{
			{Name: "impl_hash", Type: ClassHash},
			{Name: "eic_data", Type: StarknetOption{InnerType: expectedEicDataStruct}},
			{Name: "final", Type: Bool},
		},
	}

	assert.Equal(t, expectedImplementationDataStruct, structDict["src::replaceability_interface::ImplementationData"])
}

func TestEnumParsing(t *testing.T) {
	groupedAbi := GroupAbiByType(argentAccountAbi)
	typeDict, err := ParseEnumsAndStructs(groupedAbi["type_def"])
	require.NoError(t, err)

	assert.Len(t, typeDict, 5)

	escapeStatus, ok := typeDict["account::escape::EscapeStatus"].(StarknetEnum)
	require.True(t, ok)

	expectedVariants := []struct {
		Name string
		Type StarknetType
	}{
		{"None", NoneType},
		{"NotReady", NoneType},
		{"Ready", NoneType},
		{"Expired", NoneType},
	}

	assert.Equal(t, "account::escape::EscapeStatus", escapeStatus.Name)
	assert.Equal(t, expectedVariants, escapeStatus.Variants)
}

func TestTupleParsing(t *testing.T) {
	customTypes := make(map[string]interface{})

	t.Run("Single Tuple", func(t *testing.T) {
		singleTuple, err := ParseTuple("(core::felt252, core::bool)", customTypes)
		assert.NoError(t, err)
		assert.Equal(t, StarknetTuple{Members: []StarknetType{Felt, Bool}}, singleTuple)
	})

	t.Run("Nested Tuple 1", func(t *testing.T) {
		nestedTuple1, err := ParseTuple("(core::felt252, (core::bool, core::integer::u256))", customTypes)
		assert.NoError(t, err)
		expected := StarknetTuple{Members: []StarknetType{
			Felt,
			StarknetTuple{Members: []StarknetType{Bool, U256}},
		}}
		assert.Equal(t, expected, nestedTuple1)
	})

	t.Run("Nested Tuple 2", func(t *testing.T) {
		nestedTuple2, err := ParseTuple("(core::felt252, ((core::integer::u16, core::integer::u32), core::bool), core::integer::u256)", customTypes)
		assert.NoError(t, err)
		expected := StarknetTuple{Members: []StarknetType{
			Felt,
			StarknetTuple{Members: []StarknetType{
				StarknetTuple{Members: []StarknetType{U16, U32}},
				Bool,
			}},
			U256,
		}}
		assert.Equal(t, expected, nestedTuple2)
	})
}

var UnorderedStructs = []map[string]interface{}{
	{
		"type": "struct",
		"name": "betting::betting::Bet",
		"members": []map[string]interface{}{
			{"name": "expire_timestamp", "type": "core::integer::u64"},
			{"name": "bettor", "type": "betting::betting::UserData"},
			{"name": "counter_bettor", "type": "betting::betting::UserData"},
			{"name": "amount", "type": "core::integer::u256"},
		},
	},
	{
		"type": "struct",
		"name": "betting::betting::UserData",
		"members": []map[string]interface{}{
			{"name": "address", "type": "core::starknet::contract_address::ContractAddress"},
			{"name": "total_assets", "type": "core::integer::u256"},
		},
	},
	{
		"type": "struct",
		"name": "core::integer::u256",
		"members": []map[string]interface{}{
			{"name": "low", "type": "core::integer::u128"},
			{"name": "high", "type": "core::integer::u128"},
		},
	},
}

func TestBuildTypeGraph(t *testing.T) {
	typeGraph := BuildTypeGraph(UnorderedStructs)

	expectedTypeGraph := map[string]map[string]bool{
		"betting::betting::Bet": {
			"betting::betting::UserData": true,
			"core::integer::u256":        true,
		},
		"betting::betting::UserData": {
			"core::integer::u256": true,
		},
		"core::integer::u256": {},
	}

	assert.Equal(t, expectedTypeGraph, typeGraph, "The type graph should match the expected result")

	sortedDefs, err := TopoSortTypeDefs(UnorderedStructs)
	assert.NoError(t, err, "TopoSortTypeDefs should not return an error")

	expectedOrder := []string{
		"core::integer::u256",
		"betting::betting::UserData",
		"betting::betting::Bet",
	}

	actualOrder := make([]string, len(sortedDefs))
	for i, def := range sortedDefs {
		actualOrder[i] = def["name"].(string)
	}

	assert.Equal(t, expectedOrder, actualOrder, "The sorted order should match the expected result")
}

func TestStructTopoSorting(t *testing.T) {
	topoSortedTypeDefs, err := TopoSortTypeDefs(UnorderedStructs)
	assert.NoError(t, err)

	expectedOrder := []string{
		"core::integer::u256",
		"betting::betting::UserData",
		"betting::betting::Bet",
	}

	assert.Equal(t, len(expectedOrder), len(topoSortedTypeDefs), "Number of sorted structs doesn't match expected")

	for i, expectedName := range expectedOrder {
		assert.Equal(t, expectedName, topoSortedTypeDefs[i]["name"], "Struct at index %d should be %s", i, expectedName)
	}

	// Verify the contents of each struct
	assert.Equal(t, UnorderedStructs[2], topoSortedTypeDefs[0], "First struct should be core::integer::u256")
	assert.Equal(t, UnorderedStructs[1], topoSortedTypeDefs[1], "Second struct should be betting::betting::UserData")
	assert.Equal(t, UnorderedStructs[0], topoSortedTypeDefs[2], "Third struct should be betting::betting::Bet")
}
