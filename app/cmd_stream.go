package main

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
		entry = MemoryEntry{Stream: []string{id}, Value: value}
		memory[key] = entry
	} else {
		entry.Stream = append(entry.Stream, args[1].Bulk)
		memory[key] = entry
	}

	return Value{Typ: "string", Str: id}
}
