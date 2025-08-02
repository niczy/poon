# Contributing to Poon Monorepo System

Thank you for your interest in contributing to Poon! This guide will help you get started with development and ensure your contributions align with the project's standards.

## 🚀 Getting Started

### Development Environment Setup

1. **Prerequisites**
   - Go 1.23 or later
   - Node.js 20 or later
   - Protocol Buffers compiler (`protoc`)
   - Git

2. **Clone and Setup**
   ```bash
   git clone https://github.com/nic/poon.git
   cd poon
   
   # Install Go protobuf plugins
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   
   # Generate protobuf files
   cd poon-proto
   npm install
   npm run proto:generate:go
   cd ..
   ```

3. **Build and Test**
   ```bash
   # Test all components
   for component in poon-git poon-server poon-cli poon-proto poon-tests poon-web; do
     cd $component
     ./scripts/run_test.sh
     cd ..
   done
   ```

## 📋 Development Workflow

### Making Changes

1. **Fork the repository** on GitHub
2. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes**
   - Follow the coding standards (see below)
   - Add tests for new functionality
   - Update documentation as needed

4. **Test your changes**
   ```bash
   # Run tests for affected components
   cd affected-component
   ./scripts/run_test.sh
   ```

5. **Commit and push**
   ```bash
   git add .
   git commit -m "Add feature: description of your changes"
   git push origin feature/your-feature-name
   ```

6. **Create a Pull Request** on GitHub

### Code Standards

#### Go Code

- **Formatting**: Use `gofmt` for consistent formatting
- **Linting**: Code should pass `go vet` without warnings
- **Testing**: Include unit tests for new functionality
- **Documentation**: Add comments for exported functions and types

```bash
# Before committing Go code
gofmt -w .
go vet ./...
go test ./...
```

#### TypeScript/JavaScript Code

- **Linting**: Use ESLint with Next.js recommended rules
- **Formatting**: Consistent with project Prettier config
- **Type Safety**: Strict TypeScript mode enabled
- **Testing**: Include tests for new components

```bash
# Before committing TypeScript code
npm run lint
npx tsc --noEmit
```

#### Protocol Buffers

- **Style**: Follow [Protocol Buffers Style Guide](https://developers.google.com/protocol-buffers/docs/style)
- **Documentation**: Include comments for services and messages
- **Compatibility**: Maintain backward compatibility when possible

## 🧪 Testing Guidelines

### Test Structure

Each component has standardized testing via `scripts/run_test.sh`:

- **Unit Tests**: Component-specific functionality
- **Integration Tests**: Cross-component interactions
- **Formatting Tests**: Code style validation
- **Build Tests**: Compilation and build verification

### Writing Tests

#### Go Tests

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    Input
        expected Expected
    }{
        {"test case 1", input1, expected1},
        {"test case 2", input2, expected2},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := YourFunction(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

#### Integration Tests

Add integration tests to `poon-tests/` for end-to-end scenarios:

```go
func TestNewWorkflow(t *testing.T) {
    // Setup test environment
    server := testutil.StartTestServer(t)
    defer server.Stop()
    
    // Test workflow steps
    // ...
    
    // Verify results
}
```

## 📝 Documentation

### Code Documentation

- **Go**: Use standard Go doc comments
- **TypeScript**: Use JSDoc comments for complex functions
- **README**: Update component READMEs for significant changes

### API Documentation

When adding new gRPC services:

1. **Update `poon-proto/monorepo.proto`** with detailed comments
2. **Regenerate clients** with `npm run proto:generate:go`
3. **Update README** with new API usage examples

## 🔄 Pull Request Process

### PR Requirements

- [ ] **Tests Pass**: All `scripts/run_test.sh` must pass
- [ ] **Code Formatted**: Go code formatted with `gofmt`, TS linted
- [ ] **Documentation Updated**: READMEs and code comments updated
- [ ] **Backwards Compatible**: No breaking changes without discussion
- [ ] **Descriptive PR**: Clear description of changes and reasoning

### PR Template

```markdown
## Description
Brief description of changes and motivation.

## Type of Change
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] All tests pass locally

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes (or breaking changes documented)
```

### Review Process

1. **Automated Checks**: GitHub Actions will run all tests
2. **Code Review**: Maintainers will review for:
   - Code quality and style
   - Test coverage
   - Documentation completeness
   - Architecture alignment

3. **Approval**: At least one maintainer approval required
4. **Merge**: Squash and merge preferred for clean history

## 🐛 Bug Reports

### Before Reporting

1. **Search existing issues** for similar problems
2. **Test with latest version** of the main branch
3. **Isolate the issue** with minimal reproduction steps

### Bug Report Template

```markdown
## Bug Description
Clear description of what the bug is.

## To Reproduce
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

## Expected Behavior
What you expected to happen.

## Environment
- Component: [poon-server/poon-web/etc.]
- Go version: [if applicable]
- Node.js version: [if applicable]
- OS: [e.g., Ubuntu 20.04, macOS 13.0]

## Additional Context
Any other context about the problem.
```

## 💡 Feature Requests

### Before Requesting

1. **Check existing issues** for similar requests
2. **Consider the scope** - does it fit the project's goals?
3. **Think about implementation** - how would it work?

### Feature Request Template

```markdown
## Feature Description
Clear description of the feature you'd like to see.

## Motivation
Why is this feature needed? What problem does it solve?

## Proposed Solution
How do you envision this working?

## Alternatives Considered
Other approaches you've considered.

## Additional Context
Any other context or screenshots.
```

## 🏷️ Component-Specific Guidelines

### poon-server (Go gRPC Server)

- **Performance**: Consider memory and CPU impact
- **Error Handling**: Proper gRPC error codes
- **Logging**: Structured logging with context
- **Concurrency**: Thread-safe implementations

### poon-web (Next.js Frontend)

- **Performance**: Optimize bundle size and loading
- **Accessibility**: Follow WCAG guidelines
- **Mobile**: Responsive design required
- **SEO**: Proper meta tags and structure

### poon-git (Git HTTP Server)

- **Compatibility**: Follow Git protocol standards
- **Security**: Validate all inputs
- **Performance**: Handle large repositories efficiently

### poon-cli (Command Line Tool)

- **UX**: Clear, consistent command interface
- **Error Messages**: Helpful, actionable error messages
- **Cross-platform**: Work on Linux, macOS, Windows

## 📞 Getting Help

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Discord**: [Community chat] (if available)

## 🎯 Good First Issues

Look for issues labeled `good first issue` - these are designed to be approachable for new contributors and help you get familiar with the codebase.

---

Thank you for contributing to Poon! Your efforts help make this project better for everyone. 🙏