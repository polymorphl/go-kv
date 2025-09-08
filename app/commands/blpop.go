package commands

import (
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

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
func Blpop(connID string, args []shared.Value) shared.Value {
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
	checkAndPop := func() *shared.Value {
		for i := 0; i < len(args)-1; i++ {
			key := args[i].Bulk
			entry, exists := shared.Memory[key]

			if exists && len(entry.Array) > 0 {
				// Found a non-empty list, pop the first element
				value := entry.Array[0]
				entry.Array = entry.Array[1:]
				shared.Memory[key] = entry

				// Return [key, value] array
				return &shared.Value{Typ: "array", Array: []shared.Value{
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
	return shared.Value{Typ: "null_array", Str: ""}
}
