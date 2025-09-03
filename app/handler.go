package main

import (
	"strconv"
	"strings"
	"time"
)

var Handlers = map[string]func([]Value) Value{
	"PING":  ping,
	"ECHO":  echo,
	"SET":   set,
	"GET":   get,
	"RPUSH": rpush,
}

type MemoryEntry struct {
	Value   string
	Array   []string
	Expires int64 // Unix timestamp in milliseconds, 0 means no expiry
}

var memory = make(map[string]MemoryEntry)

func ping(args []Value) Value {
	if len(args) == 0 {
		return Value{Typ: "string", Str: "PONG"}
	}

	return Value{Typ: "string", Str: args[0].Bulk}
}

func echo(args []Value) Value {
	return Value{Typ: "string", Str: args[0].Bulk}
}

func set(args []Value) Value {
	if len(args) < 2 {
		return Value{Typ: "error", Str: "ERR wrong number of arguments for 'set' command"}
	}

	key := args[0].Bulk
	value := args[1].Bulk
	entry := MemoryEntry{Value: value, Expires: 0}

	for i := 2; i < len(args); i++ {
		if strings.ToUpper(args[i].Bulk) == "PX" && i+1 < len(args) {
			ms, err := strconv.ParseInt(args[i+1].Bulk, 10, 64)
			if err != nil {
				return Value{Typ: "error", Str: "ERR value is not an integer or out of range"}
			}
			entry.Expires = time.Now().UnixMilli() + ms
			i++
		}
	}

	memory[key] = entry
	return Value{Typ: "string", Str: "OK"}
}

func get(args []Value) Value {
	if len(args) != 1 {
		return Value{Typ: "error", Str: "ERR wrong number of arguments for 'get' command"}
	}

	key := args[0].Bulk
	entry, exists := memory[key]

	if !exists {
		return Value{Typ: "null", Str: ""}
	}

	// Check if key has expired
	if entry.Expires > 0 && time.Now().UnixMilli() > entry.Expires {
		// Remove expired key
		delete(memory, key)
		return Value{Typ: "null", Str: ""}
	}

	// If it's an array, return an error (GET only works with strings)
	if len(entry.Array) > 0 {
		return Value{Typ: "error", Str: "WRONGTYPE Operation against a key holding the wrong kind of value"}
	}

	return Value{Typ: "string", Str: entry.Value}
}

func rpush(args []Value) Value {
	if len(args) < 2 {
		return Value{Typ: "error", Str: "ERR wrong number of arguments for 'rpush' command"}
	}

	key := args[0].Bulk
	entry, exists := memory[key]

	// If key doesn't exist or is not an array, create a new array
	if !exists || len(entry.Array) == 0 && entry.Value != "" {
		entry = MemoryEntry{Array: []string{}, Expires: 0}
	}

	// Add all values to the end of the array
	for i := 1; i < len(args); i++ {
		entry.Array = append(entry.Array, args[i].Bulk)
	}

	memory[key] = entry
	return Value{Typ: "integer", Num: len(entry.Array)}
}
