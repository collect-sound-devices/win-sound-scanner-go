# PowerShell
<#
Usage (local dev):
  $env:GH_TOKEN = "<a GitHub token with repo read access>"
  .\scripts\fetch-native.ps1 -Tag v1.2.3 `
    -Repo eduarddanziger/SoundWinAgent `
    -IncludeDir native\include `
    -LibDir native\lib `
    -DllOutDir .\out

In GitHub Actions, set GH_TOKEN or GITHUB_TOKEN and call with -Tag "${{ github.ref_name }}".
#>

param(
  [Parameter(Mandatory = $true)][string]$Tag,                         # e.g. v1.2.3
  [string]$Repo = "eduarddanziger/SoundWinAgent",                     # owner/repo hosting the release
  [string]$IncludeDir = "native\include",
  [string]$LibDir = "native\lib",
  [string]$DllOutDir                                                  # optional, where your built .exe lives
)

$ErrorActionPreference = "Stop"

if (-not ($env:GH_TOKEN -or $env:GITHUB_TOKEN)) {
  throw "Set GH_TOKEN or GITHUB_TOKEN for GitHub release download."
}
$ghToken = if ($env:GH_TOKEN) { $env:GH_TOKEN } else { $env:GITHUB_TOKEN }
$env:GH_TOKEN = $ghToken

# Ensure destination dirs
New-Item -ItemType Directory -Force -Path $IncludeDir | Out-Null
New-Item -ItemType Directory -Force -Path $LibDir | Out-Null
if ($DllOutDir) { New-Item -ItemType Directory -Force -Path $DllOutDir | Out-Null }

# Temp dirs
$tmp = Join-Path $env:TEMP ("ghrel_" + [guid]::NewGuid())
$null = New-Item -ItemType Directory -Force -Path $tmp
$extractDir = Join-Path $tmp "extracted"
$null = New-Item -ItemType Directory -Force -Path $extractDir

# Download the zipped asset
$zipName = "soundagent-go-$Tag.zip"
try {
  gh release download $Tag -R $Repo -p $zipName -D $tmp | Out-Null
} catch {
  # Fallback to wildcard in case the exact name differs slightly
  gh release download $Tag -R $Repo -p "soundagent-go-*.zip" -D $tmp | Out-Null
}
$zipPath = Get-ChildItem -Path $tmp -Filter "soundagent-go-*.zip" | Select-Object -First 1 -ExpandProperty FullName
if (-not $zipPath) { throw "Zip asset 'soundagent-go-*.zip' not found on tag $Tag in $Repo." }

# Extract
Expand-Archive -Path $zipPath -DestinationPath $extractDir -Force

# Locate required files inside the extracted tree
$h   = Get-ChildItem -Path $extractDir -Recurse -Filter "SoundAgentApi.h" | Select-Object -First 1
$lib = Get-ChildItem -Path $extractDir -Recurse -Filter "SoundAgentApiDll.lib" | Select-Object -First 1
$dll = Get-ChildItem -Path $extractDir -Recurse -Filter "SoundAgentApiDll.dll" | Select-Object -First 1

if (-not $h)   { throw "Header 'SoundAgentApi.h' not found in the zip." }
if (-not $lib) { throw "Library 'SoundAgentApiDll.lib' not found in the zip." }
if ($DllOutDir -and -not $dll) { Write-Warning "DLL 'SoundAgentApiDll.dll' not found in the zip; skipping copy to '$DllOutDir'." }

# Copy to module layout
Copy-Item $h.FullName   -Destination (Join-Path $IncludeDir "SoundAgentApi.h") -Force
Copy-Item $lib.FullName -Destination (Join-Path $LibDir "SoundAgentApiDll.lib") -Force
if ($DllOutDir -and $dll) {
  Copy-Item $dll.FullName -Destination (Join-Path $DllOutDir "SoundAgentApiDll.dll") -Force
}

Write-Host "Native assets synced from '$(Split-Path -Leaf $zipPath)':"
Write-Host "  .h   -> '$IncludeDir'"
Write-Host "  .lib -> '$LibDir'"
Write-Host "  .dll -> '$DllOutDir'"

# Cleanup
Remove-Item -Recurse -Force $tmp