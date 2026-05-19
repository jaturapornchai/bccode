param(
  [string]$Namespace = "bc-ai-account",
  [string]$KustomizePath = "..\base"
)

$ErrorActionPreference = "Stop"

$basePath = Resolve-Path (Join-Path $PSScriptRoot $KustomizePath)

Write-Host "Rendering kustomize manifests"
kubectl kustomize $basePath | Out-Host

Write-Host "Checking current kubectl context"
kubectl config current-context | Out-Host

Write-Host "Checking workload status"
kubectl -n $Namespace get pods,svc,hpa,pdb,ingress

Write-Host "Checking rollout status"
kubectl -n $Namespace rollout status deployment/mainapi --timeout=60s
kubectl -n $Namespace rollout status deployment/transaction-consumer --timeout=60s
