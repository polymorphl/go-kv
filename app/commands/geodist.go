package commands

import (
	"fmt"
	"math"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// Constants from Redis geohash helper
const (
	DEG_TO_RAD             = 0.017453292519943295769236907684886
	EARTH_RADIUS_IN_METERS = 6372797.560856
)

// deg_rad converts degrees to radians
func deg_rad(ang float64) float64 {
	return ang * DEG_TO_RAD
}

// geohashGetLatDistance calculates distance using simplified haversine when longitude diff is 0
func geohashGetLatDistance(lat1d, lat2d float64) float64 {
	return EARTH_RADIUS_IN_METERS * math.Abs(deg_rad(lat2d)-deg_rad(lat1d))
}

// geohashGetDistance calculates distance using haversine great circle distance formula
// This is the exact implementation from Redis geohash helper
func geohashGetDistance(lon1d, lat1d, lon2d, lat2d float64) float64 {
	lon1r := deg_rad(lon1d)
	lon2r := deg_rad(lon2d)
	v := math.Sin((lon2r - lon1r) / 2)

	// if v == 0 we can avoid doing expensive math when lons are practically the same
	if v == 0.0 {
		return geohashGetLatDistance(lat1d, lat2d)
	}

	lat1r := deg_rad(lat1d)
	lat2r := deg_rad(lat2d)
	u := math.Sin((lat2r - lat1r) / 2)
	a := u*u + math.Cos(lat1r)*math.Cos(lat2r)*v*v
	return 2.0 * EARTH_RADIUS_IN_METERS * math.Asin(math.Sqrt(a))
}

// Geodist handles the GEODIST command.
// Usage: GEODIST key member1 member2 [unit]
// Returns: The distance between the two members in meters (or specified unit) as a bulk string.
//
// This command returns the distance between two members of a geospatial sorted set.
// The distance is calculated using the Haversine formula.
// If either member doesn't exist, returns null.
// If the key doesn't exist, returns null.
func Geodist(connID string, args []shared.Value) shared.Value {
	if len(args) < 3 || len(args) > 4 {
		return createErrorResponse("ERR wrong number of arguments for 'geodist' command")
	}

	key := args[0].Bulk
	member1 := args[1].Bulk
	member2 := args[2].Bulk

	// Optional unit parameter (defaults to meters)
	unit := "m"
	if len(args) == 4 {
		unit = args[3].Bulk
	}

	entry, exists := server.Memory[key]
	if !exists || entry.SortedSet == nil {
		return shared.Value{Typ: "null_bulk", Str: ""}
	}

	// Get scores for both members
	score1, member1Exists := entry.SortedSet.GetScore(member1)
	score2, member2Exists := entry.SortedSet.GetScore(member2)

	if !member1Exists || !member2Exists {
		return shared.Value{Typ: "null_bulk", Str: ""}
	}

	// Decode geohash scores to get coordinates
	lat1, lon1 := decodeGeohash(uint64(score1))
	lat2, lon2 := decodeGeohash(uint64(score2))

	// Calculate distance using Redis geohash helper implementation
	distance := geohashGetDistance(lon1, lat1, lon2, lat2)

	// Convert to requested unit
	switch unit {
	case "m", "meter", "meters":
		// Already in meters
	case "km", "kilometer", "kilometers":
		distance = distance / 1000
	case "mi", "mile", "miles":
		distance = distance / 1609.344
	case "ft", "feet":
		distance = distance * 3.28084
	default:
		return createErrorResponse("ERR unsupported unit provided")
	}

	// Format the result with appropriate precision
	result := fmt.Sprintf("%.4f", distance)
	return shared.Value{Typ: "bulk", Bulk: result}
}
