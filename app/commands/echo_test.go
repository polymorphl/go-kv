package commands

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestEcho(t *testing.T) {
	tests := []struct {
		name     string
		connID   string
		args     []shared.Value
		expected shared.Value
	}{
		{
			name:   "echo simple message",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "Hello World"},
			},
			expected: shared.Value{Typ: "string", Str: "Hello World"},
		},
		{
			name:   "echo empty message",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: ""},
			},
			expected: shared.Value{Typ: "string", Str: ""},
		},
		{
			name:   "echo special characters",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "!@#$%^&*()"},
			},
			expected: shared.Value{Typ: "string", Str: "!@#$%^&*()"},
		},
		{
			name:   "echo unicode message",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "Hello 世界 🌍"},
			},
			expected: shared.Value{Typ: "string", Str: "Hello 世界 🌍"},
		},
		{
			name:   "echo long message",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "This is a very long message that contains multiple words and should be echoed back exactly as received"},
			},
			expected: shared.Value{Typ: "string", Str: "This is a very long message that contains multiple words and should be echoed back exactly as received"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Echo(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Echo() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Str != tt.expected.Str {
				t.Errorf("Echo() string = %v, expected %v", result.Str, tt.expected.Str)
			}
		})
	}
}

func BenchmarkEcho(b *testing.B) {
	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "Hello World"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Echo(connID, args)
	}
}

func BenchmarkEchoLongMessage(b *testing.B) {
	connID := "benchmark-conn"
	longMessage := "This is a very long message that contains multiple words and should be echoed back exactly as received. " +
		"It includes various characters and symbols to test the performance of the echo command with realistic data. " +
		"The message is repeated multiple times to simulate real-world usage patterns."

	args := []shared.Value{
		{Typ: "bulk", Bulk: longMessage},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Echo(connID, args)
	}
}

func BenchmarkEchoUnicode(b *testing.B) {
	connID := "benchmark-conn"
	unicodeMessage := "Hello 世界 🌍 こんにちは مرحبا Здравствуй 你好"

	args := []shared.Value{
		{Typ: "bulk", Bulk: unicodeMessage},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Echo(connID, args)
	}
}

func BenchmarkEchoEmpty(b *testing.B) {
	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: ""},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Echo(connID, args)
	}
}
