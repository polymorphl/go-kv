package commands

import (
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/network"
	"github.com/codecrafters-io/redis-starter-go/app/pubsub"
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
		pubsub.SubscriptionsDelete("test-conn-1")
		pubsub.SubscriptionsDelete("test-conn-2")

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
		pubsub.SubscribedModeDelete(connID)
		pubsub.SubscriptionsDelete(connID)
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
		pubsub.SubscribedModeDelete(connID1)
		pubsub.SubscribedModeDelete(connID2)
		pubsub.SubscribedModeDelete(connID3)
		pubsub.SubscriptionsDelete(connID1)
		pubsub.SubscriptionsDelete(connID2)
		pubsub.SubscriptionsDelete(connID3)
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
		pubsub.SubscribedModeDelete(connID1)
		pubsub.SubscribedModeDelete(connID2)
		pubsub.SubscribedModeDelete(connID3)
		pubsub.SubscriptionsDelete(connID1)
		pubsub.SubscriptionsDelete(connID2)
		pubsub.SubscriptionsDelete(connID3)
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
		pubsub.SubscribedModeDelete(connID1)
		pubsub.SubscribedModeDelete(connID2)
		pubsub.SubscriptionsDelete(connID1)
		pubsub.SubscriptionsDelete(connID2)
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
		pubsub.SubscribedModeDelete(connID)
		pubsub.SubscriptionsDelete(connID)
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
		pubsub.SubscribedModeDelete(connID1)
		pubsub.SubscribedModeDelete(connID2)
		pubsub.SubscriptionsDelete(connID1)
		pubsub.SubscriptionsDelete(connID2)
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

// MockConnection represents a mock network connection for testing
type MockConnection struct {
	*bytes.Buffer
	remoteAddr string
	localAddr  string
}

func (m *MockConnection) RemoteAddr() net.Addr {
	return &mockAddr{addr: m.remoteAddr}
}

func (m *MockConnection) LocalAddr() net.Addr {
	return &mockAddr{addr: m.localAddr}
}

func (m *MockConnection) Close() error {
	return nil
}

func (m *MockConnection) Read(b []byte) (n int, err error) {
	return m.Buffer.Read(b)
}

func (m *MockConnection) Write(b []byte) (n int, err error) {
	return m.Buffer.Write(b)
}

func (m *MockConnection) SetDeadline(t time.Time) error {
	return nil
}

func (m *MockConnection) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *MockConnection) SetWriteDeadline(t time.Time) error {
	return nil
}

type mockAddr struct {
	addr string
}

func (m *mockAddr) Network() string { return "tcp" }
func (m *mockAddr) String() string  { return m.addr }

