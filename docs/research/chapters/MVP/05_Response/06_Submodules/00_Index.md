# 06_Submodules/ — Family Index

> **Status:** Draft v1.
> **Last updated:** 2026-04-29.
> **Purpose:** Navigation hub for the *Submodules & Containers* family.
> **Targets (R-XX):** R-03 (decoupled reusable submodules under
> `vasic-digital`), R-12 (per-submodule Ten test types), R-17 (tracking),
> R-18 (Operational Integrity inheritance).
> **Cross-links:** [`../00_Master_Plan.md`](../00_Master_Plan.md)
> §7.2 row S01-S05; [`../01_Constitution.md`](../01_Constitution.md)
> §3 (Decoupling) + §6 (Ten test types) + §11.5 (R-18); the prior
> three families (Architecture C01-C13, Latency C14-C24, Video/Audio
> C25-C37) which name every submodule reused or introduced here.

---

## 1. Position in the synthesis programme

The Submodules family is the **fourth of nine** families in the
`05_Response/` synthesis. The first three are *content* families
(Architecture, Latency, Video/Audio); they introduce 24+ public
submodules under the `vasic-digital` GitHub/GitLab organisation as
they go, with each chapter §6 defining the submodule's API surface,
imports, and R-18 integration. **This family is the *aggregation*
layer**: it resolves duplication risk (R-04), states the dependency
graph, fixes the Ten-test-type matrix per submodule, and ratifies
how the three pre-existing organisational repos (`vasic-digital/
Containers`, `vasic-digital/Challenges`, `HelixDevelopment/HelixQA`)
plug into the submodule fleet.

A submodule is **public** if it appears under `vasic-digital/`. It
is **private** only if it embeds proprietary logic that cannot ship
as open source — none of the 24 submodules introduced so far is
private. Public visibility is required by R-03; the per-submodule
descriptor §05_Submodule_Descriptors confirms each submodule's
licence (MIT or Apache 2.0 default), CI lane (per `vasic-digital/
Containers`), test matrix (Ten types), and dependency list.

The family does **not** re-introduce the submodules. Each one is
already named, scoped, and has its API surface defined in its
originating chapter (e.g. `helix-r18-safeexec` in C08 §10;
`helix-shm` in C15 §6; `helix-codec` in C26 §6). What the family
adds is:

1. A **catalog** that lists all 24+ submodules with their origin
   chapter, public path, dependency list, and R-12/R-13/R-18
   compliance status (chapter §01).
2. The **Containers integration** — how every submodule consumes
   `vasic-digital/Containers` for build, test, and CI containers
   (chapter §02).
3. The **Challenges integration** — how every submodule's §8.10
   *Challenges* test row hooks into `vasic-digital/Challenges`
   for production-like full-system runs (chapter §03).
4. The **HelixQA autonomous integration** — how every submodule's
   regression run feeds into `HelixDevelopment/HelixQA`'s
   nightly + canary + per-PR Challenge orchestration
   (chapter §04).
5. **Per-submodule descriptors** — one short file per submodule
   under `per-submodule/<name>.md` documenting that submodule's
   exact dependency list, CI lane, test matrix entries, and
   public path (chapter §05_Submodule_Descriptors).

---

## 2. Family chapter list

| ID  | Chapter                                                                                                              | Floor (lines) | Status     |
|-----|----------------------------------------------------------------------------------------------------------------------|---------------|------------|
| —   | [`00_Index.md`](00_Index.md) — this file                                                                              |     navigation| **draft**  |
| S01 | [`01_Submodule_Catalog.md`](01_Submodule_Catalog.md) — all 24+ submodules + dependency graph                          | 800           | pending    |
| S02 | [`02_Containers_Submodule.md`](02_Containers_Submodule.md) — Containers repo integration                             | 400           | pending    |
| S03 | [`03_Challenges_Submodule.md`](03_Challenges_Submodule.md) — Challenges repo integration                             | 400           | pending    |
| S04 | [`04_HelixQA_Integration.md`](04_HelixQA_Integration.md) — autonomous QA orchestration                              | 400           | pending    |
| S05 | `per-submodule/<name>.md` — one descriptor per public submodule under `vasic-digital`                                | 300 ea        | pending    |

