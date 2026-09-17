@echo off
echo ===================================================
echo   Restarting Docker Desktop (Administrator Mode)
echo ===================================================

echo [1/3] Terminating stuck Docker processes...
taskkill /F /IM "Docker Desktop.exe" /T 2>nul
taskkill /F /IM "com.docker.backend.exe" /T 2>nul
taskkill /F /IM "com.docker.build.exe" /T 2>nul
taskkill /F /IM "com.docker.service.exe" /T 2>nul

echo [2/3] Starting Docker Desktop Service...
net start com.docker.service 2>nul

echo [3/3] Launching Docker Desktop...
start "" "C:\Program Files\Docker\Docker\Docker Desktop.exe"

echo ===================================================
echo   Docker Desktop restart initiated successfully!
echo ===================================================
timeout /t 5
