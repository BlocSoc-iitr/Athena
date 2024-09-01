package readers

import (
	"github.com/DarkLord017/athena/athena/types"
	"testing"
)

func TestStringIsCorrect(t *testing.T) {
	network := types.StarkNet
	if network.String() != "StarkNet" {
		t.Errorf("Expected %s, got %s", "StarkNet", network.String())
	}
}