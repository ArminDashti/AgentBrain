#!/usr/bin/env bash
# Idempotent Cloud Agent install: Go toolchain, module download, npm ci, MCP build.
# Runs from the repository root. Does not start servers.
set -euo pipefail

GO_VERSION=1.26.10
if ! /usr/local/go/bin/go version 2>/dev/null | grep -q "go${GO_VERSION} "; then
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o "/tmp/go${GO_VERSION}.linux-amd64.tar.gz"
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf "/tmp/go${GO_VERSION}.linux-amd64.tar.gz"
  rm -f "/tmp/go${GO_VERSION}.linux-amd64.tar.gz"
fi
sudo ln -sfn /usr/local/go/bin/go /usr/local/bin/go
sudo ln -sfn /usr/local/go/bin/gofmt /usr/local/bin/gofmt

export GOTOOLCHAIN=local
export CGO_ENABLED=0
export PATH="/usr/local/go/bin:${PATH}"

cd api
go mod download
cd ../dispatcher-win
go mod download
cd ../webui
npm ci
cd ../AgentBrain
npm ci
npm run build
