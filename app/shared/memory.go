package shared

import (
	"fmt"
	"net"
	"sync"
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

// Mutexes to protect concurrent access to global maps
var connectionsMu sync.RWMutex
var transactionsMu sync.RWMutex
var replicasMu sync.RWMutex
var acknowledgedReplicasMu sync.RWMutex

// AcknowledgedReplicas tracks which replicas have acknowledged commands
var AcknowledgedReplicas = make(map[string]bool)

// CommandHandler represents a function that handles a Redis command
type CommandHandler func(string, []Value) Value

// CommandHandlers is a map of command names to their handler functions
var CommandHandlers map[string]CommandHandler

// ExecuteCommand executes a command using the shared handlers map
func ExecuteCommand(command string, connID string, args []Value) Value {
	// Check if client is in subscribed mode and command is not allowed
	if SubscribedModeGet(connID) && !IsAllowedInSubscribedMode(command) {
		return Value{Typ: "error", Str: fmt.Sprintf("ERR Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context", command)}
	}

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

	// Build the command array for propagation
	commandArray := make([]Value, len(args)+1)
	commandArray[0] = Value{Typ: "bulk", Bulk: command}
	for i, arg := range args {
		commandArray[i+1] = arg
	}

	// Snapshot replicas under read lock to avoid concurrent map iteration/writes
	replicasMu.RLock()
	snapshot := make(map[string]net.Conn, len(StoreState.Replicas))
	for id, c := range StoreState.Replicas {
		snapshot[id] = c
	}
	replicasMu.RUnlock()

	// Send to all replicas using the snapshot
	for replicaID, replicaConn := range snapshot {
		bytes := Value{Typ: "array", Array: commandArray}.Marshal()
		_, err := replicaConn.Write(bytes)
		if err != nil {
			// Remove failed replica connection
			ReplicasDelete(replicaID)
			fmt.Printf("Failed to propagate command to replica %s: %v\n", replicaID, err)
		}
	}
}

type State struct {
	Role             string
	ReplicaOf        string
	MasterReplID     string
	MasterReplOffset int64
	Replicas         map[string]net.Conn // Map of replica connection IDs to their connections
	ConfigDir        string              // Directory where Redis stores its data
	ConfigDbfilename string              // Database filename
}

var StoreState = State{
	Role:             "master",
	ReplicaOf:        "",
	MasterReplID:     "",
	MasterReplOffset: 0,
	Replicas:         make(map[string]net.Conn),
	ConfigDir:        "/tmp/redis-data",
	ConfigDbfilename: "rdbfile",
}

// Connections helpers
func ConnectionsSet(connID string, conn net.Conn) {
	connectionsMu.Lock()
	Connections[connID] = conn
	connectionsMu.Unlock()
}

func ConnectionsDelete(connID string) {
	connectionsMu.Lock()
	delete(Connections, connID)
	connectionsMu.Unlock()
}

func ConnectionsGet(connID string) (net.Conn, bool) {
	connectionsMu.RLock()
	c, ok := Connections[connID]
	connectionsMu.RUnlock()
	return c, ok
}

// Transactions helpers
func TransactionsGet(connID string) (Transaction, bool) {
	transactionsMu.RLock()
	t, ok := Transactions[connID]
	transactionsMu.RUnlock()
	return t, ok
}

func TransactionsSet(connID string, t Transaction) {
	transactionsMu.Lock()
	Transactions[connID] = t
	transactionsMu.Unlock()
}

func TransactionsDelete(connID string) {
	transactionsMu.Lock()
	delete(Transactions, connID)
	transactionsMu.Unlock()
}

// Replicas helpers (inside StoreState)
func ReplicasGet(connID string) (net.Conn, bool) {
	replicasMu.RLock()
	c, ok := StoreState.Replicas[connID]
	replicasMu.RUnlock()
	return c, ok
}

func ReplicasSet(connID string, conn net.Conn) {
	replicasMu.Lock()
	StoreState.Replicas[connID] = conn
	replicasMu.Unlock()
}

func ReplicasDelete(connID string) {
	replicasMu.Lock()
	delete(StoreState.Replicas, connID)
	replicasMu.Unlock()
}

// AcknowledgedReplicasSet marks a replica as having acknowledged
func AcknowledgedReplicasSet(connID string) {
	acknowledgedReplicasMu.Lock()
	AcknowledgedReplicas[connID] = true
	acknowledgedReplicasMu.Unlock()
}

// AcknowledgedReplicasCount returns the number of replicas that have acknowledged
func AcknowledgedReplicasCount() int {
	acknowledgedReplicasMu.RLock()
	count := len(AcknowledgedReplicas)
	acknowledgedReplicasMu.RUnlock()
	return count
}

// AcknowledgedReplicasClear clears all acknowledgments (used before sending new GETACK)
func AcknowledgedReplicasClear() {
	acknowledgedReplicasMu.Lock()
	AcknowledgedReplicas = make(map[string]bool)
	acknowledgedReplicasMu.Unlock()
}

// SendReplconfGetack sends REPLCONF GETACK * to all replicas
func SendReplconfGetack() {
	// Clear previous acknowledgments before sending new GETACK
	AcknowledgedReplicasClear()

	cmd := Value{Typ: "array", Array: []Value{
		{Typ: "bulk", Bulk: "REPLCONF"},
		{Typ: "bulk", Bulk: "GETACK"},
		{Typ: "bulk", Bulk: "*"},
	}}
	bytes := cmd.Marshal()

	replicasMu.RLock()
	replicas := make(map[string]net.Conn, len(StoreState.Replicas))
	for id, c := range StoreState.Replicas {
		replicas[id] = c
	}
	replicasMu.RUnlock()

	for replicaID, replicaConn := range replicas {
		_, err := replicaConn.Write(bytes)
		if err != nil {
			ReplicasDelete(replicaID)
			fmt.Printf("Failed to send GETACK to replica %s: %v\n", replicaID, err)
		}
	}
}

// subscriptionsMu is the mutex for the subscriptions map
var subscriptionsMu sync.RWMutex

// Subscriptions is the global map of subscriptions.
// The key is the connection ID, the value is a slice of subscribed channels.
var Subscriptions = make(map[string][]string)

// subscribedModeMu is the mutex for the subscribed mode map
var subscribedModeMu sync.RWMutex

// SubscribedMode tracks which clients are in subscribed mode.
// The key is the connection ID, the value is true if the client is in subscribed mode.
var SubscribedMode = make(map[string]bool)

// SubscriptionsSet adds a subscription for a connection ID
func SubscriptionsSet(connID string, channel string) {
	subscriptionsMu.Lock()
	defer subscriptionsMu.Unlock()

	// Check if connection already exists
	if channels, exists := Subscriptions[connID]; exists {
		// Check if channel is already subscribed
		for _, existingChannel := range channels {
			if existingChannel == channel {
				return // Already subscribed to this channel
			}
		}
		// Add new channel to existing list
		Subscriptions[connID] = append(channels, channel)
	} else {
		// Create new subscription list
		Subscriptions[connID] = []string{channel}
	}
}

// SubscriptionsGet gets all subscriptions for a connection ID
func SubscriptionsGet(connID string) ([]string, bool) {
	subscriptionsMu.RLock()
	defer subscriptionsMu.RUnlock()
	channels, ok := Subscriptions[connID]
	return channels, ok
}

// SubscriptionsDelete deletes a subscription for a connection ID
func SubscriptionsDelete(connID string) {
	subscriptionsMu.Lock()
	delete(Subscriptions, connID)
	subscriptionsMu.Unlock()
}

// SubscriptionsSetChannels sets the entire subscription list for a connection ID
func SubscriptionsSetChannels(connID string, channels []string) {
	subscriptionsMu.Lock()
	Subscriptions[connID] = channels
	subscriptionsMu.Unlock()
}

// SubscriptionsCountForChannel counts the number of clients subscribed to a specific channel
func SubscriptionsCountForChannel(channel string) int {
	subscriptionsMu.RLock()
	defer subscriptionsMu.RUnlock()

	count := 0
	for _, channels := range Subscriptions {
		for _, subscribedChannel := range channels {
			if subscribedChannel == channel {
				count++
				break // Each connection can only be counted once per channel
			}
		}
	}
	return count
}

// SubscribedMode helpers
func SubscribedModeSet(connID string) {
	subscribedModeMu.Lock()
	SubscribedMode[connID] = true
	subscribedModeMu.Unlock()
}

// SubscribedModeGet gets the subscribed mode for a connection ID
func SubscribedModeGet(connID string) bool {
	subscribedModeMu.RLock()
	defer subscribedModeMu.RUnlock()
	return SubscribedMode[connID]
}

func SubscribedModeDelete(connID string) {
	subscribedModeMu.Lock()
	delete(SubscribedMode, connID)
	subscribedModeMu.Unlock()
}

// IsAllowedInSubscribedMode checks if a command is allowed when client is in subscribed mode
func IsAllowedInSubscribedMode(command string) bool {
	allowedCommands := map[string]bool{
		"SUBSCRIBE":    true,
		"UNSUBSCRIBE":  true,
		"PSUBSCRIBE":   true,
		"PUNSUBSCRIBE": true,
		"PING":         true,
		"QUIT":         true,
		"RESET":        true,
	}
	return allowedCommands[command]
}
