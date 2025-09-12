package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/network"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// Multi handles the MULTI command.
// After a MULTI command is executed, any further commands from the same connection will be "queued" but not executed.
// The commands will be executed only after a EXEC command is issued.
// Examples:
//
//	MULTI           // Starts a transaction block
func Multi(connID string, args []shared.Value) shared.Value {
	if len(args) != 0 {
		return createErrorResponse("ERR wrong number of arguments for 'multi' command")
	}

	// Create a new transaction for this connection (concurrency-safe)
	network.TransactionsSet(connID, shared.Transaction{Commands: []shared.QueuedCommand{}})

	return shared.Value{Typ: "string", Str: "OK"}
}
