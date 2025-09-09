package commands

import "github.com/codecrafters-io/redis-starter-go/app/shared"

// psync handles the PSYNC command.
// Usage: PSYNC masterReplID masterReplOffset
// Returns: "FULLRESYNC masterReplID masterReplOffset"
// This is typically used to synchronize a replica with a master.
func Psync(connID string, args []shared.Value) shared.Value {
	if len(args) < 2 {
		return createErrorResponse("ERR wrong number of arguments for 'psync' command")
	}

	result := "FULLRESYNC " + shared.StoreState.MasterReplID + " 0"

	return shared.Value{Typ: "string", Str: result}
}
