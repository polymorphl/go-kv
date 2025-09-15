package commands

import (
	"strconv"

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
	if len(args) < 4 || (len(args)-1)%3 != 0 {
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

	newElementsCount := 0

	// Process longitude-latitude-member triplets
	for i := 1; i < len(args); i += 3 {
		longitudeStr := args[i].Bulk
		latitudeStr := args[i+1].Bulk
		member := args[i+2].Bulk

		// Parse longitude
		longitude, err := strconv.ParseFloat(longitudeStr, 64)
		if err != nil {
			return createErrorResponse("ERR invalid longitude argument")
		}

		// Parse latitude
		latitude, err := strconv.ParseFloat(latitudeStr, 64)
		if err != nil {
			return createErrorResponse("ERR invalid latitude argument")
		}

		// Validate longitude range: -180° to +180° (inclusive)
		if longitude < -180.0 || longitude > 180.0 {
			return createErrorResponse("ERR invalid longitude value")
		}

		// Validate latitude range: -85.05112878° to +85.05112878° (inclusive)
		if latitude < -85.05112878 || latitude > 85.05112878 {
			return createErrorResponse("ERR invalid latitude value")
		}

		// For GEOADD, we use a hardcoded score of 0 for now
		// In later stages, we'll implement proper geohash scoring
		if entry.SortedSet.Add(member, 0) {
			newElementsCount++
		}
	}

	server.Memory[key] = entry
	return shared.Value{Typ: "integer", Num: newElementsCount}
}
