package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// parseBlockParameter parses the BLOCK parameter and returns timeout and streams index.
func parseBlockParameter(args []shared.Value) (int, int) {
	if len(args) >= 2 && strings.ToUpper(args[0].Bulk) == "BLOCK" {
		if len(args) < 4 {
			return 0, -1 // Error case
		}

		timeoutStr := args[1].Bulk
		if timeoutStr == "0" {
			return -1, 2 // Block indefinitely
		}

		if parsedTimeout, err := strconv.Atoi(timeoutStr); err == nil {
			return parsedTimeout, 2
		}
		return 1000, 2 // Default timeout
	}
	return 0, 0 // No blocking
}

// parseStreamArguments validates streams keyword and parses stream arguments.
func parseStreamArguments(args []shared.Value, streamsIndex int) ([]shared.Value, int, error) {
	if args[streamsIndex].Bulk != "streams" {
		return nil, 0, fmt.Errorf("ERR wrong number of arguments for 'xread' command")
	}

	remainingArgs := args[streamsIndex+1:]
	if len(remainingArgs)%2 != 0 || len(remainingArgs) == 0 {
		return nil, 0, fmt.Errorf("ERR wrong number of arguments for 'xread' command")
	}

	return remainingArgs, len(remainingArgs) / 2, nil
}

// convertDollarToLastID converts $ to the actual last entry ID for each stream.
func convertDollarToLastID(remainingArgs []shared.Value, keyCount int) []shared.Value {
	processedArgs := make([]shared.Value, len(remainingArgs))
	copy(processedArgs, remainingArgs)

	for i := 0; i < keyCount; i++ {
		key := remainingArgs[i].Bulk
		startID := remainingArgs[i+keyCount].Bulk

		if startID == "$" {
			if entry, exists := shared.Memory[key]; exists && len(entry.Stream) > 0 {
				processedArgs[i+keyCount] = shared.Value{Typ: "bulk", Bulk: entry.Stream[len(entry.Stream)-1].ID}
			} else {
				processedArgs[i+keyCount] = shared.Value{Typ: "bulk", Bulk: "0-0"}
			}
		}
	}
	return processedArgs
}

// getStreamEntriesAfter returns stream entries newer than the given ID.
//
// Examples:
//
//	getStreamEntriesAfter("mystream", "0-0")     // Returns all entries newer than 0-0
//	getStreamEntriesAfter("mystream", "1526985054069-0")  // Returns entries newer than specific ID
//	getStreamEntriesAfter("nonexistent", "0-0") // Returns nil (stream doesn't exist)
func getStreamEntriesAfter(key, startID string) []shared.Value {
	entry, exists := shared.Memory[key]
	if !exists {
		return nil
	}

	// Pre-allocate slice with estimated capacity
	result := make([]shared.Value, 0, len(entry.Stream))
	for _, streamEntry := range entry.Stream {
		comparison := compareStreamIDs(streamEntry.ID, startID)
		if comparison > 0 {
			result = append(result, createStreamEntryValue(streamEntry))
		}
	}
	return result
}

// createStreamResponse creates a stream response array.
func createStreamResponse(key string, entries []shared.Value) shared.Value {
	return shared.Value{
		Typ: "array",
		Array: []shared.Value{
			{Typ: "bulk", Bulk: key},
			{Typ: "array", Array: entries},
		},
	}
}

// checkForNewEntries checks for new stream entries across multiple streams.
func checkForNewEntries(remainingArgs []shared.Value, keyCount int) []shared.Value {
	var result []shared.Value
	for i := 0; i < keyCount; i++ {
		key := remainingArgs[i].Bulk
		startID := remainingArgs[i+keyCount].Bulk

		streamEntries := getStreamEntriesAfter(key, startID)
		if len(streamEntries) > 0 {
			result = append(result, createStreamResponse(key, streamEntries))
		}
	}
	return result
}

// blockForNewEntries blocks until new entries are available or timeout occurs.
func blockForNewEntries(processedArgs []shared.Value, keyCount int, blockTimeout int) shared.Value {
	checkInterval := 10 * time.Millisecond

	if blockTimeout == -1 {
		// Block indefinitely
		for {
			time.Sleep(checkInterval)
			if result := checkForNewEntries(processedArgs, keyCount); len(result) > 0 {
				return shared.Value{Typ: "array", Array: result}
			}
		}
	}

	// Block with timeout
	totalWaitTime := time.Duration(blockTimeout) * time.Millisecond
	for elapsed := time.Duration(0); elapsed < totalWaitTime; elapsed += checkInterval {
		time.Sleep(checkInterval)
		if result := checkForNewEntries(processedArgs, keyCount); len(result) > 0 {
			return shared.Value{Typ: "array", Array: result}
		}
	}

	return shared.Value{Typ: "null_array"}
}

// xread handles the XREAD command for reading from multiple streams.
//
// Examples:
//
//	XREAD streams mystream 0-0                    // Single stream
//	XREAD streams stream1 stream2 0-0 0-1        // Multiple streams
//	XREAD BLOCK 1000 streams mystream 0-0       // Blocking with 1000ms timeout
//	XREAD BLOCK 0 streams mystream $      		// Blocking until new entries are available
func Xread(connID string, args []shared.Value) shared.Value {
	if len(args) < 3 {
		return createErrorResponse("ERR wrong number of arguments for 'xread' command")
	}

	// Parse BLOCK parameter if present
	blockTimeout, streamsIndex := parseBlockParameter(args)
	if streamsIndex == -1 {
		return createErrorResponse("ERR wrong number of arguments for 'xread' command")
	}

	// Validate streams keyword and parse arguments
	remainingArgs, keyCount, err := parseStreamArguments(args, streamsIndex)
	if err != nil {
		return createErrorResponse(err.Error())
	}

	// Convert $ to actual last entry IDs
	processedArgs := convertDollarToLastID(remainingArgs, keyCount)

	// Check for immediate results
	if result := checkForNewEntries(processedArgs, keyCount); len(result) > 0 {
		return shared.Value{Typ: "array", Array: result}
	}

	// Handle blocking or return empty array
	if blockTimeout == 0 {
		return shared.Value{Typ: "array", Array: []shared.Value{}}
	}
	return blockForNewEntries(processedArgs, keyCount, blockTimeout)
}
