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

func TestPingInSubscribedMode(t *testing.T) {
	t.Run("ping without message in subscribed mode", func(t *testing.T) {
		connID := "test-conn-subscribed-1"

		// Subscribe to enter subscribed mode
		args := []shared.Value{{Typ: "bulk", Bulk: "test-channel"}}
		Subscribe(connID, args)

		// Test PING in subscribed mode
		result := Ping(connID, []shared.Value{})

		// Should return array with "pong" and empty bulk string
		expected := shared.Value{
			Typ: "array",
			Array: []shared.Value{
				{Typ: "bulk", Bulk: "pong"},
				{Typ: "bulk", Bulk: ""},
			},
		}

		if result.Typ != expected.Typ {
			t.Errorf("Ping() in subscribed mode type = %v, expected %v", result.Typ, expected.Typ)
		}

		if len(result.Array) != 2 {
			t.Errorf("Ping() in subscribed mode array length = %v, expected 2", len(result.Array))
		}

		if result.Array[0].Typ != "bulk" || result.Array[0].Bulk != "pong" {
			t.Errorf("Ping() in subscribed mode first element = %v, expected bulk 'pong'", result.Array[0])
		}

		if result.Array[1].Typ != "bulk" || result.Array[1].Bulk != "" {
			t.Errorf("Ping() in subscribed mode second element = %v, expected empty bulk string", result.Array[1])
		}

		// Clean up
		shared.SubscribedModeDelete(connID)
		shared.SubscriptionsDelete(connID)
	})

	t.Run("ping with message in subscribed mode", func(t *testing.T) {
		connID := "test-conn-subscribed-2"

		// Subscribe to enter subscribed mode
		args := []shared.Value{{Typ: "bulk", Bulk: "test-channel"}}
		Subscribe(connID, args)

		// Test PING with message in subscribed mode
		result := Ping(connID, []shared.Value{{Typ: "bulk", Bulk: "Hello"}})

		// Should still return array with "pong" and empty bulk string (ignores message in subscribed mode)
		expected := shared.Value{
			Typ: "array",
			Array: []shared.Value{
				{Typ: "bulk", Bulk: "pong"},
				{Typ: "bulk", Bulk: ""},
			},
		}

		if result.Typ != expected.Typ {
			t.Errorf("Ping() with message in subscribed mode type = %v, expected %v", result.Typ, expected.Typ)
		}

		if len(result.Array) != 2 {
			t.Errorf("Ping() with message in subscribed mode array length = %v, expected 2", len(result.Array))
		}

		if result.Array[0].Typ != "bulk" || result.Array[0].Bulk != "pong" {
			t.Errorf("Ping() with message in subscribed mode first element = %v, expected bulk 'pong'", result.Array[0])
		}

		if result.Array[1].Typ != "bulk" || result.Array[1].Bulk != "" {
			t.Errorf("Ping() with message in subscribed mode second element = %v, expected empty bulk string", result.Array[1])
		}

		// Clean up
		shared.SubscribedModeDelete(connID)
		shared.SubscriptionsDelete(connID)
	})

	t.Run("ping works normally when not in subscribed mode", func(t *testing.T) {
		connID := "test-conn-normal"

		// Should not be in subscribed mode initially
		if shared.SubscribedModeGet(connID) {
			t.Error("Client should not be in subscribed mode initially")
		}

		// Test PING without message
		result := Ping(connID, []shared.Value{})
		expected := shared.Value{Typ: "string", Str: "PONG"}

		if result.Typ != expected.Typ || result.Str != expected.Str {
			t.Errorf("Ping() when not in subscribed mode = %v, expected %v", result, expected)
		}

		// Test PING with message
		result = Ping(connID, []shared.Value{{Typ: "bulk", Bulk: "Hello"}})
		expected = shared.Value{Typ: "string", Str: "Hello"}

		if result.Typ != expected.Typ || result.Str != expected.Str {
			t.Errorf("Ping() with message when not in subscribed mode = %v, expected %v", result, expected)
		}
	})
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
