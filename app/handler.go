package main

import (
	"github.com/codecrafters-io/redis-starter-go/app/commands"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// Handlers maps Redis command names to their corresponding handler functions.
// Each handler function takes an array of Value arguments and returns a Value response.
var Handlers = map[string]func([]shared.Value) shared.Value{
	// utils commands
	"PING": commands.Ping,
	"ECHO": commands.Echo,
	"TYPE": commands.Type,
	// string commands
	"SET":  commands.Set,
	"GET":  commands.Get,
	"INCR": commands.Incr,
	// list commands
	"LPUSH":  commands.Lpush,
	"RPUSH":  commands.Rpush,
	"LRANGE": commands.Lrange,
	"LLEN":   commands.Llen,
	"LPOP":   commands.Lpop,
	"BLPOP":  commands.Blpop,
	// stream commands
	"XADD":   commands.Xadd,
	"XRANGE": commands.Xrange,
	"XREAD":  commands.Xread,
}
