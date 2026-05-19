# K3d Local Overlay

## Objective

Run the K3s manifests on Docker Desktop through k3d for optional local/staging verification before using a real production K3s cluster. This is not the default development workflow while the project is still starting out.

## Workflow

1. Install k3d once:

```powershell
winget install k3d.k3d
```

2. Create the local K3s-in-Docker cluster:

```powershell
backend/cluster/k3s/scripts/k3d-create.ps1
```

3. Build, import, and deploy:

```powershell
backend/cluster/k3s/scripts/k3d-build-deploy.ps1 `
  -BootstrapJson backend/bootstrap.json
```

When the image was already imported and only manifests/config changed:

```powershell
backend/cluster/k3s/scripts/k3d-build-deploy.ps1 `
  -BootstrapJson backend/bootstrap.json `
  -SkipBuild
```

By default the script creates a temporary k3d bootstrap file and converts Docker Compose endpoints to Docker Desktop host endpoints:

- `clickhouse:9000` -> `host.k3d.internal:19002`
- `http://seaweedfs-filer:8333` -> `http://host.k3d.internal:18333`
- Kafka/Redis are injected through k3d overlay environment variables. The k3d cluster should join `backend_app-network` so `kafka:29092` and `redis:6379` resolve like Docker Compose services.

4. Verify:

```powershell
curl.exe -f http://localhost:18088/healthz
curl.exe -f http://localhost:18088/goapi/api/health
kubectl -n bc-ai-account get pods,svc,ingress
```

5. Delete local cluster when done:

```powershell
backend/cluster/k3s/scripts/k3d-delete.ps1
```

## Config

- Image: `bc-ai-account/mainapi:local`
- Namespace: `bc-ai-account`
- Ingress host: `localhost`
- Host port: `18088`

## Dependency

- Docker Desktop running.
- `kubectl` available.
- `k3d` installed.
- A real local `bootstrap.json`, not committed into git.

## Limitation

- This overlay is not HA.
- It does not prove 10,000 concurrent users.
- It is for manifest/runtime smoke tests only.
