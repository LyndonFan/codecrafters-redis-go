package main

import "fmt"

func runTokens(tokens []*Token) (*Token, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty input")
	}
	if !isSimple[tokens[0].Type] {
		return nil, fmt.Errorf("expected simple input, got %s", tokens[0].Type)
	}
	command := tokens[0].SimpleValue
	values := make([]any, len(tokens)-1)
	for i := 1; i < len(tokens); i++ {
		if !isSimple[tokens[i].Type] {
			return nil, fmt.Errorf("expected simple input, got %s", tokens[i].Type)
		}
		values[i-1] = tokens[i].SimpleValue
	}
	res, err := runCommand(command, values)
	if err != nil {
		return nil, err
	}
	return &Token{Type: simpleStringType, SimpleValue: res}, nil
}
