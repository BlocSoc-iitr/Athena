package athena_abi

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseV1AbisToStarknetAbi(t *testing.T) {

	v1Abis, err := GetAbisForVersion("v1")
	if err != nil {
		t.Fatalf("Failed to get v1 ABIs: %v", err)
	}

	for abiName, abi := range v1Abis {
		t.Logf("Testing ABI Name: %s, ", abiName)

		_, err := StarknetAbiFromJSON(abi, abiName, []byte{})
		if err != nil {
			t.Errorf("Failed to parse v1 ABI for %s: %v", abiName, err)
		}
	}
}

func TestParseV2AbisToStarknetAbi(t *testing.T) {
	v2Abis, err := GetAbisForVersion("v2")
	if err != nil {
		t.Fatalf("Failed to get v2 ABIs: %v", err)
	}

	for abiName, abi := range v2Abis {
		// Replace with actual ABI name from data
		decoder, err := StarknetAbiFromJSON(abi, abiName, []byte{})
		if err != nil {
			t.Errorf("Failed to parse v2 ABI: %v", err)
		}

		// Add specific assertions based on ABI name
		if abiName == "starknet_eth" {
			if _, ok := decoder.Functions["transfer"]; !ok {
				t.Errorf("Expected function 'transfer' in starknet_eth ABI")
			}
		}

		if abiName == "argent_account_v3" {
			funcDef := decoder.Functions["change_guardian_backup"]
			if len(funcDef.inputs) > 0 && funcDef.inputs[0].Type.idStr() != "StarknetStruct" {
				t.Errorf("Expected first input to be a StarknetStruct")
			}
		}
	}
}

func TestNamedTupleParsing(t *testing.T) {
	abiFile := filepath.Join("abis", "v1", "legacy_named_tuple.json")
	abiJson, err := os.ReadFile(abiFile)
	if err != nil {
		t.Fatalf("Failed to read ABI file: %v", err)
	}

	var abi []map[string]interface{}
	if err := json.Unmarshal(abiJson, &abi); err != nil {
		t.Fatalf("Failed to unmarshal ABI JSON: %v", err)
	}

	classHash, _ := hex.DecodeString("0484c163658bcce5f9916f486171ac60143a92897533aa7ff7ac800b16c63311")
	parsedAbi, err := StarknetAbiFromJSON(abi, "legacy_named_tuple", classHash)
	if err != nil {
		t.Fatalf("Failed to parse named tuple ABI: %v", err)
	}

	// Assertions based on the parsed ABI
	funcDef := parsedAbi.Functions["xor_counters"]
	if len(funcDef.inputs) == 0 || funcDef.inputs[0].Name != "index_and_x" {
		t.Errorf("Expected input 'index_and_x' in xor_counters function")
	}
}

func TestStorageAddressParsing(t *testing.T) {
	abiFile := filepath.Join("abis", "v2", "storage_address.json")
	abiJson, err := os.ReadFile(abiFile)
	if err != nil {
		t.Fatalf("Failed to read ABI file: %v", err)
	}

	var abi []map[string]interface{}
	if err := json.Unmarshal(abiJson, &abi); err != nil {
		t.Fatalf("Failed to unmarshal ABI JSON: %v", err)
	}

	classHash, _ := hex.DecodeString("0484c163658bcce5f9916f486171ac60143a92897533aa7ff7ac800b16c63311")
	parsedAbi, err := StarknetAbiFromJSON(abi, "storage_address", classHash)
	if err != nil {
		t.Fatalf("Failed to parse storage address ABI: %v", err)
	}

	// Assertions based on parsed ABI
	storageFunction := parsedAbi.Functions["storage_read"]
	if len(storageFunction.inputs) != 2 || storageFunction.inputs[0].Name != "address_domain" {
		t.Errorf("Expected two inputs with first input named 'address_domain'")
	}
}
