#!/usr/bin/env bash
# commit-submodules.sh
# Commit propagated constitution files in each submodule

set -e

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

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

for submodule in "${SUBMODULES[@]}"; do
    echo "--- Committing in $submodule ---"
    cd "$REPO_ROOT/$submodule"
    git add CLAUDE.md AGENTS.md CONSTITUTION.md
    git commit -m "feat: add Constitution v2.0.0 preamble (R-15)" || echo "  (nothing to commit or already committed)"
    cd "$REPO_ROOT"
done

echo ""
echo "All submodules committed."
