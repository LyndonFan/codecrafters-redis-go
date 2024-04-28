package main

import (
	"fmt"
	"go/token"

	"github.com/codecrafters-io/redis-starter-go/app/token"
)

func runTokens(tokens []*token.Token) (*token.Token, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty input")
	}
	for tokens[0].Type == arrayType {
		newTokens := make([]token.Token, len(tokens[0].NestedValue)+len(tokens)-1)
		copy(newTokens, tokens[0].NestedValue)
		copy(newTokens[len(tokens[0].NestedValue):], tokens[1:])
		tokens = newTokens
	}
	if valueEncoding[tokens[0].Type] == "nested" {
		return nil, fmt.Errorf("can't parse nested input of type %s", tokens[0].Type)
	}
	command := tokens[0].SimpleValue
	values := make([]any, len(tokens)-1)
	for i := 1; i < len(tokens); i++ {
		if valueEncoding[tokens[i].Type] == "nested" {
			return nil, fmt.Errorf("can't parse nested input of type %s", tokens[i].Type)
		}
		values[i-1] = tokens[i].SimpleValue
	}
	res, err := runCommand(command, values)
	return res, err
}
