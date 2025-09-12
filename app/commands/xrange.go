package commands

import (
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// compareStrings performs a simple string comparison.
func compareStrings(s1, s2 string) int {
	if s1 < s2 {
		return -1
	} else if s1 > s2 {
		return 1
	}
	return 0
}

// compareStreamIDs compares two Redis stream IDs alphabetically.
func compareStreamIDs(id1, id2 string) int {
	// Handle special Redis stream IDs
	if id1 == "$" || id2 == "$" {
		// $ means "end of stream" - any real ID is greater than $
		if id1 == "$" && id2 != "$" {
			return -1
		} else if id1 != "$" && id2 == "$" {
			return 1
		} else {
			return 0 // both are $
		}
	}

	// Parse both IDs into timestamp-sequence format
	parts1 := strings.Split(id1, "-")
	parts2 := strings.Split(id2, "-")

	// Validate ID format
	if len(parts1) != 2 || len(parts2) != 2 {
		return compareStrings(id1, id2)
	}

	// Parse timestamps
	timestamp1, err1 := strconv.ParseInt(parts1[0], 10, 64)
	timestamp2, err2 := strconv.ParseInt(parts2[0], 10, 64)

	if err1 != nil || err2 != nil {
		return compareStrings(id1, id2)
	}

	// Compare timestamps first
	if timestamp1 < timestamp2 {
		return -1
	} else if timestamp1 > timestamp2 {
		return 1
	}

	// Timestamps are equal, compare sequences
	sequence1, err1 := strconv.ParseInt(parts1[1], 10, 64)
	sequence2, err2 := strconv.ParseInt(parts2[1], 10, 64)

	if err1 != nil || err2 != nil {
		return compareStrings(id1, id2)
	}

	if sequence1 < sequence2 {
		return -1
	} else if sequence1 > sequence2 {
		return 1
	}

	return 0
}

// isInRange checks if a stream ID is within the specified range (inclusive).
func isInRange(id, start, end string) bool {
	// Handle special Redis range values
	if start == "-" && end == "+" {
		return true // All entries
	}
	if start == "-" {
		return compareStreamIDs(id, end) <= 0
	}
	if end == "+" {
		return compareStreamIDs(id, start) >= 0
	}

	// Both are specific IDs - check if id is between start and end (inclusive)
	return compareStreamIDs(id, start) >= 0 && compareStreamIDs(id, end) <= 0
}

// createStreamEntryValue creates a RESP array value representing a stream entry.
func createStreamEntryValue(entry shared.StreamEntry) shared.Value {
	// Create field-value array
	var fieldValueArray []shared.Value
	for field, value := range entry.Data {
		fieldValueArray = append(fieldValueArray,
			shared.Value{Typ: "bulk", Bulk: field},
			shared.Value{Typ: "bulk", Bulk: value},
		)
	}

	// Create entry array: [ID, [field-value pairs]]
	entryArray := []shared.Value{
		{Typ: "bulk", Bulk: entry.ID},
		{Typ: "array", Array: fieldValueArray},
	}

	return shared.Value{Typ: "array", Array: entryArray}
}

// xrange handles the XRANGE command.
//
// Examples:
//
//	XRANGE mystream 1526985054069 1526985054079    // Range between specific IDs
//	XRANGE mystream - +                            // All entries
//	XRANGE mystream 1526985054069 +                // From specific ID to end
func Xrange(connID string, args []shared.Value) shared.Value {
	if len(args) < 3 {
		return createErrorResponse("ERR wrong number of arguments for 'xrange' command")
	}

	key := args[0].Bulk
	start := args[1].Bulk
	end := args[2].Bulk

	entry, exists := server.Memory[key]
	if !exists {
		// Empty stream - return empty array
		return shared.Value{Typ: "array", Array: []shared.Value{}}
	}

	var result []shared.Value

	for _, streamEntry := range entry.Stream {
		if isInRange(streamEntry.ID, start, end) {
			entryValue := createStreamEntryValue(streamEntry)
			result = append(result, entryValue)
		}
	}

	return shared.Value{Typ: "array", Array: result}
}
