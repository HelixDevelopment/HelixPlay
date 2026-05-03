#!/usr/bin/env bash
# propagate-constitution.sh
# Propagate Constitution v2.3.0 preamble and anti-bluff enforcement to all submodules

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

CONSTITUTION_URL="https://github.com/HelixDevelopment/HelixPlay/blob/main/docs/research/chapters/MVP/05_Response/01_Constitution.md"

# Read submodules dynamically from .gitmodules
SUBMODULES=($(git config --file .gitmodules --get-regexp path | awk '{print $2}'))

echo "Propagating Constitution v2.3.0 to ${#SUBMODULES[@]} submodules..."
echo "Constitution URL: $CONSTITUTION_URL"
echo ""

PRIME_DIRECTIVE="> **NON-NEGOTIABLE PRIME DIRECTIVE:**
> **\"We had been in position that all tests do execute with success and
> all Challenges as well, but in reality the most of the features does
> not work and can't be used! This MUST NOT be the case and execution
> of tests and Challenges MUST guarantee the quality, the completion
> and full usability by end users of the product!\"**
> This statement is the foundational requirement of this project. Any
> agent dispatch, any CI configuration, any code review that allows
> green tests on broken features is a violation and MUST be rejected."

for submodule in "${SUBMODULES[@]}"; do
    echo "Processing $submodule..."

    if [ ! -d "$submodule" ]; then
        echo "  WARNING: Directory $submodule does not exist, skipping."
        continue
    fi

    cat > "$submodule/CLAUDE.md" <<EOF
# CLAUDE.md — ${submodule}

$PRIME_DIRECTIVE

> **Constitution v2.3.0**: [Read the Constitution]($CONSTITUTION_URL)
> All rules in Constitution §1-§21 are MANDATORY. No exception.
>
> **Amendments (2026-05-02):**
> - Anti-bluff: forbidden patterns include \`assert.True(t, true)\`,
>   \`assert.NotNil(t, nil)\`, constructor-only tests, mock-only
>   integration/E2E tests, and permanently skipped tests without
>   containerization plans.
> - Usability evidence mandatory per §6.7 (HelixQA visual assertion,
>   manual recording, or Challenge scenario).
> - Automatic negative-leg fault injection per §1.3 / §6.3 / §11.5.7 —
>   CI breaks each feature and verifies non-Unit tests fail.
> - \`ValidateAntiBluff\` unconditional; all challenges call \`RecordAction()\`.
> - Container verifier \`execCommand()\` executes real commands.
> - \`go vet ./...\` MUST pass with zero warnings — no suppressions, no exceptions.
> - Anti-bluff scan MUST fail the CI lane: \`scripts/anti-bluff-scan.sh\` exits
>   non-zero on any violation. Process substitution (\`< <(...)>\`) required over
>   pipes for variable state propagation; subshell-based patterns that silently
>   drop failure state are forbidden.
> - Observable behaviour assertion ratio: at least 60% of assertions must verify
>   observable behaviour per §1.2.
> - Mutation score >= 85% enforced by \`mutation_ratchet_challenge.sh\` per §6.4.
> - The 18 Contract Clauses (R-01..R-18) codified in §17.
> - Eight Architectural Pillars codified in §18 — binding architectural decisions.
> - Performance SLAs codified in §19 — <=30ms LAN, <=50ms WAN at p999.
> - Technology Stack codified in §20 — mandatory technology choices.
> - Implementation Roadmap codified in §21 — 14 phases (P00–P13).

## Project Context
This submodule is part of the HelixPlay system.
See the [feature spec](https://github.com/HelixDevelopment/HelixPlay/blob/001-helixplay-system/specs/001-helixplay-system/spec.md).

## Submodule-Specific Notes
<!-- Add submodule-specific AI agent guidance here -->
EOF

    cat > "$submodule/AGENTS.md" <<EOF
# AGENTS.md — ${submodule}

$PRIME_DIRECTIVE

> **Constitution v2.3.0**: [Read the Constitution]($CONSTITUTION_URL)
> All rules in Constitution §1-§21 are MANDATORY. No exception.
>
> **Amendments (2026-05-02):**
> - Anti-bluff enforcement strengthened: no vacuous assertions, no
>   constructor-only tests, no mock-only integration/E2E tests, no
>   untriaged skips.
> - Usability evidence mandatory per §6.7.
> - Automatic negative-leg fault injection per §1.3 / §6.3 / §11.5.7.
> - \`ValidateAntiBluff\` unconditional; all challenges call \`RecordAction()\`.
> - Container verifier \`execCommand()\` executes real commands.
> - \`go vet ./...\` MUST pass with zero warnings — no suppressions, no exceptions.
> - Anti-bluff scan MUST fail the CI lane: \`scripts/anti-bluff-scan.sh\` exits
>   non-zero on any violation. Process substitution (\`< <(...)>\`) required over
>   pipes for variable state propagation.
> - Observable behaviour assertion ratio: at least 60% of assertions must verify
>   observable behaviour per §1.2.
> - Mutation score >= 85% enforced by \`mutation_ratchet_challenge.sh\` per §6.4.
> - The 18 Contract Clauses (R-01..R-18) codified in §17.
> - Eight Architectural Pillars codified in §18 — binding architectural decisions.
> - Performance SLAs codified in §19 — <=30ms LAN, <=50ms WAN at p999.
> - Technology Stack codified in §20 — mandatory technology choices.
> - Implementation Roadmap codified in §21 — 14 phases (P00–P13).

## Repo state
This is a \`vasic-digital\` / \`HelixDevelopment\` submodule for HelixPlay.

## Critical constraints
- **Anti-bluff:** No placeholders, dead code, vacuous tests. Details in Constitution §1.
- **Containers only:** Every service, DB, build, test runs inside a container.
- **Decoupling:** Reusable components live in public \`vasic-digital\` submodules.
- **Tests:** 100% coverage across all ten types. Only Unit may use mocks.
- **R-18 Operational Integrity:** No command may suspend/hibernate/lock/terminate/crash the host.

## Git topology
\`origin\` fetch=GitHub, push=GitFlic. Four remotes configured.
Force-push requires explicit authorization. \`--no-verify\` is forbidden.
EOF

    cat > "$submodule/CONSTITUTION.md" <<EOF
# CONSTITUTION.md — ${submodule}

> **Source of truth:** [$CONSTITUTION_URL]($CONSTITUTION_URL)
>
> This submodule adopts the HelixPlay Constitution v2.3.0 in full.
> All clauses §1-§21 are binding. No local weakening permitted.
EOF

    echo "  Updated CLAUDE.md, AGENTS.md, CONSTITUTION.md"
done

echo ""
echo "Constitution v2.3.0 propagation complete."
echo "Next: git add + commit in each submodule, then push."
