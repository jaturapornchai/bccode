@echo off
REM ========================================
REM   BC AI Cloud DEV - Build and Deploy
REM ========================================

echo ========================================
echo   BC AI Cloud DEV - Build and Deploy
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

REM Step 2: Build Flutter Web
echo [2/3] Building Flutter Web (DEV)...
echo.
call flutter build web -t lib/main_bcaidev.dart --release --no-tree-shake-icons --no-wasm-dry-run --pwa-strategy none
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
echo [3/3] Deploying to Firebase (DEV)...
echo.
call firebase deploy --only hosting:bcai-dev
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
echo   URL: https://smlai-cloud-dev-2.web.app
echo ========================================
echo.
