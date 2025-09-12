package server

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

// sendPing sends a PING command to the master
func sendPing(conn net.Conn, writer *protocol.Writer, reader *protocol.Resp) {
	err := writer.Write(protocol.Value{Typ: "array", Array: []protocol.Value{
		{Typ: "bulk", Bulk: "PING"},
	}})
	if err != nil {
		fmt.Printf("Failed to send PING: %s\n", err.Error())
		conn.Close()
		return
	}
	_, err = reader.Read()
	if err != nil {
		fmt.Printf("Failed to read PONG response: %s\n", err.Error())
		conn.Close()
		return
	}
}

// sendReplConfListeningPort sends a REPLCONF listening-port command to the master
func sendReplConfListeningPort(conn net.Conn, writer *protocol.Writer, reader *protocol.Resp, port string) {
	err := writer.Write(protocol.Value{Typ: "array", Array: []protocol.Value{
		{Typ: "bulk", Bulk: "REPLCONF"},
		{Typ: "bulk", Bulk: "listening-port"},
		{Typ: "bulk", Bulk: port},
	}})
	if err != nil {
		fmt.Printf("Failed to send REPLCONF listening-port: %s\n", err.Error())
		conn.Close()
		return
	}
	_, err = reader.Read()
	if err != nil {
		fmt.Printf("Failed to read REPLCONF listening-port response: %s\n", err.Error())
		conn.Close()
		return
	}
}

// sendReplConfCapaPsync2 sends a REPLCONF capa psync2 command to the master
func sendReplConfCapaPsync2(conn net.Conn, writer *protocol.Writer, reader *protocol.Resp) {
	err := writer.Write(protocol.Value{Typ: "array", Array: []protocol.Value{
		{Typ: "bulk", Bulk: "REPLCONF"},
		{Typ: "bulk", Bulk: "capa"},
		{Typ: "bulk", Bulk: "psync2"},
	}})
	if err != nil {
		fmt.Printf("Failed to send REPLCONF capa: %s\n", err.Error())
		conn.Close()
		return
	}
	_, err = reader.Read()
	if err != nil {
		fmt.Printf("Failed to read REPLCONF capa response: %s\n", err.Error())
		conn.Close()
		return
	}
}

// sendPsync sends a PSYNC command to the master
func sendPsync(conn net.Conn, writer *protocol.Writer, reader *protocol.Resp, masterReplID string, masterReplOffset int64) {
	err := writer.Write(protocol.Value{Typ: "array", Array: []protocol.Value{
		{Typ: "bulk", Bulk: "PSYNC"},
		{Typ: "bulk", Bulk: masterReplID},
		{Typ: "bulk", Bulk: strconv.FormatInt(masterReplOffset, 10)},
	}})
	if err != nil {
		fmt.Printf("Failed to send PSYNC: %s\n", err.Error())
		conn.Close()
		return
	}

	// Read the FULLRESYNC response
	_, err = reader.Read()
	if err != nil {
		fmt.Printf("Failed to read PSYNC response: %s\n", err.Error())
		conn.Close()
		return
	}

	// Read the RDB file (this is binary data, not a command)
	// Use the RESP reader to read it as a bulk string without trailing CRLF
	_, err = reader.ReadBulkWithoutCRLF()
	if err != nil {
		fmt.Printf("Failed to read RDB file: %s\n", err.Error())
		conn.Close()
		return
	}
}

// performReplicationHandshake performs the complete replication handshake with master
func performReplicationHandshake(address, port string, executeCommand func(string, string, []protocol.Value) protocol.Value) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Printf("Failed to connect to master %s: %s\n", address, err.Error())
		return // Don't exit, just return and let the server start
	}
	// Note: We don't close the connection here - keep it alive for replication

	writer := protocol.NewWriter(conn)
	reader := protocol.NewResp(conn)

	// Step 1: Send PING and wait for PONG response
	sendPing(conn, writer, reader)

	// Step 2: Send REPLCONF listening-port and wait for OK response
	sendReplConfListeningPort(conn, writer, reader, port)

	// Step 3: Send REPLCONF capa psync2 and wait for OK response
	sendReplConfCapaPsync2(conn, writer, reader)

	// Step 4: Send PSYNC and wait for FULLRESYNC response
	sendPsync(conn, writer, reader, "?", -1)

	// Step 5: Start listening for propagated commands
	// Reuse the same RESP reader to avoid losing any buffered bytes
	go processPropagatedCommands(conn, reader, executeCommand)
}

func connectToMaster(replicaPort string, replicaOf string, executeCommand func(string, string, []protocol.Value) protocol.Value) {
	parts := strings.Split(replicaOf, " ")
	if len(parts) != 2 {
		fmt.Println("Invalid replicaof format. Expected 'host port'")
		os.Exit(1)
	}

	host := parts[0]
	masterPort := parts[1]
	address := host + ":" + masterPort

	performReplicationHandshake(address, replicaPort, executeCommand)
}

// processPropagatedCommands processes commands propagated from the master
func processPropagatedCommands(conn net.Conn, reader *protocol.Resp, executeCommand func(string, string, []protocol.Value) protocol.Value) {
	writer := protocol.NewWriter(conn)
	var processedOffset int64 = 0

	for {
		value, err := reader.Read()
		if err != nil {
			fmt.Printf("Error reading propagated command: %v\n", err)
			return
		}

		if value.Typ != "array" {
			continue
		}

		if len(value.Array) == 0 {
			continue
		}

		// Count bytes consumed by this command in RESP form
		bytesConsumed := int64(len(value.Marshal()))

		command := strings.ToUpper(value.Array[0].Bulk)
		args := value.Array[1:]

		// For REPLCONF GETACK, respond with current offset before including this command
		if command == "REPLCONF" && len(args) >= 1 && strings.ToUpper(args[0].Bulk) == "GETACK" {
			ack := protocol.Value{Typ: "array", Array: []protocol.Value{
				{Typ: "bulk", Bulk: "REPLCONF"},
				{Typ: "bulk", Bulk: "ACK"},
				{Typ: "bulk", Bulk: strconv.FormatInt(processedOffset, 10)},
			}}
			if err := writer.Write(ack); err != nil {
				fmt.Printf("Error writing REPLCONF GETACK response: %v\n", err)
				return
			}
			// Flush to ensure the response is sent immediately
			if err := writer.Flush(); err != nil {
				fmt.Printf("Error flushing REPLCONF GETACK response: %v\n", err)
				return
			}
			processedOffset += bytesConsumed
			continue
		}

		// Execute the command using the provided handler; ignore response to master
		connID := conn.RemoteAddr().String()
		_ = executeCommand(command, connID, args)
		processedOffset += bytesConsumed
	}
}

// HandleReplicaMode sets up the server as a replica and connects to master
func HandleReplicaMode(replicaPort string, role string, replicaOf string, executeCommand func(string, string, []protocol.Value) protocol.Value) {
	if role == "slave" {
		// Start replica connection in a goroutine so it doesn't block the server startup
		go connectToMaster(replicaPort, replicaOf, executeCommand)
	}
}
