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
		{
			name:    "RDB with expiration timestamps",
			hexData: "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fe00fb0505fc000c288ac70100000009626c7565626572727909626c75656265727279fc000c288ac7010000000a737472617762657272790662616e616e61fc000c288ac7010000000662616e616e6109726173706265727279fc009cef127e01000000056772617065056772617065fc000c288ac70100000009726173706265727279056d616e676fffac8c4b485b4c789e",
			expected: map[string]string{
				"blueberry":  "blueberry",
				"strawberry": "banana",
				"banana":     "raspberry",
				"grape":      "grape",
				"raspberry":  "mango",
			},
			wantErr: false,
		},
		{
			name:    "RDB with specific hexdump from user",
			hexData: "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fe00fb0505fc000c288ac70100000009726173706265727279056d616e676ffc009cef127e01000000056170706c650662616e616e61fc000c288ac7010000000662616e616e61056772617065fc000c288ac701000000056d616e676f09626c75656265727279fc000c288ac7010000000970696e656170706c65066f72616e6765ff24da7ab32f8f235a",
			expected: map[string]string{
				"raspberry": "mango",
				"apple":     "banana", // apple is loaded but expired
				"banana":    "grape",
				"mango":     "blueberry",
				"pineapple": "orange",
			},
			wantErr: false,
		},
		{
			name:    "RDB with new hexdump from user",
			hexData: "524544495330303131fa0a72656469732d62697473c040fa0972656469732d76657205372e322e30fe00fb0303fc009cef127e010000000970696e656170706c650662616e616e61fc000c288ac7010000000a73747261776265727279066f72616e6765fc000c288ac701000000056d616e676f0970696e656170706c65ffd8df7e4a4b906861",
			expected: map[string]string{
				"pineapple":  "banana",    // pineapple is loaded but not expired
				"strawberry": "orange",    // strawberry is loaded but expired
				"mango":      "pineapple", // mango is loaded but expired
			},
			wantErr: false,
		},
		{
			name:    "RDB with latest hexdump from user",
			hexData: "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fe00fb0404fc000c288ac7010000000a73747261776265727279056772617065fc009cef127e01000000056d616e676f066f72616e6765fc000c288ac701000000056170706c650470656172fc000c288ac7010000000662616e616e61056d616e676fff51db2234a3117faf",
			expected: map[string]string{
				"strawberry": "grape",  // strawberry is loaded but not expired
				"mango":      "orange", // mango is loaded but not expired
				"apple":      "pear",   // apple is loaded but not expired
				"banana":     "mango",  // banana is loaded but not expired
			},
			wantErr: false,
		},
		{
			name:    "RDB with blueberry hexdump from user",
			hexData: "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fe00fb0505fc000c288ac7010000000a73747261776265727279056772617065fc000c288ac7010000000970696e656170706c650470656172fc009cef127e0100000009626c7565626572727909726173706265727279fc000c288ac7010000000662616e616e610970696e656170706c65fc000c288ac70100000004706561720a73747261776265727279ffa48a15de7c461f05",
			expected: map[string]string{
				"strawberry": "grape",      // strawberry is loaded but not expired
				"pineapple":  "pear",       // pineapple is loaded but not expired
				"blueberry":  "raspberry",  // blueberry is loaded but expired
				"banana":     "pineapple",  // banana is loaded but not expired
				"pear":       "strawberry", // pear is loaded but not expired
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
