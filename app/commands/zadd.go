package commands

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// zadd handles the ZADD command.
// Usage: ZADD score member [score member ...]
// Returns: The number of new elements added to the sorted set.
//
// This command adds one or more members to a sorted set, or updates the score of an existing member.
// If the key does not exist, a new sorted set is created.
//
// Examples:
//
//	ZADD myzset 1 "one"                        // Adds one element to the sorted set
//	ZADD myzset 1 "one" 2 "two"                // Adds two elements to the sorted set
//	ZADD myzset 1 "one" 2 "two" 3 "three"      // Adds three elements to the sorted set
func Zadd(connID string, args []protocol.Value) protocol.Value {
	if len(args) < 3 {
		return createErrorResponse("ERR wrong number of arguments for 'zadd' command")
	}

	// Check if we have an even number of score-member pairs (excluding the key)
	if (len(args)-1)%2 != 0 {
		return createErrorResponse("ERR wrong number of arguments for 'zadd' command")
	}

	key := args[0].Bulk
	entry, exists := server.Memory[key]

	if !exists {
		entry = shared.MemoryEntry{SortedSet: shared.NewSortedSet(), Expires: 0}
	}

	// Ensure the entry has a sorted set
	if entry.SortedSet == nil {
		entry.SortedSet = shared.NewSortedSet()
	}

	newElementsCount := 0

	// Process score-member pairs
	for i := 1; i < len(args); i += 2 {
		scoreStr := args[i].Bulk
		member := args[i+1].Bulk

		// Parse score
		score, err := strconv.ParseFloat(scoreStr, 64)
		if err != nil {
			return createErrorResponse("ERR value is not a valid float")
		}

		// Add member to sorted set
		if entry.SortedSet.Add(member, score) {
			newElementsCount++
		}
	}

	// Update the entry in memory
	server.Memory[key] = entry

	return shared.Value{Typ: "integer", Num: newElementsCount}
}
