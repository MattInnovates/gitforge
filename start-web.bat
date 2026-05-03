@echo off
setlocal

cd /d "%~dp0"

echo [1/3] Building web server...
go build -o tmp\web.exe .\cmd\web
if errorlevel 1 (
  echo Build failed. Server not started.
  exit /b 1
)

echo [2/3] Stopping existing server on port 3000 (if running)...
for /f "tokens=5" %%p in ('netstat -ano ^| findstr ":3000" ^| findstr "LISTENING"') do (
  taskkill /PID %%p /F >nul 2>nul
)

echo [3/3] Starting web server...
start "GitForge Web" /D "%~dp0" cmd /c "tmp\web.exe"

echo Server started. Open http://localhost:3000
exit /b 0
