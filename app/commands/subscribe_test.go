package commands

import (
	"fmt"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestSubscribe(t *testing.T) {
	tests := []struct {
		name          string
		connID        string
		args          []shared.Value
		expectedType  string
		expectedArray []shared.Value
		expectedCount int
		wantErr       bool
	}{
		{
			name:         "subscribe to single channel",
			connID:       "conn1",
			args:         []shared.Value{{Typ: "bulk", Bulk: "strawberry"}},
			expectedType: "array",
			expectedArray: []shared.Value{
				{Typ: "bulk", Bulk: "subscribe"},
				{Typ: "bulk", Bulk: "strawberry"},
				{Typ: "integer", Num: 1},
			},
			expectedCount: 1,
			wantErr:       false,
		},
		{
			name:   "subscribe to multiple channels",
			connID: "conn2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "channel1"},
				{Typ: "bulk", Bulk: "channel2"},
				{Typ: "bulk", Bulk: "channel3"},
			},
			expectedType: "array",
			expectedArray: []shared.Value{
				{Typ: "bulk", Bulk: "subscribe"},
				{Typ: "bulk", Bulk: "channel1"},
				{Typ: "integer", Num: 3},
			},
			expectedCount: 3,
			wantErr:       false,
		},
		{
			name:   "subscribe to same channel twice",
			connID: "conn3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "duplicate"},
				{Typ: "bulk", Bulk: "duplicate"},
			},
			expectedType: "array",
			expectedArray: []shared.Value{
				{Typ: "bulk", Bulk: "subscribe"},
				{Typ: "bulk", Bulk: "duplicate"},
				{Typ: "integer", Num: 1}, // Should only count once
			},
			expectedCount: 1,
			wantErr:       false,
		},
		{
			name:         "subscribe with empty arguments",
			connID:       "conn4",
			args:         []shared.Value{},
			expectedType: "error",
			expectedArray: []shared.Value{
				{Typ: "bulk", Bulk: "ERR wrong number of arguments for 'subscribe' command"},
			},
			expectedCount: 0,
			wantErr:       true,
		},
		{
			name:   "subscribe with unicode channel names",
			connID: "conn5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "频道1"},
				{Typ: "bulk", Bulk: "café"},
				{Typ: "bulk", Bulk: "тест"},
			},
			expectedType: "array",
			expectedArray: []shared.Value{
				{Typ: "bulk", Bulk: "subscribe"},
				{Typ: "bulk", Bulk: "频道1"},
				{Typ: "integer", Num: 3},
			},
			expectedCount: 3,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear subscriptions before each test
			shared.Subscriptions = make(map[string][]string)

			result := Subscribe(tt.connID, tt.args)

			// Check response type
			if result.Typ != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, result.Typ)
			}

			// Check array content
			if result.Typ == "array" {
				if len(result.Array) != len(tt.expectedArray) {
					t.Errorf("Expected array length %d, got %d", len(tt.expectedArray), len(result.Array))
				}

				for i, expectedItem := range tt.expectedArray {
					if i < len(result.Array) {
						actualItem := result.Array[i]
						if expectedItem.Typ != actualItem.Typ {
							t.Errorf("Expected array[%d] type %s, got %s", i, expectedItem.Typ, actualItem.Typ)
						}
						if expectedItem.Typ == "bulk" && expectedItem.Bulk != actualItem.Bulk {
							t.Errorf("Expected array[%d] bulk %s, got %s", i, expectedItem.Bulk, actualItem.Bulk)
						}
						if expectedItem.Typ == "integer" && expectedItem.Num != actualItem.Num {
							t.Errorf("Expected array[%d] integer %d, got %d", i, expectedItem.Num, actualItem.Num)
						}
					}
				}
			}

			// Check subscription count
			channels, exists := shared.SubscriptionsGet(tt.connID)
			if tt.expectedCount > 0 {
				if !exists {
					t.Errorf("Expected subscription to exist for connection %s", tt.connID)
				} else if len(channels) != tt.expectedCount {
					t.Errorf("Expected %d subscriptions, got %d: %v", tt.expectedCount, len(channels), channels)
				}
			}
		})
	}
}

// Benchmarks

