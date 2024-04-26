package main

import "strconv"

func (t *Token) Value() string {
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
	if t.Type == bulkStringType && t.representNull {
		return string([]byte{firstByte[t.Type]}) + "-1" + TERMINATOR
	}
	res := string([]byte{firstByte[t.Type]})
	res += strconv.Itoa(len(t.SimpleValue)) + TERMINATOR
	if t.Type == verbatimStringType {
		res += "txt:"
	}
	res += t.SimpleValue + TERMINATOR
	return res
}

func nestedEncode(t *Token) string {
	if t.Type == arrayType && t.representNull {
		return string([]byte{firstByte[t.Type]}) + "-1" + TERMINATOR
	}
	res := string([]byte{firstByte[t.Type]})
	n := len(t.NestedValue)
	if t.Type == mapType {
		n /= 2
	}
	res += strconv.Itoa(n) + TERMINATOR
	for _, token := range t.NestedValue {
		res += token.Value() + TERMINATOR
	}
	return res
}
