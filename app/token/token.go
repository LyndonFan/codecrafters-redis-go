package token

const TERMINATOR string = "\r\n"

type TokenType string

const (
	SimpleStringType   TokenType = "simple-string"
	ErrorType          TokenType = "error"
	IntegerType        TokenType = "integer"
	BulkStringType     TokenType = "bulk-string"
	ArrayType          TokenType = "array"
	NullType           TokenType = "null"
	BooleanType        TokenType = "boolean"
	DoubleType         TokenType = "double"
	BigNumberType      TokenType = "big-number"
	MapType            TokenType = "map"
	SetType            TokenType = "set"
	PushType           TokenType = "push"
	BulkErrorType      TokenType = "bulk-error"
	VerbatimStringType TokenType = "verbatim-string"
)

type Token struct {
	Type                    TokenType
	SimpleValue             string
	NestedValue             []*Token
	representNull           bool // only for null bulk strings and null bulk arrays
	stripTrailingTerminator bool // only for RDB files
}

var FirstByteType map[byte]TokenType = map[byte]TokenType{
	'+': SimpleStringType,
	'-': ErrorType,
	':': IntegerType,
	'$': BulkStringType,
	'*': ArrayType,
	'_': NullType,
	'#': BooleanType,
	',': DoubleType,
	'(': BigNumberType,
	'!': BulkErrorType,
	'=': VerbatimStringType,
	'%': MapType,
	'~': SetType,
	'>': PushType,
}

var firstByte map[TokenType]byte = map[TokenType]byte{
	SimpleStringType:   '+',
	ErrorType:          '-',
	IntegerType:        ':',
	BulkStringType:     '$',
	ArrayType:          '*',
	NullType:           '_',
	BooleanType:        '#',
	DoubleType:         ',',
	BigNumberType:      '(',
	BulkErrorType:      '!',
	VerbatimStringType: '=',
	MapType:            '%',
	SetType:            '~',
	PushType:           '>',
}

// TokenEncoding is the encoding type of a Redis token. It can be one of:
//   - SimpleEncoding: simple string, error, integer, boolean, null or double
//   - LengthEncoding: bulk string or array
//   - NestedEncoding: nested array
type TokenEncoding string

const (
	SimpleEncoding TokenEncoding = "simple"
	LengthEncoding TokenEncoding = "length-encoded"
	NestedEncoding TokenEncoding = "nested"
)

var ValueEncoding map[TokenType]TokenEncoding = map[TokenType]TokenEncoding{
	SimpleStringType:   SimpleEncoding,
	ErrorType:          SimpleEncoding,
	IntegerType:        SimpleEncoding,
	BulkStringType:     LengthEncoding,
	ArrayType:          NestedEncoding,
	NullType:           SimpleEncoding,
	BooleanType:        SimpleEncoding,
	DoubleType:         SimpleEncoding,
	BigNumberType:      SimpleEncoding,
	BulkErrorType:      LengthEncoding,
	VerbatimStringType: LengthEncoding,
	MapType:            NestedEncoding,
	SetType:            NestedEncoding,
	PushType:           NestedEncoding,
}
