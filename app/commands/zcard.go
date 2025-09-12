package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// zcard handles the ZCARD command.
// Usage: ZCARD key
// Returns: The number of elements in the sorted set stored at key.
//
// This command returns the number of elements in the sorted set stored at key.
// If key does not exist, it is treated as an empty sorted set and 0 is returned.
// If key exists but is not a sorted set, an error is returned.
//
// Examples:
//
//	ZCARD myzset                    // Returns the number of elements in myzset
//	ZCARD nonexistent              // Returns 0 (key doesn't exist)
//	ZCARD mystring                 // Returns error (wrong type)
func Zcard(connID string, args []shared.Value) shared.Value {
	if len(args) != 1 {
		return createErrorResponse("ERR wrong number of arguments for 'zcard' command")
	}

	key := args[0].Bulk
	entry, exists := server.Memory[key]

	if !exists {
		return shared.Value{Typ: "integer", Num: 0}
	}

	if entry.SortedSet == nil {
		return createErrorResponse("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return shared.Value{Typ: "integer", Num: entry.SortedSet.Size}
}
