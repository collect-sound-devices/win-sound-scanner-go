Param(
    [Parameter(Mandatory = $true)]
    [string]$TargetPath,

    [Parameter()]
    [string]$PfxPath = $env:WIN_SOUND_SIGN_PFX_PATH,

    [Parameter()]
    [string]$CertificatePassword = $env:WIN_SOUND_SIGN_PFX_PASSWORD,

    [Parameter()]
    [string]$TimestampUrl = $(if ([string]::IsNullOrWhiteSpace($env:WIN_SOUND_SIGN_TIMESTAMP_URL)) { "http://timestamp.digicert.com" } else { $env:WIN_SOUND_SIGN_TIMESTAMP_URL }),

    [Parameter()]
    [string]$SignToolPath = $env:WIN_SOUND_SIGNTOOL_PATH
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path -LiteralPath $TargetPath)) {
    throw "Target file not found: $TargetPath"
}

$hasPfxPath = -not [string]::IsNullOrWhiteSpace($PfxPath)
$hasPassword = -not [string]::IsNullOrWhiteSpace($CertificatePassword)

if (-not $hasPfxPath -and -not $hasPassword) {
    Write-Host "Skipping signing: WIN_SOUND_SIGN_PFX_PATH and WIN_SOUND_SIGN_PFX_PASSWORD are not set."
    exit 0
}

if (-not $hasPfxPath) {
    throw "WIN_SOUND_SIGN_PFX_PATH is required when WIN_SOUND_SIGN_PFX_PASSWORD is set."
}

if (-not $hasPassword) {
    throw "WIN_SOUND_SIGN_PFX_PASSWORD is required when WIN_SOUND_SIGN_PFX_PATH is set."
}

if (-not (Test-Path -LiteralPath $PfxPath)) {
    throw "PFX file not found: $PfxPath"
}

if ([string]::IsNullOrWhiteSpace($SignToolPath)) {
    $kitsBin = Join-Path ${env:ProgramFiles(x86)} "Windows Kits\10\bin"
    $signTool = $null

    if (Test-Path -LiteralPath $kitsBin) {
        $signTool = Get-ChildItem -Path $kitsBin -Filter "signtool.exe" -Recurse -ErrorAction SilentlyContinue |
            Where-Object { $_.FullName -like "*\x64\signtool.exe" } |
            Sort-Object FullName -Descending |
            Select-Object -First 1
    }

    if ($signTool) {
        $SignToolPath = $signTool.FullName
    } else {
        $command = Get-Command "signtool.exe" -ErrorAction SilentlyContinue
        if ($command) {
            $SignToolPath = $command.Source
        }
    }
}

if ([string]::IsNullOrWhiteSpace($SignToolPath)) {
    throw "signtool.exe not found. Set WIN_SOUND_SIGNTOOL_PATH or install the Windows SDK."
}

Write-Host "Signing '$TargetPath' ..."
& $SignToolPath sign /fd sha256 /td sha256 /f $PfxPath /p $CertificatePassword /tr $TimestampUrl $TargetPath
if ($LASTEXITCODE -ne 0) {
    throw "signtool.exe failed with exit code $LASTEXITCODE."
}
