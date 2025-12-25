# Contributing to mcp-go

Thank you for your interest in contributing to mcp-go! This document provides guidelines and information about contributing.

## How to Contribute

### Reporting Bugs

Before creating a bug report, please check existing issues to avoid duplicates. When creating a bug report, include:

- A clear, descriptive title
- Steps to reproduce the issue
- Expected behavior vs actual behavior
- Go version (`go version`)
- Operating system and version
- Relevant code snippets or error messages

### Suggesting Features

Feature requests are welcome! Please include:

- A clear description of the feature
- The motivation/use case
- Example code showing how it might work
- Any relevant MCP specification references

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Write tests** for any new functionality
3. **Run the test suite** to ensure nothing is broken
4. **Update documentation** if needed
5. **Follow the code style** (enforced by golangci-lint)
6. **Write clear commit messages** following conventional commits

## Development Setup

### Prerequisites

- Go 1.21 or later
- Make (optional, for convenience commands)

### Getting Started

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/mcp-go.git
cd mcp-go

# Install dependencies
go mod download

# Run tests
go test ./...

# Run linter
golangci-lint run
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. -benchmem ./...

# Run specific package tests
go test ./middleware/...
```

### Code Style

We use [golangci-lint](https://golangci-lint.run/) for code quality. Key standards:

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Write descriptive variable and function names
- Add comments for exported functions and types
- Keep functions focused and small

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Maintenance tasks

Examples:
```
feat(middleware): add rate limiting middleware
fix(transport): handle connection timeouts properly
docs: update README with new examples
test(server): add integration tests for prompts
```

## Project Structure

```
mcp-go/
├── mcp.go              # Main package, public API
├── middleware/         # Request/response middleware
├── protocol/           # JSON-RPC message types
├── schema/             # JSON Schema generation
├── server/             # Core server implementation
├── transport/          # Stdio and HTTP transports
├── examples/           # Example servers
└── e2e/                # End-to-end compliance tests
```

## Testing Guidelines

- Write table-driven tests where appropriate
- Use subtests for related test cases
- Mock external dependencies
- Aim for >80% coverage on new code
- Include both positive and negative test cases

Example:
```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "test", "result", false},
        {"empty input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Feature(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("got = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Review Process

1. All PRs require at least one approval
2. CI must pass (tests, linting, build)
3. Coverage should not decrease significantly
4. Documentation must be updated if needed

## Getting Help

- Open an issue for questions
- Check existing issues and discussions
- Read the [MCP specification](https://spec.modelcontextprotocol.io/)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
