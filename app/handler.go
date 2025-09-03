package main

import (
	"strconv"
	"strings"
	"time"
)

// Handlers maps Redis command names to their corresponding handler functions.
// Each handler function takes an array of Value arguments and returns a Value response.
var Handlers = map[string]func([]Value) Value{
	"PING":   ping,
	"ECHO":   echo,
	"SET":    set,
	"GET":    get,
	"RPUSH":  rpush,
	"LRANGE": lrange,
}

// MemoryEntry represents a value stored in the in-memory database.
// It can hold either a string value or an array of strings, with optional expiration.
type MemoryEntry struct {
	Value   string   // String value (used when Array is empty)
	Array   []string // Array of strings (used for list operations)
	Expires int64    // Unix timestamp in milliseconds, 0 means no expiry
}

// memory is the global in-memory database that stores all key-value pairs.
var memory = make(map[string]MemoryEntry)

// ping handles the PING command.
// Usage: PING [message]
// Returns: "PONG" if no message provided, otherwise echoes the provided message.
// This is typically used to test if the server is alive and responsive.
func ping(args []Value) Value {
	if len(args) == 0 {
		return Value{Typ: "string", Str: "PONG"}
	}

	return Value{Typ: "string", Str: args[0].Bulk}
}

// echo handles the ECHO command.
// Usage: ECHO message
// Returns: The message that was sent as an argument.
// This command is useful for testing the connection and verifying that
// the server is receiving and processing commands correctly.
func echo(args []Value) Value {
	return Value{Typ: "string", Str: args[0].Bulk}
}

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

// rpush handles the RPUSH command.
// Usage: RPUSH key value [value ...]
// Returns: The length of the list after the push operation.
//
// This command inserts all the specified values at the tail of the list stored at key.
// If key does not exist, it is created as an empty list before performing the push operation.
// If key exists but is not a list, it is converted to a list before the operation.
//
// Examples:
//
//	RPUSH mylist "one"                    // Creates list with one element, returns 1
//	RPUSH mylist "two" "three"            // Adds two elements, returns 3
//	RPUSH newlist "first" "second"        // Creates new list, returns 2
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

// lrange handles the LRANGE command.
// Usage: LRANGE key start stop
// Returns: Array of elements in the specified range.
//
// This command returns the specified elements of the list stored at key.
// The offsets start and stop are zero-based indexes, with 0 being the first element.
// Negative offsets can be used to start from the end of the list.
//
// Special cases:
//   - If start > stop, returns empty array
//   - If start < 0, treated as 0
//   - If stop >= list length, treated as list length - 1
//   - Both start and stop are inclusive
//
// Examples:
//
//	LRANGE mylist 0 2      // Returns elements at index 0, 1, and 2
//	LRANGE mylist 0 -1     // Returns all elements
//	LRANGE mylist -3 -1    // Returns last 3 elements
//	LRANGE mylist 5 3      // Returns empty array (start > stop)
func lrange(args []Value) Value {
	if len(args) != 3 {
		return Value{Typ: "error", Str: "ERR wrong number of arguments for 'lrange' command"}
	}

	key := args[0].Bulk
	entry, exists := memory[key]

	if !exists {
		return Value{Typ: "array", Array: []Value{}}
	}

	// Check if it's an array
	if len(entry.Array) == 0 {
		return Value{Typ: "error", Str: "WRONGTYPE Operation against a key holding the wrong kind of value"}
	}

	// Parse start and stop indices
	start, err := strconv.Atoi(args[1].Bulk)
	if err != nil {
		return Value{Typ: "error", Str: "ERR value is not an integer or out of range"}
	}
	stop, err := strconv.Atoi(args[2].Bulk)
	if err != nil {
		return Value{Typ: "error", Str: "ERR value is not an integer or out of range"}
	}

	arrayLen := len(entry.Array)

	// Handle negative indices (count from end)
	if start < 0 {
		start = arrayLen + start
	}
	if stop < 0 {
		stop = arrayLen + stop
	}

	// Clamp indices to valid range
	if start < 0 {
		start = 0
	}
	if stop >= arrayLen {
		stop = arrayLen - 1
	}

	// Check if start is after stop (invalid range)
	if start > stop {
		return Value{Typ: "array", Array: []Value{}}
	}

	// Pre-allocate result slice with exact capacity for better performance
	result := make([]Value, 0, stop-start+1)
	for _, value := range entry.Array[start : stop+1] {
		result = append(result, Value{Typ: "string", Str: value})
	}

	return Value{Typ: "array", Array: result}
}
