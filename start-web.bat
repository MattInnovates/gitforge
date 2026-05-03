@echo off
setlocal

cd /d "%~dp0"
if not exist tmp mkdir tmp

echo [1/4] Building web server...
go build -o tmp\web.exe .\cmd\web
if errorlevel 1 (
  echo Build failed. Server not started.
  exit /b 1
)

echo [2/4] Stopping existing server on port 3000 (if running)...
for /f "tokens=5" %%p in ('netstat -ano ^| findstr ":3000" ^| findstr "LISTENING"') do (
  taskkill /PID %%p /F >nul 2>nul
)

echo [3/4] Starting web server...
start "GitForge Web" /D "%~dp0" cmd /c "tmp\web.exe 1>tmp\\web.out.log 2>tmp\\web.err.log"

echo [4/4] Verifying startup...
set "STARTED="
for /l %%i in (1,1,20) do (
  for /f "tokens=5" %%p in ('netstat -ano ^| findstr ":3000" ^| findstr "LISTENING"') do (
    set "STARTED=1"
  )
  if defined STARTED goto :ok
  timeout /t 1 >nul
)

echo Server failed to start. Check tmp\web.err.log
exit /b 1

:ok
echo Server started. Open http://localhost:3000
exit /b 0
