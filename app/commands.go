package main

import (
	"fmt"
	"strings"
)

func runCommand(commandName string, args []any) (*Token, error) {
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
		return &Token{Type: simpleStringType, SimpleValue: res}, nil
	case "get":
		res, err := cache.Get(args)
		if err == ErrNotFound {
			return &Token{Type: bulkStringType, representNull: true}, nil
		}
		if err != nil {
			return nil, err
		}
		return &Token{Type: simpleStringType, SimpleValue: res}, nil
	default:
		return nil, fmt.Errorf("unknown command: %s", commandName)
	}
}

func ping(args []any) (*Token, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("unexpected arguments: %v", args)
	}
	return &Token{Type: simpleStringType, SimpleValue: "PONG"}, nil
}

func echo(args []any) (*Token, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("expected 1 argument, got %d", len(args))
	}
	value := fmt.Sprintf("%s", args[0])
	return &Token{Type: simpleStringType, SimpleValue: value}, nil
}

func info(args []any) (*Token, error) {
	if len(args) > 1 {
		return nil, fmt.Errorf("unexpected arguments: %v", args)
	}
	return &Token{Type: bulkStringType, SimpleValue: "role:master"}, nil
}
