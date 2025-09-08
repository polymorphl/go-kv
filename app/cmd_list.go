package main

import (
	"strconv"
	"time"
)

// push handles the common logic for both LPUSH and RPUSH commands.
// Usage: push key value [value ...] (prepend bool)
// Returns: The length of the list after the push operation.
//
// This function inserts all the specified values either at the head (prepend=true)
// or tail (prepend=false) of the list stored at key.
// If key doesn't exist, it is created as an empty list before performing the push operation.
// If key exists but is not a list, it is converted to a list before the operation.
func push(args []Value, prepend bool) Value {
	if len(args) < 2 {
		cmdName := "lpush"
		if !prepend {
			cmdName = "rpush"
		}
		return createErrorResponse("ERR wrong number of arguments for '" + cmdName + "' command")
	}

	key := args[0].Bulk
	entry, exists := memory[key]

	// If key doesn't exist or is not an array, create a new array
	if !exists || len(entry.Array) == 0 && entry.Value != "" {
		entry = MemoryEntry{Array: []string{}, Expires: 0}
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

	memory[key] = entry
	return Value{Typ: "integer", Num: len(entry.Array)}
}

// lpush handles the LPUSH command.
// Usage: LPUSH key value [value ...]
// Returns: The length of the list after the push operation.
//
// This command inserts all the specified values at the head of the list stored at key.
// If key does not exist, it is created as an empty list before performing the push operation.
// If key exists but is not a list, it is converted to a list before the operation.
// Values are inserted in reverse order, so the last value becomes the first element.
//
// Examples:
//
//	LPUSH mylist "one"                    // Creates list with one element, returns 1
//	LPUSH mylist "two" "three"            // Adds two elements, returns 3
//	LPUSH newlist "first" "second"        // Creates new list, returns 2
//
// Note: LPUSH is the opposite of RPUSH - it adds elements to the beginning of the list,
// while RPUSH adds them to the end. The order of elements in the final list will be
// reversed compared to the order they were pushed.
func lpush(args []Value) Value {
	return push(args, true)
}

// rpush handles the RPUSH command.
// Usage: RPUSH key value [value ...]
// Returns: The length of the list after the push operation.
//
// This command inserts all the specified values at the tail of the list stored at key.
// If key does not exist, it is created as an empty list before performing the push operation.
// If key exists but is not a list, it is converted to a list before the operation.
//
// Examples:
//
//	RPUSH mylist "one"                    // Creates list with one element, returns 1
//	RPUSH mylist "two" "three"            // Adds two elements, returns 3
//	RPUSH newlist "first" "second"        // Creates new list, returns 2
func rpush(args []Value) Value {
	return push(args, false)
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
func lrange(args []Value) Value {
	if len(args) != 3 {
		return createErrorResponse("ERR wrong number of arguments for 'lrange' command")
	}

	key := args[0].Bulk
	entry, exists := memory[key]

	if !exists {
		return Value{Typ: "array", Array: []Value{}}
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
		return Value{Typ: "array", Array: []Value{}}
	}

	// Pre-allocate result slice with exact capacity for better performance
	result := make([]Value, 0, stop-start+1)
	for _, value := range entry.Array[start : stop+1] {
		result = append(result, Value{Typ: "string", Str: value})
	}

	return Value{Typ: "array", Array: result}
}

// llen handles the LLEN command.
// Usage: LLEN key
// Returns: The length of the list stored at key.
//
// This command returns the length of the list stored at key.
// If key does not exist, it is treated as an empty list and 0 is returned.
// If key exists but is not a list, an error is returned.
//
// Examples:
//
//	LLEN mylist                    // Returns the length of mylist
//	LLEN nonexistent              // Returns 0 (key doesn't exist)
//	LLEN mystring                 // Returns error (wrong type)
//
// Note: LLEN is a fast O(1) operation that simply returns the current length
// of the list without traversing its contents.
func llen(args []Value) Value {
	if len(args) != 1 {
		return createErrorResponse("ERR wrong number of arguments for 'llen' command")
	}

	key := args[0].Bulk
	entry, exists := memory[key]

	if !exists {
		return Value{Typ: "integer", Num: 0}
	}

	return Value{Typ: "integer", Num: len(entry.Array)}
}

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
func lpop(args []Value) Value {
	if len(args) < 1 || len(args) > 2 {
		return createErrorResponse("ERR wrong number of arguments for 'lpop' command")
	}

	key := args[0].Bulk
	entry, exists := memory[key]

	if !exists || len(entry.Array) == 0 {
		return Value{Typ: "null", Str: ""}
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
		return Value{Typ: "array", Array: []Value{}}
	}

	// If count is 1, return single string (backward compatibility)
	if count == 1 {
		tmp := entry.Array[0]
		entry.Array = entry.Array[1:]
		memory[key] = entry
		return Value{Typ: "string", Str: tmp}
	}

	// Pop multiple items and return as array
	result := make([]Value, count)
	for i := 0; i < count; i++ {
		result[i] = Value{Typ: "string", Str: entry.Array[i]}
	}
	entry.Array = entry.Array[count:]
	memory[key] = entry

	return Value{Typ: "array", Array: result}
}

// blpop handles the BLPOP command.
// Usage: BLPOP key [key ...] timeout
// Returns: The popped element from the head of the first non-empty list.
//
// This command is a blocking variant of LPOP. It blocks the client until an element
// becomes available on one of the specified lists, or until the timeout is reached.
// If timeout is 0, the command blocks indefinitely.
//
// The timeout can be specified as an integer (seconds) or float (fractional seconds).
// The command returns a two-element array containing the key name and the popped value.
// If timeout is reached before an element becomes available, null is returned.
//
// Examples:
//
//	BLPOP mylist 5                    // Wait up to 5 seconds for an element
//	BLPOP list1 list2 10              // Wait up to 10 seconds on either list
//	BLPOP mylist 0                    // Wait indefinitely
//	BLPOP mylist 0.1                  // Wait up to 0.1 seconds (100ms)
//	BLPOP mylist 1.5                  // Wait up to 1.5 seconds
//	BLPOP nonexistent 1               // Returns null after 1 second timeout
//
// Note: This implementation uses polling to check for available elements every 100ms.
// For production use, consider implementing an event-driven approach for better performance.
func blpop(args []Value) Value {
	if len(args) < 2 {
		return createErrorResponse("ERR wrong number of arguments for 'blpop' command")
	}

	// Last argument is the timeout (can be integer or float)
	timeoutStr := args[len(args)-1].Bulk
	timeout, err := strconv.ParseFloat(timeoutStr, 64)
	if err != nil || timeout < 0 {
		return createErrorResponse("ERR timeout is not a float or out of range")
	}

	// Helper function to check and pop from any available list
	checkAndPop := func() *Value {
		for i := 0; i < len(args)-1; i++ {
			key := args[i].Bulk
			entry, exists := memory[key]

			if exists && len(entry.Array) > 0 {
				// Found a non-empty list, pop the first element
				value := entry.Array[0]
				entry.Array = entry.Array[1:]
				memory[key] = entry

				// Return [key, value] array
				return &Value{Typ: "array", Array: []Value{
					{Typ: "string", Str: key},
					{Typ: "string", Str: value},
				}}
			}
		}
		return nil
	}

	// First, check if any list has elements available immediately
	if result := checkAndPop(); result != nil {
		return *result
	}

	// No elements available, block until timeout or element becomes available
	if timeout == 0 {
		// Block indefinitely
		for {
			time.Sleep(100 * time.Millisecond)
			if result := checkAndPop(); result != nil {
				return *result
			}
		}
	} else {
		// Block with timeout (convert float seconds to duration)
		deadline := time.Now().Add(time.Duration(timeout * float64(time.Second)))
		for time.Now().Before(deadline) {
			time.Sleep(100 * time.Millisecond)
			if result := checkAndPop(); result != nil {
				return *result
			}
		}
	}

	// Timeout reached, return null array
	return Value{Typ: "null_array", Str: ""}
}
