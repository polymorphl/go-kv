package commands

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
	"github.com/codecrafters-io/redis-starter-go/app/server"
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
	entry, exists := server.Memory[key]

	if !exists {
		return shared.Value{Typ: "null", Str: ""}
	}

	// Check if list is empty (either array or linked list)
	isEmpty := false
	if entry.List != nil {
		isEmpty = entry.List.Size == 0
	} else {
		isEmpty = len(entry.Array) == 0
	}

	if isEmpty {
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

	// Get the actual list size
	var listSize int
	if entry.List != nil {
		listSize = entry.List.Size
	} else {
		listSize = len(entry.Array)
	}

	// Limit count to the actual list length
	if count > listSize {
		count = listSize
	}

	// If count is 0, return empty array
	if count == 0 {
		return shared.Value{Typ: "array", Array: []shared.Value{}}
	}

	// If count is 1, return single string (backward compatibility)
	if count == 1 {
		var value string
		if entry.List != nil {
			value = entry.List.RemoveFromHead()
		} else {
			value = entry.Array[0]
			entry.Array = entry.Array[1:]
		}
		server.Memory[key] = entry
		return shared.Value{Typ: "string", Str: value}
	}

	// Pop multiple items and return as array
	result := make([]shared.Value, count)
	if entry.List != nil {
		// Use linked list
		for i := 0; i < count; i++ {
			result[i] = shared.Value{Typ: "string", Str: entry.List.RemoveFromHead()}
		}
	} else {
		// Use array
		for i := 0; i < count; i++ {
			result[i] = shared.Value{Typ: "string", Str: entry.Array[i]}
		}
		entry.Array = entry.Array[count:]
	}
	server.Memory[key] = entry

	return shared.Value{Typ: "array", Array: result}
}
