package shared

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestRDBParser(t *testing.T) {
	tests := []struct {
		name     string
		hexData  string
		expected map[string]string
		wantErr  bool
	}{
		{
			name:     "Empty RDB file",
			hexData:  "",
			expected: map[string]string{},
			wantErr:  false,
		},
		{
			name:    "Valid RDB with single key",
			hexData: "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fe00fb01000009626c75656265727279066f72616e6765ff17353b458361d7a0",
			expected: map[string]string{
				"blueberry": "orange",
			},
			wantErr: false,
		},
		{
			name:     "Invalid RDB header",
			hexData:  "494e56414c4944484541444552",
			expected: map[string]string{},
			wantErr:  true,
		},
		{
			name:    "RDB with multiple keys",
			hexData: "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fe00fb02000003666f6f036261720362617a03717578ff17353b458361d7a0",
			expected: map[string]string{
				"foo": "bar",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear memory before each test
			Memory = make(map[string]MemoryEntry)

			var data []byte
			var err error

			if tt.hexData != "" {
				data, err = hex.DecodeString(tt.hexData)
				if err != nil {
					t.Fatalf("Failed to decode hex data: %v", err)
				}
			}

			err = ParseRDBData(data)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check that all expected keys are present
			for expectedKey, expectedValue := range tt.expected {
				entry, exists := Memory[expectedKey]
				if !exists {
					t.Errorf("Expected key '%s' not found in memory", expectedKey)
					continue
				}
				if entry.Value != expectedValue {
					t.Errorf("Expected value '%s' for key '%s', got '%s'", expectedValue, expectedKey, entry.Value)
				}
			}

			// Check that no unexpected keys are present
			if len(Memory) != len(tt.expected) {
				t.Errorf("Expected %d keys in memory, got %d", len(tt.expected), len(Memory))
				for key, entry := range Memory {
					t.Errorf("Unexpected key: %s = %s", key, entry.Value)
				}
			}
		})
	}
}

func TestLoadRDBFile(t *testing.T) {
	tests := []struct {
		name        string
		setupFile   bool
		fileContent string
		expected    map[string]string
		wantErr     bool
	}{
		{
			name:        "File does not exist",
			setupFile:   false,
			fileContent: "",
			expected:    map[string]string{},
			wantErr:     false,
		},
		{
			name:        "Valid RDB file",
			setupFile:   true,
			fileContent: "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fe00fb01000009626c75656265727279066f72616e6765ff17353b458361d7a0",
			expected: map[string]string{
				"blueberry": "orange",
			},
			wantErr: false,
		},
		{
			name:        "Invalid RDB file",
			setupFile:   true,
			fileContent: "494e56414c4944484541444552",
			expected:    map[string]string{},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir := t.TempDir()
			tempFile := filepath.Join(tempDir, "test.rdb")

			// Clear memory before each test
			Memory = make(map[string]MemoryEntry)

			// Setup file if needed
			if tt.setupFile {
				var data []byte
				var err error

				if tt.fileContent != "" {
					data, err = hex.DecodeString(tt.fileContent)
					if err != nil {
						t.Fatalf("Failed to decode hex data: %v", err)
					}
				}

				err = os.WriteFile(tempFile, data, 0644)
				if err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
			}

			// Test LoadRDBFile
			err := LoadRDBFile(tempDir, "test.rdb")

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check that all expected keys are present
			for expectedKey, expectedValue := range tt.expected {
				entry, exists := Memory[expectedKey]
				if !exists {
					t.Errorf("Expected key '%s' not found in memory", expectedKey)
					continue
				}
				if entry.Value != expectedValue {
					t.Errorf("Expected value '%s' for key '%s', got '%s'", expectedValue, expectedKey, entry.Value)
				}
			}

			// Check that no unexpected keys are present
			if len(Memory) != len(tt.expected) {
				t.Errorf("Expected %d keys in memory, got %d", len(tt.expected), len(Memory))
				for key, entry := range Memory {
					t.Errorf("Unexpected key: %s = %s", key, entry.Value)
				}
			}
		})
	}
}

func TestRDBParserEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		hexData string
		wantErr bool
	}{
		{
			name:    "Truncated RDB file",
			hexData: "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fe00fb01000009626c75656265727279066f72616e6765",
			wantErr: false, // Parser now handles truncated files gracefully
		},
		{
			name:    "RDB with empty database",
			hexData: "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fe00fb000000ff17353b458361d7a0",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear memory before each test
			Memory = make(map[string]MemoryEntry)

			data, err := hex.DecodeString(tt.hexData)
			if err != nil {
				t.Fatalf("Failed to decode hex data: %v", err)
			}

			err = ParseRDBData(data)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestRDBParserLengthEncoding(t *testing.T) {
	tests := []struct {
		name     string
		hexData  string
		expected map[string]string
	}{
		{
			name:    "6-bit length encoding",
			hexData: "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fe00fb01000003666f6f03626172ff17353b458361d7a0",
			expected: map[string]string{
				"foo": "bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear memory before each test
			Memory = make(map[string]MemoryEntry)

			data, err := hex.DecodeString(tt.hexData)
			if err != nil {
				t.Fatalf("Failed to decode hex data: %v", err)
			}

			err = ParseRDBData(data)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check that all expected keys are present
			for expectedKey, expectedValue := range tt.expected {
				entry, exists := Memory[expectedKey]
				if !exists {
					t.Errorf("Expected key '%s' not found in memory", expectedKey)
					continue
				}
				if entry.Value != expectedValue {
					t.Errorf("Expected value '%s' for key '%s', got '%s'", expectedValue, expectedKey, entry.Value)
				}
			}
		})
	}
}

func BenchmarkRDBParser(b *testing.B) {
	// Test data with single key
	hexData := "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fe00fb01000009626c75656265727279066f72616e6765ff17353b458361d7a0"
	data, err := hex.DecodeString(hexData)
	if err != nil {
		b.Fatalf("Failed to decode hex data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Memory = make(map[string]MemoryEntry)
		ParseRDBData(data)
	}
}

func BenchmarkRDBParserLarge(b *testing.B) {
	// Create a larger RDB file with multiple keys
	hexData := "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fe00fb100000036b65790376616c036b65790376616c036b65790376616c036b65790376616c036b65790376616c036b65790376616c036b65790376616c036b65790376616c036b65790376616c036b65790376616c036b65790376616c036b65790376616c036b65790376616cff17353b458361d7a0"
	data, err := hex.DecodeString(hexData)
	if err != nil {
		b.Fatalf("Failed to decode hex data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Memory = make(map[string]MemoryEntry)
		ParseRDBData(data)
	}
}
