---
name: bc-deploy
description: Prepare to deploy BC Account.
disable-model-invocation: true
---

- **Rules**: Read [bc-account-core-rules.md](file:///D:/bccode/.agents/rules/bc-account-core-rules.md) for deploy facts.
- **Scope**: `deploy dev` builds both frontend and backend together.
- **Local Backend Auto Deploy**: Backend Go-code edits are auto deployed to local Docker Desktop with `cd D:\bccode\backend; .\scripts\deploy-mainapi-fast.ps1`, then `/healthz` is verified. Use full `docker-compose up -d --no-deps --build mainapi` for Dockerfile, dependency, runtime asset, compose, config, or image-content changes. This is local-only and does not replace explicit `deploy dev`.
- **Docker Desktop Buildx**: If `docker buildx` is missing on Windows Docker Desktop, first check `C:\Program Files\Docker\Docker\resources\cli-plugins\docker-buildx.exe` and copy it to `%USERPROFILE%\.docker\cli-plugins\docker-buildx.exe`; then verify with `docker buildx version` and `docker buildx ls`.
- **Backend Dockerfile Cache**: Keep backend full rebuilds on BuildKit cache mounts for `/var/cache/apk`, `/go/pkg/mod`, and `/root/.cache/go-build`; verify with `docker-compose build mainapi` before recreating `mainapi`.
- **Secrets**: Read credentials from env/secret variables only; do not commit.
- **Checks**: Verify tests, build, endpoints, and version compatibility before deploy.
