#!/usr/bin/env bash
set -euo pipefail

secret_dir=/etc/bcai-account
install_dir=/opt/bcai-account
runtime_dir=/var/lib/bcai-account/config
secret_file="$secret_dir/secrets.env"

if [[ ${EUID:-$(id -u)} -ne 0 ]]; then
  echo "provision-server.sh must run as root" >&2
  exit 1
fi

: "${MAINAPI_IMAGE:?MAINAPI_IMAGE is required}"
: "${FRONTEND_IMAGE:?FRONTEND_IMAGE is required}"
image_pattern='^[a-zA-Z0-9./:_-]+$'
if [[ ! $MAINAPI_IMAGE =~ $image_pattern || ! $FRONTEND_IMAGE =~ $image_pattern ]]; then
  echo "image name is invalid" >&2
  exit 1
fi

install -d -m 0700 "$secret_dir"
install -d -m 0750 "$install_dir"
install -d -o 10001 -g 10001 -m 0700 "$runtime_dir"
umask 077

if [[ ! -f "$secret_file" ]]; then
  {
    printf 'POSTGRES_PASSWORD=%s\n' "$(openssl rand -hex 32)"
    printf 'CLICKHOUSE_PASSWORD=%s\n' "$(openssl rand -hex 32)"
    printf 'JWT_SECRET_KEY=%s\n' "$(openssl rand -hex 48)"
    printf 'RELOAD_CONFIG_SECRET=%s\n' "$(openssl rand -hex 32)"
    printf 'KAFKA_CLUSTER_ID=%s\n' "$(openssl rand 16 | openssl base64 -A | tr '+/' '-_' | tr -d '=')"
  } >"$secret_file"
fi

set -a
# shellcheck disable=SC1090
source "$secret_file"
set +a

if [[ -z ${KAFKA_CLUSTER_ID:-} ]]; then
  KAFKA_CLUSTER_ID="$(openssl rand 16 | openssl base64 -A | tr '+/' '-_' | tr -d '=')"
  printf 'KAFKA_CLUSTER_ID=%s\n' "$KAFKA_CLUSTER_ID" >>"$secret_file"
fi

if [[ -z ${MINIO_ROOT_USER:-} ]]; then
  MINIO_ROOT_USER="bcairoot$(openssl rand -hex 6)"
  printf 'MINIO_ROOT_USER=%s\n' "$MINIO_ROOT_USER" >>"$secret_file"
fi
if [[ -z ${MINIO_ROOT_PASSWORD:-} ]]; then
  MINIO_ROOT_PASSWORD="$(openssl rand -hex 32)"
  printf 'MINIO_ROOT_PASSWORD=%s\n' "$MINIO_ROOT_PASSWORD" >>"$secret_file"
fi
if [[ -z ${MINIO_APP_ACCESS_KEY:-} ]]; then
  MINIO_APP_ACCESS_KEY="bcaiapp$(openssl rand -hex 6)"
  printf 'MINIO_APP_ACCESS_KEY=%s\n' "$MINIO_APP_ACCESS_KEY" >>"$secret_file"
fi
if [[ -z ${MINIO_APP_SECRET_KEY:-} ]]; then
  MINIO_APP_SECRET_KEY="$(openssl rand -hex 32)"
  printf 'MINIO_APP_SECRET_KEY=%s\n' "$MINIO_APP_SECRET_KEY" >>"$secret_file"
fi

if [[ ! -f "$runtime_dir/bootstrap.json" ]]; then
  printf '{}\n' >"$runtime_dir/bootstrap.json"
fi
if [[ ! -f "$runtime_dir/custom_config.json" ]]; then
  printf '{}\n' >"$runtime_dir/custom_config.json"
