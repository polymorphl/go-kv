package commands

import (
	"fmt"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/pubsub"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestUnsubscribeCommand(t *testing.T) {
	t.Run("unsubscribe from specific channels", func(t *testing.T) {
		connID := "test-conn-8"

		// Subscribe to multiple channels
		args := []shared.Value{{Typ: "bulk", Bulk: "channel1"}, {Typ: "bulk", Bulk: "channel2"}, {Typ: "bulk", Bulk: "channel3"}}
		Subscribe(connID, args)

		// Unsubscribe from two channels
		unsubArgs := []shared.Value{{Typ: "bulk", Bulk: "channel1"}, {Typ: "bulk", Bulk: "channel3"}}
		result := Unsubscribe(connID, unsubArgs)

		// Should return unsubscribe response
		if result.Typ != "array" || len(result.Array) != 3 {
			t.Errorf("Expected array response with 3 elements, got: %v", result)
		}

		// Check response format
		if result.Array[0].Bulk != "unsubscribe" {
			t.Errorf("Expected 'unsubscribe' command, got: %s", result.Array[0].Bulk)
		}

		// Should have one remaining subscription
		channels, _ := pubsub.SubscriptionsGet(connID)
		if len(channels) != 1 || channels[0] != "channel2" {
			t.Errorf("Expected 1 remaining subscription 'channel2', got %v", channels)
		}

		// Should still be in subscribed mode
		if !pubsub.SubscribedModeGet(connID) {
			t.Error("Client should still be in subscribed mode")
		}

		// Clean up
		pubsub.SubscribedModeDelete(connID)
		pubsub.SubscriptionsDelete(connID)
	})

	t.Run("unsubscribe from all channels", func(t *testing.T) {
		connID := "test-conn-9"

		// Subscribe to channels
		args := []shared.Value{{Typ: "bulk", Bulk: "channel1"}, {Typ: "bulk", Bulk: "channel2"}}
		Subscribe(connID, args)

		// Unsubscribe from all (no args)
		result := Unsubscribe(connID, []shared.Value{})

		// Should return unsubscribe response
		if result.Typ != "array" || len(result.Array) != 3 {
			t.Errorf("Expected array response with 3 elements, got: %v", result)
		}

		// Should have no remaining subscriptions
		channels, _ := pubsub.SubscriptionsGet(connID)
		if len(channels) != 0 {
			t.Errorf("Expected no remaining subscriptions, got %v", channels)
		}

		// Should not be in subscribed mode
		if pubsub.SubscribedModeGet(connID) {
			t.Error("Client should not be in subscribed mode after unsubscribing from all")
		}

		// Clean up
		pubsub.SubscribedModeDelete(connID)
		pubsub.SubscriptionsDelete(connID)
	})

	t.Run("unsubscribe from non-existent channel", func(t *testing.T) {
		connID := "test-conn-10"

		// Subscribe to a channel
		args := []shared.Value{{Typ: "bulk", Bulk: "channel1"}}
		Subscribe(connID, args)

		// Try to unsubscribe from non-existent channel
		unsubArgs := []shared.Value{{Typ: "bulk", Bulk: "nonexistent"}}
		result := Unsubscribe(connID, unsubArgs)

		// Should return response with 0 remaining subscriptions
		if result.Typ != "array" || len(result.Array) != 3 {
			t.Errorf("Expected array response with 3 elements, got: %v", result)
		}

		// Should still have original subscription
		channels, _ := pubsub.SubscriptionsGet(connID)
		if len(channels) != 1 || channels[0] != "channel1" {
			t.Errorf("Expected 1 remaining subscription 'channel1', got %v", channels)
		}

		// Clean up
		pubsub.SubscribedModeDelete(connID)
		pubsub.SubscriptionsDelete(connID)
	})

	t.Run("unsubscribe when not subscribed", func(t *testing.T) {
		connID := "test-conn-11"

		// Try to unsubscribe without being subscribed
		result := Unsubscribe(connID, []shared.Value{})

		// Should return empty response
		if result.Typ != "array" || len(result.Array) != 3 {
			t.Errorf("Expected array response with 3 elements, got: %v", result)
		}

		if result.Array[0].Bulk != "unsubscribe" || result.Array[1].Bulk != "" || result.Array[2].Num != 0 {
			t.Errorf("Expected empty unsubscribe response, got: %v", result)
		}

		// Clean up
		pubsub.SubscribedModeDelete(connID)
		pubsub.SubscriptionsDelete(connID)
	})
}

