package main

const TERMINATOR string = "\r\n"

const (
	simpleStringType   string = "simple-string"
	errorType          string = "error"
	integerType        string = "integer"
	bulkStringType     string = "bulk-string"
	arrayType          string = "array"
	nullType           string = "null"
	booleanType        string = "boolean"
	doubleType         string = "double"
	bigNumberType      string = "big-number"
	mapType            string = "map"
	setType            string = "set"
	pushType           string = "push"
	bulkErrorType      string = "bulk-error"
	verbatimStringType string = "verbatim-string"
)

type Token struct {
	Type          string
	SimpleValue   string
	NestedValue   []*Token
	representNull bool // only for null bulk strings and null bulk arrays
}

var firstByteType map[byte]string = map[byte]string{
	'+': simpleStringType,
	'-': errorType,
	':': integerType,
	'$': bulkStringType,
	'*': arrayType,
	'_': nullType,
	'#': booleanType,
	',': doubleType,
	'(': bigNumberType,
	'!': bulkErrorType,
	'=': verbatimStringType,
	'%': mapType,
	'~': setType,
	'>': pushType,
}

var firstByte map[string]byte = map[string]byte{
	simpleStringType:   '+',
	errorType:          '-',
	integerType:        ':',
	bulkStringType:     '$',
	arrayType:          '*',
	nullType:           '_',
	booleanType:        '#',
	doubleType:         ',',
	bigNumberType:      '(',
	bulkErrorType:      '!',
	verbatimStringType: '=',
	mapType:            '%',
	setType:            '~',
	pushType:           '>',
}

const (
	SimpleEncoding string = "simple"
	LengthEncoding string = "length-encoded"
	NestedEncoding string = "nested"
)

var valueEncoding map[string]string = map[string]string{
	simpleStringType:   SimpleEncoding,
	errorType:          SimpleEncoding,
	integerType:        SimpleEncoding,
	bulkStringType:     LengthEncoding,
	arrayType:          NestedEncoding,
	nullType:           SimpleEncoding,
	booleanType:        SimpleEncoding,
	doubleType:         SimpleEncoding,
	bigNumberType:      SimpleEncoding,
	bulkErrorType:      LengthEncoding,
	verbatimStringType: LengthEncoding,
	mapType:            NestedEncoding,
	setType:            NestedEncoding,
	pushType:           NestedEncoding,
}

var nullBulkString Token = Token{Type: bulkStringType, SimpleValue: ""}
