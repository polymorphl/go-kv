package main

import (
	"github.com/codecrafters-io/redis-starter-go/app/commands"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// Handlers maps Redis command names to their corresponding handler functions.
// Each handler function takes an array of Value arguments and returns a Value response.
var Handlers = map[string]func([]shared.Value) shared.Value{
	"BLPOP":  commands.Blpop,
	"ECHO":   commands.Echo,
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
