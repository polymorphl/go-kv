package shared

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

// Wrapper functions for replication functionality
// These maintain backward compatibility while delegating to the server package

// IsWriteCommand checks if a command modifies data and should be propagated to replicas
func IsWriteCommand(command string) bool {
	return server.IsWriteCommand(command)
}

// PropagateCommand sends a command to all connected replicas
func PropagateCommand(command string, args []protocol.Value) {
	server.PropagateCommand(command, args)
}

// Replicas helpers
func ReplicasGet(connID string) (net.Conn, bool) {
	return server.ReplicasGet(connID)
}

func ReplicasSet(connID string, conn net.Conn) {
	server.ReplicasSet(connID, conn)
}

func ReplicasDelete(connID string) {
	server.ReplicasDelete(connID)
}

// AcknowledgedReplicasSet marks a replica as having acknowledged
func AcknowledgedReplicasSet(connID string) {
	server.AcknowledgedReplicasSet(connID)
}

// AcknowledgedReplicasCount returns the number of replicas that have acknowledged
func AcknowledgedReplicasCount() int {
	return server.AcknowledgedReplicasCount()
}

// AcknowledgedReplicasClear clears all acknowledgments (used before sending new GETACK)
func AcknowledgedReplicasClear() {
	server.AcknowledgedReplicasClear()
}

// SendReplconfGetack sends REPLCONF GETACK * to all replicas
func SendReplconfGetack() {
	server.SendReplconfGetack()
}
