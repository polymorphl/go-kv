package main

func typeCmd(args []Value) Value {
	if len(args) != 1 {
		return Value{Typ: "error", Str: "ERR wrong number of arguments for 'type' command"}
	}

	key := args[0].Bulk
	_, exists := memory[key]

	if !exists {
		return Value{Typ: "string", Str: "none"}
	}

	return Value{Typ: "string", Str: "string"}
}
