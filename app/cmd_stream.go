package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// StreamID represents a parsed Redis stream ID with timestamp and sequence
type StreamID struct {
	Timestamp int64
	Sequence  int64
}

// parseStreamID parses a Redis stream ID string into timestamp and sequence components
func parseStreamID(id string) (*StreamID, error) {
	if id == "*" {
		return &StreamID{
			Timestamp: time.Now().UnixMilli(),
			Sequence:  0,
		}, nil
	}

	parts := strings.Split(id, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("ERR Invalid stream ID format")
	}

	timestampStr, sequenceStr := parts[0], parts[1]

	timestamp, err := parseStreamComponent(timestampStr, time.Now().UnixMilli())
	if err != nil {
		return nil, fmt.Errorf("ERR Invalid timestamp in stream ID")
	}

	sequence, err := parseStreamComponent(sequenceStr, 0)
	if err != nil {
		return nil, fmt.Errorf("ERR Invalid sequence in stream ID")
	}

	return &StreamID{Timestamp: timestamp, Sequence: sequence}, nil
}

// parseStreamComponent parses a stream ID component (timestamp or sequence), handling "*" for auto-generation
func parseStreamComponent(component string, defaultValue int64) (int64, error) {
	if component == "*" {
		return defaultValue, nil
	}
	return strconv.ParseInt(component, 10, 64)
}

// getLastStreamID extracts and parses the last ID from a stream
func getLastStreamID(stream []string) (*StreamID, error) {
	if len(stream) == 0 {
		return nil, nil
	}

	lastEntry := stream[len(stream)-1]
	parts := strings.Split(lastEntry, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("ERR Invalid stream data format")
	}

	timestamp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("ERR Invalid timestamp in existing stream entry")
	}

	sequence, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("ERR Invalid sequence in existing stream entry")
	}

	return &StreamID{Timestamp: timestamp, Sequence: sequence}, nil
}

// validateStreamKey validates a Redis stream ID against existing stream entries
// Should handle: ERR The ID specified in XADD is equal or smaller than the target stream top item
// Should handle: ERR The ID specified in XADD must be greater than 0-0
func validateStreamKey(id string, stream []string) (bool, error) {
	// Parse the new ID
	newID, err := parseStreamID(id)
	if err != nil {
		return false, err
	}

	// Check if ID is greater than 0-0
	if newID.Timestamp < 0 || newID.Sequence < 0 || (newID.Timestamp == 0 && newID.Sequence == 0) {
		return false, fmt.Errorf("ERR The ID specified in XADD must be greater than 0-0")
	}

	// If stream is empty, any valid ID is acceptable
	if len(stream) == 0 {
		return true, nil
	}

	// Get the last entry ID from the stream
	lastID, err := getLastStreamID(stream)
	if err != nil {
		return false, err
	}

	// Check if new ID is greater than the last entry
	if newID.Timestamp < lastID.Timestamp ||
		(newID.Timestamp == lastID.Timestamp && newID.Sequence <= lastID.Sequence) {
		return false, fmt.Errorf("ERR The ID specified in XADD is equal or smaller than the target stream top item")
	}

	return true, nil
}

// xadd handles the XADD command.
// Usage: XADD key ID field value [field value ...]
// Returns: The ID of the newly added entry.
//
// This command adds an entry to a stream stored at key. If the key does not exist,
// a new stream is created. Each entry consists of an ID and one or more field-value pairs.
// The ID must be unique within the stream.
//
// Examples:
//
//	XADD newstream 1-0 message "Hello"                // Creates new stream
//
// Note: This is a simplified implementation that stores stream data as a basic
// array structure.
func xadd(args []Value) Value {
	if len(args) < 3 {
		return Value{Typ: "error", Str: "ERR wrong number of arguments for 'xadd' command"}
	}

	key := args[0].Bulk
	id := args[1].Bulk
	entry, exists := memory[key]

	if !exists {
		value := args[2].Bulk
		valid, err := validateStreamKey(id, []string{})
		if !valid {
			return Value{Typ: "error", Str: err.Error()}
		}
		entry = MemoryEntry{Stream: []string{id}, Value: value}
		memory[key] = entry
	} else {
		valid, err := validateStreamKey(id, entry.Stream)
		if !valid {
			return Value{Typ: "error", Str: err.Error()}
		}
		entry.Stream = append(entry.Stream, id)
		memory[key] = entry
	}

	return Value{Typ: "string", Str: id}
}
