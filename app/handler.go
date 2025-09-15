package main

import (
	"github.com/codecrafters-io/redis-starter-go/app/commands"
	"github.com/codecrafters-io/redis-starter-go/app/network"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

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

// Handlers maps Redis command names to their corresponding handler functions.
// Each handler function takes a connection ID and an array of Value arguments, and returns a Value response.
var Handlers = map[string]func(string, []shared.Value) shared.Value{
	"BLPOP":       commands.Blpop,
	"CONFIG":      commands.Config,
	"DISCARD":     commands.Discard,
	"ECHO":        commands.Echo,
	"EXEC":        commands.Exec,
	"GET":         commands.Get,
	"GEOADD":      commands.Geoadd,
	"GEODIST":     commands.Geodist,
	"GEOPOS":      commands.Geopos,
	"INCR":        commands.Incr,
	"INFO":        commands.Info,
	"KEYS":        commands.Keys,
	"LLEN":        commands.Llen,
	"LPOP":        commands.Lpop,
	"LPUSH":       commands.Lpush,
	"LRANGE":      commands.Lrange,
	"MULTI":       commands.Multi,
	"PING":        commands.Ping,
	"PSYNC":       commands.Psync,
	"PUBLISH":     commands.Publish,
	"REPLCONF":    commands.Replconf,
	"RPUSH":       commands.Rpush,
	"SET":         commands.Set,
	"SUBSCRIBE":   commands.Subscribe,
	"TYPE":        commands.Type,
	"UNSUBSCRIBE": commands.Unsubscribe,
	"WAIT":        commands.Wait,
	"XADD":        commands.Xadd,
	"XRANGE":      commands.Xrange,
	"XREAD":       commands.Xread,
	"ZADD":        commands.Zadd,
	"ZCARD":       commands.Zcard,
	"ZRANGE":      commands.Zrange,
	"ZREM":        commands.Zrem,
	"ZSCORE":      commands.Zscore,
	"ZRANK":       commands.Zrank,
}

// init initializes the shared command handlers map
func init() {
	network.CommandHandlers = make(map[string]shared.CommandHandler)
	for cmd, handler := range Handlers {
		network.CommandHandlers[cmd] = handler
	}
}
