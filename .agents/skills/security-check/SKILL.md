---
name: security-check
description: Automatically run a security audit before every commit.
---

Check these topics before committing:
- **Injection**: Verify parameterized queries / ORM usage (no raw concatenation).
- **Access Bypass**: Ensure auth middleware covers every route.
- **Hardcoded secrets**: Grep for keys/tokens (`sk-`, `Bearer `, `password=`).
- **Isolation**: Company-owned operational reads/writes must use server-validated `holdingcode + businesscode` in filters, indexes, joins, events, caches, and projections; branch-owned data also uses `branchcode`. Reject missing/unauthorized company context and ignore mutation-body company values. Holding-wide endpoints must be authorized, read-only aggregation that preserves `businesscode`.
- **Paths**: Never accept user input directly as a file path.

Report format:
- `CRITICAL: [issue] -> [fix]`
- `HIGH: [issue] -> [fix]`
- `PASS: [topic]`
