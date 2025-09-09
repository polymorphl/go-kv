package commands

import "github.com/codecrafters-io/redis-starter-go/app/shared"

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
func Lpush(connID string, args []shared.Value) shared.Value {
	if len(args) < 2 {
		return createErrorResponse("ERR wrong number of arguments for 'lpush' command")
	}

	key := args[0].Bulk
	entry, exists := shared.Memory[key]

	// If key doesn't exist, create a new linked list
	if !exists {
		entry = shared.MemoryEntry{List: shared.NewLinkedList(), Expires: 0}
	} else if entry.List == nil {
		// If we have an array but no list, convert array to linked list
		if len(entry.Array) > 0 {
			entry.List = shared.FromArray(entry.Array)
			entry.Array = nil // Clear the array to save memory
		} else {
			// Create new linked list for empty array
			entry.List = shared.NewLinkedList()
		}
		// Clear the string value when converting to list
		entry.Value = ""
	}

	// LPUSH: O(1) insertion at head using linked list
	newCount := len(args) - 1
	if newCount == 0 {
		return shared.Value{Typ: "integer", Num: entry.List.Size}
	}

	// Add new values in reverse order (Redis LPUSH behavior)
	// We need to add them in forward order to get reverse result
	for i := 1; i < len(args); i++ {
		entry.List.AddToHead(args[i].Bulk)
	}

	shared.Memory[key] = entry
	return shared.Value{Typ: "integer", Num: entry.List.Size}
}
