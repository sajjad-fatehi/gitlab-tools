# Contributing to GitLab Tools

Thank you for your interest in contributing to GitLab Tools!

## Development Setup

### Prerequisites

- Go 1.21 or later
- Access to a self-hosted GitLab instance for testing
- GitLab personal access token with `api` scope

### Getting Started

```bash
# Clone the repository
git clone <repository-url>
cd gitlab-tools

# Install dependencies
make deps

# Build the project
make build

# Run tests
make test
```

## Project Structure

```text
gitlab-tools/
├── cmd/
│   └── gitlab-tools/
│       └── main.go           # CLI entry point and command handlers
├── internal/
│   ├── gitlab/
│   │   ├── client.go         # GitLab API client implementation
│   │   ├── types.go          # Domain models and types
│   │   └── types_test.go     # Unit tests for types
│   └── bulkmr/
│       ├── service.go        # Bulk MR creation business logic
│       └── service_test.go   # Service tests with mocks
├── go.mod                    # Go module definition
├── Makefile                  # Build automation
└── README.md                 # User documentation
```

## Code Style and Guidelines

### Go Conventions

- Follow standard Go formatting (use `gofmt` or `goimports`)
- Use meaningful variable names (avoid 1-2 character names)
- Functions should be verbs, variables should be nouns
- Keep functions small and focused (single responsibility)
- Add comments for non-obvious logic

### Error Handling

- Always handle errors explicitly
- Return wrapped errors with context: `fmt.Errorf("context: %w", err)`
- Use structured logging for operational messages
- Include helpful error messages for end users

### Testing

- Write unit tests for new business logic
- Use table-driven tests for multiple scenarios
- Mock external dependencies (GitLab API)
- Aim for high test coverage of critical paths

## Adding New Features

### Adding a New Command

1. **Create the command handler** in `cmd/gitlab-tools/main.go`
2. **Add business logic** to a new package in `internal/`
3. **Add tests** for the new functionality
4. **Update documentation** in README.md

Example structure:

```go
// In cmd/gitlab-tools/main.go
case "new-command":
    newCommand()

func newCommand() {
    // Parse flags
    // Create service
    // Execute and display results
}
```

### Adding GitLab API Methods

Add new methods to `internal/gitlab/client.go`:

```go
func (c *Client) NewMethod(params) (Result, error) {
    endpoint := fmt.Sprintf("%s/api/v4/...", c.baseURL)
    
    var result Result
    if err := c.doRequest("GET", endpoint, nil, &result); err != nil {
        return nil, fmt.Errorf("operation failed: %w", err)
    }
    
    return result, nil
}
```

### Extending the Service Layer

Business logic goes in `internal/*/service.go` files:

```go
type Service struct {
    client GitLabClient
    config Config
}

func (s *Service) ProcessSomething() ([]Result, Summary) {
    // Implementation
}
```

## Testing Guidelines

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -v -run TestName ./internal/bulkmr

# Run tests with verbose output
go test -v ./...
```

### Writing Tests

Use table-driven tests:

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    Input
        expected Output
    }{
        {
            name:     "scenario description",
            input:    Input{...},
            expected: Output{...},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := ProcessFeature(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Mocking

Use interface-based mocks for external dependencies:

```go
type MockClient struct {
    // Mock state
}

func (m *MockClient) Method(params) (Result, error) {
    // Mock implementation
}
```

## Documentation

### Code Comments

- Add package-level comments explaining purpose
- Document exported functions, types, and constants
- Explain "why" not "what" for complex logic
- Use complete sentences in comments

### User Documentation

When adding features, update:

- `README.md` - Main usage documentation
- `QUICKSTART.md` - Quick start examples
- Command help text in `main.go`

## Pull Request Process

1. **Create a feature branch**: `git checkout -b feature/your-feature`
2. **Write your code** following the guidelines above
3. **Add tests** for new functionality
4. **Update documentation** as needed
5. **Run tests**: `make test`
6. **Build successfully**: `make build`
7. **Commit with clear messages**: Use conventional commit format
8. **Submit PR** with description of changes

### Commit Message Format

```text
type(scope): brief description

Detailed explanation of changes if needed.

- Bullet points for multiple changes
- Include issue references if applicable
```

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`

## Code Review Checklist

Before submitting:

- [ ] All tests pass (`make test`)
- [ ] Code builds successfully (`make build`)
- [ ] New functionality has tests
- [ ] Documentation is updated
- [ ] Code follows project conventions
- [ ] Error handling is comprehensive
- [ ] No secrets or credentials in code

## Future Enhancement Ideas

Potential features to work on:

- Parallel processing of projects
- Configuration file support (YAML/JSON)
- Batch MR updates (labels, assignees)
- Branch cleanup commands
- MR status reporting
- Pipeline management
- Retry logic with exponential backoff
- Progress bar for bulk operations

## Getting Help

- Review existing code for patterns and examples
- Check the README for usage documentation
- Look at test files for usage examples
- Ask questions in issues or discussions

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
