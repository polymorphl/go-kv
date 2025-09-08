package commands

import (
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// get handles the GET command.
// Usage: GET key
// Returns: The string value of the key, null if key doesn't exist, or error if key holds wrong type.
//
// This command retrieves the value of a key. It only works with string values.
// If the key holds an array (from RPUSH operations), it returns a WRONGTYPE error.
// Expired keys are automatically removed and return null.
//
// Examples:
//
//	GET mykey           // Returns the value of mykey
//	GET nonexistent     // Returns null
func Get(args []shared.Value) shared.Value {
	if len(args) != 1 {
		return createErrorResponse("ERR wrong number of arguments for 'get' command")
	}

	key := args[0].Bulk
	entry, exists := shared.Memory[key]

	if !exists {
		return shared.Value{Typ: "null", Str: ""}
	}

	// Check if key has expired and remove it if so
	if entry.Expires > 0 && time.Now().UnixMilli() > entry.Expires {
		delete(shared.Memory, key)
		return shared.Value{Typ: "null", Str: ""}
	}

	// GET only works with string values, not arrays
	if len(entry.Array) > 0 {
		return createErrorResponse("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return shared.Value{Typ: "string", Str: entry.Value}
}
