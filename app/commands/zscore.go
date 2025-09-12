package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// zscore handles the ZSCORE command.
// Usage: ZSCORE key member
// Returns: The score of the member in the sorted set, or null if the member does not exist.
//
// This command returns the score of a member in a sorted set.
// If the member does not exist in the sorted set, null is returned.
// If the key does not exist, null is returned.
// If the key exists but does not hold a sorted set, null is returned.
//
// Examples:
//
//	ZSCORE myzset "one"                 // Returns the score of "one" in myzset
//	ZSCORE nonexistent "member"         // Returns null (key doesn't exist)
//	ZSCORE myzset "member"              // Returns null (member doesn't exist)
//	ZSCORE myzset "member"              // Returns null (key doesn't hold a sorted set)
func Zscore(connID string, args []shared.Value) shared.Value {
	if len(args) != 2 {
		return createErrorResponse("ERR wrong number of arguments for 'zscore' command")
	}

	key := args[0].Bulk
	entry, exists := server.Memory[key]

	if !exists {
		return shared.Value{Typ: "null", Str: ""}
	}

	score, exists := entry.SortedSet.GetScore(args[1].Bulk)
	if !exists {
		return shared.Value{Typ: "null", Str: ""}
	}

	return shared.Value{Typ: "bulk", Bulk: fmt.Sprintf("%.15g", score)}
}
