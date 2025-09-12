package shared

import (
	"fmt"
	"net"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/pubsub"
)

// Function aliases for backward compatibility
var NewResp = protocol.NewResp
var NewWriter = protocol.NewWriter

// Constant aliases for backward compatibility
const NO_RESPONSE = protocol.NO_RESPONSE

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

// CommandHandlers is a map of command names to their handler functions
var CommandHandlers map[string]CommandHandler

// ExecuteCommand executes a command using the shared handlers map
func ExecuteCommand(command string, connID string, args []protocol.Value) protocol.Value {
	// Check if client is in subscribed mode and command is not allowed
	if pubsub.SubscribedModeGet(connID) && !pubsub.IsAllowedInSubscribedMode(command) {
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
