package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
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
		r := NewResp(conn)
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
			err := writer.Write(Value{Typ: "string", Str: ""})
			if err != nil {
				fmt.Println("Error writing response:", err)
				break
			}
			continue
		}

		result := handler(args)
		writer.Write(result)
	}
}
