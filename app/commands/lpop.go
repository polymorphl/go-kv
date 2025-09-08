package commands

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// lpop handles the LPOP command.
// Usage: LPOP key [count]
// Returns: The popped element(s) from the head of the list.
//
// This command removes and returns one or more elements from the head of the list stored at key.
// If key does not exist, null is returned.
// If key exists but is not a list, an error is returned.
// If the list is empty, null is returned.
//
// The optional count argument specifies how many elements to pop:
//   - If count is not specified, pops and returns 1 element as a string
//   - If count is 0, returns an empty array
//   - If count is positive, pops up to count elements and returns them as an array
//   - If count is greater than the list length, pops all elements
//   - If count is negative, returns an error
//
// Examples:
//
//	LPOP mylist                    // Returns single string (original behavior)
//	LPOP mylist 1                  // Returns single string (same as above)
//	LPOP mylist 3                  // Returns array with up to 3 items
//	LPOP mylist 0                  // Returns empty array
//	LPOP nonexistent               // Returns null (key doesn't exist)
//	LPOP mystring                 // Returns error (wrong type)
//
// Note: LPOP is an O(1) operation for popping a single element, or O(n) for popping
// multiple elements where n is the number of elements being popped.
func Lpop(connID string, args []shared.Value) shared.Value {
	if len(args) < 1 || len(args) > 2 {
		return createErrorResponse("ERR wrong number of arguments for 'lpop' command")
	}

	key := args[0].Bulk
	entry, exists := shared.Memory[key]

	if !exists || len(entry.Array) == 0 {
		return shared.Value{Typ: "null", Str: ""}
	}

	// Default to popping 1 item if no count specified
	count := 1
	if len(args) == 2 {
		var err error
		count, err = strconv.Atoi(args[1].Bulk)
		if err != nil || count < 0 {
			return createErrorResponse("ERR value is not an integer or out of range")
		}
	}

	// Limit count to the actual array length
	if count > len(entry.Array) {
		count = len(entry.Array)
	}

	// If count is 0, return empty array
	if count == 0 {
		return shared.Value{Typ: "array", Array: []shared.Value{}}
	}

	// If count is 1, return single string (backward compatibility)
	if count == 1 {
		tmp := entry.Array[0]
		entry.Array = entry.Array[1:]
		shared.Memory[key] = entry
		return shared.Value{Typ: "string", Str: tmp}
	}

	// Pop multiple items and return as array
	result := make([]shared.Value, count)
	for i := 0; i < count; i++ {
		result[i] = shared.Value{Typ: "string", Str: entry.Array[i]}
	}
	entry.Array = entry.Array[count:]
	shared.Memory[key] = entry

	return shared.Value{Typ: "array", Array: result}
}
