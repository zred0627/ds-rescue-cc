# ds-rescue installer for Windows
$ErrorActionPreference = "Stop"

$repo = "zred0627/ds-rescue-cc"
$prefix = if ($env:PREFIX) { $env:PREFIX } else { "$env:USERPROFILE\bin" }

if (-not (Test-Path $prefix)) {
    New-Item -ItemType Directory -Path $prefix -Force | Out-Null
}

$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { throw "32-bit Windows not supported" }

$apiUrl = "https://api.github.com/repos/$repo/releases/latest"
$release = Invoke-RestMethod -Uri $apiUrl
$tag = $release.tag_name
if (-not $tag) { throw "Failed to resolve latest tag" }

$archive = "ds-rescue_$($tag.TrimStart('v'))_windows_${arch}.zip"
$url = "https://github.com/$repo/releases/download/$tag/$archive"

$tmp = Join-Path $env:TEMP "ds-rescue-install-$(Get-Random)"
New-Item -ItemType Directory -Path $tmp -Force | Out-Null

Write-Host "Downloading $url ..."
Invoke-WebRequest -Uri $url -OutFile "$tmp\$archive"
Expand-Archive -Path "$tmp\$archive" -DestinationPath $tmp

Move-Item -Path "$tmp\ds-rescue.exe" -Destination "$prefix\ds-rescue.exe" -Force
Remove-Item -Recurse -Force $tmp

Write-Host "Installed ds-rescue.exe to $prefix"
Write-Host "Add $prefix to PATH if not already present"
Write-Host ""

# --- Smart App Control self-check ---
$sacState = $null
try {
    $mpStatus = Get-MpComputerStatus -ErrorAction SilentlyContinue
    if ($mpStatus) {
        $sacState = $mpStatus.SmartAppControlState
    }
} catch {
    # Get-MpComputerStatus may not be available on all editions; treat as unknown
    $sacState = $null
}

if ($sacState -eq "On") {
    Write-Host ""
    Write-Host "WARNING: Windows Smart App Control (SAC) is ON on this machine." -ForegroundColor Yellow
    Write-Host "SAC may block the downloaded binary because it carries a Mark-of-the-Web" -ForegroundColor Yellow
    Write-Host "(MOTW) zone flag from the GitHub Releases download." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Three options (in order of preference):" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "  1. RECOMMENDED -- compile locally (no MOTW, SAC trusts by default):" -ForegroundColor Green
    Write-Host "     go install github.com/zred0627/ds-rescue-cc/cmd/ds-rescue@latest"
    Write-Host ""
    Write-Host "  2. Unblock the downloaded file (removes MOTW flag):" -ForegroundColor Yellow
    Write-Host "     Unblock-File $prefix\ds-rescue.exe"
    Write-Host ""
    Write-Host "  3. Add an exclusion (requires Administrator):" -ForegroundColor Yellow
    Write-Host "     Add-MpPreference -ExclusionPath `"$prefix\ds-rescue.exe`""
    Write-Host ""
} elseif ($sacState -eq "Off" -or $sacState -eq "Eval") {
    Write-Host "Smart App Control state: $sacState -- binary should run without issues." -ForegroundColor Green
} else {
    Write-Host "Smart App Control state: could not determine (run as Administrator for full status)." -ForegroundColor Gray
    Write-Host "If you see a blocked-app dialog, run:" -ForegroundColor Gray
    Write-Host "  go install github.com/zred0627/ds-rescue-cc/cmd/ds-rescue@latest" -ForegroundColor Gray
}

Write-Host ""
Write-Host "Next steps:"
Write-Host "  1. Set API key: `$env:DEEPSEEK_API_KEY = 'sk-...'"
Write-Host "  2. Self-check: ds-rescue --check"
Write-Host ""

# --- Verify binary runs ---
$binaryPath = Join-Path $prefix "ds-rescue.exe"
$versionOutput = & $binaryPath --version 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "Binary verified: $versionOutput" -ForegroundColor Green
} else {
    Write-Host "Binary verification failed (exit code $LASTEXITCODE)." -ForegroundColor Red
    Write-Host "If SAC is blocking it, use option 1 above: go install github.com/zred0627/ds-rescue-cc/cmd/ds-rescue@latest" -ForegroundColor Yellow
}
