package commands

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/network"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func TestExec(t *testing.T) {
	// Initialize command handlers for testing
	initCommandHandlers()

	// Clear memory and transactions before each test
	clearMemory()
	clearTransactions()

	tests := []struct {
		name     string
		connID   string
		args     []shared.Value
		setup    func() // Setup function to prepare transaction
		expected shared.Value
		verify   func() // Function to verify the result
	}{
		{
			name:     "exec without multi should fail",
			connID:   "test-conn-1",
			args:     []shared.Value{},
			setup:    func() {}, // No transaction setup
			expected: shared.Value{Typ: "error", Str: "ERR EXEC without MULTI"},
			verify: func() {
				// Transaction should not exist after error
				if _, exists := network.Transactions["test-conn-1"]; exists {
					t.Error("Transaction should not exist after error")
				}
			},
		},
		{
			name:   "exec with arguments should fail",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "extra"},
			},
			setup: func() {
				network.Transactions["test-conn-2"] = shared.Transaction{Commands: []shared.QueuedCommand{}}
			},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'exec' command"},
			verify: func() {
				// Transaction should still exist after argument error
				if _, exists := network.Transactions["test-conn-2"]; !exists {
					t.Error("Transaction should still exist after argument error")
				}
			},
		},
		{
			name:   "exec empty transaction",
			connID: "test-conn-3",
			args:   []shared.Value{},
			setup: func() {
				network.Transactions["test-conn-3"] = shared.Transaction{Commands: []shared.QueuedCommand{}}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{}},
			verify: func() {
				// Transaction should be cleared after EXEC
				if _, exists := network.Transactions["test-conn-3"]; exists {
					t.Error("Transaction should be cleared after EXEC")
				}
			},
		},
		{
			name:   "exec single command",
			connID: "test-conn-4",
			args:   []shared.Value{},
			setup: func() {
				network.Transactions["test-conn-4"] = shared.Transaction{
					Commands: []shared.QueuedCommand{
						{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "key1"}, {Typ: "bulk", Bulk: "value1"}}},
					},
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{{Typ: "string", Str: "OK"}}},
			verify: func() {
				// Transaction should be cleared after EXEC
				if _, exists := network.Transactions["test-conn-4"]; exists {
					t.Error("Transaction should be cleared after EXEC")
				}
				// Command should have been executed
				entry, exists := server.Memory["key1"]
				if !exists {
					t.Error("Key should exist after EXEC")
				}
				if entry.Value != "value1" {
					t.Errorf("Expected value 'value1', got '%s'", entry.Value)
				}
			},
		},
		{
			name:   "exec multiple commands",
			connID: "test-conn-5",
			args:   []shared.Value{},
			setup: func() {
				network.Transactions["test-conn-5"] = shared.Transaction{
					Commands: []shared.QueuedCommand{
						{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "key1"}, {Typ: "bulk", Bulk: "value1"}}},
						{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "key2"}, {Typ: "bulk", Bulk: "value2"}}},
						{Command: "GET", Args: []shared.Value{{Typ: "bulk", Bulk: "key1"}}},
					},
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{
				{Typ: "string", Str: "OK"},
				{Typ: "string", Str: "OK"},
				{Typ: "string", Str: "value1"},
			}},
			verify: func() {
				// Transaction should be cleared after EXEC
				if _, exists := network.Transactions["test-conn-5"]; exists {
					t.Error("Transaction should be cleared after EXEC")
				}
				// Commands should have been executed
				entry1, exists := server.Memory["key1"]
				if !exists {
					t.Error("Key1 should exist after EXEC")
				}
				if entry1.Value != "value1" {
					t.Errorf("Expected value 'value1', got '%s'", entry1.Value)
				}

				entry2, exists := server.Memory["key2"]
				if !exists {
					t.Error("Key2 should exist after EXEC")
				}
				if entry2.Value != "value2" {
					t.Errorf("Expected value 'value2', got '%s'", entry2.Value)
				}
			},
		},
		{
			name:   "exec with invalid command",
			connID: "test-conn-6",
			args:   []shared.Value{},
			setup: func() {
				network.Transactions["test-conn-6"] = shared.Transaction{
					Commands: []shared.QueuedCommand{
						{Command: "INVALID", Args: []shared.Value{{Typ: "bulk", Bulk: "arg1"}}},
					},
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{{Typ: "string", Str: ""}}},
			verify: func() {
				// Transaction should be cleared after EXEC
				if _, exists := network.Transactions["test-conn-6"]; exists {
					t.Error("Transaction should be cleared after EXEC")
				}
			},
		},
		{
			name:   "exec with command that has wrong arguments",
			connID: "test-conn-7",
			args:   []shared.Value{},
			setup: func() {
				network.Transactions["test-conn-7"] = shared.Transaction{
					Commands: []shared.QueuedCommand{
						{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "key1"}}}, // Missing value
					},
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{{Typ: "error", Str: "ERR wrong number of arguments for 'set' command"}}},
			verify: func() {
				// Transaction should be cleared after EXEC
				if _, exists := network.Transactions["test-conn-7"]; exists {
					t.Error("Transaction should be cleared after EXEC")
				}
				// Key should not exist due to error
				if _, exists := server.Memory["key1"]; exists {
					t.Error("Key should not exist after error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()
			clearTransactions()

			// Setup transaction if needed
			tt.setup()

			result := Exec(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Exec() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Typ == "array" {
				if len(result.Array) != len(tt.expected.Array) {
					t.Errorf("Exec() array length = %v, expected %v", len(result.Array), len(tt.expected.Array))
				}
				for i, val := range result.Array {
					if i < len(tt.expected.Array) {
						if val.Typ != tt.expected.Array[i].Typ {
							t.Errorf("Exec() array[%d] type = %v, expected %v", i, val.Typ, tt.expected.Array[i].Typ)
						}
						if val.Str != tt.expected.Array[i].Str {
							t.Errorf("Exec() array[%d] string = %v, expected %v", i, val.Str, tt.expected.Array[i].Str)
						}
						if val.Bulk != tt.expected.Array[i].Bulk {
							t.Errorf("Exec() array[%d] bulk = %v, expected %v", i, val.Bulk, tt.expected.Array[i].Bulk)
						}
					}
				}
			} else {
				if result.Str != tt.expected.Str {
					t.Errorf("Exec() string = %v, expected %v", result.Str, tt.expected.Str)
				}
			}

			tt.verify()
		})
	}
}

func TestExecTransactionClearing(t *testing.T) {
	initCommandHandlers()
	clearMemory()
	clearTransactions()

	connID := "test-conn-clear"

	// Setup transaction
	network.Transactions[connID] = shared.Transaction{
		Commands: []shared.QueuedCommand{
			{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "testkey"}, {Typ: "bulk", Bulk: "testvalue"}}},
		},
	}

	// Verify transaction exists before EXEC
	if _, exists := network.Transactions[connID]; !exists {
		t.Error("Transaction should exist before EXEC")
	}

	// Execute EXEC
	result := Exec(connID, []shared.Value{})

	// Verify transaction is cleared after EXEC
	if _, exists := network.Transactions[connID]; exists {
		t.Error("Transaction should be cleared after EXEC")
	}

	// Verify command was executed
	if result.Typ != "array" || len(result.Array) != 1 {
		t.Errorf("EXEC should return array with one result, got %v", result)
	}

	entry, exists := server.Memory["testkey"]
	if !exists {
		t.Error("Key should exist after EXEC")
	}
	if entry.Value != "testvalue" {
		t.Errorf("Expected value 'testvalue', got '%s'", entry.Value)
	}
}

