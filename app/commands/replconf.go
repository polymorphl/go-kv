package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// replconf handles the REPLCONF command.
// Usage: REPLCONF listening-port <port> or REPLCONF GETACK *
// Returns: "OK" on success, error message on failure, or REPLCONF ACK <offset> for GETACK.
//
// This command sets the listening port for the replica or responds to GETACK requests.
// It is also used to register the replica connection as soon as we receive REPLCONF.
func Replconf(connID string, args []shared.Value) shared.Value {
	if len(args) < 2 {
		return createErrorResponse("ERR wrong number of arguments for 'replconf' command")
	}

	subcommand := args[0].Bulk

	// Handle REPLCONF GETACK * command
	if subcommand == "GETACK" {
		if len(args) >= 2 && args[1].Bulk == "*" {
			// Respond with REPLCONF ACK 0 (hardcoded offset for now)
			return shared.Value{Typ: "array", Array: []shared.Value{
				{Typ: "bulk", Bulk: "REPLCONF"},
				{Typ: "bulk", Bulk: "ACK"},
				{Typ: "bulk", Bulk: "0"},
			}}
		}
		return createErrorResponse("ERR wrong number of arguments for 'replconf getack' command")
	}

	// Handle REPLCONF ACK <offset> command (from replicas to master)
	if subcommand == "ACK" {
		if len(args) >= 2 {
			// Mark this replica as having acknowledged
			server.AcknowledgedReplicasSet(connID)
			// Return NO_RESPONSE since this is an internal command
			return shared.Value{Typ: shared.NO_RESPONSE, Str: ""}
		}
		return createErrorResponse("ERR wrong number of arguments for 'replconf ack' command")
	}

	// Register replica connection as soon as we receive REPLCONF
	// This ensures the replica is registered before any commands are processed
	if conn, exists := shared.ConnectionsGet(connID); exists {
		// Only register if not already registered
		if _, alreadyRegistered := shared.StoreState.Replicas[connID]; !alreadyRegistered {
			server.ReplicasSet(connID, conn)
		}
	}

	return shared.Value{Typ: "string", Str: "OK"}
}
