package commands

import "github.com/codecrafters-io/redis-starter-go/app/shared"

// info handles the INFO command.
// Usage: INFO
// This command is used to get information about the server.
func Info(connID string, args []shared.Value) shared.Value {
	state := shared.StoreState

	// Build the info response as a bulk string with key-value pairs
	info := "role:" + state.Role + "\r\n"

	// If a section is specified, we can filter the response
	// For now, we'll return the same info regardless of section
	if len(args) > 0 {
		section := args[0].Bulk
		_ = section // Use section if needed for filtering
	}

	return shared.Value{Typ: "bulk", Bulk: info}
}
