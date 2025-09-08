package commands

import "github.com/codecrafters-io/redis-starter-go/app/shared"

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

	// Check if there's an active transaction for this connection
	transaction, exists := shared.Transactions[connID]
	if !exists {
		return createErrorResponse("ERR EXEC without MULTI")
	}

	// Clear the transaction
	delete(shared.Transactions, connID)

	if len(transaction.Commands) == 0 {
		return shared.Value{Typ: "array", Array: []shared.Value{}}
	}

	// For now, return OK - in a full implementation, we would execute the queued commands
	return shared.Value{Typ: "string", Str: "OK"}
}