func TestExecMultipleConnections(t *testing.T) {
	initCommandHandlers()
	clearMemory()
	clearTransactions()

	conn1 := "connection-1"
	conn2 := "connection-2"

	// Setup transactions for both connections
	network.Transactions[conn1] = shared.Transaction{
		Commands: []shared.QueuedCommand{
			{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "key1"}, {Typ: "bulk", Bulk: "value1"}}},
		},
	}

	network.Transactions[conn2] = shared.Transaction{
		Commands: []shared.QueuedCommand{
			{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "key2"}, {Typ: "bulk", Bulk: "value2"}}},
		},
	}

	// Execute EXEC on first connection
	result1 := Exec(conn1, []shared.Value{})
	if result1.Typ != "array" || len(result1.Array) != 1 {
		t.Errorf("EXEC on conn1 should return array with one result, got %v", result1)
	}

	// Execute EXEC on second connection
	result2 := Exec(conn2, []shared.Value{})
	if result2.Typ != "array" || len(result2.Array) != 1 {
		t.Errorf("EXEC on conn2 should return array with one result, got %v", result2)
	}

	// Verify both transactions are cleared
	if _, exists := network.Transactions[conn1]; exists {
		t.Error("Transaction for conn1 should be cleared")
	}

	if _, exists := network.Transactions[conn2]; exists {
		t.Error("Transaction for conn2 should be cleared")
	}

	// Verify both commands were executed
	entry1, exists := server.Memory["key1"]
	if !exists {
		t.Error("Key1 should exist after EXEC")
	}
	if entry1.Value != "value1" {
		t.Errorf("Expected value 'value1', got '%s'", entry1.Value)
	}

	entry2, exists := server.Memory["key2"]
	if !exists {
		t.Error("Key2 should exist after EXEC")
	}
	if entry2.Value != "value2" {
		t.Errorf("Expected value 'value2', got '%s'", entry2.Value)
	}
}

