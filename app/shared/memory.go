package shared

import (
	"fmt"
	"net"
)

// StreamEntry represents a single entry in a Redis stream
type StreamEntry struct {
	ID   string            // Stream ID (e.g., "1526985054069-0")
	Data map[string]string // Field-value pairs
}

// MemoryEntry represents a value stored in the in-memory database.
// It can hold either a string value, an array of strings, or a linked list, with optional expiration.
type MemoryEntry struct {
	Value   string        // String value (used when Array is empty)
	Array   []string      // Array of strings (used for list operations - kept for compatibility)
	List    *LinkedList   // Linked list (used for optimized list operations)
	Stream  []StreamEntry // Stream entries (used for stream operations)
	Expires int64         // Unix timestamp in milliseconds, 0 means no expiry
}

// QueuedCommand represents a command that is queued in a transaction.
type QueuedCommand struct {
	Command string
	Args    []Value
}

// Transaction represents a transaction that is being executed.
type Transaction struct {
	Commands []QueuedCommand
}

// memory is the global in-memory database that stores all key-value pairs.
var Memory = make(map[string]MemoryEntry)

// Transactions is the global map of transactions that are being executed.
// The key is the connection ID.
var Transactions = make(map[string]Transaction)

// Connections is the global map of active connections.
// The key is the connection ID.
var Connections = make(map[string]net.Conn)

// CommandHandler represents a function that handles a Redis command
type CommandHandler func(string, []Value) Value

// CommandHandlers is a map of command names to their handler functions
var CommandHandlers map[string]CommandHandler

// ExecuteCommand executes a command using the shared handlers map
func ExecuteCommand(command string, connID string, args []Value) Value {
	if handler, ok := CommandHandlers[command]; ok {
		return handler(connID, args)
	}
	return Value{Typ: "string", Str: ""}
}

// IsWriteCommand checks if a command modifies data and should be propagated to replicas
func IsWriteCommand(command string) bool {
	writeCommands := map[string]bool{
		"SET":     true,
		"LPUSH":   true,
		"RPUSH":   true,
		"LPOP":    true,
		"BLPOP":   true,
		"INCR":    true,
		"XADD":    true,
		"MULTI":   true,
		"EXEC":    true,
		"DISCARD": true,
	}
	return writeCommands[command]
}

// PropagateCommand sends a command to all connected replicas
func PropagateCommand(command string, args []Value) {
	if StoreState.Role != "master" {
		return
	}

	fmt.Printf("Propagating command: %s to %d replicas\n", command, len(StoreState.Replicas))

	// Build the command array for propagation
	commandArray := make([]Value, len(args)+1)
	commandArray[0] = Value{Typ: "bulk", Bulk: command}
	for i, arg := range args {
		commandArray[i+1] = arg
	}

	// Send to all replicas
	for replicaID, replicaConn := range StoreState.Replicas {
		// Create a simple writer for this connection
		bytes := Value{Typ: "array", Array: commandArray}.Marshal()
		_, err := replicaConn.Write(bytes)
		if err != nil {
			// Remove failed replica connection
			delete(StoreState.Replicas, replicaID)
			fmt.Printf("Failed to propagate command to replica %s: %v\n", replicaID, err)
		} else {
			fmt.Printf("Successfully propagated command to replica %s\n", replicaID)
		}
	}
}

type State struct {
	Role             string
	ReplicaOf        string
	MasterReplID     string
	MasterReplOffset int64
	Replicas         map[string]net.Conn // Map of replica connection IDs to their connections
}

var StoreState = State{
	Role:             "master",
	ReplicaOf:        "",
	MasterReplID:     "",
	MasterReplOffset: 0,
	Replicas:         make(map[string]net.Conn),
}
