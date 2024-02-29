package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// terminator string = "\r\n"

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
	Type        string
	SimpleValue string
	NestedValue []*Token
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

const (
	SimpleEncoding string = "simple"
	LengthEncoding string = "length-encoded"
	NestedEncoding string = "nested"
)

var inputEncoding map[string]string = map[string]string{
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

func parseInput(input string) ([]*Token, error) {
	reader := bufio.NewReader(strings.NewReader(input))
	tokens := make([]*Token, 0, 100)
	for _, err := reader.Peek(1); err == nil; _, err = reader.Peek(1) {
		token, err := parseToken(reader)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}

func parseToken(read *bufio.Reader) (*Token, error) {
	//returns type, value (if any), and error (if any)
	//read the first byte
	firstByte, err := read.ReadByte()
	if err != nil {
		return nil, err
	}
	tokenType, exists := firstByteType[firstByte]
	if !exists {
		return nil, fmt.Errorf("unknown first byte: %c", firstByte)
	}
	switch inputEncoding[tokenType] {
	case SimpleEncoding:
		val, err := readSimple(read)
		if err != nil {
			return nil, err
		}
		return &Token{Type: tokenType, SimpleValue: val}, nil
	case LengthEncoding:
		val, err := readLengthEncoded(read, tokenType == verbatimStringType)
		if err != nil {
			return nil, err
		}
		return &Token{Type: tokenType, SimpleValue: val}, nil
	case NestedEncoding:
		val, err := readNested(read, tokenType == arrayType)
		if err != nil {
			return nil, err
		}
		return &Token{Type: tokenType, NestedValue: val}, nil
	default:
		return nil, fmt.Errorf("not implemented parsing for type %s", tokenType)
	}
}

func checkNextChar(read *bufio.Reader, target byte) error {
	nxt, err := read.ReadByte()
	if err != nil {
		return err
	}
	if nxt != target {
		return fmt.Errorf("expected '%c', got '%c'", target, nxt)
	}
	return nil
}

func readSimple(read *bufio.Reader) (string, error) {
	val, err := read.ReadString('\r')
	if err != nil {
		return "", err
	}
	if err = checkNextChar(read, '\n'); err != nil {
		return "", err
	}
	return val[:len(val)-1], nil
}

func readLengthEncoded(read *bufio.Reader, isVerbatimString bool) (string, error) {
	lengthString, err := read.ReadString('\r')
	if err != nil {
		return "", err
	}
	if err = checkNextChar(read, '\n'); err != nil {
		return "", err
	}
	length, err := strconv.Atoi(lengthString[:len(lengthString)-1])
	if err != nil {
		return "", err
	}
	if length < 0 {
		return "", fmt.Errorf("invalid length: %d", length)
	}
	if isVerbatimString {
		length += 4 // 3 bytes for encoding, plus 1 for ':'
	}
	var buf []byte = make([]byte, length)
	if _, err = io.ReadFull(read, buf); err != nil {
		return "", err
	}
	if err = checkNextChar(read, '\r'); err != nil {
		return "", err
	}
	if err = checkNextChar(read, '\n'); err != nil {
		return "", err
	}
	return string(buf), nil
}

func readNested(read *bufio.Reader, isMap bool) ([]*Token, error) {
	lengthString, err := read.ReadString('\r')
	if err != nil {
		return nil, err
	}
	if err = checkNextChar(read, '\n'); err != nil {
		return nil, err
	}
	length, err := strconv.Atoi(lengthString[:len(lengthString)-1])
	if err != nil {
		return nil, err
	}
	if length < 0 {
		return nil, fmt.Errorf("invalid length: %d", length)
	}
	if isMap {
		length *= 2 // tokens encoded as [key1, value1, key2, value2, ...]
	}
	tokens := make([]*Token, 0, length)
	for _, err := read.Peek(1); err == nil; _, err = read.Peek(1) {
		token, err := parseToken(read)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}