func TestPublishMessageDelivery(t *testing.T) {
	t.Run("publish delivers message to single subscriber", func(t *testing.T) {
		// Clean up any existing subscriptions
		pubsub.SetSubscriptionsMap(make(map[string][]string))
		pubsub.SetSubscribedModeMap(make(map[string]bool))

		// Create mock connections
		conn1 := &MockConnection{
			Buffer:     &bytes.Buffer{},
			remoteAddr: "127.0.0.1:12345",
			localAddr:  "127.0.0.1:6379",
		}
		conn2 := &MockConnection{
			Buffer:     &bytes.Buffer{},
			remoteAddr: "127.0.0.1:12346",
			localAddr:  "127.0.0.1:6379",
		}

		// Register connections
		network.ConnectionsSet("127.0.0.1:12345", conn1)
		network.ConnectionsSet("127.0.0.1:12346", conn2)

		// Subscribe client1 to channel
		Subscribe("127.0.0.1:12345", []shared.Value{{Typ: "bulk", Bulk: "test-channel"}})

		// Publish message
		result := Publish("127.0.0.1:12347", []shared.Value{
			{Typ: "bulk", Bulk: "test-channel"},
			{Typ: "bulk", Bulk: "hello world"},
		})

		// Verify result
		if result.Typ != "integer" || result.Num != 1 {
			t.Errorf("Expected integer 1, got %v", result)
		}

		// Verify message was delivered to subscriber
		expectedMessage := "*3\r\n$7\r\nmessage\r\n$12\r\ntest-channel\r\n$11\r\nhello world\r\n"
		actualMessage := conn1.String()
		if actualMessage != expectedMessage {
			t.Errorf("Expected message %q, got %q", expectedMessage, actualMessage)
		}

		// Verify no message was sent to non-subscriber
		if conn2.Len() != 0 {
			t.Errorf("Expected no message to non-subscriber, got %q", conn2.String())
		}

		// Clean up
		network.ConnectionsDelete("127.0.0.1:12345")
		network.ConnectionsDelete("127.0.0.1:12346")
		pubsub.SubscriptionsDelete("127.0.0.1:12345")
		pubsub.SubscribedModeDelete("127.0.0.1:12345")
	})

	t.Run("publish delivers message to multiple subscribers", func(t *testing.T) {
		// Clean up any existing subscriptions
		pubsub.SetSubscriptionsMap(make(map[string][]string))
		pubsub.SetSubscribedModeMap(make(map[string]bool))

		// Create mock connections
		conn1 := &MockConnection{
			Buffer:     &bytes.Buffer{},
			remoteAddr: "127.0.0.1:12345",
			localAddr:  "127.0.0.1:6379",
		}
		conn2 := &MockConnection{
			Buffer:     &bytes.Buffer{},
			remoteAddr: "127.0.0.1:12346",
			localAddr:  "127.0.0.1:6379",
		}
		conn3 := &MockConnection{
			Buffer:     &bytes.Buffer{},
			remoteAddr: "127.0.0.1:12347",
			localAddr:  "127.0.0.1:6379",
		}

		// Register connections
		network.ConnectionsSet("127.0.0.1:12345", conn1)
		network.ConnectionsSet("127.0.0.1:12346", conn2)
		network.ConnectionsSet("127.0.0.1:12347", conn3)

		// Subscribe all clients to the same channel
		Subscribe("127.0.0.1:12345", []shared.Value{{Typ: "bulk", Bulk: "test-channel"}})
		Subscribe("127.0.0.1:12346", []shared.Value{{Typ: "bulk", Bulk: "test-channel"}})
		Subscribe("127.0.0.1:12347", []shared.Value{{Typ: "bulk", Bulk: "test-channel"}})

		// Publish message
		result := Publish("127.0.0.1:12348", []shared.Value{
			{Typ: "bulk", Bulk: "test-channel"},
			{Typ: "bulk", Bulk: "broadcast message"},
		})

		// Verify result
		if result.Typ != "integer" || result.Num != 3 {
			t.Errorf("Expected integer 3, got %v", result)
		}

		// Verify message was delivered to all subscribers
		expectedMessage := "*3\r\n$7\r\nmessage\r\n$12\r\ntest-channel\r\n$17\r\nbroadcast message\r\n"
		for i, conn := range []*MockConnection{conn1, conn2, conn3} {
			actualMessage := conn.String()
			if actualMessage != expectedMessage {
				t.Errorf("Expected message %q for conn%d, got %q", expectedMessage, i+1, actualMessage)
			}
		}

		// Clean up
		network.ConnectionsDelete("127.0.0.1:12345")
		network.ConnectionsDelete("127.0.0.1:12346")
		network.ConnectionsDelete("127.0.0.1:12347")
		pubsub.SubscriptionsDelete("127.0.0.1:12345")
		pubsub.SubscriptionsDelete("127.0.0.1:12346")
		pubsub.SubscriptionsDelete("127.0.0.1:12347")
		pubsub.SubscribedModeDelete("127.0.0.1:12345")
		pubsub.SubscribedModeDelete("127.0.0.1:12346")
		pubsub.SubscribedModeDelete("127.0.0.1:12347")
	})

	t.Run("publish delivers message only to subscribers of specific channel", func(t *testing.T) {
		// Clean up any existing subscriptions
		pubsub.SetSubscriptionsMap(make(map[string][]string))
		pubsub.SetSubscribedModeMap(make(map[string]bool))

		// Create mock connections
		conn1 := &MockConnection{
			Buffer:     &bytes.Buffer{},
			remoteAddr: "127.0.0.1:12345",
			localAddr:  "127.0.0.1:6379",
		}
		conn2 := &MockConnection{
			Buffer:     &bytes.Buffer{},
			remoteAddr: "127.0.0.1:12346",
			localAddr:  "127.0.0.1:6379",
		}
		conn3 := &MockConnection{
			Buffer:     &bytes.Buffer{},
			remoteAddr: "127.0.0.1:12347",
			localAddr:  "127.0.0.1:6379",
		}

		// Register connections
		network.ConnectionsSet("127.0.0.1:12345", conn1)
		network.ConnectionsSet("127.0.0.1:12346", conn2)
		network.ConnectionsSet("127.0.0.1:12347", conn3)

		// Subscribe clients to different channels
		Subscribe("127.0.0.1:12345", []shared.Value{{Typ: "bulk", Bulk: "channel-1"}})
		Subscribe("127.0.0.1:12346", []shared.Value{{Typ: "bulk", Bulk: "channel-2"}})
		Subscribe("127.0.0.1:12347", []shared.Value{{Typ: "bulk", Bulk: "channel-1"}})

		// Publish message to channel-1
		result := Publish("127.0.0.1:12348", []shared.Value{
			{Typ: "bulk", Bulk: "channel-1"},
			{Typ: "bulk", Bulk: "channel-1 message"},
		})

		// Verify result
		if result.Typ != "integer" || result.Num != 2 {
			t.Errorf("Expected integer 2, got %v", result)
		}

		// Verify message was delivered only to subscribers of channel-1
		expectedMessage := "*3\r\n$7\r\nmessage\r\n$9\r\nchannel-1\r\n$17\r\nchannel-1 message\r\n"
		if conn1.String() != expectedMessage {
			t.Errorf("Expected message %q for conn1, got %q", expectedMessage, conn1.String())
		}
		if conn3.String() != expectedMessage {
			t.Errorf("Expected message %q for conn3, got %q", expectedMessage, conn3.String())
		}
		if conn2.Len() != 0 {
			t.Errorf("Expected no message for conn2 (subscribed to different channel), got %q", conn2.String())
		}

		// Clean up
		network.ConnectionsDelete("127.0.0.1:12345")
		network.ConnectionsDelete("127.0.0.1:12346")
		network.ConnectionsDelete("127.0.0.1:12347")
		pubsub.SubscriptionsDelete("127.0.0.1:12345")
		pubsub.SubscriptionsDelete("127.0.0.1:12346")
		pubsub.SubscriptionsDelete("127.0.0.1:12347")
		pubsub.SubscribedModeDelete("127.0.0.1:12345")
		pubsub.SubscribedModeDelete("127.0.0.1:12346")
		pubsub.SubscribedModeDelete("127.0.0.1:12347")
	})

	t.Run("publish handles failed connections gracefully", func(t *testing.T) {
		// Clean up any existing subscriptions
		pubsub.SetSubscriptionsMap(make(map[string][]string))
		pubsub.SetSubscribedModeMap(make(map[string]bool))

		// Create mock connections
		conn1 := &MockConnection{
			Buffer:     &bytes.Buffer{},
			remoteAddr: "127.0.0.1:12345",
			localAddr:  "127.0.0.1:6379",
		}
		// conn2 will be a failed connection (not registered in Connections map)

		// Register only one connection
		network.ConnectionsSet("127.0.0.1:12345", conn1)

		// Subscribe both clients (one with connection, one without)
		Subscribe("127.0.0.1:12345", []shared.Value{{Typ: "bulk", Bulk: "test-channel"}})
		Subscribe("127.0.0.1:12346", []shared.Value{{Typ: "bulk", Bulk: "test-channel"}})

		// Publish message
		result := Publish("127.0.0.1:12347", []shared.Value{
			{Typ: "bulk", Bulk: "test-channel"},
			{Typ: "bulk", Bulk: "test message"},
		})

		// Should return 2 (both subscribers counted, but only 1 actually delivered)
		if result.Typ != "integer" || result.Num != 2 {
			t.Errorf("Expected integer 2, got %v", result)
		}

		// Verify message was delivered to the connected subscriber
		expectedMessage := "*3\r\n$7\r\nmessage\r\n$12\r\ntest-channel\r\n$12\r\ntest message\r\n"
		if conn1.String() != expectedMessage {
			t.Errorf("Expected message %q for conn1, got %q", expectedMessage, conn1.String())
		}

		// Clean up
		network.ConnectionsDelete("127.0.0.1:12345")
		pubsub.SubscriptionsDelete("127.0.0.1:12345")
		pubsub.SubscriptionsDelete("127.0.0.1:12346")
		pubsub.SubscribedModeDelete("127.0.0.1:12345")
		pubsub.SubscribedModeDelete("127.0.0.1:12346")
	})

	t.Run("publish with unicode channel and message", func(t *testing.T) {
		// Clean up any existing subscriptions
		pubsub.SetSubscriptionsMap(make(map[string][]string))
		pubsub.SetSubscribedModeMap(make(map[string]bool))

		// Create mock connection
		conn1 := &MockConnection{
			Buffer:     &bytes.Buffer{},
			remoteAddr: "127.0.0.1:12345",
			localAddr:  "127.0.0.1:6379",
		}

		// Register connection
		network.ConnectionsSet("127.0.0.1:12345", conn1)

		// Subscribe to unicode channel
		channel := "测试频道"
		message := "测试消息"
		Subscribe("127.0.0.1:12345", []shared.Value{{Typ: "bulk", Bulk: channel}})

		// Publish unicode message
		result := Publish("127.0.0.1:12346", []shared.Value{
			{Typ: "bulk", Bulk: channel},
			{Typ: "bulk", Bulk: message},
		})

		// Verify result
		if result.Typ != "integer" || result.Num != 1 {
			t.Errorf("Expected integer 1, got %v", result)
		}

		// Verify message was delivered with correct unicode encoding
		expectedMessage := fmt.Sprintf("*3\r\n$7\r\nmessage\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
			len(channel), channel, len(message), message)
		actualMessage := conn1.String()
		if actualMessage != expectedMessage {
			t.Errorf("Expected message %q, got %q", expectedMessage, actualMessage)
		}

		// Clean up
		network.ConnectionsDelete("127.0.0.1:12345")
		pubsub.SubscriptionsDelete("127.0.0.1:12345")
		pubsub.SubscribedModeDelete("127.0.0.1:12345")
	})
}

