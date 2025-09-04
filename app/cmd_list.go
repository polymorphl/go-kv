package main

import "strconv"

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
		return Value{Typ: "error", Str: "ERR wrong number of arguments for '" + cmdName + "' command"}
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
		return Value{Typ: "error", Str: "ERR wrong number of arguments for 'lrange' command"}
	}

	key := args[0].Bulk
	entry, exists := memory[key]

	if !exists {
		return Value{Typ: "array", Array: []Value{}}
	}

	// Check if it's an array
	if len(entry.Array) == 0 {
		return Value{Typ: "error", Str: "WRONGTYPE Operation against a key holding the wrong kind of value"}
	}

	// Parse start and stop indices
	start, err := strconv.Atoi(args[1].Bulk)
	if err != nil {
		return Value{Typ: "error", Str: "ERR value is not an integer or out of range"}
	}
	stop, err := strconv.Atoi(args[2].Bulk)
	if err != nil {
		return Value{Typ: "error", Str: "ERR value is not an integer or out of range"}
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
		return Value{Typ: "error", Str: "ERR wrong number of arguments for 'llen' command"}
	}

	key := args[0].Bulk
	entry, exists := memory[key]

	if !exists {
		return Value{Typ: "integer", Num: 0}
	}

	return Value{Typ: "integer", Num: len(entry.Array)}
}

func lpop(args []Value) Value {
	if len(args) != 1 {
		return Value{Typ: "error", Str: "ERR wrong number of arguments for 'lpop' command"}
	}

	key := args[0].Bulk
	entry, exists := memory[key]

	if !exists || len(entry.Array) == 0 {
		return Value{Typ: "null", Str: ""}
	}

	tmp := entry.Array[0]
	entry.Array = entry.Array[1:]
	memory[key] = entry

	return Value{Typ: "string", Str: tmp}
}
