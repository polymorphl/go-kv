package shared

import (
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

// State is an alias for server.State to maintain backward compatibility
type State = server.State

// StoreState provides access to the global server state
var StoreState = server.StoreState

// Helper functions for test compatibility
func SetStoreState(state State) {
	*server.StoreState = state
}

func GetStoreState() *State {
	return server.StoreState
}
