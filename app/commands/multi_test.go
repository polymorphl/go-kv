package commands

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestMulti(t *testing.T) {
	// Clear transactions before each test
	clearTransactions()

	tests := []struct {
		name     string
		connID   string
		args     []shared.Value
		expected shared.Value
		verify   func() // Function to verify the result
	}{
		{
			name:     "multi without arguments",
			connID:   "test-conn-1",
			args:     []shared.Value{},
			expected: shared.Value{Typ: "string", Str: "OK"},
			verify: func() {
				transaction, exists := shared.Transactions["test-conn-1"]
				if !exists {
					t.Error("Transaction should exist after MULTI")
				}
				if transaction.Commands == nil {
					t.Error("Transaction Commands should be initialized")
				}
				if len(transaction.Commands) != 0 {
					t.Error("Transaction should start with empty commands")
				}
			},
		},
		{
			name:   "multi with arguments should fail",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "extra"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'multi' command"},
			verify: func() {
				// Transaction should not exist after error
				if _, exists := shared.Transactions["test-conn-2"]; exists {
					t.Error("Transaction should not exist after error")
				}
			},
		},
		{
			name:   "multi with multiple arguments should fail",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "arg1"},
				{Typ: "bulk", Bulk: "arg2"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'multi' command"},
			verify: func() {
				// Transaction should not exist after error
				if _, exists := shared.Transactions["test-conn-3"]; exists {
					t.Error("Transaction should not exist after error")
				}
			},
		},
		{
			name:     "multi overwrites existing transaction",
			connID:   "test-conn-4",
			args:     []shared.Value{},
			expected: shared.Value{Typ: "string", Str: "OK"},
			verify: func() {
				transaction, exists := shared.Transactions["test-conn-4"]
				if !exists {
					t.Error("Transaction should exist after MULTI")
				}
				if len(transaction.Commands) != 0 {
					t.Error("Transaction should start with empty commands")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTransactions()

			// Set up initial transaction for overwrite test
			if tt.name == "multi overwrites existing transaction" {
				shared.Transactions["test-conn-4"] = shared.Transaction{
					Commands: []shared.QueuedCommand{
						{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "key"}, {Typ: "bulk", Bulk: "value"}}},
					},
				}
			}

			result := Multi(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Multi() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Str != tt.expected.Str {
				t.Errorf("Multi() string = %v, expected %v", result.Str, tt.expected.Str)
			}

			tt.verify()
		})
	}
}

func TestMultiTransactionState(t *testing.T) {
	clearTransactions()

	connID := "test-conn-state"

	// Test that MULTI creates a transaction
	result := Multi(connID, []shared.Value{})
	if result.Typ != "string" || result.Str != "OK" {
		t.Errorf("MULTI should return OK, got %v", result)
	}

	// Verify transaction exists
	transaction, exists := shared.Transactions[connID]
	if !exists {
		t.Error("Transaction should exist after MULTI")
	}

	// Verify transaction is properly initialized
	if transaction.Commands == nil {
		t.Error("Transaction Commands should be initialized")
	}

	if len(transaction.Commands) != 0 {
		t.Error("Transaction should start with empty commands")
	}
}

func TestMultiMultipleConnections(t *testing.T) {
	clearTransactions()

	conn1 := "connection-1"
	conn2 := "connection-2"

	// Start transaction on first connection
	result1 := Multi(conn1, []shared.Value{})
	if result1.Typ != "string" || result1.Str != "OK" {
		t.Errorf("MULTI on conn1 should return OK, got %v", result1)
	}

	// Start transaction on second connection
	result2 := Multi(conn2, []shared.Value{})
	if result2.Typ != "string" || result2.Str != "OK" {
		t.Errorf("MULTI on conn2 should return OK, got %v", result2)
	}

	// Verify both transactions exist independently
	if _, exists := shared.Transactions[conn1]; !exists {
		t.Error("Transaction for conn1 should exist")
	}

	if _, exists := shared.Transactions[conn2]; !exists {
		t.Error("Transaction for conn2 should exist")
	}

	// Verify transactions are independent
	if len(shared.Transactions[conn1].Commands) != 0 {
		t.Error("Transaction for conn1 should start empty")
	}

	if len(shared.Transactions[conn2].Commands) != 0 {
		t.Error("Transaction for conn2 should start empty")
	}
}

func BenchmarkMulti(b *testing.B) {
	clearTransactions()

	connID := "benchmark-conn"
	args := []shared.Value{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Multi(connID, args)
		// Clear transaction after each iteration to avoid accumulation
		delete(shared.Transactions, connID)
	}
}

func BenchmarkMultiWithExistingTransaction(b *testing.B) {
	clearTransactions()

	connID := "benchmark-conn"
	args := []shared.Value{}

	// Pre-create a transaction
	shared.Transactions[connID] = shared.Transaction{
		Commands: []shared.QueuedCommand{
			{Command: "SET", Args: []shared.Value{{Typ: "bulk", Bulk: "key"}, {Typ: "bulk", Bulk: "value"}}},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Multi(connID, args)
	}
}
