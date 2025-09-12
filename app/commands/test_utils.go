package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/network"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// clearMemory clears all entries from the shared memory for testing
func clearMemory() {
	for k := range server.Memory {
		delete(server.Memory, k)
	}
}

// clearTransactions clears all transactions for testing
func clearTransactions() {
	for k := range network.Transactions {
		delete(network.Transactions, k)
	}
}

// getListAsArray gets list as array for testing (works with both linked list and array)
func getListAsArray(key string) []string {
	entry, exists := server.Memory[key]
	if !exists {
		return nil
	}
	if entry.List != nil {
		return entry.List.ToArray()
	}
	return entry.Array
}

// initCommandHandlers initializes the shared command handlers for testing
func initCommandHandlers() {
	network.CommandHandlers = map[string]shared.CommandHandler{
		"SET":    Set,
		"GET":    Get,
		"LPUSH":  Lpush,
		"RPUSH":  Rpush,
		"LPOP":   Lpop,
		"LLEN":   Llen,
		"LRANGE": Lrange,
		"INCR":   Incr,
		"PING":   Ping,
		"ECHO":   Echo,
		"TYPE":   Type,
		"XADD":   Xadd,
		"XRANGE": Xrange,
		"XREAD":  Xread,
		"BLPOP":  Blpop,
	}
}
