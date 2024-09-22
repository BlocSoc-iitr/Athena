package athena_abi

import (
    "math/big"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestEventDecode(t *testing.T) {
    tests := []struct {
        name       string
        abiEvent   *AbiEvent
        eventData  []*big.Int
        eventKeys  []*big.Int
        expected   map[string]interface{}
    }{
        {
            name: "Approval Event",
            abiEvent: NewAbiEvent(
                "Approval",
                []string{"owner", "spender", "value"},
                map[string]StarknetType{"value": U256},
                map[string]StarknetType{"owner": U256, "spender": U256},
                "erc20_key_events",
            ),
            eventData: []*big.Int{
                mustParseBigInt("C95E3D845779376FED50", 16),
                big.NewInt(0),
            },
            eventKeys: []*big.Int{
                mustParseBigInt("0134692B230B9E1FFA39098904722134159652B09C5BC41D88D6698779D228FF", 16), // Event ID
                mustParseBigInt("060CAFC0B0E66067B3A4978E93552DE54E0CAEEB82A352A202E0DC79A41459B6", 16), // Owner low
                big.NewInt(0), // Owner high
                mustParseBigInt("04270219D365D6B017231B52E92B3FB5D7C8378B05E9ABC97724537A80E93B0F", 16), // Spender low
                big.NewInt(0), // Spender high
            },
            expected: map[string]interface{}{
                "owner":   mustParseBigInt("060CAFC0B0E66067B3A4978E93552DE54E0CAEEB82A352A202E0DC79A41459B6", 16),
                "spender": mustParseBigInt("04270219D365D6B017231B52E92B3FB5D7C8378B05E9ABC97724537A80E93B0F", 16),
                "value":   mustParseBigInt("C95E3D845779376FED50", 16), //Doubt here needs review 
            },
        },
    }

    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            decodedEvent, err := test.abiEvent.Decode(test.eventData, test.eventKeys)
            if err != nil {
                t.Fatalf("Error decoding event: %v\nEvent Data: %v\nEvent Keys: %v", err, test.eventData, test.eventKeys)
            }

            assert.Equal(t, test.abiEvent.abiName, decodedEvent.abiName)
            assert.Equal(t, test.abiEvent.name, decodedEvent.name)

            for key, expectedValue := range test.expected {
                actualValue, exists := decodedEvent.data[key]
                assert.True(t, exists, "Key %s not found in decoded event", key)
                if !exists {
                    continue
                }

                switch expected := expectedValue.(type) {
                case string:
                    actual, ok := actualValue.(*big.Int)
                    assert.True(t, ok, "Expected *big.Int for key %s, got %T", key, actualValue)
                    if ok {
                        assert.Equal(t, expected, bigIntToAddress(actual))
                    }
                case *big.Int:
                    actual, ok := actualValue.(*big.Int)
                    assert.True(t, ok, "Expected *big.Int for key %s, got %T", key, actualValue)
                    if ok {
                        assert.Equal(t, 0, expected.Cmp(actual), "For key %s, expected %s, got %s", key, expected.String(), actual.String())
                    }
                default:
                    t.Errorf("Unexpected type for key %s: %T", key, expectedValue)
                }
            }
        })
    }
}

// Helper function to parse big integers
func mustParseBigInt(s string, base int) *big.Int {
    n, ok := new(big.Int).SetString(s, base)
    if !ok {
        panic("Failed to parse big integer: " + s)
    }
    return n
}

// Helper function to convert big.Int to address string
func bigIntToAddress(n *big.Int) string {
    return "0x" + n.Text(16)
}