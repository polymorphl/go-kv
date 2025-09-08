package commands

import "github.com/codecrafters-io/redis-starter-go/app/shared"

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
func Rpush(connID string, args []shared.Value) shared.Value {
	return push(args, false)
}
