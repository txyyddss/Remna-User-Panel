#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [[ ! -d frontend/node_modules ]]; then
  echo "npm install --prefix frontend"
  npm --prefix frontend install
fi

echo "go mod download"
go mod download

echo "npm run check"
npm run check
