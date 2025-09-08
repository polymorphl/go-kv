package commands

import "github.com/codecrafters-io/redis-starter-go/app/shared"

// echo handles the ECHO command.
// Usage: ECHO message
// Returns: The message that was sent as an argument.
// This command is useful for testing the connection and verifying that
// the server is receiving and processing commands correctly.
func Echo(args []shared.Value) shared.Value {
	return shared.Value{Typ: "string", Str: args[0].Bulk}
}
