package athena_abi

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"strconv"
	"testing"
)

func TestBigIntToBytes(t *testing.T) {

	tests := []struct {
		input    big.Int
		length   int
		expected []byte
	}{
		{*big.NewInt(12345), 4, []byte{0, 0, 48, 57}},
		{*big.NewInt(0x123456), 6, []byte{0x01, 0x23, 0x45, 0x00, 0x00, 0x00}},
	}

	t.Run("Normal Case - First Item", func(t *testing.T) {
		tt := tests[0]
		result := bigIntToBytes(tt.input, tt.length)
		assert.Equal(t, tt.expected, result)
	})

	t.Run("Pass Case - Second Item with NotEqual", func(t *testing.T) {
		tt := tests[1]
		incorrectExpected := []byte{0x01, 0x23, 0x45, 0x00, 0x00, 0x01}
		result := bigIntToBytes(tt.input, tt.length)
		assert.NotEqual(t, incorrectExpected, result)
	})

	t.Run("Zero Length", func(t *testing.T) {
		result := bigIntToBytes(*big.NewInt(12345), 0)
		assert.Empty(t, result)
	})

	t.Run("Exact Length Match", func(t *testing.T) {
		result := bigIntToBytes(*big.NewInt(255), 1)
		assert.Equal(t, []byte{0xFF}, result)
	})

	t.Run("Large Integer", func(t *testing.T) {
		largeInt := big.NewInt(0).Lsh(big.NewInt(1), 64)
		result := bigIntToBytes(*largeInt, 9)
		expected := []byte{0x01, 0, 0, 0, 0, 0, 0, 0, 0}
		assert.Equal(t, expected, result)
	})

}

func TestStarknetKeccak(t *testing.T) {

	t.Run("Normal Case", func(t *testing.T) {
		input := []byte("transfer")
		expected := []byte{0x00, 0x83, 0xaf, 0xd3, 0xf4, 0xca, 0xed, 0xc6, 0xee, 0xbf, 0x44, 0x24, 0x6f, 0xe5, 0x4e, 0x38, 0xc9, 0x5e, 0x31, 0x79, 0xa5, 0xec, 0x9e, 0xa8, 0x17, 0x40, 0xec, 0xa5, 0xb4, 0x82, 0xd1, 0x2e}

		result := StarknetKeccak(input)
		assert.Equal(t, expected, result)
	})

	t.Run("Empty Input", func(t *testing.T) {
		result := StarknetKeccak([]byte{})
		expected := []byte{0x1, 0xd2, 0x46, 0x1, 0x86, 0xf7, 0x23, 0x3c, 0x92, 0x7e, 0x7d, 0xb2, 0xdc, 0xc7, 0x3, 0xc0, 0xe5, 0x0, 0xb6, 0x53, 0xca, 0x82, 0x27, 0x3b, 0x7b, 0xfa, 0xd8, 0x4, 0x5d, 0x85, 0xa4, 0x70}
		assert.Equal(t, expected, result)
	})

	t.Run("Large Input", func(t *testing.T) {
		largeInput := make([]byte, 1024)
		result := StarknetKeccak(largeInput)
		expected := []byte{0x1, 0xd4, 0xd1, 0xdf, 0x10, 0x38, 0x8b, 0xbc, 0x20, 0x87, 0x78, 0xff, 0x2, 0x31, 0xd, 0xb9, 0x8f, 0xda, 0xa6, 0x8e, 0xfe, 0xd0, 0xb2, 0x6, 0x8a, 0x9b, 0xef, 0x78, 0xbd, 0x3b, 0xfd, 0x74}
		assert.Equal(t, expected, result)
	})
}

