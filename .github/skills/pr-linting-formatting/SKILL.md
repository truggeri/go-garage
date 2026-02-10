---
name: pr-linting-formatting
description: Instructions for running linting and formatting tools when authoring Pull Requests. Use this skill when creating or modifying code in pull requests to ensure code quality standards are met.
---

# Pull Request Linting and Formatting

When authoring Pull Requests for this Go project, always run linting and formatting tools and fix any errors before finalizing the PR.

## Required Tools

This project uses:
- **gofmt** - Go code formatter (with simplify flag)
- **goimports** - Go imports organizer
- **golangci-lint** - Comprehensive Go linter with multiple analyzers

## Formatting Steps

### 1. Format Code

Run the formatting command to automatically fix code style issues:

```bash
make fmt
```

This runs `gofmt -s -w .` which:
- Formats all Go files according to Go standards
- Simplifies code where possible (-s flag)
- Writes changes directly to files (-w flag)

**Always commit formatting changes separately** from logical changes when possible.

### 2. Run Linting

After formatting, run the linter to catch potential issues:

```bash
make lint
```

This runs `golangci-lint run` using the configuration in `.golangci.yml`.

### 3. Run Go Vet

Run Go's built-in static analysis tool:

```bash
make vet
```

This runs `go vet ./...` to identify suspicious constructs.

## Enabled Linters

The project has the following linters enabled (see `.golangci.yml`):
- **govet** - Official Go static analyzer (with most checks enabled except fieldalignment)
- **ineffassign** - Detects ineffectual assignments
- **staticcheck** - Advanced Go linter
- **unused** - Finds unused code
- **misspell** - Finds commonly misspelled words
- **unparam** - Finds unused function parameters
- **unconvert** - Removes unnecessary type conversions
- **goconst** - Finds repeated strings that could be constants

## Formatters Configuration

- **gofmt**: Enabled with simplify mode
- **goimports**: Enabled for import organization

## Fixing Errors

### Common Formatting Issues
- **Import organization**: goimports will automatically group and sort imports
- **Indentation**: gofmt handles all indentation automatically
- **Spacing**: gofmt normalizes spacing around operators and keywords

### Common Linting Issues

1. **Unused variables/imports**: Remove or use them, or prefix with `_` if intentionally unused
2. **Ineffectual assignments**: Remove assignments that don't affect program behavior
3. **Misspellings**: Fix typos in comments and strings
4. **Unused parameters**: Remove or prefix with `_` if part of an interface
5. **Repeated strings**: Extract to constants when appropriate (goconst)
6. **Unnecessary conversions**: Remove redundant type conversions

### Handling Linting Errors

When golangci-lint reports errors:

1. **Read the error message carefully** - it includes the file, line, and specific issue
2. **Fix the root cause** - don't just suppress warnings
3. **Run linting again** after each fix to ensure it's resolved
4. **Run tests** to ensure fixes don't break functionality: `make test`

## Exceptions and Exclusions

Note: The configuration **excludes test files** (`*_test.go`) from linting runs. However, test files are still formatted by gofmt.

## Complete Pre-PR Checklist

Before finalizing any Pull Request, run this sequence:

```bash
# 1. Format code
make fmt

# 2. Run linter
make lint

# 3. Run vet
make vet

# 4. Run tests to ensure nothing broke
make test

# 5. Review changes
git diff
```

Fix any errors reported by steps 2-4 before committing.

## Installing Tools

If golangci-lint is not installed, run:

```bash
make install-tools
```

This installs the latest version of golangci-lint.

## Configuration Files

- **`.golangci.yml`**: Linter configuration (timeout: 5m, skips: vendor/, bin/)
- **`Makefile`**: Convenient commands for all development tasks

## Best Practices

1. **Run formatters first** - they're automatic and fast
2. **Fix linting errors before committing** - don't create PRs with known issues
3. **Commit formatting separately** - makes code review easier
4. **Test after fixes** - ensure linting fixes don't break functionality
5. **Don't suppress warnings unnecessarily** - fix the root cause when possible
