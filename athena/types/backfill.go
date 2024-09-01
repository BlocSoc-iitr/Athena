package types

import (
	"time"
)

// BlockProtocol is an interface that ensures implementing structs have block_number and timestamp fields.
type BlockProtocol interface {
	BlockNumber() int
	Timestamp() int64
}

// BackfillDataType is an enum that represents different types of blockchain data that can be backfilled.
type BackfillDataType int

const (
	FullBlocks BackfillDataType = iota
	Blocks
	Transactions
	Transfers
	SpotPrices
	Prices
	Events
	Traces
)

func (b BackfillDataType) String() string {
	return []string{
		"Full Blocks",
		"Blocks",
		"Transactions",
		"Transfers",
		"Spot Prices",
		"Prices",
		"Events",
		"Traces",
	}[b]
}

func (b BackfillDataType) Pretty() string {
	switch b {
	case FullBlocks:
		return "Full Blocks"
	case SpotPrices:
		return "Spot-Prices"
	default:
		return b.String()
	}
}

// DataSources is an enum storing supported backfill data sources.
type DataSources int

const (
	JSONRPC DataSources = iota
	Etherscan
)

func (d DataSources) String() string {
	return []string{
		"json_rpc",
		"etherscan",
	}[d]
}

// SupportedNetwork is an enum that represents supported networks that can be backfilled.
type SupportedNetwork int

const (
	StarkNet SupportedNetwork = iota
	Ethereum
	ZkSyncEra
)

func (s SupportedNetwork) String() string {
	return []string{
		"StarkNet",
		"Ethereum",
		"zkSync Era",
	}[s]
}

func (s SupportedNetwork) Pretty() string {
	switch s {
	case StarkNet:
		return "StarkNet"
	case ZkSyncEra:
		return "zkSync Era"
	default:
		return s.String()
	}
}

// BlockTimestamp is a struct that efficiently stores block timestamps.
type BlockTimestamp struct {
	BlockNumber int
	Timestamp   time.Time
}
