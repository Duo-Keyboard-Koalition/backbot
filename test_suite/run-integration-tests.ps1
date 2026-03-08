# SentinelAI Integration Test Runner (PowerShell)
# Runs all integration tests with real APIs (NO MOCKS)

param(
    [ValidateSet("all", "gemini", "tailscale", "darci", "e2e")]
    [string]$Mode = "all",
    
    [switch]$Quiet,
    
    [switch]$Verbose,
    
    [string]$Parallel,
    
    [switch]$Coverage,
    
    [string[]]$Skip,
    
    [string]$Keyword,
    
    [switch]$Help
)

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$TestDir = Join-Path $ScriptDir "integration"

# Colors
$Blue = [ConsoleColor]::Blue
$Green = [ConsoleColor]::Green
$Yellow = [ConsoleColor]::Yellow
$Red = [ConsoleColor]::Red

function Log-Info { Write-Host "ℹ $args" -ForegroundColor $Blue }
function Log-Success { Write-Host "✓ $args" -ForegroundColor $Green }
function Log-Warning { Write-Host "⚠ $args" -ForegroundColor $Yellow }
function Log-Error { Write-Host "✗ $args" -ForegroundColor $Red }

function Show-Help {
    Write-Host @"

SentinelAI Integration Test Runner

Usage: $(Split-Path $MyInvocation.ScriptName -Leaf) [OPTIONS]

Options:
    -Mode <all|gemini|tailscale|darci|e2e>  Test mode (default: all)
    -Quiet                                   Quiet mode (minimal output)
    -Verbose                                 Extra verbose output
    -Parallel <N>                            Run tests in parallel with N workers
    -Coverage                                Generate coverage report
    -Skip <marker>                           Skip tests with marker (can be used multiple times)
    -Keyword <expr>                          Run tests matching keyword expression
    -Help                                    Show this help message

Examples:
    $(Split-Path $MyInvocation.ScriptName -Leaf)                        # Run all tests
    $(Split-Path $MyInvocation.ScriptName -Leaf) -Mode gemini           # Run only Gemini tests
    $(Split-Path $MyInvocation.ScriptName -Leaf) -Mode tailscale -Parallel auto
    $(Split-Path $MyInvocation.ScriptName -Leaf) -Skip slow -Skip api_cost
    $(Split-Path $MyInvocation.ScriptName -Leaf) -Keyword "test_agent"

Test Markers:
    gemini      Tests requiring Gemini API
    tailscale   Tests requiring Tailscale connection
    e2e         End-to-end workflow tests
    slow        Slow-running tests (>30s)
    api_cost    Tests consuming API quota

"@
    exit 0
}

function Check-Prerequisites {
    Log-Info "Checking prerequisites..."
    
    # Check Python
    if (-not (Get-Command python -ErrorAction SilentlyContinue)) {
        Log-Error "Python not found. Please install Python 3.10+"
        exit 1
    }
    
    # Check pytest
    $pytestVersion = python -m pytest --version 2>&1
    if ($LASTEXITCODE -ne 0) {
        Log-Error "pytest not found. Install with: pip install pytest pytest-asyncio pytest-cov pytest-timeout"
        exit 1
    }
    Log-Success "pytest found: $pytestVersion"
    
    # Check .env.test
    $envTestPath = Join-Path $ProjectRoot ".env.test"
    if (-not (Test-Path $envTestPath)) {
        Log-Warning ".env.test not found. Copy .env.test.example and configure API keys."
        Log-Info "Creating from example..."
        $envExamplePath = Join-Path $ProjectRoot ".env.test.example"
        if (Test-Path $envExamplePath) {
            Copy-Item $envExamplePath $envTestPath
        }
    }
    
    # Check Tailscale
    if ($Mode -eq "tailscale" -or $Mode -eq "all") {
        if (Get-Command tailscale -ErrorAction SilentlyContinue) {
            $tsStatus = tailscale status 2>&1
            if ($LASTEXITCODE -ne 0) {
                Log-Warning "Tailscale not connected. Tailscale tests may be skipped."
            } else {
                Log-Success "Tailscale connected"
            }
        } else {
            Log-Warning "tailscale CLI not found. Tailscale tests may be skipped."
        }
    }
    
    Log-Success "Prerequisites check complete"
}

function Run-Tests {
    $pytestArgs = @()
    
    if ($Quiet) {
        $pytestArgs += "-q"
    } elseif ($Verbose) {
        $pytestArgs += "-vv"
    } else {
        $pytestArgs += "-v"
    }
    
    # Add test directory
    $pytestArgs += $TestDir
    
    # Add mode-specific markers
    switch ($Mode) {
        "gemini" {
            $pytestArgs += "-m", "gemini"
            Log-Info "Running Gemini API tests..."
        }
        "tailscale" {
            $pytestArgs += "-m", "tailscale"
            Log-Info "Running Tailscale tests..."
        }
        "darci" {
            $pytestArgs += "-k", "darci"
            Log-Info "Running DarCI tests..."
        }
        "e2e" {
            $pytestArgs += "-m", "e2e"
            Log-Info "Running E2E tests..."
        }
        "all" {
            Log-Info "Running all integration tests..."
        }
    }
    
    # Add parallel execution
    if ($Parallel) {
        $pytestArgs += "-n", $Parallel
        Log-Info "Parallel execution enabled ($Parallel workers)"
    }
    
    # Add coverage
    if ($Coverage) {
        $pytestArgs += "--cov=backend", "--cov=darci", "--cov-report=html", "--cov-report=term"
        Log-Info "Coverage report enabled"
    }
    
    # Add skip markers
    foreach ($marker in $Skip) {
        $pytestArgs += "-m", "not $marker"
    }
    
    # Add keyword filter
    if ($Keyword) {
        $pytestArgs += "-k", $Keyword
    }
    
    # Run tests
    Log-Info "Executing: pytest $($pytestArgs -join ' ')"
    Write-Host ""
    
    python -m pytest @pytestArgs
    $exitCode = $LASTEXITCODE
    
    if ($exitCode -eq 0) {
        Log-Success "All tests passed!"
    } else {
        Log-Error "Some tests failed (exit code: $exitCode)"
        exit $exitCode
    }
}

# Main execution
if ($Help) {
    Show-Help
}

Write-Host ""
Write-Host "========================================" -ForegroundColor $Green
Write-Host "  SentinelAI Integration Test Runner   " -ForegroundColor $Green
Write-Host "========================================" -ForegroundColor $Green
Write-Host ""

Check-Prerequisites
Write-Host ""
Run-Tests

Write-Host ""
Log-Success "Test run complete!"