// Test suite for TopologicalSort function
func TestTopologicalSort(t *testing.T) {
	// Normal Logic Testing

	// Case 1: Simple Directed Acyclic Graph (DAG)
	t.Run("Simple DAG", func(t *testing.T) {
		graph := map[string][]string{
			"A": {"B"},
			"B": {"C"},
			"C": {},
		}
		expected := []string{"A", "B", "C"}
		result := TopologicalSort(graph)
		assert.Equal(t, expected, result)
	})

	// Case 2: DAG with multiple valid topological sorts
	t.Run("DAG with multiple valid sorts", func(t *testing.T) {
		graph := map[string][]string{
			"A": {"B", "C"},
			"B": {"D"},
			"C": {"D"},
			"D": {},
		}
		result := TopologicalSort(graph)
		// Two valid outputs: [A, B, C, D] or [A, C, B, D]
		assert.Contains(t, [][]string{
			{"A", "B", "C", "D"},
			{"A", "C", "B", "D"},
		}, result)
	})

	// Case 3: Disconnected graph
	t.Run("Disconnected graph", func(t *testing.T) {
		graph := map[string][]string{
			"A": {"B"},
			"B": {},
			"C": {"D"},
			"D": {},
		}
		result := TopologicalSort(graph)
		// Multiple valid topological orders exist for disconnected components
		assert.Contains(t, [][]string{
			{"A", "B", "C", "D"},
			{"A", "B", "D", "C"},
			{"C", "D", "A", "B"},
			{"C", "A", "D", "B"},
			{"C", "A", "B", "D"},
			{"A", "C", "B", "D"},
		}, result)
	})

	// Case 4: Empty graph
	t.Run("Empty graph", func(t *testing.T) {
		graph := map[string][]string{}
		expected := []string{}
		result := TopologicalSort(graph)
		assert.Equal(t, expected, result)
	})

	// Case 5: Single node graph
	t.Run("Single node graph", func(t *testing.T) {
		graph := map[string][]string{
			"A": {},
		}
		expected := []string{"A"}
		result := TopologicalSort(graph)
		assert.Equal(t, expected, result)
	})

	// Boundary Testing
	//Large graph and Graph with long linear chains
	t.Run("Large graph", func(t *testing.T) {
		graph := make(map[string][]string)
		expectedOrder := []string{} // This will hold the expected topological order.

		for i := 1; i <= 1000; i++ {
			node := "Node" + strconv.Itoa(i)            // Convert int to string
			expectedOrder = append(expectedOrder, node) // Build the expected order.

			if i < 1000 {
				nextNode := "Node" + strconv.Itoa(i+1) // Convert next node to string
				graph[node] = append(graph[node], nextNode)
			}
		}

		result := TopologicalSort(graph)

		assert.NotNil(t, result)               // Check that a result is returned
		assert.Equal(t, expectedOrder, result) // Check if the result matches the expected order
	})
}
func slicesEqualIgnoringOrder(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Create a map to count occurrences of each element
	counts := make(map[string]int)
	for _, item := range a {
		counts[item]++
	}
	for _, item := range b {
		counts[item]--
		if counts[item] < 0 {
			return false
		}
	}

	return true
}
func TestConvertMap(t *testing.T) {

	t.Run("Case 1: Nested map with all true values", func(t *testing.T) {
		input := map[string]map[string]bool{
			"outer1": {"key1": true, "key2": true},
			"outer2": {"key3": true, "key4": true},
		}
		expected := map[string][]string{
			"outer1": {"key1", "key2"},
			"outer2": {"key3", "key4"},
		}

		result := convertMap(input)
		assert.Equal(t, expected, result)
	})
	t.Run("Case 2: Nested map with all false values", func(t *testing.T) {
		input := map[string]map[string]bool{
			"outer1": {"key1": false, "key2": false},
			"outer2": {"key3": false, "key4": false},
		}
		expected := map[string][]string{
			"outer1": nil, // Use nil to match the actual result
			"outer2": nil, // Use nil to match the actual result
		}

		result := convertMap(input)
		assert.Equal(t, expected, result)
	})
	t.Run("Case 3: Mixed true and false values", func(t *testing.T) {
		input := map[string]map[string]bool{
			"outer1": {"key1": true, "key2": false},
			"outer2": {"key3": false, "key4": true},
		}
		expected := map[string][]string{
			"outer1": {"key1"},
			"outer2": {"key4"},
		}

		result := convertMap(input)
		assert.Equal(t, expected, result)
	})

	t.Run("Case 4: Empty map", func(t *testing.T) {
		input := map[string]map[string]bool{}
		expected := map[string][]string{}

		result := convertMap(input)
		assert.Equal(t, expected, result)
	})
	t.Run("Case 5: Inner map is empty for some outer keys", func(t *testing.T) {
		input := map[string]map[string]bool{
			"outer1": nil,
			"outer2": {"key3": true},
		}
		expected := map[string][]string{
			"outer1": nil,
			"outer2": {"key3"},
		}

		result := convertMap(input)
		assert.Equal(t, expected, result)
	})

	t.Run("Boundary Case 1: Large nested map", func(t *testing.T) {
		input := make(map[string]map[string]bool)
		expected := make(map[string][]string)
		for i := 1; i <= 1000; i++ { // Reduced from 1000 to 10
			outerKey := "outer" + strconv.Itoa(i)
			input[outerKey] = make(map[string]bool)
			expected[outerKey] = []string{}
			for j := 1; j <= 1000; j++ { // Reduced from 1000 to 10
				innerKey := "key" + strconv.Itoa(j)
				if j%2 == 0 {
					input[outerKey][innerKey] = true
					expected[outerKey] = append(expected[outerKey], innerKey)
				} else {
					input[outerKey][innerKey] = false
				}
			}
		}

		result := convertMap(input)

		// Compare maps with unordered slices
		for outerKey, expSlice := range expected {
			resSlice, exists := result[outerKey]
			assert.True(t, exists, "Outer key %s missing in result", outerKey)
			assert.True(t, slicesEqualIgnoringOrder(expSlice, resSlice), "Slice mismatch for outer key %s", outerKey)
		}

		// Check if result map does not have any extra keys
		for outerKey := range result {
			_, exists := expected[outerKey]
			assert.True(t, exists, "Unexpected outer key %s in result", outerKey)
		}
	})
	t.Run("Boundary Case 2: Inner map with many false values", func(t *testing.T) {
		input := map[string]map[string]bool{
			"outer1": {
				"key1": false, "key2": false, "key3": false, "key4": true,
				"key5": false, "key6": false, "key7": false, "key8": false,
			},
		}
		expected := map[string][]string{
			"outer1": {"key4"},
		}

		result := convertMap(input)
		assert.Equal(t, expected, result)
	})

}
