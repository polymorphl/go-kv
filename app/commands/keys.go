package commands

import (
	"path/filepath"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// Keys handles the KEYS command.
// Usage: KEYS pattern
// Returns: A RESP array of keys that match the given pattern.
//
// This command returns all keys in the database that match the specified pattern.
// The pattern uses glob-style wildcards:
// - * matches any number of characters (including zero)
// - ? matches exactly one character
// - [abc] matches any character in the set
// - [a-z] matches any character in the range
//
// Examples:
//
//	KEYS *           // Returns all keys
//	KEYS "f*"        // Returns keys starting with 'f'
//	KEYS "foo?"      // Returns keys like 'foo1', 'foo2', etc.
//	KEYS "test[0-9]" // Returns keys like 'test0', 'test1', etc.
func Keys(connID string, args []shared.Value) shared.Value {
	if len(args) != 1 {
		return createErrorResponse("ERR wrong number of arguments for 'keys' command")
	}

	pattern := args[0].Bulk
	var matchingKeys []string

	// Iterate through all keys in memory
	for key, entry := range shared.Memory {
		// Check if key has expired and skip it if so
		if entry.Expires > 0 && time.Now().UnixMilli() > entry.Expires {
			continue
		}

		// Check if key matches the pattern
		matched, err := filepath.Match(pattern, key)
		if err != nil {
			// If pattern is invalid, return error
			return createErrorResponse("ERR invalid pattern")
		}
		if matched {
			matchingKeys = append(matchingKeys, key)
		}
	}

	// Convert to RESP array format
	result := make([]shared.Value, len(matchingKeys))
	for i, key := range matchingKeys {
		result[i] = shared.Value{Typ: "bulk", Bulk: key}
	}

	return shared.Value{Typ: "array", Array: result}
}
