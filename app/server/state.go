package server

import (
	"net"
)

// State represents the server state including replication information
type State struct {
	Role             string
	ReplicaOf        string
	MasterReplID     string
	MasterReplOffset int64
	Replicas         map[string]net.Conn // Map of replica connection IDs to their connections
	ConfigDir        string              // Directory where Redis stores its data
	ConfigDbfilename string              // Database filename
}

// Global server state
var StoreState = &State{
	Role:             "master",
	ReplicaOf:        "",
	MasterReplID:     "",
	MasterReplOffset: 0,
	Replicas:         make(map[string]net.Conn),
	ConfigDir:        "/tmp/redis-data",
	ConfigDbfilename: "rdbfile",
}
