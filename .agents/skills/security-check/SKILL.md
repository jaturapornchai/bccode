---
name: security-check
description: Automatically run a security audit before every commit.
---

Check these topics before committing:
- **Injection**: Verify parameterized queries / ORM usage (no raw concatenation).
- **Access Bypass**: Ensure auth middleware covers every route.
- **Hardcoded secrets**: Grep for keys/tokens (`sk-`, `Bearer `, `password=`).
- **Isolation**: Check that all customer queries use the validated `tenant_id` (shop ID).
- **Paths**: Never accept user input directly as a file path.

Report format:
- `CRITICAL: [issue] -> [fix]`
- `HIGH: [issue] -> [fix]`
- `PASS: [topic]`
