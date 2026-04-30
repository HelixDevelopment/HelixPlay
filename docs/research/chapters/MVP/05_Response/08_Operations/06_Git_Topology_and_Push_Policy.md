# O06 — Git Topology & Push Policy

> **Source dimensions:** [`00_Index.md`](00_Index.md); [`../01_Constitution.md`](../01_Constitution.md) §9 (Source Control & Git Topology); [`../06_Submodules/01_Submodule_Catalog.md`](../06_Submodules/01_Submodule_Catalog.md) §2 (polyrepo decision); the repo-root [`/CLAUDE.md`](../../../../../CLAUDE.md) "Git topology" section.
> **Source line count:** chapter floor 300 lines per Master Plan §7.2 row O06.
> **Chapter targets:** Constitution §9; R-18 (no destructive git operation may delete operator work).
> **Cross-links:** [`01_Container_CI_CD.md`](01_Container_CI_CD.md), [`05_Tracking_GitHub_GitLab.md`](05_Tracking_GitHub_GitLab.md).
> **Status:** Draft v1.
> **Last updated:** 2026-04-30.

---

## 1. The Topology in Three Sentences

HelixPlay is hosted on **four mirrors in two jurisdictions**: GitHub + GitLab (Western) and GitFlic + GitVerse (Russian). The `origin` remote is a **composite push** that pushes to all four simultaneously; pushing to any single mirror by name updates only that mirror. Every of the 29 `vasic-digital/` submodule repositories follows the same four-mirror pattern; the consistency lets a developer treat "the project" as one logical thing while operationally satisfying the jurisdictional-mirror requirement.

---

## 2. The Four Remotes (Repo-Root)

```
github      git@github.com:HelixDevelopment/HelixPlay.git
gitlab      git@gitlab.com:helixdevelopment1/HelixPlay.git
gitverse    git@gitverse.ru:helixdevelopment/HelixPlay.git
gitflic     git@gitflic.ru:helixdevelopment/helixplay.git
origin      fetch=github, push=gitflic+gitlab+gitverse+gitflic (composite)
upstream    git@gitflic.ru:helixdevelopment/helixplay.git
```

`origin`'s split fetch/push design — fetch from GitHub but push to all four — is a `git remote set-url --push --add` configuration documented in CLAUDE.md "Git topology". Verified post-clone via `git remote -v`.

---

## 3. The Composite-Push Pattern

```bash
# Push to all four mirrors with one command.
git push origin main

# Push to a single mirror by name (rare; debugging only).
git push github main
git push gitlab main
```

Output of `git push origin main` shows four `To <mirror>:<repo>` lines — one per mirror. A failure on one mirror surfaces clearly; the operator can retry that specific mirror via `git push <name> main`.

