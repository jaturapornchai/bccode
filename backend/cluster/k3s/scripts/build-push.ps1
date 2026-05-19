param(
  [Parameter(Mandatory = $true)]
  [string]$Image,

  [Parameter(Mandatory = $true)]
  [string]$Tag,

  [string]$Context = "..\..\..",
  [string]$Dockerfile = "..\..\..\Dockerfile"
)

$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot $Context)
$dockerfilePath = Resolve-Path (Join-Path $PSScriptRoot $Dockerfile)
$fullImage = "${Image}:${Tag}"

Write-Host "Building $fullImage from $dockerfilePath"
docker build -t $fullImage -f $dockerfilePath $root

Write-Host "Pushing $fullImage"
docker push $fullImage

Write-Host "Done: $fullImage"
