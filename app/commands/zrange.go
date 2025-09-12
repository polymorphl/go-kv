package commands

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// zrange handles the ZRANGE command.
// Usage: ZRANGE key start stop [WITHSCORES]
// Returns: Array of elements in the specified range.
//
// This command returns the specified elements of the sorted set stored at key.
// The offsets start and stop are zero-based indexes, with 0 being the first element.
// Negative offsets can be used to start from the end of the sorted set.
//
// Special cases:
//   - If start > stop, returns empty array
//   - If start < 0, treated as 0
//   - If stop >= sorted set length, treated as sorted set length - 1
//   - Both start and stop are inclusive
//
// Examples:
//
//	ZRANGE myzset 0 2      // Returns elements at index 0, 1, and 2
//	ZRANGE myzset 0 -1     // Returns all elements
//	ZRANGE myzset -3 -1    // Returns last 3 elements
func Zrange(connID string, args []protocol.Value) protocol.Value {
	if len(args) < 3 {
		return createErrorResponse("ERR wrong number of arguments for 'zrange' command")
	}

	key := args[0].Bulk
	entry, exists := server.Memory[key]

	if !exists {
		return shared.Value{Typ: "array", Array: []shared.Value{}}
	}

	if entry.SortedSet == nil {
		return shared.Value{Typ: "array", Array: []shared.Value{}}
	}

	start, err := strconv.Atoi(args[1].Bulk)
	if err != nil {
		return createErrorResponse("ERR value is not an integer or out of range")
	}
	stop, err := strconv.Atoi(args[2].Bulk)
	if err != nil {
		return createErrorResponse("ERR value is not an integer or out of range")
	}

	members := entry.SortedSet.GetSortedMembers()

	// Handle negative indices
	if start < 0 {
		start = len(members) + start
	}
	if stop < 0 {
		stop = len(members) + stop
	}

	// Clamp indices to valid range
	if start < 0 {
		start = 0
	}
	if stop >= len(members) {
		stop = len(members) - 1
	}

	// If start > stop, return empty array
	if start > stop {
		return shared.Value{Typ: "array", Array: []shared.Value{}}
	}

	result := make([]shared.Value, 0, stop-start+1)
	for i := start; i <= stop; i++ {
		result = append(result, shared.Value{Typ: "string", Str: members[i]})
	}

	return shared.Value{Typ: "array", Array: result}

}
