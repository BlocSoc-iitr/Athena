package athenaabi

import (
	"fmt"
	"math/big"
	"strings"
)

type StarknetType interface {
	idStr() string
}

type AbiMemberType string

const (
	Function    AbiMemberType = "function"
	L1Handler   AbiMemberType = "l1Handler"
	AbiStruct   AbiMemberType = "struct"
	Constructor AbiMemberType = "constructor"
	Event       AbiMemberType = "event"
	Interface   AbiMemberType = "interface"
	Impl        AbiMemberType = "impl"
	TypeDef     AbiMemberType = "typeDef" // Internal Definition: typeDef = Union[struct, enum]
)

type StarknetAbiEventKind string

const (
	Enum   StarknetAbiEventKind = "enum"
	Struct StarknetAbiEventKind = "struct"
	Data   StarknetAbiEventKind = "data"
	Nested StarknetAbiEventKind = "nested"
	Key    StarknetAbiEventKind = "key"
	Flat   StarknetAbiEventKind = "flat"
)

type StarknetCoreType int

const (
	U8   StarknetCoreType = 1
	U16  StarknetCoreType = 2
	U32  StarknetCoreType = 4
	U64  StarknetCoreType = 8
	U128 StarknetCoreType = 16
	U256 StarknetCoreType = 32
	// Random Enum values for the rest
	Bool            StarknetCoreType = 3
	Felt            StarknetCoreType = 5
	ContractAddress StarknetCoreType = 6
	EthAddress      StarknetCoreType = 7
	ClassHash       StarknetCoreType = 9
	StorageAddress  StarknetCoreType = 10
	Bytes31         StarknetCoreType = 11
	NoneType        StarknetCoreType = 12
)

func (t StarknetCoreType) String() string {
	switch t {
	case U8:
		return "U8"
	case U16:
		return "U16"
	case U32:
		return "U32"
	case U64:
		return "U64"
	case U128:
		return "U128"
	case U256:
		return "U256"
	case Bool:
		return "Bool"
	case Felt:
		return "Felt"
	case ContractAddress:
		return "ContractAddress"
	case EthAddress:
		return "EthAddress"
	case ClassHash:
		return "ClassHash"
	case StorageAddress:
		return "StorageAddress"
	case Bytes31:
		return "Bytes31"
	case NoneType:
		return "NoneType"
	default:
		return "Unknown"
	}
}

// intFromString converts a string like 'u16' to the corresponding StarknetCoreType
func intFromString(typeStr string) (StarknetCoreType, error) {
	switch strings.ToLower(typeStr) {
	case "u8":
		return U8, nil
	case "u16":
		return U16, nil
	case "u32":
		return U32, nil
	case "u64":
		return U64, nil
	case "u128":
		return U128, nil
	case "u256":
		return U256, nil
	default:
		return 0, fmt.Errorf("invalid integer type: %s", typeStr)
	}
}

func (t StarknetCoreType) idStr() string {
	return t.String()
}

// maxValue returns the maximum value for the corresponding StarknetCoreType
func (t StarknetCoreType) maxValue() (*big.Int, error) {
	switch t {
	case U8, U16, U32, U64, U128, U256:
		return new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(8*t)), big.NewInt(1)), nil
	case Felt, ContractAddress, ClassHash:
		// Felt Prime = 2^251 + 17*2^192 + 1
		// ContractAddress is computed by the pedersen hash function and ClassHash is computed by the posiedon hash function, both of which are taken modulo Felt Prime.
		return new(big.Int).Add(big.NewInt(1), new(big.Int).Add(new(big.Int).Mul(big.NewInt(17), new(big.Int).Lsh(big.NewInt(1), 192)), new(big.Int).Lsh(big.NewInt(1), 251))), nil
	case EthAddress:
		return new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 160), big.NewInt(1)), nil
	case Bytes31:
		return new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 248), big.NewInt(1)), nil
	default:
		return nil, fmt.Errorf("cannot get max value for type: %s", t.String())
	}
}

// Dataclass representing a Starknet ABI Array. Both core::array::Array and core::array::Span are mapped to this dataclass since their ABI Encoding & Decoding are identical
type StarknetArray struct {
	InnerType StarknetType
}

func (t StarknetArray) idStr() string {
	return fmt.Sprintf("[%s]", t.InnerType.idStr())
}

type StarknetOption struct {
	InnerType StarknetType
}

func (t StarknetOption) idStr() string {
	return fmt.Sprintf("Option[%s]", t.InnerType.idStr())
}

type StarknetNonZero struct {
	InnerType StarknetType
}

func (t StarknetNonZero) idStr() string {
	return fmt.Sprintf("NonZero[%s]", t.InnerType.idStr())
}

type StarknetEnum struct {
	Name     string
	Variants []struct {
		Name string
		Type StarknetType
	}
}

// idStr returns the string representation of the enum in the format "Enum[variant-name:variant-type,...]"
func (e StarknetEnum) idStr() string {
	var variants []string
	for _, variant := range e.Variants {
		var variantStr string
		if variant.Type.idStr() == "NoneType" {
			variantStr = variant.Name
		} else {
			variantStr = fmt.Sprintf("%s:%s", variant.Name, variant.Type.idStr())
		}
		variants = append(variants, variantStr)
	}
	return fmt.Sprintf("Enum[%s]", strings.Join(variants, ","))
}

type StarknetTuple struct {
	Members []StarknetType
}

func (e StarknetTuple) idStr() string {
	var members []string
	for _, member := range e.Members {
		members = append(members, member.idStr())
	}
	return fmt.Sprintf("(%s)", strings.Join(members, ","))
}

type abiParameter struct {
	Name string
	Type StarknetType
}

func (p abiParameter) idStr() string {
	return fmt.Sprintf("%s:%s", p.Name, p.Type.idStr())
}

type StarknetStruct struct {
	Name    string
	Members []abiParameter
}

func (s StarknetStruct) idStr() string {
	var members []string
	for _, member := range s.Members {
		members = append(members, fmt.Sprintf("%s:%s", member.Name, member.Type.idStr()))
	}
	return fmt.Sprintf("{%s}", strings.Join(members, ","))
}
