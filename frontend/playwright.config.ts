import { defineConfig } from "@playwright/test";

// e2e/UAT regression suite. Assumes the dev server is already running at BASE_URL (this project's
// convention is `build && start`, never `next dev` — see CLAUDE.md). This config does not manage
// the server lifecycle itself; start it separately before running `npm run test:e2e`.
export default defineConfig({
  testDir: "./e2e",
  timeout: 60_000,
  fullyParallel: false,
  retries: 0,
  reporter: "list",
  use: {
    baseURL: process.env.E2E_BASE_URL ?? "http://localhost:3000",
    channel: process.env.E2E_BROWSER_CHANNEL || undefined,
    trace: "retain-on-failure",
  },
});
