package commands

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestWait(t *testing.T) {
	// Reset store state for clean test
	shared.StoreState = shared.State{
		Role:             "master",
		MasterReplID:     "test-repl-id",
		MasterReplOffset: 0,
		Replicas:         make(map[string]net.Conn),
	}

	// Clear acknowledged replicas
	shared.AcknowledgedReplicasClear()

	tests := []struct {
		name        string
		args        []shared.Value
		expectError bool
		expectedNum int
	}{
		{
			name: "WAIT with valid arguments",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "1"},
				{Typ: "bulk", Bulk: "100"},
			},
			expectError: false,
			expectedNum: 0, // No replicas connected initially
		},
		{
			name: "WAIT with higher replica count",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "3"},
				{Typ: "bulk", Bulk: "500"},
			},
			expectError: false,
			expectedNum: 0, // No replicas connected initially
		},
		{
			name: "WAIT with insufficient arguments",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "1"},
			},
			expectError: true,
		},
		{
			name:        "WAIT with no arguments",
			args:        []shared.Value{},
			expectError: true,
		},
		{
			name: "WAIT with invalid numreplicas",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "invalid"},
				{Typ: "bulk", Bulk: "100"},
			},
			expectError: true,
		},
		{
			name: "WAIT with invalid timeout",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "1"},
				{Typ: "bulk", Bulk: "invalid"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Wait("test-conn", tt.args)

			if tt.expectError {
				if result.Typ != "error" {
					t.Errorf("Expected error type, got %s", result.Typ)
				}
			} else {
				if result.Typ != "integer" {
					t.Errorf("Expected integer type, got %s", result.Typ)
				}
				if result.Num != tt.expectedNum {
					t.Errorf("Expected %d, got %d", tt.expectedNum, result.Num)
				}
			}
		})
	}
}

func TestWaitWithConnectedReplicas(t *testing.T) {
	// Set up store state with mock replicas
	shared.StoreState = shared.State{
		Role:             "master",
		MasterReplID:     "test-repl-id",
		MasterReplOffset: 0,
		Replicas:         make(map[string]net.Conn),
	}

	// Clear acknowledged replicas
	shared.AcknowledgedReplicasClear()

	// Add mock replica connections
	mockConn1 := &mockConn{}
	mockConn2 := &mockConn{}
	mockConn3 := &mockConn{}

	shared.ReplicasSet("replica-1", mockConn1)
	shared.ReplicasSet("replica-2", mockConn2)
	shared.ReplicasSet("replica-3", mockConn3)

	tests := []struct {
		name        string
		numReplicas int
		timeoutMs   int
		expectedMin int // Minimum expected result
		expectedMax int // Maximum expected result
	}{
		{
			name:        "WAIT 1 replica with 100ms timeout",
			numReplicas: 1,
			timeoutMs:   100,
			expectedMin: 0,
			expectedMax: 3, // Should return total replicas if no ACKs
		},
		{
			name:        "WAIT 3 replicas with 200ms timeout",
			numReplicas: 3,
			timeoutMs:   200,
			expectedMin: 0,
			expectedMax: 3, // Should return total replicas if no ACKs
		},
		{
			name:        "WAIT 5 replicas with 50ms timeout",
			numReplicas: 5,
			timeoutMs:   50,
			expectedMin: 0,
			expectedMax: 3, // Should return total replicas if no ACKs
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []shared.Value{
				{Typ: "bulk", Bulk: strconv.Itoa(tt.numReplicas)},
				{Typ: "bulk", Bulk: strconv.Itoa(tt.timeoutMs)},
			}

			start := time.Now()
			result := Wait("test-conn", args)
			duration := time.Since(start)

			if result.Typ != "integer" {
				t.Errorf("Expected integer type, got %s", result.Typ)
			}

			if result.Num < tt.expectedMin || result.Num > tt.expectedMax {
				t.Errorf("Expected result between %d and %d, got %d", tt.expectedMin, tt.expectedMax, result.Num)
			}

			// Check that it respects timeout (with some tolerance)
			// WAIT should always wait for the timeout when no acknowledgments are received
			if duration < time.Duration(tt.timeoutMs-10)*time.Millisecond {
				t.Errorf("WAIT returned too quickly, expected at least %dms, got %v", tt.timeoutMs-10, duration)
			}
		})
	}
}

func TestWaitWithAcknowledgedReplicas(t *testing.T) {
	// Set up store state
	shared.StoreState = shared.State{
		Role:             "master",
		MasterReplID:     "test-repl-id",
		MasterReplOffset: 0,
		Replicas:         make(map[string]net.Conn),
	}

	// Clear acknowledged replicas
	shared.AcknowledgedReplicasClear()

	// Add mock replica connections
	mockConn1 := &mockConn{}
	mockConn2 := &mockConn{}

	shared.ReplicasSet("replica-1", mockConn1)
	shared.ReplicasSet("replica-2", mockConn2)

	// Simulate acknowledgments
	shared.AcknowledgedReplicasSet("replica-1")
	shared.AcknowledgedReplicasSet("replica-2")

	args := []shared.Value{
		{Typ: "bulk", Bulk: "2"},
		{Typ: "bulk", Bulk: "100"},
	}

	result := Wait("test-conn", args)

	if result.Typ != "integer" {
		t.Errorf("Expected integer type, got %s", result.Typ)
	}

	if result.Num != 2 {
		t.Errorf("Expected 2 acknowledged replicas, got %d", result.Num)
	}
}

// mockConn is a simple mock implementation of net.Conn for testing
type mockConn struct{}

func (m *mockConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (m *mockConn) Write(b []byte) (n int, err error)  { return len(b), nil }
func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }


// BenchmarkWait benchmarks the WAIT command
func BenchmarkWait(b *testing.B) {
	// Reset store state for clean benchmark
	shared.StoreState = shared.State{
		Role:            "master",
		MasterReplID:    "bench-repl-id",
		MasterReplOffset: 0,
		Replicas:        make(map[string]net.Conn),
	}
	
	// Clear acknowledged replicas
	shared.AcknowledgedReplicasClear()

	args := []shared.Value{
		{Typ: "bulk", Bulk: "1"},
		{Typ: "bulk", Bulk: "10"}, // Short timeout for benchmark
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Wait("bench-conn", args)
	}
}
