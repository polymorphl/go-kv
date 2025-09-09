package commands

import "github.com/codecrafters-io/redis-starter-go/app/shared"

func Replconf(connID string, args []shared.Value) shared.Value {
	return shared.Value{Typ: "string", Str: "OK"}
}
