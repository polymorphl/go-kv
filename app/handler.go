package main

var Handlers = map[string]func([]Value) Value{
	"PING": ping,
	"ECHO": echo,
	"SET":  set,
	"GET":  get,
}

var memory = make(map[string]string)

func ping(args []Value) Value {
	if len(args) == 0 {
		return Value{Typ: "string", Str: "PONG"}
	}

	return Value{Typ: "string", Str: args[0].Bulk}
}

func echo(args []Value) Value {
	return Value{Typ: "string", Str: args[0].Bulk}
}

func set(args []Value) Value {
	memory[args[0].Bulk] = args[1].Bulk
	return Value{Typ: "string", Str: "OK"}
}

func get(args []Value) Value {
	tmp := memory[args[0].Bulk]
	if tmp == "" {
		return Value{Typ: "string", Str: ""}
	}

	return Value{Typ: "string", Str: tmp}
}
