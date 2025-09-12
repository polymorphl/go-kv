package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/network"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// Exec handles the EXEC command.
// Executes all commands that were queued since the MULTI command was issued.
// Examples:
//
//	MULTI           // Starts a transaction block
//	SET mykey "Hello"
//	GET mykey
//	EXEC            // Executes the transaction block
func Exec(connID string, args []shared.Value) shared.Value {
	if len(args) != 0 {
		return createErrorResponse("ERR wrong number of arguments for 'exec' command")
	}

	// Check if there's an active transaction for this connection (concurrency-safe)
	transaction, exists := network.TransactionsGet(connID)
	if !exists {
		return createErrorResponse("ERR EXEC without MULTI")
	}

	// Clear the transaction (concurrency-safe)
	network.TransactionsDelete(connID)

	if len(transaction.Commands) == 0 {
		return shared.Value{Typ: "array", Array: []shared.Value{}}
	}

	// Execute all queued commands
	results := make([]shared.Value, len(transaction.Commands))
	for i, queuedCmd := range transaction.Commands {
		results[i] = network.ExecuteCommand(queuedCmd.Command, connID, queuedCmd.Args)
	}

	return shared.Value{Typ: "array", Array: results}
}
