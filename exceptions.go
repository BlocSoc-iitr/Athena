package athena

// ArchivalNodeRequired is raised when archival features of an RPC connection are inadequate,
// and simulations need to be reworked, or node infrastructure needs to be upgraded.
type ArchivalNodeRequired struct{}

func (e *ArchivalNodeRequired) Error() string {
	return "archival features of the RPC connection are inadequate; simulations need reworking or node infrastructure upgrade"
}

// BackfillError is raised when issues occur with backfilling data.
type BackfillError struct{}

func (e *BackfillError) Error() string {
	return "issues occurred with backfilling data"
}

// BackfillRateLimitError is raised when gateway rate limits are implemented by the remote host.
type BackfillRateLimitError struct {
	BackfillError
}

func (e *BackfillRateLimitError) Error() string {
	return "gateway rate limits are implemented by the remote host"
}

// BackfillHostError is raised when the remote host returns an error, fails to provide correct data, or a timeout occurs.
type BackfillHostError struct {
	BackfillError
}

func (e *BackfillHostError) Error() string {
	return "remote host error, incorrect data, or timeout occurred"
}

// DatabaseError is raised when issues occur with database operations.
type DatabaseError struct{}

func (e *DatabaseError) Error() string {
	return "issues occurred with database operations"
}

// DecodingError is raised when issues occur with input decoding during data backfills.
type DecodingError struct {
	Message string
}

// Error implements the error interface for DecodingError.
func (e *DecodingError) Error() string {
	return e.Message
}

// NewDecodingError creates a new DecodingError with a custom message.
func NewDecodingError(message string) error {
	return &DecodingError{
		Message: message,
	}
}

// UniswapV3Revert is raised when a pool action causes behavior that would throw a revert on-chain.
type UniswapV3Revert struct{}

func (e *UniswapV3Revert) Error() string {
	return `Uniswap V3 Revert: 
    - Ticks exceed the maximum tick value of 887272 or the minimum tick value of -887272
    - uint values are set to a negative value
    - operations cause uints and ints to overflow or underflow
    - Minting & Burning positions with zero liquidity
    - Executing Swaps with invalid sqrt_price bounds or zero input`
}

// FullMathRevert is raised when the result of (a * b) / c overflows the maximum value of a uint256.
type FullMathRevert struct{}

func (e *FullMathRevert) Error() string {
	return "result of (a * b) / c overflows the maximum value of a uint256; triggers UniswapV3Revert"
}

// TickMathRevert is raised when a tick value is out of bounds or a sqrt_price exceeds the maximum sqrt_price.
type TickMathRevert struct{}

func (e *TickMathRevert) Error() string {
	return "tick value out of bounds or sqrt_price exceeds the maximum sqrt_price"
}

// SqrtPriceMathRevert is raised when a sqrt_price value is out of bounds or the inputs to a price calculation are invalid.
type SqrtPriceMathRevert struct{}

func (e *SqrtPriceMathRevert) Error() string {
	return "sqrt_price value out of bounds or invalid price calculation inputs"
}

// OracleError is raised when the PricingOracle fails to return a valid price.
type OracleError struct{}

func (e *OracleError) Error() string {
	return `PricingOracle failed to return a valid price:
    - Verify all addresses are checksummed with eth_utils.to_checksum_address
    - Check the token address on Etherscan to ensure it is a valid ERC20 token
    - Double-check that RPC connection is working & node is synced to the correct chain`
}
