package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Entry struct {
	Value       any
	TimeCreated time.Time
	ExpiresAt   time.Time
}

type Cache map[string]*Entry

var cache = Cache{}

func (c *Cache) Set(args []any) error {
	key, expiresAt := "", time.Time{}
	var value any
	if len(args) != 2 && len(args) != 4 {
		return fmt.Errorf("expected 2 or 4 arguments, got %d", len(args))
	}
	var ok bool
	key, ok = args[0].(string)
	if !ok {
		return fmt.Errorf("unable to process key: %v", args[0])
	}
	value = args[1]
	now := time.Now()
	if len(args) == 4 {
		cmd, ok := args[2].(string)
		if !ok {
			return fmt.Errorf("unable to process command: %v", args[2])
		}
		var duration int64
		durationString, ok := args[3].(string)
		var err error
		if ok {
			duration, err = strconv.ParseInt(durationString, 10, 64)
		}
		if !ok || err != nil {
			return fmt.Errorf("unable to process duration: %v", args[3])
		}
		switch strings.ToLower(cmd) {
		case "ex":
			expiresAt = now.Add(time.Duration(duration) * time.Second)
		case "px":
			expiresAt = now.Add(time.Duration(duration) * time.Millisecond)
		default:
			return fmt.Errorf("unknown unit %s, should be EX or PX", cmd)
		}
	}
	cache[key] = &Entry{
		Value:       value,
		TimeCreated: now,
		ExpiresAt:   expiresAt,
	}
	return nil
}

var ErrNotFound error = fmt.Errorf("not found")

func (c Cache) Get(args []any) (any, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("expected 1 argument, got %d", len(args))
	}
	key, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("unable to process key: %v", args[0])
	}
	entry, found := c[key]
	if !found || (entry.ExpiresAt.Before(time.Now()) && entry.ExpiresAt != time.Time{}) {
		return "", ErrNotFound
	}
	return entry.Value, nil
}
