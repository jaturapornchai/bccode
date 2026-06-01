param(
  [switch]$RebuildBuilder,
  [int]$HealthTimeoutSeconds = 45
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$backendRoot = (Resolve-Path (Join-Path $scriptDir "..")).Path
$builderImage = "backend-mainapi-local-builder:latest"
$containerName = "mainapi"
$goBuildCache = "backend-mainapi-go-build-cache"
$goModCache = "backend-mainapi-go-mod-cache"
$outputDir = Join-Path $backendRoot "tmp\fastdeploy"
$outputBinary = Join-Path $outputDir "go-app"

function Invoke-Step {
  param(
    [string]$Name,
    [scriptblock]$Action
  )
  $elapsed = Measure-Command { & $Action }
  [pscustomobject]@{
    step = $Name
    seconds = [math]::Round($elapsed.TotalSeconds, 2)
  }
}

function Test-DockerImage {
  param([string]$ImageName)
  $existing = docker image inspect $ImageName 2>$null
  return $LASTEXITCODE -eq 0 -and $existing
}

function Wait-Healthz {
  param([int]$TimeoutSeconds)
  $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
  do {
    try {
      $result = curl.exe --max-time 2 -s http://localhost:8888/healthz
      if ($result -eq "ok") {
        return
      }
    } catch {
      Start-Sleep -Milliseconds 500
    }
    Start-Sleep -Milliseconds 500
  } while ((Get-Date) -lt $deadline)
  throw "mainapi healthz did not return ok within $TimeoutSeconds seconds"
}

New-Item -ItemType Directory -Force -Path $outputDir | Out-Null
$backendMount = $backendRoot.Replace("\", "/")
$outputMount = $outputDir.Replace("\", "/")

if ($RebuildBuilder -or -not (Test-DockerImage $builderImage)) {
  $tempContext = Join-Path ([System.IO.Path]::GetTempPath()) ("mainapi-builder-empty-" + [guid]::NewGuid().ToString("N"))
  New-Item -ItemType Directory -Force -Path $tempContext | Out-Null
  try {
    $dockerfile = @"
FROM golang:1.26-alpine
RUN apk add --no-cache alpine-sdk librdkafka-dev build-base
WORKDIR /src
"@
    $dockerfile | docker build -t $builderImage -f - $tempContext
    if ($LASTEXITCODE -ne 0) {
      throw "failed to build $builderImage"
    }
  } finally {
    Remove-Item -LiteralPath $tempContext -Recurse -Force -ErrorAction SilentlyContinue
  }
}

$containerExists = docker ps -a --format "{{.Names}}" | Where-Object { $_ -eq $containerName }
if (-not $containerExists) {
  Push-Location $backendRoot
  try {
    docker-compose up -d --no-deps --build mainapi
    if ($LASTEXITCODE -ne 0) {
      throw "docker-compose build/start failed"
    }
    Wait-Healthz -TimeoutSeconds $HealthTimeoutSeconds
    Write-Output ([pscustomobject]@{
      mode = "compose-build"
      reason = "mainapi container did not exist"
    } | ConvertTo-Json -Compress)
    return
  } finally {
    Pop-Location
  }
}

$results = @()
$results += Invoke-Step "compile" {
  docker run --rm `
    -v "${backendMount}:/src" `
    -v "${goBuildCache}:/root/.cache/go-build" `
    -v "${goModCache}:/go/pkg/mod" `
    -v "${outputMount}:/out" `
    -w /src `
    $builderImage `
    sh -c "CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /out/go-app -tags musl main.go"
  if ($LASTEXITCODE -ne 0) {
    throw "go build failed"
  }
}

$results += Invoke-Step "swap-restart-healthz" {
  docker stop $containerName | Out-Null
  if ($LASTEXITCODE -ne 0) {
    throw "failed to stop $containerName"
  }
  docker cp $outputBinary "${containerName}:/app/go-app" | Out-Null
  if ($LASTEXITCODE -ne 0) {
    throw "failed to copy binary to $containerName"
  }
  docker start $containerName | Out-Null
  if ($LASTEXITCODE -ne 0) {
    throw "failed to start $containerName"
  }
  Wait-Healthz -TimeoutSeconds $HealthTimeoutSeconds
}

$totalSeconds = [math]::Round(($results | Measure-Object -Property seconds -Sum).Sum, 2)
Write-Output ([pscustomobject]@{
  mode = "fast-binary-swap"
  total_seconds = $totalSeconds
  steps = $results
  binary = $outputBinary
} | ConvertTo-Json -Depth 4 -Compress)
