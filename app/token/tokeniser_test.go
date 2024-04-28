package main

import (
	"bufio"
	"strings"
	"testing"
)

// returns a string without error if the input ends with '\r\n'
func TestReadSimple_ValidInput(t *testing.T) {
	input := "hello world\r\n"
	reader := bufio.NewReader(strings.NewReader(input))
	result, err := readSimple(reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", result)
	}
}

// returns an error if the input ends with '\n' instead of '\r\n'
func TestReadSimple_InvalidInput(t *testing.T) {
	input := "hello world\n"
	reader := bufio.NewReader(strings.NewReader(input))
	_, err := readSimple(reader)
	if err == nil {
		t.Error("Expected an error, got nil")
	}
}

func TestParseInput_ValidSimpleInput(t *testing.T) {
	input := "+hello world\r\n"
	tokens, err := ParseInput(input)
	expectedTokens := []*Token{{SimpleStringType, "hello world", nil, false}}
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(tokens) != len(expectedTokens) {
		t.Errorf("Expected %d token(s), got %d", len(expectedTokens), len(tokens))
	}
	for i := 0; i < len(expectedTokens); i++ {
		if tokens[i].Type != expectedTokens[i].Type {
			t.Errorf("Expected token %d to have type %s, got %s", i, expectedTokens[i].Type, tokens[i].Type)
		}
		if tokens[i].SimpleValue != expectedTokens[i].SimpleValue {
			t.Errorf("Expected token %d to have simple value %s, got %s", i, expectedTokens[i].SimpleValue, tokens[i].SimpleValue)
		}
		if len(tokens[i].NestedValue) != len(expectedTokens[i].NestedValue) {
			t.Errorf("Expected token %d to have %d nested value(s), got %d", i, len(expectedTokens[i].NestedValue), len(tokens[i].NestedValue))
		}
	}
}