// BenchmarkSubscribeSingleChannel benchmarks subscribing to a single channel
func BenchmarkSubscribeSingleChannel(b *testing.B) {
	// Clear subscriptions before benchmark
	shared.Subscriptions = make(map[string][]string)

	args := []shared.Value{{Typ: "bulk", Bulk: "test-channel"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connID := fmt.Sprintf("conn%d", i)
		Subscribe(connID, args)
	}
}

// BenchmarkSubscribeMultipleChannels benchmarks subscribing to multiple channels
func BenchmarkSubscribeMultipleChannels(b *testing.B) {
	// Clear subscriptions before benchmark
	shared.Subscriptions = make(map[string][]string)

	args := []shared.Value{
		{Typ: "bulk", Bulk: "channel1"},
		{Typ: "bulk", Bulk: "channel2"},
		{Typ: "bulk", Bulk: "channel3"},
		{Typ: "bulk", Bulk: "channel4"},
		{Typ: "bulk", Bulk: "channel5"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connID := fmt.Sprintf("conn%d", i)
		Subscribe(connID, args)
	}
}

// BenchmarkSubscribeDuplicateChannels benchmarks subscribing to duplicate channels
func BenchmarkSubscribeDuplicateChannels(b *testing.B) {
	// Clear subscriptions before benchmark
	shared.Subscriptions = make(map[string][]string)

	args := []shared.Value{
		{Typ: "bulk", Bulk: "duplicate"},
		{Typ: "bulk", Bulk: "duplicate"},
		{Typ: "bulk", Bulk: "duplicate"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connID := fmt.Sprintf("conn%d", i)
		Subscribe(connID, args)
	}
}

// BenchmarkSubscribeConcurrent benchmarks concurrent subscriptions
func BenchmarkSubscribeConcurrent(b *testing.B) {
	// Clear subscriptions before benchmark
	shared.Subscriptions = make(map[string][]string)

	args := []shared.Value{{Typ: "bulk", Bulk: "concurrent-channel"}}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			connID := fmt.Sprintf("conn%d", i)
			Subscribe(connID, args)
			i++
		}
	})
}

// BenchmarkSubscribeLargeChannelList benchmarks subscribing to many channels at once
func BenchmarkSubscribeLargeChannelList(b *testing.B) {
	// Clear subscriptions before benchmark
	shared.Subscriptions = make(map[string][]string)

	// Create a large list of channels
	args := make([]shared.Value, 100)
	for i := 0; i < 100; i++ {
		args[i] = shared.Value{Typ: "bulk", Bulk: fmt.Sprintf("channel%d", i)}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connID := fmt.Sprintf("conn%d", i)
		Subscribe(connID, args)
	}
}

