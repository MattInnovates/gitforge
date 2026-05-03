# Contributing to GitForge

Thank you for your interest in contributing!

## Development Setup

1. Fork and clone the repository
2. Install Go 1.22+
3. Run `go mod tidy` to install dependencies
4. Build all binaries:
   ```bash
   go build ./...
   ```
5. Run tests:
   ```bash
   go test ./...
   ```

## Project Structure

See [Architecture Overview](../architecture/overview.md) to understand how the codebase is organized before making changes.

## Making Changes

- Keep PRs focused — one feature or fix per PR
- Write tests for new functionality
- Follow standard Go formatting (`gofmt`)
- Run `go vet ./...` before submitting

## Submitting a Pull Request

1. Create a branch: `git checkout -b feature/my-feature`
2. Commit your changes with clear messages
3. Push and open a PR against `main`
4. Describe what your change does and why

## Reporting Issues

Open a GitHub issue with:
- What you expected to happen
- What actually happened
- Steps to reproduce
- Go version and OS
