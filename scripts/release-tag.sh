#!/usr/bin/env bash
# release-tag.sh
# Tag all 46 submodules with a release version and push to all remotes.
#
# Usage:
#   bash scripts/release-tag.sh v1.0.0

set -euo pipefail

VERSION="${1:-}"
if [[ -z "$VERSION" ]]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

SUBMODULES=($(git config --file .gitmodules --get-regexp path | awk '{print $2}'))

echo "Tagging ${#SUBMODULES[@]} submodules with $VERSION..."

for submodule in "${SUBMODULES[@]}"; do
    if [[ ! -d "$submodule" ]]; then
        echo "  SKIP: $submodule (not found)"
        continue
    fi

    cd "$submodule"

    if git rev-parse "$VERSION" >/dev/null 2>&1; then
        echo "  SKIP: $submodule already tagged $VERSION"
    else
        git tag -a "$VERSION" -m "Release $VERSION"
        echo "  TAGGED: $submodule"

        # Push tag to all remotes.
        for remote in $(git remote); do
            if git push "$remote" "$VERSION" 2>/dev/null; then
                echo "    PUSHED to $remote"
            fi
        done
    fi

    cd "$REPO_ROOT"
done

# Tag root repo.
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo "SKIP: root already tagged $VERSION"
else
    git tag -a "$VERSION" -m "Release $VERSION"
    echo "TAGGED: root"

    for remote in origin github gitlab gitverse; do
        if git push "$remote" "$VERSION" 2>/dev/null; then
            echo "  PUSHED to $remote"
        fi
    done
fi

echo "Release $VERSION complete."
