$ErrorActionPreference = "Stop"

# Detect arch -> asset
$arch = $env:PROCESSOR_ARCHITECTURE.ToUpper()
$asset = switch ($arch) {
    "AMD64" { "yapi_windows_amd64.zip" }
    "ARM64" { "yapi_windows_arm64.zip" }
    default { throw "Unsupported arch: $arch" }
}

$baseUrl = "https://github.com/jamierpond/yapi/releases/latest/download"

# Temp dir
$tmp = Join-Path $env:TEMP ([IO.Path]::GetRandomFileName())
New-Item -ItemType Directory $tmp | Out-Null

# Download
$zip = Join-Path $tmp $asset
$checksums = Join-Path $tmp "checksums.txt"
Invoke-WebRequest "$baseUrl/$asset" -OutFile $zip
Invoke-WebRequest "$baseUrl/checksums.txt" -OutFile $checksums

# Verify checksum
Write-Host "Verifying checksum..."
$expected = (Get-Content $checksums | Select-String $asset).ToString().Split()[0]
$actual = (Get-FileHash $zip -Algorithm SHA256).Hash.ToLower()
if ($expected -ne $actual) {
    Write-Host "Checksum verification failed!"
    Write-Host "Expected: $expected"
    Write-Host "Actual:   $actual"
    Remove-Item -Recurse -Force $tmp
    exit 1
}
Write-Host "Checksum verified."

# Extract
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

