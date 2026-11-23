@echo off
setlocal
cd /d "%~dp0\.."

echo ==========================================================
echo   Building Auto Typer for Windows (AMD64)...
echo ==========================================================

go build -ldflags="-H=windowsgui -s -w" -o windows\server.exe .\cmd\server
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Failed to compile server.exe
    exit /b %ERRORLEVEL%
)

go build -ldflags="-s -w" -o windows\stop.exe .\cmd\stop
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Failed to compile stop.exe
    exit /b %ERRORLEVEL%
)

echo.
echo ==========================================================
echo [SUCCESS] Windows binaries built successfully in windows\
echo ==========================================================
