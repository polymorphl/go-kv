package commands

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

const (
	MIN_LATITUDE  = -85.05112878
	MAX_LATITUDE  = 85.05112878
	MIN_LONGITUDE = -180.0
	MAX_LONGITUDE = 180.0

	LATITUDE_RANGE  = MAX_LATITUDE - MIN_LATITUDE
	LONGITUDE_RANGE = MAX_LONGITUDE - MIN_LONGITUDE
)

// spreadInt32ToInt64 spreads the bits of a 32-bit integer to a 64-bit integer
// This is used for interleaving latitude and longitude bits
func spreadInt32ToInt64(v uint32) uint64 {
	result := uint64(v)
	result = (result | (result << 16)) & 0x0000FFFF0000FFFF
	result = (result | (result << 8)) & 0x00FF00FF00FF00FF
	result = (result | (result << 4)) & 0x0F0F0F0F0F0F0F0F
	result = (result | (result << 2)) & 0x3333333333333333
	result = (result | (result << 1)) & 0x5555555555555555
	return result
}

// interleave interleaves two 32-bit integers into a single 64-bit integer
// This creates a geohash-like encoding where latitude and longitude bits are interleaved
func interleave(x, y uint32) uint64 {
	xSpread := spreadInt32ToInt64(x)
	ySpread := spreadInt32ToInt64(y)
	yShifted := ySpread << 1
	return xSpread | yShifted
}

// encodeGeohash converts latitude and longitude to a geohash score
// This is the core algorithm for storing geospatial data in sorted sets
func encodeGeohash(latitude, longitude float64) uint64 {
	// Use integer arithmetic for better precision
	const scale = 1 << 26 // 2^26 as integer

	// Normalize to the range 0-2^26
	normalizedLatitude := float64(scale) * (latitude - MIN_LATITUDE) / LATITUDE_RANGE
	normalizedLongitude := float64(scale) * (longitude - MIN_LONGITUDE) / LONGITUDE_RANGE

	// Convert to integers (truncate, don't round)
	latInt := uint32(normalizedLatitude)
	lonInt := uint32(normalizedLongitude)

	return interleave(latInt, lonInt)
}

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
		if longitude < MIN_LONGITUDE || longitude > MAX_LONGITUDE {
			return createErrorResponse("ERR invalid longitude value")
		}

		// Validate latitude range: -85.05112878° to +85.05112878° (inclusive)
		if latitude < MIN_LATITUDE || latitude > MAX_LATITUDE {
			return createErrorResponse("ERR invalid latitude value")
		}

		// Convert latitude and longitude to geohash score
		score := encodeGeohash(latitude, longitude)

		// Add member to sorted set with geohash score
		if entry.SortedSet.Add(member, float64(score)) {
			newElementsCount++
		}
	}

	server.Memory[key] = entry
	return shared.Value{Typ: "integer", Num: newElementsCount}
}
