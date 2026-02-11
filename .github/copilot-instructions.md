# Copilot Instructions for Go-Garage

## Project Overview

Go-Garage is a vehicle management web application written in Go that helps users track and manage their vehicles, maintenance records, and fuel consumption.

## Skills

This project has custom skills in `.github/skills/` that provide detailed instructions for specific tasks:

| Skill | Description |
|-------|-------------|
| `pr-linting-formatting` | Instructions for running linting and formatting tools when authoring Pull Requests |

**Always check for and use relevant skills when working on this codebase.**

## Key Resources

Before making changes, use code search to explore:
- **Specifications**: Check `/spec/` for requirements, architecture, and API documentation
- **Existing Patterns**: Search for similar implementations in the codebase
- **Tests**: Review existing tests to understand expected behavior

## Technology Stack

- **Language**: Go 1.24+
- **Web Framework**: Standard library `net/http` with `gorilla/mux`
- **Database**: SQLite with `database/sql`
- **Frontend**: Go `html/template` with htmx
- **Authentication**: JWT tokens
- **Testing**: `testify` toolkit

## Code Conventions

### Naming
- `PascalCase` for exported, `camelCase` for unexported
- Descriptive names: `vehicleRepository` not `vr`
- Test pattern: `TestFunctionName_Scenario_ExpectedBehavior`

### Error Handling
- Always handle errors explicitly
- Wrap with context: `fmt.Errorf("context: %w", err)`
- Use custom error types for domain-specific errors

### Database
- Use parameterized queries (prevent SQL injection)
- Use transactions for multi-table operations
- Always `defer` closing connections and rows

### Security
- Validate all user input
- Escape HTML template output
- Use prepared statements for SQL

## Domain Model

- **User** → owns multiple **Vehicles**
- **Vehicle** → has multiple **Maintenance Records** and **Fuel Records**

## Working with This Codebase

When asked to make changes:
1. Search for existing patterns and similar implementations first
2. Review related specs in `/spec/` directory
3. Follow established conventions found in the codebase
4. Include appropriate tests following existing test patterns

When exploring the codebase:
- Use semantic search to understand how features are implemented
- Use lexical search to find specific symbols, functions, or patterns
- Check `/spec/README.md` for documentation index

## Pull Request Requirements

### Updating Spec Documents

Pull requests **must** update the relevant spec document(s) in `/spec/` to reflect completed work. The milestone documents use checkboxes to track progress:

| Status | Syntax | Meaning |
|--------|--------|---------|
| ⬜ | `- [ ]` | Not started |
| ✅ | `- [x]` | Complete |

When completing a task:
1. Find the corresponding checkbox item in the relevant milestone document
2. Change `- [ ]` to `- [x]`
3. Add a PR reference: `- [x] Task description (PR #123)`

**Example:**
```markdown
### Before (in spec/milestone-2-data-layer.md)
- [ ] Configure database connection with connection pooling

### After
- [x] Configure database connection with connection pooling (PR #24)
```

### PR Checklist
- [ ] Code follows project conventions
- [ ] Tests added/updated for changes
- [ ] Linting and formatting pass (`make fmt`, `make lint`, `make vet`)
- [ ] Relevant spec document updated with completion status
- [ ] PR description references the spec/milestone being addressed
