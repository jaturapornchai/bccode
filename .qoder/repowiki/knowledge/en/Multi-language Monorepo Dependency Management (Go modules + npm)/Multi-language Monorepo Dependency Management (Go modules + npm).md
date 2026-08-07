---
kind: dependency_management
name: Multi-language Monorepo Dependency Management (Go modules + npm)
category: dependency_management
scope:
    - '**'
source_files:
    - backend/go.mod
    - backend/go.sum
    - backend/.dockerignore
    - backend/Makefile
    - frontend/package.json
    - frontend/package-lock.json
    - jarvisui/package.json
---

This monorepo manages dependencies across three independent sub-projects, each using its own language-native package manager with no cross-project sharing:

Backend (Go) - backend/go.mod declares a single Go module (smlcloudplatform) at Go 1.26 with ~30 direct dependencies and ~80 indirect ones. Dependencies are pinned via go.sum, which is copied into every Docker image (COPY go.mod go.sum ./). There is no vendoring directory, no GOPRIVATE or replace directives, and no private Go proxy configured in the module file itself. The Makefile installs tooling via go install github.com/swaggo/swag/cmd/swag@latest rather than declaring it as a module dependency.

Frontend (Next.js) - frontend/package.json uses npm with a package-lock.json lockfile. It pins React 19.2.6, Next.js 16.2.6, and Tailwind CSS v4; dev/test tooling includes Vitest 4, Playwright, and ESLint 9. An overrides block forces postcss to 8.5.10 to resolve a version conflict.

Jarvis UI (Vite/React) - jarvisui/package.json is an independent project (React 18, Vite 5, Three.js) with its own package-lock.json. It has no relationship to the frontend's dependency tree.

Key conventions and constraints:
- Each sub-project is fully self-contained; there is no shared workspace, pnpm workspaces, or Go multi-module setup.
- Lockfiles (go.sum, package-lock.json) are committed and used by CI/Docker builds for reproducible installs.
- No vendoring strategy is used for either Go or Node; dependencies are downloaded fresh during build.
- Private/internal packages are not referenced via replace or GOPRIVATE; all imports come from public GitHub or Go proxy sources.
- Tooling dependencies (e.g., swag) are installed ad-hoc via go install in Makefile targets rather than declared in go.mod.

Rules developers should follow:
- Add new Go dependencies only under backend/go.mod; run go mod tidy and commit both go.mod and go.sum.
- For Node projects, edit the appropriate package.json and commit the generated package-lock.json; do not mix package managers between sub-projects.
- Do not introduce cross-project dependency sharing - keep backend, frontend, and jarvisui isolated.
- When adding CLI tools, prefer go install in Makefile targets over adding them to go.mod require blocks.