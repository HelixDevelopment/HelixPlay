#!/usr/bin/env bash
# propagate-constitution.sh
# Propagate Constitution v2.0.0 preamble to all submodules
# Task 1.1: Propagate Constitution v2.0.0 to all submodules

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

CONSTITUTION_URL="https://github.com/HelixDevelopment/HelixPlay/blob/main/docs/research/chapters/MVP/05_Response/01_Constitution.md"

# Read submodules dynamically from .gitmodules
SUBMODULES=($(git config --file .gitmodules --get-regexp path | awk '{print $2}'))

echo "Propagating Constitution v2.0.0 to ${#SUBMODULES[@]} submodules..."
echo "Constitution URL: $CONSTITUTION_URL"
echo ""

for submodule in "${SUBMODULES[@]}"; do
    echo "Processing $submodule..."

    if [ ! -d "$submodule" ]; then
        echo "  WARNING: Directory $submodule does not exist, skipping."
        continue
    fi

    # CLAUDE.md - keep as is (already fine)
    cat > "$submodule/CLAUDE.md" <<EOF
# CLAUDE.md — ${submodule}

> **Constitution v2.0.0**: [Read the Constitution]($CONSTITUTION_URL)
> All rules in Constitution §1-§18 are MANDATORY. No exception.

## Project Context
This submodule is part of the HelixPlay system. See the [feature spec](https://github.com/HelixDevelopment/HelixPlay/blob/001-helixplay-system/specs/001-helixplay-system/spec.md).

## Submodule-Specific Notes
<!-- Add submodule-specific AI agent guidance here -->
EOF

    # AGENTS.md - enhanced with git topology and critical constraints
    cat > "$submodule/AGENTS.md" <<EOF
# AGENTS.md — ${submodule}

> **Constitution v2.0.0**: [Read the Constitution]($CONSTITUTION_URL)
> All rules in Constitution §1-§18 are MANDATORY. No exception.

## Repo state
This is a \`vasic-digital\` / \`HelixDevelopment\` submodule for HelixPlay.
Specs live in \`docs/research/chapters/MVP/\` — treat as source of truth.

## Git topology
Four remotes; \`origin\` is **split**: fetch from GitHub, push to GitFlic.

\`\`\`bash
github      git@github.com:HelixDevelopment/HelixPlay.git
gitlab      git@gitlab.com:helixdevelopment1/HelixPlay.git
gitverse    git@gitverse.ru:helixdevelopment/HelixPlay.git
gitflic     git@gitflic.ru:helixdevelopment/helixplay.git
origin      fetch=github, push=gitflic
\`\`\`

When operator says "push", confirm which mirror — \`origin\` only updates GitFlic. Force-push requires explicit authorization. \`--no-verify\` is forbidden.

## Critical constraints

These are mandatory project-wide rules, not suggestions:

- **Anti-bluff:** No \`TODO\`, \`FIXME\`, \`XXX\`, \`placeholder\`, empty function bodies, dead code, or tests that pass without exercising real behavior. Details in Constitution §1.
- **Containers only:** Every service, DB, build step, test runner, and scanner runs inside a container. Definitions live in \`vasic-digital/Containers\` — never vendor a \`Dockerfile\` outside that submodule. No faking a local toolchain.
- **Decoupling:** Reusable components live in **public** \`vasic-digital\` Git/Go submodules. Reuse before recreating.

## Agent instructions
<!-- Add submodule-specific agent instructions here -->
EOF

    # CONSTITUTION.md — proper reference file
    cat > "$submodule/CONSTITUTION.md" <<EOF
# Constitution Reference

This submodule follows [HelixPlay Constitution v2.0.0]($CONSTITUTION_URL).

All rules in Constitution §1-§18 are MANDATORY. No exception.

> **Source of truth:** [$CONSTITUTION_URL]($CONSTITUTION_URL)
EOF

    echo "  Done."
done

echo ""
echo "Propagation complete. All ${#SUBMODULES[@]} submodules processed."
