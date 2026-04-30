# O05 — Tracking on GitHub Projects + GitLab (R-17)

> **Source dimensions:** [`00_Index.md`](00_Index.md); [`../00_Master_Plan.md`](../00_Master_Plan.md) §6 (R-17 mirroring); [`../01_Constitution.md`](../01_Constitution.md) §8 (R-16/R-17); [`../06_Submodules/04_HelixQA_Integration.md`](../06_Submodules/04_HelixQA_Integration.md) §10.5; [`../07_Testing/12_HelixQA_Autonomous.md`](../07_Testing/12_HelixQA_Autonomous.md) §6 + §9b.
> **Source line count:** chapter floor 300 lines per Master Plan §7.2 row O05.
> **Chapter targets:** R-16 (fine-grained phases / tasks / subtasks), R-17 (mirrored to GitHub Projects + GitLab via `gh` + `glab`).
> **Cross-links:** [`02_Quality_Gates_SonarQube_Snyk.md`](02_Quality_Gates_SonarQube_Snyk.md), [`04_Observability_and_Events.md`](04_Observability_and_Events.md), [`06_Git_Topology_and_Push_Policy.md`](06_Git_Topology_and_Push_Policy.md).
> **Status:** Draft v1.
> **Last updated:** 2026-04-30.

---

## 1. The Mandate in Three Sentences

R-16 + R-17 (Constitution §8) require **every phase / task / subtask** to be mirrored on **both** GitHub Projects (`HelixDevelopment/HelixPlay`) and the GitLab equivalent (`gitlab.com/helixdevelopment1/HelixPlay`). The CLI tooling is `gh` (GitHub) + `glab` (GitLab). This chapter operationalises the W07 task from Master Plan §7.2 — the actual `gh project item-create` + `glab project issue create` invocations that bring the documented architecture into operational use.

---

## 2. The Naming Convention

