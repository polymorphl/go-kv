package main

import (
	"github.com/codecrafters-io/redis-starter-go/app/commands"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// Handlers maps Redis command names to their corresponding handler functions.
// Each handler function takes a connection ID and an array of Value arguments, and returns a Value response.
var Handlers = map[string]func(string, []shared.Value) shared.Value{
	"BLPOP":   commands.Blpop,
	"DISCARD": commands.Discard,
	"ECHO":    commands.Echo,
	"EXEC":    commands.Exec,
	"GET":     commands.Get,
	"INCR":    commands.Incr,
	"LLEN":    commands.Llen,
	"LPOP":    commands.Lpop,
	"LPUSH":   commands.Lpush,
	"LRANGE":  commands.Lrange,
	"MULTI":   commands.Multi,
	"PING":    commands.Ping,
	"RPUSH":   commands.Rpush,
	"SET":     commands.Set,
	"TYPE":    commands.Type,
	"XADD":    commands.Xadd,
	"XRANGE":  commands.Xrange,
	"XREAD":   commands.Xread,
}

// TransactionCommands contains commands that should be executed normally even during a transaction
var TransactionCommands = []string{"MULTI", "EXEC", "DISCARD"}

// IsTransactionCommand checks if a command should be executed normally during a transaction
func IsTransactionCommand(command string) bool {
	for _, cmd := range TransactionCommands {
		if cmd == command {
			return true
		}
	}
	return false
}

// init initializes the shared command handlers map
func init() {
	shared.CommandHandlers = make(map[string]shared.CommandHandler)
	for cmd, handler := range Handlers {
		shared.CommandHandlers[cmd] = handler
	}
}
