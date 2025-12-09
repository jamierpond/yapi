$ErrorActionPreference = "Stop"

Write-Host "Installing yapi for Windows..."

# Detect architecture
$arch = $env:PROCESSOR_ARCHITECTURE
if (-not $arch) {
    # Fallback for non-Windows systems (e.g., PowerShell on Linux)
    $uname = uname -m 2>$null
    if ($uname -eq "x86_64") {
        $arch = "AMD64"
    } elseif ($uname -eq "aarch64" -or $uname -eq "arm64") {
        $arch = "ARM64"
    }
}

if ($arch -eq "ARM64") {
    $asset = "yapi_windows_arm64.zip"
} elseif ($arch -eq "AMD64") {
    $asset = "yapi_windows_amd64.zip"
} else {
    Write-Error "Unsupported architecture: $arch"
    exit 1
}

# Download
$tmpDir = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }
$zipPath = Join-Path $tmpDir "yapi.zip"
Invoke-WebRequest -Uri "https://github.com/jamierpond/yapi/releases/latest/download/$asset" -OutFile $zipPath

# Extract
Expand-Archive -Path $zipPath -DestinationPath $tmpDir

# Install
if ($env:LOCALAPPDATA) {
    $installDir = "$env:LOCALAPPDATA\Microsoft\WindowsApps"
} else {
    # Fallback for non-Windows systems
    $installDir = "/usr/local/bin"
}
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}
$exeName = if ($env:OS -eq "Windows_NT") { "yapi.exe" } else { "yapi" }
Move-Item -Path (Join-Path $tmpDir $exeName) -Destination $installDir -Force

# Cleanup
Remove-Item -Recurse -Force $tmpDir

Write-Host "yapi installed successfully!"
yapi version
