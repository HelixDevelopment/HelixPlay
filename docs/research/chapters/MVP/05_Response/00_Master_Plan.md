# HelixPlay MVP — Ultimate Documentation Master Plan

> **Slogan:** "Ultimate gaming experience!"
>
> **Status:** Living document. Session 1 (foundation) committed. Subsequent
> sessions will fill the chapter directories per the queue at the bottom of
> this file.
>
> **For agentic workers:** This plan governs the synthesis effort that turns
> the three MVP research streams (`01_base`, `02_latency`, `03_video_technology`)
> into a single, executable, fully-specified implementation programme stored
> under `docs/research/chapters/MVP/05_Response/`. Use
> `superpowers:subagent-driven-development` to execute the chapter tasks task-by-task.
> Use `superpowers:executing-plans` for sequential inline execution.

---

## 1. Contract — what `04_Request.md` Demands

The authoritative MVP brief at `docs/research/chapters/MVP/04_Request.md`
imposes a contract that every artifact under `05_Response/` must satisfy.
The clauses below are reproduced verbatim in the project Constitution
(`05_Response/01_Constitution.md`) and are referenced everywhere else by
their short identifier (`R-NN`).

| ID    | Clause                                                                                                                                                                                                                            |
|-------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| R-01  | The 05_Response synthesis MUST contain at least as much material as `01_base`+`02_latency`+`03_video_technology` combined, and MUST extend (never simplify) every point.                                                       |
| R-02  | No simplification, no bluffing, no skipping. No `TODO`/`FIXME` placeholders, no dummy/placeholder classes, no dead code, nothing hanging.                                                                                       |
| R-03  | Everything fully decoupled and reusable; reusable components live in **public** Git/Go submodules under the `vasic-digital` GitHub & GitLab organisations.                                                                       |
| R-04  | Reuse existing `vasic-digital` submodules; if features are missing, **extend** the submodule. Do not duplicate.                                                                                                                  |
| R-05  | All container work flows through `https://github.com/vasic-digital/Containers`.                                                                                                                                                   |
| R-06  | Every Service, infra component (DBs included), build, test, scan runs **inside containers**. Local-only CI/CD inside containers.                                                                                                  |
| R-07  | Service discovery on the LAN. Dynamic port assignment. gRPC preferred. REST is a separate microservice. HTTP/3 (QUIC/Cronet). Brotli compression.                                                                                |
| R-08  | NATS / Redis / RabbitMQ used wherever they replace ad-hoc plumbing. Heavy use of events + observability for real-time propagation.                                                                                                 |
| R-09  | Concurrency: non-blocking by default; lazy initialization preferred over eager; semaphores/backpressure to prevent clogging.                                                                                                       |
| R-10  | Heavy quality/security scanning: SonarQube, Snyk, plus supplemental scanners.                                                                                                                                                      |
| R-11  | Every submodule and every file of code MUST be covered 100% with tests of ten types: Unit, Integration, E2E, Security, Benchmarking, Chaos, Stress, Smoke, Full Automation, **Challenges**.                                        |
| R-12  | Only **Unit** tests may use mocks/stubs/hardcoded values. Every other test type must drive a real, production-like system with all containers running.                                                                            |
| R-13  | Tests must execute in **anti-bluff** mode — green tests must guarantee real, end-user-usable behaviour. Past incidents had green tests on broken features; this MUST NOT recur.                                                  |
| R-14  | The `Challenges` discipline (`git@github.com:vasic-digital/Challenges.git`) is integrated as in HelixAgent and Catalogizer. The autonomous QA system `git@github.com:HelixDevelopment/HelixQA.git` is fully integrated.            |
| R-15  | Every submodule pulls in **all of its own** dependency submodules and propagates the Constitution into its `CLAUDE.md` / `AGENTS.md`.                                                                                              |
| R-16  | Implementation is split into fine-grained **phases → tasks → subtasks**, every detail captured. Final state: a clean board, nothing skipped, simplified, omitted, or disabled.                                                    |
| R-17  | Every phase/task/subtask is mirrored on **GitHub Projects** AND its **GitLab equivalent** through their CLIs (`gh`, `glab`).                                                                                                       |
| R-18  | **Operational Integrity** — no command, hook, container entrypoint, CI lane, or agent prompt may suspend, hibernate, lock, terminate, or crash the operator's active development host. Forbidden-command list and container hazards documented in Constitution §11.5; enforced by the `host-integrity-scan` CI sub-lane. Added 2026-04-28 after a Session-2/3 incident.                                                                              |

These clauses are **non-negotiable** and must appear, verbatim or by reference,
in every submodule's Constitution / CLAUDE.md / AGENTS.md. See
`05_Response/01_Constitution.md` and `05_Response/06_Submodules/`.

---

## 2. Source Inventory

The synthesis must absorb and extend every byte of the existing MVP research
streams. Below is the full inventory with line counts established at session
start (2026-04-28).

### 2.1 Stream 1 — `01_base/` (Cloud Gaming System Architecture)

**Authoritative request:** `01_base/01_Request.md` (overall system).
**Per-stream final:** `01_base/02_response/cloudgaming.agent.final/cloudgaming.agent.final.md` (2,817 lines, plus diagrams).

12 dimensions (`Research/research/cloudgaming_dim01..12.md`):

