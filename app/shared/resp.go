package shared

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

type Value struct {
	Typ     string
	Str     string
	Num     int
	Bulk    string
	Array   []Value
	Expires int64
}

// Special value types
const (
	NO_RESPONSE = "no_response"
)

type Resp struct {
	reader *bufio.Reader
}

func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

func (r *Resp) readArray() (Value, error) {
	v := Value{}
	v.Typ = "array"

	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}
	for range len {
		val, err := r.Read()
		if err != nil {
			return v, err
		}

		v.Array = append(v.Array, val)
	}
	return v, nil
}

func (r *Resp) readLine() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		n += 1
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}
	return line[:len(line)-2], n, nil
}

// returns the integer from the buffer and the number of bytes in the buffer.
func (r *Resp) readInteger() (x int, n int, err error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}
	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}
	return int(i64), n, nil
}

func (r *Resp) readBulk() (Value, error) {
	v := Value{}
	v.Typ = "bulk"

	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	bulk := make([]byte, len)
	r.reader.Read(bulk)
	v.Bulk = string(bulk)

	// read the CRLF
	r.readLine()

	return v, nil
}

// ReadBulkWithoutCRLF reads a bulk string without expecting trailing CRLF
// This is used for reading RDB files which don't have trailing CRLF
func (r *Resp) ReadBulkWithoutCRLF() (Value, error) {
	v := Value{}
	v.Typ = "bulk"

	// Read the bulk string length (format: $<length>\r\n)
	line, _, err := r.readLine()
	if err != nil {
		return v, err
	}

	// Parse the length from the line (should be $<length>)
	if len(line) < 2 || line[0] != '$' {
		return v, fmt.Errorf("invalid bulk string header: %s", string(line))
	}

	lengthStr := string(line[1:])
	len, err := strconv.Atoi(lengthStr)
	if err != nil {
		return v, fmt.Errorf("failed to parse bulk string length: %s", lengthStr)
	}

	bulk := make([]byte, len)
	r.reader.Read(bulk)
	v.Bulk = string(bulk)

	// Don't read the CRLF - RDB files don't have it

	return v, nil
}

func (r *Resp) readString() (Value, error) {
	v := Value{}
	v.Typ = "string"

	line, _, err := r.readLine()
	if err != nil {
		return v, err
	}

	v.Str = string(line)
	return v, nil
}

func (r *Resp) Read() (Value, error) {
	_type, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch _type {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	case STRING:
		return r.readString()
	default:
		fmt.Printf("Unknown type: %v\n", string(_type))
		return Value{}, nil
	}
}
