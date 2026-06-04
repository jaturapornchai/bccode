#!/usr/bin/env bash
# AI-team dispatch: Claude (leader) -> Codex (code worker).
# Codex writes backend + frontend IMPLEMENTATION code from a 5-section handoff spec.
#
# Usage:
#   dispatch-codex.sh "<spec>"          # default: gpt-5.5  + reasoning medium (complex)
#   dispatch-codex.sh --mini "<spec>"   # fast/small/parallel: gpt-5.4-mini + reasoning low
#   dispatch-codex.sh --deep "<spec>"   # hardest/architecture: gpt-5.5 + reasoning high
#
# Speed/reliability tuning (verified 2026-06-04 on this machine):
#   - keep user config (project TRUST) so workspace-write can actually edit files
#   - clear MCP servers (`-c mcp_servers={}`) -> no slow MCP load, no github-copilot auth error noise
#   - reasoning effort matched to tier (config default xhigh is the main latency culprit)
#   - --ephemeral: no session files persisted
# Auth: ChatGPT subscription (no per-token API key).
set -euo pipefail

model="gpt-5.5"; effort="medium"
case "${1:-}" in
  --mini) model="gpt-5.4-mini"; effort="low";  shift ;;
  --deep) model="gpt-5.5";      effort="high"; shift ;;
esac

if [ "$#" -lt 1 ] || [ -z "${1:-}" ]; then
  echo "usage: dispatch-codex.sh [--mini|--deep] \"<5-section handoff spec>\"" >&2
  exit 1
fi

# Subscription-only guard (no per-token billing).
if [ -n "${OPENAI_API_KEY:-}" ]; then
  echo "[ABORT] OPENAI_API_KEY is set — policy is subscription-only. Unset it before dispatch." >&2
  exit 2
fi

echo "[dispatch-codex] model=$model effort=$effort sandbox=workspace-write mcp=off" >&2
exec codex exec \
  --skip-git-repo-check --ephemeral \
  -m "$model" -s workspace-write \
  -c model_reasoning_effort="$effort" \
  -c 'mcp_servers={}' \
  "$1"
