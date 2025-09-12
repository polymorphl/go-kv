package commands

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/network"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func TestDiscard(t *testing.T) {
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
			name:     "discard without multi should fail",
			connID:   "test-conn-1",
			args:     []shared.Value{},
			setup:    func() {}, // No transaction setup
			expected: shared.Value{Typ: "error", Str: "ERR DISCARD without MULTI"},
			verify: func() {
				// Transaction should not exist after error
				if _, exists := network.Transactions["test-conn-1"]; exists {
					t.Error("Transaction should not exist after error")
				}
			},
		},
		{
			name:   "discard with arguments should fail",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "extra"},
			},
			setup: func() {
				network.Transactions["test-conn-2"] = shared.Transaction{Commands: []shared.QueuedCommand{}}
			},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'discard' command"},
			verify: func() {
				// Transaction should still exist after argument error
				if _, exists := network.Transactions["test-conn-2"]; !exists {
					t.Error("Transaction should still exist after argument error")
				}
			},
		},
		{
			name:   "discard with multiple arguments should fail",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "arg1"},
				{Typ: "bulk", Bulk: "arg2"},
			},
			setup: func() {
				network.Transactions["test-conn-3"] = shared.Transaction{Commands: []shared.QueuedCommand{}}
			},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'discard' command"},
			verify: func() {
				// Transaction should still exist after argument error
				if _, exists := network.Transactions["test-conn-3"]; !exists {
					t.Error("Transaction should still exist after argument error")
				}
			},
		},
		{
			name:   "discard empty transaction",
			connID: "test-conn-4",
			args:   []shared.Value{},
			setup: func() {
				network.Transactions["test-conn-4"] = shared.Transaction{Commands: []shared.QueuedCommand{}}
			},
			expected: shared.Value{Typ: "string", Str: "OK"},
			verify: func() {
				// Transaction should be cleared after DISCARD
				if _, exists := network.Transactions["test-conn-4"]; exists {
					t.Error("Transaction should be cleared after DISCARD")
				}
			},
		},
		{
			name:   "discard transaction with commands",
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
			expected: shared.Value{Typ: "string", Str: "OK"},
			verify: func() {
				// Transaction should be cleared after DISCARD
				if _, exists := network.Transactions["test-conn-5"]; exists {
					t.Error("Transaction should be cleared after DISCARD")
				}
				// Commands should NOT have been executed
				if _, exists := server.Memory["key1"]; exists {
					t.Error("Key1 should not exist after DISCARD")
				}
				if _, exists := server.Memory["key2"]; exists {
					t.Error("Key2 should not exist after DISCARD")
				}
			},
		},
		{
			name:   "discard transaction with mixed commands",
			connID: "test-conn-6",
			args:   []shared.Value{},
			setup: func() {
				network.Transactions["test-conn-6"] = shared.Transaction{
					Commands: []shared.QueuedCommand{
						{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "stringkey"}, {Typ: "bulk", Bulk: "stringvalue"}}},
						{Command: "LPUSH", Args: []shared.Value{{Typ: "bulk", Bulk: "listkey"}, {Typ: "bulk", Bulk: "item1"}}},
						{Command: "INCR", Args: []shared.Value{{Typ: "bulk", Bulk: "numkey"}}},
					},
				}
			},
			expected: shared.Value{Typ: "string", Str: "OK"},
			verify: func() {
				// Transaction should be cleared after DISCARD
				if _, exists := network.Transactions["test-conn-6"]; exists {
					t.Error("Transaction should be cleared after DISCARD")
				}
				// Commands should NOT have been executed
				if _, exists := server.Memory["stringkey"]; exists {
					t.Error("String key should not exist after DISCARD")
				}
				if _, exists := server.Memory["listkey"]; exists {
					t.Error("List key should not exist after DISCARD")
				}
				if _, exists := server.Memory["numkey"]; exists {
					t.Error("Number key should not exist after DISCARD")
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

			result := Discard(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Discard() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Str != tt.expected.Str {
				t.Errorf("Discard() string = %v, expected %v", result.Str, tt.expected.Str)
			}

			tt.verify()
		})
	}
}