func TestUnsubscribeSubscribedMode(t *testing.T) {
	t.Run("client exits subscribed mode after UNSUBSCRIBE all", func(t *testing.T) {
		connID := "test-conn-2"

		// Subscribe to channels
		args := []shared.Value{{Typ: "bulk", Bulk: "channel1"}, {Typ: "bulk", Bulk: "channel2"}}
		Subscribe(connID, args)

		// Should be in subscribed mode
		if !pubsub.SubscribedModeGet(connID) {
			t.Error("Client should be in subscribed mode after SUBSCRIBE")
		}

		// Unsubscribe from all channels
		Unsubscribe(connID, []shared.Value{})

		// Should no longer be in subscribed mode
		if pubsub.SubscribedModeGet(connID) {
			t.Error("Client should not be in subscribed mode after UNSUBSCRIBE all")
		}

		// Clean up
		pubsub.SubscribedModeDelete(connID)
		pubsub.SubscriptionsDelete(connID)
	})

	t.Run("client exits subscribed mode after UNSUBSCRIBE last channel", func(t *testing.T) {
		connID := "test-conn-3"

		// Subscribe to a channel
		args := []shared.Value{{Typ: "bulk", Bulk: "test-channel"}}
		Subscribe(connID, args)

		// Should be in subscribed mode
		if !pubsub.SubscribedModeGet(connID) {
			t.Error("Client should be in subscribed mode after SUBSCRIBE")
		}

		// Unsubscribe from the channel
		unsubArgs := []shared.Value{{Typ: "bulk", Bulk: "test-channel"}}
		Unsubscribe(connID, unsubArgs)

		// Should no longer be in subscribed mode
		if pubsub.SubscribedModeGet(connID) {
			t.Error("Client should not be in subscribed mode after UNSUBSCRIBE last channel")
		}

		// Clean up
		pubsub.SubscribedModeDelete(connID)
		pubsub.SubscriptionsDelete(connID)
	})

	t.Run("client stays in subscribed mode after UNSUBSCRIBE some channels", func(t *testing.T) {
		connID := "test-conn-4"

		// Subscribe to multiple channels
		args := []shared.Value{{Typ: "bulk", Bulk: "channel1"}, {Typ: "bulk", Bulk: "channel2"}}
		Subscribe(connID, args)

		// Should be in subscribed mode
		if !pubsub.SubscribedModeGet(connID) {
			t.Error("Client should be in subscribed mode after SUBSCRIBE")
		}

		// Unsubscribe from one channel
		unsubArgs := []shared.Value{{Typ: "bulk", Bulk: "channel1"}}
		Unsubscribe(connID, unsubArgs)

		// Should still be in subscribed mode
		if !pubsub.SubscribedModeGet(connID) {
			t.Error("Client should still be in subscribed mode after UNSUBSCRIBE some channels")
		}

		// Should have one remaining subscription
		channels, _ := pubsub.SubscriptionsGet(connID)
		if len(channels) != 1 || channels[0] != "channel2" {
			t.Errorf("Expected 1 remaining subscription 'channel2', got %v", channels)
		}

		// Clean up
		pubsub.SubscribedModeDelete(connID)
		pubsub.SubscriptionsDelete(connID)
	})
}

// Benchmark functions for UNSUBSCRIBE command
func BenchmarkUnsubscribeSingleChannel(b *testing.B) {
	connID := "bench-conn-1"

	// Pre-subscribe to a channel
	args := []shared.Value{{Typ: "bulk", Bulk: "test-channel"}}
	Subscribe(connID, args)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Unsubscribe and re-subscribe for next iteration
		Unsubscribe(connID, args)
		Subscribe(connID, args)
	}

	// Clean up
	pubsub.SubscribedModeDelete(connID)
	pubsub.SubscriptionsDelete(connID)
}

func BenchmarkUnsubscribeMultipleChannels(b *testing.B) {
	connID := "bench-conn-2"

	// Pre-subscribe to multiple channels
	args := []shared.Value{{Typ: "bulk", Bulk: "channel1"}, {Typ: "bulk", Bulk: "channel2"}, {Typ: "bulk", Bulk: "channel3"}}
	Subscribe(connID, args)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Unsubscribe from all and re-subscribe for next iteration
		Unsubscribe(connID, []shared.Value{})
		Subscribe(connID, args)
	}

	// Clean up
	pubsub.SubscribedModeDelete(connID)
	pubsub.SubscriptionsDelete(connID)
}

func BenchmarkUnsubscribeConcurrent(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		connID := "bench-conn-concurrent"

		// Pre-subscribe to a channel
		args := []shared.Value{{Typ: "bulk", Bulk: "test-channel"}}
		Subscribe(connID, args)

		for pb.Next() {
			// Unsubscribe and re-subscribe
			Unsubscribe(connID, args)
			Subscribe(connID, args)
		}

		// Clean up
		pubsub.SubscribedModeDelete(connID)
		pubsub.SubscriptionsDelete(connID)
	})
}

func BenchmarkUnsubscribeLargeChannelList(b *testing.B) {
	connID := "bench-conn-3"

	// Create a large list of channels
	var channels []shared.Value
	for i := 0; i < 100; i++ {
		channels = append(channels, shared.Value{Typ: "bulk", Bulk: fmt.Sprintf("channel%d", i)})
	}

	// Pre-subscribe to all channels
	Subscribe(connID, channels)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Unsubscribe from all and re-subscribe for next iteration
		Unsubscribe(connID, []shared.Value{})
		Subscribe(connID, channels)
	}

	// Clean up
	pubsub.SubscribedModeDelete(connID)
	pubsub.SubscriptionsDelete(connID)
}

func BenchmarkUnsubscribeMemoryUsage(b *testing.B) {
	connID := "bench-conn-4"

	// Pre-subscribe to a channel
	args := []shared.Value{{Typ: "bulk", Bulk: "test-channel"}}
	Subscribe(connID, args)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Unsubscribe and re-subscribe
		Unsubscribe(connID, args)
		Subscribe(connID, args)
	}

	// Clean up
	pubsub.SubscribedModeDelete(connID)
	pubsub.SubscriptionsDelete(connID)
}

func BenchmarkUnsubscribeMixedWorkload(b *testing.B) {
	connID := "bench-conn-5"

	// Pre-subscribe to multiple channels
	args := []shared.Value{{Typ: "bulk", Bulk: "channel1"}, {Typ: "bulk", Bulk: "channel2"}, {Typ: "bulk", Bulk: "channel3"}}
	Subscribe(connID, args)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Mixed workload: unsubscribe from some, then all, then re-subscribe
		Unsubscribe(connID, []shared.Value{{Typ: "bulk", Bulk: "channel1"}})
		Unsubscribe(connID, []shared.Value{})
		Subscribe(connID, args)
	}

	// Clean up
	pubsub.SubscribedModeDelete(connID)
	pubsub.SubscriptionsDelete(connID)
}