| Dim | Title                                                       | Lines |
|-----|-------------------------------------------------------------|------:|
| 01  | Low-Latency Video Streaming Protocols & Codecs              |   812 |
| 02  | Cross-Platform Controller Input Capture & Forwarding        |   557 |
| 03  | Host OS Game Capture Technologies                           |   909 |
| 04  | Go Ecosystem for Cross-Platform Client Development          | 1,380 |
| 05  | Real-Time Communication APIs in Go                          |   966 |
| 06  | Game Catalog, Metadata & 4K Asset Management                | 1,431 |
| 07  | Host Agent Architecture & Game Lifecycle Management         | 1,449 |
| 08  | Scalability, Load Balancing & Multi-Region Infrastructure   | 1,003 |
| 09  | Security, Authentication & Host Isolation                   |   974 |
| 10  | White-Label, Theming & Customization Architecture           | 1,353 |
| 11  | TV-First UI/UX & Living Room Experience                     | 1,155 |
| 12  | Performance Optimization & End-to-End Latency Engineering   | 1,340 |
|     | `cloudgaming_dim_decomposition.md`                          |    70 |
|     | `cloudgaming_insight.md` (8 cross-dim insights)             |   156 |
|     | `cloudgaming_cross_verification.md` (10 HC, 5 MC, 5 CZ)     |   130 |

**Plus** all diagrams under `cloudgaming.agent.final/` (PNG): C4 context, C4
container, theme pipeline, token architecture, protocol abstraction, frame
pacing pipeline, infrastructure, load balancing, latency budget figure 12.1,
encoder bandwidth figure 12.2, gantt timeline.

### 2.2 Stream 2 — `02_latency/` (Zero-Latency Communication)

**Authoritative request:** `02_latency/01_Request.md`.
**Per-stream final:** `02_latency/02_Response/ZeroLatency_Communication_CloudGaming.md` (2,199 lines).

10 dimensions (`Agent_results/research/latency_dim01..10.md`):

| Dim | Title                                                | Lines |
|-----|------------------------------------------------------|------:|
| 01  | Shared Memory & Zero-Copy IPC                        |   126 |
| 02  | io_uring & Kernel Bypass I/O                         |   117 |
| 03  | Lock-Free Data Structures & Algorithms               |   108 |
| 04  | GPU Direct & Hardware Accelerated Pipelines          |   136 |
| 05  | Ultra-Low-Latency Network Protocols                  |   103 |
| 06  | Real-Time OS & Scheduling                            |   129 |
| 07  | Controller Input Optimization                        |   118 |
| 08  | Frame Pacing & Synchronization                       |    91 |
| 09  | Memory & Cache Optimization                          |    92 |
| 10  | Testing, Benchmarking & Validation Frameworks        |   128 |
|     | `latency_dim_decomposition.md`                       |    33 |
|     | `latency_insight.md` (5 cross-dim insights)          |   100 |
|     | `latency_cross_verification.md` (10 HC, 4 CZ)        |   101 |

### 2.3 Stream 3 — `03_video_technology/` (Capture, Codec, Encode, Audio)

**Authoritative request:** `03_video_technology/Request.md` (note: no `01_` prefix — preserved verbatim).
**Per-stream final:** `03_video_technology/02_Response/video-tech.agent.final/video-tech.agent.final.md` (2,588 lines).

12 dimensions (`Agent_Results/research/video-tech_dim01..12.md`):

| Dim | Title (paraphrased)                              | Lines |
|-----|--------------------------------------------------|------:|
| 01  | Codec selection & low-latency encoding (HEVC/AV1/JPEG-XS/PyroWave) | 1,151 |
| 02  | Hardware encoders (NVENC, QSV, AMF, VideoToolbox, V4L2)            |   934 |
| 03  | Capture pipelines per OS                                           | 1,009 |
| 04  | Dual-path encoding (stream + record)                               | 1,013 |
| 05  | Recording storage & containers (MKV, fMP4, network targets)        | 1,234 |
| 06  | Audio capture/codec (Opus MultiStream, AC3, EAC3, DTS, Atmos)      | 1,141 |
| 07  | HDR / colour pipeline (HDR10, HDR10+, Dolby Vision, HLG)           | 1,058 |
| 08  | Adaptive bitrate, congestion control, FEC, SQP                     | 1,329 |
| 09  | Thermal-aware quality and GPU load-balancing                       | 1,181 |
| 10  | Test methodology, latency measurement, reference instrumentation    | 1,689 |
| 11  | Go pipeline implementation patterns (goroutines, sync.Pool, CGO)   | 1,466 |
| 12  | Network transport (WebRTC vs custom UDP, QUIC, AF_XDP, Pion)       | 1,593 |
|     | `video-tech_insight.md` (10 cross-dim insights)                    |   243 |
|     | `video-tech_cross_verification.md` (15 HC, 6 MC, 3 LC, 6 CZ)       |   206 |

### 2.4 Aggregate Source Volume

| Stream                  | Per-dim files | Final synthesis | Insight + CV | Total |
|-------------------------|--------------:|----------------:|-------------:|------:|
| 01 base                 |        12,329 |           2,817 |          286 | 15,432 |
| 02 latency              |         1,148 |           2,199 |          201 |  3,548 |
| 03 video_technology     |        14,798 |           2,588 |          449 | 17,835 |
| **Combined**            |    **28,275** |       **7,604** |      **936** | **36,815** |

The remaining lines in the original 84,473-line count are prompts (`*_prompt.txt`),
non-MD assets, and converted `.docx` outputs. The synthesis target stated in
clause **R-01** therefore translates to **≥36,815 lines of new prose under
`05_Response/`** (excluding figures, tables, and code samples — those are
additive). The web research addenda in `99_Web_Research_Addenda/` count
toward the total only when integrated as references in the chapter prose.

---

## 3. Output Structure

