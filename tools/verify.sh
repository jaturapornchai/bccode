#!/bin/sh
# BC Ai Account — local verification runner.
#
# This replaces .github/workflows/ci.yml, deleted on 2026-09-09 when the owner decided GitHub is
# code storage only (the account is locked for a billing issue, so Actions never ran again anyway —
# GitHub's own annotation: "The job was not started because your account is locked due to a billing
# issue."). Nothing on the GitHub side checks this repo any more: THIS SCRIPT IS THE CHECK, and it
# only runs when a human runs it. The commands below are copied verbatim from the deleted workflow
# so the coverage is identical; see docs/kms/14-cmd-tools-scripts.md.
#
#   sh tools/verify.sh              # fast: codemap + frontend lint/typecheck/test  (~2-4 min)
#   sh tools/verify.sh all          # everything, including the Docker integration suites (~20 min)
#   sh tools/verify.sh backend      # one target: codemap|frontend|frontend-build|backend|postgres
#
# Requirements: Docker Desktop running, pwsh (for codemap), Node/npm (for frontend).
#
# Two deliberate deviations from the workflow, both in the frontend target:
#   * npm ci runs only when node_modules is missing (CI always reinstalled from scratch).
#   * node is whatever the machine has (this box: v24.14.0); CI pinned 24.18.0 via setup-node.

set -u

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT" || exit 1

# Docker bind mounts need a Windows-style path under Git Bash, and MSYS must not rewrite /src.
if command -v cygpath >/dev/null 2>&1; then
  WINROOT=$(cygpath -m "$ROOT")
else
  WINROOT=$ROOT
fi
export MSYS_NO_PATHCONV=1

GO_IMAGE=golang:1.26
FAILED=""
PASSED=""

hr() { printf '\n== %s ==\n' "$1"; }
ok() { PASSED="$PASSED $1"; printf 'PASS %s\n' "$1"; }
ko() { FAILED="$FAILED $1"; printf 'FAIL %s\n' "$1"; }
record() { if [ "$2" -eq 0 ]; then ok "$1"; else ko "$1"; fi; }

need_docker() {
  if ! docker info >/dev/null 2>&1; then
    echo "docker ไม่ตอบสนอง — เปิด Docker Desktop ก่อน แล้วรันใหม่" >&2
    return 1
  fi
  return 0
}

# --- codemap: mirrors job code-map-check ------------------------------------
run_codemap() {
  hr "codemap — docs/reference/CODE-MAP.md ตรงกับซอร์ส"
  if ! command -v pwsh >/dev/null 2>&1; then
    echo "ไม่พบ pwsh — ตรวจไม่ได้ ถือว่าไม่ผ่าน (ติดตั้ง PowerShell 7)" >&2
    return 1
  fi
  pwsh -NoProfile -File tools/gen-code-map.ps1 -Check
}

t_codemap() {
  run_codemap
  record codemap $?
}

# --- frontend: mirrors job frontend-test (build split out, it is the slow half) ---
run_frontend() {
  hr "frontend — lint / typecheck / vitest"
  ( cd frontend || exit 1
    if [ ! -d node_modules ]; then
      echo "node_modules หาย — npm ci ก่อน"
      npm ci || exit 1
    fi
    BCAI_LOCAL_BACKEND_URL=http://localhost:8888 npm run lint || exit 1
    BCAI_LOCAL_BACKEND_URL=http://localhost:8888 npm run typecheck || exit 1
    BCAI_LOCAL_BACKEND_URL=http://localhost:8888 npm test -- --run || exit 1
  )
  local status=$?
  # A client that forgets the token does not fail loudly — the screen just says
  # "ไม่พบข้อมูล". Nothing else catches that, so it is checked on every verify.
  if [ $status -eq 0 ]; then
    node tools/audit-auth-fetch.mjs || status=1
  fi
  return $status
}

t_frontend() {
  run_frontend
  record frontend $?
}

# --- fast: run codemap + frontend concurrently -----------------------------
t_fast() {
  local codemap_log frontend_log codemap_pid frontend_pid codemap_rc frontend_rc
  codemap_log=$(mktemp)
  frontend_log=$(mktemp)

  ( run_codemap > "$codemap_log" 2>&1 ) &
  codemap_pid=$!

  ( run_frontend > "$frontend_log" 2>&1 ) &
  frontend_pid=$!

  wait "$codemap_pid"
  codemap_rc=$?

  wait "$frontend_pid"
  frontend_rc=$?

  cat "$codemap_log"
  record codemap "$codemap_rc"
  rm -f "$codemap_log"

  cat "$frontend_log"
  record frontend "$frontend_rc"
  rm -f "$frontend_log"
}

