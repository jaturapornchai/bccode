param(
  [string]$ClusterName = "bc-ai-account",
  [string]$Image = "bc-ai-account/mainapi",
  [string]$Tag = "local",
  [int]$HostPort = 18088,
  [switch]$UseDockerDesktopEndpoints = $true,
  [switch]$SkipBuild,
  [Parameter(Mandatory = $true)]
  [string]$BootstrapJson
)

$ErrorActionPreference = "Stop"

if (-not (Get-Command k3d -ErrorAction SilentlyContinue)) {
  throw "k3d is not installed. Install with: winget install k3d.k3d"
}

$backendRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..\..")
$overlayPath = Resolve-Path (Join-Path $PSScriptRoot "..\overlays\k3d")
$bootstrapPath = Resolve-Path $BootstrapJson
$fullImage = "${Image}:${Tag}"
$secretBootstrapPath = $bootstrapPath

function Convert-BootstrapForK3d {
  param(
    [Parameter(Mandatory = $true)]
    [string]$SourcePath
  )

  $tempPath = Join-Path $env:TEMP "bc-ai-account-k3d-bootstrap.json"
  $json = Get-Content -Raw $SourcePath | ConvertFrom-Json

  if ($json.clickhouse) {
    if ($json.clickhouse.host -eq "clickhouse") {
      $json.clickhouse.host = "host.k3d.internal"
    }
    if ($json.clickhouse.port -eq "9000") {
      $json.clickhouse.port = "19002"
    }
  }

  if ($json.storage -and $json.storage.s3_endpoint -eq "http://seaweedfs-filer:8333") {
    $json.storage.s3_endpoint = "http://host.k3d.internal:18333"
  }

  $json | ConvertTo-Json -Depth 20 | Set-Content -Path $tempPath -Encoding UTF8
  return (Resolve-Path $tempPath)
}

if ($UseDockerDesktopEndpoints) {
  $secretBootstrapPath = Convert-BootstrapForK3d -SourcePath $bootstrapPath
  Write-Host "Using temporary k3d bootstrap with Docker Desktop endpoints."
}

kubectl config use-context "k3d-$ClusterName"

if (-not $SkipBuild) {
  docker build -t $fullImage -f (Join-Path $backendRoot "Dockerfile") $backendRoot
  k3d image import $fullImage -c $ClusterName
} else {
  Write-Host "Skipping image build/import. Using image already available in k3d: $fullImage"
}

& (Join-Path $PSScriptRoot "create-bootstrap-secret.ps1") `
  -BootstrapJson $secretBootstrapPath `
  -Apply

kubectl apply -k $overlayPath
kubectl -n bc-ai-account rollout restart deployment/mainapi deployment/transaction-consumer
kubectl -n bc-ai-account rollout status deployment/mainapi --timeout=180s
kubectl -n bc-ai-account rollout status deployment/transaction-consumer --timeout=180s
kubectl -n bc-ai-account get pods,svc,ingress

Write-Host "Local URLs:"
Write-Host "  http://localhost:$HostPort/healthz"
Write-Host "  http://localhost:$HostPort/goapi/api/health"