func BenchmarkExec(b *testing.B) {
	clearMemory()
	clearTransactions()

	connID := "benchmark-conn"
	args := []shared.Value{}

	// Setup transaction
	network.Transactions[connID] = shared.Transaction{
		Commands: []shared.QueuedCommand{
			{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "key"}, {Typ: "bulk", Bulk: "value"}}},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Exec(connID, args)
		// Re-setup transaction after each iteration
		network.Transactions[connID] = shared.Transaction{
			Commands: []shared.QueuedCommand{
				{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "key"}, {Typ: "bulk", Bulk: "value"}}},
			},
		}
	}
}

func BenchmarkExecEmptyTransaction(b *testing.B) {
	clearTransactions()

	connID := "benchmark-conn"
	args := []shared.Value{}

	// Setup empty transaction
	network.Transactions[connID] = shared.Transaction{Commands: []shared.QueuedCommand{}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Exec(connID, args)
		// Re-setup empty transaction after each iteration
		network.Transactions[connID] = shared.Transaction{Commands: []shared.QueuedCommand{}}
	}
}

func BenchmarkExecMultipleCommands(b *testing.B) {
	clearMemory()
	clearTransactions()

	connID := "benchmark-conn"
	args := []shared.Value{}

	// Setup transaction with multiple commands
	network.Transactions[connID] = shared.Transaction{
		Commands: []shared.QueuedCommand{
			{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "key1"}, {Typ: "bulk", Bulk: "value1"}}},
			{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "key2"}, {Typ: "bulk", Bulk: "value2"}}},
			{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "key3"}, {Typ: "bulk", Bulk: "value3"}}},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Exec(connID, args)
		// Re-setup transaction after each iteration
		network.Transactions[connID] = shared.Transaction{
			Commands: []shared.QueuedCommand{
				{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "key1"}, {Typ: "bulk", Bulk: "value1"}}},
				{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "key2"}, {Typ: "bulk", Bulk: "value2"}}},
				{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "key3"}, {Typ: "bulk", Bulk: "value3"}}},
			},
		}
	}
}
