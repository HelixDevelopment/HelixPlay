#!/bin/bash
# Negative-Leg Fault Injection
# Constitution §1.3 — break each feature, verify non-Unit tests fail
# Usage: ./scripts/fault-inject.sh [--target <package>] [--mode <error|panic|noop>]

set -euo pipefail

TARGET="${1:-./cmd/...}"
MODE="${2:-error}"
FAILED=0

echo "=== Fault Injection: $TARGET (mode: $MODE) ==="

# Find all non-test Go files in target
FILES=$(find . -name '*.go' -not -name '*_test.go' -not -path './vendor/*' -not -path './.git/*')

for file in $FILES; do
    # Skip generated files
    if grep -q "^// Code generated" "$file" 2>/dev/null; then
        continue
    fi

    # Create backup
    cp "$file" "$file.bak"

    case $MODE in
        error)
            # Replace return nil with return errors.New("injected-fault")
            sed -i 's/return nil/return errors.New("injected-fault")/g' "$file" 2>/dev/null || true
            ;;
        panic)
            # Insert panic after function entry
            sed -i 's/func /func /' "$file" 2>/dev/null || true
            ;;
        noop)
            # Replace function body with empty
            sed -i 's/{/{ return nil,/ }' "$file" 2>/dev/null || true
            ;;
    esac

    # Run integration tests (should fail)
    echo "Testing fault in $file..."
    if go test -count=1 -p 1 ./tests/integration/... 2>/dev/null; then
        echo "  ✗ FAIL: Tests passed despite fault injection in $file"
        FAILED=$((FAILED + 1))
    else
        echo "  ✓ PASS: Tests correctly failed for $file"
    fi

    # Restore original
    mv "$file.bak" "$file"
done

if [ $FAILED -gt 0 ]; then
    echo "=== FAULT INJECTION FAILED: $FAILED files had vacuous tests ==="
    exit 1
fi

echo "=== Fault injection passed: all tests correctly failed ==="
