package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var Handlers = map[string]func([]Value) Value{
	"PING": ping,
	"ECHO": echo,
	"SET":  set,
	"GET":  get,
}

type MemoryEntry struct {
	Value   string
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

	fmt.Println("entry", entry)

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

	return Value{Typ: "string", Str: entry.Value}
}
