# Variables
SERVER_BIN=bin/server
CLI_BIN=bin/job-cli
GO_FILES=$(shell find . -name '*.go')


.PHONY: all build clean run

all: clean setup proto build

# Build both Server and CLI
build: build-server build-cli

build-server:
	@echo "Building Server..."
	@go build -o $(SERVER_BIN) cmd/server/*.go

build-cli:
	@echo "Building CLI..."
	@go build -o $(CLI_BIN) cmd/cli/main.go

# Run the server
run:
	@echo "Running Server..."
	-@go run cmd/server/*.go


# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin
	@rm -rf proto/*.pb.go
	@rm -rf proto/*.gw.go


.PHONY: fmt lint tidy

fmt:
	@echo "Formatting code..."
	@go fmt ./...

lint:
	@echo "Running Linter..."
	@golangci-lint run ./...

tidy:
	@echo "Tidying modules..."
	@go mod tidy


.PHONY: proto setup

# Generate Go code from Proto files
proto:
	@echo "Rb  Generating gRPC code..."
	@protoc -I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
		proto/scheduler.proto
	@echo "Proto generation complete."

# One-time setup: installs tools and downloads google/api protos
setup:
	@echo "🔧 Installing Protoc Plugins..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@echo "Downloading Google API definitions..."
	@mkdir -p google/api
	@curl -s -o google/api/annotations.proto https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto
	@curl -s -o google/api/http.proto https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto
	@echo "Setup complete."
