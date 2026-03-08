# Multi-Agent Tailscale Test Runner (PowerShell)
# This script runs the multi-agent integration test and displays IP addresses and chat logs

param(
    [switch]$NoCleanup,
    [switch]$Verbose
)

$ErrorActionPreference = "Stop"

# Configuration
$env:TS_AUTH_KEY = $env:TS_AUTH_KEY ?? "tskey-auth-k7Q1t39ZWj11CNTRL-FbRR2tKLRcPn5L246vsAcP7LP2YCUxWD"
$env:GEMINI_API_KEY = $env:GEMINI_API_KEY ?? "AIzaSyC_9M8im8z5F0ING2W3Hu2aQiunJhhWUXI"
$env:AGENT1_URL = $env:AGENT1_URL ?? "http://localhost:8081"
$env:AGENT2_URL = $env:AGENT2_URL ?? "http://localhost:8082"
$env:AGENT3_URL = $env:AGENT3_URL ?? "http://localhost:8083"

Write-Host "╔═══════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║     MULTI-AGENT TAILSCALE COMMUNICATION TEST RUNNER      ║" -ForegroundColor Cyan
Write-Host "╚═══════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# Check if Docker is running
try {
    docker ps | Out-Null
    Write-Host "✓" -ForegroundColor Green -NoNewline
    Write-Host " Docker is running"
} catch {
    Write-Host "✗ Docker is not running. Please start Docker Desktop." -ForegroundColor Red
    exit 1
}

# Check if .env file exists
if (-not (Test-Path "docker\.env")) {
    Write-Host "!" -ForegroundColor Yellow -NoNewline
    Write-Host " Creating docker/.env file with auth keys..."
    @"
TS_AUTH_KEY_1=$($env:TS_AUTH_KEY)
TS_AUTH_KEY_2=$($env:TS_AUTH_KEY)
TS_AUTH_KEY_3=$($env:TS_AUTH_KEY)
TAILNET_NAME=test.ts.net
"@ | Out-File -FilePath "docker\.env" -Encoding utf8
    Write-Host "✓" -ForegroundColor Green -NoNewline
    Write-Host " Created docker/.env"
}

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Blue
Write-Host "Configuration:" -ForegroundColor Blue
Write-Host "  Tailscale Auth Key: $($env:TS_AUTH_KEY.Substring(0, [Math]::Min(30, $env:TS_AUTH_KEY.Length)))..."
Write-Host "  Agent 1 URL: $($env:AGENT1_URL)"
Write-Host "  Agent 2 URL: $($env:AGENT2_URL)"
Write-Host "  Agent 3 URL: $($env:AGENT3_URL)"
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Blue
Write-Host ""

# Start the test environment
Write-Host "📦" -NoNewline
Write-Host " Starting Docker test environment..." -ForegroundColor Yellow
docker-compose -f docker/docker-compose.test.yml up -d

Write-Host ""
Write-Host "⏳" -NoNewline
Write-Host " Waiting for agents to connect to Tailscale (this may take 2-3 minutes)..." -ForegroundColor Yellow
Write-Host ""

# Wait for agents
$MAX_ATTEMPTS = 30
$ATTEMPT = 0

while ($ATTEMPT -lt $MAX_ATTEMPTS) {
    $ATTEMPT++
    Write-Host "  Attempt $ATTEMPT/$MAX_ATTEMPTS..." -ForegroundColor Cyan -NoNewline
    Write-Host "`r" -NoNewline
    
    $ALL_HEALTHY = $true
    
    for ($i = 1; $i -le 3; $i++) {
        $PORT = 8080 + $i
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:$PORT/health" -TimeoutSec 2 -UseBasicParsing -ErrorAction SilentlyContinue
            if ($response.StatusCode -ne 200) {
                $ALL_HEALTHY = $false
                break
            }
        } catch {
            $ALL_HEALTHY = $false
            break
        }
    }
    
    if ($ALL_HEALTHY) {
        Write-Host "  ✓ All agents are healthy!          " -ForegroundColor Green
        break
    }
    
    Start-Sleep -Seconds 10
}

if (-not $ALL_HEALTHY) {
    Write-Host "✗ Timeout waiting for agents to become healthy" -ForegroundColor Red
    Write-Host "!" -ForegroundColor Yellow -NoNewline
    Write-Host " Showing agent logs:"
    docker-compose -f docker/docker-compose.test.yml logs --tail=20
    exit 1
}

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Blue
Write-Host "AGENT IP ADDRESSES:" -ForegroundColor Blue
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Blue

# Get agent information
for ($i = 1; $i -le 3; $i++) {
    $PORT = 8080 + $i
    Write-Host ""
    Write-Host "Agent $i:" -ForegroundColor Cyan
    
    # Get status from agent
    try {
        $STATUS = Invoke-RestMethod -Uri "http://localhost:$PORT/status" -TimeoutSec 5 -ErrorAction SilentlyContinue
        $NAME = $STATUS.name ?? "agent$i"
        $IP = $STATUS.ip ?? "unknown"
        $TS_IP = $STATUS.tailscale_ip ?? "connecting..."
        $ONLINE = $STATUS.online ?? $false
    } catch {
        $NAME = "agent$i"
        $IP = "unknown"
        $TS_IP = "unknown"
        $ONLINE = $false
    }
    
    Write-Host "  Hostname:      $NAME"
    Write-Host "  Local IP:      $IP"
    Write-Host "  Tailscale IP:  $TS_IP"
    Write-Host "  Status:        $ONLINE"
}

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Blue
Write-Host "RUNNING MULTI-AGENT COMMUNICATION TEST:" -ForegroundColor Blue
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Blue
Write-Host ""

# Run the integration test
$testLogPath = "test_logs\multiagent-test-$(Get-Date -Format 'yyyyMMdd-HHmmss').log"
if (-not (Test-Path "test_logs")) {
    New-Item -ItemType Directory -Path "test_logs" | Out-Null
}

$testOutput = go test ./integration/multiagent/... -v -tags=integration -timeout=10m 2>&1
$TEST_EXIT_CODE = $LASTEXITCODE

$testOutput | Tee-Object -FilePath $testLogPath

Write-Host ""
if ($TEST_EXIT_CODE -eq 0) {
    Write-Host "╔═══════════════════════════════════════════════════════════╗" -ForegroundColor Green
    Write-Host "║              TEST PASSED SUCCESSFULLY                     ║" -ForegroundColor Green
    Write-Host "╚═══════════════════════════════════════════════════════════╝" -ForegroundColor Green
} else {
    Write-Host "╔═══════════════════════════════════════════════════════════╗" -ForegroundColor Red
    Write-Host "║                    TEST FAILED                            ║" -ForegroundColor Red
    Write-Host "╚═══════════════════════════════════════════════════════════╝" -ForegroundColor Red
}

Write-Host ""
Write-Host "!" -ForegroundColor Yellow -NoNewline
Write-Host " Test logs saved to: $testLogPath"
Write-Host ""

# Optionally cleanup
if (-not $NoCleanup) {
    $response = Read-Host "Stop Docker containers? (y/n)"
    if ($response -match '^[Yy]$') {
        Write-Host "🧹" -NoNewline
        Write-Host " Stopping Docker containers..." -ForegroundColor Yellow
        docker-compose -f docker/docker-compose.test.yml down
        Write-Host "✓" -ForegroundColor Green -NoNewline
        Write-Host " Containers stopped"
    }
}

exit $TEST_EXIT_CODE
