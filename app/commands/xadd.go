package commands

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// StreamID represents a parsed Redis stream ID with timestamp and sequence
type StreamID struct {
	Timestamp int64
	Sequence  int64
}

// Object pools for reducing allocations
var (
	streamIDPool = sync.Pool{
		New: func() interface{} {
			return &StreamID{}
		},
	}
)

// Helper functions for pool management
func getStreamID() *StreamID {
	return streamIDPool.Get().(*StreamID)
}

func putStreamID(id *StreamID) {
	id.Timestamp = 0
	id.Sequence = 0
	streamIDPool.Put(id)
}

// parseStreamComponent parses a stream ID component, handling "*" for auto-generation.
func parseStreamComponent(component string, defaultValue int64) (int64, error) {
	if component == "*" {
		return defaultValue, nil
	}
	return strconv.ParseInt(component, 10, 64)
}

// generateActualIDOptimized is an optimized version of generateActualID
func generateActualIDOptimized(id string, stream []string) string {
	if !strings.Contains(id, "*") {
		return id
	}

	if id == "*" {
		timestamp := time.Now().UnixMilli()
		sequence := generateSequenceForTimestampOptimized(timestamp, stream)
		return fmt.Sprintf("%d-%d", timestamp, sequence)
	}

	parts := strings.Split(id, "-")
	if len(parts) != 2 {
		return id // Invalid format, return as-is
	}

	timestampStr, sequenceStr := parts[0], parts[1]

	// Handle timestamp
	var timestamp int64
	if timestampStr == "*" {
		timestamp = time.Now().UnixMilli()
	} else {
		timestamp, _ = strconv.ParseInt(timestampStr, 10, 64)
	}

	// Handle sequence
	var sequence int64
	if sequenceStr == "*" {
		sequence = generateSequenceForTimestampOptimized(timestamp, stream)
	} else {
		sequence, _ = strconv.ParseInt(sequenceStr, 10, 64)
	}

	return fmt.Sprintf("%d-%d", timestamp, sequence)
}

// generateSequenceForTimestampOptimized is an optimized version
func generateSequenceForTimestampOptimized(timestamp int64, stream []string) int64 {
	if len(stream) == 0 {
		if timestamp == 0 {
			return 1
		}
		return 0
	}

	// Optimize: Get last entry more efficiently
	lastEntry := stream[len(stream)-1]
	lastParts := strings.Split(lastEntry, "-")
	if len(lastParts) != 2 {
		return 0
	}

	lastTimestamp, _ := strconv.ParseInt(lastParts[0], 10, 64)
	lastSequence, _ := strconv.ParseInt(lastParts[1], 10, 64)

	if timestamp == lastTimestamp {
		return lastSequence + 1
	}
	return 0
}

// validateStreamKeyOptimized is an optimized version
func validateStreamKeyOptimized(id string, stream []string) (bool, error) {
	// Parse the new ID efficiently
	newID := getStreamID()
	defer putStreamID(newID)

	var err error
	if id == "*" {
		newID.Timestamp = time.Now().UnixMilli()
		newID.Sequence = 0
	} else {
		parts := strings.Split(id, "-")
		if len(parts) != 2 {
			return false, fmt.Errorf("ERR Invalid stream ID format")
		}

		newID.Timestamp, err = parseStreamComponent(parts[0], time.Now().UnixMilli())
		if err != nil {
			return false, fmt.Errorf("ERR Invalid timestamp in stream ID")
		}

		newID.Sequence, err = parseStreamComponent(parts[1], 0)
		if err != nil {
			return false, fmt.Errorf("ERR Invalid sequence in stream ID")
		}
	}

	// Check if ID is greater than 0-0
	if newID.Timestamp < 0 || newID.Sequence < 0 {
		return false, fmt.Errorf("ERR The ID specified in XADD must be greater than 0-0")
	}

	// Reject exactly "0-0" only if it was explicitly provided
	if newID.Timestamp == 0 && newID.Sequence == 0 {
		if strings.Contains(id, "*") {
			return true, nil
		}
		return false, fmt.Errorf("ERR The ID specified in XADD must be greater than 0-0")
	}

	// If stream is empty, any valid ID is acceptable
	if len(stream) == 0 {
		return true, nil
	}

	// Get the last entry ID from the stream efficiently
	lastEntry := stream[len(stream)-1]
	parts := strings.Split(lastEntry, "-")
	if len(parts) != 2 {
		return false, fmt.Errorf("ERR Invalid stream data format")
	}

	lastTimestamp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return false, fmt.Errorf("ERR Invalid timestamp in existing stream entry")
	}

	lastSequence, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false, fmt.Errorf("ERR Invalid sequence in existing stream entry")
	}

	// Check if new ID is greater than the last entry
	if newID.Timestamp < lastTimestamp ||
		(newID.Timestamp == lastTimestamp && newID.Sequence <= lastSequence) {
		return false, fmt.Errorf("ERR The ID specified in XADD is equal or smaller than the target stream top item")
	}

	return true, nil
}

// xadd handles the XADD command.
//
// Examples:
//
//	XADD mystream 1-0 message "Hello"           // Explicit ID
//	XADD mystream * message "Hello"             // Auto-generate both timestamp and sequence
//	XADD mystream 0-* message "Hello"           // Auto-generate sequence only
//	XADD mystream *-0 message "Hello"           // Auto-generate timestamp only
func Xadd(connID string, args []shared.Value) shared.Value {
	if len(args) < 3 {
		return createErrorResponse("ERR wrong number of arguments for 'xadd' command")
	}

	key := args[0].Bulk
	id := args[1].Bulk
	entry, exists := server.Memory[key]

	// Parse field-value pairs efficiently
	streamData := make(map[string]string, (len(args)-2)/2)
	for i := 2; i < len(args); i += 2 {
		if i+1 < len(args) {
			streamData[args[i].Bulk] = args[i+1].Bulk
		}
	}

	var actualID string
	var streamIDs []string

	if !exists {
		streamIDs = []string{}
	} else {
		// Optimize: Pre-allocate slice with known capacity
		streamIDs = make([]string, 0, len(entry.Stream))
		for _, streamEntry := range entry.Stream {
			streamIDs = append(streamIDs, streamEntry.ID)
		}
	}

	actualID = generateActualIDOptimized(id, streamIDs)
	valid, err := validateStreamKeyOptimized(actualID, streamIDs)
	if !valid {
		return createErrorResponse(err.Error())
	}

	// Create stream entry efficiently
	streamEntry := shared.StreamEntry{ID: actualID, Data: streamData}

	if !exists {
		// Optimize: Pre-allocate with capacity
		entry = shared.MemoryEntry{Stream: make([]shared.StreamEntry, 0, 1)}
	}
	entry.Stream = append(entry.Stream, streamEntry)
	server.Memory[key] = entry

	return shared.Value{Typ: "bulk", Bulk: actualID}
}
