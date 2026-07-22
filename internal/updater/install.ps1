# vizb installer for Windows
# Usage: irm https://vizb.goptics.org/install.ps1 | iex

param(
    [int]$Retain = 3,
    [string]$VersionTag,
    [string]$SourceExecutable
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

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
function LogError($msg) { Write-Host "  $E[31m>$E[0m error: $msg" }

$Repo = "goptics/vizb"
$Bin = "vizb.exe"
$InstallDir = Join-Path $env:LOCALAPPDATA "vizb"
$StandaloneRoot = Join-Path $env:LOCALAPPDATA "vizb-packages\standalone"
$ReleasesDir = Join-Path $StandaloneRoot "releases"

function Test-PathStartsWith {
    param(
        [string]$Path,
        [string]$Prefix
    )

    if ([string]::IsNullOrWhiteSpace($Path) -or [string]::IsNullOrWhiteSpace($Prefix)) {
        return $false
    }

    try {
        $normalizedPath = [System.IO.Path]::GetFullPath($Path)
        $normalizedPrefix = [System.IO.Path]::GetFullPath($Prefix).TrimEnd("\") + "\"
        return $normalizedPath.StartsWith($normalizedPrefix, [System.StringComparison]::OrdinalIgnoreCase)
    } catch {
        return $false
    }
}

function Test-PathContains {
    param(
        [string]$PathValue,
        [string]$Entry
    )

    if ([string]::IsNullOrWhiteSpace($PathValue)) {
        return $false
    }

    $needle = $Entry.TrimEnd("\")
    foreach ($segment in $PathValue.Split(";", [System.StringSplitOptions]::RemoveEmptyEntries)) {
        if ($segment.TrimEnd("\") -ieq $needle) {
            return $true
        }
    }
    return $false
}

function Test-IsJunction {
    param([string]$Path)

    if (-not (Test-Path -LiteralPath $Path)) {
        return $false
    }
    $item = Get-Item -LiteralPath $Path -Force
    return ($item.Attributes -band [IO.FileAttributes]::ReparsePoint) -and $item.LinkType -eq "Junction"
}

function Move-LegacyInstallDirectory {
    param([string]$Path)

    $entries = @(Get-ChildItem -LiteralPath $Path -Force)
    $unexpected = @($entries | Where-Object { $_.PSIsContainer -or $_.Name -ine $Bin })
    if (($entries | Where-Object { -not $_.PSIsContainer -and $_.Name -ieq $Bin } | Select-Object -First 1) -eq $null -or $unexpected.Count -gt 0) {
        return $null
    }

    $legacyPath = "$Path.legacy.$([System.Guid]::NewGuid().ToString("N"))"
    Move-Item -LiteralPath $Path -Destination $legacyPath
    Log "migrated legacy installation to $legacyPath"
    return $legacyPath
}

function Set-ManagedJunction {
    param(
        [string]$LinkPath,
        [string]$TargetPath,
        [string]$ManagedTargetPrefix,
        [bool]$AllowLegacyMigration = $false
    )

    $previousTarget = $null
    $legacyPath = $null

    if (Test-Path -LiteralPath $LinkPath) {
        $item = Get-Item -LiteralPath $LinkPath -Force
        if (Test-IsJunction -Path $LinkPath) {
            $previousTarget = [string]$item.Target
            if (-not (Test-PathStartsWith -Path $previousTarget -Prefix $ManagedTargetPrefix)) {
                throw "Refusing to retarget unmanaged junction at $LinkPath."
            }
            if ($previousTarget.Equals($TargetPath, [System.StringComparison]::OrdinalIgnoreCase)) {
                return
            }
            Remove-Item -LiteralPath $LinkPath -Recurse -Force
        } elseif ($item.PSIsContainer) {
            $firstEntry = Get-ChildItem -LiteralPath $LinkPath -Force | Select-Object -First 1
            if ($null -eq $firstEntry) {
                Remove-Item -LiteralPath $LinkPath -Recurse -Force
            } elseif ($AllowLegacyMigration) {
                $legacyPath = Move-LegacyInstallDirectory -Path $LinkPath
                if ([string]::IsNullOrWhiteSpace($legacyPath)) {
                    throw "Refusing to replace non-empty directory at $LinkPath because it is not a legacy Vizb installation."
                }
            } else {
                throw "Refusing to replace non-empty directory at $LinkPath with a junction."
            }
        } else {
            throw "Refusing to replace file at $LinkPath with a junction."
        }
    }

    try {
        New-Item -ItemType Directory -Force -Path (Split-Path -Parent $LinkPath) | Out-Null
        New-Item -ItemType Junction -Path $LinkPath -Target $TargetPath | Out-Null
    } catch {
        if (-not (Test-Path -LiteralPath $LinkPath)) {
            if (-not [string]::IsNullOrWhiteSpace($previousTarget)) {
                New-Item -ItemType Junction -Path $LinkPath -Target $previousTarget | Out-Null
            } elseif (-not [string]::IsNullOrWhiteSpace($legacyPath)) {
                Move-Item -LiteralPath $legacyPath -Destination $LinkPath
            }
        }
        throw
    }
}

function Invoke-WithInstallLock {
    param([scriptblock]$Script)

    $mutex = New-Object System.Threading.Mutex($false, "Local\VizbStandaloneInstaller")
    $acquired = $false
    try {
        try {
            $acquired = $mutex.WaitOne([TimeSpan]::FromMinutes(5))
        } catch [System.Threading.AbandonedMutexException] {
            $acquired = $true
        }
        if (-not $acquired) {
            throw "Timed out waiting for another Vizb installation to finish."
        }
        & $Script
    } finally {
        if ($acquired) {
            $mutex.ReleaseMutex()
        }
        $mutex.Dispose()
    }
}

function Remove-OldReleases {
    param(
        [string]$CurrentReleaseDir,
        [int]$Keep
    )

    if ($Keep -lt 1 -or -not (Test-Path -LiteralPath $ReleasesDir -PathType Container)) {
        return
    }

    $currentFullPath = [System.IO.Path]::GetFullPath($CurrentReleaseDir)
    $releaseDirs = Get-ChildItem -LiteralPath $ReleasesDir -Force -Directory |
        Where-Object { -not $_.Name.StartsWith(".staging.") } |
        Sort-Object LastWriteTimeUtc -Descending
    $kept = 0
    foreach ($directory in $releaseDirs) {
        $directoryFullPath = [System.IO.Path]::GetFullPath($directory.FullName)
        if ($directoryFullPath.Equals($currentFullPath, [System.StringComparison]::OrdinalIgnoreCase)) {
            $kept += 1
            continue
        }
        if ($kept -lt $Keep) {
            $kept += 1
            continue
        }
        Remove-Item -LiteralPath $directory.FullName -Recurse -Force -ErrorAction SilentlyContinue
    }
}

function Get-LatestVersion {
    $request = [System.Net.WebRequest]::Create("https://github.com/$Repo/releases/latest")
    $request.AllowAutoRedirect = $false
    $response = $request.GetResponse()
    try {
        $redirect = $response.GetResponseHeader("Location")
    } finally {
        $response.Close()
    }

    $versionTag = $redirect -replace '.*/', ''
    if ([string]::IsNullOrWhiteSpace($versionTag)) {
        throw "Failed to determine the latest Vizb release."
    }
    return $versionTag
}

function Get-ReleaseChecksum {
    param(
        [string]$ChecksumsPath,
        [string]$ArchiveName
    )

    $escapedArchive = [regex]::Escape($ArchiveName)
    foreach ($line in Get-Content -LiteralPath $ChecksumsPath) {
        if ($line -match "^([0-9a-fA-F]{64})\s+\*?$escapedArchive$") {
            return $Matches[1].ToLowerInvariant()
        }
    }
    throw "Checksum for $ArchiveName was not found in checksums.txt."
}

function Test-VizbBinary {
    param(
        [string]$Path,
        [string]$ExpectedVersion
    )

    try {
        $output = @(& $Path --version 2>&1)
        if ($LASTEXITCODE -ne 0) {
            return $false
        }
        return (($output -join "`n") -match ([regex]::Escape($ExpectedVersion)))
    } catch {
        return $false
    }
}

$TmpDir = $null
try {
    if (-not [string]::IsNullOrWhiteSpace($SourceExecutable) -and
        -not (Test-PathStartsWith -Path $SourceExecutable -Prefix $InstallDir) -and
        -not (Test-PathStartsWith -Path $SourceExecutable -Prefix $StandaloneRoot)) {
        throw "This standalone Vizb binary is outside the Windows installer directories. Reinstall it with irm https://vizb.goptics.org/install.ps1 | iex"
    }

    $architecture = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()
    switch ($architecture) {
        "X64" { $Arch = "amd64" }
        "Arm64" { $Arch = "arm64" }
        default { throw "Unsupported Windows architecture: $architecture" }
    }

    Log "detected windows/$Arch"
    if ([string]::IsNullOrWhiteSpace($VersionTag)) {
        Log "fetching latest release..."
        $VersionTag = Get-LatestVersion
    } elseif ($VersionTag -notmatch '^v[0-9]+\.[0-9]+\.[0-9]+(?:[-+][0-9A-Za-z.-]+)?$') {
        throw "Invalid Vizb release tag: $VersionTag"
    }
    $Version = $VersionTag -replace '^v', ''
    $Archive = "vizb@$Version-windows-$Arch.zip"
    $ReleaseBase = "https://github.com/$Repo/releases/download/$VersionTag"
    $ReleaseName = "$Version-windows-$Arch"
    $ReleaseDir = Join-Path $ReleasesDir $ReleaseName

    $TmpDir = Join-Path $env:TEMP "vizb-install-$([System.IO.Path]::GetRandomFileName())"
    New-Item -ItemType Directory -Path $TmpDir | Out-Null

    Invoke-WithInstallLock -Script {
        Get-ChildItem -LiteralPath $ReleasesDir -Force -Directory -Filter ".staging.*" -ErrorAction SilentlyContinue |
            Remove-Item -Recurse -Force -ErrorAction SilentlyContinue

        $ReleaseBinary = Join-Path $ReleaseDir $Bin
        if ((Test-Path -LiteralPath $ReleaseBinary -PathType Leaf) -and -not (Test-VizbBinary -Path $ReleaseBinary -ExpectedVersion $VersionTag)) {
            Remove-Item -LiteralPath $ReleaseDir -Recurse -Force
        }
        if (-not (Test-Path -LiteralPath $ReleaseBinary -PathType Leaf)) {
            if (Test-Path -LiteralPath $ReleaseDir) {
                Remove-Item -LiteralPath $ReleaseDir -Recurse -Force
            }

            $ZipPath = Join-Path $TmpDir $Archive
            $ChecksumsPath = Join-Path $TmpDir "checksums.txt"
            Log "downloading v$Version..."
            Invoke-WebRequest -Uri "$ReleaseBase/$Archive" -OutFile $ZipPath
            Invoke-WebRequest -Uri "$ReleaseBase/checksums.txt" -OutFile $ChecksumsPath

            $ExpectedChecksum = Get-ReleaseChecksum -ChecksumsPath $ChecksumsPath -ArchiveName $Archive
            $ActualChecksum = (Get-FileHash -LiteralPath $ZipPath -Algorithm SHA256).Hash.ToLowerInvariant()
            if ($ActualChecksum -ne $ExpectedChecksum) {
                throw "Downloaded archive checksum did not match. Expected $ExpectedChecksum but got $ActualChecksum."
            }

            Log "extracting..."
            $ExtractDir = Join-Path $TmpDir "extract"
            Expand-Archive -LiteralPath $ZipPath -DestinationPath $ExtractDir
            $ExtractedBinary = Join-Path $ExtractDir $Bin
            if (-not (Test-Path -LiteralPath $ExtractedBinary -PathType Leaf)) {
                throw "Binary was not found in the release archive."
            }

            New-Item -ItemType Directory -Force -Path $ReleasesDir | Out-Null
            $StagingDir = Join-Path $ReleasesDir ".staging.$ReleaseName.$PID"
            New-Item -ItemType Directory -Path $StagingDir | Out-Null
            Copy-Item -LiteralPath $ExtractedBinary -Destination (Join-Path $StagingDir $Bin)
            Move-Item -LiteralPath $StagingDir -Destination $ReleaseDir
        }

        if (-not (Test-VizbBinary -Path $ReleaseBinary -ExpectedVersion $VersionTag)) {
            throw "Installed Vizb binary failed verification: $ReleaseBinary --version"
        }
        Set-ManagedJunction -LinkPath $InstallDir -TargetPath $ReleaseDir -ManagedTargetPrefix $ReleasesDir -AllowLegacyMigration $true

        $ExePath = Join-Path $InstallDir $Bin
        if (-not (Test-Path -LiteralPath $ExePath -PathType Leaf)) {
            throw "Installed Vizb command is not reachable at $ExePath"
        }

        Remove-OldReleases -CurrentReleaseDir $ReleaseDir -Keep $Retain
    }

    Log "installed vizb to $(Join-Path $InstallDir $Bin)"
    Write-Host ""

    $CurrentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if (-not (Test-PathContains -PathValue $CurrentPath -Entry $InstallDir)) {
        $NewPath = if ([string]::IsNullOrWhiteSpace($CurrentPath)) { $InstallDir } else { "$CurrentPath;$InstallDir" }
        [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
        Log "note: added $InstallDir to user PATH"
    }
    if (-not (Test-PathContains -PathValue $env:Path -Entry $InstallDir)) {
        $env:Path += ";$InstallDir"
    }

    Log "ready. run 'vizb' to get started"
} catch {
    LogError $_.Exception.Message
    exit 1
} finally {
    if (-not [string]::IsNullOrWhiteSpace($TmpDir)) {
        Remove-Item -LiteralPath $TmpDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}