func TestPublishMessageFormat(t *testing.T) {
	t.Run("message format matches Redis protocol", func(t *testing.T) {
		// Clean up any existing subscriptions
		pubsub.SetSubscriptionsMap(make(map[string][]string))
		pubsub.SetSubscribedModeMap(make(map[string]bool))

		// Create mock connection
		conn1 := &MockConnection{
			Buffer:     &bytes.Buffer{},
			remoteAddr: "127.0.0.1:12345",
			localAddr:  "127.0.0.1:6379",
		}

		// Register connection
		network.ConnectionsSet("127.0.0.1:12345", conn1)

		// Subscribe to channel
		Subscribe("127.0.0.1:12345", []shared.Value{{Typ: "bulk", Bulk: "channel_1"}})

		// Publish message
		Publish("127.0.0.1:12346", []shared.Value{
			{Typ: "bulk", Bulk: "channel_1"},
			{Typ: "bulk", Bulk: "hello"},
		})

		// Verify the exact RESP format
		expected := "*3\r\n$7\r\nmessage\r\n$9\r\nchannel_1\r\n$5\r\nhello\r\n"
		actual := conn1.String()
		if actual != expected {
			t.Errorf("Expected RESP format %q, got %q", expected, actual)
		}

		// Clean up
		network.ConnectionsDelete("127.0.0.1:12345")
		pubsub.SubscriptionsDelete("127.0.0.1:12345")
		pubsub.SubscribedModeDelete("127.0.0.1:12345")
	})
}

