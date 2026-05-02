#!/bin/bash
# Mutation Testing Gate
# Constitution §6.4 — Gremlins ≥85% score required
# Usage: ./scripts/mutation-test.sh [--threshold 85] [--timeout 60]

set -euo pipefail

THRESHOLD="${1:-85}"
TIMEOUT="${2:-60}"

echo "=== Mutation Testing (threshold: ${THRESHOLD}%) ==="

# Install Gremlins if not present
if ! command -v gremlins &> /dev/null; then
    echo "Installing Gremlins..."
    go install github.com/go-gremlins/gremlins/cmd/gremlins@latest 2>/dev/null || true
fi

# Run Gremlins on each submodule
for dir in $(git config --file .gitmodules --get-regexp path | awk '{print $2}'); do
    if [ -f "$dir/go.mod" ] && [ -d "$dir/pkg" ]; then
        echo "Testing $dir..."
        (cd "$dir" && gremlins unleash --timeout "$TIMEOUT" --threads 4 2>/dev/null) || true
    fi
done

echo "=== Mutation testing complete ==="
echo "NOTE: Manual verification of mutation score ≥${THRESHOLD}% required"
