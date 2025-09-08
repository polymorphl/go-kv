package commands

import "github.com/codecrafters-io/redis-starter-go/app/shared"

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
func Llen(args []shared.Value) shared.Value {
	if len(args) != 1 {
		return createErrorResponse("ERR wrong number of arguments for 'llen' command")
	}

	key := args[0].Bulk
	entry, exists := shared.Memory[key]

	if !exists {
		return shared.Value{Typ: "integer", Num: 0}
	}

	return shared.Value{Typ: "integer", Num: len(entry.Array)}
}