func TestPublishConcurrentAccess(t *testing.T) {
	t.Run("publish handles concurrent subscriptions and publishing", func(t *testing.T) {
		// Clean up any existing subscriptions
		pubsub.SetSubscriptionsMap(make(map[string][]string))
		pubsub.SetSubscribedModeMap(make(map[string]bool))

		// Create multiple mock connections
		connections := make([]*MockConnection, 10)
		for i := 0; i < 10; i++ {
			conn := &MockConnection{
				Buffer:     &bytes.Buffer{},
				remoteAddr: fmt.Sprintf("127.0.0.1:1234%d", i),
				localAddr:  "127.0.0.1:6379",
			}
			connections[i] = conn
			network.ConnectionsSet(conn.remoteAddr, conn)
		}

		// Subscribe all connections to the same channel concurrently
		done := make(chan bool, 10)
		for i, conn := range connections {
			go func(conn *MockConnection, i int) {
				Subscribe(conn.remoteAddr, []shared.Value{{Typ: "bulk", Bulk: "concurrent-channel"}})
				done <- true
			}(conn, i)
		}

		// Wait for all subscriptions to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Publish message
		result := Publish("127.0.0.1:12350", []shared.Value{
			{Typ: "bulk", Bulk: "concurrent-channel"},
			{Typ: "bulk", Bulk: "concurrent message"},
		})

		// Verify result
		if result.Typ != "integer" || result.Num != 10 {
			t.Errorf("Expected integer 10, got %v", result)
		}

		// Verify all connections received the message
		expectedMessage := "*3\r\n$7\r\nmessage\r\n$18\r\nconcurrent-channel\r\n$18\r\nconcurrent message\r\n"
		for i, conn := range connections {
			actualMessage := conn.String()
			if actualMessage != expectedMessage {
				t.Errorf("Expected message %q for conn%d, got %q", expectedMessage, i, actualMessage)
			}
		}

		// Clean up
		for _, conn := range connections {
			network.ConnectionsDelete(conn.remoteAddr)
			pubsub.SubscriptionsDelete(conn.remoteAddr)
			pubsub.SubscribedModeDelete(conn.remoteAddr)
		}
	})
}
