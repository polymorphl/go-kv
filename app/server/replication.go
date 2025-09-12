package server

import (
	"fmt"
	"net"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

// Mutexes to protect concurrent access to replication data
var replicasMu sync.RWMutex
var acknowledgedReplicasMu sync.RWMutex

// AcknowledgedReplicas tracks which replicas have acknowledged commands
var AcknowledgedReplicas = make(map[string]bool)

// IsWriteCommand checks if a command modifies data and should be propagated to replicas
func IsWriteCommand(command string) bool {
	writeCommands := map[string]bool{
		"SET":     true,
		"LPUSH":   true,
		"RPUSH":   true,
		"LPOP":    true,
		"BLPOP":   true,
		"INCR":    true,
		"XADD":    true,
		"MULTI":   true,
		"EXEC":    true,
		"DISCARD": true,
	}
	return writeCommands[command]
}

// PropagateCommand sends a command to all connected replicas
func PropagateCommand(command string, args []protocol.Value) {
	if StoreState.Role != "master" {
		return
	}

	// Build the command array for propagation
	commandArray := make([]protocol.Value, len(args)+1)
	commandArray[0] = protocol.Value{Typ: "bulk", Bulk: command}
	for i, arg := range args {
		commandArray[i+1] = arg
	}

	// Snapshot replicas under read lock to avoid concurrent map iteration/writes
	replicasMu.RLock()
	snapshot := make(map[string]net.Conn, len(StoreState.Replicas))
	for id, c := range StoreState.Replicas {
		snapshot[id] = c
	}
	replicasMu.RUnlock()

	// Send to all replicas using the snapshot
	for replicaID, replicaConn := range snapshot {
		bytes := protocol.Value{Typ: "array", Array: commandArray}.Marshal()
		_, err := replicaConn.Write(bytes)
		if err != nil {
			// Remove failed replica connection
			ReplicasDelete(replicaID)
			fmt.Printf("Failed to propagate command to replica %s: %v\n", replicaID, err)
		}
	}
}

// Replicas helpers
func ReplicasGet(connID string) (net.Conn, bool) {
	replicasMu.RLock()
	c, ok := StoreState.Replicas[connID]
	replicasMu.RUnlock()
	return c, ok
}

func ReplicasSet(connID string, conn net.Conn) {
	replicasMu.Lock()
	StoreState.Replicas[connID] = conn
	replicasMu.Unlock()
}

func ReplicasDelete(connID string) {
	replicasMu.Lock()
	delete(StoreState.Replicas, connID)
	replicasMu.Unlock()
}

// AcknowledgedReplicasSet marks a replica as having acknowledged
func AcknowledgedReplicasSet(connID string) {
	acknowledgedReplicasMu.Lock()
	AcknowledgedReplicas[connID] = true
	acknowledgedReplicasMu.Unlock()
}

// AcknowledgedReplicasCount returns the number of replicas that have acknowledged
func AcknowledgedReplicasCount() int {
	acknowledgedReplicasMu.RLock()
	count := len(AcknowledgedReplicas)
	acknowledgedReplicasMu.RUnlock()
	return count
}

// AcknowledgedReplicasClear clears all acknowledgments (used before sending new GETACK)
func AcknowledgedReplicasClear() {
	acknowledgedReplicasMu.Lock()
	AcknowledgedReplicas = make(map[string]bool)
	acknowledgedReplicasMu.Unlock()
}

// SendReplconfGetack sends REPLCONF GETACK * to all replicas
func SendReplconfGetack() {
	// Clear previous acknowledgments before sending new GETACK
	AcknowledgedReplicasClear()

	cmd := protocol.Value{Typ: "array", Array: []protocol.Value{
		{Typ: "bulk", Bulk: "REPLCONF"},
		{Typ: "bulk", Bulk: "GETACK"},
		{Typ: "bulk", Bulk: "*"},
	}}
	bytes := cmd.Marshal()

	replicasMu.RLock()
	replicas := make(map[string]net.Conn, len(StoreState.Replicas))
	for id, c := range StoreState.Replicas {
		replicas[id] = c
	}
	replicasMu.RUnlock()

	for replicaID, replicaConn := range replicas {
		_, err := replicaConn.Write(bytes)
		if err != nil {
			ReplicasDelete(replicaID)
			fmt.Printf("Failed to send GETACK to replica %s: %v\n", replicaID, err)
		}
	}
}
