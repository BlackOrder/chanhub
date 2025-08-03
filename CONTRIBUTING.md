# Contributing to ChanHub

Thank you for your interest in contributing to ChanHub! We welcome contributions from the community and are pleased to have you join us.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Code Style](#code-style)
- [Submitting Changes](#submitting-changes)
- [Reporting Issues](#reporting-issues)
- [Community Guidelines](#community-guidelines)

## Getting Started

### Prerequisites

- Go 1.22 or later
- Git
- golangci-lint (for code quality checks)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/your-username/chanhub.git
   cd chanhub
   ```

3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/BlackOrder/chanhub.git
   ```

## Development Setup

1. Ensure you have Go installed and properly configured
2. Install golangci-lint:
   ```bash
   # On Linux/macOS
   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.64.8
   
   # Or using Go install
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
   ```

3. Verify your setup:
   ```bash
   go version
   golangci-lint version
   ```

## Making Changes

### Development Workflow

1. Create a new branch for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following the guidelines below
3. Test your changes thoroughly
4. Commit your changes with a clear commit message
5. Push to your fork and create a pull request

### Code Organization

The project structure is simple:
- `hub.go` - Core implementation
- `hub_test.go` - Test suite
- `doc.go` - Package documentation
- `README.md` - Main documentation
- `EXAMPLES.md` - Usage examples

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -race -coverprofile=coverage.out -covermode=atomic ./...

# View coverage report
go tool cover -html=coverage.out
```

### Writing Tests

- All new functionality must include comprehensive tests
- Aim for 100% code coverage
- Include both unit tests and integration tests where appropriate
- Test concurrent scenarios and edge cases
- Use table-driven tests for multiple similar test cases

Example test structure:
```go
func TestNewFeature(t *testing.T) {
    // Arrange
    hub := New()
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Act
    result := hub.NewFeature(ctx)

    // Assert
    if result == nil {
        t.Fatal("NewFeature() returned nil")
    }
}
```

### Benchmarks

Include benchmarks for performance-critical code:
```go
func BenchmarkNewFeature(b *testing.B) {
    hub := New()
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        hub.NewFeature(ctx)
    }
}
```

## Code Style

### Go Formatting

- Use `gofmt` to format your code
- Follow standard Go conventions
- Use meaningful variable and function names
- Add comments for exported functions and complex logic

### Linting

We use golangci-lint with a comprehensive configuration. Before submitting:

```bash
# Run the linter
golangci-lint run

# Auto-fix some issues
golangci-lint run --fix
```

### Code Quality Guidelines

1. **Simplicity**: Keep code simple and readable
2. **Error Handling**: Handle errors appropriately
3. **Thread Safety**: Ensure concurrent safety where needed
4. **Performance**: Consider performance implications
5. **Documentation**: Comment exported functions and complex logic

### Function Documentation

Document all exported functions:
```go
// Subscribe creates a new subscription to the hub that will receive signals
// when Broadcast is called. The subscription is automatically cleaned up
// when the provided context is canceled.
//
// The returned channel is buffered and will receive a signal each time
// Broadcast() is called, until the context is canceled.
func (h *Hub) Subscribe(ctx context.Context) <-chan struct{} {
    // implementation
}
```

## Submitting Changes

### Pull Request Process

1. **Update Documentation**: Ensure README.md and other docs are updated if needed
2. **Add Tests**: Include tests for new functionality
3. **Check CI**: Ensure all CI checks pass
4. **Rebase**: Rebase your branch on the latest main branch
5. **Clear Description**: Provide a clear description of your changes

### Pull Request Template

When creating a pull request, include:

- **What**: Brief description of the changes
- **Why**: Explanation of why the changes are needed
- **How**: Description of how the changes work
- **Testing**: Details about testing performed
- **Breaking Changes**: Any breaking changes and migration notes

### Commit Message Format

Use clear, concise commit messages:
```
type(scope): brief description

Longer description if needed, explaining the what and why.

Fixes #123
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Maintenance tasks

## Reporting Issues

### Bug Reports

When reporting bugs, include:
1. **Go version** and operating system
2. **Expected behavior** vs **actual behavior**
3. **Minimal reproduction case**
4. **Error messages** and stack traces
5. **Additional context** that might be helpful

### Feature Requests

For feature requests, include:
1. **Use case**: Describe the problem you're trying to solve
2. **Proposed solution**: Your suggested approach
3. **Alternatives**: Other solutions you've considered
4. **Additional context**: Any other relevant information

## Community Guidelines

### Code of Conduct

We follow the [Go Community Code of Conduct](https://golang.org/conduct). Please be respectful and inclusive in all interactions.

### Communication

- **Issues**: Use GitHub issues for bug reports and feature requests
- **Discussions**: Use GitHub discussions for questions and ideas
- **Pull Requests**: Use PR comments for code review discussions

### Review Process

1. All submissions require review before merging
2. We aim to review PRs within a few days
3. Address review feedback promptly
4. Multiple rounds of review may be needed

### Acceptance Criteria

For a contribution to be accepted:
- [ ] All tests pass
- [ ] Code coverage is maintained or improved
- [ ] Linting passes without errors
- [ ] Documentation is updated if needed
- [ ] Code follows project conventions
- [ ] PR has clear description and reasoning

## Getting Help

If you need help or have questions:
1. Check existing issues and documentation
2. Create a GitHub issue with your question
3. Be specific about what you're trying to achieve

Thank you for contributing to ChanHub! Your contributions help make this project better for everyone.
