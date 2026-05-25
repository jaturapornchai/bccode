---
name: go-expert
description: Auto-activate when writing, reviewing, or modifying Go code. Enforces Go best practices for BC Account backend.
---
When working with Go code:

## Mandatory Steps
1. LSP goToDefinition before modifying an unfamiliar function
2. LSP findReferences before renaming or refactoring
3. LSP getDiagnostics after every change
4. After backend code/config changes, automatically deploy the affected service to Docker Desktop before completion: `cd D:\bccode\backend; docker-compose up -d --no-deps --build mainapi`
5. Verify the running Docker Desktop backend with `http://localhost:8888/healthz` or the changed route.

## Required Patterns
- Context: always pass ctx through every function
- Error: return errors — do not use panic in production
- Goroutine: every goroutine must have a done channel or context cancel
- SQL: use $1,$2 parameterized queries — never string concat
- Gin route: use middleware for auth and logging

## Pre-commit Checks
```bash
go vet ./...
go build ./...
```

## Forbidden Anti-patterns
- _ = err (ignoring errors)
- time.Sleep in production code
- Global variables without a mutex
- SELECT * in SQL queries
