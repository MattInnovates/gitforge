# Getting Started with GitForge

## What is GitForge?

GitForge is a lightweight, self-hosted Git platform written in Go. It provides Git repository hosting over both HTTP and SSH, a web interface, and a background worker for async tasks.

## Prerequisites

- Go 1.22 or later
- Git installed on the host machine

## Quick Start

1. Clone the repository:
   ```bash
   git clone https://github.com/yourname/gitforge.git
   cd gitforge
   ```

2. Initialize and build:
   ```bash
   go mod tidy
   go build ./cmd/web
   go build ./cmd/git-http
   go build ./cmd/git-ssh
   go build ./cmd/worker
   ```

3. Start the web server:
   ```bash
   ./web
   ```

## Next Steps

- [Configuration](../configuration/configuration.md)
- [Architecture Overview](../architecture/overview.md)
- [Creating Your First Repository](./first-repository.md)
