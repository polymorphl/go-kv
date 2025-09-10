package commands

import "github.com/codecrafters-io/redis-starter-go/app/shared"

// discard handles the DISCARD command.
// Usage: DISCARD
// Returns: OK
// This command is used to discard all commands that have been queued since the MULTI command was issued.
// If no MULTI command has been issued, it returns an error.
func Discard(connID string, args []shared.Value) shared.Value {
	if len(args) != 0 {
		return createErrorResponse("ERR wrong number of arguments for 'discard' command")
	}

	if _, exists := shared.TransactionsGet(connID); !exists {
		return createErrorResponse("ERR DISCARD without MULTI")
	}

	shared.TransactionsDelete(connID)

	return shared.Value{Typ: "string", Str: "OK"}
}
