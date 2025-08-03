# Poon Monorepo System

[![Presubmit Tests](https://github.com/nic/poon/actions/workflows/presubmit.yml/badge.svg)](https://github.com/nic/poon/actions/workflows/presubmit.yml)
[![Integration Tests](https://github.com/nic/poon/actions/workflows/integration.yml/badge.svg)](https://github.com/nic/poon/actions/workflows/integration.yml)

A modern, gRPC-powered monorepo management system designed for internet-scale development workflows. Poon provides Git-compatible interfaces, web-based browsing, and CLI tools for efficient monorepo operations.

## 🏗️ Architecture

The Poon system consists of six interconnected components:

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  poon-web   │    │  poon-cli   │    │  poon-git   │
│   (Next.js) │    │   (CLI)     │    │ (Git Server)│
└─────┬───────┘    └─────┬───────┘    └─────┬───────┘
      │                  │                  │
      │            gRPC  │                  │ Git HTTP
      │                  │                  │ Protocol
      └──────────────────┼──────────────────┘
                         │                  
                ┌────────▼────────┐          
                │  poon-server    │          
                │   (gRPC API)    │          
                │ ┌─────────────┐ │          
                │ │  Workspace  │ │          
                │ │   Storage   │ │          
                │ └─────────────┘ │          
                └─────────────────┘          
```

### Components

- **🖥️ poon-server** - Core gRPC API server with content-addressable storage and workspace management
- **🌐 poon-web** - Modern Next.js web interface with gRPC-Web client
- **⚡ poon-git** - Git-compatible HTTP server providing Git protocol endpoints
- **🛠️ poon-cli** - Command-line interface for workspace creation and developer workflows
- **📦 poon-proto** - Protocol Buffer definitions and generated clients
- **🧪 poon-tests** - Comprehensive integration test suite

## 🚀 Quick Start

### Prerequisites

- **Go 1.23+** (for server components)
- **Node.js 20+** (for web interface)
- **Protocol Buffers compiler** (`protoc`)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/nic/poon.git
   cd poon
   ```

2. **Generate protobuf files**
   ```bash
   cd poon-proto
   npm install
   npm run proto:generate:go
   cd ..
   ```

3. **Build all components**
   ```bash
   # Build Go components
   cd poon-server && go build -o poon-server . && cd ..
   cd poon-git && go build -o poon-git . && cd ..
   cd poon-cli && go build -o poon-cli . && cd ..
   
   # Build web interface
   cd poon-web && npm install && npm run build && cd ..
   ```

4. **Start the system**
   ```bash
   # Terminal 1: Start gRPC server
   cd poon-server && ./poon-server
   
   # Terminal 2: Start Git server
   cd poon-git && ./poon-git
   
   # Terminal 3: Start web interface
   cd poon-web && npm start
   ```

5. **Access the system**
   - 🌐 Web interface: http://localhost:3000
   - 🔧 Git server: http://localhost:3000 (Git HTTP protocol)
   - 🚀 gRPC server: localhost:50051

## 💻 CLI Workflow

The Poon CLI provides a streamlined workflow for working with internet-scale monorepos:

### Creating a Workspace

```bash
# Create a workspace and track initial path
poon-cli start src/frontend

# This will:
# 1. Generate a UUID workspace name (e.g., "a1b2c3d4-e5f6-...")
# 2. Create server-side git repository with tracked files
# 3. Set up local workspace configuration
# 4. Initialize local git repository connected to poon-git server
```

### Managing Tracked Paths

```bash
# Check workspace status
poon-cli status

# Track additional directories
poon-cli track src/backend
poon-cli track docs

# Work with standard git workflow
git add .
git commit -m "Your changes"
git push

# Sync with latest monorepo state
poon-cli sync
```

### Key Features

- **UUID-based Workspace Names**: Server generates unique identifiers for workspaces
- **Server-side Git Repositories**: Workspaces are created with full git history in temporary directories
- **Selective Path Tracking**: Only download and track the directories you need
- **Standard Git Workflow**: Use familiar git commands for version control
- **Automatic Synchronization**: Keep your workspace in sync with the monorepo
- **Temporary Storage**: Workspace git repositories are stored in temporary directories for security

## 🧪 Testing

Each component includes comprehensive tests via standardized `run_test.sh` scripts:

```bash
# Test individual components
cd poon-server && ./scripts/run_test.sh
cd poon-git && ./scripts/run_test.sh
cd poon-cli && ./scripts/run_test.sh
cd poon-proto && ./scripts/run_test.sh
cd poon-web && ./scripts/run_test.sh

# Run integration tests
cd poon-tests && ./scripts/run_test.sh
```

### GitHub Actions

The project includes automated CI/CD workflows:

- **Presubmit Tests** - Run on every PR and push
- **Integration Tests** - Full system testing with all components
- **Release Builds** - Multi-platform binary releases

## 📚 Component Documentation

### poon-server
Core gRPC API server providing:
- Content-addressable storage with SHA-256 hashing
- Workspace creation with UUID generation in temporary directories
- Git repository initialization and file copying
- Directory listing and file reading
- Patch merging and conflict resolution
- Branch management operations
- File history and commit tracking

**Technology**: Go 1.23, gRPC, Protocol Buffers, Content-addressable storage

### poon-web  
Modern web interface featuring:
- Interactive file browser with breadcrumb navigation
- Real-time file viewing with syntax detection
- Responsive design with Tailwind CSS
- gRPC-Web client with fallback mock data

**Technology**: Next.js 15, React 19, TypeScript, Tailwind CSS v4

### poon-git
Git-compatible HTTP server providing:
- Standard Git protocol support (`git clone`, `git push`)
- Git HTTP endpoints (`/info/refs`, `/git-upload-pack`)
- REST API for directory and file access
- Sparse checkout capabilities
- Workspace repository serving

**Technology**: Go 1.23, HTTP server, Git protocol

### poon-cli
Command-line interface supporting:
- Workspace creation with server-side UUID generation
- Initial path tracking during workspace creation
- Directory tracking from monorepo
- Standard git workflow integration
- Push/sync operations with monorepo
- Direct gRPC server communication

**Technology**: Go 1.23, Cobra CLI framework

### poon-proto
Protocol Buffer definitions containing:
- gRPC service definitions
- Multi-language client generation (Go, TypeScript, Python)
- Consistent API contracts across components

**Technology**: Protocol Buffers, protoc

### poon-tests
Integration test suite providing:
- End-to-end workspace creation workflow testing
- Multi-component integration validation
- CLI command validation with new UUID-based workflow
- Error handling and recovery testing
- Server-side workspace creation verification

**Technology**: Go 1.23, Testify

## 🛠️ Development

### Code Standards

- **Go**: Standard formatting (`gofmt`), linting, and `go vet`
- **TypeScript**: ESLint with Next.js rules, strict type checking
- **Testing**: Comprehensive unit and integration tests
- **Documentation**: Inline code documentation and README files

### Project Structure

```
poon/
├── .github/workflows/     # GitHub Actions CI/CD
├── poon-server/          # gRPC API server
│   ├── scripts/          # Test scripts
│   ├── storage/          # Content-addressable storage implementation
│   ├── main.go          # Server implementation
│   └── server_test.go   # Unit tests
├── poon-web/            # Next.js web interface
│   ├── src/app/         # Next.js App Router
│   ├── src/components/  # React components
│   ├── src/proto/       # gRPC-Web client
│   └── scripts/         # Test scripts
├── poon-git/            # Git HTTP server
├── poon-cli/            # CLI tool
├── poon-proto/          # Protocol definitions
│   ├── monorepo.proto   # Service definitions
│   └── gen/             # Generated clients
└── poon-tests/          # Integration tests
    ├── testutil/        # Test utilities and helpers
    └── scripts/         # Test scripts
```

### Adding New Features

1. **Update Protocol Buffers** (if needed)
   ```bash
   cd poon-proto
   # Edit monorepo.proto
   npm run proto:generate:go
   ```

2. **Implement in poon-server**
   ```bash
   cd poon-server
   # Add gRPC handler
   ./scripts/run_test.sh  # Verify tests pass
   ```

3. **Update clients** (poon-web, poon-cli, poon-git)
   ```bash
   # Update client code to use new APIs
   # Add tests and verify functionality
   ```

4. **Add integration tests**
   ```bash
   cd poon-tests
   # Add end-to-end test scenarios
   ./scripts/run_test.sh
   ```

## 🐛 Troubleshooting

### Common Issues

1. **gRPC connection errors**
   - Ensure poon-server is running on port 50051
   - Check firewall settings

2. **Web interface not loading**
   - Verify Next.js build completed successfully
   - Check for TypeScript compilation errors

3. **Git operations failing**
   - Ensure poon-git server is running on port 3000
   - Verify poon-server is accessible

4. **Test failures**
   - Run `./scripts/run_test.sh` in each component directory
   - Check protobuf files are generated correctly

### Debug Mode

Enable verbose logging by setting environment variables:

```bash
# Enable gRPC debug logging
export GRPC_GO_LOG_VERBOSITY_LEVEL=99
export GRPC_GO_LOG_SEVERITY_LEVEL=info

# Enable Next.js debug mode
export DEBUG=next:*

# Run components with debug output
```

## 🤝 Contributing

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/new-feature`
3. **Make your changes** with proper tests
4. **Run all tests**: `./scripts/run_test.sh` in each component
5. **Ensure formatting**: `gofmt -w .` for Go, `npm run lint` for TypeScript
6. **Submit a pull request** with detailed description

### Pull Request Requirements

- [ ] All tests pass (`scripts/run_test.sh` in each component)
- [ ] Code is properly formatted and linted
- [ ] New features include tests
- [ ] Documentation is updated
- [ ] Integration tests pass

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🎯 Roadmap

- [ ] **Performance Optimization** - Implement caching and connection pooling
- [ ] **Security Enhancements** - Add authentication and authorization
- [ ] **Scalability** - Support for distributed deployments
- [ ] **Plugin System** - Extensible architecture for custom workflows
- [ ] **Monitoring** - Comprehensive metrics and observability
- [ ] **Documentation** - Interactive API documentation

---

**Built with ❤️ using Go, Next.js, gRPC, and Protocol Buffers**

For questions, issues, or contributions, please visit our [GitHub repository](https://github.com/nic/poon).