#!/usr/bin/env bash
set -euo pipefail

# Requires tinygo installed: https://tinygo.org/getting-started/
# Example install:
#   go install github.com/tinygo-org/tinygo@latest

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

tinygo build \
  -o "${SCRIPT_DIR}/hello-job.wasm" \
  -target=wasi \
  "${SCRIPT_DIR}/main.go"

echo "Built ${SCRIPT_DIR}/hello-job.wasm"