S01 is the canonical lookup — every other chapter in the programme
that names a submodule should resolve to a row in S01's catalog.
S02-S04 cover the three pre-existing organisational repos; the
authority chain is: per-submodule rule → S05 descriptor → S02-S04
integration → S01 catalog row.

---

## 3. Submodule inventory (provisional snapshot, finalised in S01)

24 submodules introduced by chapters C01-C37 to date. The S01
*Submodule Catalog* will canonicalise the list, dedupe any
overlapping scope flagged by R-04 (RK08 in Master Plan §8 risk
register), and freeze the dependency graph.

Architecture-family submodules (C01-C13):
- `helix-r18-safeexec` — origin C08 §10. **Most-reused submodule
  in the programme** (every later chapter imports it for subprocess
  R-18 wrapping; its deny-list is non-overridable).
- `helix-grpc-frame` — origin C06 §6 (real-time gRPC streaming).
- `helix-tv-input` — origin C12 §6 (TV remote / D-pad input).
- `helix-vault` — origin C10 §6 (Vault wrapper for KEK/DEK + GDPR).
- `helix-tenant` — origin C11 §6 (white-label theming).

Latency-family submodules (C14-C24):
- `helix-shm` — origin C15 §6 (shared-memory NV12/I420 pages).
- `helix-iouring` — origin C16 §4 (io_uring async I/O).
- `helix-xdp` — origin C16 §6 (XDP eBPF kernel-bypass).
- `helix-lockfree` — origin C17 §3 (SPSC ringbuffer + MPSC queue).
- `helix-gpu-direct` — origin C18 §3 (GPUDirect RDMA + p2p PCIe).
- `helix-network` — origin C19 §6 (DSCP/L4S/jitter buffer).
- `helix-rtos` — origin C20 §3 (SCHED_FIFO + cgroup pinning).
- `helix-input` — origin C21 §6 (controller polling + Reflex).
- `helix-display` — origin C22 §6 (frame pacing + VRR).
- `helix-mempool` — origin C23 §3 (mempool + arena allocator).
- `helix-allocator` — origin C23 §6 (alloc-free hot path enforcer).
- `helix-bench` — origin C24 §6 (benchmark harness; ≥10K samples).

Video/Audio-family submodules (C25-C37):
- `helix-codec` — origin C26 §6 (codec ladder + GOP cadence).
- `helix-encoder` — origin C27 §6 (vendor SDK wrappers).
- `helix-capture` — origin C28 §6 (DXGI/Metal/X11/PipeWire capture).
- `helix-dualpath` — origin C29 §6 (dual-rung NAL feed).
- `helix-record` — origin C30 §6 (fMP4 + MKV recording + S3/SMB/NFS sync).
- `helix-audio` — origin C31 §6 (Opus MultiStream + eARC + ALLM).
- `helix-hdr` — origin C32 §6 (PQ/HLG + HDR10/10+/Dolby Vision + tone-map).
- `helix-abr` — origin C33 §6 (8-tier ladder + GCC/SCReAM/SQP).
- `helix-thermal` — origin C34 §6 (NVML/ADL/Level Zero + DVFS).
- `helix-vqa` — origin C35 §6 (VMAF + LDAT + change-point).
- `helix-pipeline` — origin C36 §8 (goroutine topology + cgo).
- `helix-transport` — origin C37 §9 (RTP/SRTP/ICE/QUIC + io_uring/XDP send).

R-04 duplication watch: S01 §3 will scan `vasic-digital` org for
existing repos overlapping any new name (`helix-shm` vs an existing
`shm-go`, etc.); none of the 27 names listed above currently
collides with an existing repo, but the catalog re-runs the check
before commit per RK08 in Master Plan §8.

---

## 4. R-18 inheritance ladder (canonical)

Every submodule that wraps a subprocess imports `r18.SafeExec` from
`vasic-digital/helix-r18-safeexec` (origin C08 §10). The deny-list
is **non-overridable** per Constitution §11.5.4 — submodules MAY
extend the allow-list with chapter-specific tooling but MUST NOT
duplicate or relax the deny-list.

