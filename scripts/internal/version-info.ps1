Param(
    [Parameter(Mandatory = $true)]
    [string]$appName,

    [Parameter(Mandatory = $true)]
    [string]$appVersion,

    [Parameter()]
    [string]$mingwPath = "",

    [Parameter(HelpMessage = "Respect existing CC/CXX values instead of forcing an auto-detected GNU toolchain.")]
    [switch]$respectExistingCompiler

)

$scriptDir = Split-Path -Parent $PSCommandPath
$repoRoot = [System.IO.Directory]::GetParent([System.IO.Directory]::GetParent($scriptDir).FullName).FullName

function Resolve-RequiredCommand {
    Param(
        [Parameter(Mandatory = $true)]
        [string]$commandName,

        [Parameter(Mandatory = $true)]
        [string]$missingMessage
    )

    $command = Get-Command $commandName -ErrorAction SilentlyContinue
    if ($command) {
        return $command.Source
    }

    $command = Resolve-VisualStudioCommand $commandName
    if ($command) {
        return $command
    }

    throw $missingMessage
}

function Resolve-VisualStudioCommand {
    Param(
        [Parameter(Mandatory = $true)]
        [string]$commandName
    )

    if ($commandName -eq "rc.exe") {
        $kitsRoots = @()
        if (-not [string]::IsNullOrWhiteSpace($Env:WindowsSdkDir)) {
            $kitsRoots += Join-Path $Env:WindowsSdkDir "bin"
        }
        $kitsRoots += "C:\Program Files (x86)\Windows Kits\10\bin"

        foreach ($kitsRoot in $kitsRoots | Select-Object -Unique) {
            if (-not (Test-Path -LiteralPath $kitsRoot)) {
                continue
            }

            $candidate = Get-ChildItem -LiteralPath $kitsRoot -Directory -ErrorAction SilentlyContinue |
                Sort-Object Name -Descending |
                ForEach-Object { Join-Path $_.FullName "x64\rc.exe" } |
                Where-Object { Test-Path -LiteralPath $_ } |
                Select-Object -First 1

            if ($candidate) {
                return $candidate
            }
        }
    }

    if ($commandName -eq "cvtres.exe") {
        if (-not [string]::IsNullOrWhiteSpace($Env:VCToolsInstallDir)) {
            $candidate = Join-Path $Env:VCToolsInstallDir "bin\Hostx64\x64\cvtres.exe"
            if (Test-Path -LiteralPath $candidate) {
                return $candidate
            }
        }

        $vswherePaths = @(
            "C:\Program Files (x86)\Microsoft Visual Studio\Installer\vswhere.exe",
            "C:\Program Files\Microsoft Visual Studio\Installer\vswhere.exe"
        )

        foreach ($vswherePath in $vswherePaths) {
            if (-not (Test-Path -LiteralPath $vswherePath)) {
                continue
            }

            $installationPath = (& $vswherePath -latest -products * -property installationPath | Select-Object -First 1).Trim()
            if (-not $installationPath) {
                continue
            }

            $msvcRoot = Join-Path $installationPath "VC\Tools\MSVC"
            if (-not (Test-Path -LiteralPath $msvcRoot)) {
                continue
            }

            $candidate = Get-ChildItem -LiteralPath $msvcRoot -Directory -ErrorAction SilentlyContinue |
                Sort-Object Name -Descending |
                ForEach-Object { Join-Path $_.FullName "bin\Hostx64\x64\cvtres.exe" } |
                Where-Object { Test-Path -LiteralPath $_ } |
                Select-Object -First 1

            if ($candidate) {
                return $candidate
            }
        }
    }

    return $null
}

$versionText = $appVersion.TrimStart("v")
$versionParts = @(0, 0, 0, 0)
if ($versionText -match '^(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:.*?(\d+))?$') {
    for ($i = 1; $i -le 4; $i++) {
        if ($Matches[$i]) {
            $versionParts[$i - 1] = [int]$Matches[$i]
        }
    }
}

$rcTemplatePath = Join-Path $repoRoot "cmd/win-sound-scanner/versioninfo.rc"
$rcBuildPath = Join-Path $env:TEMP "win-sound-scanner-versioninfo.rc"
$resBuildPath = Join-Path $env:TEMP "win-sound-scanner-versioninfo.res"
$sysoPath = Join-Path $repoRoot "cmd/win-sound-scanner/versioninfo_windows.syso"

if (-not $respectExistingCompiler) {
    if ($mingwPath -ne "") {
        $windresPath = Join-Path $mingwPath "bin/x86_64-w64-mingw32-windres.exe"
    }
    else {
        $windresPath = (Get-Command "x86_64-w64-mingw32-windres.exe" -ErrorAction Stop).Source
    }
}
else {
    $developerShellMessage = "Run this script from a Visual Studio Developer PowerShell/Command Prompt, or add the required Visual Studio build tools to PATH."
    $rcPath = Resolve-RequiredCommand "rc.exe" "rc.exe was not found. $developerShellMessage"
    $cvtresPath = Resolve-RequiredCommand "cvtres.exe" "cvtres.exe was not found. $developerShellMessage"
}


$rcContent = Get-Content -LiteralPath $rcTemplatePath -Raw
$rcContent = $rcContent.Replace("__APP_NAME__", $appName)
$rcContent = $rcContent.Replace("__APP_VERSION__", $versionText)
$rcContent = $rcContent.Replace("__FILE_VERSION__", ($versionParts -join ","))
[System.IO.File]::WriteAllText($rcBuildPath, $rcContent, [System.Text.Encoding]::ASCII)

if (-not $respectExistingCompiler) {
    & $windresPath --input $rcBuildPath --output $sysoPath --output-format coff
    if ($LASTEXITCODE -ne 0) {
        throw "windres failed with exit code $LASTEXITCODE."
    }
}
else {
    & $rcPath /nologo /fo $resBuildPath $rcBuildPath
    if ($LASTEXITCODE -ne 0) {
        throw "rc.exe failed with exit code $LASTEXITCODE."
    }

    & $cvtresPath /NOLOGO /MACHINE:X64 /OUT:$sysoPath $resBuildPath
    if ($LASTEXITCODE -ne 0) {
        throw "cvtres.exe failed with exit code $LASTEXITCODE."
    }
}
