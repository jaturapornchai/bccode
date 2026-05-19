---
name: go-expert
description: Auto-activate when writing, reviewing, or modifying Go code. Enforces Go best practices for BC Account backend.
---
When working with Go code:

## Mandatory Steps
1. LSP goToDefinition before modifying an unfamiliar function
2. LSP findReferences before renaming or refactoring
3. LSP getDiagnostics after every change

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
