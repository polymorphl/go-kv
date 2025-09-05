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
		valid, err := validateStreamKey(id, []string{})
		if !valid {
			return Value{Typ: "error", Str: err.Error()}
		}
		actualID = generateActualID(id, []string{})

		// Parse field-value pairs from args[2:]
		streamData := make(map[string]string)
		for i := 2; i < len(args); i += 2 {
			if i+1 < len(args) {
				field := args[i].Bulk
				value := args[i+1].Bulk
				streamData[field] = value
			}
		}

		streamEntry := StreamEntry{ID: actualID, Data: streamData}
		entry = MemoryEntry{Stream: []StreamEntry{streamEntry}}
		memory[key] = entry
	} else {
		// Convert existing stream to string slice for validation
		streamIDs := make([]string, len(entry.Stream))
		for i, streamEntry := range entry.Stream {
			streamIDs[i] = streamEntry.ID
		}

		actualID = generateActualID(id, streamIDs)
		valid, err := validateStreamKey(actualID, streamIDs)
		if !valid {
			return Value{Typ: "error", Str: err.Error()}
		}

		// Parse field-value pairs from args[2:]
		streamData := make(map[string]string)
		for i := 2; i < len(args); i += 2 {
			if i+1 < len(args) {
				field := args[i].Bulk
				value := args[i+1].Bulk
				streamData[field] = value
			}
		}

		streamEntry := StreamEntry{ID: actualID, Data: streamData}
		entry.Stream = append(entry.Stream, streamEntry)
		memory[key] = entry
	}

	return Value{Typ: "bulk", Bulk: actualID}
}

// xrange handles the XRANGE command.
// Usage: XRANGE key start end [COUNT count]
// Returns: Array of stream entries within the specified ID range.
//
// This command returns entries from a stream within a specified ID range.
// Each entry is returned as an array containing [ID, [field1, value1, field2, value2, ...]]
//
// ID Range Format:
// - start/end: Can be stream IDs (e.g., "1526985054069-0") or "-"/"+" for beginning/end
// - "-": Start from the beginning of the stream
// - "+": End at the end of the stream
//
// Examples:
//
//	XRANGE mystream 1526985054069 1526985054079    // Range between specific IDs
//	XRANGE mystream - +                            // All entries
//	XRANGE mystream 1526985054069 +                // From specific ID to end
//
// Returns: Array of entries, where each entry is [ID, [field-value pairs]]
func xrange(args []Value) Value {
	if len(args) < 3 {
		return Value{Typ: "error", Str: "ERR wrong number of arguments for 'xrange' command"}
	}

	key := args[0].Bulk
	start := args[1].Bulk
	end := args[2].Bulk

	entry, exists := memory[key]
	if !exists {
		// Empty stream - return empty array
		return Value{Typ: "array", Array: []Value{}}
	}

	var result []Value

	for _, streamEntry := range entry.Stream {
		if isInRange(streamEntry.ID, start, end) {
			entryValue := createStreamEntryValue(streamEntry)
			result = append(result, entryValue)
		}
	}

	return Value{Typ: "array", Array: result}
}

// createStreamEntryValue creates a RESP array value representing a stream entry.
// It formats the stream entry as [ID, [field1, value1, field2, value2, ...]]
// which matches the Redis XRANGE response format.
//
// Parameters:
//   - entry: The stream entry containing ID and field-value pairs
//
// Returns: A RESP array value representing the stream entry
func createStreamEntryValue(entry StreamEntry) Value {
	// Create field-value array
	var fieldValueArray []Value
	for field, value := range entry.Data {
		fieldValueArray = append(fieldValueArray,
			Value{Typ: "bulk", Bulk: field},
			Value{Typ: "bulk", Bulk: value},
		)
	}

	// Create entry array: [ID, [field-value pairs]]
	entryArray := []Value{
		{Typ: "bulk", Bulk: entry.ID},
		{Typ: "array", Array: fieldValueArray},
	}

	return Value{Typ: "array", Array: entryArray}
}

// compareStreamIDs compares two Redis stream IDs lexicographically.
// It implements Redis stream ID comparison rules:
// 1. Compare timestamps first (numerically)
// 2. If timestamps are equal, compare sequences (numerically)
// 3. Fallback to string comparison for invalid formats
//
// Parameters:
//   - id1: First stream ID to compare
//   - id2: Second stream ID to compare
//
// Returns:
//   - -1: if id1 < id2
//   - 0:  if id1 == id2
//   - 1:  if id1 > id2
func compareStreamIDs(id1, id2 string) int {
	// Parse both IDs into timestamp-sequence format
	parts1 := strings.Split(id1, "-")
	parts2 := strings.Split(id2, "-")

	// Validate ID format
	if len(parts1) != 2 || len(parts2) != 2 {
		return compareStrings(id1, id2)
	}

	// Parse timestamps
	timestamp1, err1 := strconv.ParseInt(parts1[0], 10, 64)
	timestamp2, err2 := strconv.ParseInt(parts2[0], 10, 64)

	if err1 != nil || err2 != nil {
		return compareStrings(id1, id2)
	}

	// Compare timestamps first
	if timestamp1 < timestamp2 {
		return -1
	} else if timestamp1 > timestamp2 {
		return 1
	}

	// Timestamps are equal, compare sequences
	sequence1, err1 := strconv.ParseInt(parts1[1], 10, 64)
	sequence2, err2 := strconv.ParseInt(parts2[1], 10, 64)

	if err1 != nil || err2 != nil {
		return compareStrings(id1, id2)
	}

	if sequence1 < sequence2 {
		return -1
	} else if sequence1 > sequence2 {
		return 1
	}

	return 0
}

