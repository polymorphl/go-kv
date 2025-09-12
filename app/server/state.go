package server

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// Global server state
var StoreState = &shared.State{
	Role:             "master",
	ReplicaOf:        "",
	MasterReplID:     "",
	MasterReplOffset: 0,
	Replicas:         make(map[string]net.Conn),
	ConfigDir:        "/tmp/redis-data",
	ConfigDbfilename: "rdbfile",
}

// Memory is the global in-memory database that stores all key-value pairs.
var Memory = make(map[string]shared.MemoryEntry)

// InitializeSharedState initializes the shared.StoreState with our server state
func InitializeSharedState() {
	shared.StoreState = StoreState
}