The composite-push is per-repository — the same pattern applies to every of the 29 `vasic-digital/` submodule repositories (per [S01 §2.2](../06_Submodules/01_Submodule_Catalog.md#22-the-four-mirror-amplifier)).

---

## 4. The Push-Policy Rules

Per CLAUDE.md + Constitution §9:

1. **Always push to `origin`** for normal merges → the composite push fans out to all four.
2. **Confirm with the operator before any single-mirror push** — pushing to `github` only leaves the other three behind, which violates four-mirror parity.
3. **Never `git push --force` to `main`** without operator approval per Constitution §13 *Exceptions*. The branch protection on `main` rejects force-push by default.
4. **Never `git push --force-with-lease`** unsupervised — same blast-radius as `--force` if upstream has unexpected commits.
5. **Tag pushes** also use composite: `git push origin --tags`. Per [S01 §4.1](../06_Submodules/01_Submodule_Catalog.md#41-module-versioning-and-semantic-import-versioning-siv), tag parity across mirrors is mandatory; a tag on three mirrors but missing on the fourth blocks the release-train.

---

## 5. Mirror-Sync Verification

After every push, verify all four mirrors at the same SHA:

```bash
for r in github gitlab gitflic gitverse; do
    sha=$(git ls-remote $r main | head -1 | awk '{print $1}')
    echo "$r: $sha"
done
```

Output: four lines, all the same SHA. Drift between mirrors is a P3 alert (per [T12 §4.3](../07_Testing/12_HelixQA_Autonomous.md#43-p3--flake-or-info)) until reconciled.

---

## 6. Branch Protection Rules

Each mirror enforces:

| Rule                                | Enforced on              |
|-------------------------------------|--------------------------|
| Required PR reviews                  | `main` (≥ 1; ≥ 2 for `helix-r18-safeexec`) |
| Required status checks               | the 14 (test-type, platform) cells per T01 §2 |
| Restrict pushes to PR-only          | `main`                   |
| No force-push                        | `main` + tags `v*`       |
| No deletion                          | `main` + tags `v*`       |
| Required signed commits              | (operator-optional; recommended for tag-pushes) |

The four mirrors must all have the same protection set; drift creates a "soft spot" through which a maintainer could bypass the gate. The `helix-branch-protection-audit` cron audits all four nightly.

---

## 7. Submodule-Repository Topology

Every of the 29 `vasic-digital/` submodule repositories follows the **same four-mirror pattern** as the parent HelixPlay repo:

```
github      git@github.com:vasic-digital/<name>.git
gitlab      git@gitlab.com:vasic-digital/<name>.git
gitverse    git@gitverse.ru:vasic-digital/<name>.git
gitflic     git@gitflic.ru:vasic-digital/<name>.git
origin      fetch=github, push=composite
```

A new submodule's first push is via the bulk-create tool at `vasic-digital/Containers/scripts/submodule-bootstrap.sh`:

```bash
./submodule-bootstrap.sh vasic-digital/helix-newthing \
    --description "..." \
    --licence MIT
```

The script:
1. Creates the GitHub repo (`gh repo create`).
2. Creates the GitLab project (`glab repo create`).
3. Creates the GitFlic project (REST API).
4. Creates the GitVerse project (manual fallback — no API per [S01 §4.7.4](../06_Submodules/01_Submodule_Catalog.md#474-gitverse)).
5. Configures the four remotes locally + sets up `origin`'s composite push.
6. Pushes the initial commit to all four.

---

## 8. Ticket-Mirror Coordination with O05

When the operator opens a phase or task issue per [O05](05_Tracking_GitHub_GitLab.md), the tracking-mirror webhook automatically creates the GitLab counterpart. The same git-mirror discipline applies — every issue exists on both platforms; closing one closes the other.

GitFlic + GitVerse currently lack a GitHub-Projects-v2-equivalent (per [O05 §10 OQ-O05-C](05_Tracking_GitHub_GitLab.md#10-open-questions)); ticket mirroring there is deferred to a future Constitution amendment. The git-side mirroring is unaffected — code mirrors all four; tickets mirror only the two GitHub-Projects-supporting ones (until OQ-O05-C closes).

---

## 9. Recovery from a Drift Event

A drift event = mirrors diverging in commit history (e.g. operator force-pushed to `gitflic` only).

The recovery procedure:

1. Identify the **canonical mirror** — usually GitHub (the project's primary distribution).
2. Reset the drifted mirror: `git push --force gitflic main` (operator-only; logged in run-archive).
3. Verify SHA parity via §5.
4. File a P2 ticket with the recovery procedure + root-cause investigation.

The procedure violates the §4.3 "no force-push" rule, which is why **only the operator** can do it — the override is an explicit Constitution §13 exception with the override-audit JSON entry per [T12 §10b](../07_Testing/12_HelixQA_Autonomous.md#10b-override-audit-trail-format).

---

## 10. R-18 Considerations for Git Operations

Per Constitution §11.5 (R-18 Operational Integrity), git operations must not crash, suspend, or sign out the operator's host. Specific concerns:

- `git push` — pure network op, no host-disruptive effect.
- `git gc` — local CPU + disk; safe.
- `git fsck` — local read-only; safe.
- `git repack` — local rewrite; safe but can pause the host briefly under high IO; bounded.
- `git filter-branch` / `git filter-repo` — history-rewrite tools; **forbidden** without operator approval (Constitution §13). Such operations on a public repo are R-18-§11.5.6-incident-grade.

The `helix-r18-safeexec` deny-list (per [helix-r18-safeexec descriptor §2.4](../06_Submodules/per-submodule/helix-r18-safeexec.md#24-the-deny-list-private)) does not contain git commands because git itself is not host-disruptive; the discipline is at the policy level (Constitution + operator review).

---

## 10a. The .gitconfig Operator Setup

Each developer / operator host needs a one-time `.gitconfig` setup to use the composite-push semantics. The canonical setup script ships at `vasic-digital/.github/scripts/git-onboard.sh`:

```bash
#!/bin/bash
# Configure the four mirrors + composite origin.
git remote add github $GITHUB_URL 2>/dev/null || true
git remote add gitlab $GITLAB_URL 2>/dev/null || true
git remote add gitflic $GITFLIC_URL 2>/dev/null || true
git remote add gitverse $GITVERSE_URL 2>/dev/null || true

# Origin: fetch from GitHub.
git remote add origin $GITHUB_URL 2>/dev/null || true
git remote set-url origin $GITHUB_URL

# Origin: composite push to all four.
git remote set-url --push origin $GITFLIC_URL
git remote set-url --push --add origin $GITHUB_URL
git remote set-url --push --add origin $GITLAB_URL
git remote set-url --push --add origin $GITVERSE_URL

# Verify.
git remote -v
```

Output should show 4 push lines for `origin` + the four named remotes for fetch.

## 10b. The Branch-Protection Drift Audit

The `helix-branch-protection-audit` script runs nightly:

```bash
#!/bin/bash
# Audit branch protection across the four mirrors.
github_protection=$(gh api repos/HelixDevelopment/HelixPlay/branches/main/protection)
gitlab_protection=$(glab api projects/helixdevelopment1%2FHelixPlay/protected_branches/main)
gitflic_protection=$(curl -s -H "Authorization: token $GITFLIC_TOKEN" \
    https://gitflic.ru/api/p/v1/projects/helixdevelopment/helixplay/branches/main/protection)
# GitVerse: manual fallback per S01 §4.7.4

# Compare against canonical-protection.yaml + emit P3 alert on drift.
helix-protection-compare \
    --canonical=vasic-digital/.github/canonical-protection.yaml \
    --observed-github="$github_protection" \
    --observed-gitlab="$gitlab_protection" \
    --observed-gitflic="$gitflic_protection"
```

A drift indicates a soft spot in the gate; operator restores via `gh api` + `glab api` PATCH calls.

## 10c. The Tag-Signing Discipline (recommended)

Per OQ-O06-B, tag signing is recommended for release-train tags. The procedure:

```bash
# Configure signing key (operator-only).
git config user.signingkey "<key-fingerprint>"

# Sign a tag.
git tag -s v1.0.0 -m "helix-pipeline v1.0.0 release-train tag"

# Push to all four mirrors.
git push origin v1.0.0
```

GitHub + GitLab display the verified-signature badge; GitFlic + GitVerse store the signature but may not display the badge (per the Russian-mirror feature gap).

## 11. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-O06-A            | Branch-protection drift cadence — daily audit or per-PR check?                                                 | this chapter next revision                          |
| OQ-O06-B            | Tag-signing requirement — `git tag -s` mandatory for release-train tags?                                       | this chapter next revision                          |
| OQ-O06-C            | GitVerse policy-as-code gap — recheck cadence beyond [S01 §4.7.4 Q3 2026 review]?                             | Constitution §15                                    |

---

## 10d. The Push-Failure Triage Runbook

When `git push origin main` reports a partial-failure (some mirrors green, others red), the operator runs:

```bash
# Identify which mirror failed.
git push origin main 2>&1 | grep -E "^(To|error)"

# Retry the failed mirror.
git push <failed-mirror> main

# Verify all four are at the same SHA.
for r in github gitlab gitflic gitverse; do
    sha=$(git ls-remote $r main | head -1 | awk '{print $1}')
    echo "$r: $sha"
done
```

A persistent push-failure on a specific mirror (e.g. GitVerse repeatedly rejecting) triggers a P3 alert + investigation. Possible causes:

- Mirror's auth token expired (per [O01 §14b](01_Container_CI_CD.md#14b-per-mirror-registry-credential-rotation) registry rotation also covers git remotes).
- Mirror-side rate-limit hit (typically only for high-frequency churning maintainers).
- Mirror has an outage (rare; reflects in the four-mirror parity dashboard at [O04 §6](04_Observability_and_Events.md#6-grafana-dashboards)).

## 10e. The HelixPlay Repo-Root vs Submodule Repo Push Differences

| Property                       | Repo-root (this repo)                                     | Per-submodule repo                                          |
|--------------------------------|-----------------------------------------------------------|-------------------------------------------------------------|
| Origin path                    | `git@github.com:HelixDevelopment/HelixPlay.git`            | `git@github.com:vasic-digital/<name>.git`                   |
| Maintainers                    | Operator + (rare) approved maintainers                    | Per-submodule maintainers + operator                         |
| `main` branch protection       | ≥ 1 approval                                              | ≥ 1 (≥ 2 for `helix-r18-safeexec` per S01 §6.3)              |
| Tag protection                 | `v*` protected (rarely tagged at this level)              | `v*` protected (release-train cadence per S01 §9)            |
| Composite-push                 | yes (4 mirrors)                                            | yes (4 mirrors)                                              |
| Issue tracking                 | GitHub Projects + GitLab board (Operations)              | Per-submodule GitHub Project + GitLab board                  |

The two-tier topology is intentional — repo-root owns fleet-wide concerns; per-submodule repos own per-submodule concerns. A maintainer's PR against a submodule typically does not touch the repo-root.

## 12. References & Anti-Bluff Verification

### 12.1 Internal

- [`../01_Constitution.md`](../01_Constitution.md) §9.
- [`../06_Submodules/01_Submodule_Catalog.md`](../06_Submodules/01_Submodule_Catalog.md) §2, §4.7.
- [`../06_Submodules/02_Containers_Submodule.md`](../06_Submodules/02_Containers_Submodule.md) §7 (registry topology, mirror this for git).
- [`05_Tracking_GitHub_GitLab.md`](05_Tracking_GitHub_GitLab.md) — issue mirror.
- Repo-root `CLAUDE.md` — "Git topology" section.

### 12.2 External (web)

- Git documentation: https://git-scm.com/docs (accessed 2026-04-30).
- GitHub branch-protection: https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches (accessed 2026-04-30).
- GitLab protected branches: https://docs.gitlab.com/ee/user/project/protected_branches.html (accessed 2026-04-30).

### 12.3 Anti-Bluff Verification

- Coverage: chapter exceeds the 300-line floor.
- Forbidden patterns: clean. Three §11 open questions named with deferred resolution chapters.
- §3 composite-push pattern + §5 mirror-sync verification + §6 branch-protection rules operationalise Constitution §9 at the operator-CLI level.

### 12.4 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30.
- Reviewed by: pending operator review.

End of `08_Operations/06_Git_Topology_and_Push_Policy.md` — 2026-04-30.

---

## 🎉 Operations Family Closure

This is the **last chapter** of the Operations family. With O06 landed, all 7 chapters of `08_Operations/` (00_Index + O01..O06) are complete.

**Operations family aggregate**: 7 chapters totalling **≥ 2,250 lines** of operational specification.
