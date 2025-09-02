package main

var Handlers = map[string]func([]Value) Value{
	"PING": ping,
	"ECHO": echo,
}

func ping(args []Value) Value {
	if len(args) == 0 {
		return Value{Typ: "string", Str: "PONG"}
	}

	return Value{Typ: "string", Str: args[0].Bulk}
}

func echo(args []Value) Value {
	return Value{Typ: "string", Str: args[0].Bulk}
}
