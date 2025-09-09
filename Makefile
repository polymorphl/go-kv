# Redis Go Implementation Makefile
# Provides convenient commands for testing, benchmarking, and development

.PHONY: help test test-all test-basic test-list test-stream test-transaction bench bench-all bench-basic bench-list bench-stream clean build run lint format

# Default target
.DEFAULT_GOAL := help

# Colors for output
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
RED := \033[31m
RESET := \033[0m

# Go command with common flags
GO := go
GO_TEST := $(GO) test
GO_BENCH := $(GO_TEST) -bench=. -benchmem
GO_BUILD := $(GO) build

# Test timeout (in seconds)
TEST_TIMEOUT := 30s

help: ## Show this help message
	@echo "$(BLUE)Redis Go Implementation$(RESET)"
	@echo "$(BLUE)=====================$(RESET)"
	@echo ""
	@echo "$(GREEN)Available commands:$(RESET)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(YELLOW)%-15s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(GREEN)Examples:$(RESET)"
	@echo "  make test-basic     # Run basic command tests"
	@echo "  make bench-list     # Benchmark list commands"
	@echo "  make test-all       # Run all tests"
	@echo "  make clean          # Clean build artifacts"

# Testing commands
test: ## Run all tests
	@echo "$(BLUE)Running all tests...$(RESET)"
	$(GO_TEST) ./app/commands -v -timeout $(TEST_TIMEOUT)

test-all: ## Run all tests with coverage
	@echo "$(BLUE)Running all tests with coverage...$(RESET)"
	$(GO_TEST) ./app/commands -v -cover -timeout $(TEST_TIMEOUT)

test-basic: ## Run basic command tests (PING, ECHO, GET, SET, INCR, TYPE)
	@echo "$(BLUE)Running basic command tests...$(RESET)"
	$(GO_TEST) ./app/commands -v -run "TestPing|TestEcho|TestGet|TestSet|TestIncr|TestType" -timeout $(TEST_TIMEOUT)

test-list: ## Run list command tests (LPUSH, RPUSH, LRANGE, LPOP, LLEN, BLPOP)
	@echo "$(BLUE)Running list command tests...$(RESET)"
	$(GO_TEST) ./app/commands -v -run "TestLpush|TestRpush|TestLrange|TestLpop|TestLlen|TestBlpop" -timeout $(TEST_TIMEOUT)

test-stream: ## Run stream command tests (XADD, XRANGE, XREAD)
	@echo "$(BLUE)Running stream command tests...$(RESET)"
	$(GO_TEST) ./app/commands -v -run "TestXadd|TestXrange|TestXread" -timeout $(TEST_TIMEOUT)

test-transaction: ## Run transaction command tests (MULTI, EXEC, DISCARD)
	@echo "$(BLUE)Running transaction command tests...$(RESET)"
	$(GO_TEST) ./app/commands -v -run "TestMulti|TestExec|TestDiscard" -timeout $(TEST_TIMEOUT)

# Benchmarking commands
bench: ## Run all benchmarks
	@echo "$(BLUE)Running all benchmarks...$(RESET)"
	$(GO_BENCH) ./app/commands

bench-all: ## Run all benchmarks with detailed output
	@echo "$(BLUE)Running all benchmarks with detailed output...$(RESET)"
	$(GO_BENCH) ./app/commands -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof

bench-basic: ## Run basic command benchmarks (PING, ECHO, GET, SET, INCR, TYPE)
	@echo "$(BLUE)Running basic command benchmarks...$(RESET)"
	$(GO_TEST) ./app/commands -bench="BenchmarkPing|BenchmarkEcho|BenchmarkGet|BenchmarkSet|BenchmarkIncr|BenchmarkType" -benchmem

bench-list: ## Run list command benchmarks (LPUSH, RPUSH, LRANGE, LPOP, LLEN)
	@echo "$(BLUE)Running list command benchmarks...$(RESET)"
	$(GO_TEST) ./app/commands -bench="BenchmarkLpush|BenchmarkRpush|BenchmarkLrange|BenchmarkLpop|BenchmarkLlen" -benchmem

bench-stream: ## Run stream command benchmarks (XADD, XRANGE, XREAD)
	@echo "$(BLUE)Running stream command benchmarks...$(RESET)"
	$(GO_TEST) ./app/commands -bench="BenchmarkXadd|BenchmarkXrange|BenchmarkXread" -benchmem

# Development commands
build: ## Build the Redis server
	@echo "$(BLUE)Building Redis server...$(RESET)"
	$(GO_BUILD) -o redis-server ./app

run: ## Run the Redis server
	@echo "$(BLUE)Running Redis server...$(RESET)"
	$(GO) run ./app

lint: ## Run linter
	@echo "$(BLUE)Running linter...$(RESET)"
	golangci-lint run

format: ## Format Go code
	@echo "$(BLUE)Formatting Go code...$(RESET)"
	$(GO) fmt ./...

# Utility commands
clean: ## Clean build artifacts and temporary files
	@echo "$(BLUE)Cleaning build artifacts...$(RESET)"
	rm -f redis-server
	rm -f *.prof
	rm -f *.log
	$(GO) clean

deps: ## Download dependencies
	@echo "$(BLUE)Downloading dependencies...$(RESET)"
	$(GO) mod download
	$(GO) mod tidy

# Quick test commands for development
quick-test: ## Run quick tests (excluding slow BLPOP tests)
	@echo "$(BLUE)Running quick tests...$(RESET)"
	$(GO_TEST) ./app/commands -v -run "TestPing|TestEcho|TestGet|TestSet|TestIncr|TestType|TestLpush|TestRpush|TestLrange|TestLpop|TestLlen|TestXadd|TestXrange|TestXread|TestMulti|TestExec|TestDiscard" -timeout $(TEST_TIMEOUT)

# CI/CD helpers
ci-test: ## Run tests suitable for CI (no blocking operations)
	@echo "$(BLUE)Running CI tests...$(RESET)"
	$(GO_TEST) ./app/commands -v -run "TestPing|TestEcho|TestGet|TestSet|TestIncr|TestType|TestLpush|TestRpush|TestLrange|TestLpop|TestLlen|TestXadd|TestXrange|TestMulti|TestExec|TestDiscard" -timeout $(TEST_TIMEOUT)

ci-bench: ## Run benchmarks suitable for CI
	@echo "$(BLUE)Running CI benchmarks...$(RESET)"
	$(GO_TEST) ./app/commands -bench="BenchmarkPing|BenchmarkEcho|BenchmarkGet|BenchmarkSet|BenchmarkIncr|BenchmarkType|BenchmarkLpush|BenchmarkRpush|BenchmarkLrange|BenchmarkLpop|BenchmarkLlen|BenchmarkXadd|BenchmarkXrange|BenchmarkXread" -benchmem

# Documentation
test-coverage: ## Generate test coverage report
	@echo "$(BLUE)Generating test coverage report...$(RESET)"
	$(GO_TEST) ./app/commands -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(RESET)"

# Show project status
status: ## Show project status and test results
	@echo "$(BLUE)Project Status$(RESET)"
	@echo "$(BLUE)=============$(RESET)"
	@echo "$(GREEN)Commands implemented:$(RESET)"
	@echo "  • Basic: PING, ECHO, GET, SET, INCR, TYPE"
	@echo "  • Lists: LPUSH, RPUSH, LRANGE, LPOP, LLEN, BLPOP"
	@echo "  • Streams: XADD, XRANGE, XREAD"
	@echo "  • Transactions: MULTI, EXEC, DISCARD"
	@echo ""
	@echo "$(GREEN)Test coverage:$(RESET)"
	@echo "  • $(shell find ./app/commands -name "*_test.go" | wc -l | tr -d ' ') test files"
	@echo "  • $(shell grep -r "func Test" ./app/commands | wc -l | tr -d ' ') test functions"
	@echo "  • $(shell grep -r "func Benchmark" ./app/commands | wc -l | tr -d ' ') benchmark functions"
	@echo ""
	@echo "$(GREEN)Run 'make help' for available commands$(RESET)"
