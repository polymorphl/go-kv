package shared

// Memory is the global in-memory database that stores all key-value pairs.
var Memory = make(map[string]MemoryEntry)
