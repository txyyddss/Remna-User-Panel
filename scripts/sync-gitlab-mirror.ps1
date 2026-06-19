param(
    [string]$PrimaryRemote,
    [string]$MirrorRemote,
    [string[]]$Branches,
    [switch]$Force,
    [switch]$Prune,
    [switch]$NoTags,
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"

if (-not $PrimaryRemote) {
    $PrimaryRemote = if ($env:PRIMARY_REMOTE) { $env:PRIMARY_REMOTE } else { "origin" }
}
if (-not $MirrorRemote) {
    $MirrorRemote = if ($env:MIRROR_REMOTE) { $env:MIRROR_REMOTE } else { "gitlab" }
}
if (-not $Branches -and $env:SYNC_BRANCHES) {
    $Branches = @($env:SYNC_BRANCHES -split "[,;\s]+" | Where-Object { $_ })
}
if ($env:FORCE -eq "true") {
    $Force = $true
}
if ($env:PRUNE -eq "true") {
    $Prune = $true
}
if ($env:SYNC_TAGS -eq "false") {
    $NoTags = $true
}
if ($env:DRY_RUN -eq "true") {
    $DryRun = $true
}

function Invoke-Git {
    param([string[]]$Arguments)

    if ($DryRun) {
        Write-Host ("git " + ($Arguments -join " "))
        return
    }

    & git @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "git $($Arguments -join ' ') failed with exit code $LASTEXITCODE"
    }
}

Invoke-Git -Arguments @("fetch", $PrimaryRemote, "--prune", "--tags")

if ($Force) {
    Invoke-Git -Arguments @("fetch", $MirrorRemote, "--prune")
}

if (-not $Branches) {
    $Branches = @(
        git for-each-ref "--format=%(refname:strip=3)" "refs/remotes/$PrimaryRemote" |
            Where-Object { $_ -and $_ -ne "HEAD" }
    )
}

foreach ($branch in $Branches) {
    if (-not $branch) {
        continue
    }

    $src = "refs/remotes/$PrimaryRemote/$branch"
    $dst = "refs/heads/$branch"
    $args = @("push", $MirrorRemote)
    if ($Force) {
        $args += "--force-with-lease=$dst"
    }
    $args += "${src}:${dst}"
    Invoke-Git -Arguments $args
}

if (-not $NoTags) {
    Invoke-Git -Arguments @("push", $MirrorRemote, "--tags")
}

if ($Prune) {
    $primaryBranchSet = @{}
    foreach ($branch in $Branches) {
        $primaryBranchSet[$branch] = $true
    }

    $mirrorBranches = @(
        git ls-remote --heads $MirrorRemote |
            ForEach-Object {
                if ($_ -match "refs/heads/(.+)$") {
                    $Matches[1]
                }
            }
    )

    foreach ($mirrorBranch in $mirrorBranches) {
        if (-not $primaryBranchSet.ContainsKey($mirrorBranch)) {
            Invoke-Git -Arguments @("push", $MirrorRemote, ":refs/heads/$mirrorBranch")
        }
    }
}
