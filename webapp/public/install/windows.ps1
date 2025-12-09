$ErrorActionPreference = "Stop"

# Detect arch -> asset
$arch = $env:PROCESSOR_ARCHITECTURE.ToUpper()
$asset = switch ($arch) {
    "AMD64" { "yapi_windows_amd64.zip" }
    "ARM64" { "yapi_windows_arm64.zip" }
    default { throw "Unsupported arch: $arch" }
}

# Temp dir
$tmp = Join-Path $env:TEMP ([IO.Path]::GetRandomFileName())
New-Item -ItemType Directory $tmp | Out-Null

# Download + extract
$zip = Join-Path $tmp "yapi.zip"
Invoke-WebRequest "https://github.com/jamierpond/yapi/releases/latest/download/$asset" -OutFile $zip
Expand-Archive $zip $tmp -Force

# Install dir
$install = "$env:LOCALAPPDATA\yapi"
New-Item -ItemType Directory -Force $install | Out-Null

# Move binary
Move-Item (Join-Path $tmp "yapi.exe") $install -Force

# Ensure PATH
if (-not ($env:PATH -split ";" | Where-Object { $_ -eq $install })) {
    $userPath = [Environment]::GetEnvironmentVariable("PATH","User")
    [Environment]::SetEnvironmentVariable("PATH", "$userPath;$install", "User")
    $env:PATH += ";$install"
}

# Cleanup
Remove-Item -Recurse -Force $tmp

yapi version