```
05_Response/
├── 00_Master_Plan.md                         ← THIS FILE
├── 01_Constitution.md                        ← R-01 … R-17 codified
├── 02_System_Overview.md                     ← scope, vision, dataflow
├── 03_Architecture/                          ← from Stream 1 (12 dims)
│   ├── 00_Index.md
│   ├── 01_Streaming_Protocols_and_Codecs.md
│   ├── 02_Controller_Input_Pipeline.md
│   ├── 03_Host_OS_Capture.md
│   ├── 04_Go_Client_Ecosystem.md
│   ├── 05_RealTime_APIs.md
│   ├── 06_Catalog_and_Assets.md
│   ├── 07_Host_Agent_and_Game_Lifecycle.md
│   ├── 08_Scalability_and_MultiRegion.md
│   ├── 09_Security_and_Isolation.md
│   ├── 10_WhiteLabel_and_Theming.md
│   ├── 11_TV_UX.md
│   └── 12_Latency_Engineering_Overview.md
├── 04_Latency/                               ← from Stream 2 (10 dims)
│   ├── 00_Index.md
│   ├── 01_Shared_Memory_and_Zero_Copy_IPC.md
│   ├── 02_io_uring_and_Kernel_Bypass.md
│   ├── 03_LockFree_Data_Structures.md
│   ├── 04_GPU_Direct_and_Hardware_Pipelines.md
│   ├── 05_UltraLowLatency_Network_Protocols.md
│   ├── 06_RealTime_OS_and_Scheduling.md
│   ├── 07_Controller_Input_Optimization.md
│   ├── 08_Frame_Pacing_and_VRR.md
│   ├── 09_Memory_and_Cache_Optimization.md
│   └── 10_Latency_Testing_and_Validation.md
├── 05_Video_Audio/                           ← from Stream 3 (12 dims)
│   ├── 00_Index.md
│   ├── 01_Codec_Selection.md
│   ├── 02_Hardware_Encoders.md
│   ├── 03_Capture_Pipelines.md
│   ├── 04_DualPath_Encoding.md
│   ├── 05_Recording_Storage.md
│   ├── 06_Audio_Pipeline.md
│   ├── 07_HDR_and_Color.md
│   ├── 08_ABR_FEC_Congestion.md
│   ├── 09_Thermal_and_GPU_Balancing.md
│   ├── 10_Measurement_and_QA.md
│   ├── 11_Go_Pipeline_Implementation.md
│   └── 12_Network_Transport.md
├── 06_Submodules/                            ← vasic-digital decomposition
│   ├── 00_Index.md
│   ├── 01_Submodule_Catalog.md
│   ├── 02_Containers_Submodule.md
│   ├── 03_Challenges_Submodule.md
│   ├── 04_HelixQA_Integration.md
│   └── per-submodule/<name>.md (one per HelixPlay submodule)
├── 07_Testing/
│   ├── 00_Index.md
│   ├── 01_Test_Matrix.md
│   ├── 02_Unit_Tests.md
│   ├── 03_Integration_Tests.md
│   ├── 04_E2E_Tests.md
│   ├── 05_Security_Tests.md
│   ├── 06_Benchmarking.md
│   ├── 07_Chaos.md
│   ├── 08_Stress.md
│   ├── 09_Smoke.md
│   ├── 10_Full_Automation.md
│   ├── 11_Challenges.md
│   └── 12_HelixQA_Autonomous.md
├── 08_Operations/
│   ├── 00_Index.md
│   ├── 01_Container_CI_CD.md
│   ├── 02_Quality_Gates_SonarQube_Snyk.md
│   ├── 03_Service_Discovery_and_Ports.md
│   ├── 04_Observability_and_Events.md
│   ├── 05_Tracking_GitHub_GitLab.md
│   └── 06_Git_Topology_and_Push_Policy.md
├── 09_Implementation_Phases/                 ← the executable plan
│   ├── 00_Phase_Index.md
│   ├── Phase_00_Foundation.md
│   ├── Phase_01_Containers_and_CI.md
│   ├── Phase_02_Core_Submodules.md
│   ├── Phase_03_Backend_Services.md
│   ├── Phase_04_Streaming_Pipeline.md
│   ├── Phase_05_Clients.md
│   ├── Phase_06_Host_Agent.md
│   ├── Phase_07_Latency_Optimization.md
│   ├── Phase_08_Audio_Surround.md
│   ├── Phase_09_Recording_and_Replay.md
│   ├── Phase_10_Monetization_and_Auth.md
│   ├── Phase_11_Hardening_and_Security.md
│   ├── Phase_12_Beta_Launch.md
│   └── Phase_13_GA.md
└── 99_Web_Research_Addenda/                  ← dated extension notes
    ├── 00_Index.md
    └── YYYY-MM-DD-<topic>.md (one per addendum)
```

Every chapter file is a **first-class artifact** with its own header, table
of contents, prose, diagrams (referenced from the source `cloudgaming.agent.final/`,
`video-tech.agent.final/`, etc.), tables, code listings, references, and a
`## Anti-Bluff Verification` block confirming what evidence was reviewed.

---

## 4. Synthesis Methodology

### 4.1 Per-Chapter Workflow

For each `05_Response/<chapter>/<file>.md`, an executor (subagent or
inline session) performs the following ten-step sequence. **No step may be
skipped.** All ten steps are mirrored as subtasks on the GitHub Project
and GitLab equivalent (R-17) before the executor begins.

1. **Read the dimension's request file** (e.g. `01_base/01_Request.md`).
2. **Read the dimension's per-dim research file** in full
   (e.g. `cloudgaming_dim03.md`).
3. **Read the dimension's slice of the per-stream final** (the
   `cloudgaming.agent.final.md` / `ZeroLatency_Communication_CloudGaming.md` /
   `video-tech.agent.final.md`).
4. **Read the relevant Insights** in `*_insight.md` and **the relevant
   Conflict Zones** in `*_cross_verification.md`.
5. **Run targeted web research** (≥3 queries, official docs / RFCs / IEEE /
   ACM / vendor SDKs / OSS repos) to extend the dimension with material not
   in the source. Record findings dated under
   `99_Web_Research_Addenda/YYYY-MM-DD-<topic>.md`.
