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
#   sh tools/verify.sh backend      # one target: codemap|frontend|frontend-build|backend|outbox|projection
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
t_codemap() {
  hr "codemap — docs/reference/CODE-MAP.md ตรงกับซอร์ส"
  if ! command -v pwsh >/dev/null 2>&1; then
    echo "ไม่พบ pwsh — ตรวจไม่ได้ ถือว่าไม่ผ่าน (ติดตั้ง PowerShell 7)" >&2
    record codemap 1
    return
  fi
  pwsh -NoProfile -File tools/gen-code-map.ps1 -Check
  record codemap $?
}

# --- frontend: mirrors job frontend-test (build split out, it is the slow half) ---
t_frontend() {
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
  record frontend $?
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

# --- outbox: mirrors job backend-outbox-integration -------------------------
t_outbox() {
  hr "outbox — transaction rollback + barcode replay (Mongo rs0 + PG แยกต่างหาก)"
  need_docker || { record outbox 1; return; }
  rc=0
  docker rm -fv bc-outbox-ci bc-barcode-ci >/dev/null 2>&1

  docker run -d --name bc-outbox-ci mongo:7 --replSet rs0 --bind_ip_all >/dev/null || rc=1
  attempt=1
  while [ $rc -eq 0 ] && [ $attempt -le 30 ]; do
    echo "MongoDB readiness attempt $attempt"
    if docker exec bc-outbox-ci mongosh --quiet --eval 'db.adminCommand({ping:1})' >/dev/null 2>&1; then break; fi
    attempt=$((attempt + 1))
    sleep 1
  done
  [ $rc -eq 0 ] && docker exec bc-outbox-ci mongosh --quiet --eval 'rs.initiate({_id:"rs0",members:[{_id:0,host:"localhost:27017"}]})' >/dev/null 2>&1

  if [ $rc -eq 0 ]; then
    docker run --rm --network container:bc-outbox-ci \
      -e SERVERLESS=serverless \
      -e BC_OUTBOX_TEST_MONGODB_URI='mongodb://127.0.0.1:27017/?replicaSet=rs0' \
      -v "$WINROOT/backend:/src" -w /src "$GO_IMAGE" \
      bash -c 'set -euo pipefail; go test -tags=integration -count=1 -timeout=120s -json -run "^TestProduct(Outbox|ServiceOutbox)Integration$" ./internal/product/product/outbox ./internal/product/product/services | tee outbox-test-results.json' || rc=1
  fi

  if [ $rc -eq 0 ]; then
    docker run -d --name bc-barcode-ci -e POSTGRES_HOST_AUTH_METHOD=trust postgres:17-alpine >/dev/null || rc=1
    attempt=1
    while [ $rc -eq 0 ] && [ $attempt -le 30 ]; do
      echo "PostgreSQL readiness attempt $attempt"
      if docker exec bc-barcode-ci pg_isready -U postgres >/dev/null 2>&1; then break; fi
      attempt=$((attempt + 1))
      sleep 1
    done
    [ $rc -eq 0 ] && docker run --rm --network container:bc-barcode-ci \
      -e SERVERLESS=serverless \
      -e BC_BARCODE_TEST_POSTGRES_DSN='postgres://postgres@127.0.0.1:5432/postgres?sslmode=disable' \
      -v "$WINROOT/backend:/src" -w /src "$GO_IMAGE" \
      bash -c 'set -euo pipefail; go test -tags=integration -count=1 -timeout=120s -json -run "^TestBarcodeBatchIntegration$" ./internal/goapi/handlers/kafka | tee barcode-test-results.json' || rc=1
  fi

  # Always clean up, exactly like the workflow's `if: always()` step.
  docker rm -fv bc-outbox-ci bc-barcode-ci >/dev/null 2>&1
  echo "ผลดิบ: backend/outbox-test-results.json, backend/barcode-test-results.json"
  record outbox $rc
}

# --- projection: mirrors job backend-projection-kafka-integration -----------
t_projection() {
  hr "projection — cross-topic fences + legacy writers (Kafka + Mongo + PG18)"
  need_docker || { record projection 1; return; }
  COMPOSE=backend/.ci/projection.compose.yml
  rc=0
  docker compose -p bc-projection-ci -f "$COMPOSE" up -d mongo mongo-init postgres kafka || rc=1
  if [ $rc -eq 0 ]; then
    docker compose -p bc-projection-ci -f "$COMPOSE" up -d --wait mongo postgres kafka || rc=1
  fi
  if [ $rc -eq 0 ]; then
    # `... | tee` would report tee's exit status, hiding a failing test — the workflow got away
    # with it only because its step ran under `set -o pipefail`. Carry the real status by hand.
    status_file=$(mktemp)
    {
      docker compose -p bc-projection-ci -f "$COMPOSE" run --rm --no-deps tests \
        'go test -tags=integration -count=1 -timeout=180s -json -run "^Test(Projection(Kafka|Rebalance)|BarcodeServiceOutbox|LegacyBarcode(Reconcile|PrimarySource))Integration$" ./internal/goapi/handlers/kafka ./internal/product/productbarcode/repositories ./internal/product/productbarcode/services ./internal/product/projection'
      echo $? > "$status_file"
    } | tee backend/projection-test-results.json
    [ "$(cat "$status_file")" = "0" ] || rc=1
    rm -f "$status_file"
  fi
  docker compose -p bc-projection-ci -f "$COMPOSE" down -v --remove-orphans >/dev/null 2>&1
  echo "ผลดิบ: backend/projection-test-results.json"
  record projection $rc
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
  echo "  outbox          integration: outbox rollback + barcode replay"
  echo "  projection      integration: Kafka projection fences"
}

targets=${*:-fast}
case " $targets " in
  *" help "*|*" -h "*|*" --help "*) usage; exit 0 ;;
esac

started=$(date '+%H:%M:%S')
for t in $targets; do
  case "$t" in
    fast)           t_codemap; t_frontend ;;
    all)            t_codemap; t_frontend; t_frontend_build; t_backend; t_outbox; t_projection ;;
    codemap)        t_codemap ;;
    frontend)       t_frontend ;;
    frontend-build) t_frontend_build ;;
    backend)        t_backend ;;
    outbox)         t_outbox ;;
    projection)     t_projection ;;
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
