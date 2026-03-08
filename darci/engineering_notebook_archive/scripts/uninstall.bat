@echo off
REM darci Uninstall Script for Windows
REM Usage: uninstall.bat [-y]
REM   -y  Skip confirmation prompt

setlocal
echo 🐈 darci - Uninstallation (Windows)
echo =======================================

REM Check for -y flag
set "CONFIRM=n"
if "%~1"=="-y" set "CONFIRM=y"
if "%~1"=="-Y" set "CONFIRM=y"

if "%CONFIRM%"=="n" (
    echo.
    echo This will remove:
    echo   - Python dependencies (virtual environment)
    echo   - Go binary (darci-go.exe)
    echo   - Configuration files
    echo   - Workspace files
    echo   - Start Menu shortcuts
    echo.
    set /p CONFIRM="Are you sure you want to continue? (y/N): "
)

if /i not "%CONFIRM%"=="y" (
    echo Uninstallation cancelled.
    goto :end
)

REM ========== Stop Running Processes ==========
echo.
echo [1/5] Stopping running processes...
taskkill /F /IM darci-go.exe 2>nul
taskkill /F /IM python.exe /FI "WINDOWTITLE eq darci*" 2>nul

REM ========== Remove Python Bot ==========
echo.
echo [2/5] Removing Python Bot...
cd /d "%~dp0..\darci-python"
if exist ".venv" rmdir /S /Q .venv
if exist "darci_env" rmdir /S /Q darci_env

REM ========== Remove Go Bot ==========
echo.
echo [3/5] Removing Go Bot...
cd /d "%~dp0..\darci-go"
if exist "darci-go.exe" del /Q darci-go.exe
if exist "darci-go" del /Q darci-go

REM ========== Remove Configuration ==========
echo.
echo [4/5] Removing configuration...
set "CONFIG_DIR=%APPDATA%\darci"
if exist "%CONFIG_DIR%" rmdir /S /Q "%CONFIG_DIR%"

REM ========== Remove Workspace ==========
echo.
echo [5/5] Removing workspace...
set "WORKSPACE=%USERPROFILE%\darci-workspace"
if exist "%WORKSPACE%" rmdir /S /Q "%WORKSPACE%"

REM ========== Remove Start Menu Shortcuts ==========
set "SHORTCUT_DIR=%APPDATA%\Microsoft\Windows\Start Menu\Programs\darci"
if exist "%SHORTCUT_DIR%" rmdir /S /Q "%SHORTCUT_DIR%"

echo.
echo =======================================
echo ✓ Uninstallation complete!
echo =======================================
echo.
echo To reinstall, run: install.bat [API_KEY]
echo Example: install.bat AIzaSyCSvcSZsC8Bg1k343y9l3as3vlOrhsXRSw
echo.

:end
endlocal
