package commands

import (
	"sort"
	"strconv"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// Object pool for result slices to reduce allocations
var (
	zrangeResultPool = sync.Pool{
		New: func() interface{} {
			return make([]shared.Value, 0, 16)
		},
	}
)

// Helper functions for pool management
func getZrangeResult() []shared.Value {
	return zrangeResultPool.Get().([]shared.Value)
}

func putZrangeResult(s []shared.Value) {
	s = s[:0] // Reset length but keep capacity
	zrangeResultPool.Put(s)
}

// getSortedRange efficiently gets a range from a sorted set without full sorting
func getSortedRange(ss *shared.SortedSet, start, stop int) []string {
	if start > stop || start >= ss.Size || stop < 0 {
		return []string{}
	}

	// Clamp indices to valid range
	if start < 0 {
		start = 0
	}
	if stop >= ss.Size {
		stop = ss.Size - 1
	}

	// Create slice of members with scores for efficient sorting
	type memberScore struct {
		member string
		score  float64
	}

	members := make([]memberScore, 0, len(ss.Members))
	for m, s := range ss.Members {
		members = append(members, memberScore{m, s})
	}

	// Use Go's efficient sort instead of bubble sort
	sort.Slice(members, func(i, j int) bool {
		if members[i].score != members[j].score {
			return members[i].score < members[j].score
		}
		return members[i].member < members[j].member
	})

	// Extract only the requested range
	result := make([]string, 0, stop-start+1)
	for i := start; i <= stop; i++ {
		result = append(result, members[i].member)
	}

	return result
}

// zrange handles the ZRANGE command.
// Usage: ZRANGE key start stop [WITHSCORES]
// Returns: Array of elements in the specified range.
//
// This command returns the specified elements of the sorted set stored at key.
// The offsets start and stop are zero-based indexes, with 0 being the first element.
// Negative offsets can be used to start from the end of the sorted set.
//
// Special cases:
//   - If start > stop, returns empty array
//   - If start < 0, treated as 0
//   - If stop >= sorted set length, treated as sorted set length - 1
//   - Both start and stop are inclusive
//
// Examples:
//
//	ZRANGE myzset 0 2      // Returns elements at index 0, 1, and 2
//	ZRANGE myzset 0 -1     // Returns all elements
//	ZRANGE myzset -3 -1    // Returns last 3 elements
func Zrange(connID string, args []protocol.Value) protocol.Value {
	if len(args) < 3 {
		return createErrorResponse("ERR wrong number of arguments for 'zrange' command")
	}

	key := args[0].Bulk
	entry, exists := server.Memory[key]

	if !exists {
		return shared.Value{Typ: "array", Array: []shared.Value{}}
	}

	if entry.SortedSet == nil {
		return shared.Value{Typ: "array", Array: []shared.Value{}}
	}

	start, err := strconv.Atoi(args[1].Bulk)
	if err != nil {
		return createErrorResponse("ERR value is not an integer or out of range")
	}
	stop, err := strconv.Atoi(args[2].Bulk)
	if err != nil {
		return createErrorResponse("ERR value is not an integer or out of range")
	}

	// Handle negative indices
	if start < 0 {
		start = entry.SortedSet.Size + start
	}
	if stop < 0 {
		stop = entry.SortedSet.Size + stop
	}

	// If start > stop, return empty array
	if start > stop {
		return shared.Value{Typ: "array", Array: []shared.Value{}}
	}

	// Get range efficiently using optimized method
	rangeMembers := getSortedRange(entry.SortedSet, start, stop)

	// Use object pool for result slice
	result := getZrangeResult()
	defer putZrangeResult(result)

	// Pre-allocate with exact capacity
	result = result[:0]
	if cap(result) < len(rangeMembers) {
		result = make([]shared.Value, 0, len(rangeMembers))
	}

	// Convert to shared.Value array efficiently
	for _, member := range rangeMembers {
		result = append(result, shared.Value{Typ: "string", Str: member})
	}

	return shared.Value{Typ: "array", Array: result}
}
