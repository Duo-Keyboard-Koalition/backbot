# 🐈 scorpion - Installation Scripts

Unified installation and uninstallation scripts for both Python and Go scorpion bots.

## Quick Start

### Windows

```powershell
# Navigate to scripts folder
cd C:\Users\darcy\repos\sentinelai\scorpion\scripts

# Install with API key
.\install.bat AIzaSyCSvcSZsC8Bg1k343y9l3as3vlOrhsXRSw

# Install without API key (configure later)
.\install.bat

# Uninstall (with confirmation)
.\uninstall.bat

# Uninstall (skip confirmation)
.\uninstall.bat -y
```

### Linux/macOS

```bash
# Navigate to scripts folder
cd /path/to/scorpion/scripts

# Install with API key
./install.sh AIzaSyCSvcSZsC8Bg1k343y9l3as3vlOrhsXRSw

# Install without API key (configure later)
./install.sh

# Uninstall (with confirmation)
./uninstall.sh

# Uninstall (skip confirmation)
./uninstall.sh -y
```

## What Gets Installed

### Prerequisites (auto-installed if missing)
- **Python 3.11+** (via winget/apt/brew/dnf/pacman)
- **uv** package manager (auto-installed)
- **Go 1.22+** (via winget/apt/brew/dnf/pacman)

### Python Bot (scorpion-python)
- Virtual environment with all dependencies
- Configuration at:
  - Windows: `%APPDATA%\scorpion\config.json`
  - Linux/macOS: `~/.config/scorpion/config.json`

### Go Bot (scorpion-go)
- Compiled binary `scorpion-go.exe` (Windows) or `scorpion-go` (Linux/macOS)
- Self-contained with no external dependencies
- Commands: `/time`, `/ls <path>`, `/cat <path>`

### Additional
- Workspace directory with bootstrap files
- Start Menu shortcuts (Windows) or launcher scripts (Linux/macOS)

## What Gets Removed on Uninstall

- Python virtual environment
- Go binary
- Configuration files
- Workspace directory
- Start Menu/Desktop shortcuts
- Launcher scripts

## Usage

### Python Bot
```bash
# Interactive mode (from anywhere)
scorpion-python

# Single message
scorpion-python -m "Hello!"

# With logs
scorpion-python --logs -m "Hello!"

# Manual
cd scorpion-python && uv run scorpion agent
```

### Go Bot
```bash
# Interactive mode (from anywhere)
scorpion-go

# Commands:
# /time          - Show current UTC time
# /ls <path>     - List directory contents
# /cat <path>    - Read file contents
# exit/quit      - Exit the bot
```

## Requirements

### Python Bot
- Python 3.11+ (auto-installed if missing)
- uv package manager (auto-installed)
- Gemini API key (get from https://aistudio.google.com/apikey)

### Go Bot
- Go 1.22+ (auto-installed if missing)
- No API key required (uses local rule-based model)

## Troubleshooting

### Go Installation Fails (Windows)
If winget fails to install Go, download manually from https://go.dev/dl/

### Python Dependencies Fail
Delete the `.venv` folder and run `uv sync` again

### API Key Issues
Edit the config file and ensure `providers.gemini.apiKey` is set correctly

### Prerequisites Not Found
The installer will attempt to auto-install missing prerequisites. If automatic installation fails, manual installation instructions are provided.
