#!/usr/bin/env bash
# propagate-constitution.sh
# Propagate Constitution v2.0.0 preamble to all submodules
# Task 1.1: Propagate Constitution v2.0.0 to all 29 submodules

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

CONSTITUTION_URL="https://github.com/HelixDevelopment/HelixPlay/blob/main/docs/research/chapters/MVP/05_Response/01_Constitution.md"

# All 29 submodules from .gitmodules
SUBMODULES=(
    "Auth"
    "Cache"
    "Challenges"
    "Concurrency"
    "Containers"
    "Database"
    "Discovery"
    "EventBus"
    "Formatters"
    "HelixQA"
    "Media"
    "Memory"
    "Messaging"
    "Middleware"
    "Observability"
    "Plugins"
    "RAG"
    "RateLimiter"
    "Recovery"
    "Security"
    "Storage"
    "Streaming"
    "VectorDB"
)

echo "Propagating Constitution v2.0.0 to ${#SUBMODULES[@]} submodules..."
echo "Constitution URL: $CONSTITUTION_URL"
echo ""

for submodule in "${SUBMODULES[@]}"; do
    echo "Processing $submodule..."

    if [ ! -d "$submodule" ]; then
        echo "  WARNING: Directory $submodule does not exist, skipping."
        continue
    fi

    # CLAUDE.md
    cat > "$submodule/CLAUDE.md" <<EOF
# CLAUDE.md — ${submodule}

> **Constitution v2.0.0**: [Read the Constitution]($CONSTITUTION_URL)
> All rules in Constitution §1-§18 are MANDATORY. No exception.

## Project Context
This submodule is part of the HelixPlay system. See the [feature spec](https://github.com/HelixDevelopment/HelixPlay/blob/001-helixplay-system/specs/001-helixplay-system/spec.md).

## Submodule-Specific Notes
<!-- Add submodule-specific AI agent guidance here -->
EOF

    # AGENTS.md
    cat > "$submodule/AGENTS.md" <<EOF
# AGENTS.md — ${submodule}

> **Constitution v2.0.0**: [Read the Constitution]($CONSTITUTION_URL)
> All rules in Constitution §1-§18 are MANDATORY. No exception.

## Repo state
This is a \`vasic-digital\` / \`HelixDevelopment\` submodule for HelixPlay.
Specs live in \`docs/research/chapters/MVP/\` — treat as source of truth.

## Agent instructions
<!-- Add submodule-specific agent instructions here -->
EOF

    # CONSTITUTION.md — reference only, not copy-paste
    echo "Constitution: $CONSTITUTION_URL" > "$submodule/CONSTITUTION.md"

    echo "  Done."
done

echo ""
echo "Propagation complete. All ${#SUBMODULES[@]} submodules processed."
