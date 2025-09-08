package main

import (
	"github.com/codecrafters-io/redis-starter-go/app/commands"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// Handlers maps Redis command names to their corresponding handler functions.
// Each handler function takes a connection ID and an array of Value arguments, and returns a Value response.
var Handlers = map[string]func(string, []shared.Value) shared.Value{
	"BLPOP":  commands.Blpop,
	"ECHO":   commands.Echo,
	"EXEC":   commands.Exec,
	"GET":    commands.Get,
	"INCR":   commands.Incr,
	"LLEN":   commands.Llen,
	"LPOP":   commands.Lpop,
	"LPUSH":  commands.Lpush,
	"LRANGE": commands.Lrange,
	"MULTI":  commands.Multi,
	"PING":   commands.Ping,
	"RPUSH":  commands.Rpush,
	"SET":    commands.Set,
	"TYPE":   commands.Type,
	"XADD":   commands.Xadd,
	"XRANGE": commands.Xrange,
	"XREAD":  commands.Xread,
}

// init initializes the shared command handlers map
func init() {
	shared.CommandHandlers = make(map[string]shared.CommandHandler)
	for cmd, handler := range Handlers {
		shared.CommandHandlers[cmd] = handler
	}
}