// BenchmarkSubscribeUnicodeChannels benchmarks subscribing to unicode channel names
func BenchmarkSubscribeUnicodeChannels(b *testing.B) {
	// Clear subscriptions before benchmark
	shared.Subscriptions = make(map[string][]string)

	args := []shared.Value{
		{Typ: "bulk", Bulk: "频道1"},
		{Typ: "bulk", Bulk: "café"},
		{Typ: "bulk", Bulk: "тест"},
		{Typ: "bulk", Bulk: "🚀"},
		{Typ: "bulk", Bulk: "مرحبا"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connID := fmt.Sprintf("conn%d", i)
		Subscribe(connID, args)
	}
}

// BenchmarkSubscribeMemoryUsage benchmarks memory usage with many subscriptions
func BenchmarkSubscribeMemoryUsage(b *testing.B) {
	// Clear subscriptions before benchmark
	shared.Subscriptions = make(map[string][]string)

	args := []shared.Value{{Typ: "bulk", Bulk: "memory-test"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connID := fmt.Sprintf("conn%d", i)
		Subscribe(connID, args)
	}
}

// BenchmarkSubscribeGetSubscriptions benchmarks getting subscription data
func BenchmarkSubscribeGetSubscriptions(b *testing.B) {
	// Setup: create many subscriptions
	shared.Subscriptions = make(map[string][]string)

	// Create 1000 connections with subscriptions
	for i := 0; i < 1000; i++ {
		connID := fmt.Sprintf("conn%d", i)
		args := []shared.Value{
			{Typ: "bulk", Bulk: fmt.Sprintf("channel%d", i)},
			{Typ: "bulk", Bulk: fmt.Sprintf("channel%d-2", i)},
		}
		Subscribe(connID, args)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connID := fmt.Sprintf("conn%d", i%1000)
		shared.SubscriptionsGet(connID)
	}
}

// BenchmarkSubscribeMixedWorkload benchmarks a mixed workload of different subscription patterns
func BenchmarkSubscribeMixedWorkload(b *testing.B) {
	// Clear subscriptions before benchmark
	shared.Subscriptions = make(map[string][]string)

	// Different subscription patterns
	patterns := [][]shared.Value{
		{{Typ: "bulk", Bulk: "single"}},
		{
			{Typ: "bulk", Bulk: "multi1"},
			{Typ: "bulk", Bulk: "multi2"},
			{Typ: "bulk", Bulk: "multi3"},
		},
		{
			{Typ: "bulk", Bulk: "duplicate"},
			{Typ: "bulk", Bulk: "duplicate"},
		},
		{
			{Typ: "bulk", Bulk: "unicode频道"},
			{Typ: "bulk", Bulk: "café"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connID := fmt.Sprintf("conn%d", i)
		pattern := patterns[i%len(patterns)]
		Subscribe(connID, pattern)
	}
}

func TestSubscribeMultipleConnections(t *testing.T) {
	// Clear subscriptions
	shared.Subscriptions = make(map[string][]string)

	// Test multiple connections subscribing to different channels
	conn1 := "connection1"
	conn2 := "connection2"
	conn3 := "connection3"

	// Subscribe conn1 to channel1
	result1 := Subscribe(conn1, []shared.Value{{Typ: "bulk", Bulk: "channel1"}})
	if result1.Typ != "array" || len(result1.Array) != 3 {
		t.Errorf("Expected valid array response for conn1")
	}

	// Subscribe conn2 to channel1 and channel2
	result2 := Subscribe(conn2, []shared.Value{
		{Typ: "bulk", Bulk: "channel1"},
		{Typ: "bulk", Bulk: "channel2"},
	})
	if result2.Typ != "array" || len(result2.Array) != 3 {
		t.Errorf("Expected valid array response for conn2")
	}

	// Subscribe conn3 to channel2 and channel3
	result3 := Subscribe(conn3, []shared.Value{
		{Typ: "bulk", Bulk: "channel2"},
		{Typ: "bulk", Bulk: "channel3"},
	})
	if result3.Typ != "array" || len(result3.Array) != 3 {
		t.Errorf("Expected valid array response for conn3")
	}

	// Verify subscription counts
	channels1, _ := shared.SubscriptionsGet(conn1)
	channels2, _ := shared.SubscriptionsGet(conn2)
	channels3, _ := shared.SubscriptionsGet(conn3)

	if len(channels1) != 1 {
		t.Errorf("Expected conn1 to have 1 subscription, got %d", len(channels1))
	}
	if len(channels2) != 2 {
		t.Errorf("Expected conn2 to have 2 subscriptions, got %d", len(channels2))
	}
	if len(channels3) != 2 {
		t.Errorf("Expected conn3 to have 2 subscriptions, got %d", len(channels3))
	}

	// Verify specific channels
	if channels1[0] != "channel1" {
		t.Errorf("Expected conn1 to be subscribed to channel1, got %s", channels1[0])
	}
	if channels2[0] != "channel1" || channels2[1] != "channel2" {
		t.Errorf("Expected conn2 to be subscribed to channel1 and channel2, got %v", channels2)
	}
	if channels3[0] != "channel2" || channels3[1] != "channel3" {
		t.Errorf("Expected conn3 to be subscribed to channel2 and channel3, got %v", channels3)
	}
}

func TestSubscribeConcurrent(t *testing.T) {
	// Clear subscriptions
	shared.Subscriptions = make(map[string][]string)

	// Test concurrent subscriptions
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(connID string) {
			defer func() { done <- true }()

			// Each connection subscribes to multiple channels
			channels := []string{"channel1", "channel2", "channel3"}
			args := make([]shared.Value, len(channels))
			for j, channel := range channels {
				args[j] = shared.Value{Typ: "bulk", Bulk: channel}
			}

			result := Subscribe(connID, args)
			if result.Typ != "array" {
				t.Errorf("Expected array response for concurrent subscription")
			}
		}(fmt.Sprintf("conn%d", i))
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all connections have correct subscriptions
	for i := 0; i < 10; i++ {
		connID := fmt.Sprintf("conn%d", i)
		channels, exists := shared.SubscriptionsGet(connID)
		if !exists {
			t.Errorf("Expected subscription to exist for %s", connID)
		} else if len(channels) != 3 {
			t.Errorf("Expected %s to have 3 subscriptions, got %d", connID, len(channels))
		}
	}
}

func TestSubscribeEdgeCases(t *testing.T) {
	// Clear subscriptions
	shared.Subscriptions = make(map[string][]string)

	t.Run("empty channel name", func(t *testing.T) {
		result := Subscribe("conn1", []shared.Value{{Typ: "bulk", Bulk: ""}})
		if result.Typ != "array" {
			t.Errorf("Expected array response for empty channel name")
		}

		channels, _ := shared.SubscriptionsGet("conn1")
		if len(channels) != 1 || channels[0] != "" {
			t.Errorf("Expected empty channel name to be stored")
		}
	})

	t.Run("very long channel name", func(t *testing.T) {
		longChannel := string(make([]byte, 1000))
		for i := range longChannel {
			longChannel = longChannel[:i] + "a" + longChannel[i+1:]
		}

		result := Subscribe("conn2", []shared.Value{{Typ: "bulk", Bulk: longChannel}})
		if result.Typ != "array" {
			t.Errorf("Expected array response for long channel name")
		}

		channels, _ := shared.SubscriptionsGet("conn2")
		if len(channels) != 1 || channels[0] != longChannel {
			t.Errorf("Expected long channel name to be stored correctly")
		}
	})

	t.Run("special characters in channel name", func(t *testing.T) {
		specialChannel := "channel!@#$%^&*()_+-=[]{}|;':\",./<>?"
		result := Subscribe("conn3", []shared.Value{{Typ: "bulk", Bulk: specialChannel}})
		if result.Typ != "array" {
			t.Errorf("Expected array response for special characters")
		}

		channels, _ := shared.SubscriptionsGet("conn3")
		if len(channels) != 1 || channels[0] != specialChannel {
			t.Errorf("Expected special characters to be stored correctly")
		}
	})
}
