---
name: bc-deploy
description: Prepare to deploy BC Account.
disable-model-invocation: true
---

- **Rules**: Read [bc-account-core-rules.md](file:///D:/bccode/.agents/rules/bc-account-core-rules.md) for deploy facts.
- **Scope**: `deploy dev` builds both frontend and backend together.
- **Secrets**: Read credentials from env/secret variables only; do not commit.
- **Checks**: Verify tests, build, endpoints, and version compatibility before deploy.
