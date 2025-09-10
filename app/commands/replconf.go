package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// replconf handles the REPLCONF command.
// Usage: REPLCONF listening-port <port>
// Returns: "OK" on success, error message on failure.
//
// This command sets the listening port for the replica.
// It is used to configure the replica to listen on a specific port.
// It is also used to register the replica connection as soon as we receive REPLCONF.
func Replconf(connID string, args []shared.Value) shared.Value {
	if len(args) < 2 {
		return createErrorResponse("ERR wrong number of arguments for 'replconf' command")
	}

	// Register replica connection as soon as we receive REPLCONF
	// This ensures the replica is registered before any commands are processed
	if conn, exists := shared.Connections[connID]; exists {
		// Only register if not already registered
		if _, alreadyRegistered := shared.StoreState.Replicas[connID]; !alreadyRegistered {
			shared.StoreState.Replicas[connID] = conn
		}
	}

	return shared.Value{Typ: "string", Str: "OK"}
}
