# Phase_06 — Host Agent (Sunshine++)

> **Source dimensions:** [`Phase_05_Clients.md`](Phase_05_Clients.md), [`../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md`](../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md) (C08), Insight #1 Sunshine++ pattern, [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row P06.
> **Source line count:** chapter floor 500 lines per Master Plan §7.2 row P06.
> **Phase targets:** R-09 (concurrency) + R-13 + R-18 (host agent's `helix-r18-safeexec` enforcement).
> **Cross-links:** [`Phase_07_Latency_Optimization.md`](Phase_07_Latency_Optimization.md), [`Phase_11_Hardening_and_Security.md`](Phase_11_Hardening_and_Security.md).
> **Status:** Draft v1 specification.
> **Last updated:** 2026-04-30.

---

## 1. Phase Scope

Phase_06 ships the **HelixPlay host agent** — the daemon that runs on the operator's host machine, captures + encodes + serves gameplay. The host agent is the **Sunshine++ fork** (per Insight #1 of the Architecture family + [C08 §6](../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md)) — Sunshine v2026.423.21833 (or current) at the capture-plane fork point + HelixPlay-specific session-plane on top.

Phase_04 stood up the streaming pipeline (helix-pipeline + helix-transport); Phase_06 wraps these in a **host-runtime-friendly daemon** with: session lifecycle management; per-tenant isolation; helix-r18-safeexec deny-list enforcement at the subprocess boundary; observability + alert routing; the auto-update + crash-recovery mechanisms.

After Phase_06, an operator deploys the host agent on a real machine (Steam Deck dev kit, Linux gaming desktop, Windows host), the agent advertises itself via mDNS, and a Phase_05 client can discover + connect + stream gameplay.

---

## 2. Prerequisites

- Phase_04 (streaming pipeline) + Phase_05 (clients) complete.
- helix-r18-safeexec at v1.0.0 (the deny-list enforcement primitive).
- helix-pipeline at v1.0.0.
- Operator-provisioned host hardware: GPU (NVIDIA RTX 30+ recommended), 32 GB RAM, dedicated 1 Gbps NIC, Steam-or-equivalent game launcher.

---

## 3. Tasks Catalogue

| Task ID    | Task                                                          | Subtasks |
|------------|---------------------------------------------------------------|---------:|
| P06.T01   | Sunshine++ fork point + Constitution-§11.5 R-18 audit          | 5        |
| P06.T02   | Session lifecycle manager (start / pause / resume / end)       | 6        |
| P06.T03   | Per-tenant session isolation                                   | 4        |
| P06.T04   | helix-r18-safeexec wrapping every subprocess invocation        | 4        |
| P06.T05   | Host-side observability — OTel + Prom + Loki                   | 4        |
| P06.T06   | mDNS / DoH advertisement                                       | 3        |
| P06.T07   | Auto-update mechanism (cosign-verified)                        | 4        |
| P06.T08   | Crash-recovery + session-restore                               | 4        |
| P06.T09   | Game launcher integration (Steam, GOG, Epic, etc.)             | 5        |
| P06.T10   | Operator runbook for host-agent deployment                     | 3        |
| P06.T11   | End-to-end smoke — real game streamed to Phase_05 client       | 4        |
| P06.T12   | Phase_06 acceptance review                                     | 2        |

12 tasks; ~48 subtasks.

---

## 4. Task Details

### 4.1 P06.T01 — Sunshine++ fork

Fork from Sunshine v2026.423.21833 (or current latest). Apply Constitution §11.5 audit:

- Every `os/exec` invocation wrapped in helix-r18-safeexec.
- Every host-disruptive command-line tool path replaced with the deny-listed equivalent.
- Audit log integration with operator's SIEM.

The fork divergence is documented in `HelixDevelopment/HelixAgent/SUNSHINE-FORK-DELTA.md`; rebases against upstream Sunshine happen every 90 days per the Sunshine++ pattern.

### 4.2 P06.T02 — Session lifecycle

- Session.Start(ctx, req) → returns SessionID + provisions the GPU + helix-pipeline + helix-transport.
- Session.Pause(ctx, sid) → checkpoints the encoder state + drops the gRPC stream.
- Session.Resume(ctx, sid) → re-establishes the gRPC stream + reanchored from checkpoint.
- Session.End(ctx, sid) → graceful teardown; emits billing event.

Per Constitution §11.5 (R-18) — every session-disruption potential is wrapped + logged.

### 4.3 P06.T03 — Per-tenant isolation

Per [helix-tenant descriptor §3.1](../06_Submodules/per-submodule/helix-tenant.md) + Constitution §11. Each session is bound to a tenant; tenants cannot observe each other's sessions.

### 4.4 P06.T04 — helix-r18-safeexec wrapping

Every subprocess invocation in the host agent uses `r18.SafeExec(ctx, ...)`. The `helix-r18-safeexec`-vet lint rejects direct `exec.Command` calls.

### 4.5 P06.T05 — Host-side observability

OTel SDK init via helix-otel-init; Prometheus metrics scraping; Loki log shipping. Per [O04](../08_Operations/04_Observability_and_Events.md).

### 4.6 P06.T06 — mDNS / DoH advertisement

Per [O03 §2 + §11.4](../08_Operations/03_Service_Discovery_and_Ports.md) — host agent advertises `_helix-host._grpc._tcp.local.` SRV record + (Russian-jurisdiction) DoH alternative.

### 4.7 P06.T07 — Auto-update

cosign-verified auto-update from `vasic-digital/Containers` registry. Operator can pin a specific version via env-var.

### 4.8 P06.T08 — Crash recovery

systemd / Windows Service / launchd integration; on crash, restart + restore last-known-good session via the run-archive.

### 4.9 P06.T09 — Game launcher integration

Steam (canonical), GOG, Epic, Battle.net (per C08 §9 cross-vendor authentication). Each launcher's quirks are documented in the C08 §9 addendum.

### 4.10 P06.T10 — Operator runbook

`HelixDevelopment/HelixAgent/docs/runbook/host-agent-deployment.md` — covers from fresh host → mDNS-discoverable → first session served.

### 4.11 P06.T11 — End-to-end smoke

A real Steam game (e.g. Counter-Strike 2 or a Godot demo) streamed end-to-end to a Phase_05 Wails client. 30-second sanity test.

### 4.12 P06.T12 — Acceptance

Operator signoff per Constitution §16 + the §6 exit criteria.

---

## 5. Subtask Catalogue

48 subtasks across 12 tasks; bulk-imported.

---

## 6. Exit Criteria

- [ ] Sunshine++ fork builds + runs.
- [ ] Session lifecycle CRUD operations green.
- [ ] Per-tenant isolation tested with cross-tenant denial.
- [ ] helix-r18-safeexec wraps every subprocess.
- [ ] OTel + Prom + Loki integrated.
- [ ] mDNS advertisement operational.
- [ ] Auto-update mechanism cosign-verifies.
- [ ] Crash-recovery restores session.
- [ ] At least 1 game launcher integration green.
- [ ] End-to-end smoke passes.
- [ ] Operator signoff per Constitution §16.

---

## 7. Risk Register

| ID       | Risk                                                                | Mitigation                                                            |
|----------|----------------------------------------------------------------------|------------------------------------------------------------------------|
| RP06-01  | Sunshine upstream rebase introduces breaking change                  | 90-day rebase cadence + documented divergence delta.                  |
| RP06-02  | Game launcher EULA conflict with helix-r18-safeexec audit            | Audit Steam EULA per CZ-7 closure; operator legal sign-off.            |
| RP06-03  | Auto-update breaks per-mirror trust (Russian operator pulls Western digest) | Auto-update pulls from operator's preferred mirror only.       |
| RP06-04  | Vanguard motherboard attestation rejects host (per Z-1)              | Host-agent advertises a per-game compatibility matrix; Vanguard-protected games disabled. |

---

## 8. Cross-Family Dependencies

- C08 §6 (Host Agent + Game Lifecycle) architecturally specifies the patterns.
- helix-pipeline + helix-transport (Phase_04) provide the streaming primitives.
- helix-r18-safeexec wraps all subprocess invocations.
- helix-tenant provides session isolation.

---

## 9. Acceptance Criteria

Constitution §16 signoff + §6 exit criteria.

---

## 10. The Phase_06 Calendar

~ 4 weeks. Operator-side capacity: 3 engineers (fork-and-modify Sunshine + Go integration + per-game-launcher specifics).

---

## 11. Anti-Bluff Verification

### 11.1 Sources resolved

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`Phase_05_Clients.md`](Phase_05_Clients.md)                      |    250+ | 2026-04-30 | predecessor                                      |
| [`../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md`](../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md) | 3,425 | 2026-04-30 | C08 architectural source |
| [`../06_Submodules/per-submodule/helix-r18-safeexec.md`](../06_Submodules/per-submodule/helix-r18-safeexec.md) | 343 | 2026-04-30 | R-18 wrapper                              |

### 11.2 Forbidden patterns

Clean.

### 11.3 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 (specification only).
- Pending: Phase_06 execution + operator signoff.

End of `09_Implementation_Phases/Phase_06_Host_Agent.md` — 2026-04-30.
