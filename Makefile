.PHONY: test test-unit test-integration test-coverage test-clean help

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Test targets
test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	go test -v -short ./...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	go test -v -run=Integration ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-clean: ## Clean up test artifacts
	@echo "Cleaning up test artifacts..."
	rm -f coverage.out coverage.html

# Development targets
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

# CI/CD targets
ci-test: ## Run tests for CI
	@echo "Running CI tests..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

ci-lint: ## Run linter for CI
	@echo "Running CI linter..."
	golangci-lint run --out-format=github-actions