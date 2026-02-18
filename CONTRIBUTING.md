# Contributing to Go-Garage

Thank you for considering contributing to Go-Garage! This document provides guidelines and instructions for contributing to the project.

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on constructive feedback
- Accept that disagreement happens and work toward resolution

## Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

- Go 1.24 or later
- SQLite3
- Make (optional, but recommended)
- Docker and Docker Compose (for containerized development)
- Git

### Setting Up Your Development Environment

1. **Fork the repository** on GitHub

2. **Clone your fork**:

   ```shell
   git clone https://github.com/YOUR_USERNAME/go-garage.git
   cd go-garage
   ```

3. **Add the upstream repository**:

   ```shell
   git remote add upstream https://github.com/truggeri/go-garage.git
   ```

4. **Install dependencies**:

   ```shell
   go mod download
   ```

5. **Set up pre-commit hooks**:

   ```shell
   cp .githooks/pre-commit .git/hooks/pre-commit
   chmod +x .git/hooks/pre-commit
   ```

6. **Copy the environment file**:

   ```shell
   cp .env.example .env
   ```

7. **Install development tools** (optional):

   ```shell
   make install-tools
   ```

## Development Workflow

### Creating a Feature Branch

Always create a new branch for your work:

```shell
git checkout -b feature/your-feature-name
```

Branch naming conventions:

- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Adding or updating tests

### Making Changes

1. **Write code** following the [coding standards](#coding-standards)

2. **Format your code**:

   ```shell
   make fmt
   ```

3. **Run tests**:

   ```shell
   make test
   ```

4. **Run linters**:

   ```shell
   make lint
   make vet
   ```

5. **Build the application**:

   ```shell
   make build
   ```

### Running the Application

Test your changes by running the application:

```shell
# Run directly
make run

# Or run with Docker
docker compose up --build
```

## Coding Standards

### Go Style Guidelines

- Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` for code formatting (run `make fmt`)
- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Keep functions small and focused on a single responsibility
- Use meaningful variable and function names

### Code Organization

- Place new features in appropriate packages:
  - `internal/handlers/` - HTTP handlers
  - `internal/services/` - Business logic
  - `internal/repositories/` - Data access layer
  - `internal/models/` - Data structures
  - `internal/middleware/` - HTTP middleware
  - `pkg/` - Public, reusable packages

### Error Handling

- Always handle errors explicitly
- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Never ignore errors (avoid `_` for error returns)
- Return errors as the last return value

### Tests

- Write tests for all new code
- Aim for minimum 80% code coverage
- Use table-driven tests where appropriate
- Test both success and error cases
- Use meaningful test names: `TestFunctionName_Scenario_ExpectedBehavior`

Example test structure:

```go
func TestVehicleService_Create_ValidInput_Success(t *testing.T) {
    // Arrange
    // ... setup

    // Act
    // ... call function

    // Assert
    // ... verify results
}
```

### Documentation

- Add package comments for all packages
- Document all exported functions, types, and methods
- Use GoDoc format
- Keep comments concise and meaningful
- Update documentation when changing functionality

### Security

- Never commit secrets, API keys, or credentials
- Validate all user input
- Use parameterized queries for database operations
- Escape output in HTML templates
- Follow OWASP security best practices

## Commit Guidelines

### Commit Messages

Write clear, descriptive commit messages:

```text
Short summary (50 chars or less)

More detailed explanation if needed (wrap at 72 chars).
Explain the problem this commit solves and why you chose
this solution.

Fixes #123
```

**Good commit messages**:

- `Add vehicle creation endpoint`
- `Fix panic in maintenance record handler`
- `Refactor database connection management`
- `Update README with Docker instructions`

**Bad commit messages**:

- `fix bug`
- `update`
- `wip`

### Commits Should Be

- **Atomic**: Each commit should represent one logical change
- **Tested**: Code should pass all tests
- **Formatted**: Code should be properly formatted
- **Documented**: Include necessary documentation updates

## Pull Request Process

### Before Submitting

1. **Sync with upstream**:

   ```shell
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run all checks**:

   ```shell
   make fmt
   make vet
   make lint
   make test
   ```

3. **Ensure build succeeds**:

   ```shell
   make build
   ```

4. **Update documentation** if you changed:
   - Configuration options
   - API endpoints
   - Command-line flags
   - Environment variables

### Submitting a Pull Request

1. **Push your branch**:

   ```shell
   git push origin feature/your-feature-name
   ```

2. **Create a Pull Request** on GitHub with:
   - Clear title describing the change
   - Description of what changed and why
   - Reference to related issues (e.g., "Fixes #123")
   - Screenshots for UI changes (if applicable)

3. **PR Description Template**:

   ```markdown
   ## Description
   Brief description of the changes

   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update

   ## Testing
   - [ ] Unit tests added/updated
   - [ ] Manual testing completed
   - [ ] All tests pass

   ## Checklist
   - [ ] Code follows project style guidelines
   - [ ] Self-review completed
   - [ ] Comments added for complex logic
   - [ ] Documentation updated
   - [ ] No new warnings generated
   ```

### Review Process

- Maintainers will review your PR
- Address feedback and requested changes
- Keep your PR up to date with the main branch
- Be patient and responsive to comments
- Once approved, a maintainer will merge your PR

## Testing

### Running Tests

```shell
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./internal/handlers/...

# Run a specific test
go test -run TestFunctionName ./path/to/package
```

### Writing Tests

- Place tests in the same package as the code
- Use `_test.go` suffix for test files
- Use the `testify` assertion library
- Mock external dependencies
- Test edge cases and error conditions

## Reporting Issues

### Bug Reports

Include:

- Clear, descriptive title
- Steps to reproduce
- Expected behavior
- Actual behavior
- Environment details (Go version, OS, etc.)
- Relevant logs or error messages

### Feature Requests

Include:

- Clear description of the feature
- Use case and motivation
- Proposed implementation (if you have ideas)
- Any alternatives you've considered

## Questions and Support

- Open an issue for questions about the project
- Check existing issues before creating a new one
- Tag questions with the `question` label

## License

By contributing to Go-Garage, you agree that your contributions will be licensed under the MIT License.

## Recognition

Contributors will be recognized in:

- Release notes for significant contributions
- The project's contributor list

Thank you for contributing to Go-Garage! 🚗
