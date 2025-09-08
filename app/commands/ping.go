package commands

import "github.com/codecrafters-io/redis-starter-go/app/shared"

// ping handles the PING command.
// Usage: PING [message]
// Returns: "PONG" if no message provided, otherwise echoes the provided message.
// This is typically used to test if the server is alive and responsive.
func Ping(args []shared.Value) shared.Value {
	if len(args) == 0 {
		return shared.Value{Typ: "string", Str: "PONG"}
	}

	return Echo(args)
}
