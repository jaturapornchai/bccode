#!/usr/bin/env bash
# AI-team dispatch: Claude (leader) -> Gemini (UX/UI design worker; Antigravity engine).
# Produces a UX/UI design + spec in READ-ONLY (plan) mode. Hand the result to Codex to implement.
#
# Usage:
#   dispatch-design.sh "<design brief: who uses it / what for / tone; must be สวย + ใช้ง่าย + เข้าใจง่าย>"
#
# Auth : Google OAuth (free preview, no per-token API key).
# Note : agy (Antigravity CLI) headless stdout is broken on Windows (typing-animation needs a real TTY),
#        so automated design runs use `gemini` (same Gemini engine). Use `agy` INTERACTIVELY for iterative design.
set -euo pipefail

if [ "$#" -lt 1 ] || [ -z "${1:-}" ]; then
  echo "usage: dispatch-design.sh \"<design brief>\"" >&2
  exit 1
fi

# Subscription/OAuth-only guard (no per-token billing).
if [ -n "${GEMINI_API_KEY:-}" ] || [ -n "${GOOGLE_API_KEY:-}" ]; then
  echo "[ABORT] GEMINI_API_KEY/GOOGLE_API_KEY is set — policy is OAuth-only. Unset it before dispatch." >&2
  exit 2
fi

export GEMINI_CLI_TRUST_WORKSPACE=true
echo "[dispatch-design] engine=gemini mode=plan(read-only)" >&2
# plan = read-only design/spec; Codex implements the code afterwards.
exec gemini -p "$1" --approval-mode plan -o text