6. **Identify cross-references** with already-written chapters and link
   them with `[<short>](path)` style links.
7. **Draft the chapter file** with the canonical header (see §4.2), full
   prose, all tables, all listings, all diagrams referenced.
8. **Anti-bluff verification** — fill the chapter's
   `## Anti-Bluff Verification` block listing every source artifact
   reviewed by absolute path, every web source consulted by URL+date,
   every conflict zone resolved with its decision and rationale.
9. **Cross-link & integrate** — update parent `00_Index.md` and add
   bidirectional links to other chapters that touch this dimension.
10. **Anti-bluff line-count check** — chapter prose must be **≥** the
    sum of the source dimension's per-dim file lines and the dimension's
    slice of the per-stream final. If shorter, expansion is required
    (more web research, more code samples, deeper rationale) — not
    prose padding.

### 4.2 Canonical Chapter Header

````markdown
# <Chapter Title>

> **Source dimensions:** <list of source files with absolute paths>
> **Source line count:** N (target: ≥N lines of synthesized prose)
> **Chapter targets (R-XX):** <list of contract clauses this chapter satisfies>
> **Cross-links:** <list of sibling chapters>
> **Status:** Draft v1 / Reviewed / Verified
> **Last updated:** YYYY-MM-DD

## Table of Contents
…
````

### 4.3 Anti-Bluff Verification Block (R-13)

Every chapter ends with:

````markdown
## Anti-Bluff Verification

### Source Evidence Reviewed
- `<absolute path 1>` — N lines, reviewed YYYY-MM-DD
- `<absolute path 2>` — N lines, reviewed YYYY-MM-DD
- …

### Web Sources Consulted
- [Title](URL) — accessed YYYY-MM-DD — N words extracted
- …

### Insights Incorporated
- Insight #N (`<insight file>`) — incorporated in §X.Y
- …

### Conflict Zones Resolved
| CZ-ID | Conflict | Decision | Rationale |
|-------|----------|----------|-----------|
| CZ-XX | …        | …        | …         |

### Coverage Confirmation
- Source line count: N
- Synthesized line count: M
- M ≥ N: ✅ / ❌ (if ❌, executor must expand before merge)

### Sign-off
Executed by: <agent / human handle>
Reviewed by: <agent / human handle>
Date: YYYY-MM-DD
````

### 4.4 Forbidden Outputs (R-02, R-13)

Any chapter that contains any of the following must be rejected and rewritten:

- `TODO`, `FIXME`, `tbd`, `xxx`, `???`, `placeholder`, "fill in later".
- Phrases like "and similar", "etc.", "and so on", "as appropriate" used
  to dodge specifying behaviour.
- Section headers with no body, or bodies under one paragraph.
- Tables with empty cells (mark `N/A` only if genuinely not applicable and
  add a footnote explaining why).
- Code listings with `pass` / `panic("not implemented")` / equivalent.
- Type signatures referenced but never defined.
- Configuration knobs mentioned but never given values & ranges.

---

## 5. Execution Model — Section-Stitched Subagent Dispatch (R1)

**Authority:** the original "one chapter per fresh subagent" model
specified at the top of session 1 (recorded as the §5.2 dispatch
template) was **superseded** during session 2 after C02, C03, and C04
each stalled on the runtime's 600 s stream watchdog at the chapter-
write step (RK09 realised). The Group A pilot of the new model
returned 360 lines in 220 s; the full C02 chapter (2,327 lines, 2.45×
the R-01 floor) stitched cleanly and is filed at
[`03_Architecture/01_Streaming_Protocols_and_Codecs.md`](03_Architecture/01_Streaming_Protocols_and_Codecs.md).
The new model — **R1, section-stitched dispatch** — is now the
canonical execution model for every queued chapter.

### 5.0 Why the original model failed

A single chapter dispatch had to: (a) read 5–10 source artifacts, (b)
issue ≥ 3 web searches, (c) write the addendum, (d) write a 700–1500-
line chapter, all within the runtime's 600-second streaming window.
The cumulative work fit the *token* budget but not the *streaming-time*
budget — long Write calls late in the trajectory tripped the watchdog
before completion. The Web Research Addenda were written first and
survived; the chapter Write did not. Salvage in
`99_Web_Research_Addenda/` proves the read-and-research stages
completed; only the final assembly failed.

### 5.1 The R1 model

Each chapter is decomposed into **3–4 section groups** of ~250–500
lines each. The orchestrator dispatches one **fresh subagent per
section group** with a narrow source-list scoped to that group. Each
subagent's only output is a single scratch file at
`/tmp/helixplay_chapter_scratch/<chapter>_section_<id>.md`. After all
section groups for a chapter return, the orchestrator:

1. Authors the canonical header (per §4.2) and the §12 References /
   Anti-Bluff Verification footer (per §4.3) inline — both reuse
   information already in the orchestrator's context (the dispatch
   prompts and the subagent reports).
2. Concatenates header + section-A + section-B + … + footer via
   `cat` redirection in a single Bash command, writing the final
   chapter file under `05_Response/`.
3. Verifies on disk: `wc -l`, `grep` for the §4.4 forbidden patterns
   (whitelisting the self-referential mentions inside the verification
   block), `grep -nE "^## "` to confirm section ordering.
4. On any failure (line floor not met, missing section, forbidden
   pattern in body), the offending group is re-dispatched with a
   targeted instruction.

**Properties of R1:**

- Each subagent's wall-clock < 600 s by design (small scope, no web
  research per subagent — addenda are written *once*, separately, before
  section dispatch).
- The orchestrator never reads the 2K+ lines of section content into
  its own context — `cat` does the assembly in shell.
- Per-section parallelism is the rule: 4 groups of one chapter run as
  4 simultaneous background dispatches. Multi-chapter parallelism
  (e.g. C03 and C04 sections together) is also acceptable when source
  files don't overlap in interesting ways and total subagent count
  stays ≤ 8.
