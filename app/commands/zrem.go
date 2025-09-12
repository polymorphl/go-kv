package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// zrem handles the ZREM command.
// Usage: ZREM key member [member ...]
// Returns: The number of members removed from the sorted set.
//
// This command removes one or more members from a sorted set.
// If the key does not exist, it is treated as an empty sorted set and 0 is returned.
// If the key exists but is not a sorted set, an error is returned.
//
// Examples:
//
//	ZREM myzset "one"                 // Removes one element from myzset
//	ZREM myzset "one" "two"           // Removes two elements from myzset
//	ZREM nonexistent "member"         // Returns 0 (key doesn't exist)
//	ZREM mystring "member"            // Returns error (wrong type)
func Zrem(connID string, args []shared.Value) shared.Value {
	if len(args) < 2 {
		return createErrorResponse("ERR wrong number of arguments for 'zrem' command")
	}

	key := args[0].Bulk
	entry, exists := server.Memory[key]

	if !exists {
		return shared.Value{Typ: "integer", Num: 0}
	}

	if entry.SortedSet == nil {
		return createErrorResponse("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	removedCount := 0
	for i := 1; i < len(args); i++ {
		member := args[i].Bulk
		if entry.SortedSet.Remove(member) {
			removedCount++
		}
	}

	return shared.Value{Typ: "integer", Num: removedCount}
}
