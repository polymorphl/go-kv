package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// zrank handles the ZRANK command.
// Usage: ZRANK key member
// Returns: The rank of the member in the sorted set, or null if the member does not exist.
//
// This command returns the rank of a member in a sorted set.
// If the member does not exist in the sorted set, null is returned.
// If the key does not exist, null is returned.
// If the key exists but does not hold a sorted set, null is returned.
//
// Examples:
//
//	ZRANK myzset "one"                 // Returns the rank of "one" in myzset
//	ZRANK nonexistent "member"         // Returns null (key doesn't exist)
//	ZRANK myzset "member"              // Returns null (member doesn't exist)
//	ZRANK myzset "member"              // Returns null (key doesn't hold a sorted set)
func Zrank(connID string, args []protocol.Value) protocol.Value {
	if len(args) != 2 {
		return createErrorResponse("ERR wrong number of arguments for 'zrank' command")
	}

	key := args[0].Bulk
	entry, exists := server.Memory[key]

	if !exists {
		return shared.Value{Typ: "null", Str: ""}
	}

	if entry.SortedSet == nil {
		return shared.Value{Typ: "null", Str: ""}
	}

	member := args[1].Bulk
	rank, exists := entry.SortedSet.GetRank(member)
	if !exists {
		return shared.Value{Typ: "null", Str: ""}
	}

	return shared.Value{Typ: "integer", Num: rank}
}
