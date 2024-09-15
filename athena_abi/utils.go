package athena_abi

import (
	"math/big"

	"golang.org/x/crypto/sha3"
)

func bigIntToBytes(value big.Int, length int) []byte {
	b := make([]byte, length)
	for i := length - 1; i >= 0; i-- {
		b[i] = byte(new(big.Int).And(&value, big.NewInt(0xff)).Int64())
		value = *value.Rsh(&value, 8)
	}
	return b
}

func starknetKeccak(data []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)

	var masked big.Int
	masked.SetBytes(hash.Sum(nil))
	masked = *masked.And(&masked, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 250), big.NewInt(1)))
	return bigIntToBytes(masked, 32)
}
