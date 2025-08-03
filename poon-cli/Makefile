.PHONY: build test install clean fmt vet lint

# Build CLI binary
build:
	go build -o bin/poon cmd/poon/main.go

# Install CLI globally
install:
	go install ./cmd/poon

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -cover ./...

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf bin/

# Build for multiple platforms
cross-compile:
	GOOS=linux GOARCH=amd64 go build -o bin/poon-linux-amd64 cmd/poon/main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/poon-darwin-amd64 cmd/poon/main.go
	GOOS=windows GOARCH=amd64 go build -o bin/poon-windows-amd64.exe cmd/poon/main.go

# Development build with debug info
dev:
	go build -race -o bin/poon cmd/poon/main.go

# Show project structure
structure:
	@echo "Project structure:"
	@echo "├── cmd/"
	@echo "│   └── poon/          # Main CLI entry point"
	@echo "├── internal/"
	@echo "│   ├── cli/           # Root CLI setup"
	@echo "│   └── commands/      # Command implementations"
	@echo "│       ├── start/"
	@echo "│       ├── track/"
	@echo "│       ├── push/"
	@echo "│       ├── sync/"
	@echo "│       ├── status/"
	@echo "│       ├── ls/"
	@echo "│       ├── cat/"
	@echo "│       ├── branches/"
	@echo "│       └── workspace/"
	@echo "├── pkg/"
	@echo "│   ├── config/        # Configuration management"
	@echo "│   ├── client/        # gRPC client wrapper"
	@echo "│   └── util/          # Utility functions"
	@echo "├── bin/               # Built binaries"
	@echo "└── Makefile           # Build automation"