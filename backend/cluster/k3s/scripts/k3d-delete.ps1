param(
  [string]$ClusterName = "bc-ai-account"
)

$ErrorActionPreference = "Stop"

if (-not (Get-Command k3d -ErrorAction SilentlyContinue)) {
  throw "k3d is not installed."
}

k3d cluster delete $ClusterName
