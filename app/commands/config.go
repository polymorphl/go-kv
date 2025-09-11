package commands

import (
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// Config handles the CONFIG command
// Usage: CONFIG GET parameter
// Returns: The value of the configuration parameter, or error if parameter is unknown.
//
// Examples:
//
//	CONFIG GET dir           // Returns the value of the directory where Redis stores its data
//	CONFIG GET dbfilename    // Returns the value of the database file name
//	CONFIG GET unknown       // Returns an error if the parameter is unknown
func Config(connID string, args []shared.Value) shared.Value {
	if len(args) < 1 {
		return createErrorResponse("ERR wrong number of arguments for 'config' command")
	}

	subcommand := strings.ToUpper(args[0].Bulk)

	switch subcommand {
	case "GET":
		return configGet(args[1:])
	default:
		return createErrorResponse("ERR unknown subcommand for 'config' command")
	}
}

// configGet handles the CONFIG GET subcommand
func configGet(args []shared.Value) shared.Value {
	if len(args) == 0 {
		return createErrorResponse("ERR wrong number of arguments for 'config get' command")
	}

	// Create array to hold key-value pairs
	var result []shared.Value

	// Process each configuration parameter
	for _, arg := range args {
		param := arg.Bulk
		paramUpper := strings.ToUpper(param)

		// Get the configuration value
		value := getConfigValue(paramUpper)

		// Add key-value pair to result (preserve original case)
		result = append(result, shared.Value{Typ: "bulk", Bulk: param})
		result = append(result, shared.Value{Typ: "bulk", Bulk: value})
	}

	return shared.Value{Typ: "array", Array: result}
}

// getConfigValue returns the value for a given configuration parameter
func getConfigValue(param string) string {
	switch param {
	case "DIR":
		return getConfigDir()
	case "DBFILENAME":
		return getConfigDbfilename()
	default:
		return ""
	}
}

// getConfigDir returns the current directory configuration
func getConfigDir() string {
	return shared.StoreState.ConfigDir
}

// getConfigDbfilename returns the current database filename configuration
func getConfigDbfilename() string {
	return shared.StoreState.ConfigDbfilename
}
