# Configuration

GitForge is configured via environment variables or a config file (`gitforge.toml`).

## Config File Location

By default, GitForge looks for `gitforge.toml` in the working directory.

## Options

> Full configuration reference will be documented once the config system is implemented.

### Planned Options

| Key | Default | Description |
|-----|---------|-------------|
| `server.http_port` | `3000` | Port for the web server |
| `server.git_http_port` | `8080` | Port for the Git HTTP backend |
| `server.ssh_port` | `2222` | Port for the Git SSH backend |
| `storage.repos_path` | `./repos` | Directory where repositories are stored |
| `database.driver` | `sqlite3` | Database driver (`sqlite3`, `postgres`) |
| `database.dsn` | `gitforge.db` | Database connection string |

## Environment Variables

All config keys can be overridden with environment variables using the prefix `GITFORGE_` and dots replaced with underscores:

```bash
GITFORGE_SERVER_HTTP_PORT=4000 ./web
```
