package main

import "strconv"

func (t *Token) EncodedString() string {
	switch valueEncoding[t.Type] {
	case SimpleEncoding:
		return simpleEncode(t)
	case LengthEncoding:
		return lengthEncode(t)
	case NestedEncoding:
		return nestedEncode(t)
	default:
		return ""
	}
}

func simpleEncode(t *Token) string {
	return string([]byte{firstByte[t.Type]}) + t.SimpleValue + TERMINATOR
}

func lengthEncode(t *Token) string {
	if t.Type == BulkStringType && t.representNull {
		return string([]byte{firstByte[t.Type]}) + "-1" + TERMINATOR
	}
	res := string([]byte{firstByte[t.Type]})
	res += strconv.Itoa(len(t.SimpleValue)) + TERMINATOR
	if t.Type == VerbatimStringType {
		res += "txt:"
	}
	res += t.SimpleValue + TERMINATOR
	return res
}

func nestedEncode(t *Token) string {
	if t.Type == ArrayType && t.representNull {
		return string([]byte{firstByte[t.Type]}) + "-1" + TERMINATOR
	}
	res := string([]byte{firstByte[t.Type]})
	n := len(t.NestedValue)
	if t.Type == MapType {
		n /= 2
	}
	res += strconv.Itoa(n) + TERMINATOR
	for _, token := range t.NestedValue {
		res += token.EncodedString() + TERMINATOR
	}
	return res
}
