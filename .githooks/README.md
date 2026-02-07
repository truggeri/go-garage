# Go-Garage Git Hooks

This directory contains Git hooks for the Go-Garage project.

## Setup

To enable these hooks, run:

```bash
git config core.hooksPath .githooks
```

## Available Hooks

### pre-commit

Runs before each commit to:
- Format all Go files with `gofmt`
- Run `go vet` for static analysis
- Run `go fmt` to check formatting

If any check fails, the commit will be aborted.
