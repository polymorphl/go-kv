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

### Transaction Operations
- `MULTI` - Start a transaction block
- `EXEC` - Execute all commands in a transaction block

## Architecture

The server is built with a clean, modular architecture:

- **Network Layer**: TCP server listening on port 6379
- **Protocol Layer**: RESP (Redis Serialization Protocol) implementation
- **Command Handler**: Extensible command routing system
- **Storage**: In-memory key-value store with support for strings, lists, and streams

## Project Structure

```
app/
├── main.go                    # Server entry point and connection handling
├── handler.go                 # Command routing and request handling
├── writer.go                  # Response writing utilities
├── commands/                  # Individual command implementations
└── shared/                    # Shared utilities and data structures
```

## Getting Started

### Prerequisites
- Go 1.24 or later
- Make (optional, for using Makefile commands)

### Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd <project-name>
   ```

2. **Run the server**
   ```bash
   make run
   ```

3. **Test the server**
   ```bash
   redis-cli -p 6379
   ```

### Development Workflow

The project includes a comprehensive Makefile for easy development:

#### **Testing Commands**
```bash
make help                    # Show all available commands
make quick-test             # Run fast tests (excludes blocking operations)
make test-all               # Run all tests with coverage
make test-basic             # Test basic commands (PING, ECHO, GET, SET, INCR, TYPE)
make test-list              # Test list commands (LPUSH, RPUSH, LRANGE, LPOP, LLEN, BLPOP)
make test-stream             # Test stream commands (XADD, XRANGE, XREAD)
make test-transaction        # Test transaction commands (MULTI, EXEC)
```

#### **Benchmarking Commands**
```bash
make bench                  # Run all benchmarks
make bench-basic            # Benchmark basic commands
make bench-list             # Benchmark list commands
make bench-stream           # Benchmark stream commands
```

#### **Development Commands**
```bash
make build                  # Build the Redis server binary
make run                    # Run the Redis server
make clean                  # Clean build artifacts
make format                 # Format Go code
make lint                   # Run linter (requires golangci-lint)
make deps                   # Download and tidy dependencies
```



### Example Usage

#### **Basic Redis Operations**
```redis
# Basic commands
PING
ECHO "Hello World"
SET mykey "Hello Redis"
GET mykey
INCR counter

# List operations
LPUSH mylist "item1" "item2" "item3"
RPUSH mylist "item4"
LRANGE mylist 0 -1
LLEN mylist
LPOP mylist

# Stream operations
XADD mystream * field1 "value1" field2 "value2"
XRANGE mystream - +
XREAD STREAMS mystream 0

```

#### **Performance Testing**
```bash
# Test specific command performance (ex: bench-basic, bench-list, bench-stream)
make bench-basic 

# Generate coverage report
make test-coverage
```

## Implementation Details

- **Concurrent Connections**: Each client connection is handled in a separate goroutine
- **Memory Management**: In-memory storage with optional expiration support
- **Protocol Compliance**: Full RESP protocol implementation for Redis compatibility
- **Error Handling**: Robust error handling with graceful connection management
- **Transaction Support**: Connection-specific transaction state management
- **Stream Support**: Full Redis stream implementation with ID generation and blocking reads
- **Unicode Support**: Complete UTF-8 string support across all operations

## Test Coverage

### Test Statistics
```bash
make status
```


## Development

This project was developed as part of the [CodeCrafters Redis Challenge](https://codecrafters.io/challenges/redis), which provides an excellent learning experience for understanding distributed systems, network programming, and the Redis protocol.
