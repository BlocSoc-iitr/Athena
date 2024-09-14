package types

import (
	"fmt"
	"math"
	"math/big"
	"strings"
)

// AbiMemberType represents the type of ABI member.
type AbiMemberType string

const (
	Abi_Function   AbiMemberType = "function"
	AbiL1Handler   AbiMemberType = "l1_handler"
	AbiStruct      AbiMemberType = "struct"
	AbiConstructor AbiMemberType = "constructor"
	Abi_Event      AbiMemberType = "event"
	Abi_Interface  AbiMemberType = "interface"
	AbiImpl        AbiMemberType = "impl"
	AbiTypeDef     AbiMemberType = "type_def"
)

type StarknetAbiEventKind int

const (
	EventEnum StarknetAbiEventKind = iota
	EventStruct
	EventData
	EventNested
	EventKey
	EventFlat
)

func (e StarknetAbiEventKind) String() string {
	return [...]string{"enum", "struct", "data", "nested", "key", "flat"}[e]
}

type StarknetCoreType int

const (
	U8 StarknetCoreType = iota + 1
	U16
	U32
	U64
	U128
	U256
	Bool
	Felt
	ContractAddress
	EthAddress
	ClassHash
	StorageAddress
	Bytes31
	NoneType
)

func (t StarknetCoreType) IDStr() string {
	return t.String()
}

func (t StarknetCoreType) String() string {
	names := [...]string{
		"", "U8", "U16", "U32", "U64", "U128", "U256", "Bool",
		"Felt", "ContractAddress", "EthAddress", "ClassHash",
		"StorageAddress", "Bytes31", "NoneType",
	}

	if t < U8 || t > NoneType {
		return "Unknown"
	}

	return names[t]
}

func IntFromString(typeStr string) (StarknetCoreType, error) {
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

func (t StarknetCoreType) MaxValue() (*big.Int, error) {
	switch t {
	case U8:
		return big.NewInt(math.MaxUint8), nil
	case U16:
		return big.NewInt(math.MaxUint16), nil
	case U32:
		return new(big.Int).SetUint64(math.MaxUint32), nil
	case U64:
		return new(big.Int).SetUint64(math.MaxUint64), nil
	case U128:
		return new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)), nil
	case U256:
		return new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)), nil
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

type StarknetType interface {
	IDStr() string
}

type StarknetArray struct {
	InnerType StarknetType
}

func (sa StarknetArray) IDStr() string {
	return fmt.Sprintf("[%s]", sa.InnerType.IDStr())
}

type StarknetOption struct {
	InnerType StarknetType
}

func (so StarknetOption) IDStr() string {
	return fmt.Sprintf("Option[%s]", so.InnerType.IDStr())
}

type StarknetEnum struct {
	Name     string
	Variants []Variant
}

func (se StarknetEnum) IDStr() string {
	var membersStr []string
	for _, member := range se.Variants {
		membersStr = append(membersStr, member.VariantType.IDStr())
	}
	return fmt.Sprintf("Option[%s]", strings.Join(membersStr, ","))
}

type Variant struct {
	VariantName string
	VariantType StarknetType
}

type StarknetTuple struct {
	Members []StarknetType
}

func (st StarknetTuple) IDStr() string {
	var membersStr []string
	for _, member := range st.Members {
		membersStr = append(membersStr, member.IDStr())
	}
	return fmt.Sprintf("(%s)", strings.Join(membersStr, ","))
}

type AbiParameter struct {
	Name string
	Type StarknetType
}

func (ap AbiParameter) IDStr() string {
	return fmt.Sprintf("%s:%s", ap.Name, ap.Type.IDStr())
}

type StarknetStruct struct {
	Name    string
	Members []AbiParameter
}

func (ss StarknetStruct) IDStr() string {
	var membersStr []string
	for _, member := range ss.Members {
		membersStr = append(membersStr, member.IDStr())
	}
	return fmt.Sprintf("{%s}", strings.Join(membersStr, ","))
}