fi
chown 10001:10001 "$runtime_dir"/*.json
chmod 0600 "$runtime_dir"/*.json

cat >"$secret_dir/postgres.env" <<EOF
POSTGRES_DB=bcai_projection
POSTGRES_USER=bcai
POSTGRES_PASSWORD=$POSTGRES_PASSWORD
EOF

cat >"$secret_dir/clickhouse.env" <<EOF
CLICKHOUSE_DB=bcai_analytics
CLICKHOUSE_USER=bcai
CLICKHOUSE_PASSWORD=$CLICKHOUSE_PASSWORD
CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT=1
EOF

cat >"$secret_dir/kafka.env" <<EOF
CLUSTER_ID=$KAFKA_CLUSTER_ID
EOF

cat >"$secret_dir/minio.env" <<EOF
MINIO_ROOT_USER=$MINIO_ROOT_USER
MINIO_ROOT_PASSWORD=$MINIO_ROOT_PASSWORD
MINIO_APP_ACCESS_KEY=$MINIO_APP_ACCESS_KEY
MINIO_APP_SECRET_KEY=$MINIO_APP_SECRET_KEY
MINIO_BUCKET_NAME=bcai-account
MINIO_BROWSER=off
EOF

cat >"$secret_dir/backend.env" <<EOF
MODE=production
GO_ENV=production
BC_ENV=production
DEV_API_MODE=2
SERVICE_PORT=8888
LOG_LEVEL=INFO
TZ=Asia/Bangkok
HOST_API=account.bcaicloud.com
HTTP_CORS=https://account.bcaicloud.com
MONGODB_PRO_URI=mongodb://mongo:27017/?replicaSet=rs0
MONGODB_PRO_DB=bcai_account
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=bcai
POSTGRES_USERNAME=bcai
POSTGRES_PASSWORD=$POSTGRES_PASSWORD
POSTGRES_DB=bcai_projection
POSTGRES_DATABASE=bcai_projection
POSTGRES_DB_NAME=bcai_projection
POSTGRES_SSL_MODE=disable
POSTGRES_TIMEZONE=UTC
CH_SERVER_ADDRESS=clickhouse
CLICKHOUSE_HOST=clickhouse
CLICKHOUSE_PORT=9000
CLICKHOUSE_USER=bcai
CLICKHOUSE_PASSWORD=$CLICKHOUSE_PASSWORD
CLICKHOUSE_DATABASE=bcai_analytics
CH_USERNAME=bcai
CH_PASSWORD=$CLICKHOUSE_PASSWORD
CH_DATABASE_NAME=bcai_analytics
REDIS_CACHE_URI=redis:6379
ENABLE_KAFKA=true
KAFKA_SERVER_URL=kafka:29092
JWT_SECRET_KEY=$JWT_SECRET_KEY
RELOAD_CONFIG_SECRET=$RELOAD_CONFIG_SECRET
GOOGLE_CLIENT_ID=212036599086-c7aqvm005jiv2kqi4duju8spd9b3jb94.apps.googleusercontent.com
BCAI_DEV_LOGIN_ENABLED=false
BCAI_DEMO_LOGIN_ENABLED=true
BCAI_DEMO_USERNAME=demo
GODEBUG=x509negativeserial=1
S3_ENDPOINT=http://minio:9000
S3_REGION=us-east-1
S3_ACCESS_KEY_ID=$MINIO_APP_ACCESS_KEY
S3_SECRET_ACCESS_KEY=$MINIO_APP_SECRET_KEY
S3_BUCKET_NAME=bcai-account
S3_FORCE_PATH_STYLE=true
STORAGE_ALLOW_PRESIGNED_URL=false
EOF

cat >"$secret_dir/frontend.env" <<EOF
NODE_ENV=production
BCAI_LOCAL_BACKEND_URL=http://mainapi:8888
NEXT_PUBLIC_GOOGLE_CLIENT_ID=212036599086-c7aqvm005jiv2kqi4duju8spd9b3jb94.apps.googleusercontent.com
GOOGLE_CLIENT_ID=212036599086-c7aqvm005jiv2kqi4duju8spd9b3jb94.apps.googleusercontent.com
BCAI_DEV_LOGIN_ENABLED=false
EOF

cat >"$secret_dir/release.env" <<EOF
MAINAPI_IMAGE=$MAINAPI_IMAGE
FRONTEND_IMAGE=$FRONTEND_IMAGE
EOF

chmod 0600 "$secret_dir"/*.env
echo "server configuration prepared"
