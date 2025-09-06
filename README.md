# Redis Implementation in Go

A lightweight Redis-compatible server implementation written in Go, featuring core Redis commands and the RESP (Redis Serialization Protocol).

**⚠️ Work in Progress** - This is an ongoing implementation with features being added incrementally.

## Features

This implementation supports the following Redis commands:

### Basic Commands
- `PING` - Test server connectivity
- `ECHO` - Echo back the provided message
- `TYPE` - Get the type of a key

### String Operations
- `SET` - Set a key-value pair with optional expiration
- `GET` - Retrieve a value by key
- `INCR` - Increment the value of a key by 1

### List Operations
- `LPUSH` - Push elements to the left of a list
- `RPUSH` - Push elements to the right of a list
- `LRANGE` - Get a range of elements from a list
- `LLEN` - Get the length of a list
- `LPOP` - Remove and return the leftmost element
- `BLPOP` - Blocking left pop operation

### Stream Operations
- `XADD` - Add entries to a stream with auto-generated or specified IDs
- `XRANGE` - Retrieve entries from a stream within a specified ID range
- `XREAD` - Read entries from one or more streams newer than specified IDs

## Architecture

The server is built with a clean, modular architecture:

- **Network Layer**: TCP server listening on port 6379
- **Protocol Layer**: RESP (Redis Serialization Protocol) implementation
- **Command Handler**: Extensible command routing system
- **Storage**: In-memory key-value store with support for strings, lists, and streams

## Project Structure

```
app/
├── main.go          # Server entry point and connection handling
├── handler.go       # Command handlers and in-memory storage
├── cmd_string.go    # String operation implementations
├── cmd_list.go      # List operation implementations
├── cmd_stream.go    # Stream operation implementations
├── resp.go          # RESP protocol implementation
├── marshal.go       # Data marshaling utilities
└── writer.go        # Response writing utilities
```

## Getting Started

### Prerequisites
- Go 1.24 or later

### Running the Server

1. Clone the repository
2. Navigate to the project directory
3. Run the server:
   ```bash
   ./your_program.sh
   ```

The server will start listening on `localhost:6379`.

### Testing

You can test the server using any Redis client or the `redis-cli`:

```bash
redis-cli -p 6379
```

Example commands:
```redis
PING
SET mykey "Hello World"
GET mykey
LPUSH mylist "item1" "item2"
LRANGE mylist 0 -1
```

## Implementation Details

- **Concurrent Connections**: Each client connection is handled in a separate goroutine
- **Memory Management**: In-memory storage with optional expiration support
- **Protocol Compliance**: Full RESP protocol implementation for Redis compatibility
- **Error Handling**: Robust error handling with graceful connection management

## Development

This project was developed as part of the [CodeCrafters Redis Challenge](https://codecrafters.io/challenges/redis), which provides an excellent learning experience for understanding distributed systems, network programming, and the Redis protocol.
