package shared

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

// Transaction represents a transaction that is being executed.
type Transaction struct {
	Commands []Value
}

// memory is the global in-memory database that stores all key-value pairs.
var Memory = make(map[string]MemoryEntry)

// Transactions is the global map of transactions that are being executed.
// The key is the connection ID.
var Transactions = make(map[string]Transaction)
