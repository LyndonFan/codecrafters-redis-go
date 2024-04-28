package main

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/token"
)

func runCommand(commandName string, args []any) (*token.Token, error) {
	fmt.Println(commandName, args)
	switch strings.ToLower(commandName) {
	case "ping":
		return ping(args)
	case "echo":
		return echo(args)
	case "info":
		return info(args)
	case "set":
		res, err := cache.Set(args)
		if err != nil {
			return nil, err
		}
		return &token.Token{Type: token.SimpleStringType, SimpleValue: res}, nil
	case "get":
		res, err := cache.Get(args)
		if err == ErrNotFound {
			return &token.NullBulkString, nil
		}
		if err != nil {
			return nil, err
		}
		return &token.Token{Type: token.SimpleStringType, SimpleValue: res}, nil
	default:
		return nil, fmt.Errorf("unknown command: %s", commandName)
	}
}

func ping(args []any) (*token.Token, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("unexpected arguments: %v", args)
	}
	return &token.Token{Type: token.SimpleStringType, SimpleValue: "PONG"}, nil
}

func echo(args []any) (*token.Token, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("expected 1 argument, got %d", len(args))
	}
	value := fmt.Sprintf("%s", args[0])
	return &token.Token{Type: token.SimpleStringType, SimpleValue: value}, nil
}

func info(args []any) (*token.Token, error) {
	if len(args) > 1 {
		return nil, fmt.Errorf("unexpected arguments: %v", args)
	}
	lines := make([]string, 0, 4)
	for k, v := range repl.InfoMap() {
		lines = append(lines, fmt.Sprintf("%s:%s", k, v))
	}
	return &token.Token{Type: token.BulkStringType, SimpleValue: strings.Join(lines, token.TERMINATOR)}, nil
}
