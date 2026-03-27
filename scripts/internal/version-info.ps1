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

    throw $missingMessage
}

function Resolve-OptionalCommand {
    Param(
        [Parameter(Mandatory = $true)]
        [string]$commandName
    )

    $command = Get-Command $commandName -ErrorAction SilentlyContinue
    if ($command) {
        return $command.Source
    }

    return $null
}

function Resolve-WindresPath {
    Param(
        [Parameter()]
        [string]$compilerCommand,

        [Parameter()]
        [string]$toolchainRoot,

        [Parameter(Mandatory = $true)]
        [string]$windresExeName
    )

    if (-not [string]::IsNullOrWhiteSpace($compilerCommand)) {
        $compilerPath = $compilerCommand.Trim().Trim('"')
        if (-not (Test-Path -LiteralPath $compilerPath)) {
            $compilerPath = Resolve-OptionalCommand $compilerPath
        }

        if ($compilerPath -and (Test-Path -LiteralPath $compilerPath)) {
            $windresCandidatePath = Join-Path (Split-Path -Parent $compilerPath) $windresExeName
            if (Test-Path -LiteralPath $windresCandidatePath) {
                return $windresCandidatePath
            }
        }
    }

    if (-not [string]::IsNullOrWhiteSpace($toolchainRoot)) {
        $windresCandidatePath = Join-Path $toolchainRoot (Join-Path "bin" $windresExeName)
        if (Test-Path -LiteralPath $windresCandidatePath) {
            return $windresCandidatePath
        }
    }

    return Resolve-OptionalCommand $windresExeName
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
$windresExeName = "x86_64-w64-mingw32-windres.exe"

if (-not $respectExistingCompiler) {
    $windresPath = Resolve-WindresPath -compilerCommand $Env:CC -toolchainRoot $mingwPath -windresExeName $windresExeName
    if (-not $windresPath) {
        throw "$windresExeName was not found. Set -mingwPath to a valid llvm-mingw root or add $windresExeName to PATH."
    }

    $resourceCompilerMode = "windres"
}
else {
    $windresPath = Resolve-WindresPath -compilerCommand $Env:CC -toolchainRoot "" -windresExeName $windresExeName
    if ($windresPath) {
        $resourceCompilerMode = "windres"
    }
    else {
        $developerShellMessage = "Run this script from a Visual Studio Developer PowerShell/Command Prompt, or add the required Visual Studio build tools to PATH."
        $rcPath = Resolve-OptionalCommand "rc.exe"
        $cvtresPath = Resolve-OptionalCommand "cvtres.exe"

        if ($rcPath -and $cvtresPath) {
            $resourceCompilerMode = "msvc"
        }
        else {
            throw "Neither $windresExeName next to CC='$Env:CC' nor rc.exe/cvtres.exe were found. $developerShellMessage"
        }
    }
}


$rcContent = Get-Content -LiteralPath $rcTemplatePath -Raw
$rcContent = $rcContent.Replace("__APP_NAME__", $appName)
$rcContent = $rcContent.Replace("__APP_VERSION__", $versionText)
$rcContent = $rcContent.Replace("__FILE_VERSION__", ($versionParts -join ","))
[System.IO.File]::WriteAllText($rcBuildPath, $rcContent, [System.Text.Encoding]::ASCII)

if ($resourceCompilerMode -eq "windres") {
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
