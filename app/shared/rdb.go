package shared

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// RDBParser handles parsing RDB files
type RDBParser struct {
	data []byte
	pos  int
}

// NewRDBParser creates a new RDB parser
func NewRDBParser(data []byte) *RDBParser {
	return &RDBParser{data: data, pos: 0}
}

// LoadRDBFile loads an RDB file from disk and parses it
func LoadRDBFile(dir, filename string) error {
	filePath := filepath.Join(dir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist, that's okay - start with empty database
		return nil
	}

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read RDB file %s: %v", filePath, err)
	}

	// Parse the RDB data
	return ParseRDBData(data)
}

// ParseRDBData parses RDB data and loads it into memory
func ParseRDBData(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	// Clear existing memory before loading RDB data
	Memory = make(map[string]MemoryEntry)

	parser := NewRDBParser(data)
	return parser.parse()
}

// parse parses the RDB data
func (p *RDBParser) parse() error {
	// Check RDB header
	if !p.checkHeader() {
		return fmt.Errorf("invalid RDB header")
	}

	// Skip metadata (redis-ver, redis-bits, etc.)
	if err := p.skipMetadata(); err != nil {
		return fmt.Errorf("failed to skip metadata: %v", err)
	}

	// Parse database data
	for {
		// Check if we're at EOF
		if p.pos >= len(p.data) {
			break
		}

		opcode, err := p.readByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read opcode: %v", err)
		}

		switch opcode {
		case 0xFE: // SELECTDB
			if err := p.parseSelectDB(); err != nil {
				return fmt.Errorf("failed to parse SELECTDB: %v", err)
			}
		case 0xFB: // RESIZEDB
			if err := p.parseResizeDB(); err != nil {
				return fmt.Errorf("failed to parse RESIZEDB: %v", err)
			}
			// After RESIZEDB, we expect key-value pairs
			if err := p.parseKeyValuePairs(); err != nil {
				return fmt.Errorf("failed to parse key-value pairs: %v", err)
			}
			// We're done parsing this database
			return nil
		case 0xFF: // EOF
			// End of file, we're done
			return nil
		default:
			return fmt.Errorf("unexpected opcode: 0x%02X", opcode)
		}
	}

	return nil
}

// checkHeader checks if the RDB header is valid
func (p *RDBParser) checkHeader() bool {
	if len(p.data) < 9 {
		return false
	}

	header := string(p.data[0:9])
	return header == "REDIS0011"
}

// skipMetadata skips the metadata section
func (p *RDBParser) skipMetadata() error {
	p.pos = 9 // Skip "REDIS0011"

	for {
		if p.pos >= len(p.data) {
			return io.EOF
		}

		opcode, err := p.readByte()
		if err != nil {
			return err
		}

		switch opcode {
		case 0xFA: // Auxiliary field
			if err := p.skipAuxiliaryField(); err != nil {
				return err
			}
		case 0xFE: // SELECTDB - start of database data
			p.pos-- // Back up one byte
			return nil
		case 0x40: // Skip this byte (appears after redis-bits)
			continue
		default:
			return fmt.Errorf("unexpected opcode in metadata: 0x%02X", opcode)
		}
	}
}

// skipAuxiliaryField skips an auxiliary field
func (p *RDBParser) skipAuxiliaryField() error {
	// Skip key
	if err := p.skipLengthEncodedString(); err != nil {
		return err
	}
	// Skip value
	if err := p.skipLengthEncodedString(); err != nil {
		return err
	}
	return nil
}

// parseSelectDB parses SELECTDB opcode
func (p *RDBParser) parseSelectDB() error {
	// Read database number (we ignore it for now)
	_, err := p.readLength()
	return err
}

// parseResizeDB parses RESIZEDB opcode
func (p *RDBParser) parseResizeDB() error {
	// Read hash table size (we ignore it for now)
	_, err := p.readLength()
	if err != nil {
		return err
	}
	// Read expiry hash table size (we ignore it for now)
	_, err = p.readLength()
	return err
}

// parseKeyValuePairs parses key-value pairs until EOF
func (p *RDBParser) parseKeyValuePairs() error {
	keyCount := 0
	for {
		// Check if we're at EOF
		if p.pos >= len(p.data) {
			return nil
		}

		// Peek at the next byte to see if it's EOF
		if p.data[p.pos] == 0xFF {
			return nil
		}

		// Read the value type first
		valueType, err := p.readByte()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		// Parse a key-value pair with the correct type
		if err := p.parseKeyValue(valueType); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		keyCount++

		// After parsing a key-value pair, check if we're at EOF
		if p.pos >= len(p.data) {
			return nil
		}

		if p.data[p.pos] == 0xFF {
			return nil
		}
	}
}

