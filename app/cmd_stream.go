package main

// typeCmd handles the TYPE command.
// Usage: TYPE key
// Returns: The type of the value stored at key.
//
// This command returns the type of the value stored at key.
// The type can be one of: string, list, set, zset, hash, stream, or none.
// If the key does not exist, "none" is returned.
//
// Examples:
//
//	TYPE mykey                    // Returns the type of mykey
//	TYPE nonexistent              // Returns "none" (key doesn't exist)
//	TYPE mystring                 // Returns "string"
//	TYPE mylist                   // Returns "list"
func typeCmd(args []Value) Value {
	if len(args) != 1 {
		return Value{Typ: "error", Str: "ERR wrong number of arguments for 'type' command"}
	}

	key := args[0].Bulk
	entry, exists := memory[key]

	if !exists {
		return Value{Typ: "string", Str: "none"}
	}

	// Determine the actual type based on the entry structure
	switch {
	case len(entry.Array) > 0:
		return Value{Typ: "string", Str: "list"}
	case len(entry.Stream) > 0:
		return Value{Typ: "string", Str: "stream"}
	case entry.Value != "":
		return Value{Typ: "string", Str: "string"}
	default:
		return Value{Typ: "string", Str: "none"}
	}
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
		entry = MemoryEntry{Stream: []string{id}, Value: value}
		memory[key] = entry
	} else {
		entry.Stream = append(entry.Stream, args[1].Bulk)
		memory[key] = entry
	}

	return Value{Typ: "string", Str: id}
}
