package commands

import (
	"fmt"
	"math"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// decodeGeohash converts a geohash score back to latitude and longitude
// This is the inverse of the encodeGeohash function
func decodeGeohash(score uint64) (latitude, longitude float64) {
	// Deinterleave the bits
	x := uint32(0)
	y := uint32(0)

	for i := 0; i < 32; i++ {
		if (score & (1 << (2 * i))) != 0 {
			x |= (1 << i)
		}
		if (score & (1 << (2*i + 1))) != 0 {
			y |= (1 << i)
		}
	}

	// Convert back to coordinates
	latitude = MIN_LATITUDE + (float64(x) * LATITUDE_RANGE / math.Pow(2, 26))
	longitude = MIN_LONGITUDE + (float64(y) * LONGITUDE_RANGE / math.Pow(2, 26))

	return latitude, longitude
}

// geopos handles the GEOPOS command.
// Usage: GEOPOS key member [member ...]
// Returns: An array with one entry for each requested member.
//
// This command returns the longitude and latitude of members in a sorted set.
// For each member:
// - If the member exists: returns [longitude, latitude] as bulk strings
// - If the member doesn't exist: returns null
// - If the key doesn't exist: returns null for all members
func Geopos(connID string, args []shared.Value) shared.Value {
	if len(args) < 2 {
		return createErrorResponse("ERR wrong number of arguments for 'geopos' command")
	}

	key := args[0].Bulk
	entry, exists := server.Memory[key]

	// Create result array with one entry for each requested member
	result := make([]shared.Value, len(args)-1)

	for i := 1; i < len(args); i++ {
		member := args[i].Bulk

		// If key doesn't exist or doesn't hold a sorted set, return null array
		if !exists || entry.SortedSet == nil {
			result[i-1] = shared.Value{Typ: "null_array", Str: ""}
			continue
		}

		// Get the score for this member
		score, memberExists := entry.SortedSet.GetScore(member)

		if !memberExists {
			// Member doesn't exist, return null array
			result[i-1] = shared.Value{Typ: "null_array", Str: ""}
			continue
		}

		// Decode the geohash score back to latitude and longitude
		latitude, longitude := decodeGeohash(uint64(score))

		// Return [longitude, latitude] as bulk strings
		coordinateArray := []shared.Value{
			{Typ: "bulk", Bulk: fmt.Sprintf("%.15g", longitude)},
			{Typ: "bulk", Bulk: fmt.Sprintf("%.15g", latitude)},
		}

		result[i-1] = shared.Value{Typ: "array", Array: coordinateArray}
	}

	return shared.Value{Typ: "array", Array: result}
}
