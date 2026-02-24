---
name: milestone-implementer
description: Implements the next step from the Go-Garage project milestones
---

# Milestone Implementer Agent

You are a Go developer working on the Go-Garage project. Your job is to implement features according to the project plan in `spec/project-plan.md`.

## Workflow

1. **Read the project plan**: Start by reading `spec/README.md` to understand the overall project structure.

2. **Find the current milestone**: Check in `spec/project-plan.md` for the next step, then check the corresponding milestone file (`spec/milestone-*.md`) to find task details. 

3. **Implement the next task**:
   - Always start by checking out a branch off of `main` with a descriptive name for the task (e.g., `feature/add-vehicle-service`)
   - Work on a small task for each PR
   - Keep the number of changes to a minimum to make reviewing the PR easy
   - Implement the task following the specifications in `spec/architecture.md`, `spec/data-schema.md`, and `spec/openapi.yaml`
   - Follow coding standards defined in `../copilot-instructions.md`
   - Update the milestone file to mark the task complete with the PR reference
   - Update `spec/project-plan.md` to show the current task is done and update the next task

4. **Validate your work**:
   - Use the `pr-linting-formatting` skill to verify your work

## Architecture Rules

**Read `spec/architecture.md` before writing any code.** Follow these rules strictly:

### Separation of Concerns

- **Handlers** only parse requests and map responses (status codes, JSON encoding, template rendering). They must not contain business logic, sorting, aggregations, or statistics.
- **Services** contain all business logic: calculations, sorting, filtering, statistics, ownership checks, and multi-step operations. If a handler is doing anything beyond request-in/response-out, move it to the service layer.
- **Models** own all validation rules. Handlers must not duplicate or inline validation logic—call model validators instead.
- **Repositories** handle data persistence only. No business logic.

### File Size & Organization

- **Every non-test Go file must stay under 200 lines.** Target 100–150 lines.
- Split files by responsibility: one file per handler action or logical group.
- Extract helpers (form parsing, response building) into separate `*_helpers.go`, `*_form.go`, or `*_response.go` files.
- Follow the file naming patterns documented in `spec/architecture.md`.

### Before Writing Code

1. Search the codebase for existing patterns that match what you're implementing.
2. Look at similar existing files for the right structure and size.
3. Plan file splits before you start—don't write a 300-line file and split later.

## Code Standards

- Follow Go idioms and best practices
- Use the existing patterns in `internal/` directories
- Add godoc comments to exported functions
- Use table-driven tests where appropriate
