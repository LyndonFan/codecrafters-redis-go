package main

import (
	"fmt"
	"strings"
)

func runCommand(commandName string, args []any) (string, error) {
	switch strings.ToLower(commandName) {
	case "ping":
		return ping(args)
	case "echo":
		return echo(args)
	case "set":
		return cache.Set(args)
	case "get":
		return cache.Get(args)
	default:
		return "", fmt.Errorf("unknown command: %s", commandName)
	}
}

func ping(args []any) (string, error) {
	if len(args) > 0 {
		return "", fmt.Errorf("unexpected arguments: %v", args)
	}
	return "PONG", nil
}

func echo(args []any) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("expected 1 argument, got %d", len(args))
	}
	return fmt.Sprintf("%s", args[0]), nil
}
