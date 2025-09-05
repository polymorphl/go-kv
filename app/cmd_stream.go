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

// parseStreamID parses a Redis stream ID string into timestamp and sequence components.
// It handles both explicit IDs and wildcard patterns:
// - "*": Returns current timestamp with sequence 0 (for auto-generation)
// - "timestamp-sequence": Parses explicit values
// - "timestamp-*" or "*-sequence": Handles partial wildcards
//
// Parameters:
//   - id: The stream ID string to parse
//
// Returns:
//   - *StreamID: Parsed timestamp and sequence components
//   - error: Error if ID format is invalid
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

// parseStreamComponent parses a stream ID component (timestamp or sequence), handling "*" for auto-generation.
// It's a utility function that handles the common pattern of parsing stream ID components.
//
// Parameters:
//   - component: The component string to parse (e.g., "1234" or "*")
//   - defaultValue: The value to return if component is "*"
//
// Returns:
//   - int64: The parsed value or default value
//   - error: Error if parsing fails
func parseStreamComponent(component string, defaultValue int64) (int64, error) {
	if component == "*" {
		return defaultValue, nil
	}
	return strconv.ParseInt(component, 10, 64)
}

// getLastStreamID extracts and parses the last ID from a stream.
// It's used for validation to ensure new IDs are greater than existing ones.
//
// Parameters:
//   - stream: The stream entries (slice of ID strings)
//
// Returns:
//   - *StreamID: The parsed last stream ID, or nil if stream is empty
//   - error: Error if the last entry has invalid format
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

// generateSequenceForTimestamp generates the appropriate sequence number for a given timestamp.
// It handles the Redis stream sequence generation rules:
// - Empty stream with timestamp 0: returns 1 (avoids generating 0-0)
// - Empty stream with other timestamps: returns 0
// - Existing stream with same timestamp: increments from last sequence
// - Existing stream with different timestamp: returns 0
//
// Parameters:
//   - timestamp: The timestamp for which to generate a sequence
//   - stream: The existing stream entries (slice of ID strings)
//
// Returns: The generated sequence number
func generateSequenceForTimestamp(timestamp int64, stream []string) int64 {
	if len(stream) == 0 {
		// For empty stream, start at 1 if timestamp is 0, otherwise 0
		if timestamp == 0 {
			return 1
		}
		return 0
	}

	// Get the last entry's sequence and increment
	lastEntry := stream[len(stream)-1]
	lastParts := strings.Split(lastEntry, "-")
	if len(lastParts) != 2 {
		return 0
	}

	lastTimestamp, _ := strconv.ParseInt(lastParts[0], 10, 64)
	lastSequence, _ := strconv.ParseInt(lastParts[1], 10, 64)

	if timestamp == lastTimestamp {
		return lastSequence + 1 // Same timestamp, increment sequence
	}
	return 0 // Different timestamp, start sequence at 0
}

// generateActualID generates the actual Redis stream ID by replacing wildcards (*) with appropriate values.
// It handles various ID formats:
// - "*": Auto-generates both timestamp (current time) and sequence
// - "timestamp-*": Uses provided timestamp, auto-generates sequence
// - "*-sequence": Auto-generates timestamp, uses provided sequence
// - "timestamp-sequence": Returns as-is (no auto-generation)
//
// Auto-generation rules:
// - Timestamp: Uses current Unix timestamp in milliseconds
// - Sequence: Uses generateSequenceForTimestamp() logic
//
// Parameters:
//   - id: The input ID string (may contain wildcards)
//   - stream: The existing stream entries for sequence calculation
//
// Returns: The generated actual ID in "timestamp-sequence" format
func generateActualID(id string, stream []string) string {
	if !strings.Contains(id, "*") {
		return id
	}

	if id == "*" {
		timestamp := time.Now().UnixMilli()
		sequence := generateSequenceForTimestamp(timestamp, stream)
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
		sequence = generateSequenceForTimestamp(timestamp, stream)
	} else {
		sequence, _ = strconv.ParseInt(sequenceStr, 10, 64)
	}

	return fmt.Sprintf("%d-%d", timestamp, sequence)
}

// validateStreamKey validates a Redis stream ID against existing stream entries.
// It enforces Redis XADD validation rules:
// - ID must be greater than "0-0" (unless auto-generated)
// - ID must be greater than the last entry in the stream
// - Auto-generated IDs (containing "*") are allowed to be "0-0"
//
// Validation checks:
// 1. Parse the ID into timestamp and sequence components
// 2. Ensure timestamp and sequence are non-negative
// 3. Reject explicit "0-0" but allow auto-generated "0-0"
// 4. For existing streams, ensure new ID is greater than last entry
//
// Parameters:
//   - id: The stream ID to validate (may contain wildcards)
//   - stream: The existing stream entries for comparison
//
// Returns:
//   - bool: true if valid, false if invalid
//   - error: Error message if validation fails
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

	// Reject exactly "0-0" only if it was explicitly provided
	// Allow "0-0" if it was auto-generated (e.g., from "*" or "0-*")
	if newID.Timestamp == 0 && newID.Sequence == 0 {
		// Check if this was auto-generated by looking for "*" in the original ID
		if strings.Contains(id, "*") {
			// This was auto-generated, so "0-0" is valid
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

	return Value{Typ: "bulk", Bulk: actualID}
}
