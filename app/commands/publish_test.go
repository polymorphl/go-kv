package commands

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestPublish(t *testing.T) {
	t.Run("publish with insufficient arguments", func(t *testing.T) {
		result := Publish("test-conn", []shared.Value{})
		if result.Typ != "error" {
			t.Errorf("Expected error for insufficient arguments, got: %v", result)
		}
		if result.Str != "ERR wrong number of arguments for 'publish' command" {
			t.Errorf("Expected specific error message, got: %s", result.Str)
		}
	})

	t.Run("publish with single argument", func(t *testing.T) {
		result := Publish("test-conn", []shared.Value{{Typ: "bulk", Bulk: "channel"}})
		if result.Typ != "error" {
			t.Errorf("Expected error for single argument, got: %v", result)
		}
	})

	t.Run("publish to channel with no subscribers", func(t *testing.T) {
		// Clean up any existing subscriptions
		shared.SubscriptionsDelete("test-conn-1")
		shared.SubscriptionsDelete("test-conn-2")

		result := Publish("test-conn", []shared.Value{
			{Typ: "bulk", Bulk: "test-channel"},
			{Typ: "bulk", Bulk: "test-message"},
		})

		expected := shared.Value{Typ: "integer", Num: 0}
		if result.Typ != expected.Typ || result.Num != expected.Num {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("publish to channel with one subscriber", func(t *testing.T) {
		connID := "test-conn-1"
		channel := "test-channel-1"

		// Subscribe to channel
		Subscribe(connID, []shared.Value{{Typ: "bulk", Bulk: channel}})

		// Publish to channel
		result := Publish("test-conn", []shared.Value{
			{Typ: "bulk", Bulk: channel},
			{Typ: "bulk", Bulk: "test-message"},
		})

		expected := shared.Value{Typ: "integer", Num: 1}
		if result.Typ != expected.Typ || result.Num != expected.Num {
			t.Errorf("Expected %v, got %v", expected, result)
		}

		// Clean up
		shared.SubscribedModeDelete(connID)
		shared.SubscriptionsDelete(connID)
	})

	t.Run("publish to channel with multiple subscribers", func(t *testing.T) {
		connID1 := "test-conn-1"
		connID2 := "test-conn-2"
		connID3 := "test-conn-3"
		channel := "test-channel-2"

		// Subscribe multiple clients to the same channel
		Subscribe(connID1, []shared.Value{{Typ: "bulk", Bulk: channel}})
		Subscribe(connID2, []shared.Value{{Typ: "bulk", Bulk: channel}})
		Subscribe(connID3, []shared.Value{{Typ: "bulk", Bulk: channel}})

		// Publish to channel
		result := Publish("test-conn", []shared.Value{
			{Typ: "bulk", Bulk: channel},
			{Typ: "bulk", Bulk: "test-message"},
		})

		expected := shared.Value{Typ: "integer", Num: 3}
		if result.Typ != expected.Typ || result.Num != expected.Num {
			t.Errorf("Expected %v, got %v", expected, result)
		}

		// Clean up
		shared.SubscribedModeDelete(connID1)
		shared.SubscribedModeDelete(connID2)
		shared.SubscribedModeDelete(connID3)
		shared.SubscriptionsDelete(connID1)
		shared.SubscriptionsDelete(connID2)
		shared.SubscriptionsDelete(connID3)
	})

	t.Run("publish to channel with mixed subscriptions", func(t *testing.T) {
		connID1 := "test-conn-1"
		connID2 := "test-conn-2"
		connID3 := "test-conn-3"
		channel1 := "test-channel-3"
		channel2 := "test-channel-4"

		// Subscribe clients to different channels
		Subscribe(connID1, []shared.Value{{Typ: "bulk", Bulk: channel1}})
		Subscribe(connID2, []shared.Value{{Typ: "bulk", Bulk: channel1}})
		Subscribe(connID3, []shared.Value{{Typ: "bulk", Bulk: channel2}})

		// Publish to channel1 (should have 2 subscribers)
		result1 := Publish("test-conn", []shared.Value{
			{Typ: "bulk", Bulk: channel1},
			{Typ: "bulk", Bulk: "test-message-1"},
		})

		expected1 := shared.Value{Typ: "integer", Num: 2}
		if result1.Typ != expected1.Typ || result1.Num != expected1.Num {
			t.Errorf("Expected %v for channel1, got %v", expected1, result1)
		}

		// Publish to channel2 (should have 1 subscriber)
		result2 := Publish("test-conn", []shared.Value{
			{Typ: "bulk", Bulk: channel2},
			{Typ: "bulk", Bulk: "test-message-2"},
		})

		expected2 := shared.Value{Typ: "integer", Num: 1}
		if result2.Typ != expected2.Typ || result2.Num != expected2.Num {
			t.Errorf("Expected %v for channel2, got %v", expected2, result2)
		}

		// Clean up
		shared.SubscribedModeDelete(connID1)
		shared.SubscribedModeDelete(connID2)
		shared.SubscribedModeDelete(connID3)
		shared.SubscriptionsDelete(connID1)
		shared.SubscriptionsDelete(connID2)
		shared.SubscriptionsDelete(connID3)
	})

	t.Run("publish to channel with client subscribed to multiple channels", func(t *testing.T) {
		connID1 := "test-conn-1"
		connID2 := "test-conn-2"
		channel1 := "test-channel-5"
		channel2 := "test-channel-6"

		// Subscribe client1 to both channels
		Subscribe(connID1, []shared.Value{
			{Typ: "bulk", Bulk: channel1},
			{Typ: "bulk", Bulk: channel2},
		})
		// Subscribe client2 to only channel1
		Subscribe(connID2, []shared.Value{{Typ: "bulk", Bulk: channel1}})

		// Publish to channel1 (should have 2 subscribers)
		result1 := Publish("test-conn", []shared.Value{
			{Typ: "bulk", Bulk: channel1},
			{Typ: "bulk", Bulk: "test-message-1"},
		})

		expected1 := shared.Value{Typ: "integer", Num: 2}
		if result1.Typ != expected1.Typ || result1.Num != expected1.Num {
			t.Errorf("Expected %v for channel1, got %v", expected1, result1)
		}

		// Publish to channel2 (should have 1 subscriber)
		result2 := Publish("test-conn", []shared.Value{
			{Typ: "bulk", Bulk: channel2},
			{Typ: "bulk", Bulk: "test-message-2"},
		})

		expected2 := shared.Value{Typ: "integer", Num: 1}
		if result2.Typ != expected2.Typ || result2.Num != expected2.Num {
			t.Errorf("Expected %v for channel2, got %v", expected2, result2)
		}

		// Clean up
		shared.SubscribedModeDelete(connID1)
		shared.SubscribedModeDelete(connID2)
		shared.SubscriptionsDelete(connID1)
		shared.SubscriptionsDelete(connID2)
	})

	t.Run("publish with unicode channel and message", func(t *testing.T) {
		connID := "test-conn-1"
		channel := "测试频道"
		message := "测试消息"

		// Subscribe to channel
		Subscribe(connID, []shared.Value{{Typ: "bulk", Bulk: channel}})

		// Publish to channel
		result := Publish("test-conn", []shared.Value{
			{Typ: "bulk", Bulk: channel},
			{Typ: "bulk", Bulk: message},
		})

		expected := shared.Value{Typ: "integer", Num: 1}
		if result.Typ != expected.Typ || result.Num != expected.Num {
			t.Errorf("Expected %v, got %v", expected, result)
		}

		// Clean up
		shared.SubscribedModeDelete(connID)
		shared.SubscriptionsDelete(connID)
	})

	t.Run("publish after unsubscribe", func(t *testing.T) {
		connID1 := "test-conn-1"
		connID2 := "test-conn-2"
		channel := "test-channel-7"

		// Subscribe both clients to channel
		Subscribe(connID1, []shared.Value{{Typ: "bulk", Bulk: channel}})
		Subscribe(connID2, []shared.Value{{Typ: "bulk", Bulk: channel}})

		// Publish to channel (should have 2 subscribers)
		result1 := Publish("test-conn", []shared.Value{
			{Typ: "bulk", Bulk: channel},
			{Typ: "bulk", Bulk: "test-message-1"},
		})

		expected1 := shared.Value{Typ: "integer", Num: 2}
		if result1.Typ != expected1.Typ || result1.Num != expected1.Num {
			t.Errorf("Expected %v before unsubscribe, got %v", expected1, result1)
		}

		// Unsubscribe one client
		Unsubscribe(connID1, []shared.Value{{Typ: "bulk", Bulk: channel}})

		// Publish to channel (should have 1 subscriber)
		result2 := Publish("test-conn", []shared.Value{
			{Typ: "bulk", Bulk: channel},
			{Typ: "bulk", Bulk: "test-message-2"},
		})

		expected2 := shared.Value{Typ: "integer", Num: 1}
		if result2.Typ != expected2.Typ || result2.Num != expected2.Num {
			t.Errorf("Expected %v after unsubscribe, got %v", expected2, result2)
		}

		// Clean up
		shared.SubscribedModeDelete(connID1)
		shared.SubscribedModeDelete(connID2)
		shared.SubscriptionsDelete(connID1)
		shared.SubscriptionsDelete(connID2)
	})
}

func BenchmarkPublish(b *testing.B) {
	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "test-channel"},
		{Typ: "bulk", Bulk: "test-message"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Publish(connID, args)
	}
}
