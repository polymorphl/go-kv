package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/pubsub"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// ping handles the PING command.
// Usage: PING [message]
// Returns: "PONG" if no message provided, otherwise echoes the provided message.
// In subscribed mode, returns a RESP array with "pong" and empty bulk string.
// This is typically used to test if the server is alive and responsive.
func Ping(connID string, args []shared.Value) shared.Value {
	// Check if client is in subscribed mode
	if pubsub.SubscribedModeGet(connID) {
		// In subscribed mode, return array with "pong" and empty bulk string
		return shared.Value{
			Typ: "array",
			Array: []shared.Value{
				{Typ: "bulk", Bulk: "pong"},
				{Typ: "bulk", Bulk: ""},
			},
		}
	}

	if len(args) == 0 {
		return shared.Value{Typ: "string", Str: "PONG"}
	}

	return Echo(connID, args)
}
