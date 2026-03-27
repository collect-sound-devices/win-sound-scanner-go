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
    $rcPath = (Get-Command "rc.exe" -ErrorAction Stop).Source
    $cvtresPath = (Get-Command "cvtres.exe" -ErrorAction Stop).Source
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
