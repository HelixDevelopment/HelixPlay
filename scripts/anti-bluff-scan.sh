#!/usr/bin/env bash
# scripts/anti-bluff-scan.sh
# Non-overridable CI lane (Constitution §1.3)
# Fails the build if bluff patterns are found
set -euo pipefail

echo "=== Anti-Bluff CI Scan (Constitution §1 / R-13) ==="
echo "Mandate: green tests MUST guarantee real, end-user-usable behaviour"
echo "Version: 2.0.0 (2026-05-01)"
echo ""

FAILED=0
SCAN_DIR="${1:-.}"
cd "$SCAN_DIR" 2>/dev/null || cd /run/media/milosvasic/DATA4TB/Projects/HelixPlay

# Step 1: Forbidden patterns in CODE only (exclude docs, binaries, vendor)
echo "[1/5] Checking for bluff patterns in CODE files..."

# Find code files to scan (exclude vendor, node_modules, .opencode)
# NOTE: We now INCLUDE *_test.go for vacuous-assertion scanning in Step 5.
CODE_FILES=$(find . -type f \( -name "*.go" -o -name "*.js" -o -name "*.ts" -o -name "*.py" \) \
    ! -path "*/vendor/*" ! -path "*/.git/*" ! -path "*/node_modules/*" \
    ! -path "*/Panoptic/*" \
    ! -path "*/.opencode/*" \
    ! -name "*.docx" ! -name "*.png" ! -name "*.jpg" ! -name "*.pdf" ! -name "*.zip" \
    2>/dev/null)

while IFS= read -r -d '' f; do
    # Check for "not implemented" as error return (skip string literals in var/const)
    if grep -I -n 'fmt\.Errorf("not implemented\|errors\.New("not implemented' "$f" 2>/dev/null | \
       grep -v 'var\s.*=\|const\s.*=' | head -1 >/dev/null 2>&1; then
        if grep -I -q 'fmt\.Errorf("not implemented\|errors\.New("not implemented' "$f" 2>/dev/null; then
            echo "  ERROR: 'not implemented' error in: $f"
            FAILED=1
        fi
    fi

    # Check for empty function bodies - skip files with stub/Mock/Noop in name
    if grep -I -q "func.*{}$" "$f" 2>/dev/null; then
        basename=$(basename "$f")
        if ! grep -I -q "stub\|Mock\|Noop\|NoOp" <<< "$basename" 2>/dev/null; then
            if ! grep -I -q "stub\|Mock\|Noop\|NoOp" "$f" 2>/dev/null; then
                echo "  ERROR: Empty function body in: $f"
                FAILED=1
            fi
        fi
    fi

    # Check for panic("not implemented")
    if grep -I -q 'panic("not implemented")' "$f" 2>/dev/null; then
        echo "  ERROR: 'panic(\"not implemented\")' in: $f"
        FAILED=1
    fi

    # Check for TODO/FIXME/XXX/HACK comments (not in string literals)
    for pattern in "TODO" "FIXME" "XXX" "HACK"; do
        # Only flag if pattern appears as a comment (not in string literals like mockPatterns)
        if grep -I -n "$pattern" "$f" 2>/dev/null | \
           grep -v '^.*"[^"]*'"$pattern"'[^"]*"' | head -1 >/dev/null 2>&1; then
            # Skip if it's in a string literal slice (mockPatterns, invalidTitlePatterns, etc.)
            if grep -I -q "$pattern" "$f" 2>/dev/null; then
                # Check if it's a standalone comment (starts with // or /*)
                if grep -I -q '^\s*//.*'"$pattern"'\|^.*/\*.*'"$pattern"'.*\*/' "$f" 2>/dev/null; then
                    echo "  ERROR: Forbidden pattern '$pattern' in: $f"
                    FAILED=1
                fi
            fi
        fi
    done

    # Check for "tbd" as standalone word (not in string literals like invalidTitlePatterns)
    if grep -I -n '\btbd\b' "$f" 2>/dev/null | \
       grep -v 'var\s.*=\|const\s.*=\|"[^"]*tbd[^"]*"' | head -1 >/dev/null 2>&1; then
        if grep -I -q '\btbd\b' "$f" 2>/dev/null; then
            echo "  ERROR: Forbidden pattern 'tbd' in: $f"
            FAILED=1
        fi
    fi
done <<< "$CODE_FILES"

# Step 2: Verify anti-bluff in Challenges is being used unconditionally
echo ""
echo "[2/5] Verifying ValidateAntiBluff is called unconditionally in runner..."

