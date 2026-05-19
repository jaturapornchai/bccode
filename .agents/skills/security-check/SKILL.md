---
name: security-check
description: Automatically run a security audit before every commit — detect vulnerabilities in BC Account codebase
---
Check every item and report any issues found:

## Critical (must fix before commit)
1. SQL Injection: every query uses parameterized or ORM
2. Auth bypass: verify authorization middleware covers every route
3. Hardcoded secrets: no passwords/keys/tokens in code
4. Input validation: every API endpoint validates input
5. Path traversal: never accept user input directly as a file path

## High (fix soon)
6. Error exposure: do not leak stack traces to users
7. Rate limiting: important endpoints have rate limits
8. CORS: config is appropriate — no wildcard in production
9. Logging: do not log sensitive data (password, token, PII)

## Multi-tenant Security
10. Tenant isolation: every customer-data query has a `tenant_id` filter or an explicitly authorized `tenant_id IN (...)` list. Legacy `shop_id` / `shopid` filters must map from `tenant_id`.
11. Cross-tenant: verify no data leaks across tenants

Report as:
- CRITICAL: [issue] -> [fix]
- HIGH: [issue] -> [fix]
- PASS: [topic]
