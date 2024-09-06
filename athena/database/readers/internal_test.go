package readers

import (
	"github.com/BlocSoc-iitr/Athena/athena/types"
	
	"testing"
)

func TestStringIsCorrect(t *testing.T) {
	network := types.StarkNet
	if network.String() != "StarkNet" {
		t.Errorf("Expected %s, got %s", "StarkNet", network.String())
	}
}