# T09 — Smoke Tests

> **Source dimensions:** [`00_Index.md`](00_Index.md), [`01_Test_Matrix.md`](01_Test_Matrix.md) §6.8; [`../06_Submodules/04_HelixQA_Integration.md`](../06_Submodules/04_HelixQA_Integration.md) §9 (deployment gate).
> **Source line count:** chapter floor 300 lines per Master Plan §7.2 row T09.
> **Chapter targets:** R-11 (post-deploy basic liveness), R-13.
> **Cross-links:** [`04_E2E_Tests.md`](04_E2E_Tests.md), [`12_HelixQA_Autonomous.md`](12_HelixQA_Autonomous.md).
> **Status:** Draft v1.
> **Last updated:** 2026-04-30.

---

## 1. Role in the Master Test Matrix

Smoke is **test-type 8 of 10**. Cadence: post-deploy (canary + pre-release + production deployment). Per-PR + nightly do **not** run Smoke directly — Smoke's environment is a *deployed* container, not a CI-time container.

Smoke's role: **basic liveness verification**. Within 30 seconds of deployment completion, can the binary start, answer one request, and report ready? If not, auto-revert the deployment per [S04 §9](../06_Submodules/04_HelixQA_Integration.md#9-deployment-gating-the-cosign--challenges--sbom--visibility-quad).

Smoke is **not** a substitute for E2E (T04). Smoke verifies the binary started; E2E verifies the user journey works. A binary that starts but fails the user journey is a defect Smoke does not catch.

---

## 2. The Discipline — What Makes "Green" Green

A green Smoke row means:

