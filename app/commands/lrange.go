package commands

import (
	"strconv"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// Object pool for result slices to reduce allocations
var (
	lrangeResultPool = sync.Pool{
		New: func() interface{} {
			return make([]shared.Value, 0, 16)
		},
	}
)

// Helper functions for pool management
func getLrangeResult() []shared.Value {
	return lrangeResultPool.Get().([]shared.Value)
}

func putLrangeResult(s []shared.Value) {
	s = s[:0] // Reset length but keep capacity
	lrangeResultPool.Put(s)
}

// getRangeFromArray efficiently extracts a range from an array without copying
func getRangeFromArray(arr []string, start, stop int) []string {
	if start > stop || start >= len(arr) || stop < 0 {
		return []string{}
	}

	// Clamp indices to array bounds
	if start < 0 {
		start = 0
	}
	if stop >= len(arr) {
		stop = len(arr) - 1
	}

	return arr[start : stop+1]
}

// getRangeFromLinkedList efficiently extracts a range from a linked list without full conversion
func getRangeFromLinkedList(ll *shared.LinkedList, start, stop int) []string {
	if start > stop || start >= ll.Size || stop < 0 {
		return []string{}
	}

	// Clamp indices to list bounds
	if start < 0 {
		start = 0
	}
	if stop >= ll.Size {
		stop = ll.Size - 1
	}

	// Pre-allocate result with exact capacity
	result := make([]string, 0, stop-start+1)

	// Navigate to start position
	current := ll.Head
	for i := 0; i < start; i++ {
		current = current.Next
	}

	// Extract range elements
	for i := start; i <= stop; i++ {
		result = append(result, current.Value)
		current = current.Next
	}

	return result
}

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
func Lrange(connID string, args []shared.Value) shared.Value {
	if len(args) != 3 {
		return createErrorResponse("ERR wrong number of arguments for 'lrange' command")
	}

	key := args[0].Bulk
	entry, exists := server.Memory[key]

	if !exists {
		return shared.Value{Typ: "array", Array: []shared.Value{}}
	}

	// Check if it's a list (either array or linked list)
	if entry.List == nil && len(entry.Array) == 0 {
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

	// Get list length and handle negative indices
	var listLen int
	if entry.List != nil {
		listLen = entry.List.Size
	} else {
		listLen = len(entry.Array)
	}

	// Handle negative indices (count from end)
	if start < 0 {
		start = listLen + start
	}
	if stop < 0 {
		stop = listLen + stop
	}

	// Check if start is after stop (invalid range)
	if start > stop {
		return shared.Value{Typ: "array", Array: []shared.Value{}}
	}

	// Get range elements efficiently
	var rangeElements []string
	if entry.List != nil {
		rangeElements = getRangeFromLinkedList(entry.List, start, stop)
	} else {
		rangeElements = getRangeFromArray(entry.Array, start, stop)
	}

	// Use object pool for result slice
	result := getLrangeResult()
	defer putLrangeResult(result)

	// Pre-allocate with exact capacity
	result = result[:0]
	if cap(result) < len(rangeElements) {
		result = make([]shared.Value, 0, len(rangeElements))
	}

	// Convert to shared.Value array efficiently
	for _, value := range rangeElements {
		result = append(result, shared.Value{Typ: "string", Str: value})
	}

	return shared.Value{Typ: "array", Array: result}
}
