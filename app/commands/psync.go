package commands

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// psync handles the PSYNC command.
// Usage: PSYNC masterReplID masterReplOffset
// Returns: "FULLRESYNC masterReplID masterReplOffset" followed by RDB file
// This is typically used to synchronize a replica with a master.
func Psync(connID string, args []shared.Value) shared.Value {
	if len(args) < 2 {
		return createErrorResponse("ERR wrong number of arguments for 'psync' command")
	}

	fullResyncResponse := fmt.Sprintf("FULLRESYNC %s %d", shared.StoreState.MasterReplID, shared.StoreState.MasterReplOffset)

	// Find the connection to send the RDB file
	if conn, exists := shared.ConnectionsGet(connID); exists {
		// Register this replica connection for command propagation
		shared.ReplicasSet(connID, conn)

		// Send the FULLRESYNC response first
		response := shared.Value{Typ: "string", Str: fullResyncResponse}
		conn.Write(response.Marshal())

		// Send empty RDB file
		rdbData, err := GetRDBData()
		if err != nil {
			return createErrorResponse("Failed to get RDB data")
		}
		rdbHeader := fmt.Sprintf("$%d\r\n", len(rdbData))
		conn.Write([]byte(rdbHeader))
		conn.Write(rdbData)

		// Return NO_RESPONSE to indicate we've already sent the response directly
		return shared.Value{Typ: shared.NO_RESPONSE, Str: ""}
	}

	// For other PSYNC requests, return FULLRESYNC with master info
	return shared.Value{Typ: "string", Str: fullResyncResponse}
}

// GetRDBData returns the RDB file data, loading from disk if available
func GetRDBData() ([]byte, error) {
	// Try to load the actual RDB file first
	filePath := filepath.Join(shared.StoreState.ConfigDir, shared.StoreState.ConfigDbfilename)

	if data, err := os.ReadFile(filePath); err == nil {
		// Successfully loaded RDB file, parse it into memory
		if parseErr := shared.ParseRDBData(data); parseErr != nil {
			fmt.Printf("Warning: Failed to parse RDB file %s: %v\n", filePath, parseErr)
		}
		return data, nil
	}

	// Fall back to empty RDB file if no file exists or can't be read
	emptyRDBFileHex := "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"
	emptyRDB, err := hex.DecodeString(emptyRDBFileHex)
	return emptyRDB, err
}
