#!/usr/bin/env bash
set -euo pipefail

IMAGE_REGISTRY="${IMAGE_REGISTRY:-ghcr.io}"
IMAGE_NAMESPACE="${IMAGE_NAMESPACE:-local}"
IMAGE_TAG="${IMAGE_TAG:?Set IMAGE_TAG to the release tag you want to push}"
IMAGE_PREFIX="${IMAGE_PREFIX:-remna-user-panel}"

push_image() {
  local target="$1"
  local image="$IMAGE_REGISTRY/$IMAGE_NAMESPACE/$IMAGE_PREFIX-$target:$IMAGE_TAG"
  echo "Pushing $image"
  docker push "$image"
}

push_image backend
push_image worker
push_image frontend
