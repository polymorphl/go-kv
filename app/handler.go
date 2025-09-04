package main

// Handlers maps Redis command names to their corresponding handler functions.
// Each handler function takes an array of Value arguments and returns a Value response.
var Handlers = map[string]func([]Value) Value{
	"PING": ping,
	"ECHO": echo,
	// string commands
	"SET": set,
	"GET": get,
	// list commands
	"LPUSH":  lpush,
	"RPUSH":  rpush,
	"LRANGE": lrange,
	"LLEN":   llen,
	"LPOP":   lpop,
	"BLPOP":  blpop,
	// stream commands
	"TYPE": typeCmd,
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
