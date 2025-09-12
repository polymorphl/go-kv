package shared

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

// Type aliases for backward compatibility
type Value = protocol.Value
type Resp = protocol.Resp
type Writer = protocol.Writer

// StreamEntry represents a single entry in a Redis stream
type StreamEntry struct {
	ID   string            // Stream ID (e.g., "1526985054069-0")
	Data map[string]string // Field-value pairs
}

// MemoryEntry represents a value stored in the in-memory database.
// It can hold either a string value, an array of strings, or a linked list, with optional expiration.
type MemoryEntry struct {
	Value     string        // String value (used when Array is empty)
	Array     []string      // Array of strings (used for list operations - kept for compatibility)
	List      *LinkedList   // Linked list (used for optimized list operations)
	Stream    []StreamEntry // Stream entries (used for stream operations)
	SortedSet *SortedSet    // Sorted set (used for sorted set operations)
	Expires   int64         // Unix timestamp in milliseconds, 0 means no expiry
}

// QueuedCommand represents a command that is queued in a transaction.
type QueuedCommand struct {
	Command string
	Args    []protocol.Value
}

// Transaction represents a transaction that is being executed.
type Transaction struct {
	Commands []QueuedCommand
}

// CommandHandler represents a function that handles a Redis command
type CommandHandler func(string, []protocol.Value) protocol.Value

// State represents the server state including replication information
type State struct {
	Role             string
	ReplicaOf        string
	MasterReplID     string
	MasterReplOffset int64
	Replicas         map[string]net.Conn // Map of replica connection IDs to their connections
	ConfigDir        string              // Directory where Redis stores its data
	ConfigDbfilename string              // Database filename
}
