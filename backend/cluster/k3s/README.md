# BC Ai Account K3s Production Path

## Objective

Use K3s as the production/high-concurrency runtime for BC Ai Account backend services. Docker Desktop and `docker-compose.yml` remain useful for local development, but they are not the target for 10,000 concurrent screens.

## What K3s Solves

- Runs multiple `mainapi` replicas behind a Kubernetes Service and Ingress.
- Runs transaction consumers separately from interactive API pods so Kafka/PG processing does not steal menu/login capacity.
- Adds rolling update, self-healing, readiness/liveness probes, and PodDisruptionBudget.
- Enables HorizontalPodAutoscaler when metrics-server is available.
- Gives one deployment model for single customer servers, VPS clusters, or on-prem nodes.

## What K3s Does Not Solve Alone

- It does not make PostgreSQL, MongoDB, Kafka, ClickHouse, Redis, or object storage fast by itself.
- It does not prove 10,000 concurrent screens. That needs load testing with real API paths and realistic report workloads.
- It does not fix mutable per-pod config. Current `bootstrap.json` writes from Settings are not safe when many replicas run.

## Minimum Production Shape

| Layer | Recommendation |
| --- | --- |
| K3s control plane | 3 server nodes, SSD-backed datastore, secrets encryption enabled |
| Worker nodes | Start with 3 workers, add nodes after load test evidence |
| App | `mainapi` Deployment, Service, Ingress, HPA, PDB |
| Config | Kubernetes Secret for `bootstrap.json`; later move to shared config service |
| Database | PostgreSQL and MongoDB must be external HA or clustered |
| Cache/queue | Redis and Kafka must be HA, not single local containers |
| Reports | Report generation should be separated into worker pods before heavy production use |
| Storage | Use HA object storage; single SeaweedFS volume is not enough for confidential large-image workloads |

## Files

- `base/namespace.yaml` - namespace
- `base/mainapi-secret.example.yaml` - example only, replace all values before use
- `base/mainapi-deployment.yaml` - Deployment, Service, HPA, PDB, Ingress
- `base/transaction-consumer-deployment.yaml` - Kafka consumer Deployment, HPA, PDB
- `base/kustomization.yaml` - apply entrypoint
- `overlays/k3d` - local K3s-on-Docker overlay for Docker Desktop/k3d
- `scripts/build-push.ps1` - build and push the backend image
- `scripts/create-bootstrap-secret.ps1` - generate/apply a Kubernetes Secret from a real local bootstrap file
- `scripts/deploy.ps1` - apply manifests, set images, and wait for rollout
- `scripts/verify.ps1` - render manifests and inspect live workload status
- `scripts/k3d-create.ps1` - create a local k3d cluster on Docker Desktop
- `scripts/k3d-build-deploy.ps1` - build local image, import it into k3d, create secret, and deploy
- `scripts/k3d-delete.ps1` - delete the local k3d cluster
- `scripts/verify-manifests.ps1` - render all kustomize bases/overlays
- `CAPACITY_10000.md` - load-test gate before claiming 10,000 concurrent screens

## Docker Desktop / k3d Local Test

Docker Desktop can run this K3s path through k3d. This is optional and is not the default development path while the system is still starting out. Use it only for smoke testing manifests before production:

```powershell
winget install k3d.k3d
backend/cluster/k3s/scripts/k3d-create.ps1
backend/cluster/k3s/scripts/k3d-build-deploy.ps1 `
  -BootstrapJson backend/bootstrap.json
curl.exe -f http://localhost:18088/healthz
curl.exe -f http://localhost:18088/goapi/api/health
```

When only manifests/config changed and the local image is already imported, redeploy faster:

```powershell
backend/cluster/k3s/scripts/k3d-build-deploy.ps1 `
  -BootstrapJson backend/bootstrap.json `
  -SkipBuild
