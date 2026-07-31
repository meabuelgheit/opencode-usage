#Requires -Version 5.1
<#
.SYNOPSIS
    Installs opencode-usage on Windows.

.DESCRIPTION
    Downloads the latest Windows release binary from GitHub Releases and
    installs it to a user-local directory, adding it to the user's PATH.

    Usage:
        powershell -ExecutionPolicy Bypass -c "irm https://raw.githubusercontent.com/abuelgheit/opencode-usage/main/install.ps1 | iex"
#>

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$Repo = 'abuelgheit/opencode-usage'
$Bin  = 'opencode-usage.exe'

# Install under %LOCALAPPDATA% (fall back to %USERPROFILE%\AppData\Local)
$BaseDir = if ($env:LOCALAPPDATA) { $env:LOCALAPPDATA } else { Join-Path $env:USERPROFILE 'AppData\Local' }
$InstallDir = Join-Path $BaseDir 'Programs\opencode-usage'

# Detect native architecture (handles 32-bit PowerShell on 64-bit OS)
$NativeArch = $env:PROCESSOR_ARCHITEW6432
if (-not $NativeArch) { $NativeArch = $env:PROCESSOR_ARCHITECTURE }
switch ($NativeArch) {
    'AMD64' { $Arch = 'amd64' }
    'ARM64' { $Arch = 'arm64' }
    default { throw "Unsupported architecture: $NativeArch" }
}

$Latest = (Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -Headers @{ 'User-Agent' = 'opencode-usage-installer' }).tag_name
if (-not $Latest) {
    throw 'Failed to fetch latest release. Check https://github.com/abuelgheit/opencode-usage/releases'
}

Write-Host "Installing opencode-usage $Latest for windows/$Arch..." -ForegroundColor Cyan

$ZipName = "opencode-usage_windows_$Arch.zip"
$BaseUrl = "https://github.com/$Repo/releases/download/$Latest"
$TmpDir = Join-Path $env:TEMP ("opencode-usage-" + [System.Guid]::NewGuid().ToString('N'))
$ZipPath = Join-Path $TmpDir $ZipName
$ChecksumsPath = Join-Path $TmpDir 'checksums.txt'
$ExtractDir = Join-Path $TmpDir 'extract'

try {
    New-Item -ItemType Directory -Path $TmpDir | Out-Null

    Invoke-WebRequest -Uri "$BaseUrl/$ZipName" -OutFile $ZipPath -UseBasicParsing
    Invoke-WebRequest -Uri "$BaseUrl/checksums.txt" -OutFile $ChecksumsPath -UseBasicParsing

    # Reject if the download does not match the published SHA-256 checksum
    $Expected = Get-Content $ChecksumsPath | Where-Object { $_ -match "\s+$([regex]::Escape($ZipName))$" } | ForEach-Object { ($_ -split '\s+')[0] }
    if (-not $Expected) {
        throw "No checksum found for $ZipName"
    }
    $Actual = (Get-FileHash -Path $ZipPath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($Actual -ne $Expected.ToLowerInvariant()) {
        throw "Checksum verification failed for $ZipName (expected $Expected, got $Actual)"
    }

    Expand-Archive -Path $ZipPath -DestinationPath $ExtractDir

    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    Copy-Item -Path (Join-Path $ExtractDir $Bin) -Destination (Join-Path $InstallDir $Bin) -Force

    Write-Host "Installed to $InstallDir\$Bin" -ForegroundColor Green
}
finally {
    Remove-Item -Recurse -Force -Path $TmpDir -ErrorAction SilentlyContinue
}

# Add to user PATH if not already present
$UserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
$AlreadyOnPath = if ($UserPath) { ($UserPath -split ';') -contains $InstallDir } else { $false }
if (-not $AlreadyOnPath) {
    $NewPath = if ($UserPath) { "$UserPath;$InstallDir" } else { $InstallDir }
    [Environment]::SetEnvironmentVariable('Path', $NewPath, 'User')
    Write-Host "Added $InstallDir to your user PATH. Open a new terminal to use it." -ForegroundColor Yellow
}
else {
    Write-Host "$InstallDir is already on your PATH." -ForegroundColor Yellow
}
