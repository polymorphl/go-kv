package main

// echo handles the ECHO command.
// Usage: ECHO message
// Returns: The message that was sent as an argument.
// This command is useful for testing the connection and verifying that
// the server is receiving and processing commands correctly.
func echo(args []Value) Value {
	return Value{Typ: "string", Str: args[0].Bulk}
}

// ping handles the PING command.
// Usage: PING [message]
// Returns: "PONG" if no message provided, otherwise echoes the provided message.
// This is typically used to test if the server is alive and responsive.
func ping(args []Value) Value {
	if len(args) == 0 {
		return Value{Typ: "string", Str: "PONG"}
	}

	return echo(args)
}

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
		return createErrorResponse("ERR wrong number of arguments for 'type' command")
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
