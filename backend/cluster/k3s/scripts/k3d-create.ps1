param(
  [string]$ClusterName = "bc-ai-account",
  [int]$Agents = 2,
  [int]$HostPort = 18088,
  [int]$ApiPort = 6550,
  [string]$DockerNetwork = "backend_app-network"
)

$ErrorActionPreference = "Stop"

if (-not (Get-Command k3d -ErrorAction SilentlyContinue)) {
  throw "k3d is not installed. Install with: winget install k3d.k3d"
}

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
  throw "docker is not available. Start Docker Desktop first."
}

if (-not (Get-Command kubectl -ErrorAction SilentlyContinue)) {
  throw "kubectl is not available."
}

$existing = k3d cluster list 2>$null | Select-String -Pattern "^\s*$ClusterName\s"
if (-not $existing) {
  $args = @(
    "cluster", "create", $ClusterName,
    "--servers", "1",
    "--agents", "$Agents",
    "--api-port", "127.0.0.1:${ApiPort}",
    "--port", "127.0.0.1:${HostPort}:80@loadbalancer",
    "--wait"
  )

  if ($DockerNetwork -and (docker network inspect $DockerNetwork 2>$null)) {
    $args += @("--network", $DockerNetwork)
  } elseif ($DockerNetwork) {
    Write-Warning "Docker network '$DockerNetwork' not found. Creating k3d cluster with its own network."
  }

  k3d @args
}

kubectl config use-context "k3d-$ClusterName"
kubectl cluster-info
