package token

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func ParseInput(input string) ([]*Token, error) {
	reader := bufio.NewReader(strings.NewReader(input))
	tokens := make([]*Token, 0, 100)
	_, peekErr := reader.Peek(1)
	for peekErr == nil {
		token, err := parseToken(reader)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
		_, peekErr = reader.Peek(1)
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
	tknType, exists := FirstByteType[firstByte]
	if !exists {
		return nil, fmt.Errorf("unknown first byte: %c", firstByte)
	}
	switch ValueEncoding[tknType] {
	case SimpleEncoding:
		val, err := readSimple(read)
		if err != nil {
			return nil, err
		}
		return &Token{Type: tknType, SimpleValue: val}, nil
	case LengthEncoding:
		val, err := readLengthEncoded(read, tknType == VerbatimStringType)
		if err != nil {
			return nil, err
		}
		return &Token{Type: tknType, SimpleValue: val}, nil
	case NestedEncoding:
		val, err := readNested(read, tknType == MapType)
		if err != nil {
			return nil, err
		}
		return &Token{Type: tknType, NestedValue: val}, nil
	default:
		return nil, fmt.Errorf("not implemented parsing for type %s", tknType)
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
	_, peekErr := read.Peek(1)
	for peekErr == nil {
		token, err := parseToken(read)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
		_, peekErr = read.Peek(1)
	}
	return tokens, nil
}
