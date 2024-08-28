package main

type DecodedFunction struct {
	ABIName           string // Function Decoding Result
	Name              string
	FunctionSignature string
	Input             map[string]interface{} // map representing a dictionary of string to any type
	Output            []interface{}          // slice representing a list of any type
}

type DecodedEvent struct {
	ABIName        string // Event Decoding Result
	Name           string
	EventSignature string
	Data           map[string]interface{} // map representing a dictionary of string to any type
}

type DecodedTrace struct {
	ABIName       string // Decoded Trace with decoded inputs and outputs
	Name          string
	Signature     string
	DecodedInput  map[string]interface{} // map representing a dictionary of string to any type
	DecodedOutput map[string]interface{} // map representing a dictionary of string to any type
}
