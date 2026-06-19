$ErrorActionPreference = "Stop"

if (-not $env:IMAGE_TAG) {
    throw "Set IMAGE_TAG to the release tag you want to push"
}

$imageRegistry = if ($env:IMAGE_REGISTRY) { $env:IMAGE_REGISTRY } else { "ghcr.io" }
$imageNamespace = if ($env:IMAGE_NAMESPACE) { $env:IMAGE_NAMESPACE } else { "local" }
$imageTag = $env:IMAGE_TAG
$imagePrefix = if ($env:IMAGE_PREFIX) { $env:IMAGE_PREFIX } else { "remna-user-panel" }

function Push-Image {
    param([string]$Target)
    $image = "$imageRegistry/$imageNamespace/$imagePrefix-$Target`:$imageTag"
    Write-Host "Pushing $image" -ForegroundColor Cyan
    docker push $image
}

Push-Image backend
Push-Image worker
