package models

import (
	"github.com/BlocSoc-iitr/Athena/athena/types"
	
)

type ContractABI struct {
	AbiName   string                   `gorm:"primaryKey;column:abi_name"`
	AbiJson   []map[string]interface{} `gorm:"column:abi_json;type:json"`
	Priority  int                      `gorm:"column:priority"`
	DecoderOS string                   `gorm:"column:decoder_os"`
}

func (ContractABI) TableName() string {
	return "contract_abis"
}

type BackfilledRange struct {
	BackfillID   string                 `gorm:"primaryKey;column:backfill_id"`
	DataType     types.BackfillDataType `gorm:"primaryKey;column:data_type"`
	Network      types.SupportedNetwork `gorm:"primaryKey;column:network"`
	StartBlock   int                    `gorm:"primaryKey;column:start_block"`
	EndBlock     int                    `gorm:"primaryKey;column:end_block"`
	FilterData   map[string]interface{} `gorm:"column:filter_data;type:json"`
	MetadataDict map[string]interface{} `gorm:"column:metadata_dict;type:json"`
	DecodedAbis  []string               `gorm:"column:decoded_abis;type:json"`
}

func (BackfilledRange) TableName() string {
	return "backfilled_ranges"
}
