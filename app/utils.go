package main

// createErrorResponse creates a standardized error response.
func createErrorResponse(message string) Value {
	return Value{Typ: "error", Str: message}
}