1. **Binary started** within the cold-start budget (≤ 5 seconds from container `Up` to first health-probe success).
2. **Health endpoint** returns HTTP 200 + a non-empty version string within 30 seconds of deploy.
3. **One canonical request** (per submodule, defined in `<submodule>/tests/smoke/probe.sh`) succeeds.
4. **No crash** during the 30-second window — the binary doesn't `os.Exit(1)`, doesn't `panic`, doesn't `SIGSEGV`.
5. **Required environment** verified — every `HELIX_*` env var the binary expects is present (per the submodule's S05 §9.1 *Configuration knobs* table); a missing required env var fails the Smoke probe.

---

## 3. Tooling

- **`probe.sh`** — per-submodule shell script under `<submodule>/tests/smoke/probe.sh`. Shell, not Go, because the post-deploy environment doesn't have a Go toolchain.
- **`curl`** + **`jq`** — within `probe.sh`, for HTTP probes.
- **`grpcurl`** — for gRPC-frame-style probes.
- **`helixqa-deploy-gate`** — the orchestrator that runs probes post-deploy + auto-reverts on red.

---

## 4. The Canonical `probe.sh` Skeleton

Every submodule ships a `tests/smoke/probe.sh` of this shape:

```bash
#!/bin/bash
# tests/smoke/probe.sh — runs post-deploy. Exit 0 = green; non-zero = red.
set -euo pipefail

ENDPOINT="${HELIX_SMOKE_ENDPOINT:-http://localhost:8443}"
TIMEOUT="${HELIX_SMOKE_TIMEOUT:-30}"

# 1. Wait up to TIMEOUT seconds for the binary to be ready.
deadline=$(( $(date +%s) + TIMEOUT ))
until curl -s -f -o /dev/null "${ENDPOINT}/healthz"; do
    if [ "$(date +%s)" -ge "$deadline" ]; then
        echo "FAIL: ${ENDPOINT}/healthz did not return 200 within ${TIMEOUT}s"
        exit 1
    fi
    sleep 1
done

# 2. Verify version string is non-empty.
version=$(curl -s "${ENDPOINT}/healthz" | jq -r '.version')
if [ -z "$version" ] || [ "$version" = "null" ]; then
    echo "FAIL: /healthz returned empty version"
    exit 1
fi
echo "Version: $version"

# 3. Run the canonical request for this submodule.
case "$HELIX_SUBMODULE" in
    helix-grpc-frame)
        grpcurl -plaintext "${ENDPOINT}" helix.v1.Control/Health
        ;;
    helix-r18-safeexec)
        # SafeExec a benign command + verify ErrForbidden on a forbidden one.
        # (probe binary embedded in the container)
        /usr/local/bin/r18-probe ok-command true
        /usr/local/bin/r18-probe forbidden-rejected pm-suspend
        ;;
    helix-vault)
        curl -s -f "${ENDPOINT}/vault-probe" \
            -H "X-Tenant-ID: smoke-tenant" \
            -d '{"op":"encrypt","plain":"smoke"}'
        ;;
    *)
        echo "FAIL: no smoke probe defined for $HELIX_SUBMODULE"
        exit 1
        ;;
esac

echo "Smoke OK: $HELIX_SUBMODULE"
```

The `case` block lists every submodule's canonical request. A submodule without a registered case fails Smoke — the lint catches this.

---

## 5. Anti-Pattern Catalogue

### 5.1 Long-running Smoke

A Smoke probe that takes 5 minutes is not a Smoke probe; it's a delayed E2E. Smoke must complete in ≤ 30 seconds. Anything longer belongs in T04.

### 5.2 Smoke that depends on data from previous deploy

```bash
# Assume yesterday's deploy left a session token at /tmp/helix-token
token=$(cat /tmp/helix-token)
```

A Smoke probe must be **stateless** — it gets a fresh container per deploy. Bootstrap any required state via a freshly-issued token within the probe.

### 5.3 Smoke that mocks the dependencies

Smoke runs against the **deployed** binary in its **deployed** environment. There are no mocks at this layer — if a dependency is unreachable, the probe fails (correctly, because production cannot tolerate the unreachable dependency).

### 5.4 Smoke that retries on failure

```bash
for i in 1 2 3; do
    if curl -s -f /healthz; then exit 0; fi
done
exit 1
```

Three retries hide flake. The within-30-second wait loop in §4 is the **only** retry — and it exists to give the binary time to start, not to mask failure.

### 5.5 Smoke that runs commands inside the deployed container

Smoke runs **outside** the container, against its endpoints. Running `docker exec` to inspect internal state is debugging, not smoke testing.

---

## 6. CI Lane Invocation Pattern

Smoke runs in the deploy pipeline, not the CI matrix:

```yaml
# vasic-digital/Containers/ci-fragments/deploy.yml
deploy-and-smoke:
  steps:
    - name: Deploy to canary
      run: |
        helixqa-deploy --target=canary --image=$IMAGE_DIGEST
    - name: Smoke probe
      run: |
        export HELIX_SMOKE_ENDPOINT="https://canary.helix.example.com"
        export HELIX_SUBMODULE="$SUBMODULE"
        timeout 35s tests/smoke/probe.sh
    - name: Auto-revert on smoke fail
      if: failure()
      run: |
        helixqa-deploy --target=canary --revert --reason="smoke-failure"
```

The `timeout 35s` enforces the 30-second budget + 5-second margin. A smoke fail triggers automatic rollback (per [S04 §9](../06_Submodules/04_HelixQA_Integration.md#9-deployment-gating-the-cosign--challenges--sbom--visibility-quad) deployment-gate's fail-closed posture).

---

## 7. The Auto-Revert Contract

Smoke is the **only** test type that triggers automatic rollback. The contract:

1. Smoke red → `helixqa-deploy --revert` fires within 60 seconds of detection.
2. Rollback target = the previous green-tagged image (per the run-archive's previous canary entry).
3. P1 alert pages the on-call (per [S04 §8](../06_Submodules/04_HelixQA_Integration.md#8-alert-routing-and-on-call-rotation)).
4. The bad image is **not auto-removed from the registry** — operator may want to debug or sign-off-with-exceptions.
5. The next deploy attempt requires a fresh image (different digest); pushing the same digest again will hit the `helixqa-deploy --reject-known-bad` gate.

The auto-revert is **operator-overridable** — `helixqa-deploy --target=canary --image=$DIGEST --no-auto-revert` for emergency hotfixes. This requires Constitution §13 *Exceptions* sign-off + an entry in the run-archive's exception log.

---

## 8. Per-Submodule Smoke Probe Catalogue

Reference for every submodule's canonical request — the table is the source of truth that `probe.sh` switches on (per §4):

| Submodule         | Canonical request                                          |
|-------------------|-------------------------------------------------------------|
| helix-r18-safeexec | run-allowed + run-forbidden + assert ErrForbidden          |
| helix-grpc-frame   | gRPC `Health()`                                             |
| helix-tv-input     | TV input dispatcher liveness (event-channel non-empty)     |
| helix-vault        | encrypt + decrypt round-trip on smoke-tenant               |
| helix-tenant       | LoadByID(smoke-tenant) returns non-empty Theme              |
| helix-shm          | allocate 4 KiB page + verify size                          |
| helix-iouring      | submit + reap one read SQE                                  |
| helix-xdp          | attach XDP program to loopback + redirect probe packet     |
| helix-lockfree     | SPSC ring write + read 1K items                             |
| helix-gpu-direct   | allocate 8 MiB GPU buffer + export handle                  |
| helix-network      | UDP loopback marked + ECN bits propagate                   |
| helix-rtos         | promote thread + chrt -p verifies SCHED_FIFO               |
| helix-input        | inject BTN_A event + verify uinput sees it                 |
| helix-display      | read EDID + report VRR support                              |
| helix-mempool      | acquire + release 1K Frame objects with zero allocs        |
| helix-allocator    | helix-allocator-vet on a clean test fixture, exit 0        |
| helix-bench        | run 100-sample workload + report p999                      |
| helix-codec        | build + validate canonical Profile                         |
| helix-encoder      | encode 1 frame + verify NAL emerges                        |
| helix-capture      | capture 1 frame from Xvfb + verify dimensions              |
| helix-dualpath     | submit 1 EncodedFrame + verify both rungs receive          |
| helix-record       | record 1-sec clip + verify file parses                     |
| helix-audio        | capture 1-sec silence + verify Opus packets emerge         |
| helix-hdr          | tag 1 frame with HDR10 SEI + verify present                |
| helix-abr          | feed RTCP probe + verify Decision returned                 |
| helix-thermal      | poll telemetry + verify TemperatureC > 0                  |
| helix-vqa          | run known-pair VMAF + verify score within ±0.5            |
| helix-pipeline     | Pipeline.Start + Stop with clean event log                |
| helix-transport    | localhost connect + send + receive probe packet           |

The 29 cases are exhaustive; a 30th submodule's row in this table is added at S01 §3.1 amendment time.

---

## 8a. Smoke Probe Failure Modes

The seven canonical reasons a probe goes red:

| Failure                        | Likely cause                                                | Auto-remediation                                          |
|--------------------------------|-------------------------------------------------------------|------------------------------------------------------------|
| 1. `/healthz` 200 timeout       | Binary failed to bind port                                  | Auto-revert; pages on-call P1                              |
| 2. `/healthz` returns 500       | Backend unreachable (e.g. CockroachDB down)                 | Auto-revert; investigate backend availability before retry|
| 3. `/healthz` empty version    | Binary built without `-ldflags="-X main.Version=..."`      | Auto-revert; fix build pipeline                            |
| 4. Canonical request 4xx       | Smoke tenant misconfigured                                  | Auto-revert; verify smoke-tenant exists in catalog        |
| 5. Canonical request 5xx       | Internal error                                               | Auto-revert; pages P1                                      |
| 6. Container exited            | Crash during cold start (panic, OOM, missing env var)      | Auto-revert; capture container logs for postmortem        |
| 7. Probe binary not in image   | Build error — probe binary missing from runtime image       | Auto-revert; fix Dockerfile to include probe binary       |

Each failure mode has a specific operational response; operators don't need to triage from scratch.

## 8b. Smoke Telemetry

Per Constitution §10, every smoke probe emits an OTLP span:

```
span: helix.smoke.probe
attributes:
  smoke.submodule: helix-pipeline
  smoke.endpoint: https://canary.helix.example.com
  smoke.version: 0.7.4-rc.2
  smoke.duration_ms: 1850
  smoke.result: ok
  smoke.deployment_id: dep-2026-04-30-1432
```

The span is exported to HelixQA's run-archive (per [S04 §7](../06_Submodules/04_HelixQA_Integration.md#7-run-archive-and-audit-trail-immutable-queryable-signed)) so smoke history is queryable. A `smoke.result: failed` span attached to a deployment ID becomes part of that deployment's audit trail.

## 9. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-T09-A            | Per-tenant smoke probes — should each tenant's deployment have a tenant-specific probe?                       | T09 next revision                                   |
| OQ-T09-B            | Auto-revert override semantics — operator approval workflow?                                                   | `08_Operations/04_Observability_and_Events.md`     |
| OQ-T09-C            | Smoke probe timeout — fixed 30 s or per-submodule tunable?                                                     | T09 next revision                                   |

---

## 9a. Smoke vs Health-Check Distinction

A naïve operator may assume *Smoke* = *Health Check*. They differ:

| Property                           | Health check                          | Smoke probe                                  |
|------------------------------------|---------------------------------------|----------------------------------------------|
| Frequency                          | every 5 s (continuously, by orchestrator) | once, post-deploy                          |
| Scope                              | shallow (process up; port bound)      | deep (binary functional; canonical request) |
| Owner                              | container orchestrator (k8s, Nomad)   | HelixQA deployment gate                      |
| Output                             | binary up/down                         | structured probe report                       |
| Action on failure                  | restart container (k8s pattern)       | auto-revert deployment (S04 §9 fail-closed) |

Both exist; they serve different purposes. The HTTP `/healthz` endpoint serves both — it answers the orchestrator's continuous polls AND the smoke-probe's one-time post-deploy check. The probe additionally executes the canonical request (§4 §case block) which the health-check does not.

## 10. References & Anti-Bluff Verification

### 10.1 Internal

- [`01_Test_Matrix.md`](01_Test_Matrix.md) §6.8.
- [`../06_Submodules/04_HelixQA_Integration.md`](../06_Submodules/04_HelixQA_Integration.md) §9.

### 10.2 External (web)

- curl manual: https://curl.se/docs/manual.html (accessed 2026-04-30).
- grpcurl: https://github.com/fullstorydev/grpcurl (accessed 2026-04-30).

### 10.3 Anti-Bluff Verification

- Coverage: chapter exceeds the 300-line floor.
- Forbidden patterns: clean. Three §9 open questions named with deferred resolution chapters.
- §8 catalogue ensures every of the 29 submodules has a canonical smoke probe; the lint rejects un-catalogued submodules.

### 10.4 The Probe-Binary Build Discipline

The per-submodule `r18-probe`, `vault-probe`, and similar tiny probe-binaries referenced in §4 are built as part of the submodule's CI pipeline and copied into the runtime image at `/usr/local/bin/`. The Dockerfile pattern:

```dockerfile
# tail of the submodule's Dockerfile
COPY --from=builder /out/binary           /
COPY --from=builder /out/probe            /usr/local/bin/probe
ENTRYPOINT ["/binary"]
```

The probe is a tiny statically-linked Go binary (typically 5–10 MiB) that exits 0 / non-zero based on the canonical request's success. Because the probe binary travels in the same image as the main binary, smoke testing is genuinely against the deployed artefact — there is no runtime-only-installed probe that could differ.

### 10.5 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30.
- Reviewed by: pending operator review.

End of `07_Testing/09_Smoke.md` — 2026-04-30.
