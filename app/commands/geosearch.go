package commands

import (
	"sort"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// Geosearch handles the GEOSEARCH command.
// Usage: GEOSEARCH key FROMLONLAT longitude latitude BYRADIUS radius unit
// Returns: A RESP array of member names that are within the specified radius.
//
// This command searches for members in a geospatial sorted set within a circular area.
// Only supports FROMLONLAT and BYRADIUS options in this implementation.
func Geosearch(connID string, args []shared.Value) shared.Value {
	if len(args) < 6 {
		return createErrorResponse("ERR wrong number of arguments for 'geosearch' command")
	}

	// Parse FROMLONLAT longitude latitude
	if args[1].Bulk != "FROMLONLAT" {
		return createErrorResponse("ERR only FROMLONLAT mode is supported")
	}

	longitude, err := strconv.ParseFloat(args[2].Bulk, 64)
	if err != nil {
		return createErrorResponse("ERR invalid longitude")
	}

	latitude, err := strconv.ParseFloat(args[3].Bulk, 64)
	if err != nil {
		return createErrorResponse("ERR invalid latitude")
	}

	// Parse BYRADIUS radius unit
	if args[4].Bulk != "BYRADIUS" {
		return createErrorResponse("ERR only BYRADIUS mode is supported")
	}

	radius, err := strconv.ParseFloat(args[5].Bulk, 64)
	if err != nil {
		return createErrorResponse("ERR invalid radius")
	}

	unit := "m" // default to meters
	if len(args) > 6 {
		unit = args[6].Bulk
	}

	key := args[0].Bulk
	entry, exists := server.Memory[key]

	if !exists || entry.SortedSet == nil {
		return shared.Value{Typ: "array", Array: []shared.Value{}}
	}

	// Convert radius to meters
	radiusInMeters := convertToMeters(radius, unit)

	// Search for members within the radius
	var results []shared.Value
	for member, score := range entry.SortedSet.Members {
		memberLat, memberLon := decodeGeohash(uint64(score))

		// Calculate distance from search center to this member
		distance := geohashGetDistance(longitude, latitude, memberLon, memberLat)

		// If within radius, add to results
		if distance <= radiusInMeters {
			results = append(results, shared.Value{Typ: "bulk", Bulk: member})
		}
	}

	// Sort results alphabetically for consistent output
	sort.Slice(results, func(i, j int) bool {
		return results[i].Bulk < results[j].Bulk
	})

	return shared.Value{Typ: "array", Array: results}
}

// convertToMeters converts a distance value to meters based on the unit
func convertToMeters(distance float64, unit string) float64 {
	switch unit {
	case "m", "meter", "meters":
		return distance
	case "km", "kilometer", "kilometers":
		return distance * 1000
	case "mi", "mile", "miles":
		return distance * 1609.344
	case "ft", "feet", "foot":
		return distance * 0.3048
	default:
		return distance // default to meters
	}
}