- The orchestrator's stitching work is small enough to fit in a single
  Edit/Write/Bash call per chapter close-out.

### 5.2 Dispatch templates

#### 5.2.1 Web-research addendum (one per chapter, dispatched first)

```
Task: produce web-research addendum for <05_Response/<chapter>/<file>.md>

Output: /run/.../05_Response/99_Web_Research_Addenda/YYYY-MM-DD-<topic>.md

Inputs:
- Source per-dim file (absolute path)
- Insight + CV files for the dimension
- The list of suggested 2026 web topics for this dimension

Constraints:
- ≥3 WebSearch calls with year=2026 in the query
- Every URL captured in a structured table with title + access date
- Every claim sourced — no rephrasing that drops a source's constraint
- Single file output at the path above
- Subagent's own report ≤120 words

Stop conditions:
- File exists with ≥3 distinct URLs
- All URLs accessed 2026-04-28 or later
```

If the addendum already exists from a prior failed dispatch attempt,
this stage is skipped — addenda are append-only and salvageable.

#### 5.2.2 Section-group dispatch (3–4 per chapter, dispatched in parallel)

```
Task: produce section group <id> for <05_Response/<chapter>/<file>.md>

Output: /tmp/helixplay_chapter_scratch/<chapter>_section_<id>.md

Inputs:
- Master Plan path (this file)
- Constitution path
- System Overview path
- Architecture / Latency / Video-Audio Index path (sibling)
- Web-research addendum path
- Source per-dim file (focused — NOT all of it)
- The narrowed list of relevant source-research files (Stream-1/2/3
  per-dim, insight ID, cross-verification ID)

Constraints:
- §4.4 forbidden outputs absent
- Body prose ≥ the per-group minimum (typically 220–350 lines)
- All cross-links use relative paths from the chapter's perspective
- Constitution clauses cited by short ID (R-NN)
- Insights cited by ID
- Conflict zones explicitly resolved when in-scope for the group
- Output exactly ONE file at the scratch path
- No WebSearch calls (addendum already covers it)
- Subagent's own report ≤120 words

Stop conditions:
- File exists at scratch path with the required section headings
- Body prose ≥ minimum
- Forbidden patterns absent
```

#### 5.2.3 Orchestrator-side stitching

After all section groups return:

1. Author `c<NN>_header.md` and `c<NN>_footer.md` in the same scratch
   directory. Header includes canonical header per §4.2 + ToC. Footer
   includes §12 References + Anti-Bluff Verification block per §4.3.
2. `cat header section_a section_b … section_d footer >
   05_Response/<chapter>/<file>.md`.
3. `wc -l`, `grep -nE "^##? "`, forbidden-pattern grep.
4. Update Task tracker: chapter task → completed.
5. If any section is short or any forbidden pattern is in chapter
   prose (not in self-referential blocks), re-dispatch the offending
   group with a targeted instruction — do not patch by hand, the
   chapter is the contract.

### 5.3 When NOT to use the R1 model

The orchestrator must do these inline (no subagent):

- Cross-chapter integration (linking, index updates).
- Constitution amendments (`01_Constitution.md`).
- Master Plan updates (this file).
- GitHub / GitLab ticket creation.
- Git operations on the four configured remotes.
- Edits that touch fewer than ~250 lines and are not chapter content.

### 5.4 Failure handling

If a section group's subagent stalls (the same RK09 watchdog), the
section's scratch path will be missing or short. Recovery:

1. Re-dispatch the section group with the *exact same prompt* —
   subagents are stateless and the runtime's stall is non-deterministic.
2. If the second attempt also stalls, narrow the section group further
   (split into two smaller groups).
3. Never accept a partial section as if it were complete; the line-
   floor check would fail and the chapter close-out would block.

The R1 model has been proven on C02 (sections A, B, C, D) and is the
canonical pattern for the remaining queue (C03..C13 plus the rest of
§7.2). The §5.2 templates above replace all earlier "one chapter per
subagent" guidance.

---

## 6. Tracking — GitHub Projects + GitLab (R-17)

Every phase, every task, every subtask is mirrored on **both** GitHub Projects
(`HelixDevelopment/HelixPlay`) and the GitLab equivalent
(`gitlab.com/helixdevelopment1/HelixPlay`). The CLI tooling is `gh` and
`glab` respectively.

### 6.1 Naming Convention

- Phase: `[P00] Foundation`, `[P01] Containers & CI`, …
- Task: `[P03.T05] Implement Pion WebRTC adapter`
- Subtask: `[P03.T05.S02] Add ICE-Lite endpoint`

### 6.2 Cross-Linking

Every GitHub issue body links to its GitLab counterpart and vice versa.
The link format is fixed so a `grep` on either platform finds the pair.

### 6.3 Closure Criteria

A ticket is closed only when its corresponding chapter section has its
Anti-Bluff Verification block signed off **and** the linked code (when
implementation has begun) is merged through `simplify` review and passes
the test matrix in `07_Testing/`.

### 6.4 Local Mirror

`08_Operations/05_Tracking_GitHub_GitLab.md` documents the exact `gh` /
`glab` commands used to create, update, and close tickets, plus the
required label/milestone/iteration structure on each platform. The script
is committed in the future `Tooling` submodule.

---

## 7. Work Queue

### 7.1 Done (Session 1, 2026-04-28)

| ID  | Artifact                                              |
|-----|-------------------------------------------------------|
| W01 | `05_Response/00_Master_Plan.md` (this file)           |
| W02 | `05_Response/01_Constitution.md`                      |
| W03 | `05_Response/02_System_Overview.md`                   |
| W04 | `05_Response/` directory skeleton (`03_*` … `99_*`)    |

### 7.2 Queued

