package main

import (
	"fmt"
	"strings"
	"time"
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

type Entry struct {
	Value       any
	TimeCreated time.Time
	ExpiresAt   time.Time
}

type Cache map[string]*Entry

var cache = Cache{}

func (c *Cache) Set(args []any) (string, error) {
	key, expiresAt := "", time.Time{}
	var value any
	if len(args) != 2 && len(args) != 4 {
		return "", fmt.Errorf("expected 2 or 4 arguments, got %d", len(args))
	}
	var ok bool
	key, ok = args[0].(string)
	if !ok {
		return "", fmt.Errorf("unable to process key: %v", args[0])
	}
	value = args[1]
	now := time.Now()
	if len(args) == 4 {
		cmd, ok := args[2].(string)
		if !ok {
			return "", fmt.Errorf("unable to process command: %v", args[2])
		}
		duration, ok := args[3].(int64)
		if !ok {
			return "", fmt.Errorf("unable to process duration: %v", args[3])
		}
		switch strings.ToLower(cmd) {
		case "ex":
			expiresAt = now.Add(time.Duration(duration) * time.Second)
		case "px":
			expiresAt = now.Add(time.Duration(duration) * time.Millisecond)
		default:
			return "", fmt.Errorf("unknown unit %s, should be EX or PX", cmd)
		}
	}
	cache[key] = &Entry{
		Value:       value,
		TimeCreated: now,
		ExpiresAt:   expiresAt,
	}
	return "OK", nil
}

var ErrNotFound error = fmt.Errorf("not found")

func (c Cache) Get(args []any) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("expected 1 argument, got %d", len(args))
	}
	key, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("unable to process key: %v", args[0])
	}
	entry, found := c[key]
	fmt.Println("key:", key)
	fmt.Println("cache:", c)
	if !found || entry.ExpiresAt.Before(time.Now()) {
		return "", ErrNotFound
	}
	return fmt.Sprintf("%v", entry.Value), nil
}
