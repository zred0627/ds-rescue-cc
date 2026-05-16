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
Write-Host "Next steps:"
Write-Host "  1. Set API key: `$env:DEEPSEEK_API_KEY = 'sk-...'"
Write-Host "  2. Self-check: ds-rescue --check"