t_frontend_build() {
  hr "frontend-build — next build"
  ( cd frontend || exit 1
    BCAI_LOCAL_BACKEND_URL=http://localhost:8888 npm run build )
  record frontend-build $?
}

# --- backend: mirrors job backend-test --------------------------------------
# Quarantined packages (backend/.ci/test-quarantine.txt) are compiled but their tests are not run.
t_backend() {
  hr "backend — go build + unit tests (quarantine list ไม่ถูกรัน)"
  need_docker || { record backend 1; return; }
  echo "แพ็กเกจที่ถูกกักไว้ (compile อย่างเดียว ไม่รัน test):"
  sed '/^#/d; /^[[:space:]]*$/d; s/^/  - /' backend/.ci/test-quarantine.txt
  # The deleted workflow printed this into the GitHub step summary; keep the warning where the
  # person running the tests will actually see it.
  echo "ผ่านชุดนี้ไม่ได้แปลว่า business contract ของแพ็กเกจข้างบนถูกต้อง (a green run does not certify them)"
  docker run --rm \
    -e SERVERLESS=serverless \
    -v "$WINROOT/backend:/src" \
    -w /src \
    "$GO_IMAGE" \
    bash -c '
      set -euo pipefail
      go test -run "^$" -count=1 ./cmd/... ./pkg/... ./internal/...
      mapfile -t unit_packages < <(go list ./cmd/... ./pkg/... ./internal/... | grep -vFx -f <(tr -d "\r" < .ci/test-quarantine.txt))
      go test -short -count=1 "${unit_packages[@]}"
      go test -tags=integration -run "^$" -count=1 ./cmd/... ./pkg/... ./internal/...
    '
  record backend $?
}

# --- postgres: integration tests against a throwaway PostgreSQL 18 ------------
# PostgreSQL is the only database (MongoDB/Kafka/Redis/ClickHouse removed 2026-09-23); every
# integration test gated by BC_GL_TEST_POSTGRES_DSN / GL_AUTH_TEST_DSN runs here.
t_postgres() {
  hr "postgres — integration tests (PostgreSQL 18 แยกต่างหาก ลบทิ้งหลังจบ)"
  need_docker || { record postgres 1; return; }
  rc=0
  docker rm -fv bc-pg-ci >/dev/null 2>&1
  docker run -d --name bc-pg-ci -e POSTGRES_HOST_AUTH_METHOD=trust postgres:18-alpine >/dev/null || rc=1
  if [ $rc -eq 0 ]; then
    for attempt in $(seq 1 30); do
      if docker exec bc-pg-ci pg_isready -U postgres >/dev/null 2>&1; then break; fi
      sleep 1
    done
    docker run --rm --network container:bc-pg-ci \
      -e BC_GL_TEST_POSTGRES_DSN='postgres://postgres@127.0.0.1:5432/postgres?sslmode=disable' \
      -e GL_AUTH_TEST_DSN='postgres://postgres@127.0.0.1:5432/postgres?sslmode=disable' \
      -v "$WINROOT/backend:/src" -w /src "$GO_IMAGE" \
      bash -c 'set -euo pipefail; go test -tags=integration -count=1 -timeout=300s ./pkg/... ./internal/...' || rc=1
  fi
  docker rm -fv bc-pg-ci >/dev/null 2>&1
  record postgres $rc
}

usage() {
  echo "ใช้: sh tools/verify.sh [target ...]"
  echo ""
  echo "  (ไม่ใส่)        = fast   -> codemap + frontend"
  echo "  fast            codemap + frontend (lint/typecheck/test) — ใช้ก่อน push ทุกครั้ง"
  echo "  all             ทุกอย่าง รวม frontend-build และชุด Docker integration (~20 นาที)"
  echo "  codemap         ตรวจ docs/reference/CODE-MAP.md ตรงกับซอร์ส"
  echo "  frontend        eslint + tsc --noEmit + vitest"
  echo "  frontend-build  next build"
  echo "  backend         go build/test ใน golang:1.26 (ข้าม quarantine list)"
  echo "  postgres        integration tests กับ PostgreSQL 18 ชั่วคราว"
}

targets=${*:-fast}
case " $targets " in
  *" help "*|*" -h "*|*" --help "*) usage; exit 0 ;;
esac

started=$(date '+%H:%M:%S')
for t in $targets; do
  case "$t" in
    fast)           t_fast ;;
    all)            t_codemap; t_frontend; t_frontend_build; t_backend; t_postgres ;;
    codemap)        t_codemap ;;
    frontend)       t_frontend ;;
    frontend-build) t_frontend_build ;;
    backend)        t_backend ;;
    postgres)       t_postgres ;;
    *) echo "ไม่รู้จัก target: $t" >&2; usage; exit 2 ;;
  esac
done

hr "สรุป (เริ่ม $started จบ $(date '+%H:%M:%S'))"
if [ -n "$PASSED" ]; then printf 'ผ่าน:%s\n' "$PASSED"; fi
if [ -n "$FAILED" ]; then
  printf 'ไม่ผ่าน:%s\n' "$FAILED"
  echo "GitHub ไม่ได้ตรวจอะไรให้แล้ว — ต้องแก้ให้ผ่านก่อน push"
  exit 1
fi
echo "ผ่านทั้งหมด"
