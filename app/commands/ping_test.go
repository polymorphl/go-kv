package commands

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestPing(t *testing.T) {
	tests := []struct {
		name     string
		connID   string
		args     []shared.Value
		expected shared.Value
	}{
		{
			name:     "ping without message",
			connID:   "test-conn-1",
			args:     []shared.Value{},
			expected: shared.Value{Typ: "string", Str: "PONG"},
		},
		{
			name:   "ping with message",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "Hello"},
			},
			expected: shared.Value{Typ: "string", Str: "Hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Ping(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Ping() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Str != tt.expected.Str {
				t.Errorf("Ping() string = %v, expected %v", result.Str, tt.expected.Str)
			}
		})
	}
}

func BenchmarkPing(b *testing.B) {
	connID := "benchmark-conn"
	args := []shared.Value{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Ping(connID, args)
	}
}

func BenchmarkPingWithMessage(b *testing.B) {
	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "Hello World"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Ping(connID, args)
	}
}
