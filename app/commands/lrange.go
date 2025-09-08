package commands

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// lrange handles the LRANGE command.
// Usage: LRANGE key start stop
// Returns: Array of elements in the specified range.
//
// This command returns the specified elements of the list stored at key.
// The offsets start and stop are zero-based indexes, with 0 being the first element.
// Negative offsets can be used to start from the end of the list.
//
// Special cases:
//   - If start > stop, returns empty array
//   - If start < 0, treated as 0
//   - If stop >= list length, treated as list length - 1
//   - Both start and stop are inclusive
//
// Examples:
//
//	LRANGE mylist 0 2      // Returns elements at index 0, 1, and 2
//	LRANGE mylist 0 -1     // Returns all elements
//	LRANGE mylist -3 -1    // Returns last 3 elements
//	LRANGE mylist 5 3      // Returns empty array (start > stop)
func Lrange(args []shared.Value) shared.Value {
	if len(args) != 3 {
		return createErrorResponse("ERR wrong number of arguments for 'lrange' command")
	}

	key := args[0].Bulk
	entry, exists := shared.Memory[key]

	if !exists {
		return shared.Value{Typ: "array", Array: []shared.Value{}}
	}

	// Check if it's an array
	if len(entry.Array) == 0 {
		return createErrorResponse("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Parse start and stop indices
	start, err := strconv.Atoi(args[1].Bulk)
	if err != nil {
		return createErrorResponse("ERR value is not an integer or out of range")
	}
	stop, err := strconv.Atoi(args[2].Bulk)
	if err != nil {
		return createErrorResponse("ERR value is not an integer or out of range")
	}

	arrayLen := len(entry.Array)

	// Handle negative indices (count from end)
	if start < 0 {
		start = arrayLen + start
	}
	if stop < 0 {
		stop = arrayLen + stop
	}

	// Clamp indices to valid range
	if start < 0 {
		start = 0
	}
	if stop >= arrayLen {
		stop = arrayLen - 1
	}

	// Check if start is after stop (invalid range)
	if start > stop {
		return shared.Value{Typ: "array", Array: []shared.Value{}}
	}

	// Pre-allocate result slice with exact capacity for better performance
	result := make([]shared.Value, 0, stop-start+1)
	for _, value := range entry.Array[start : stop+1] {
		result = append(result, shared.Value{Typ: "string", Str: value})
	}

	return shared.Value{Typ: "array", Array: result}
}
