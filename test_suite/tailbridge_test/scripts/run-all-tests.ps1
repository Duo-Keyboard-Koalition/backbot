# Tailbridge Test Platform - PowerShell Test Runner
# Run all tests for the Tailbridge test platform

$ErrorActionPreference = "Stop"
$ScriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$TestPlatformRoot = Split-Path -Parent $ScriptRoot

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Tailbridge Test Platform" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Track test results
$MockTestsPassed = $true
$IntegrationTestsPassed = $true

# Function to run mock tests
function Invoke-MockTests {
    Write-Host "Running Mock Tests..." -ForegroundColor Yellow
    Write-Host "----------------------" -ForegroundColor Yellow
    
    try {
        Set-Location $TestPlatformRoot
        $result = go test ./mock/... -v -coverprofile=coverage.out 2>&1
        Write-Host $result
        
        if ($LASTEXITCODE -ne 0) {
            $MockTestsPassed = $false
            Write-Host "Mock tests FAILED" -ForegroundColor Red
        } else {
            Write-Host "Mock tests PASSED" -ForegroundColor Green
            
            # Show coverage
            go tool cover -func=coverage.out | Select-Object -Last 1
        }
    } catch {
        $MockTestsPassed = $false
        Write-Host "Mock tests ERROR: $_" -ForegroundColor Red
    }
    
    Write-Host ""
}

# Function to run Docker integration tests
function Invoke-IntegrationTests {
    Write-Host "Running Integration Tests..." -ForegroundColor Yellow
    Write-Host "-----------------------------" -ForegroundColor Yellow
    
    $DockerComposeFile = Join-Path $TestPlatformRoot "docker\docker-compose.test.yml"
    
    # Check if .env file exists
    $EnvFile = Join-Path $TestPlatformRoot "docker\.env"
    if (Test-Path $EnvFile) {
        Write-Host "Loading environment from .env file" -ForegroundColor Green
    } else {
        Write-Host "WARNING: No .env file found. Integration tests may fail without TS_AUTHKEY values." -ForegroundColor Yellow
        Write-Host "Create docker\.env with TS_AUTH_KEY_1, TS_AUTH_KEY_2, TS_AUTH_KEY_3" -ForegroundColor Yellow
        Write-Host ""
    }
    
    try {
        # Start Docker containers
        Write-Host "Starting Docker containers..." -ForegroundColor Yellow
        docker-compose -f $DockerComposeFile up -d
        
        if ($LASTEXITCODE -ne 0) {
            throw "Failed to start Docker containers"
        }
        
        Write-Host "Waiting for agents to be ready..." -ForegroundColor Yellow
        Start-Sleep -Seconds 30
        
        # Run integration tests
        Set-Location $TestPlatformRoot
        $result = go test ./integration/... -v -tags=integration -timeout=10m 2>&1
        Write-Host $result
        
        if ($LASTEXITCODE -ne 0) {
            $IntegrationTestsPassed = $false
            Write-Host "Integration tests FAILED" -ForegroundColor Red
        } else {
            Write-Host "Integration tests PASSED" -ForegroundColor Green
        }
    } catch {
        $IntegrationTestsPassed = $false
        Write-Host "Integration tests ERROR: $_" -ForegroundColor Red
    } finally {
        # Cleanup
        Write-Host ""
        Write-Host "Cleaning up Docker containers..." -ForegroundColor Yellow
        docker-compose -f $DockerComposeFile down
    }
    
    Write-Host ""
}

# Function to show test summary
function Show-Summary {
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "Test Summary" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""
    
    if ($MockTestsPassed) {
        Write-Host "[PASS] Mock Tests" -ForegroundColor Green
    } else {
        Write-Host "[FAIL] Mock Tests" -ForegroundColor Red
    }
    
    if ($IntegrationTestsPassed) {
        Write-Host "[PASS] Integration Tests" -ForegroundColor Green
    } else {
        Write-Host "[FAIL] Integration Tests" -ForegroundColor Red
    }
    
    Write-Host ""
    
    if ($MockTestsPassed -and $IntegrationTestsPassed) {
        Write-Host "All tests PASSED!" -ForegroundColor Green
        exit 0
    } else {
        Write-Host "Some tests FAILED!" -ForegroundColor Red
        exit 1
    }
}

# Parse command line arguments
$RunMockTests = $true
$RunIntegrationTests = $false
$RunAllTests = $false

if ($args -contains "-mock") {
    $RunMockTests = $true
    $RunIntegrationTests = $false
}

if ($args -contains "-integration") {
    $RunMockTests = $false
    $RunIntegrationTests = $true
}

if ($args -contains "-all") {
    $RunAllTests = $true
    $RunMockTests = $true
    $RunIntegrationTests = $true
}

if ($args -contains "-help" -or $args -contains "-h") {
    Write-Host "Usage: .\run-all-tests.ps1 [options]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -mock         Run only mock tests (default)"
    Write-Host "  -integration  Run only integration tests"
    Write-Host "  -all          Run all tests"
    Write-Host "  -help, -h     Show this help"
    Write-Host ""
    exit 0
}

# Run tests
if ($RunMockTests) {
    Invoke-MockTests
}

if ($RunIntegrationTests -or $RunAllTests) {
    Invoke-IntegrationTests
}

# Show summary
Show-Summary
