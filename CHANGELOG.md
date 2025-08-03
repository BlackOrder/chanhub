# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of chanhub package
- Core Hub functionality with Subscribe and Broadcast methods
- Context-aware automatic cleanup
- Thread-safe operations with read-write mutex
- Non-blocking broadcast implementation
- Comprehensive test suite with 100% code coverage
- Benchmark tests for performance validation
- Complete documentation with examples
- CI/CD pipeline with GitHub Actions
- Linting configuration with golangci-lint
- Security scanning with gosec

### Features
- **New()** - Creates a new Hub instance
- **Subscribe(ctx)** - Subscribe to the hub with context-based cleanup
- **Broadcast()** - Send signals to all subscribers non-blockingly

### Documentation
- README.md with comprehensive usage guide
- EXAMPLES.md with real-world usage scenarios
- Package documentation with detailed API descriptions
- Code examples for common patterns

### Testing
- Unit tests covering all functionality
- Race condition testing
- Concurrent operation testing
- Benchmark tests for performance validation
- 100% code coverage

### Quality Assurance
- golangci-lint configuration with comprehensive rules
- GitHub Actions CI/CD pipeline
- Automated testing across multiple Go versions (1.22.x, 1.23.x, 1.24.x)
- Code coverage reporting
- Security scanning

## [1.0.0] - TBD

### Added
- Initial stable release
- Full API documentation
- Production-ready implementation

## Release Notes

### Version 1.0.0
This is the initial stable release of chanhub. The package provides a simple,
efficient, and thread-safe way to implement pub/sub patterns in Go applications.

Key features include:
- Zero dependencies (uses only Go standard library)
- Automatic cleanup with context cancellation
- Non-blocking broadcasts
- High performance with minimal overhead
- Comprehensive test coverage
- Production-ready implementation

The API is considered stable and will follow semantic versioning for future changes.
