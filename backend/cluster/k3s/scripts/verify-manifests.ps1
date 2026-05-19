$ErrorActionPreference = "Stop"

$paths = @(
  "..\base",
  "..\overlays\k3d"
)

foreach ($path in $paths) {
  $resolved = Resolve-Path (Join-Path $PSScriptRoot $path)
  Write-Host "Rendering $resolved"
  kubectl kustomize $resolved | Out-Null
}

Write-Host "All manifests rendered successfully."
