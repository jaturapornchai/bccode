param(
  [Parameter(Mandatory = $true)]
  [string]$ImageTag,

  [string]$Image = "ghcr.io/jaturapornchai/bccode-mainapi",
  [string]$Namespace = "bc-ai-account",
  [string]$KustomizePath = "..\base"
)

$ErrorActionPreference = "Stop"

$basePath = Resolve-Path (Join-Path $PSScriptRoot $KustomizePath)

kubectl config current-context | Out-Host
kubectl apply -k $basePath
kubectl -n $Namespace set image deployment/mainapi mainapi="${Image}:${ImageTag}"
kubectl -n $Namespace set image deployment/transaction-consumer transaction-consumer="${Image}:${ImageTag}"
kubectl -n $Namespace rollout status deployment/mainapi --timeout=180s
kubectl -n $Namespace rollout status deployment/transaction-consumer --timeout=180s
kubectl -n $Namespace get pods,svc,hpa,pdb,ingress