Chapters in **strict dependency order**. The executor must not skip
ahead — earlier chapters supply the vocabulary later chapters reuse.

| ID  | Chapter                                                  | Source dim → output | Min lines | Subagent |
|-----|----------------------------------------------------------|---------------------|----------:|----------|
| C01 | `03_Architecture/00_Index.md`                            | overview            |       400 | yes      |
| C02 | `03_Architecture/01_Streaming_Protocols_and_Codecs.md`   | base dim01          |       950 | yes      |
| C03 | `03_Architecture/02_Controller_Input_Pipeline.md`        | base dim02          |       700 | yes      |
| C04 | `03_Architecture/03_Host_OS_Capture.md`                  | base dim03          |     1,050 | yes      |
| C05 | `03_Architecture/04_Go_Client_Ecosystem.md`              | base dim04          |     1,500 | yes      |
| C06 | `03_Architecture/05_RealTime_APIs.md`                    | base dim05          |     1,100 | yes      |
| C07 | `03_Architecture/06_Catalog_and_Assets.md`               | base dim06          |     1,550 | yes      |
| C08 | `03_Architecture/07_Host_Agent_and_Game_Lifecycle.md`    | base dim07          |     1,600 | yes      |
| C09 | `03_Architecture/08_Scalability_and_MultiRegion.md`      | base dim08          |     1,150 | yes      |
| C10 | `03_Architecture/09_Security_and_Isolation.md`           | base dim09          |     1,100 | yes      |
| C11 | `03_Architecture/10_WhiteLabel_and_Theming.md`           | base dim10          |     1,500 | yes      |
| C12 | `03_Architecture/11_TV_UX.md`                            | base dim11          |     1,250 | yes      |
| C13 | `03_Architecture/12_Latency_Engineering_Overview.md`     | base dim12          |     1,450 | yes      |
| C14 | `04_Latency/00_Index.md`                                 | overview            |       300 | yes      |
| C15 | `04_Latency/01_Shared_Memory_and_Zero_Copy_IPC.md`       | latency dim01       |       250 | yes      |
| C16 | `04_Latency/02_io_uring_and_Kernel_Bypass.md`            | latency dim02       |       250 | yes      |
| C17 | `04_Latency/03_LockFree_Data_Structures.md`              | latency dim03       |       250 | yes      |
| C18 | `04_Latency/04_GPU_Direct_and_Hardware_Pipelines.md`     | latency dim04       |       300 | yes      |
| C19 | `04_Latency/05_UltraLowLatency_Network_Protocols.md`     | latency dim05       |       250 | yes      |
| C20 | `04_Latency/06_RealTime_OS_and_Scheduling.md`            | latency dim06       |       300 | yes      |
| C21 | `04_Latency/07_Controller_Input_Optimization.md`         | latency dim07       |       250 | yes      |
| C22 | `04_Latency/08_Frame_Pacing_and_VRR.md`                  | latency dim08       |       250 | yes      |
| C23 | `04_Latency/09_Memory_and_Cache_Optimization.md`         | latency dim09       |       250 | yes      |
| C24 | `04_Latency/10_Latency_Testing_and_Validation.md`        | latency dim10       |       300 | yes      |
| C25 | `05_Video_Audio/00_Index.md`                             | overview            |       400 | yes      |
| C26 | `05_Video_Audio/01_Codec_Selection.md`                   | video dim01         |     1,250 | yes      |
| C27 | `05_Video_Audio/02_Hardware_Encoders.md`                 | video dim02         |     1,050 | yes      |
| C28 | `05_Video_Audio/03_Capture_Pipelines.md`                 | video dim03         |     1,150 | yes      |
| C29 | `05_Video_Audio/04_DualPath_Encoding.md`                 | video dim04         |     1,150 | yes      |
| C30 | `05_Video_Audio/05_Recording_Storage.md`                 | video dim05         |     1,350 | yes      |
| C31 | `05_Video_Audio/06_Audio_Pipeline.md`                    | video dim06         |     1,250 | yes      |
| C32 | `05_Video_Audio/07_HDR_and_Color.md`                     | video dim07         |     1,150 | yes      |
| C33 | `05_Video_Audio/08_ABR_FEC_Congestion.md`                | video dim08         |     1,450 | yes      |
| C34 | `05_Video_Audio/09_Thermal_and_GPU_Balancing.md`         | video dim09         |     1,300 | yes      |
| C35 | `05_Video_Audio/10_Measurement_and_QA.md`                | video dim10         |     1,800 | yes      |
| C36 | `05_Video_Audio/11_Go_Pipeline_Implementation.md`        | video dim11         |     1,600 | yes      |
| C37 | `05_Video_Audio/12_Network_Transport.md`                 | video dim12         |     1,750 | yes      |
| S01 | `06_Submodules/01_Submodule_Catalog.md`                  | new                 |       800 | inline   |
| S02 | `06_Submodules/02_Containers_Submodule.md`               | new                 |       400 | inline   |
| S03 | `06_Submodules/03_Challenges_Submodule.md`               | new                 |       400 | inline   |
| S04 | `06_Submodules/04_HelixQA_Integration.md`                | new                 |       400 | inline   |
| S05 | `06_Submodules/per-submodule/<name>.md` (×N)             | new                 |    300 ea | inline   |
| T01 | `07_Testing/01_Test_Matrix.md`                           | new                 |       600 | inline   |
| T02 | `07_Testing/02..12` (one per test type)                  | new                 |  300-500ea| inline   |
| O01 | `08_Operations/01_Container_CI_CD.md`                    | new                 |       600 | inline   |
| O02 | `08_Operations/02..06`                                   | new                 |  300-500ea| inline   |
| P00 | `09_Implementation_Phases/Phase_00_Foundation.md`        | plan                |       800 | inline   |
| P01 | … `Phase_13_GA.md`                                       | plan                |  500-1000 | inline   |
| W05 | Update repo-root `CLAUDE.md` to point at Constitution    | repo                |    n/a    | inline   |
| W06 | Create repo-root `AGENTS.md`                             | repo                |    n/a    | inline   |
| W07 | Mirror to GitHub Projects + GitLab                       | tracking            |    n/a    | inline   |
| W08 | Sweep stray `HelixAgent/.git/index.lock` rule from `.claude/settings.local.json` | repo | n/a | inline |

