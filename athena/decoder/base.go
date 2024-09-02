package decoder
type DecodedFuncDataclass struct {
	ABIName string
	Name    string
	Input   map[string]interface{} // Use map for dict in Python
	Output  []interface{}          // Use slice for list in Python
}
type DecodedEventDataclass struct {
	ABIName string
	Name    string
	Data    map[string]interface{} // Use map for dict in Python
}
type AbiFunctionDecoder interface {
	Name() string
	Signature() []byte
	ABIName() string
	Priority() int
	Decode(calldata [][]byte, result [][]byte) (*DecodedFuncDataclass, error)//returned DecodedEventDataclass in golang 
	IDStr(fullSignature bool) string
}
type AbiEventDecoder interface {
	Name() string
	Signature() []byte
	ABIName() string
	Priority() int
	IndexedParams() int
	Decode(data [][]byte, keys [][]byte) (*DecodedEventDataclass, error)
	IDStr(fullSignature bool) string
}