package types

import (
	"fmt"
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

func (t StarknetCoreType) IDString() string {
	return t.String()
}

func (t StarknetCoreType) MaxValue() (*big.Int, error) {
	switch t {
	case U8, U16, U32, U64, U128, U256:
		bytes := int(t)
		return new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(8*bytes)), big.NewInt(1)), nil
	case Felt, ContractAddress, ClassHash:
		value := new(big.Int)
		value.Exp(big.NewInt(2), big.NewInt(251), nil).
			Add(value, new(big.Int).Mul(big.NewInt(17), new(big.Int).Exp(big.NewInt(2), big.NewInt(192), nil))).
			Add(value, big.NewInt(1))
		return value, nil
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
	return fmt.Sprintf("Option[%s]")
}

// Variant represents a tuple of a variant name and type.
type Variant struct {
	VariantName string
	VariantType StarknetType
}

// IDStr returns the string representation of the StarknetEnum.
//
//	func (se StarknetEnum) IDStr() string {
//		var variantsStr []string
//		var none_type StarknetCoreType = NoneType
//		for _, variant := range se.Variants {
//			if variant.VariantType == none_type {
//				variantsStr = append(variantsStr, fmt.Sprintf("'%s'", variant.VariantName))
//			} else {
//				variantsStr = append(variantsStr, fmt.Sprintf("%s:%s", variant.VariantName, variant.VariantType.IDStr()))
//			}
//		}
//		return fmt.Sprintf("Enum[%s]", strings.Join(variantsStr, ","))
//	}
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
