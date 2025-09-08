package commands

import "github.com/codecrafters-io/redis-starter-go/app/shared"

// typeCmd handles the TYPE command.
// Usage: TYPE key
// Returns: The type of the value stored at key.
//
// This command returns the type of the value stored at key.
// The type can be one of: string, list, set, zset, hash, stream, or none.
// If the key does not exist, "none" is returned.
//
// Examples:
//
//	TYPE mykey                    // Returns the type of mykey
//	TYPE nonexistent              // Returns "none" (key doesn't exist)
//	TYPE mystring                 // Returns "string"
//	TYPE mylist                   // Returns "list"
func Type(connID string, args []shared.Value) shared.Value {
	if len(args) != 1 {
		return createErrorResponse("ERR wrong number of arguments for 'type' command")
	}

	key := args[0].Bulk
	entry, exists := shared.Memory[key]

	if !exists {
		return shared.Value{Typ: "string", Str: "none"}
	}

	// Determine the actual type based on the entry structure
	switch {
	case len(entry.Array) > 0:
		return shared.Value{Typ: "string", Str: "list"}
	case len(entry.Stream) > 0:
		return shared.Value{Typ: "string", Str: "stream"}
	case entry.Value != "":
		return shared.Value{Typ: "string", Str: "string"}
	default:
		return shared.Value{Typ: "string", Str: "none"}
	}
}
