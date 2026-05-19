param(
  [Parameter(Mandatory = $true)]
  [string]$BootstrapJson,

  [string]$Namespace = "bc-ai-account",
  [string]$SecretName = "mainapi-bootstrap",
  [switch]$Apply
)

$ErrorActionPreference = "Stop"

$bootstrapPath = Resolve-Path $BootstrapJson

if (-not (Test-Path $bootstrapPath)) {
  throw "bootstrap.json not found: $BootstrapJson"
}

if ($Apply) {
  kubectl create namespace $Namespace --dry-run=client -o yaml | kubectl apply -f -
}

$command = @(
  "create", "secret", "generic", $SecretName,
  "--namespace", $Namespace,
  "--from-file=bootstrap.json=$bootstrapPath",
  "--dry-run=client",
  "-o", "yaml"
)

if ($Apply) {
  kubectl @command | kubectl apply -f -
} else {
  kubectl @command
}
