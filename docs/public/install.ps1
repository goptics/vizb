# vizb installer for Windows
# Usage: irm https://vizb.goptics.org/install.ps1 | iex

$ErrorActionPreference = "Stop"

function Show-Logo {
Write-Host @'
\    ####
 \ ++++        vizb installer
  \**   izb    vizb.goptics.org
   =
'@
}
Show-Logo

$InstallDir = "$env:LOCALAPPDATA\vizb"
$Repo = "goptics/vizb"
$Bin = "vizb.exe"

Write-Host " info: fetching latest release..." -ForegroundColor Cyan
$LatestUrl = "https://github.com/$Repo/releases/latest"
$Request = [System.Net.WebRequest]::Create($LatestUrl)
$Request.AllowAutoRedirect = $false
$Response = $Request.GetResponse()
$Redirect = $Response.GetResponseHeader("Location")
$Response.Close()
$VersionTag = $Redirect -replace '.*/', ''
$Version = $VersionTag -replace '^v', ''

if (-not $Version) {
    Write-Host "error: failed to determine latest version" -ForegroundColor Red
    exit 1
}
Write-Host " info: latest version: $Version"

$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "arm64" }
$Archive = "vizb@$Version-windows-$Arch.zip"
$Url = "https://github.com/$Repo/releases/download/$VersionTag/$Archive"

$TmpDir = Join-Path $env:TEMP "vizb-install-$([System.IO.Path]::GetRandomFileName())"
New-Item -ItemType Directory -Path $TmpDir | Out-Null

try {
    $ZipPath = Join-Path $TmpDir $Archive
    Write-Host " info: downloading $Url..." -ForegroundColor Cyan
    Invoke-WebRequest -Uri $Url -OutFile $ZipPath

    Write-Host " info: extracting..." -ForegroundColor Cyan
    Expand-Archive -Path $ZipPath -DestinationPath $TmpDir

    $ExePath = Join-Path $InstallDir $Bin
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null

    Write-Host " info: installing to $InstallDir..." -ForegroundColor Cyan
    Copy-Item (Join-Path $TmpDir $Bin) $ExePath -Force

    Write-Host " info: verifying installation..." -ForegroundColor Cyan
    $Output = & $ExePath --version 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host " info: vizb $($Output) installed successfully to $ExePath" -ForegroundColor Green
    } else {
        Write-Host "error: installation verification failed" -ForegroundColor Red
        exit 1
    }

    $CurrentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($CurrentPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$CurrentPath;$InstallDir", "User")
        $env:Path += ";$InstallDir"
        Write-Host " info: added $InstallDir to user PATH" -ForegroundColor Cyan
    }
} finally {
    Remove-Item -Recurse -Force $TmpDir -ErrorAction SilentlyContinue
}
