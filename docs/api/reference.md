# API Reference

GitForge exposes a REST API used by the web interface and available for external tooling.

## Base URL

```
http://localhost:3000/api
```

## Authentication

> To be documented once auth is implemented.

## Endpoints

### Repositories

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/repos` | List all repositories |
| `POST` | `/api/repos` | Create a repository |
| `GET` | `/api/repos/:owner/:name` | Get repository details |
| `DELETE` | `/api/repos/:owner/:name` | Delete a repository |

### Users

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/users/:username` | Get user profile |
| `POST` | `/api/users` | Register a user |

> More endpoints will be added as features are implemented.
