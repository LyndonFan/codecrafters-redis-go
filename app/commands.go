package main

import "fmt"

func runCommand(commandName string, args []any) (string, error) {
	switch commandName {
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
