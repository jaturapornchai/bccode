# Generate build-info.json automatically
# This script creates a version file based on pubspec.yaml, timestamp, and git info

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Generating build-info.json" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Get current timestamp
$buildTime = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
$buildTimestamp = [int][double]::Parse((Get-Date -UFormat %s))

# Get git information (if available)
$gitHash = "unknown"
$gitBranch = "unknown"
try {
    $gitHash = git rev-parse --short HEAD 2>$null
    $gitBranch = git rev-parse --abbrev-ref HEAD 2>$null
    if ($LASTEXITCODE -ne 0) {
        $gitHash = "unknown"
        $gitBranch = "unknown"
    }
} catch {
    Write-Host "[WARNING] Git not available, using default values" -ForegroundColor Yellow
}

# Get version from pubspec.yaml and auto-increment build number
$pubspecPath = "pubspec.yaml"
if (Test-Path $pubspecPath) {
    $pubspecContent = Get-Content $pubspecPath -Raw
    if ($pubspecContent -match "version:\s*(\d+\.\d+\.\d+)\+(\d+)") {
        $version = $matches[1]
        $buildNumber = [int]$matches[2]
        
        # Auto-increment build number
        $newBuildNumber = $buildNumber + 1
        
        # Update pubspec.yaml with new build number
        $newVersionString = "version: $version+$newBuildNumber"
        $pubspecContent = $pubspecContent -replace "version:\s*\d+\.\d+\.\d+\+\d+", $newVersionString
        Set-Content -Path $pubspecPath -Value $pubspecContent -NoNewline
        
        Write-Host "[AUTO-INCREMENT] Build number: $buildNumber -> $newBuildNumber" -ForegroundColor Yellow
        
        # Use new build number
        $buildNumber = $newBuildNumber
    } else {
        $version = "1.0.0"
        $buildNumber = "1"
    }
} else {
    Write-Host "[ERROR] pubspec.yaml not found!" -ForegroundColor Red
    exit 1
}

# Create build-info.json
$buildInfo = @{
    version = $version
    buildNumber = [int]$buildNumber
    buildTime = $buildTime
    buildTimestamp = $buildTimestamp
    gitHash = $gitHash
    gitBranch = $gitBranch
} | ConvertTo-Json -Depth 10

# Ensure web directory exists
if (-not (Test-Path "web")) {
    Write-Host "[ERROR] web directory not found!" -ForegroundColor Red
    exit 1
}

# Write to web folder
$buildInfo | Out-File -FilePath "web/build-info.json" -Encoding UTF8 -NoNewline

Write-Host "[OK] Generated build-info.json:" -ForegroundColor Green
Write-Host ""
Write-Host "  Version:        $version" -ForegroundColor White
Write-Host "  Build Number:   $buildNumber" -ForegroundColor White
Write-Host "  Build Time:     $buildTime" -ForegroundColor White
Write-Host "  Build Timestamp: $buildTimestamp" -ForegroundColor White
Write-Host "  Git Hash:       $gitHash" -ForegroundColor White
Write-Host "  Git Branch:     $gitBranch" -ForegroundColor White
Write-Host ""
Write-Host "[SAVED] File saved to: web/build-info.json" -ForegroundColor Cyan
Write-Host ""
