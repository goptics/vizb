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

$E = [char]27
function Log($msg) { Write-Host "  $E[32m>$E[0m $msg" }
function LogError($msg) { Write-Host "  $E[31m>$E[0m error: $msg"; exit 1 }

$InstallDir = "$env:LOCALAPPDATA\vizb"
$Repo = "goptics/vizb"
$Bin = "vizb.exe"

Log "detected windows/$([Environment]::Is64BitOperatingSystem ? 'amd64' : 'arm64')"
Log "fetching latest release..."

$LatestUrl = "https://github.com/$Repo/releases/latest"
$Request = [System.Net.WebRequest]::Create($LatestUrl)
$Request.AllowAutoRedirect = $false
$Response = $Request.GetResponse()
$Redirect = $Response.GetResponseHeader("Location")
$Response.Close()
$VersionTag = $Redirect -replace '.*/', ''
$Version = $VersionTag -replace '^v', ''

if (-not $Version) {
    LogError "failed to determine latest version"
}

$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "arm64" }
$Archive = "vizb@$Version-windows-$Arch.zip"
$Url = "https://github.com/$Repo/releases/download/$VersionTag/$Archive"

$TmpDir = Join-Path $env:TEMP "vizb-install-$([System.IO.Path]::GetRandomFileName())"
New-Item -ItemType Directory -Path $TmpDir | Out-Null

try {
    $ZipPath = Join-Path $TmpDir $Archive
    Log "downloading v$Version..."
    Invoke-WebRequest -Uri $Url -OutFile $ZipPath

    Log "extracting..."
    Expand-Archive -Path $ZipPath -DestinationPath $TmpDir

    $ExePath = Join-Path $InstallDir $Bin
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null

    Copy-Item (Join-Path $TmpDir $Bin) $ExePath -Force

    $Output = & $ExePath --version 2>&1
    if ($LASTEXITCODE -ne 0) {
        LogError "installation verification failed"
    }

    Log "installed vizb to $ExePath"
    Write-Host ""

    $CurrentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($CurrentPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$CurrentPath;$InstallDir", "User")
        $env:Path += ";$InstallDir"
        Log "note: added $InstallDir to user PATH"
    }

    Log "ready. run 'vizb' to get started"
} finally {
    Remove-Item -Recurse -Force $TmpDir -ErrorAction SilentlyContinue
}
