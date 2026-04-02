# Security Rules

## Config
- **bootstrap.json only** — never use .env, .env.*, or godotenv.Load()
- `bootstrap.json` is in `.gitignore` (never commit)
- All credentials live in bootstrap.json

## Docker
- Only **port 8888** exposed externally (`0.0.0.0`)
- All infrastructure binds to `127.0.0.1` (internal only)
- MongoDB/PostgreSQL = native on host (not in Docker)
- Docker containers access host DB via `host.docker.internal`

## Git
- Never commit: `bootstrap.json`, `.env`, credentials, API keys
- Never force push to main/master
- Use `dev` branch for development

## API Keys
- MCP API keys use prefix `bc_live_` or `bc_test_`
- Keys stored in MongoDB (encrypted)
- Permission levels: readonly, developer (*), custom

## Code
- Never panic — return error instead
- Never use raw SQL without parameterization (prevent SQL injection)
- Never store passwords as plain text
- Never use --no-verify or --force without explicit permission
