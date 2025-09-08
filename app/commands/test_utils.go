package commands

import "github.com/codecrafters-io/redis-starter-go/app/shared"

// clearMemory clears all entries from the shared memory for testing
func clearMemory() {
	for k := range shared.Memory {
		delete(shared.Memory, k)
	}
}
