package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Base struct {
	gorm.Model
}

type AbstractBlock struct {
	BlockNumber uint64 `gorm:"column:block_number;primaryKey"`
	BlockHash   string `gorm:"column:block_hash;type:text"`
	Timestamp   int64  `gorm:"column:timestamp;type:bigint"`
}

type AbstractTransaction struct {
	TransactionHash  string   `gorm:"column:transaction_hash;primaryKey;type:varchar(66);index"`
	BlockNumber      uint64   `gorm:"column:block_number;type:bigint;index"`
	TransactionIndex int      `gorm:"column:transaction_index;type:int"`
	Timestamp        int64    `gorm:"column:timestamp;type:bigint"`
	GasUsed          *float64 `gorm:"column:gas_used;type:numeric;nullable:true"`
}

type AbstractEvent struct {
	BlockNumber      uint64 `gorm:"column:block_number;type:bigint;index"`
	EventIndex       int    `gorm:"column:event_index;type:int"`
	TransactionIndex int    `gorm:"column:transaction_index;type:int"`
	ContractAddress  string `gorm:"column:contract_address;type:varchar(42);index"`
}

type AbstractTrace struct {
	BlockNumber      uint64         `gorm:"column:block_number;type:bigint;index"`
	TransactionHash  string         `gorm:"column:transaction_hash;type:text;index"`
	TransactionIndex int            `gorm:"column:transaction_index;type:int"`
	TraceAddress     datatypes.JSON `gorm:"column:trace_address;type:json"`
	GasUsed          int64          `gorm:"column:gas_used;type:bigint"`
	Error            string         `gorm:"column:error;type:text"`
}

type AbstractERC20Transfer struct {
	BlockNumber      uint64 `gorm:"column:block_number;type:bigint;index"`
	TransactionHash  string `gorm:"column:transaction_hash;type:text"`
	TransactionIndex int    `gorm:"column:transaction_index;type:int"`
	EventIndex       int    `gorm:"column:event_index;type:int"`

	TokenAddress string   `gorm:"column:token_address;type:text;index"`
	FromAddress  string   `gorm:"column:from_address;type:text;index"`
	ToAddress    string   `gorm:"column:to_address;type:text;index"`
	Value        *float64 `gorm:"column:value;type:numeric(78, 0)"`
}