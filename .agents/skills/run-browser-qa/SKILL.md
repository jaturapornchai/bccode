---
name: run-browser-qa
description: Test a website through an available Browser MCP, Playwright MCP, Stagehand, or Skyvern like a real user while actively finding reproducible UI, console, and network bugs. Use for localhost UAT, login and CRUD workflow verification, post-deploy browser checks, visual regression investigation, or any request to browse/test a web flow.
---

# Run Browser QA

## Inputs

Establish the target URL, user persona, goal, and special notes. Infer only low-risk missing details from the active task and source; ask only when credentials, destructive actions, money, or another required choice cannot be discovered safely.

## Tool Choice

- Prefer a real Browser MCP when an existing signed-in browser session is useful.
- Prefer Playwright MCP for accessibility snapshots, repeatable selectors, console logs, and network requests.
- Use natural-language actions with Stagehand or Skyvern.
- State which browser tool is actually available; never claim console or network evidence the tool did not expose.

## Human-Like Flow

1. Navigate, wait for the page to settle, then observe or snapshot before acting.
2. Read visible context. Scroll naturally when the target is off-screen; do not jump or click blindly.
3. Hover for roughly 200-400 ms before clicking when the tool supports hover.
4. Focus and clear fields before typing. Use a 30-100 ms keystroke delay when supported.
5. After navigation, submit, modal open, or another major state change, wait for network idle or about 1-2 seconds and observe again.
6. Do not spam or rapid-double-click an unresponsive control unless that exact edge case is under test. Use browser history when a real user would go back.

## Audit Every Significant Step

- Console: flag unexpected errors, warnings, uncaught exceptions, and missing JS/CSS assets.
- Network: flag failed requests, 4xx/5xx responses, CORS failures, and responses slower than 5 seconds.
- Visual: inspect screenshots or snapshots for broken images, overlap, clipping, overflow, layout shifts, incorrect enabled/disabled states, and missing post-request data.
- Preserve the exact safe error, request path/status, step, and screenshot when available. Never expose credentials, tokens, PII, or customer data.

## Bug Handling

Stop the affected flow and record:

```text
BUG FOUND
Title: <one line>
Severity: Critical | Major | Minor
Step: <number and action>
Expected: <expected result>
Actual: <actual result>
Evidence: <console, network, and UI evidence available>
Repro: <minimal numbered steps>
```

For review/diagnosis requests, report without mutating the system. For an authorized build/fix task, fix the source-backed root cause, rerun the same reproduction, and then continue the goal.

## Completion

Report the tool used, pages visited, key actions, bugs found, console status, network status, and production recommendation. Mark unavailable evidence as `not exposed by tool`, not `clean`.

