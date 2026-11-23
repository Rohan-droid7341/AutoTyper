@echo off
:: Batch script to add Windows Defender Firewall rule for Auto Typer
echo ========================================================
echo   Auto Typer - Windows Defender Firewall Configurator
echo ========================================================
echo.

:: Check for administrative permissions
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo [*] Requesting administrator elevation...
    powershell -Command "Start-Process cmd -ArgumentList '/c \"%~f0\"' -Verb RunAs"
    exit /b
)

echo Adding inbound firewall rule for Auto Typer (Port 8080-8090)...
netsh advfirewall firewall delete rule name="AutoTyper" >nul 2>&1
netsh advfirewall firewall add rule name="AutoTyper" dir=in action=allow protocol=TCP localport=8080-8090 profile=any >nul 2>&1

if %errorLevel% equ 0 (
    echo.
    echo ========================================================
    echo  [SUCCESS] Windows Firewall rule added successfully!
    echo  Phones on your Wi-Fi can now connect without issues.
    echo ========================================================
) else (
    echo.
    echo [ERROR] Failed to add firewall rule. Please run as Administrator.
)

echo.
pause