The minimum-line targets above sum to **47,250 lines of new prose** —
significantly above the 36,815-line floor mandated by R-01. This headroom
exists because the synthesis must add web-research extensions, deeper
rationale, and bidirectional cross-links beyond what the sources contained.

### 7.3 Out-of-Scope for This Plan

- **Implementation source code** itself. The plan produces phases/tasks
  with executable detail; the actual code lives under each submodule's
  repository once Phase 02 begins. The `04_Request.md` clause "Once full
  and final documentation … is created, validated and verified in
  multiple-passes we can implement the whole System!" is a **gate**: no
  implementation work begins until the documentation set in §7.1+§7.2
  is signed off.

- **V1 phase research** under `docs/research/chapters/V1/`. That phase
  is queued separately. When V1 begins, mirror the MVP layout
  (`01_*/01_Request.md`, …, `04_Request.md`) and produce a `V1/05_Response/`
  with its own master plan.

---

## 8. Risk Register

| ID  | Risk                                                                                       | Mitigation                                                                                               |
|-----|--------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------|
| RK01| Subagent produces shorter prose than the line target (R-01 violation).                     | Anti-Bluff line-count check in §4.1 step 10. Reject and re-dispatch with explicit expansion instructions.|
| RK02| Subagent invents API names / file paths not in the sources (R-02 violation).               | Forbidden-outputs check (§4.4). Cross-verify cited identifiers against source files via grep.            |
| RK03| Web research returns outdated material.                                                    | All `WebSearch` calls must include `2026` in the query. Sources older than 2024 must have a justification footnote. |
| RK04| Conflict Zones get re-introduced silently in the synthesis.                                | Every chapter MUST resolve its CZ entries explicitly in the verification block.                          |
| RK05| GitHub Projects / GitLab CLI rate-limits during bulk ticket creation.                       | Batch creation in chunks of ≤50 with 60s spacing; idempotent issue creation keyed on `[Pxx.Tyy.Szz]`.    |
| RK06| `.claude/settings.json` `Stop` hook references missing `scripts/claim-check.sh` (CLAUDE.md). | Track as W08-adjacent; either implement the script in the future Tooling submodule or remove the hook.   |
| RK07| Push to `origin` only updates GitFlic, not GitHub.                                          | All push operations must list every remote explicitly. Documented in `08_Operations/06_Git_Topology_and_Push_Policy.md`. |
| RK08| Submodule explosion: many small `vasic-digital` repos with overlapping scope (R-04).        | The `06_Submodules/01_Submodule_Catalog.md` must include a "duplication check" section that grep-searches existing `vasic-digital` repos before any new submodule is proposed. |
| RK09| Subagent context window saturated mid-chapter, producing truncated output.                  | Chapter dispatches use `isolation: worktree`. Truncation detected by missing canonical footer block.     |
| RK10| Anti-bluff regression: tests added that pass without exercising real behaviour (R-13).      | `07_Testing/11_Challenges.md` defines a meta-test that boots the full container stack and sanity-checks every advertised feature end-to-end.                                                |

---

## 9. Definitions of Done

The synthesis programme is **complete** when **all** of the following hold:

1. Every row in §7.2 is checked off and committed under `05_Response/`.
2. The aggregate prose line count under `05_Response/` is ≥ 36,815
   (R-01 floor) and ideally ≥ 47,250 (per-chapter targets).
3. `01_Constitution.md` is referenced from repo-root `CLAUDE.md`,
   repo-root `AGENTS.md`, every submodule's `CLAUDE.md`/`AGENTS.md`.
4. Every phase, task, and subtask in `09_Implementation_Phases/`
   has a matching ticket on **both** GitHub Projects and GitLab,
   with cross-links resolved.
5. No file under `05_Response/` matches any of the forbidden patterns
   in §4.4 (a CI script in the Tooling submodule grep-checks).
6. Every chapter's Anti-Bluff Verification block is fully signed off.
7. The four git remotes (github, gitlab, gitverse, gitflic) all show
   the same head; `origin`'s split push has been used to update
   GitFlic explicitly.

Once **§9** is satisfied, and only then, implementation work
(`09_Implementation_Phases/Phase_00_Foundation.md` onward) may begin.

---

## 10. Session Log

