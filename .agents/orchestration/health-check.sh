#!/usr/bin/env bash
# Pre-flight check: is the AI team ready? (run before a big orchestration session)
# Verifies subscription/OAuth auth, no per-token API keys, and live reachability.
set -uo pipefail
fail=0

echo "== 1. env (must be subscription/OAuth — NO per-token key) =="
for v in OPENAI_API_KEY GEMINI_API_KEY GOOGLE_API_KEY GOOGLE_GENAI_API_KEY; do
  if [ -z "${!v:-}" ]; then echo "  ok   $v unset"; else echo "  FAIL $v is SET (per-token billing risk!)"; fail=$((fail+1)); fi
done

echo "== 2. binaries =="
command -v codex  >/dev/null 2>&1 && echo "  ok   codex"  || { echo "  FAIL codex missing";  fail=$((fail+1)); }
command -v gemini >/dev/null 2>&1 && echo "  ok   gemini" || { echo "  FAIL gemini missing"; fail=$((fail+1)); }
test -f "$LOCALAPPDATA/agy/bin/agy.exe" && echo "  ok   agy ($("$LOCALAPPDATA/agy/bin/agy.exe" --version 2>/dev/null))" || echo "  warn agy not installed (interactive design unavailable)"

echo "== 3. live auth (~20s) =="
if codex exec --skip-git-repo-check --ephemeral -m gpt-5.4-mini -s read-only \
     -c model_reasoning_effort=low -c 'mcp_servers={}' \
     "Reply with exactly this token: PONG" 2>/dev/null | grep -q PONG; then
  echo "  ok   codex auth (ChatGPT subscription)"
else echo "  FAIL codex auth — run: codex login"; fail=$((fail+1)); fi

if GEMINI_CLI_TRUST_WORKSPACE=true gemini -p "Reply with exactly this token: PONG" \
     --approval-mode plan -o text 2>/dev/null | grep -q PONG; then
  echo "  ok   gemini auth (Google OAuth)"
else echo "  FAIL gemini auth — run: gemini (sign in once)"; fail=$((fail+1)); fi

echo "== result =="
if [ "$fail" -eq 0 ]; then echo "TEAM READY"; exit 0; else echo "TEAM NOT READY ($fail problem(s))"; exit 1; fi
