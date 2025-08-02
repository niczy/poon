# CI Setup Guide for Poon Monorepo

This guide explains how to set up CI/CD for the Poon monorepo system to avoid the "protoc-gen-go: program not found" error.

## Quick Fix for CI

The error occurs because the `protoc-gen-go` and `protoc-gen-go-grpc` tools are not in the PATH. Here are the solutions:

### Option 1: Use Makefile Targets (Recommended)

```yaml
name: CI Build and Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.23'
    
    - name: Set up Node.js
      uses: actions/setup-node@v4
      with:
        node-version: '18'
    
    - name: Install protoc
      run: |
        sudo apt-get update
        sudo apt-get install -y protobuf-compiler
    
    # Use Makefile targets (handles tools automatically)
    - name: Build and test
      run: |
        make ci-setup
        make ci-build
        make ci-test
```

### Option 2: Manual Tool Installation

```yaml
steps:
- uses: actions/checkout@v4

- name: Set up Go
  uses: actions/setup-go@v4
  with:
    go-version: '1.23'

- name: Install protoc
  run: |
    sudo apt-get update
    sudo apt-get install -y protobuf-compiler

- name: Install protoc-gen-go tools
  run: |
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    echo "$HOME/go/bin" >> $GITHUB_PATH

- name: Test specific component
  run: |
    cd poon-proto
    chmod +x scripts/run_test.sh
    ./scripts/run_test.sh
```

### Option 3: Component-Specific Testing

```yaml
strategy:
  matrix:
    component: [poon-git, poon-server, poon-cli, poon-proto, poon-web]

steps:
- uses: actions/checkout@v4

- name: Set up Go
  uses: actions/setup-go@v4
  with:
    go-version: '1.23'
  if: matrix.component != 'poon-web'

- name: Set up Node.js
  uses: actions/setup-node@v4
  with:
    node-version: '18'
  if: matrix.component == 'poon-web'

- name: Install protoc
  run: |
    sudo apt-get update
    sudo apt-get install -y protobuf-compiler

- name: Set up tools and test
  run: |
    make install-protoc-tools
    make proto
    cd ${{ matrix.component }}
    chmod +x scripts/run_test.sh
    ./scripts/run_test.sh
```

## Available Make Targets

- `make ci-setup` - Install all dependencies for CI
- `make ci-build` - Build all components with proper PATH
- `make ci-test` - Run all tests with proper PATH
- `make install-protoc-tools` - Just install protoc-gen-go tools
- `make proto` - Generate protobuf files (auto-installs tools)
- `make build` - Build all components
- `make test` - Run all tests

## Key Points

1. **Always install protoc-gen-go tools** before running protobuf generation
2. **Add Go bin to PATH** using `echo "$HOME/go/bin" >> $GITHUB_PATH`
3. **Use Makefile targets** for consistent tool management
4. **Install protoc compiler** with `sudo apt-get install protobuf-compiler`

## Troubleshooting

If you still get "protoc-gen-go: program not found":

1. Check that `protoc-gen-go` is installed: `which protoc-gen-go`
2. Check PATH includes Go bin: `echo $PATH | grep go/bin`
3. Manually add to PATH: `export PATH="$PATH:$(go env GOPATH)/bin:$HOME/go/bin"`

## Example Working CI Configuration

```yaml
name: Build and Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: '1.23'
    - uses: actions/setup-node@v4
      with:
        node-version: '18'
    
    - name: Install system dependencies
      run: |
        sudo apt-get update
        sudo apt-get install -y protobuf-compiler
    
    - name: Install Go protobuf tools
      run: |
        go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
        go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
        echo "$HOME/go/bin" >> $GITHUB_PATH
    
    - name: Install Node dependencies
      run: |
        cd poon-proto && npm ci
        cd ../poon-web && npm ci
    
    - name: Generate protobuf files
      run: make proto
    
    - name: Build all components
      run: make build
    
    - name: Run tests
      run: make test
```

This configuration ensures all tools are properly installed and available in PATH before any protobuf generation occurs.