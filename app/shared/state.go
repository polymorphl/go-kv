package shared

// StoreState provides access to the global server state
// This will be initialized by the server package
var StoreState *State

// Helper functions for test compatibility
func SetStoreState(state State) {
	if StoreState != nil {
		*StoreState = state
	}
}

func GetStoreState() *State {
	return StoreState
}
