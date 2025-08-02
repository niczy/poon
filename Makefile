# Poon Monorepo System Makefile

.PHONY: all build test clean install proto help

# Default target
all: proto build test

# Help target
help:
	@echo "Poon Monorepo System Build Commands"
	@echo "=================================="
	@echo "make all      - Build everything (proto + build + test)"
	@echo "make build    - Build all components"
	@echo "make test     - Run all tests"
	@echo "make proto    - Generate protobuf files"
	@echo "make install  - Install all dependencies"
	@echo "make clean    - Clean build artifacts"
	@echo "make start    - Start all services in background"
	@echo "make stop     - Stop all services"
	@echo "make help     - Show this help message"

# Install dependencies
install:
	@echo "Installing Go protobuf plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	
	@echo "Installing Node.js dependencies..."
	cd poon-proto && npm install
	cd poon-web && npm install

# Generate protobuf files
proto:
	@echo "Generating protobuf files..."
	cd poon-proto && \
	mkdir -p gen/go gen/js gen/python gen/ts && \
	protoc --go_out=gen --go_opt=paths=source_relative \
	       --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
	       --proto_path=. monorepo.proto && \
	mv gen/monorepo*.pb.go gen/go/ 2>/dev/null || true

# Build all components
build: proto
	@echo "Building Go components..."
	cd poon-server && go build -o poon-server .
	cd poon-git && go build -o poon-git .
	cd poon-cli && go build -o poon-cli .
	
	@echo "Building web interface..."
	cd poon-web && npm run build

# Run all tests
test:
	@echo "Running tests for all components..."
	cd poon-git && chmod +x scripts/run_test.sh && ./scripts/run_test.sh
	cd poon-server && chmod +x scripts/run_test.sh && ./scripts/run_test.sh
	cd poon-cli && chmod +x scripts/run_test.sh && ./scripts/run_test.sh
	cd poon-proto && chmod +x scripts/run_test.sh && ./scripts/run_test.sh
	cd poon-web && chmod +x scripts/run_test.sh && ./scripts/run_test.sh
	cd poon-tests && chmod +x scripts/run_test.sh && ./scripts/run_test.sh

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f poon-server/poon-server
	rm -f poon-git/poon-git
	rm -f poon-cli/poon-cli
	rm -f poon-cli/poon
	rm -rf poon-web/.next
	rm -rf poon-web/out
	find . -name "*.test" -delete
	find . -name "*.log" -delete
	find . -name ".DS_Store" -delete

# Development helpers
dev-server:
	@echo "Starting poon-server in development mode..."
	cd poon-server && go run .

dev-git:
	@echo "Starting poon-git server in development mode..."
	cd poon-git && go run .

dev-web:
	@echo "Starting poon-web in development mode..."
	cd poon-web && npm run dev

# Start all services in background (requires build first)
start: build
	@echo "Starting all services..."
	cd poon-server && ./poon-server & echo $$! > poon-server.pid
	cd poon-git && ./poon-git & echo $$! > poon-git.pid
	cd poon-web && npm start & echo $$! > poon-web.pid
	@echo "Services started. Use 'make stop' to stop them."

# Stop all services
stop:
	@echo "Stopping all services..."
	-kill `cat poon-server/poon-server.pid 2>/dev/null` 2>/dev/null || true
	-kill `cat poon-git/poon-git.pid 2>/dev/null` 2>/dev/null || true
	-kill `cat poon-web/poon-web.pid 2>/dev/null` 2>/dev/null || true
	-rm -f poon-server/poon-server.pid poon-git/poon-git.pid poon-web/poon-web.pid
	@echo "Services stopped."

# Format code
format:
	@echo "Formatting Go code..."
	find . -name "*.go" -not -path "./poon-proto/gen/*" -exec gofmt -w {} \;
	
	@echo "Formatting TypeScript code..."
	cd poon-web && npm run lint --fix || true

# Development shortcuts
server: proto
	cd poon-server && go build -o poon-server . && ./poon-server

git: proto  
	cd poon-git && go build -o poon-git . && ./poon-git

cli: proto
	cd poon-cli && go build -o poon-cli .

web:
	cd poon-web && npm run dev