The inheritance ladder, observed across every chapter §6 in the
programme:

```
helix-r18-safeexec (C08 §10 — origin)
  │
  ├── helix-shm + helix-iouring + helix-xdp + helix-lockfree (Latency)
  ├── helix-gpu-direct + helix-network + helix-rtos (Latency)
  ├── helix-input + helix-display + helix-mempool + helix-bench (Latency)
  ├── helix-codec + helix-encoder + helix-capture (Video/Audio)
  ├── helix-dualpath + helix-record + helix-audio + helix-hdr (Video/Audio)
  └── helix-abr + helix-thermal + helix-vqa + helix-pipeline + helix-transport (Video/Audio)
```

Tests inherit `host-integrity-scan` from C08 §12.11 verbatim
(Constitution §11.5.4 non-overridable). Every submodule's §8.11
test row references this inheritance instead of redefining the scan.

---

## 5. Ten test types per submodule (R-12)

Per Constitution §6, every submodule supports the Ten test types:
Unit / Integration / E2E / Security / Benchmarking / Chaos / Stress
/ Smoke / Full automation / Challenges. Only Unit may use mocks /
stubs / hardcoded values; the other nine MUST hit real systems.

S01 §5 will tabulate per-submodule which test rows are owned
in-tree vs which delegate to a sibling submodule (e.g. `helix-shm`
delegates Challenges to `helix-pipeline`). S02 documents the
container that runs the tests; S03 documents the Challenges full-
system topologies; S04 documents the autonomous orchestration.

---

## 6. Family-level cross-links

- **Architecture family**: [`../03_Architecture/`](../03_Architecture/)
  — C08 §10 originates `r18.SafeExec`; C10 §6 originates `helix-vault`.
- **Latency family**: [`../04_Latency/`](../04_Latency/) — every
  chapter §6 originates 1-2 submodules.
- **Video/Audio family**: [`../05_Video_Audio/`](../05_Video_Audio/)
  — every chapter §6 (or §8 for C36 / §9 for C37) originates
  1-2 submodules.
- **Testing family**: [`../07_Testing/`](../07_Testing/) — pulls
  the per-submodule §8 test surfaces into a single test matrix
  (T01 §1 indexed against S01).
- **Operations family**: [`../08_Operations/`](../08_Operations/)
  — pulls the per-submodule CI lanes into a single container CI/CD
  pipeline (O01 §1 indexed against S01 + S02).
- **Implementation Phases**: [`../09_Implementation_Phases/`](../09_Implementation_Phases/)
  — Phase_02 (Core Submodules) deliverables are exactly the rows
  in S01 §3, ordered by dependency depth.

---

## 7. Anti-Bluff Verification (family-level)

Per [Master Plan §4.3](../00_Master_Plan.md#43-anti-bluff-verification-block-r-13).

### Sources resolved in this index

| Path                                                     | Lines  | Role               |
|----------------------------------------------------------|-------:|--------------------|
| [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2       | 1k+    | family chapter list (S01-S05) |
| [`../01_Constitution.md`](../01_Constitution.md) §3, §6, §11.5 | 1k+ | R-03 + R-12 + R-18 |
| 13× Architecture chapters (C01-C13) §6                    |    —   | submodule origins  |
| 11× Latency chapters (C14-C24) §6                         |    —   | submodule origins  |
| 13× Video/Audio chapters (C25-C37) §6 / §8 / §9           |    —   | submodule origins  |

### Forbidden patterns

Index prose scanned: clean. No TODO/FIXME/placeholder/XXX/stub.
The family-level "S05 per-submodule/<name>.md (×N)" entry in §2
is a structural placeholder for a directory of files (one per
submodule), not a TODO marker; S01 §3 freezes the N value at
commit time.

### Sign-off

- Index drafted by orchestrator (Claude) on 2026-04-29.
- Pending: S01-S05 chapters, each with their own Anti-Bluff block.
- Reviewed by: pending operator review.

End of `06_Submodules/00_Index.md` — 2026-04-29.
