package athenaabi

import (
    "fmt"
    "math/big"
    "strings"
)

type StarknetType interface {
    idStr() string
}

type abiMemberType string

const (
    Function   abiMemberType = "function"
    L1Handler  abiMemberType = "l1Handler"
    AbiStruct     abiMemberType = "struct"
    Constructor abiMemberType = "constructor"
    Event      abiMemberType = "event"
    Interface  abiMemberType = "interface"
    Impl       abiMemberType = "impl"
    TypeDef    abiMemberType = "typeDef" // Internal Definition: typeDef = Union[struct, enum]
)

type starknetAbiEventKind string

const (
    Enum   starknetAbiEventKind = "enum"
    Struct starknetAbiEventKind = "struct"
    Data   starknetAbiEventKind = "data"
    Nested starknetAbiEventKind = "nested"
    Key    starknetAbiEventKind = "key"
    Flat   starknetAbiEventKind = "flat"
)

type starknetCoreType int

const (
    U8             starknetCoreType = 1
    U16            starknetCoreType = 2
    U32            starknetCoreType = 4
    U64            starknetCoreType = 8
    U128           starknetCoreType = 16
    U256           starknetCoreType = 32
	// Random Enum values for the rest
    Bool           starknetCoreType = 3
    Felt           starknetCoreType = 5
    ContractAddress starknetCoreType = 6
    EthAddress     starknetCoreType = 7
    ClassHash      starknetCoreType = 9
    StorageAddress starknetCoreType = 10
    Bytes31        starknetCoreType = 11
    NoneType       starknetCoreType = 12
)

func (t starknetCoreType) String() string {
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

// intFromString converts a string like 'u16' to the corresponding starknetCoreType
func intFromString(typeStr string) (starknetCoreType, error) {
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

func (t starknetCoreType) idStr() string {
    return t.String()
}

// maxValue returns the maximum value for the corresponding starknetCoreType
func (t starknetCoreType) maxValue() (*big.Int, error) {
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
type starknetArray struct {
	InnerType StarknetType
}

func (t starknetArray) idStr() string {
	return fmt.Sprintf("[%s]", t.InnerType.idStr())
}

type starknetOption struct {
	InnerType StarknetType
}

func (t starknetOption) idStr() string {
	return fmt.Sprintf("Option[%s]", t.InnerType.idStr())
}

type starknetNonZero struct {
	InnerType StarknetType
}

func (t starknetNonZero) idStr() string {
	return fmt.Sprintf("NonZero[%s]", t.InnerType.idStr())
}

type starknetEnum struct {
    Name     string
    Variants []struct {
        Name  string
        Type  StarknetType
    }
}

// idStr returns the string representation of the enum in the format "Enum[variant-name:variant-type,...]"
func (e starknetEnum) idStr() string {
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

type starknetTuple struct {
	Members []StarknetType
}

func (e starknetTuple) idStr() string {
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

type starknetStruct struct {
	Name    string
	Members []abiParameter
}

func (s starknetStruct) idStr() string {
	var members []string
	for _, member := range s.Members {
		members = append(members, fmt.Sprintf("%s:%s", member.Name, member.Type.idStr()))
	}
	return fmt.Sprintf("{%s}", strings.Join(members, ","))
}
