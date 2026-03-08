@echo off
REM darci Install Script for Windows
REM Usage: install.bat [GEMINI_API_KEY]
REM Example: install.bat AIzaSyCSvcSZsC8Bg1k343y9l3as3vlOrhsXRSw

setlocal EnableDelayedExpansion
echo 🐈 darci - Installation (Windows)
echo =====================================

REM Check for API key argument
set "API_KEY=%~1"

REM ========== Prerequisites Check ==========
echo.
echo [0/5] Checking prerequisites...
echo.

REM Check Python (3.11+)
echo Checking Python...
where python >nul 2>nul
if %errorlevel% neq 0 (
    echo Python is not installed. Installing via winget...
    winget install --id Python.Python.3.11 --silent --accept-package-agreements --accept-source-agreements
    if %errorlevel% neq 0 (
        echo Failed to install Python automatically.
        echo Please install Python 3.11+ from https://www.python.org/downloads/
        pause
        exit /b 1
    )
    REM Refresh PATH
    set "PATH=%PATH%;%USERPROFILE%\AppData\Local\Programs\Python\Python311\;%USERPROFILE%\AppData\Local\Programs\Python\Python311\Scripts\"
)

python --version >nul 2>nul
if %errorlevel% neq 0 (
    echo ERROR: Python still not found after installation attempt.
    echo Please install Python 3.11+ from https://www.python.org/downloads/
    pause
    exit /b 1
)

REM Check uv
echo Checking uv...
where uv >nul 2>nul
if %errorlevel% neq 0 (
    echo uv is not installed. Installing...
    powershell -ExecutionPolicy ByPass -c "irm https://astral.sh/uv/install.ps1 | iex"
    if %errorlevel% neq 0 (
        echo Failed to install uv.
        echo Please install uv from https://docs.astral.sh/uv/getting-started/installation/
        pause
        exit /b 1
    )
    REM Refresh PATH
    set "PATH=%PATH%;%USERPROFILE%\.local\bin"
)

REM Check Go
echo Checking Go...
set "GO_INSTALLED=0"

REM Check if go is in PATH
where go >nul 2>nul
if %errorlevel% equ 0 (
    set "GO_INSTALLED=1"
)

REM If not in PATH, check common installation locations
if %GO_INSTALLED% equ 0 (
    if exist "C:\Program Files\Go\bin\go.exe" (
        set "GO_INSTALLED=1"
        set "PATH=%PATH%;C:\Program Files\Go\bin"
    )
)

if %GO_INSTALLED% equ 0 (
    if exist "%USERPROFILE%\go\bin\go.exe" (
        set "GO_INSTALLED=1"
        set "PATH=%PATH%;%USERPROFILE%\go\bin"
    )
)

if %GO_INSTALLED% equ 0 (
    echo Go is not installed. Installing via winget...
    winget install --id GoLang.Go --silent --accept-package-agreements --accept-source-agreements
    REM winget may return non-zero even on success if already installed
    timeout /t 5 /nobreak >nul

    REM Check again after installation attempt
    if exist "C:\Program Files\Go\bin\go.exe" (
        set "GO_INSTALLED=1"
        set "PATH=%PATH%;C:\Program Files\Go\bin"
    )
)

REM Final check
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo ERROR: Go not found. Please install Go from https://go.dev/dl/
    pause
    exit /b 1
)

go version
echo Go is installed!

REM Check winget (for future use)
where winget >nul 2>nul
if %errorlevel% neq 0 (
    echo WARNING: winget not available. Some automatic installations may fail.
)

echo All prerequisites are installed!

REM ========== Python Bot Installation ==========
echo.
echo [1/5] Installing Python Bot...
echo.

cd /d "%~dp0..\darci-python"

REM Sync dependencies
echo Installing Python dependencies...
uv sync
if %errorlevel% neq 0 (
    echo Failed to install Python dependencies.
    pause
    exit /b 1
)

REM ========== Go Bot Installation ==========
echo.
echo [2/5] Building Go Bot...
echo.

cd /d "%~dp0..\darci-go"

REM Build Go bot
echo Building darci-go...
go build -o darci-go.exe ./cmd/darci-go
if %errorlevel% neq 0 (
    echo Failed to build Go bot.
    pause
    exit /b 1
)

echo Go bot built successfully!

REM ========== Configuration ==========
echo.
echo [3/5] Configuring...
echo.

REM Create config directory
set "CONFIG_DIR=%APPDATA%\darci"
if not exist "%CONFIG_DIR%" mkdir "%CONFIG_DIR%"

