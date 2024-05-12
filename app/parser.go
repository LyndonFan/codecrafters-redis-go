package main

import (
	"fmt"
	"log"

	"github.com/codecrafters-io/redis-starter-go/app/token"
)

/*
TODO: update logic to handle 2 ways requests are sent
1. single command, each token as a separate token
2. multiple commands, at least first token is an array and all its values are command & args
*/

func runTokens(tokens []*token.Token) ([]*token.Token, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty input")
	}
	res := make([]*token.Token, 0, len(tokens)/2)
	for len(tokens) > 0 && tokens[0].Type == token.ArrayType {
		log.Println("processing", tokens[0].Value())
		subResult, err := runTokensSingleCommand(tokens[0].NestedValue)
		if err != nil {
			res = append(res, token.TokeniseError(err))
		} else {
			res = append(res, subResult)
		}
		repl.BytesProcessed += len(tokens[0].EncodedString())
		tokens = tokens[1:]
	}
	if len(tokens) == 0 {
		return res, nil
	}
	finalRes, err := runTokensSingleCommand(tokens)
	if err != nil {
		res = append(res, token.TokeniseError(err))
	} else {
		res = append(res, finalRes)
	}
	for _, tkn := range tokens {
		repl.BytesProcessed += len(tkn.EncodedString())
	}
	return res, nil
}

func runTokensSingleCommand(tokens []*token.Token) (*token.Token, error) {
	var err error
	if tokens[0].Type != token.SimpleStringType && tokens[0].Type != token.VerbatimStringType && tokens[0].Type != token.BulkStringType {
		err = fmt.Errorf("expected first token to be of string type, got %s", tokens[0].Type)
		return nil, err
	}
	log.Print("processing")
	for _, tkn := range tokens {
		log.Printf(" %v", tkn.Value())
	}
	log.Println()
	command, ok := tokens[0].Value().(string)
	if !ok {
		return nil, fmt.Errorf("expected first token to be string, but can't cast this: %v", tokens[0].Value())
	}
	values := make([]any, len(tokens)-1)
	for i := 1; i < len(tokens); i++ {
		if token.ValueEncoding[tokens[i].Type] == "nested" {
			return nil, fmt.Errorf("can't parse nested input of type %s", tokens[i].Type)
		}
		values[i-1] = tokens[i].Value()
	}
	return runCommand(command, values)
}
