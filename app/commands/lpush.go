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
func Lpush(args []shared.Value) shared.Value {
	return push(args, true)
}