| Session | Date       | Author              | Outcome                                                                |
|---------|------------|---------------------|------------------------------------------------------------------------|
| 1       | 2026-04-28 | Claude (orchestrator)| Skeleton created. `00_Master_Plan.md`, `01_Constitution.md`, `02_System_Overview.md` written. Tasks #1–#13 created in TaskCreate. Subagent dispatch of `03_*`/`04_*`/`05_*` queued. |
| 2       | 2026-04-28 | Claude (orchestrator)| C01 (`03_Architecture/00_Index.md`, 617 lines) landed via the §5.2 dispatch template. C02/C03/C04 dispatched under the same template each stalled on the 600 s watchdog at the chapter-write step (RK09 realised), but each subagent **did** complete its Web Research Addendum first — addenda salvaged at `99_Web_Research_Addenda/2026-04-28-streaming-protocols-and-codecs.md`, `…-controller-input-pipeline.md`, `…-host-os-capture.md`. Operator authorised the **R1 section-stitched dispatch model** as the recovery: orchestrator scaffolds each chapter, narrow-scope per-section subagents write to `/tmp/helixplay_chapter_scratch/<chapter>_section_<id>.md`, orchestrator stitches via Edit/Write. Pilot C02 Group A returned 360 lines in 220 s — well under watchdog — validating the per-section dispatch shape. C02 Groups B/C/D launched in parallel; awaiting completion. Wait-window work: stray `HelixAgent/.git/index.lock` allow-rule swept from `.claude/settings.local.json`; root `CLAUDE.md` updated to point at the Constitution; root `AGENTS.md` created (250 lines). Tasks #1–#3 and #12 marked `completed`; #14 (C02 stitch), #15 (C03 re-dispatch), #16 (C04 re-dispatch), #17 (capture R1 model in §5 once stitching validates) created. |
| 2-cont. | 2026-04-28 | Claude (orchestrator)| Continued: C02 stitched (2,327 lines, 2.45× floor); R1 model documented in §5 (replacing the original §5.2 template); C03 re-dispatched and stitched (2,819 lines, 4.03× floor); C04 re-dispatched and stitched (2,887 lines, 2.75× floor); C05 addendum landed (320 lines, 26 URLs) and closed Architecture-Index OQ-01 (Wails v2 default, Tauri-Go Phase 2) and OQ-02 (Compose for TV primary, Flutter fallback) and introduced CZ-CW1 (TinyGo vs `GOOS=js GOARCH=wasm` for Pion); C05 sections stitched (3,336 lines, 2.22× floor); C06 addendum landed (426 lines, 58 URLs) and introduced CZ-RA1..CZ-RA4 (`coder/websocket` over Gorilla, Redis cache+rate-limit only, HTTP/3 via Connect-Go, Valkey default); C06 sections stitched (3,450 lines, 2.93× floor); C07 addendum landed (471 lines, 85 URLs, 9 clusters) and refined Insight #4 with seven Z-N specifics (IGDB tier ladder, SteamGridDB pagination, Epic no public catalog API, JPEG XL Chrome 145 flag, Redis Stack EOL → Valkey, RAWG attribution gates, EU DSA Article 17 binding); C07 sections stitched (2,991 lines, 1.93× floor); C08 addendum landed (358 lines, 71 URLs, 9 clusters) confirming Sunshine++ (Sunshine v2026.423.21833 only 5 days old at audit, active maintenance) and surfacing 7 Z-contradictions (Vanguard motherboard attestation, Steam Input licensing unclear, Battle.net URI broken since 2024, Riot unified client no per-game URI, EAC vs Win11 24H2 KMHESP regression, ViGEmBus 1.22.0 vs Virtual Pad commercial, Sunshine multi-session removal). Architecture chapters delivered to date: 7 of 13 (C01-C07), totalling 18,427 lines under `03_Architecture/` plus 7 web-research addenda totalling 2,541 lines under `99_Web_Research_Addenda/`. |
| 3       | 2026-04-28 | Claude (orchestrator)| **Host-disruption incident.** Operator's development host suspended / signed out mid-session, terminating the active orchestrator and its in-flight subagents. Read-only investigation conducted: `git status` (working tree showed C01-C07 chapters + 7 addenda + foundational docs as expected); `git submodule status` (empty; no `.gitmodules`); `which docker podman` (Docker absent, Podman installed at `/usr/bin/podman` but never invoked in this session); `ls Dockerfile docker-compose.yml compose.yml` (none exist in repo); `.claude/settings.json` and `.claude/settings.local.json` reviewed (only the `Stop` hook on missing `scripts/claim-check.sh` and the four trivial `Bash(...)` allow-rules, none of which can suspend/sign-out the host); `git remote -v` (origin is composite push to all four remotes; an `upstream` → gitflic remote also present). **Conclusion**: no causal link from any tool call we issued to the host disruption. Bash invocations in the session were limited to `find`/`wc`/`grep`/`cat`/`mkdir -p`/`ls`/`git status`/`cat`-stitching — none have OS privilege to suspend or sign out. Most likely actual cause: host-side power management policy, battery exhaustion, system update reboot, or display-server crash. **Action**: added Constitution §11.5 (Operational Integrity, R-18) with sub-sections covering forbidden host-disruptive commands (§11.5.1), container guard rails (§11.5.2), container-runtime hazards (§11.5.3), hooks/CI obligation (§11.5.4), subagent isolation (§11.5.5), recovery posture (§11.5.6), and anti-bluff testing reinforcement (§11.5.7); added R-18 to §1 contract table; updated repo-root `CLAUDE.md` and `AGENTS.md` to reference R-18; marked task #20 (C07) `completed`; created tasks #22 (R-18 directive) and #23 (commit+push); committed and pushed session 1 + 2 + 3 work to all four remotes via `git push origin main` (origin is composite push). C08 sections re-dispatch resumes after the push lands. |

Future sessions append rows here. **Never** rewrite earlier rows;
the log is append-only and forms part of the audit trail.

---

## 11. Cross-References

- Project brief: `docs/research/chapters/MVP/04_Request.md`
- Stream 1 brief: `docs/research/chapters/MVP/01_base/01_Request.md`
- Stream 2 brief: `docs/research/chapters/MVP/02_latency/01_Request.md`
- Stream 3 brief: `docs/research/chapters/MVP/03_video_technology/Request.md`
- Constitution: `docs/research/chapters/MVP/05_Response/01_Constitution.md`
- System overview: `docs/research/chapters/MVP/05_Response/02_System_Overview.md`
- Repo CLAUDE.md: `CLAUDE.md`
- Containers submodule: <https://github.com/vasic-digital/Containers>
- Challenges submodule: `git@github.com:vasic-digital/Challenges.git`
- HelixQA: `git@github.com:HelixDevelopment/HelixQA.git`

End of Master Plan v1 — 2026-04-28.
