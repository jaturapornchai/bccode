# install-pgvector-win.ps1
# ติดตั้ง pgvector extension สำหรับ PostgreSQL 18 บน Windows
# ใช้ vcpkg build หรือ download pre-built binary
#
# วิธีใช้ (Run as Administrator):
#   powershell -ExecutionPolicy Bypass -File install-pgvector-win.ps1

$ErrorActionPreference = "Stop"
$PG_VERSION = "18"
$PGVECTOR_VERSION = "0.8.0"
$PG_DIR = "C:\Program Files\PostgreSQL\$PG_VERSION"

# ตรวจว่า PostgreSQL ติดตั้งอยู่
if (-not (Test-Path "$PG_DIR\bin\pg_config.exe")) {
    Write-Host "ERROR: PostgreSQL $PG_VERSION not found at $PG_DIR" -ForegroundColor Red
    Write-Host "Please update PG_DIR variable in this script" -ForegroundColor Yellow
    exit 1
}

Write-Host "=== Installing pgvector $PGVECTOR_VERSION for PostgreSQL $PG_VERSION ===" -ForegroundColor Cyan

# ตรวจว่ามี Visual Studio Build Tools
$vsWhere = "${env:ProgramFiles(x86)}\Microsoft Visual Studio\Installer\vswhere.exe"
$hasVS = $false
if (Test-Path $vsWhere) {
    $vsPath = & $vsWhere -latest -property installationPath 2>$null
    if ($vsPath) { $hasVS = $true }
}

# Method 1: ลอง download pre-built binary จาก GitHub releases
Write-Host "`n--- Trying pre-built binary ---" -ForegroundColor Yellow
$downloadUrl = "https://github.com/pgvector/pgvector/releases/download/v$PGVECTOR_VERSION/pgvector-v$PGVECTOR_VERSION-pg$PG_VERSION-windows-x64.zip"
$tempDir = "$env:TEMP\pgvector-install"
$zipFile = "$tempDir\pgvector.zip"

New-Item -ItemType Directory -Force -Path $tempDir | Out-Null

try {
    Write-Host "Downloading from $downloadUrl ..."
    Invoke-WebRequest -Uri $downloadUrl -OutFile $zipFile -UseBasicParsing
    Write-Host "Download successful!" -ForegroundColor Green

    # Extract
    Expand-Archive -Path $zipFile -DestinationPath "$tempDir\pgvector" -Force

    # Copy files to PostgreSQL directories
    $extractDir = Get-ChildItem "$tempDir\pgvector" -Directory | Select-Object -First 1
    if (-not $extractDir) { $extractDir = Get-Item "$tempDir\pgvector" }

    # Find and copy .dll, .sql, .control files
    Get-ChildItem $extractDir.FullName -Recurse -Filter "*.dll" | ForEach-Object {
        Copy-Item $_.FullName "$PG_DIR\lib\" -Force
        Write-Host "  Copied $($_.Name) -> lib\" -ForegroundColor Green
    }
    Get-ChildItem $extractDir.FullName -Recurse -Filter "vector.control" | ForEach-Object {
        Copy-Item $_.FullName "$PG_DIR\share\extension\" -Force
        Write-Host "  Copied $($_.Name) -> share\extension\" -ForegroundColor Green
    }
    Get-ChildItem $extractDir.FullName -Recurse -Filter "vector--*.sql" | ForEach-Object {
        Copy-Item $_.FullName "$PG_DIR\share\extension\" -Force
        Write-Host "  Copied $($_.Name) -> share\extension\" -ForegroundColor Green
    }

    Write-Host "`npgvector installed successfully from pre-built binary!" -ForegroundColor Green
}
catch {
    Write-Host "Pre-built binary not available: $_" -ForegroundColor Yellow

    if (-not $hasVS) {
        Write-Host "`nERROR: Visual Studio Build Tools required for source build" -ForegroundColor Red
        Write-Host "Install from: https://visualstudio.microsoft.com/downloads/#build-tools-for-visual-studio-2022" -ForegroundColor Yellow
        Write-Host "Or download pgvector manually from: https://github.com/pgvector/pgvector/releases" -ForegroundColor Yellow
        exit 1
    }

    # Method 2: Build from source
    Write-Host "`n--- Building from source ---" -ForegroundColor Yellow
    $sourceUrl = "https://github.com/pgvector/pgvector/archive/refs/tags/v$PGVECTOR_VERSION.zip"

    Write-Host "Downloading source from $sourceUrl ..."
    Invoke-WebRequest -Uri $sourceUrl -OutFile "$tempDir\pgvector-src.zip" -UseBasicParsing
    Expand-Archive -Path "$tempDir\pgvector-src.zip" -DestinationPath "$tempDir\src" -Force

    $srcDir = Get-ChildItem "$tempDir\src" -Directory | Select-Object -First 1

    # Use nmake with pg_config
    Push-Location $srcDir.FullName
    try {
        # Find vcvarsall.bat
        $vcvars = & $vsWhere -latest -find "VC\Auxiliary\Build\vcvarsall.bat" 2>$null
        if (-not $vcvars) {
            throw "vcvarsall.bat not found"
        }

        # Build
        $env:PGROOT = $PG_DIR
        cmd /c "`"$vcvars`" x64 && set `"PG_CONFIG=$PG_DIR\bin\pg_config.exe`" && nmake /F Makefile.win && nmake /F Makefile.win install"

        if ($LASTEXITCODE -eq 0) {
            Write-Host "`npgvector built and installed successfully!" -ForegroundColor Green
        } else {
            throw "Build failed with exit code $LASTEXITCODE"
        }
    }
    finally {
        Pop-Location
    }
}

# Cleanup
Remove-Item -Recurse -Force $tempDir -ErrorAction SilentlyContinue

# Verify installation
Write-Host "`n--- Verifying installation ---" -ForegroundColor Yellow
$controlFile = "$PG_DIR\share\extension\vector.control"
if (Test-Path $controlFile) {
    Write-Host "vector.control found" -ForegroundColor Green
} else {
    Write-Host "WARNING: vector.control not found!" -ForegroundColor Red
}

$dllFile = "$PG_DIR\lib\vector.dll"
if (Test-Path $dllFile) {
    Write-Host "vector.dll found" -ForegroundColor Green
} else {
    Write-Host "WARNING: vector.dll not found!" -ForegroundColor Red
}

Write-Host "`n=== Done ===" -ForegroundColor Cyan
Write-Host "Now run in psql:" -ForegroundColor Yellow
Write-Host "  CREATE EXTENSION vector;" -ForegroundColor White
Write-Host "`nThen rebuild embeddings:" -ForegroundColor Yellow
Write-Host "  rebuild_embeddings { holdingcode: 'YOUR_HOLDING_CODE', force_all: true }" -ForegroundColor White
