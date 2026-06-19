#!/usr/bin/env bash
# Pre-flight check: are the AI advisors ready? (run before a big orchestration session)
# Team (2026-06-19): Claude (Claude Code) = lead/worker · GLM 5.2 Think = primary advisor · Codex = secondary advisor + image gen.
# Verifies subscription/OAuth auth, no per-token API keys, and live reachability.
set -uo pipefail
fail=0

echo "== 1. env (per-token billing guard — these MUST be unset) =="
for v in OPENAI_API_KEY GEMINI_API_KEY GOOGLE_API_KEY GOOGLE_GENAI_API_KEY; do
  if [ -z "${!v:-}" ]; then echo "  ok   $v unset"; else echo "  FAIL $v is SET (per-token billing risk!)"; fail=$((fail+1)); fi
done
# GLM (Z.AI) uses a subscription/coding-plan key, so ZAI_API_KEY SHOULD be set (not per-token).
if [ -n "${ZAI_API_KEY:-}" ]; then echo "  ok   ZAI_API_KEY set (GLM primary advisor)"; else echo "  warn ZAI_API_KEY unset — GLM advisor unavailable (Claude runs solo)"; fi

echo "== 2. binaries / helpers =="
command -v codex >/dev/null 2>&1 && echo "  ok   codex" || { echo "  FAIL codex missing"; fail=$((fail+1)); }
command -v powershell >/dev/null 2>&1 && echo "  ok   powershell" || echo "  warn powershell missing (advisor helpers need it)"
test -f "tools/ai/glm52-think-planner.ps1" && echo "  ok   glm52-think-planner.ps1" || echo "  warn glm52-think-planner.ps1 missing"
test -f "tools/ai/codex-advisor.ps1" && echo "  ok   codex-advisor.ps1" || echo "  warn codex-advisor.ps1 missing"

echo "== 3. live auth (~20s) =="
if codex exec --skip-git-repo-check --ephemeral -m gpt-5.4-mini -s read-only \
     -c model_reasoning_effort=low -c 'mcp_servers={}' \
     "Reply with exactly this token: PONG" 2>/dev/null | grep -q PONG; then
  echo "  ok   codex auth (ChatGPT subscription)"
else echo "  FAIL codex auth — run: codex login"; fail=$((fail+1)); fi

echo "== result =="
if [ "$fail" -eq 0 ]; then echo "TEAM READY"; exit 0; else echo "TEAM NOT READY ($fail problem(s))"; exit 1; fi
