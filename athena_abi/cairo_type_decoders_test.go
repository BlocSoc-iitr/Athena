package athena_abi

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayDecoding(t *testing.T) {
	tests := []struct {
		starknetType StarknetType
		calldata     []*big.Int
		decoded      []*big.Int
	}{
		{StarknetArray{U256}, []*big.Int{big.NewInt(0)}, []*big.Int{}},
		{StarknetArray{U256}, []*big.Int{big.NewInt(2), big.NewInt(16), big.NewInt(0), big.NewInt(48), big.NewInt(0)}, []*big.Int{big.NewInt(16), big.NewInt(48)}},
	}

	for _, test := range tests {
		_calldata := make([]*big.Int, len(test.calldata))
		copy(_calldata, test.calldata)
		decodedValues, err := DecodeFromTypes([]StarknetType{test.starknetType}, &_calldata)
		assert.Equal(t, nil, err)
		for i := 0; i < len(test.decoded); i++ {
			assert.Equal(t, test.decoded[i], decodedValues[i])
		}
		// encodedCalldata, err := EncodeFromTypes([]StarknetType{test.starknetType}, []interface{}{test.decoded})
		// assert.Equal(t, err, nil)
		// assert.Equal(t, test.calldata, encodedCalldata)
		// assert.Equal(t, 0, len(_calldata))
	}
}
