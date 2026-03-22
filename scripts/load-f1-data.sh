#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")/.."

DATA_DIR="${1:-db/f1-data}"

echo "Loading F1 data from ${DATA_DIR}..."
go run ./cmd/f1-loader --data-dir "${DATA_DIR}"
