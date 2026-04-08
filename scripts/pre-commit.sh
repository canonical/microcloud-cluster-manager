#!/bin/sh
set -e

echo "Running pre-commit checks..."

# Ensure Go binaries are in PATH
export PATH="$(go env GOPATH)/bin:$PATH"

# --- Go ---
echo "→ Golang golangci-lint"
# Checking only new changes since the last commit
golangci-lint run --timeout 10m --new-from-rev=HEAD

# --- React ---
echo "→ UI lint-staged"
cd ui
yarn lint-staged
cd ..
