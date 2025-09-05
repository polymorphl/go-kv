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

// generateActualID generates the actual ID, replacing * with appropriate values
func generateActualID(id string, stream []string) string {
	if !strings.Contains(id, "*") {
		return id // No auto-generation needed
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
		if len(stream) == 0 {
			sequence = 1 // First entry gets sequence 1
		} else {
			// Get the last entry's sequence and increment
			lastEntry := stream[len(stream)-1]
			lastParts := strings.Split(lastEntry, "-")
			if len(lastParts) == 2 {
				lastTimestamp, _ := strconv.ParseInt(lastParts[0], 10, 64)
				lastSequence, _ := strconv.ParseInt(lastParts[1], 10, 64)

				if timestamp == lastTimestamp {
					sequence = lastSequence + 1 // Same timestamp, increment sequence
				} else {
					sequence = 0 // Different timestamp, start sequence at 0
				}
			} else {
				sequence = 1
			}
		}
	} else {
		sequence, _ = strconv.ParseInt(sequenceStr, 10, 64)
	}

	return fmt.Sprintf("%d-%d", timestamp, sequence)
}

func validateStreamKey(id string, stream []string) (bool, error) {
	// Parse the new ID
	newID, err := parseStreamID(id)
	if err != nil {
		return false, err
	}

	// Check if ID is greater than 0-0
	if newID.Timestamp < 0 || newID.Sequence < 0 {
		return false, fmt.Errorf("ERR The ID specified in XADD must be greater than 0-0")
	}

	// Reject exactly "0-0" but allow "0-*" (auto-generated sequence)
	// For "0-*", sequence will be auto-generated to be > 0, so it's valid
	if newID.Timestamp == 0 && newID.Sequence == 0 {
		// Check if this is "0-*" by looking at the original ID
		if strings.HasSuffix(id, "-*") {
			// This is "0-*", which is valid (sequence will be auto-generated)
			return true, nil
		}
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
// The ID must be unique within the stream and greater than the last entry.
//
// ID Format:
// - Explicit: "timestamp-sequence" (e.g., "1526919030474-0")
// - Auto-generation: "*" for timestamp (current time) or sequence (incremented)
// - Mixed: "timestamp-*" or "*-sequence" (e.g., "0-*", "*-0")
//
// Examples:
//
//	XADD mystream 1-0 message "Hello"           // Explicit ID
//	XADD mystream * message "Hello"             // Auto-generate both timestamp and sequence
//	XADD mystream 0-* message "Hello"           // Auto-generate sequence only
//	XADD mystream *-0 message "Hello"           // Auto-generate timestamp only
//
// Validation Rules:
// - ID must be greater than "0-0"
// - ID must be greater than the last entry in the stream
// - Sequence auto-generation starts at 1 for empty streams
// - Sequence auto-generation starts at 0 for new timestamps
// - Sequence auto-generation increments for same timestamps
//
// Note: This is a simplified implementation that stores stream data as a basic
// array structure with field-value pairs stored in the Value field.
func xadd(args []Value) Value {
	if len(args) < 3 {
		return Value{Typ: "error", Str: "ERR wrong number of arguments for 'xadd' command"}
	}

	key := args[0].Bulk
	id := args[1].Bulk
	entry, exists := memory[key]

	var actualID string

	if !exists {
		value := args[2].Bulk
		valid, err := validateStreamKey(id, []string{})
		if !valid {
			return Value{Typ: "error", Str: err.Error()}
		}
		actualID = generateActualID(id, []string{})
		entry = MemoryEntry{Stream: []string{actualID}, Value: value}
		memory[key] = entry
	} else {
		actualID = generateActualID(id, entry.Stream)
		valid, err := validateStreamKey(actualID, entry.Stream)
		if !valid {
			return Value{Typ: "error", Str: err.Error()}
		}
		entry.Stream = append(entry.Stream, actualID)
		memory[key] = entry
	}

	return Value{Typ: "string", Str: actualID}
}
