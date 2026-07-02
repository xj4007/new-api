param(
    [string]$GoExe = 'C:\Program Files\Go\bin\go.exe',
    [string]$GoProxy = 'https://goproxy.cn,direct',
    [switch]$SkipInstall
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = (Resolve-Path (Join-Path $scriptDir '..\..\..')).Path
$webRoot = Join-Path $repoRoot 'web'
$defaultRoot = Join-Path $webRoot 'default'
$classicRoot = Join-Path $webRoot 'classic'
$releaseDir = Join-Path $scriptDir 'release'
$binaryPath = Join-Path $releaseDir 'new-api'
$versionFile = Join-Path $repoRoot 'VERSION'
$buildInfoPath = Join-Path $releaseDir 'build-info.txt'

if (!(Test-Path $GoExe)) {
    throw "Go executable not found: $GoExe"
}

New-Item -ItemType Directory -Force $releaseDir | Out-Null

if (!$SkipInstall) {
    bun install --cwd $webRoot --frozen-lockfile
}

Push-Location $defaultRoot
try {
    bun run build
}
finally {
    Pop-Location
}

# classic currently needs its own filtered install to keep Semi/date-fns deps compatible.
# Run this after default has finished building so it doesn't perturb default's dependency graph.
if (!$SkipInstall) {
    bun install --cwd $webRoot --filter ./classic --frozen-lockfile
}

Push-Location $classicRoot
try {
    bun run build
}
finally {
    Pop-Location
}

$env:GOPROXY = $GoProxy
$env:CGO_ENABLED = '0'
$env:GOOS = 'linux'
$env:GOARCH = 'amd64'

$version = ''
if (Test-Path $versionFile) {
    $rawVersion = Get-Content $versionFile -Raw
    if ($null -ne $rawVersion) {
        $version = $rawVersion.Trim()
    }
}
$ldflags = "-s -w -X 'github.com/QuantumNous/new-api/common.Version=$version'"

Push-Location $repoRoot
try {
    & $GoExe build -ldflags $ldflags -o $binaryPath
}
finally {
    Pop-Location
}

$sha256 = (Get-FileHash $binaryPath -Algorithm SHA256).Hash.ToLowerInvariant()
$commit = (& git -C $repoRoot rev-parse HEAD).Trim()
$branch = (& git -C $repoRoot branch --show-current).Trim()
$timestamp = (Get-Date).ToString('yyyy-MM-dd HH:mm:ss K')

@(
    "branch=$branch"
    "commit=$commit"
    "sha256=$sha256"
    "built_at=$timestamp"
    "go_proxy=$GoProxy"
) | Set-Content -Path $buildInfoPath -Encoding ascii

Write-Host "Built release binary: $binaryPath"
Write-Host "SHA256: $sha256"
