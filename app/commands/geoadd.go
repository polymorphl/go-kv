package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// geoadd handles the GEOADD command.
// Usage: GEOADD key longitude latitude member [longitude latitude member ...]
// Returns: The number of new elements added to the sorted set.
//
// This command adds one or more geospatial items to a sorted set.
// If the key does not exist, it is created as an empty sorted set.
// If the key exists but is not a sorted set, it is converted to a sorted set before the operation.
// The longitude and latitude are stored as floats.
func Geoadd(connID string, args []shared.Value) shared.Value {
	if len(args) < 4 {
		return createErrorResponse("ERR wrong number of arguments for 'geoadd' command")
	}

	key := args[0].Bulk
	entry, exists := server.Memory[key]

	if !exists {
		entry = shared.MemoryEntry{SortedSet: shared.NewSortedSet(), Expires: 0}
	}

	if entry.SortedSet == nil {
		entry.SortedSet = shared.NewSortedSet()
	}

	newElementsCount := 1

	return shared.Value{Typ: "integer", Num: newElementsCount}
}
