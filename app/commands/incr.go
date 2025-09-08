package commands

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// incr handles the INCR command.
//
// Examples:
//
//	INCR counter      // Increments counter from 5 to 6
func Incr(args []shared.Value) shared.Value {
	if len(args) != 1 {
		return createErrorResponse("ERR wrong number of arguments for 'incr' command")
	}

	key := args[0].Bulk
	entry, exists := shared.Memory[key]

	if !exists {
		shared.Memory[key] = shared.MemoryEntry{Value: "1", Expires: 0}
		return shared.Value{Typ: "integer", Num: 1}
	}

	value, err := strconv.Atoi(entry.Value)
	if err != nil {
		return createErrorResponse("ERR value is not an integer or out of range")
	}

	entry.Value = strconv.Itoa(value + 1)
	shared.Memory[key] = entry
	return shared.Value{Typ: "integer", Num: value + 1}
}