Per [Master Plan §6.1](../00_Master_Plan.md#61-naming-convention):

- Phase: `[P00] Foundation`, `[P01] Containers & CI`, …, `[P13] GA`.
- Task: `[P03.T05] Implement Pion WebRTC adapter`.
- Subtask: `[P03.T05.S02] Add ICE-Lite endpoint`.

Every issue title begins with the bracketed identifier; this is mandatory for the cross-platform grep + the helix-tracking-audit lint.

---

## 3. Cross-Linking Issues + Tickets

Every GitHub issue body links to its GitLab counterpart and vice versa per [Master Plan §6.2](../00_Master_Plan.md#62-cross-linking). The link block format:

```markdown
## Cross-references

- GitHub: https://github.com/HelixDevelopment/HelixPlay/issues/1234
- GitLab: https://gitlab.com/helixdevelopment1/HelixPlay/-/issues/567
- HelixQA run-archive: helixqa://run/2026-04-30T02:00:00Z (if applicable)
- PR (or MR): https://github.com/HelixDevelopment/HelixPlay/pull/2345
```

The cross-reference block is parsed by the `helix-tracking-audit` lint nightly + flags any unidirectional link (e.g. GitHub→GitLab present but GitLab→GitHub missing).

---

## 4. The `gh` Invocation Pattern

```bash
# Create a phase top-level item.
gh project item-create $PROJECT_ID \
    --owner HelixDevelopment \
    --title "[P03] Backend Services" \
    --body "Phase 03 — Backend Services. See 09_Implementation_Phases/Phase_03_Backend_Services.md for scope."

# Create a task under the phase.
gh issue create \
    --repo HelixDevelopment/HelixPlay \
    --title "[P03.T05] Implement Pion WebRTC adapter" \
    --body "$(cat tasks/P03.T05.md)" \
    --label "phase-03,task,priority-medium"

# Add the issue to the project board.
gh project item-add $PROJECT_ID \
    --owner HelixDevelopment \
    --url https://github.com/HelixDevelopment/HelixPlay/issues/1234

# Update field (status, priority, assignee).
gh project item-edit $ITEM_ID --field=status=InProgress
gh project item-edit $ITEM_ID --field=assignee=@username

# Close on resolution.
gh issue close 1234 --comment "Closed in PR #2345 (commit abc123). Cross-closed at https://gitlab.com/helixdevelopment1/HelixPlay/-/issues/567"
```

The `$PROJECT_ID` is the GitHub Projects v2 numeric ID — operator-provisioned + recorded in `HelixDevelopment/HelixQA/config/project-ids.yaml`.

---

## 5. The `glab` Invocation Pattern

```bash
# Create the phase issue.
glab issue create \
    --repo helixdevelopment1/HelixPlay \
    --title "[P03] Backend Services" \
    --description "$(cat phases/P03.md)" \
    --label "phase-03"

# Create a task.
glab issue create \
    --repo helixdevelopment1/HelixPlay \
    --title "[P03.T05] Implement Pion WebRTC adapter" \
    --description "$(cat tasks/P03.T05.md)" \
    --label "phase-03,task,priority-medium"

# Add to a board.
glab board move --issue=567 --column="In Progress"

# Close on resolution.
glab issue close 567 --message "Closed in MR !!2345 (commit abc123). Cross-closed at https://github.com/HelixDevelopment/HelixPlay/issues/1234"
```

---

## 6. The Bidirectional Mirror Webhook

A webhook keeps GitHub + GitLab in sync without operator intervention:

1. Operator opens an issue on GitHub via `gh issue create`.
2. GitHub webhook fires → HelixQA's `tracking-mirror` service.
3. Service issues `glab issue create` with the same title + body + cross-reference link.
4. Service updates the GitHub issue's body to add the GitLab link.

The mirror is **bidirectional** — closing the GitHub issue closes the GitLab one + vice versa. Webhook payloads + the mirror's idempotency key prevent infinite-loop fan-out.

The service is shipped as a container under `HelixDevelopment/HelixQA/cmd/tracking-mirror/`; deployed alongside the rest of HelixQA's control-plane.

---

## 7. The Required Labels

Every issue carries a fixed label set:

| Label                     | Values                                                  |
|---------------------------|---------------------------------------------------------|
| Phase                     | `phase-00` .. `phase-13`                                |
| Type                      | `phase`, `task`, `subtask`, `bug`, `tech-debt`, `chore` |
| Priority                  | `priority-low`, `priority-medium`, `priority-high`, `priority-critical` |
| Status                    | `triage`, `in-progress`, `review`, `done`, `wontfix`    |
| Severity (alerts)         | `severity-p1`, `severity-p2`, `severity-p3`             |
| HelixQA cross-ref         | `helixqa-run-id:<run-id>` (if applicable)               |
| Submodule                 | `submodule:<name>` (if applicable)                       |

The `helix-tracking-audit` lint verifies every issue has the minimum label set (Phase + Type + Status); missing labels are flagged.

---

## 8. The Project-Board Layout

GitHub Projects v2 board shape:

| Column         | Auto-rule                                                   |
|----------------|-------------------------------------------------------------|
| Triage          | new issue without status label                              |
| Backlog         | status=triage moved manually                                |
| Ready           | depends on resolved issues                                  |
| In Progress     | status=in-progress                                          |
| Review          | PR linked + status=review                                   |
| Done            | issue closed + linked PR merged                             |
| Wontfix         | status=wontfix                                              |

GitLab equivalent at `helixdevelopment1/HelixPlay/-/boards`. The columns are operator-configured; rules + automation are part of the project provisioning runbook at `HelixDevelopment/HelixQA/docs/runbook/project-board-provisioning.md`.

---

## 9. The Operator's Bulk-Issue Creation Workflow

For operationalising the W07 task at scale (every phase × every task × every subtask), operator uses a CSV import:

```csv
title,body,labels
[P00] Foundation,See 09_Implementation_Phases/Phase_00_Foundation.md,phase-00;phase
[P00.T01] Containers v1.0.0 graduation,...,phase-00;task;priority-high
[P00.T01.S01] vasic-digital/Containers ≥ v1.0.0,...,phase-00;subtask;priority-medium
[P00.T01.S02] Per-mirror registry credential rotation,...,phase-00;subtask;priority-medium
...
```

The `helix-bulk-issue-create` tool (shipped in `HelixDevelopment/HelixQA/cmd/`) reads the CSV + creates issues + adds them to the project board + cross-mirrors to GitLab in one invocation. Idempotent via the `[Pxx.Tyy.Szz]` identifier prefix.

---

## 9a. The HelixQA-Created Tracking Tickets

Per [T12 §6 + §9b](../07_Testing/12_HelixQA_Autonomous.md#6-tracking-mirror-to-github-projects--gitlab-r-17), HelixQA auto-creates tickets on:

- **P1 alert** → severity-p1 + helixqa-run-id label.
- **P2 alert** → severity-p2.
- **P3 alert** → severity-p3.
- **Deployment-gate VerdictBlock** → gate-block + image-digest label.
- **Operator-override** → override + reason label.

The auto-created tickets include the OTLP trace URL + run-archive entry URL for postmortem. Operator closes the ticket once remediated; auto-close webhook keeps GitHub + GitLab in sync.

## 9b. Cross-Mirror Issue-ID Mapping

Since GitHub and GitLab issue numbers are independent, the cross-reference label `[helixqa-run-id:<run-id>]` is the **canonical join key**. An operator searching for a specific run can grep both platforms via the run-id and find the same logical issue regardless of platform-specific number.

The mapping is stored in HelixQA's tracking-mirror service's database for fast lookup (`helixqa-cli ticket-find <run-id>`).

## 9c. The Operator's Bulk-Provisioning One-Time Workflow

For the W07 task — provisioning every of the 14 phases × N tasks × M subtasks — operator runs the bulk-provisioning script once at Phase_00:

```bash
./vasic-digital/.github/scripts/provision-tracking.sh \
    --phases-source=09_Implementation_Phases/ \
    --github-project-owner=HelixDevelopment \
    --github-project-id=$PROJECT_ID \
    --gitlab-project=helixdevelopment1/HelixPlay
```

The script reads every `Phase_NN_*.md` file's tasks/subtasks (encoded in a fixed YAML front-matter format) + creates the corresponding GitHub Project items + GitLab issues. Idempotent via the `[Pxx.Tyy.Szz]` identifier prefix.

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-O05-A            | GitHub Projects v2 vs old Projects API — which API does `helix-tracking-audit` rely on?                      | this chapter next revision                          |
| OQ-O05-B            | Webhook idempotency-key cadence — operator-tunable per high-volume cadence?                                  | this chapter next revision                          |
| OQ-O05-C            | GitFlic + GitVerse ticket mirroring — currently no equivalent of GitHub Projects; document deferment?         | Constitution §15 *Amendment Procedure*              |

---

## 9d. The W07 Task Status

Master Plan §7.2 row W07: "Mirror to GitHub Projects + GitLab".

| Sub-task                                    | Status                                                          |
|---------------------------------------------|------------------------------------------------------------------|
| Document the architecture                    | ✅ done in [S04 §10.5](../06_Submodules/04_HelixQA_Integration.md#105-tracking-on-github-projects--gitlab-r-17) |
| Document the operator playbook              | ✅ done in [T12 §6 + §9b](../07_Testing/12_HelixQA_Autonomous.md#6-tracking-mirror-to-github-projects--gitlab-r-17) |
| Document the gh + glab command catalogue    | ✅ done in this chapter §4 + §5                                  |
| Provision the GitHub Project + GitLab board | **⏸️ deferred** to Phase_00 of the Implementation Phases family |
| Bulk-import every of the 14 phases × tasks  | **⏸️ deferred** to Phase_00                                       |
| Wire up the bidirectional webhook           | **⏸️ deferred** to Phase_00                                       |

The four ⏸️-deferred items are operator-side provisioning that can only happen once the Implementation Phases family lands + the operator decides to start Phase_00. This chapter is the operational specification; Phase_00 is the execution.

## 9e. The Per-Submodule Issue Tracking

Each of the 29 `vasic-digital/` submodule repositories has its own GitHub Project + GitLab board for per-submodule issues (bug reports, feature requests, dependency-bump PRs from Renovate). These are **separate** from the Phase × Task × Subtask boards above; the org-level Operations board references the per-submodule boards via cross-link.

The per-submodule boards are auto-provisioned by the submodule-bootstrap script per [O06 §7](06_Git_Topology_and_Push_Policy.md#7-submodule-repository-topology). New issues + PRs land on the per-submodule board first; if they relate to a Phase × Task, operator manually links them.

## 11. References & Anti-Bluff Verification

### 11.1 Internal

- [`../00_Master_Plan.md`](../00_Master_Plan.md) §6.
- [`../06_Submodules/04_HelixQA_Integration.md`](../06_Submodules/04_HelixQA_Integration.md) §10.5.
- [`../07_Testing/12_HelixQA_Autonomous.md`](../07_Testing/12_HelixQA_Autonomous.md) §6, §9b.
- [`../01_Constitution.md`](../01_Constitution.md) §8.

### 11.2 External (web)

- gh CLI: https://cli.github.com/manual/ (accessed 2026-04-30).
- glab CLI: https://gitlab.com/gitlab-org/cli (accessed 2026-04-30).
- GitHub Projects v2 API: https://docs.github.com/en/graphql/reference/objects#projectv2 (accessed 2026-04-30).
- GitLab Issues API: https://docs.gitlab.com/ee/api/issues.html (accessed 2026-04-30).

### 11.3 Anti-Bluff Verification

- Coverage: chapter exceeds the 300-line floor.
- Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.
- §4 + §5 command catalogues operationalise R-17 at the actual-CLI level (vs the architectural "should be mirrored" claim in S04 §10.5).
- The W07 task is **documented**; **execution** of W07 (creating the actual board + bulk-importing the issues) lands in Phase_00 of the Implementation Phases family.

### 11.4 The Per-Mirror Tracking Drift Audit

Beyond the bidirectional webhook (§6), nightly `helix-tracking-drift-audit` reconciles GitHub Projects vs GitLab boards:

```bash
# For each open issue on GitHub:
gh project item-list $PROJECT_ID --format=json | jq -r '.[] | select(.status != "Done") | .id' | \
while read item_id; do
    # Find matching GitLab issue by `[Pxx.Tyy.Szz]` prefix.
    title=$(gh project item-show $item_id --format=json | jq -r '.title')
    prefix=$(echo "$title" | grep -oE '\[P[0-9]{2}(\.T[0-9]{2}(\.S[0-9]{2})?)?\]')
    gitlab_match=$(glab issue list --search "$prefix" --output=json | jq -r '.[].iid // empty')
    if [ -z "$gitlab_match" ]; then
        echo "DRIFT: GitHub item $item_id ($prefix) has no GitLab counterpart"
    fi
done
```

A drift opens a P3 ticket. Operator either creates the missing counterpart or closes the orphaned one.

### 11.5 The Operator's Tracking Telemetry

| Metric                                          | Type    | Description                                            |
|-------------------------------------------------|---------|--------------------------------------------------------|
| `helix_tracking_issues_open`                    | gauge   | Open issues on each platform.                          |
| `helix_tracking_drift_total`                    | counter | Cumulative drift events caught by the audit.           |
| `helix_tracking_webhook_failures_total`         | counter | Bidirectional webhook delivery failures.               |
| `helix_tracking_severity_distribution`          | gauge   | Open issues bucketed by severity label.                |

Dashboards at [O04 §6](04_Observability_and_Events.md#6-grafana-dashboards) include a "Tracking Health" panel surfacing these.

### 11.6 The R-17 Operational Definition of Done

R-17's operational completion is reached when:

- [ ] All 14 phases × N tasks × M subtasks have GitHub Project items + GitLab issues.
- [ ] Every cross-reference link is bidirectional + verified by the nightly drift audit.
- [ ] The bidirectional webhook is operational + has < 1 % delivery-failure rate.
- [ ] HelixQA-auto-created tickets (per §9a) are flowing without operator intervention.
- [ ] Every auto-created ticket has the canonical label set (per §7).
- [ ] The "Tracking Health" dashboard panel (per §11.5) is green for ≥ 7 consecutive days.

The above checklist is the W07 Definition of Done at the operational level.

### 11.7 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30.
- Reviewed by: pending operator review.

End of `08_Operations/05_Tracking_GitHub_GitLab.md` — 2026-04-30.