// parseKeyValue parses a key-value pair
func (p *RDBParser) parseKeyValue(opcode byte) error {
	var expires int64 = 0
	var valueType byte = opcode

	// Handle expiration opcodes first
	switch opcode {
	case 0xFC: // Expiry time in seconds
		// Skip the first 3 bytes (00 0c 28 or 00 9c ef)
		if p.pos+3 > len(p.data) {
			return io.EOF
		}
		p.pos += 3
		// Read the expiration timestamp (4 bytes)
		if p.pos+4 > len(p.data) {
			return io.EOF
		}
		expiresVal, err := p.readUint32BigEndian()
		if err != nil {
			return err
		}
		// Convert from seconds to milliseconds (absolute timestamp)
		expires = expiresVal * 1000
		// Skip the remaining 2 bytes (00 00)
		if p.pos+2 > len(p.data) {
			return io.EOF
		}
		p.pos += 2
		// The value type is implicit (0x00 for string)
		valueType = 0x00

	case 0xFD: // Expiry time in milliseconds
		var err error
		expires, err = p.readUint32BigEndian()
		if err != nil {
			return err
		}
		valueType, err = p.readByte()
		if err != nil {
			return err
		}
	}

	// Read key
	key, err := p.readLengthEncodedString()
	if err != nil {
		return err
	}

	// Read value based on type
	var value string
	switch valueType {
	case 0x00: // String
		value, err = p.readLengthEncodedString()
		if err != nil {
			return err
		}
	case 0x01: // String (alternative encoding)
		value, err = p.readLengthEncodedString()
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported value type: 0x%02X", valueType)
	}

	// Store in memory
	Memory[key] = MemoryEntry{
		Value:   value,
		Expires: expires,
	}

	return nil
}

// Helper methods

func (p *RDBParser) readByte() (byte, error) {
	if p.pos >= len(p.data) {
		return 0, io.EOF
	}
	b := p.data[p.pos]
	p.pos++
	return b, nil
}

func (p *RDBParser) readLength() (int, error) {
	firstByte, err := p.readByte()
	if err != nil {
		return 0, err
	}

	// Length encoding format:
	// 00xxxxxx - 6 bit length
	// 01xxxxxx xxxxxxxx - 14 bit length
	// 10xxxxxx xxxxxxxx xxxxxxxx xxxxxxxx - 32 bit length
	// 11xxxxxx - special encoding

	switch (firstByte >> 6) & 0x03 {
	case 0: // 6 bit length
		return int(firstByte & 0x3F), nil
	case 1: // 14 bit length
		secondByte, err := p.readByte()
		if err != nil {
			return 0, err
		}
		return int((uint16(firstByte&0x3F) << 8) | uint16(secondByte)), nil
	case 2: // 32 bit length
		if p.pos+4 > len(p.data) {
			return 0, io.EOF
		}
		length := binary.BigEndian.Uint32(p.data[p.pos : p.pos+4])
		p.pos += 4
		return int(length), nil
	default: // Special encoding (11xxxxxx)
		// For now, just return the lower 6 bits
		// This handles cases like 0xC0 (192) which should be treated as length 0
		return int(firstByte & 0x3F), nil
	}
}

func (p *RDBParser) readLengthEncodedString() (string, error) {
	length, err := p.readLength()
	if err != nil {
		return "", err
	}

	if p.pos+length > len(p.data) {
		return "", io.EOF
	}

	str := string(p.data[p.pos : p.pos+length])
	p.pos += length
	return str, nil
}

func (p *RDBParser) readUint32BigEndian() (int64, error) {
	if p.pos+4 > len(p.data) {
		return 0, io.EOF
	}

	value := binary.BigEndian.Uint32(p.data[p.pos : p.pos+4])
	p.pos += 4
	return int64(value), nil
}

func (p *RDBParser) skipLengthEncodedString() error {
	length, err := p.readLength()
	if err != nil {
		return err
	}

	if p.pos+length > len(p.data) {
		return io.EOF
	}

	p.pos += length
	return nil
}