set "CONFIG_FILE=%CONFIG_DIR%\config.json"

REM Create or update config with API key
if "%API_KEY%"=="" (
    echo No API key provided. You can add it later to %CONFIG_FILE%
) else (
    echo Setting API key...
    if exist "%CONFIG_FILE%" (
        powershell -Command "$config = Get-Content '%CONFIG_FILE%' -Raw; $config = $config -replace '\"apiKey\": \"\"', '\"apiKey\": '%API_KEY%''; Set-Content '%CONFIG_FILE%' $config -NoNewline"
    ) else (
        powershell -Command "$json = @{agents=@{defaults=@{model='gemini-2.5-flash';provider='gemini'}}; providers=@{gemini=@{apiKey='%API_KEY%'}}; gateway=@{port=18790;heartbeat=@{enabled=$true;intervalS=1800}}; tools=@{web=@{search=@{apiKey=''}};restrictToWorkspace=$false}}; $json | ConvertTo-Json -Depth 10 | Set-Content '%CONFIG_FILE%' -NoNewline"
    )
)

REM Create workspace
echo Creating workspace...
set "WORKSPACE=%USERPROFILE%\darci-workspace"
if not exist "%WORKSPACE%" mkdir "%WORKSPACE%"

REM Copy bootstrap files
xcopy /E /I /Y "%~dp0..\darci-python\docs\*.md" "%WORKSPACE%\" 2>nul

REM ========== Create Start Menu Shortcuts ==========
echo.
echo [4/5] Creating shortcuts...
echo.

set "SHORTCUT_DIR=%APPDATA%\Microsoft\Windows\Start Menu\Programs\darci"
if not exist "%SHORTCUT_DIR%" mkdir "%SHORTCUT_DIR%"

REM Create batch files for easy launching
echo @echo off.
echo cd /d "%%~dp0..\..\..\repos\sentinelai\darci\darci-python".
echo uv run darci agent.
echo pause. > "%SHORTCUT_DIR%\darci-python.bat"

echo @echo off.
echo cd /d "%%~dp0..\..\..\repos\sentinelai\darci\darci-go".
echo darci-go.exe.
echo pause. > "%SHORTCUT_DIR%\darci-go.bat"

echo @echo off.
echo cd /d "%%~dp0".
echo uninstall.bat.
echo pause. > "%SHORTCUT_DIR%\Uninstall darci.bat"

REM ========== Add to PATH (optional) ==========
echo.
echo [5/5] Adding to PATH...
echo.

REM Create launcher scripts in user's bin directory
set "USER_BIN=%USERPROFILE%\darci-bin"
if not exist "%USER_BIN%" mkdir "%USER_BIN%"

REM Get absolute paths
set "DARCI_DIR=%~dp0.."
set "DARCI_PYTHON_DIR=%DARCI_DIR%\darci-python"
set "DARCI_GO_DIR=%DARCI_DIR%\darci-go"

REM Create Python launcher that activates venv automatically
(
echo @echo off
echo REM darci-python launcher - Windows
echo cd /d "%DARCI_PYTHON_DIR%"
echo call "%DARCI_PYTHON_DIR%\.venv\Scripts\activate.bat"
echo python -m darci agent %%*
) > "%USER_BIN%\darci-python.bat"

REM Create Go launcher
(
echo @echo off
echo REM darci-go launcher - Windows
echo cd /d "%DARCI_GO_DIR%"
echo darci-go.exe %%*
) > "%USER_BIN%\darci-go.bat"

REM Create uninstall launcher
(
echo @echo off
echo cd /d "%~dp0"
echo uninstall.bat %%*
) > "%USER_BIN%\uninstall-darci.bat"

echo Added launcher scripts to %USER_BIN%

echo.
echo =====================================
echo ✓ Installation complete!
echo =====================================
echo.
echo To run the bots:
echo   From anywhere: darci-python or darci-go (after adding to PATH)
echo   From Start Menu: darci-python or darci-go
echo   Manual: cd darci-python ^&^& uv run darci agent -m "Hello!"
echo           cd darci-go ^&^& darci-go.exe
echo.
echo To add to PATH permanently, add this folder:
echo   %USER_BIN%
echo.
echo To uninstall:
echo   Run: uninstall-darci
echo.
if "%API_KEY%"=="" (
    echo Remember to add your Gemini API key to:
    echo   %CONFIG_FILE%
    echo   Get one at: https://aistudio.google.com/apikey
)
echo.

endlocal
