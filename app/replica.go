package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// performReplicationHandshake performs the complete replication handshake with master
func performReplicationHandshake(address, port string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Printf("Failed to connect to master %s: %s\n", address, err.Error())
		os.Exit(1)
	}
	// Note: We don't close the connection here - keep it alive for replication

	writer := NewWriter(conn)
	reader := shared.NewResp(conn)

	// Step 1: Send PING and wait for PONG response
	err = writer.Write(shared.Value{Typ: "array", Array: []shared.Value{
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

	// Step 2: Send REPLCONF listening-port and wait for OK response
	err = writer.Write(shared.Value{Typ: "array", Array: []shared.Value{
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

	// Step 3: Send REPLCONF capa psync2 and wait for OK response
	err = writer.Write(shared.Value{Typ: "array", Array: []shared.Value{
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

	// Handshake completed successfully - connection remains open
	fmt.Println("Replication handshake completed successfully")
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

// handleReplicaMode sets up the server as a replica and connects to master
func handleReplicaMode(replicaPort string) {
	if shared.StoreState.Role == "slave" {
		connectToMaster(replicaPort)
	}
}
