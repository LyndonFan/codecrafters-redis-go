package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/token"
)

func runCommand(ctx context.Context, commandName string, args []any) (*token.Token, error) {
	log.Println(strings.ToLower(commandName), args)
	var err error
	reconstructedToken, err := reconstructCommandToken(commandName, args)
	if err != nil {
		return nil, fmt.Errorf("unable to reconstruct token: %v", err)
	}
	switch strings.ToLower(commandName) {
	case "ping":
		return ping(args)
	case "echo":
		return echo(args)
	case "info":
		return info(args)
	case "replconf":
		return repl.RespondToReplconf(ctx, args)
	case "psync":
		return repl.RespondToPsync(ctx, args)
	case "wait":
		return repl.RespondToWait(ctx, args)
	case "set":
		go repl.PropagateCommandToken(reconstructedToken)
		err = cache.Set(args)
		if err != nil {
			return nil, err
		}
		return &token.OKToken, nil
	case "get":
		res, err := cache.Get(args)
		if err == ErrNotFound {
			return &token.NullBulkString, nil
		}
		if err != nil {
			return nil, err
		}
		resToken, err := token.CreateToken(res)
		if err != nil {
			return nil, fmt.Errorf("got %v from cache, but unable to cast to token: %v", res, err)
		}
		return resToken, nil
	default:
		return nil, fmt.Errorf("unknown command: %s", commandName)
	}
}

func reconstructCommandToken(commandName string, args []any) (*token.Token, error) {
	nestedValues := make([]*token.Token, len(args)+1)
	nestedValues[0] = &token.Token{Type: token.BulkStringType, SimpleValue: commandName}
	for i, arg := range args {
		nestedValues[i+1] = &token.Token{Type: token.BulkStringType, SimpleValue: fmt.Sprintf("%v", arg)}
	}
	tkn := token.Token{Type: token.ArrayType, NestedValue: nestedValues}
	return &tkn, nil
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
