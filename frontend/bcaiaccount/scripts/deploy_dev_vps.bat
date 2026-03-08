@echo off
REM ========================================
REM   BC AI Cloud DEV - Deploy to VPS
REM   Target: https://dev.bcaicloud.com/
REM   VPS: 5.223.69.66 (Hetzner)
REM   Frontend: nginx on host (no Docker)
REM   Backend: Docker containers (independent)
REM ========================================

echo ========================================
echo   BC AI Cloud DEV - Deploy to VPS
echo   Target: https://dev.bcaicloud.com/
echo ========================================
echo.

SET VPS_HOST=root@5.223.69.66
SET SSH_KEY=%USERPROFILE%\.ssh\hetzner_deploy
SET VPS_WEB_DIR=/opt/bcaicloud/frontend/web
SET GOAPI_URL=https://dev.bcaicloud.com/goapi

REM Step 1: Generate build-info.json
echo [1/5] Generating build-info.json...
powershell -ExecutionPolicy Bypass -File scripts/generate_build_info.ps1
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo   [ERROR] Failed to generate build-info.json
    pause
    exit /b 1
)

REM Step 2: Build Flutter Web
echo [2/5] Building Flutter Web (DEV)...
echo.
call flutter build web -t lib/main_bcaidev.dart --release --no-tree-shake-icons --no-wasm-dry-run --pwa-strategy none
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo   [ERROR] Flutter build failed
    pause
    exit /b 1
)

REM Step 3: Create VPS config.json (override localhost with VPS URL)
echo [3/5] Creating VPS config.json...
echo {"goapi_url": "%GOAPI_URL%"} > build\web\config.json

REM Step 4: Upload to VPS
echo [4/5] Uploading to VPS...
echo.
scp -i %SSH_KEY% -r build/web/* %VPS_HOST%:%VPS_WEB_DIR%/
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo   [ERROR] Upload to VPS failed
    pause
    exit /b 1
)

REM Step 5: Fix font filenames (space -> %%20) + reload nginx
echo [5/5] Fixing fonts and reloading nginx...
ssh -i %SSH_KEY% %VPS_HOST% "cd %VPS_WEB_DIR% && find . -name '* *' -type f -exec bash -c 'for f; do dir=$(dirname \"$f\"); base=$(basename \"$f\"); newname=$(echo \"$base\" | sed \"s/ /%%20/g\"); if [ \"$base\" != \"$newname\" ]; then mv \"$f\" \"$dir/$newname\"; fi; done' _ {} + ; nginx -s reload"
if %ERRORLEVEL% NEQ 0 (
    echo   [WARNING] Post-deploy step failed - check VPS
)

echo.
echo ========================================
echo   [SUCCESS] Deploy completed!
echo   URL: https://dev.bcaicloud.com/
echo   Note: nginx serves files independently of Docker backend
echo ========================================
echo.