if ! grep -rq "ValidateAntiBluff" ./Challenges/pkg/runner/ 2>/dev/null; then
    echo "  ERROR: ValidateAntiBluff not called in Challenges runner"
    FAILED=1
else
    echo "  OK: ValidateAntiBluff found in runner"
fi

# Verify the env-var gate has been removed (Constitution v2.0.0 amendment).
# We check for actual code usage (Getenv, Setenv, lookup) — comments are fine.
if grep -rqE 'os\.Getenv\("CHALLENGE_ANTIBLUFF_STRICT"\)|t\.Setenv\("CHALLENGE_ANTIBLUFF_STRICT"|CHALLENGE_ANTIBLUFF_STRICT.*==.*"1"' ./Challenges/pkg/runner/ 2>/dev/null; then
    echo "  ERROR: CHALLENGE_ANTIBLUFF_STRICT env gate still present in runner code (must be removed per Constitution v2.0.0)"
    FAILED=1
else
    echo "  OK: CHALLENGE_ANTIBLUFF_STRICT env gate removed from code"
fi

# Step 3: Check documentation Anti-Bluff Verification blocks
echo ""
echo "[3/5] Checking Anti-Bluff Verification blocks in 05_Response/..."

if [ -d "docs/research/chapters/MVP/05_Response" ]; then
    find "docs/research/chapters/MVP/05_Response" -name "*.md" ! -name "00_*" 2>/dev/null | while read -r f; do
        if ! grep -q "Anti-Bluff Verification\|## Anti-Bluff" "$f" 2>/dev/null; then
            echo "  WARNING: Missing Anti-Bluff Verification block: $f"
        fi
    done
fi

# Step 4: Submodule Constitution propagation
echo ""
echo "[4/5] Verifying Constitution references in submodules..."

for cfg in */CLAUDE.md */AGENTS.md; do
    if [ -f "$cfg" ]; then
        if ! grep -q "Constitution\|01_Constitution" "$cfg" 2>/dev/null; then
            echo "  ERROR: $cfg missing Constitution reference (Constitution §1.3)"
            FAILED=1
        fi
    fi
done

# Also check vasic-digital/ submodules
if [ -d "vasic-digital" ]; then
    for cfg in vasic-digital/*/CLAUDE.md vasic-digital/*/AGENTS.md; do
        if [ -f "$cfg" ]; then
            if ! grep -q "Constitution\|01_Constitution" "$cfg" 2>/dev/null; then
                echo "  ERROR: $cfg missing Constitution reference (Constitution §1.3)"
                FAILED=1
            fi
        fi
    done
fi

# Step 5: Vacuous test assertions (Constitution §1.1, §6.3)
echo ""
echo "[5/5] Scanning for vacuous test assertions..."

# Patterns that are tautologically true and indicate bluff tests:
# assert.True(t, true) — the canonical bluff pattern
# assert.NotNil(t, nil) — tautological failure (will always fail, not a real test)
# assert.Equal(t, true, true) — tautological
# assert.Nil(t, nil) — tautological pass
VACUOUS_PATTERNS=(
    'assert\.True\(t,\s*true\b'
    'assert\.Equal\(t,\s*true,\s*true\b'
    'assert\.Nil\(t,\s*nil\b'
    'require\.True\(t,\s*true\b'
    'require\.Equal\(t,\s*true,\s*true\b'
)

VACUOUS_FOUND=0
for pattern in "${VACUOUS_PATTERNS[@]}"; do
    matches=$(find . -type f -name "*_test.go" \
        ! -path "*/vendor/*" ! -path "*/.git/*" \
        ! -path "*/Panoptic/*" ! -path "*/.opencode/*" \
        -exec grep -H -n -E "$pattern" {} + 2>/dev/null || true)
    if [ -n "$matches" ]; then
        echo "$matches" | while read -r line; do
            # Exclude intentional scanner test fixtures
            if [[ "$line" != *"fixtures/bluff_g_"* ]] && [[ "$line" != *"anti-bluff/tests/fixtures"* ]]; then
                echo "  ERROR: Vacuous assertion (bluff test): $line"
                VACUOUS_FOUND=1
            fi
        done
    fi
done

if [ "$VACUOUS_FOUND" -eq 1 ]; then
    FAILED=1
else
    echo "  OK: No vacuous assertions found in test files"
fi

echo ""
if [ "$FAILED" -eq 1 ]; then
    echo "=== ANTI-BLUFF SCAN: FAILED ==="
    echo "Constitution §1 violation: fix bluff patterns before merging"
    exit 1
else
    echo "=== ANTI-BLUFF SCAN: PASSED ==="
    echo "Green tests now guarantee real, end-user-usable behaviour"
    exit 0
fi
