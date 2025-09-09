package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

const DEFAULT_PORT = "6379"

var port = ""
var replicaOf = ""

// generateReplID generates a random 40-character alphanumeric string for replication ID
func generateReplID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	const length = 40

	b := make([]byte, length)
	rand.Read(b)

	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}

	return string(b)
}

// Parse command line arguments
func parseArgs() string {
	flag.StringVar(&port, "port", DEFAULT_PORT, "Port to listen on")
	flag.StringVar(&replicaOf, "replicaof", "", "Replica of")
	flag.Parse()

	if replicaOf != "" && strings.Contains(replicaOf, " ") {
		shared.StoreState = shared.State{
			Role:      "slave",
			ReplicaOf: replicaOf,
		}
	}

	if shared.StoreState.Role == "master" {
		shared.StoreState.MasterReplID = generateReplID()
	}

	return port
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:"+parseArgs())
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		r := shared.NewResp(conn)
		value, err := r.Read()
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client disconnected: ", conn.RemoteAddr().String())
			} else {
				fmt.Println("ERR IS", err)
			}
			return
		} else {
			fmt.Println("Client connected: ", conn.RemoteAddr().String())
		}

		if value.Typ != "array" {
			fmt.Println("Invalid request, expected array")
			continue
		}

		command := strings.ToUpper(value.Array[0].Bulk)
		args := value.Array[1:]

		writer := NewWriter(conn)
		handler, ok := Handlers[command]

		if !ok {
			fmt.Println("Invalid command: ", command)
			err := writer.Write(shared.Value{Typ: "string", Str: ""})
			if err != nil {
				fmt.Println("Error writing response:", err)
				break
			}
			continue
		}

		// Use connection remote address as connection ID
		connID := conn.RemoteAddr().String()

		// Check if this connection is in a transaction
		if transaction, exists := shared.Transactions[connID]; exists {
			// If it's a transaction command, execute it normally
			if IsTransactionCommand(command) {
				result := handler(connID, args)
				writer.Write(result)
			} else {
				// Queue the command instead of executing it
				transaction.Commands = append(transaction.Commands, shared.QueuedCommand{
					Command: command,
					Args:    args,
				})
				shared.Transactions[connID] = transaction

				// Return QUEUED response
				result := shared.Value{Typ: "string", Str: "QUEUED"}
				writer.Write(result)
			}
		} else {
			// No active transaction, execute command normally
			result := handler(connID, args)
			writer.Write(result)
		}
	}
}
