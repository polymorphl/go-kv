package commands

import "github.com/codecrafters-io/redis-starter-go/app/shared"

// createErrorResponse creates a standardized error response.
func createErrorResponse(message string) shared.Value {
	return shared.Value{Typ: "error", Str: message}
}

// push handles the common logic for both LPUSH and RPUSH commands.
// Usage: push key value [value ...] (prepend bool)
// Returns: The length of the list after the push operation.
//
// This function inserts all the specified values either at the head (prepend=true)
// or tail (prepend=false) of the list stored at key.
// If key doesn't exist, it is created as an empty list before performing the push operation.
// If key exists but is not a list, it is converted to a list before the operation.
func push(args []shared.Value, prepend bool) shared.Value {
	if len(args) < 2 {
		cmdName := "lpush"
		if !prepend {
			cmdName = "rpush"
		}
		return createErrorResponse("ERR wrong number of arguments for '" + cmdName + "' command")
	}

	key := args[0].Bulk
	entry, exists := shared.Memory[key]

	// If key doesn't exist or is not an array, create a new array
	if !exists || len(entry.Array) == 0 && entry.Value != "" {
		entry = shared.MemoryEntry{Array: []string{}, Expires: 0}
	}

	// Add all values to the array
	for i := 1; i < len(args); i++ {
		if prepend {
			// Add to beginning (LPUSH behavior)
			entry.Array = append([]string{args[i].Bulk}, entry.Array...)
		} else {
			// Add to end (RPUSH behavior)
			entry.Array = append(entry.Array, args[i].Bulk)
		}
	}

	shared.Memory[key] = entry
	return shared.Value{Typ: "integer", Num: len(entry.Array)}
}
