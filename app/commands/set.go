package commands

import (
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

// set handles the SET command.
// Usage: SET key value [PX milliseconds]
// Returns: "OK" on success, error message on failure.
//
// This command sets a key to hold a string value. If the key already exists,
// it is overwritten. The PX option sets an expiration time in milliseconds.
//
// Examples:
//
//	SET mykey "Hello"           // Sets key without expiration
//	SET mykey "Hello" PX 1000   // Sets key with 1 second expiration
func Set(connID string, args []shared.Value) shared.Value {
	if len(args) < 2 {
		return createErrorResponse("ERR wrong number of arguments for 'set' command")
	}

	key := args[0].Bulk
	value := args[1].Bulk
	entry := shared.MemoryEntry{Value: value, Expires: 0}

	// Parse optional PX (expiration) argument
	for i := 2; i < len(args); i++ {
		if strings.ToUpper(args[i].Bulk) == "PX" && i+1 < len(args) {
			ms, err := strconv.ParseInt(args[i+1].Bulk, 10, 64)
			if err != nil {
				return createErrorResponse("ERR value is not an integer or out of range")
			}
			entry.Expires = time.Now().UnixMilli() + ms
			i++ // Skip the next argument since we've processed it
		}
	}

	server.Memory[key] = entry
	return shared.Value{Typ: "string", Str: "OK"}
}
