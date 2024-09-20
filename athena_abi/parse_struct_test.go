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
