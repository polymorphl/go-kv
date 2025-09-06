package main

import (
	"strconv"
	"strings"
	"time"
)

// set handles the SET command.
// Usage: SET key value [PX milliseconds]
// Returns: "OK" on success, error message on failure.
//
// This command sets a key to hold a string value. If the key already exists,
// it is overwritten. The PX option sets an expiration time in milliseconds.
//
// Examples:
//
//	SET mykey "Hello"           // Sets key without expiration
//	SET mykey "Hello" PX 1000   // Sets key with 1 second expiration
func set(args []Value) Value {
	if len(args) < 2 {
		return Value{Typ: "error", Str: "ERR wrong number of arguments for 'set' command"}
	}

	key := args[0].Bulk
	value := args[1].Bulk
	entry := MemoryEntry{Value: value, Expires: 0}

	// Parse optional PX (expiration) argument
	for i := 2; i < len(args); i++ {
		if strings.ToUpper(args[i].Bulk) == "PX" && i+1 < len(args) {
			ms, err := strconv.ParseInt(args[i+1].Bulk, 10, 64)
			if err != nil {
				return Value{Typ: "error", Str: "ERR value is not an integer or out of range"}
			}
			entry.Expires = time.Now().UnixMilli() + ms
			i++ // Skip the next argument since we've processed it
		}
	}

	memory[key] = entry
	return Value{Typ: "string", Str: "OK"}
}

// get handles the GET command.
// Usage: GET key
// Returns: The string value of the key, null if key doesn't exist, or error if key holds wrong type.
//
// This command retrieves the value of a key. It only works with string values.
// If the key holds an array (from RPUSH operations), it returns a WRONGTYPE error.
// Expired keys are automatically removed and return null.
//
// Examples:
//
//	GET mykey           // Returns the value of mykey
//	GET nonexistent     // Returns null
func get(args []Value) Value {
	if len(args) != 1 {
		return Value{Typ: "error", Str: "ERR wrong number of arguments for 'get' command"}
	}

	key := args[0].Bulk
	entry, exists := memory[key]

	if !exists {
		return Value{Typ: "null", Str: ""}
	}

	// Check if key has expired and remove it if so
	if entry.Expires > 0 && time.Now().UnixMilli() > entry.Expires {
		delete(memory, key)
		return Value{Typ: "null", Str: ""}
	}

	// GET only works with string values, not arrays
	if len(entry.Array) > 0 {
		return Value{Typ: "error", Str: "WRONGTYPE Operation against a key holding the wrong kind of value"}
	}

	return Value{Typ: "string", Str: entry.Value}
}

// incr handles the INCR command.
//
// Examples:
//
//	INCR counter      // Increments counter from 5 to 6
func incr(args []Value) Value {
	if len(args) != 1 {
		return Value{Typ: "error", Str: "ERR wrong number of arguments for 'incr' command"}
	}

	key := args[0].Bulk
	entry, exists := memory[key]

	if !exists {
		return Value{Typ: "error", Str: "ERR not exists"}
	}

	value, err := strconv.Atoi(entry.Value)
	if err != nil {
		return Value{Typ: "error", Str: "ERR value is not an integer or out of range"}
	}

	entry.Value = strconv.Itoa(value + 1)
	memory[key] = entry
	return Value{Typ: "integer", Num: value + 1}
}
