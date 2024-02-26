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
	NestedValue []Token
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

var isSimple map[string]bool = map[string]bool{
	simpleStringType:   true,
	errorType:          true,
	integerType:        true,
	bulkStringType:     false,
	arrayType:          false,
	nullType:           true,
	booleanType:        true,
	doubleType:         true,
	bigNumberType:      true,
	bulkErrorType:      false,
	verbatimStringType: true,
	mapType:            false,
	setType:            false,
	pushType:           false,
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
	if isSimple[tokenType] {
		val, err := readSimple(read)
		if err != nil {
			return nil, err
		}
		return &Token{Type: tokenType, SimpleValue: val}, nil
	}
	return nil, fmt.Errorf("not implemented parsing for type %s", tokenType)
}

func readSimple(read *bufio.Reader) (string, error) {
	val, err := read.ReadString('\r')
	if err != nil {
		return "", err
	}
	nxt, err := read.ReadByte()
	if err != nil {
		return "", err
	}
	if nxt != '\n' {
		return "", fmt.Errorf("expected '\\n', got '%c'", nxt)
	}
	return val[:len(val)-1], nil
}

func readLengthEncoded(read *bufio.Reader) (string, error) {
	lengthString, err := read.ReadString('\r')
	if err != nil {
		return "", err
	}
	nxt, err := read.ReadByte()
	if err != nil {
		return "", err
	}
	if nxt != '\n' {
		return "", fmt.Errorf("expected '\\n', got '%c'", nxt)
	}
	length, err := strconv.Atoi(lengthString[:len(lengthString)-1])
	if err != nil {
		return "", err
	}
	if length < 0 {
		return "", fmt.Errorf("invalid length: %d", length)
	}
	var buf []byte = make([]byte, length)
	if _, err = io.ReadFull(read, buf); err != nil {
		return "", err
	}
	nxt, err = read.ReadByte()
	if err != nil {
		return "", err
	}
	if nxt != '\n' {
		return "", fmt.Errorf("expected '\\n', got '%c'", nxt)
	}
	return string(buf), nil
}
