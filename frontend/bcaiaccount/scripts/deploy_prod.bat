@echo off
REM ========================================
REM   BC AI Cloud PRODUCTION - Build and Deploy
REM ========================================

echo ========================================
echo   BC AI Cloud PRODUCTION - Build and Deploy
echo ========================================
echo.
echo [WARNING] You are about to deploy to PRODUCTION!
echo.
set /p CONFIRM="Type 'YES' to continue: "

if not "%CONFIRM%"=="YES" (
    echo.
    echo [CANCELLED] Deployment cancelled.
    echo.
    pause
    exit /b 0
)

echo.
echo Proceeding with PRODUCTION deployment...
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
echo [2/3] Building Flutter Web (PRODUCTION)...
echo.
set DART_VM_OPTIONS=--max-heap-size=4096m
call flutter build web -t lib/main_bcaiprod.dart --release --no-tree-shake-icons --pwa-strategy none
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
echo [3/3] Deploying to Firebase (PRODUCTION)...
echo.
call firebase deploy --only hosting:bcaicloud
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
echo   [SUCCESS] PRODUCTION Deploy completed!
echo   URL: https://bcai-cloud.web.app
echo ========================================
echo.
pause
