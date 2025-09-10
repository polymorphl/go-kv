package commands

import "github.com/codecrafters-io/redis-starter-go/app/shared"

// wait handles the WAIT command.
func Wait(connID string, args []shared.Value) shared.Value {
	if len(args) != 2 {
		return createErrorResponse("ERR wrong number of arguments for 'wait' command")
	}

	return shared.Value{Typ: "integer", Num: 0}
}
