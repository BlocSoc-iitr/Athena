package models

import (
	"database/sql"
)

type BlockDataAvailabilityMode string

const (
	Blob     BlockDataAvailabilityMode = "BLOB"
	Calldata BlockDataAvailabilityMode = "CALLDATA"
)

type StarknetFeeUnit string

const (
	Wei StarknetFeeUnit = "WEI"
	Fri StarknetFeeUnit = "FRI"
)

type StarknetTxType string

const (
	Invoke        StarknetTxType = "INVOKE"
	Declare       StarknetTxType = "DECLARE"
	Deploy        StarknetTxType = "DEPLOY"
	DeployAccount StarknetTxType = "DEPLOY_ACCOUNT"
	L1Handler     StarknetTxType = "L1_HANDLER"
)

type TransactionStatus string

const (
	NotReceived  TransactionStatus = "not_received"
	Received     TransactionStatus = "received"
	Rejected     TransactionStatus = "rejected"
	Reverted     TransactionStatus = "reverted"
	AcceptedOnL2 TransactionStatus = "accepted_on_l2"
	AcceptedOnL1 TransactionStatus = "accepted_on_l1"
)

type DecodedOperation struct {
	OperationName   string                 `json:"operation_name"`
	OperationParams map[string]interface{} `json:"operation_params"`
}

type Block struct {
	AbstractBlock
	ParentHash             string                    `gorm:"column:parent_hash;type:text;not null"`
	StateRoot              string                    `gorm:"column:state_root;type:text;not null"`
	SequencerAddress       string                    `gorm:"column:sequencer_address;type:text;not null"`
	L1GasPriceWei          float64                   `gorm:"column:l1_gas_price_wei;type:numeric;not null"`
	L1GasPriceFri          float64                   `gorm:"column:l1_gas_price_fri;type:numeric;not null"`
	L1DataGasPriceWei      sql.NullFloat64           `gorm:"column:l1_data_gas_price_wei;type:numeric"`
	L1DataGasPriceFri      sql.NullFloat64           `gorm:"column:l1_data_gas_price_fri;type:numeric"`
	L1DataAvailabilityMode BlockDataAvailabilityMode `gorm:"column:l1_da_mode;type:varchar(10);not null"`
	StarknetVersion        string                    `gorm:"column:starknet_version;type:text;not null"`
	TransactionCount       int                       `gorm:"column:transaction_count;type:int;not null"`
	TotalFee               float64                   `gorm:"column:total_fee;type:numeric;not null"`
}

type DefaultEvent struct {
	AbstractEvent
	Keys          []string               `gorm:"column:keys;type:jsonb;not null"`
	Data          []string               `gorm:"column:data;type:jsonb"`
	ClassHash     sql.NullString         `gorm:"column:class_hash;type:text"`
	EventName     sql.NullString         `gorm:"column:event_name;type:text;index"`
	DecodedParams map[string]interface{} `gorm:"column:decoded_params;type:jsonb"`
}

type Transaction struct {
	AbstractTransaction
	Type                  StarknetTxType         `gorm:"column:type;type:varchar(20);not null"`
	Nonce                 int                    `gorm:"column:nonce;type:int;not null"`
	Signature             []string               `gorm:"column:signature;type:jsonb;not null"`
	Version               int                    `gorm:"column:version;type:int;not null"`
	Status                TransactionStatus      `gorm:"column:status;type:varchar(20);not null"`
	MaxFee                float64                `gorm:"column:max_fee;type:numeric;not null"`
	ActualFee             float64                `gorm:"column:actual_fee;type:numeric;not null"`
	FeeUnit               StarknetFeeUnit        `gorm:"column:fee_unit;type:varchar(5);not null"`
	ExecutionResources    map[string]interface{} `gorm:"column:execution_resources;type:jsonb;not null"`
	Tip                   float64                `gorm:"column:tip;type:numeric"`
	ResourceBounds        map[string]int         `gorm:"column:resource_bounds;type:jsonb"`
	PaymasterData         []string               `gorm:"column:paymaster_data;type:jsonb"`
	AccountDeploymentData []string               `gorm:"column:account_deployment_data;type:jsonb"`
	ContractAddress       sql.NullString         `gorm:"column:contract_address;type:text;index"`
	Selector              string                 `gorm:"column:selector;type:text;not null"`
	Calldata              []string               `gorm:"column:calldata;type:jsonb;not null"`
	ClassHash             sql.NullString         `gorm:"column:class_hash;type:text"`
	UserOperations        []DecodedOperation     `gorm:"column:user_operations;type:jsonb"`
	RevertError           sql.NullString         `gorm:"column:revert_error;type:text"`
}

type ERC20Transfer struct {
	AbstractERC20Transfer
}
