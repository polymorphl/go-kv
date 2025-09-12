package commands

import (
	"fmt"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func TestKeys(t *testing.T) {
	tests := []struct {
		name     string
		args     []shared.Value
		setup    func()
		expected shared.Value
	}{
		{
			name: "KEYS with no arguments",
			args: []shared.Value{},
			setup: func() {
				// Clear memory
				server.Memory = make(map[string]shared.MemoryEntry)
			},
			expected: createErrorResponse("ERR wrong number of arguments for 'keys' command"),
		},
		{
			name: "KEYS with too many arguments",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "*"},
				{Typ: "bulk", Bulk: "extra"},
			},
			setup: func() {
				// Clear memory
				server.Memory = make(map[string]shared.MemoryEntry)
			},
			expected: createErrorResponse("ERR wrong number of arguments for 'keys' command"),
		},
		{
			name: "KEYS * with no keys",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "*"},
			},
			setup: func() {
				// Clear memory
				server.Memory = make(map[string]shared.MemoryEntry)
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{}},
		},
		{
			name: "KEYS * with some keys",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "*"},
			},
			setup: func() {
				// Clear memory and add some keys
				server.Memory = make(map[string]shared.MemoryEntry)
				server.Memory["foo"] = shared.MemoryEntry{Value: "bar"}
				server.Memory["baz"] = shared.MemoryEntry{Value: "qux"}
				server.Memory["test"] = shared.MemoryEntry{Value: "value"}
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{Typ: "bulk", Bulk: "baz"},
					{Typ: "bulk", Bulk: "foo"},
					{Typ: "bulk", Bulk: "test"},
				},
			},
		},
		{
			name: "KEYS with pattern 'f*'",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "f*"},
			},
			setup: func() {
				// Clear memory and add some keys
				server.Memory = make(map[string]shared.MemoryEntry)
				server.Memory["foo"] = shared.MemoryEntry{Value: "bar"}
				server.Memory["baz"] = shared.MemoryEntry{Value: "qux"}
				server.Memory["fizz"] = shared.MemoryEntry{Value: "buzz"}
				server.Memory["test"] = shared.MemoryEntry{Value: "value"}
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{Typ: "bulk", Bulk: "fizz"},
					{Typ: "bulk", Bulk: "foo"},
				},
			},
		},
		{
			name: "KEYS with pattern 'foo?'",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "foo?"},
			},
			setup: func() {
				// Clear memory and add some keys
				server.Memory = make(map[string]shared.MemoryEntry)
				server.Memory["foo"] = shared.MemoryEntry{Value: "bar"}
				server.Memory["foo1"] = shared.MemoryEntry{Value: "bar1"}
				server.Memory["foo2"] = shared.MemoryEntry{Value: "bar2"}
				server.Memory["foobar"] = shared.MemoryEntry{Value: "baz"}
				server.Memory["test"] = shared.MemoryEntry{Value: "value"}
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{Typ: "bulk", Bulk: "foo1"},
					{Typ: "bulk", Bulk: "foo2"},
				},
			},
		},
		{
			name: "KEYS with pattern '[abc]*'",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "[abc]*"},
			},
			setup: func() {
				// Clear memory and add some keys
				server.Memory = make(map[string]shared.MemoryEntry)
				server.Memory["apple"] = shared.MemoryEntry{Value: "fruit"}
				server.Memory["banana"] = shared.MemoryEntry{Value: "fruit"}
				server.Memory["cherry"] = shared.MemoryEntry{Value: "fruit"}
				server.Memory["dog"] = shared.MemoryEntry{Value: "animal"}
				server.Memory["test"] = shared.MemoryEntry{Value: "value"}
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{Typ: "bulk", Bulk: "apple"},
					{Typ: "bulk", Bulk: "banana"},
					{Typ: "bulk", Bulk: "cherry"},
				},
			},
		},
		{
			name: "KEYS with pattern 'test[0-9]'",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "test[0-9]"},
			},
			setup: func() {
				// Clear memory and add some keys
				server.Memory = make(map[string]shared.MemoryEntry)
				server.Memory["test0"] = shared.MemoryEntry{Value: "value0"}
				server.Memory["test1"] = shared.MemoryEntry{Value: "value1"}
				server.Memory["test2"] = shared.MemoryEntry{Value: "value2"}
				server.Memory["test10"] = shared.MemoryEntry{Value: "value10"}
				server.Memory["test"] = shared.MemoryEntry{Value: "value"}
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{Typ: "bulk", Bulk: "test0"},
					{Typ: "bulk", Bulk: "test1"},
					{Typ: "bulk", Bulk: "test2"},
				},
			},
		},
		{
			name: "KEYS with no matches",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "nonexistent*"},
			},
			setup: func() {
				// Clear memory and add some keys
				server.Memory = make(map[string]shared.MemoryEntry)
				server.Memory["foo"] = shared.MemoryEntry{Value: "bar"}
				server.Memory["baz"] = shared.MemoryEntry{Value: "qux"}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{}},
		},
		{
			name: "KEYS with expired keys",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "*"},
			},
			setup: func() {
				// Clear memory and add some keys with expiration
				server.Memory = make(map[string]shared.MemoryEntry)
				server.Memory["foo"] = shared.MemoryEntry{Value: "bar", Expires: 0}     // Not expired
				server.Memory["expired"] = shared.MemoryEntry{Value: "old", Expires: 1} // Expired (timestamp 1 is in the past)
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{Typ: "bulk", Bulk: "foo"},
				},
			},
		},
		{
			name: "KEYS with invalid pattern",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "[invalid"},
			},
			setup: func() {
				// Add a key so the loop executes
				server.Memory = make(map[string]shared.MemoryEntry)
				server.Memory["test"] = shared.MemoryEntry{Value: "value"}
			},
			expected: createErrorResponse("ERR invalid pattern"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result := Keys("test-conn", tt.args)

			if tt.expected.Typ == "error" {
				if result.Typ != "error" {
					t.Errorf("Expected error response, got %s", result.Typ)
				}
				if result.Str != tt.expected.Str {
					t.Errorf("Expected error message '%s', got '%s'", tt.expected.Str, result.Str)
				}
			} else {
				if result.Typ != tt.expected.Typ {
					t.Errorf("Expected type %s, got %s", tt.expected.Typ, result.Typ)
				}
				if len(result.Array) != len(tt.expected.Array) {
					t.Errorf("Expected %d keys, got %d", len(tt.expected.Array), len(result.Array))
				}

				// Check that all expected keys are present (order doesn't matter for KEYS)
				expectedKeys := make(map[string]bool)
				for _, key := range tt.expected.Array {
					expectedKeys[key.Bulk] = true
				}

				actualKeys := make(map[string]bool)
				for _, key := range result.Array {
					actualKeys[key.Bulk] = true
				}

				for expectedKey := range expectedKeys {
					if !actualKeys[expectedKey] {
						t.Errorf("Expected key '%s' not found in result", expectedKey)
					}
				}

				for actualKey := range actualKeys {
					if !expectedKeys[actualKey] {
						t.Errorf("Unexpected key '%s' found in result", actualKey)
					}
				}
			}
		})
	}
}

func BenchmarkKeys(b *testing.B) {
	// Setup: add many keys to memory
	server.Memory = make(map[string]shared.MemoryEntry)
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		server.Memory[key] = shared.MemoryEntry{Value: "value"}
	}

	args := []shared.Value{{Typ: "bulk", Bulk: "*"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Keys("test-conn", args)
	}
}

func BenchmarkKeysPattern(b *testing.B) {
	// Setup: add many keys to memory
	server.Memory = make(map[string]shared.MemoryEntry)
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		server.Memory[key] = shared.MemoryEntry{Value: "value"}
	}

	args := []shared.Value{{Typ: "bulk", Bulk: "key[0-9]*"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Keys("test-conn", args)
	}
}
