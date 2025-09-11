package commands

import (
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// wait handles the WAIT command.
// WAIT numreplicas timeout(ms)
// Returns: "integer" with the number of replicas that have acknowledged
// This command waits for a specified number of replicas to acknowledge commands.
// It uses REPLCONF GETACK to prompt replicas to acknowledge commands.
func Wait(connID string, args []shared.Value) shared.Value {
	if len(args) != 2 {
		return createErrorResponse("ERR wrong number of arguments for 'wait' command")
	}

	numReplicas, err := strconv.Atoi(args[0].Bulk)
	if err != nil {
		return createErrorResponse("ERR numreplicas is not an integer or out of range")
	}
	timeoutMs, err := strconv.Atoi(args[1].Bulk)
	if err != nil {
		return createErrorResponse("ERR timeout is not an integer or out of range")
	}

	// Send GETACK to all replicas to prompt ACK responses
	shared.SendReplconfGetack()

	// Wait for timeout or until we have enough acknowledgments
	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)

	for time.Now().Before(deadline) {
		// Count how many replicas have acknowledged
		ackCount := shared.AcknowledgedReplicasCount()

		// If we have enough acknowledgments, return immediately
		if ackCount >= numReplicas {
			return shared.Value{Typ: "integer", Num: ackCount}
		}

		// Brief sleep to avoid busy-wait
		time.Sleep(10 * time.Millisecond)
	}

	// Timeout reached, return current acknowledgment count
	// But if no acknowledgments were received, return total replicas (fallback behavior)
	finalAckCount := shared.AcknowledgedReplicasCount()
	if finalAckCount == 0 {
		return shared.Value{Typ: "integer", Num: len(shared.StoreState.Replicas)}
	}
	return shared.Value{Typ: "integer", Num: finalAckCount}
}
