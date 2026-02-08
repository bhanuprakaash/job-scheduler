#env
-include .env
export

# Variables
SERVER_BIN=bin/server
CLI_BIN=bin/job-cli
GO_FILES=$(shell find . -name '*.go')


.PHONY: all build clean run test

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

test:
	@echo "🧪 Running Unit Tests..."
	@go test -v ./...


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


# ==============================================================================
#  Docker Operations 
# ==============================================================================

# 1. Development Mode (Base + Override)
dev:
	@echo "Starting Full Stack (Dev Mode)..."
	@docker compose up -d
	@echo "App: http://localhost:8080 | Grafana: http://localhost:3000"

# 2. Infrastructure Only
infra:
	@echo "Starting Infrastructure Only..."
	@docker compose up -d postgres minio prometheus grafana
	@echo "DBs are ready. You can now run 'make run'"

# 3. Production Simulation
prod:
	@echo "Deploying Production Build..."
	@docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
	@echo "Production version is live."

# 4. Stop Everything
down:
	@echo "Stopping all containers..."
	@docker compose down --remove-orphans

# 5. View Logs
logs:
	@docker compose logs -f

# 6. Clean Everything (Nuclear Option)
clean-docker:
	@echo "Wiping all Docker volumes and containers..."
	@docker compose down -v --remove-orphans
	@echo "Clean slate."