```

The local k3d default port is `18088` so it does not collide with the existing Docker Compose `mainapi` on `8888`. The create script joins `backend_app-network` when that Docker Compose network exists, so Kafka can use the internal listener `kafka:29092`. The deploy script also creates a temporary bootstrap file for k3d so Docker Compose service names such as `clickhouse` and `seaweedfs-filer` are converted to Docker Desktop host endpoints. This is not a production HA environment. It only proves the manifests can run on local Docker Desktop.

## Current Startup Policy

Do not force k3s/k3d into the normal startup workflow yet. For early development:

- Backend stays on Docker Compose/local services.
- Frontend stays local Next.js.
- K3s manifests stay maintained and renderable.
- Any new backend service must be deployable as a stateless container with external config/secrets.
- Avoid local-only assumptions such as hardcoded Docker Compose service names in business code.

## First Install Notes

Install K3s with production flags on server nodes:

```bash
curl -sfL https://get.k3s.io | sh -s - server \
  --secrets-encryption \
  --cluster-cidr=10.42.0.0/16 \
  --service-cidr=10.43.0.0/16
```

For HA, use multiple server nodes and a supported external datastore or embedded etcd. Do not use default SQLite for HA.

## Deploy

Build and push the backend image:

```powershell
backend/cluster/k3s/scripts/build-push.ps1 `
  -Image ghcr.io/jaturapornchai/bccode-mainapi `
  -Tag 2026.05.17-001
```

Create the Kubernetes Secret from a real local `bootstrap.json`. Do not commit the rendered Secret:

```powershell
backend/cluster/k3s/scripts/create-bootstrap-secret.ps1 `
  -BootstrapJson backend/bootstrap.json `
  -Apply
```

For documentation only, `base/mainapi-secret.example.yaml` shows the expected shape after replacing placeholders:

```bash
kubectl apply -f backend/cluster/k3s/base/mainapi-secret.example.yaml
```

Then deploy the app manifests:

```powershell
backend/cluster/k3s/scripts/deploy.ps1 `
  -Image ghcr.io/jaturapornchai/bccode-mainapi `
  -ImageTag 2026.05.17-001
```

## Verification

```bash
kubectl -n bc-ai-account rollout status deployment/mainapi
kubectl -n bc-ai-account rollout status deployment/transaction-consumer
kubectl -n bc-ai-account get hpa mainapi
kubectl -n bc-ai-account get hpa transaction-consumer
kubectl -n bc-ai-account logs deploy/mainapi --tail=100
kubectl -n bc-ai-account logs deploy/transaction-consumer --tail=100
curl -f https://YOUR_DOMAIN/healthz
curl -f https://YOUR_DOMAIN/goapi/api/health
```

## 10,000 Screen Gate

Do not mark the system as ready for 10,000 concurrent screens until all pass:

- p95 API latency stays within the product target during peak load.
- Error rate stays below the target threshold.
- PostgreSQL pool wait time and slow queries are under control.
- MongoDB indexes and query plans are checked.
- ClickHouse report queries do not block interactive API traffic.
- Kafka lag stays bounded.
- Object storage upload/download latency is acceptable.
- HPA scales up and down without repeated pod restarts.

## Current Blockers Before Production Scale

- `bootstrap.json` currently contains runtime config and can be written by Settings. Multi-replica production needs a shared config store or a read-only Secret plus a controlled rollout process.
- Existing `docker-compose.yml` infrastructure is single-node. It is acceptable for dev, not for 10,000 concurrent production users.
- Load tests in `backend/loadtest` exist but need a real 10,000-screen scenario and backend data set before capacity can be claimed.
- Transaction consumers must be idempotent before scaling replicas aggressively because Kafka can redeliver messages.

## References

- K3s requirements and large-cluster sizing: https://docs.k3s.io/installation/requirements
- K3s HA datastore options: https://docs.k3s.io/datastore
- K3s HA with external DB: https://docs.k3s.io/datastore/ha
- K3s secrets encryption: https://docs.k3s.io/security/secrets-encryption
- Kubernetes HPA: https://kubernetes.io/docs/concepts/workloads/autoscaling/horizontal-pod-autoscale/
- Kubernetes Ingress: https://kubernetes.io/docs/concepts/services-networking/ingress/
