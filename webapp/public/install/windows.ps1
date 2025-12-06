$ErrorActionPreference = "Stop"

Write-Host "Installing yapi for Windows..."

# Detect architecture
$arch = $env:PROCESSOR_ARCHITECTURE
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
$installDir = "$env:LOCALAPPDATA\Microsoft\WindowsApps"
Move-Item -Path (Join-Path $tmpDir "yapi.exe") -Destination $installDir -Force

# Cleanup
Remove-Item -Recurse -Force $tmpDir

Write-Host "yapi installed successfully!"
yapi --version
