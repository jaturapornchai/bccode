[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"

$backendRoot = Split-Path -Parent $PSScriptRoot
$minioEnvPath = Join-Path $backendRoot "minio.local.env"
$storageEnvPath = Join-Path $backendRoot "storage.local.env"
$utf8NoBom = [Text.UTF8Encoding]::new($false)

function New-RandomHex([int]$byteCount) {
    $buffer = New-Object byte[] $byteCount
    $generator = [Security.Cryptography.RandomNumberGenerator]::Create()
    try {
        $generator.GetBytes($buffer)
    } finally {
        $generator.Dispose()
    }
    return ([BitConverter]::ToString($buffer) -replace "-", "").ToLowerInvariant()
}

function Read-EnvFile([string]$path) {
    $values = @{}
    foreach ($line in [IO.File]::ReadAllLines($path)) {
        $trimmed = $line.Trim()
        if ($trimmed.Length -eq 0 -or $trimmed.StartsWith("#")) {
            continue
        }
        $parts = $trimmed.Split("=", 2)
        if ($parts.Count -eq 2) {
            $values[$parts[0]] = $parts[1]
        }
    }
    return $values
}

function Protect-LocalSecretFile([string]$path) {
    if ($env:OS -ne "Windows_NT") {
        return
    }
    $identity = [Security.Principal.WindowsIdentity]::GetCurrent().Name
    & icacls.exe $path /inheritance:r /grant:r "${identity}:(F)" | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "Unable to restrict ACL on $path"
    }
}

if (Test-Path -LiteralPath $minioEnvPath) {
    $minioValues = Read-EnvFile $minioEnvPath
    foreach ($required in @(
        "MINIO_ROOT_USER",
        "MINIO_ROOT_PASSWORD",
        "MINIO_APP_ACCESS_KEY",
        "MINIO_APP_SECRET_KEY",
        "MINIO_BUCKET_NAME"
    )) {
        if (-not $minioValues.ContainsKey($required) -or [string]::IsNullOrWhiteSpace($minioValues[$required])) {
            throw "$minioEnvPath is missing $required"
        }
    }
} else {
    if (Test-Path -LiteralPath $storageEnvPath) {
        throw "$storageEnvPath exists without $minioEnvPath; refusing to overwrite inconsistent credentials"
    }

    $minioValues = @{
        MINIO_ROOT_USER      = "bcairoot$(New-RandomHex 6)"
        MINIO_ROOT_PASSWORD  = New-RandomHex 32
        MINIO_APP_ACCESS_KEY = "bcaiapp$(New-RandomHex 6)"
        MINIO_APP_SECRET_KEY = New-RandomHex 32
        MINIO_BUCKET_NAME    = "bcai-account"
    }
    $minioLines = @(
        "MINIO_ROOT_USER=$($minioValues.MINIO_ROOT_USER)",
        "MINIO_ROOT_PASSWORD=$($minioValues.MINIO_ROOT_PASSWORD)",
        "MINIO_APP_ACCESS_KEY=$($minioValues.MINIO_APP_ACCESS_KEY)",
        "MINIO_APP_SECRET_KEY=$($minioValues.MINIO_APP_SECRET_KEY)",
        "MINIO_BUCKET_NAME=$($minioValues.MINIO_BUCKET_NAME)"
    )
    [IO.File]::WriteAllLines($minioEnvPath, $minioLines, $utf8NoBom)
    Protect-LocalSecretFile $minioEnvPath
}

if (-not (Test-Path -LiteralPath $storageEnvPath)) {
    $storageLines = @(
        "S3_ENDPOINT=http://minio:9000",
        "S3_REGION=us-east-1",
        "S3_ACCESS_KEY_ID=$($minioValues.MINIO_APP_ACCESS_KEY)",
        "S3_SECRET_ACCESS_KEY=$($minioValues.MINIO_APP_SECRET_KEY)",
        "S3_BUCKET_NAME=$($minioValues.MINIO_BUCKET_NAME)",
        "S3_FORCE_PATH_STYLE=true",
        "STORAGE_ALLOW_PRESIGNED_URL=false"
    )
    [IO.File]::WriteAllLines($storageEnvPath, $storageLines, $utf8NoBom)
    Protect-LocalSecretFile $storageEnvPath
}

Write-Host "Local MinIO credentials are ready (secret values were not printed)."
