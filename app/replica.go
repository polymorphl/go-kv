package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// sendPing sends a PING command to the master
func sendPing(conn net.Conn, writer *Writer, reader *shared.Resp) {
	err := writer.Write(shared.Value{Typ: "array", Array: []shared.Value{
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
func sendReplConfListeningPort(conn net.Conn, writer *Writer, reader *shared.Resp, port string) {
	err := writer.Write(shared.Value{Typ: "array", Array: []shared.Value{
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
func sendReplConfCapaPsync2(conn net.Conn, writer *Writer, reader *shared.Resp) {
	err := writer.Write(shared.Value{Typ: "array", Array: []shared.Value{
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
func sendPsync(conn net.Conn, writer *Writer, reader *shared.Resp, masterReplID string, masterReplOffset int64) {
	err := writer.Write(shared.Value{Typ: "array", Array: []shared.Value{
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
func performReplicationHandshake(address, port string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Printf("Failed to connect to master %s: %s\n", address, err.Error())
		return // Don't exit, just return and let the server start
	}
	// Note: We don't close the connection here - keep it alive for replication

	writer := NewWriter(conn)
	reader := shared.NewResp(conn)

	// Step 1: Send PING and wait for PONG response
	sendPing(conn, writer, reader)

	// Step 2: Send REPLCONF listening-port and wait for OK response
	sendReplConfListeningPort(conn, writer, reader, port)

	// Step 3: Send REPLCONF capa psync2 and wait for OK response
	sendReplConfCapaPsync2(conn, writer, reader)

	// Step 4: Send PSYNC and wait for FULLRESYNC response
	sendPsync(conn, writer, reader, "?", -1)

	// Step 5: Start listening for propagated commands
	go processPropagatedCommands(conn, shared.NewResp(conn))
}

func connectToMaster(replicaPort string) {
	parts := strings.Split(shared.StoreState.ReplicaOf, " ")
	if len(parts) != 2 {
		fmt.Println("Invalid replicaof format. Expected 'host port'")
		os.Exit(1)
	}

	host := parts[0]
	masterPort := parts[1]
	address := host + ":" + masterPort

	performReplicationHandshake(address, replicaPort)
}

// processPropagatedCommands processes commands propagated from the master
func processPropagatedCommands(conn net.Conn, reader *shared.Resp) {
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

		command := strings.ToUpper(value.Array[0].Bulk)
		args := value.Array[1:]

		// Execute the command using the shared handlers
		// Use a dummy connection ID for replica commands
		connID := "replica"
		shared.ExecuteCommand(command, connID, args)
	}
}

// handleReplicaMode sets up the server as a replica and connects to master
func handleReplicaMode(replicaPort string) {
	if shared.StoreState.Role == "slave" {
		// Start replica connection in a goroutine so it doesn't block the server startup
		go connectToMaster(replicaPort)
	}
}
