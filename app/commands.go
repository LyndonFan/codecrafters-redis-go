package main

import (
	"fmt"
	"strings"
	"time"
)

func runCommand(commandName string, args []any) (string, error) {
	switch strings.ToLower(commandName) {
	case "ping":
		command := ping{}
		return command.execute(args)
	case "echo":
		command := echo{}
		return command.execute(args)
	default:
		return "", fmt.Errorf("unknown command: %s", commandName)
	}
}

type ping struct{}

func (p *ping) execute(args []any) (string, error) {
	if len(args) > 0 {
		return "", fmt.Errorf("unexpected arguments: %v", args)
	}
	return "PONG", nil
}

type echo struct{}

func (e *echo) execute(args []any) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("expected 1 argument, got %d", len(args))
	}
	return fmt.Sprintf("%s", args[0]), nil
}

type Entry struct {
	Value       any
	TimeCreated time.Time
}

var cache = make(map[string]*Entry)

type set struct{}

func (s *set) execute(args []any) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("expected 2 arguments, got %d", len(args))
	}
	key, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("expected string as first argument, got %T", args[0])
	}
	cache[args[0].(string)] = &Entry{
		Value:       args[1],
		TimeCreated: time.Now(),
	}
	return "OK", nil
}
