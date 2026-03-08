@echo off
REM ========================================
REM   BC AI Cloud UAT - Build and Deploy
REM ========================================

echo ========================================
echo   BC AI Cloud UAT - Build and Deploy
echo ========================================
echo.

REM Step 1: Generate build-info.json
echo [1/3] Generating build-info.json...
powershell -ExecutionPolicy Bypass -File scripts/generate_build_info.ps1
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ========================================
    echo   [ERROR] Failed to generate build-info.json
    echo ========================================
    pause
    exit /b 1
)

REM Step 2: Build Flutter Web with increased heap size
echo [2/3] Building Flutter Web (UAT)...
echo.
set DART_VM_OPTIONS=--max-heap-size=4096m
call flutter build web -t lib/main_bcaiuat.dart --release --no-tree-shake-icons --no-wasm-dry-run --pwa-strategy none
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ========================================
    echo   [ERROR] Flutter build failed
    echo ========================================
    pause
    exit /b 1
)

REM Step 3: Deploy to Firebase
echo.
echo [3/3] Deploying to Firebase (UAT)...
echo.
call firebase deploy --only hosting:bcai-uat
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ========================================
    echo   [ERROR] Firebase deploy failed
    echo ========================================
    pause
    exit /b 1
)

echo.
echo ========================================
echo   [SUCCESS] Deploy completed successfully!
echo   URL: https://smlai-cloud-uat.web.app
echo ========================================
echo.
pause
