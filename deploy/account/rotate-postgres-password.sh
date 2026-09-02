#!/usr/bin/env bash
set -euo pipefail

secret_dir=/etc/bcai-account
install_dir=/opt/bcai-account/deploy
secret_file="$secret_dir/secrets.env"
release_file="$secret_dir/release.env"
compose_file="$install_dir/compose.yml"
error_file="$secret_dir/.postgres-password-rotation-error"

if [[ ${EUID:-$(id -u)} -ne 0 ]]; then
  echo "rotate-postgres-password.sh must run as root" >&2
  exit 1
fi

for required_file in "$secret_file" "$release_file" "$compose_file" "$install_dir/provision-server.sh"; do
  if [[ ! -f $required_file ]]; then
    echo "missing required deployment file: $required_file" >&2
    exit 1
  fi
done

new_password="$(openssl rand -hex 32)"
rm -f "$error_file"

if ! {
  printf '\\set new_password %s\n' "$new_password"
  printf "ALTER ROLE bcai PASSWORD :'new_password';\n"
} | docker compose --env-file "$release_file" -f "$compose_file" exec -T postgres \
  psql -v ON_ERROR_STOP=1 -U bcai -d bcai_projection >/dev/null 2>"$error_file"; then
  chmod 0600 "$error_file"
  echo "PostgreSQL password rotation failed; inspect the root-only error file" >&2
  exit 1
fi
rm -f "$error_file"

temp_secret="$(mktemp "$secret_dir/.secrets.env.XXXXXX")"
trap 'rm -f "$temp_secret"' EXIT
while IFS= read -r line || [[ -n $line ]]; do
  if [[ $line == POSTGRES_PASSWORD=* ]]; then
    printf 'POSTGRES_PASSWORD=%s\n' "$new_password"
  else
    printf '%s\n' "$line"
  fi
done <"$secret_file" >"$temp_secret"
chmod 0600 "$temp_secret"
mv -f "$temp_secret" "$secret_file"
trap - EXIT

set -a
# shellcheck disable=SC1090
source "$release_file"
set +a
"$install_dir/provision-server.sh" >/dev/null

docker compose --env-file "$release_file" -f "$compose_file" rm -sf mainapi worker migrate >/dev/null
docker compose --env-file "$release_file" -f "$compose_file" up -d --force-recreate --wait postgres >/dev/null
docker compose --env-file "$release_file" -f "$compose_file" up -d migrate >/dev/null

for attempt in $(seq 1 60); do
  migrate_id="$(docker compose --env-file "$release_file" -f "$compose_file" ps --all -q migrate)"
  if [[ -n $migrate_id ]]; then
    migrate_status="$(docker inspect --format '{{.State.Status}}:{{.State.ExitCode}}' "$migrate_id")"
    if [[ $migrate_status == exited:0 ]]; then
      break
    fi
    if [[ $migrate_status == exited:* ]]; then
      echo "migration failed after credential rotation" >&2
      exit 1
    fi
  fi
  sleep 2
done
if [[ ${migrate_status:-} != exited:0 ]]; then
  echo "migration timed out after credential rotation" >&2
  exit 1
fi

docker compose --env-file "$release_file" -f "$compose_file" up -d --no-deps mainapi worker >/dev/null

for service in mainapi worker; do
  healthy=false
  for attempt in $(seq 1 90); do
    container_id="$(docker compose --env-file "$release_file" -f "$compose_file" ps -q "$service")"
    if [[ -n $container_id ]]; then
      health="$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$container_id")"
      if [[ $health == healthy ]]; then
        healthy=true
        break
      fi
      if [[ $health == unhealthy || $health == exited || $health == dead ]]; then
        echo "$service failed health validation" >&2
        exit 1
      fi
    fi
    sleep 2
  done
  if [[ $healthy != true ]]; then
    echo "$service health validation timed out" >&2
    exit 1
  fi
done

echo "PostgreSQL credential rotated and application services restarted"