func TestDiscardTransactionClearing(t *testing.T) {
	initCommandHandlers()
	clearMemory()
	clearTransactions()

	connID := "test-conn-clear"

	// Setup transaction with commands
	network.Transactions[connID] = shared.Transaction{
		Commands: []shared.QueuedCommand{
			{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "testkey"}, {Typ: "bulk", Bulk: "testvalue"}}},
			{Command: "LPUSH", Args: []shared.Value{{Typ: "bulk", Bulk: "testlist"}, {Typ: "bulk", Bulk: "item"}}},
		},
	}

	// Verify transaction exists before DISCARD
	if _, exists := network.Transactions[connID]; !exists {
		t.Error("Transaction should exist before DISCARD")
	}

	// Execute DISCARD
	result := Discard(connID, []shared.Value{})

	// Verify transaction is cleared after DISCARD
	if _, exists := network.Transactions[connID]; exists {
		t.Error("Transaction should be cleared after DISCARD")
	}

	// Verify DISCARD returns OK
	if result.Typ != "string" || result.Str != "OK" {
		t.Errorf("DISCARD should return OK, got %v", result)
	}

	// Verify commands were NOT executed
	if _, exists := server.Memory["testkey"]; exists {
		t.Error("Key should not exist after DISCARD")
	}
	if _, exists := server.Memory["testlist"]; exists {
		t.Error("List should not exist after DISCARD")
	}
}

func TestDiscardMultipleConnections(t *testing.T) {
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

	// Execute DISCARD on first connection
	result1 := Discard(conn1, []shared.Value{})
	if result1.Typ != "string" || result1.Str != "OK" {
		t.Errorf("DISCARD on conn1 should return OK, got %v", result1)
	}

	// Execute DISCARD on second connection
	result2 := Discard(conn2, []shared.Value{})
	if result2.Typ != "string" || result2.Str != "OK" {
		t.Errorf("DISCARD on conn2 should return OK, got %v", result2)
	}

	// Verify both transactions are cleared
	if _, exists := network.Transactions[conn1]; exists {
		t.Error("Transaction for conn1 should be cleared")
	}

	if _, exists := network.Transactions[conn2]; exists {
		t.Error("Transaction for conn2 should be cleared")
	}

	// Verify commands were NOT executed
	if _, exists := server.Memory["key1"]; exists {
		t.Error("Key1 should not exist after DISCARD")
	}

	if _, exists := server.Memory["key2"]; exists {
		t.Error("Key2 should not exist after DISCARD")
	}
}

func TestDiscardVsExec(t *testing.T) {
	initCommandHandlers()
	clearMemory()
	clearTransactions()

	connID := "test-conn-compare"

	// Setup transaction
	network.Transactions[connID] = shared.Transaction{
		Commands: []shared.QueuedCommand{
			{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "comparekey"}, {Typ: "bulk", Bulk: "comparevalue"}}},
		},
	}

	// Test DISCARD first
	result1 := Discard(connID, []shared.Value{})
	if result1.Typ != "string" || result1.Str != "OK" {
		t.Errorf("DISCARD should return OK, got %v", result1)
	}

	// Verify transaction is cleared
	if _, exists := network.Transactions[connID]; exists {
		t.Error("Transaction should be cleared after DISCARD")
	}

	// Verify command was NOT executed
	if _, exists := server.Memory["comparekey"]; exists {
		t.Error("Key should not exist after DISCARD")
	}

	// Now test EXEC on a new transaction
	network.Transactions[connID] = shared.Transaction{
		Commands: []shared.QueuedCommand{
			{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "comparekey"}, {Typ: "bulk", Bulk: "comparevalue"}}},
		},
	}

	result2 := Exec(connID, []shared.Value{})
	if result2.Typ != "array" || len(result2.Array) != 1 {
		t.Errorf("EXEC should return array with one result, got %v", result2)
	}

	// Verify transaction is cleared
	if _, exists := network.Transactions[connID]; exists {
		t.Error("Transaction should be cleared after EXEC")
	}

	// Verify command WAS executed
	entry, exists := server.Memory["comparekey"]
	if !exists {
		t.Error("Key should exist after EXEC")
	}
	if entry.Value != "comparevalue" {
		t.Errorf("Expected value 'comparevalue', got '%s'", entry.Value)
	}
}

func BenchmarkDiscard(b *testing.B) {
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
		Discard(connID, args)
		// Re-setup transaction after each iteration
		network.Transactions[connID] = shared.Transaction{
			Commands: []shared.QueuedCommand{
				{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "key"}, {Typ: "bulk", Bulk: "value"}}},
			},
		}
	}
}

func BenchmarkDiscardEmptyTransaction(b *testing.B) {
	clearTransactions()

	connID := "benchmark-conn"
	args := []shared.Value{}

	// Setup empty transaction
	network.Transactions[connID] = shared.Transaction{Commands: []shared.QueuedCommand{}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Discard(connID, args)
		// Re-setup empty transaction after each iteration
		network.Transactions[connID] = shared.Transaction{Commands: []shared.QueuedCommand{}}
	}
}

func BenchmarkDiscardMultipleCommands(b *testing.B) {
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
		Discard(connID, args)
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
