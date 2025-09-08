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
	if prepend {
		// LPUSH: Optimized approach - pre-allocate and use copy for efficiency
		newCount := len(args) - 1
		if newCount == 0 {
			return shared.Value{Typ: "integer", Num: len(entry.Array)}
		}

		// For single value, use simple prepend
		if newCount == 1 {
			entry.Array = append([]string{args[1].Bulk}, entry.Array...)
			shared.Memory[key] = entry
			return shared.Value{Typ: "integer", Num: len(entry.Array)}
		}

		// For multiple values, use efficient bulk prepend
		oldCount := len(entry.Array)
		totalCount := newCount + oldCount

		// Pre-allocate with exact capacity to avoid reallocations
		newArray := make([]string, 0, totalCount)

		// Add new values in reverse order (Redis LPUSH behavior)
		for i := len(args) - 1; i >= 1; i-- {
			newArray = append(newArray, args[i].Bulk)
		}

		// Add existing values
		newArray = append(newArray, entry.Array...)

		entry.Array = newArray
	} else {
		// RPUSH: Add values one by one to the end
		for i := 1; i < len(args); i++ {
			entry.Array = append(entry.Array, args[i].Bulk)
		}
	}

	shared.Memory[key] = entry
	return shared.Value{Typ: "integer", Num: len(entry.Array)}
}
