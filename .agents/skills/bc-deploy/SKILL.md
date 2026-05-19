---
name: bc-deploy
description: Prepare to deploy BC Account
disable-model-invocation: true
---

Prepare BC Account for deployment:

1. Verify all tests pass
2. Verify lint has no errors
3. Build Docker images
4. Verify migration files
5. Summarize changes to be deployed

IMPORTANT: This skill must only be invoked via /bc-deploy — do not auto-invoke.