// compareStrings performs a simple string comparison.
// Helper function for fallback comparison when ID parsing fails.
func compareStrings(s1, s2 string) int {
	if s1 < s2 {
		return -1
	} else if s1 > s2 {
		return 1
	}
	return 0
}

// isInRange checks if a stream ID is within the specified range (inclusive).
// It handles Redis XRANGE range semantics:
// - "-": Start from the beginning of the stream (matches any ID)
// - "+": End at the end of the stream (matches any ID)
// - Specific IDs: Uses lexicographic comparison via compareStreamIDs()
//
// Range Examples:
// - isInRange("0-2", "0-1", "0-3") → true (0-2 is between 0-1 and 0-3)
// - isInRange("0-1", "0-2", "0-3") → false (0-1 is before 0-2)
// - isInRange("0-2", "-", "+") → true (any ID matches "-" to "+")
//
// Parameters:
//   - id: The stream ID to check
//   - start: The start of the range (or "-" for beginning)
//   - end: The end of the range (or "+" for end)
//
// Returns: true if the ID is within the range (inclusive), false otherwise
func isInRange(id, start, end string) bool {
	// Handle special Redis range values
	if start == "-" && end == "+" {
		return true // All entries
	}
	if start == "-" {
		return compareStreamIDs(id, end) <= 0
	}
	if end == "+" {
		return compareStreamIDs(id, start) >= 0
	}

	// Both are specific IDs - check if id is between start and end (inclusive)
	return compareStreamIDs(id, start) >= 0 && compareStreamIDs(id, end) <= 0
}

// xread handles the XREAD command for reading from multiple streams.
// Usage: XREAD streams key1 key2 ... id1 id2 ...
// Returns: Array of streams with entries newer than the specified IDs.
//
// Examples:
//
//	XREAD streams mystream 0-0                    // Single stream
//	XREAD streams stream1 stream2 0-0 0-1        // Multiple streams
func xread(args []Value) Value {
	// Validate basic arguments
	if len(args) < 3 || args[0].Bulk != "streams" {
		return Value{Typ: "error", Str: "ERR wrong number of arguments for 'xread' command"}
	}

	// Parse: ["streams", "key1", "key2", ..., "id1", "id2", ...]
	// Must have equal number of keys and IDs
	totalArgs := len(args) - 1 // Exclude "streams"
	if totalArgs%2 != 0 || totalArgs == 0 {
		return Value{Typ: "error", Str: "ERR wrong number of arguments for 'xread' command"}
	}

	keyCount := totalArgs / 2
	var result []Value

	// Process each stream-key pair
	for i := 0; i < keyCount; i++ {
		key := args[i+1].Bulk
		startID := args[i+keyCount+1].Bulk

		// Get stream entries newer than startID
		streamEntries := getStreamEntriesAfter(key, startID)
		if len(streamEntries) > 0 {
			result = append(result, Value{
				Typ: "array",
				Array: []Value{
					{Typ: "bulk", Bulk: key},
					{Typ: "array", Array: streamEntries},
				},
			})
		}
	}

	return Value{Typ: "array", Array: result}
}

// getStreamEntriesAfter returns entries from a stream that are newer than the given ID.
//
// This function retrieves all stream entries for a given key that have IDs greater than
// the specified startID. It's used by the XREAD command to find entries newer than
// a particular stream position.
//
// Parameters:
//   - key: The stream key to search in
//   - startID: The stream ID to compare against (entries must be newer than this)
//
// Returns:
//   - []Value: Array of stream entries formatted as RESP values, or nil if stream doesn't exist
//   - Each entry is formatted as [ID, [field-value pairs]] using createStreamEntryValue
//
// Behavior:
//   - Returns nil if the stream key doesn't exist
//   - Only includes entries where compareStreamIDs(entry.ID, startID) > 0
//   - Entries are returned in the order they appear in the stream
//   - Empty array is returned if no entries are newer than startID
//
// Examples:
//
//	getStreamEntriesAfter("mystream", "0-0")     // Returns all entries newer than 0-0
//	getStreamEntriesAfter("mystream", "1526985054069-0")  // Returns entries newer than specific ID
//	getStreamEntriesAfter("nonexistent", "0-0") // Returns nil (stream doesn't exist)
//
// Stream ID Comparison:
//   - Uses lexicographical comparison of timestamp-sequence format
//   - "1526985054079-0" > "1526985054069-0" (newer timestamp)
//   - "1526985054069-1" > "1526985054069-0" (same timestamp, higher sequence)
func getStreamEntriesAfter(key, startID string) []Value {
	entry, exists := memory[key]
	if !exists {
		return nil
	}

	var result []Value
	for _, streamEntry := range entry.Stream {
		if compareStreamIDs(streamEntry.ID, startID) > 0 {
			result = append(result, createStreamEntryValue(streamEntry))
		}
	}
	return result
}
