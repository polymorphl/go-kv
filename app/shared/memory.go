package shared

import (
	"fmt"
	"net"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/pubsub"
)

// Type aliases for backward compatibility
type Value = protocol.Value
type Resp = protocol.Resp
type Writer = protocol.Writer

// Function aliases for backward compatibility
var NewResp = protocol.NewResp
var NewWriter = protocol.NewWriter

// Constant aliases for backward compatibility
const NO_RESPONSE = protocol.NO_RESPONSE

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
	Args    []protocol.Value
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

// CommandHandler represents a function that handles a Redis command
type CommandHandler func(string, []protocol.Value) protocol.Value

// CommandHandlers is a map of command names to their handler functions
var CommandHandlers map[string]CommandHandler

// ExecuteCommand executes a command using the shared handlers map
func ExecuteCommand(command string, connID string, args []protocol.Value) protocol.Value {
	// Check if client is in subscribed mode and command is not allowed
	if SubscribedModeGet(connID) && !IsAllowedInSubscribedMode(command) {
		return protocol.Value{Typ: "error", Str: fmt.Sprintf("ERR Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context", command)}
	}

	if handler, ok := CommandHandlers[command]; ok {
		return handler(connID, args)
	}
	return protocol.Value{Typ: "string", Str: ""}
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

// Wrapper functions for pubsub functionality
// These maintain backward compatibility while delegating to the pubsub package

// SubscriptionsSet adds a subscription for a connection ID
func SubscriptionsSet(connID string, channel string) {
	pubsub.SubscriptionsSet(connID, channel)
}

// SubscriptionsGet gets all subscriptions for a connection ID
func SubscriptionsGet(connID string) ([]string, bool) {
	return pubsub.SubscriptionsGet(connID)
}

// SubscriptionsDelete deletes a subscription for a connection ID
func SubscriptionsDelete(connID string) {
	pubsub.SubscriptionsDelete(connID)
}

// SubscriptionsSetChannels sets the entire subscription list for a connection ID
func SubscriptionsSetChannels(connID string, channels []string) {
	pubsub.SubscriptionsSetChannels(connID, channels)
}

// SubscriptionsCountForChannel counts the number of clients subscribed to a specific channel
func SubscriptionsCountForChannel(channel string) int {
	return pubsub.SubscriptionsCountForChannel(channel)
}

// SubscriptionsGetSubscribersForChannel returns all connection IDs subscribed to a specific channel
func SubscriptionsGetSubscribersForChannel(channel string) []string {
	return pubsub.SubscriptionsGetSubscribersForChannel(channel)
}

// SendMessageToSubscribers sends a message to all subscribers of a channel
func SendMessageToSubscribers(channel string, message string) int {
	return pubsub.SendMessageToSubscribers(channel, message, ConnectionsGet, ConnectionsDelete, SubscriptionsDelete, SubscribedModeDelete)
}

// SubscribedMode helpers
func SubscribedModeSet(connID string) {
	pubsub.SubscribedModeSet(connID)
}

// SubscribedModeGet gets the subscribed mode for a connection ID
func SubscribedModeGet(connID string) bool {
	return pubsub.SubscribedModeGet(connID)
}

func SubscribedModeDelete(connID string) {
	pubsub.SubscribedModeDelete(connID)
}

// IsAllowedInSubscribedMode checks if a command is allowed when client is in subscribed mode
func IsAllowedInSubscribedMode(command string) bool {
	return pubsub.IsAllowedInSubscribedMode(command)
}

// Test helper functions to access global variables for testing
// These are needed for backward compatibility with existing tests

// GetSubscriptionsMap returns the global subscriptions map for testing
func GetSubscriptionsMap() map[string][]string {
	return pubsub.Subscriptions
}

// GetSubscribedModeMap returns the global subscribed mode map for testing
func GetSubscribedModeMap() map[string]bool {
	return pubsub.SubscribedMode
}

// SetSubscriptionsMap sets the global subscriptions map for testing
func SetSubscriptionsMap(subscriptions map[string][]string) {
	pubsub.Subscriptions = subscriptions
}

// SetSubscribedModeMap sets the global subscribed mode map for testing
func SetSubscribedModeMap(subscribedMode map[string]bool) {
	pubsub.SubscribedMode = subscribedMode
}
