package commands

import "github.com/codecrafters-io/redis-starter-go/app/shared"

// createErrorResponse creates a standardized error response.
func createErrorResponse(message string) shared.Value {
	return shared.Value{Typ: "error", Str: message}
}
