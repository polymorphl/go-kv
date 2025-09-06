package main

// StreamEntry represents a single entry in a Redis stream
type StreamEntry struct {
	ID   string            // Stream ID (e.g., "1526985054069-0")
	Data map[string]string // Field-value pairs
}

// MemoryEntry represents a value stored in the in-memory database.
// It can hold either a string value or an array of strings, with optional expiration.
type MemoryEntry struct {
	Value   string        // String value (used when Array is empty)
	Array   []string      // Array of strings (used for list operations)
	Stream  []StreamEntry // Stream entries (used for stream operations)
	Expires int64         // Unix timestamp in milliseconds, 0 means no expiry
}

// Handlers maps Redis command names to their corresponding handler functions.
// Each handler function takes an array of Value arguments and returns a Value response.
var Handlers = map[string]func([]Value) Value{
	// utils commands
	"PING": ping,
	"ECHO": echo,
	"TYPE": typeCmd, // type is a reserved word
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
	"XADD":   xadd,
	"XRANGE": xrange,
	"XREAD":  xread,
}

// memory is the global in-memory database that stores all key-value pairs.
var memory = make(map[string]MemoryEntry)
