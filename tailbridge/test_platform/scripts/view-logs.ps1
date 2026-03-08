# View Test Logs - PowerShell Script
# Usage: .\view-logs.ps1 [suite_name]

param(
    [string]$SuiteName = "",
    [switch]$Report,
    [switch]$Help
)

if ($Help) {
    Write-Host "Usage: .\view-logs.ps1 [options]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  <suite_name>  View logs for specific suite (e.g., NetworkDiscovery)"
    Write-Host "  -Report       View test report instead of individual logs"
    Write-Host "  -Help         Show this help"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\view-logs.ps1 NetworkDiscovery"
    Write-Host "  .\view-logs.ps1 -Report"
    Write-Host ""
    exit 0
}

$LogsDir = Join-Path $PSScriptRoot "test_logs"

if (-not (Test-Path $LogsDir)) {
    Write-Host "No test logs found. Run tests first." -ForegroundColor Red
    exit 1
}

function Show-TestLog {
    param($LogFile)
    
    $json = Get-Content $LogFile | ConvertFrom-Json
    
    Write-Host "`n========================================" -ForegroundColor Cyan
    Write-Host "Test: $($json.test_name)" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "Test ID:   $($json.test_id)"
    Write-Host "Duration:  $($json.duration)"
    Write-Host "Start:     $($json.start_time)"
    Write-Host "End:       $($json.end_time)"
    Write-Host ""
    
    Write-Host "Log Entries:" -ForegroundColor Yellow
    Write-Host "------------" -ForegroundColor Yellow
    
    foreach ($entry in $json.log_entries) {
        $color = switch ($entry.level) {
            "INFO"  { "Green" }
            "WARN"  { "Yellow" }
            "ERROR" { "Red" }
            default { "White" }
        }
        
        Write-Host "[$($entry.timestamp)] " -NoNewline
        Write-Host "[$($entry.level)] " -ForegroundColor $color -NoNewline
        Write-Host "[$($entry.category)] " -ForegroundColor Blue -NoNewline
        Write-Host "$($entry.message)"
        
        if ($entry.details) {
            $entry.details | ConvertTo-Json -Depth 3 | ForEach-Object {
                Write-Host "  $_"
            }
        }
    }
}

function Show-TestReport {
    param($ReportFile)
    
    $json = Get-Content $ReportFile | ConvertFrom-Json
    
    Write-Host "`n========================================" -ForegroundColor Cyan
    Write-Host "Test Report: $($json.suite_name)" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "Generated:    $($json.generated_at)"
    Write-Host "Total Tests:  $($json.total_tests)"
    Write-Host "Passed:       $($json.passed)" -ForegroundColor Green
    Write-Host "Failed:       $($json.failed)" -ForegroundColor $(if ($json.failed -gt 0) { "Red" } else { "Green" })
    Write-Host "Duration:     $($json.total_duration)"
    Write-Host ""
    
    Write-Host "Test Details:" -ForegroundColor Yellow
    Write-Host "-------------" -ForegroundColor Yellow
    
    foreach ($test in $json.tests) {
        $status = if ($test.passed) { "PASS" } else { "FAIL" }
        $color = if ($test.passed) { "Green" } else { "Red" }
        
        Write-Host "[$status] " -ForegroundColor $color -NoNewline
        Write-Host "$($test.test_id) - $($test.duration)"
    }
}

function List-AvailableSuites {
    Write-Host "`nAvailable Test Suites:" -ForegroundColor Cyan
    Write-Host "----------------------" -ForegroundColor Cyan
    
    Get-ChildItem $LogsDir -Directory | ForEach-Object {
        $suite = $_.Name
        $logCount = (Get-ChildItem $_.FullName -Filter "*.json" -Exclude "test_report.json").Count
        $hasReport = Test-Path (Join-Path $_.FullName "test_report.json")
        
        Write-Host "  $suite" -NoNewline
        Write-Host " ($logCount logs" -ForegroundColor Gray -NoNewline
        if ($hasReport) {
            Write-Host ", has report)" -ForegroundColor Gray
        } else {
            Write-Host ")" -ForegroundColor Gray
        }
    }
}

# Main logic
if ($SuiteName -eq "") {
    List-AvailableSuites
    exit 0
}

$SuiteDir = Join-Path $LogsDir $SuiteName

if (-not (Test-Path $SuiteDir)) {
    Write-Host "Suite '$SuiteName' not found." -ForegroundColor Red
    List-AvailableSuites
    exit 1
}

if ($Report) {
    $ReportFile = Join-Path $SuiteDir "test_report.json"
    if (Test-Path $ReportFile) {
        Show-TestReport $ReportFile
    } else {
        Write-Host "No report found for suite '$SuiteName'." -ForegroundColor Red
    }
} else {
    # Get latest log file
    $LatestLog = Get-ChildItem $SuiteDir -Filter "*.json" -Exclude "test_report.json" | 
        Sort-Object LastWriteTime -Descending | 
        Select-Object -First 1
    
    if ($LatestLog) {
        Show-TestLog $LatestLog.FullName
    } else {
        Write-Host "No logs found for suite '$SuiteName'." -ForegroundColor Red
    }
}
