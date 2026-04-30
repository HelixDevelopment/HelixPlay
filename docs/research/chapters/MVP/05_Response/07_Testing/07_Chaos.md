# T07 — Chaos Tests

> **Source dimensions:** [`00_Index.md`](00_Index.md), [`01_Test_Matrix.md`](01_Test_Matrix.md) §6.6; [`../06_Submodules/per-submodule/`](../06_Submodules/per-submodule/) §9.3 (per-submodule *Common errors* tables — the canonical chaos-injection target list); [`../06_Submodules/03_Challenges_Submodule.md`](../06_Submodules/03_Challenges_Submodule.md) §5.4 (network-impairment composition).
> **Source line count:** chapter floor 300 lines per Master Plan §7.2 row T07.
> **Chapter targets:** R-11 (every documented failure mode injected + recovery verified), R-13.
> **Cross-links:** [`08_Stress.md`](08_Stress.md), [`11_Challenges.md`](11_Challenges.md).
> **Status:** Draft v1.
> **Last updated:** 2026-04-30.

---

## 1. Role in the Master Test Matrix

Chaos is **test-type 6 of 10**. Cadence: nightly + canary + pre-release. Per-PR does **not** run chaos by default (the wall-clock + flake risk are too high for the per-PR budget); operators may opt in via PR labels for high-risk submodule changes.

Chaos's role: **prove the system tolerates the documented failure modes**. Each submodule's [S05 §9.3 *Common errors* table](../06_Submodules/per-submodule/) enumerates expected failure conditions; the Chaos row injects each condition and verifies recovery. Untested error paths rot — the submodule's error handling for `vault: ErrVaultSealed` is only known to work if a test actually seals the Vault during a run.

---

## 2. The Discipline — What Makes "Green" Green

A green Chaos row means:

1. **Every documented failure mode in §9.3 is injected** by at least one chaos test (verified by `helix-chaos-coverage` lint walking the §9.3 table).
2. **Recovery succeeds** — after the injection ends, the submodule returns to a healthy steady state within the documented budget.
3. **No collateral damage** — adjacent submodules don't crash, leak, or report errors during the targeted submodule's chaos.
4. **Chaos itself is observable** — every injection emits an OTLP span with `chaos.target=<submodule>`, `chaos.fault=<kind>`, `chaos.duration=<seconds>`. The operator can audit the chaos history.
5. **Bounded blast radius** — the chaos test runs in a hermetic test topology; never against shared dev resources.

---

## 3. Tooling

### 3.1 Toxiproxy (network impairment)

Process-level network proxy with on-the-fly fault injection:

- **Latency**: add fixed or random delay.
- **Bandwidth**: cap throughput.
- **Slow close**: hold the close-handshake.
- **Slicer**: split TCP segments.
- **Timeout**: drop after N seconds.
- **Reset peer**: send RST.

Used for the network-impairment family of chaos tests. Pinned to ≥ v2.10.0 per [T01 §10](01_Test_Matrix.md#10-tooling-lockstep--pinned-versions-across-the-fleet).

### 3.2 chaos-mesh (Kubernetes-orchestrated)

Used for the larger-blast-radius chaos tests where the failure mode is a host-level event (network partition between nodes; pod kill; CPU/IO/memory pressure; DNS chaos; time chaos; HTTP chaos; kernel chaos). Requires a Kubernetes cluster — chaos-mesh is the default for the canary cadence's chaos topologies.

Pinned to ≥ v2.7.0.

### 3.3 Custom syscall-rejection harness (R-18)

The [host-integrity-scan](../06_Submodules/02_Containers_Submodule.md#34-the-host-integrity-scan-shared-lane) lane is itself a chaos-style test — it injects forbidden subprocess invocations and asserts the wrapper rejects them. This is technically Chaos rather than Security because it tests **resilience to misuse**, not **resilience to attack**.

---

## 4. Per-Submodule Chaos Targets

Each submodule's [S05 §9.3 *Common errors*](../06_Submodules/per-submodule/) table is the canonical chaos-injection target list. Excerpts:

| Submodule         | Failure mode                          | Injection method                                |
|-------------------|---------------------------------------|-------------------------------------------------|
| helix-vault       | `ErrVaultSealed`                       | `vault operator seal` mid-test                  |
| helix-vault       | `ErrConnectionRefused` (Vault HTTP)   | Toxiproxy partition                             |
| helix-shm         | RLIMIT_MEMLOCK exhaustion             | `prlimit --memlock=0` on the test process      |
| helix-iouring     | CQE overflow                           | drive submission rate above ring capacity       |
| helix-network     | Jitter ≥ 50 ms p99                     | Toxiproxy `latency`-toxic with jitter           |
| helix-network     | Loss rate 5 %                          | Toxiproxy `loss`-toxic                          |
| helix-encoder     | NVENC session-limit reached            | spawn N+1 sessions where N is the SKU cap      |
| helix-pipeline    | Per-stage failure (capture disconnect)| `xrandr --setoutput primary off`                |
| helix-record      | S3 partition during upload            | Toxiproxy partition between record and MinIO   |
| helix-thermal     | TJunction critical                     | `nvidia-smi --gpu-reset` synthetic temp ramp   |
| helix-transport   | Network partition                       | Toxiproxy + chaos-mesh `network-partition`     |

The full list lives across the 29 S05 §9.3 tables; `helix-chaos-coverage` ensures none is missed.

---

## 5. Anti-Pattern Catalogue

### 5.1 Chaos test that proves nothing

```go
func TestPartition(t *testing.T) {
    toxi.AddToxic("partition", "...")
    time.Sleep(5 * time.Second)
    toxi.RemoveToxic()
    // no assertions
}
```

A chaos test must **assert something** about the system's behaviour during + after the injection. Without assertions, a successful run says nothing.

### 5.2 Chaos with insufficient cleanup

```go
func TestLatency(t *testing.T) {
    toxi.AddToxic("latency", "...")
    runWorkload()
    // forgot to remove the toxic — next test inherits it
}
```

Always `t.Cleanup(func(){ toxi.RemoveToxic() })`. The next test in the suite is a different submodule; chaos contamination = false-positive failures elsewhere.

### 5.3 Chaos with global side effect

```go
func TestPartition(t *testing.T) {
    iptables -A OUTPUT -p tcp --dport 8200 -j DROP   // affects host, not container
}
```

Never modify the **host**; modify the **container** (or the per-test Toxiproxy proxy). Host-side iptables rules can outlive the test container and contaminate the next CI run on the same runner.

### 5.4 Long sleep in chaos test

```go
toxi.AddToxic("latency", "5s")
time.Sleep(60 * time.Second)   // 60-second wait
```

Long sleeps inflate CI cost. Use `eventually(func() bool { ... })` patterns — wait for the actual condition rather than a fixed timeout.

### 5.5 Chaos that depends on real internet

```go
chaos.PartitionFromInternet()   // forbidden — implies real-internet baseline
```

Same rule as E2E §6.2: tests are hermetic. Use `chaos-mesh` to partition between two pods inside the test cluster.

---

## 6. CI Lane Invocation Pattern

```yaml
- name: Chaos (nightly)
  if: matrix.test-type == 'chaos' && github.event_name == 'schedule'
  steps:
    - name: Start topology + Toxiproxy
      run: |
        cd tests/chaos
        docker compose up -d --wait
    - name: Run chaos-mesh experiments
      if: hashFiles('tests/chaos/chaos-mesh/*.yaml') != ''
      run: |
        kubectl apply -f tests/chaos/chaos-mesh/
        sleep 600  # let the experiments run their declared duration
        kubectl get podchaos,networkchaos,iochaos -A -o json > experiments.json
    - name: Run Go-side chaos tests
      run: go test -tags=chaos -timeout=20m ./tests/chaos/...
    - name: helix-chaos-coverage lint
      run: helix-chaos-coverage ./tests/chaos/ ../06_Submodules/per-submodule/
    - name: Tear down
      if: always()
      run: docker compose down -v
```

The lint at `helix-chaos-coverage` parses the §9.3 table of each consumed submodule + asserts every documented error has a corresponding chaos test.

---

## 7. Recovery Time Objectives (RTOs)

Per-failure-mode RTOs are documented per submodule. Excerpts:

| Failure mode                         | RTO (recovery time)         | Verification                       |
|--------------------------------------|-----------------------------|------------------------------------|
| Vault seal during active session     | ≤ 30 s after operator unseal| session resumes with new DEK       |
| S3 partition (helix-record)          | ≤ 5 s detection + retry     | local-buffer queue grows; drains on reconnect |
| Encoder OOM                           | ≤ 1 s session-error         | client receives `EncoderError`     |
| Network partition (transport)         | QUIC migration ≤ 1 s         | session continues post-migration    |
| Capture display-disconnect            | ≤ 500 ms re-init             | helix-pipeline re-acquires capture |

A failure-mode injection that exceeds its RTO blocks the merge.

---

## 8. Blast-Radius Containment

Chaos tests must not affect:

- The host's networking (iptables, routing).
- The host's filesystem outside the test container's volumes.
- Other test containers running concurrently.
- The CI runner's services (cron jobs, agent processes).

The `helix-chaos-blast-audit` tool monitors the host's namespaces during a chaos run + reports any out-of-container modification. A non-empty report fails CI.

---

## 8a. Sample Toxiproxy Test Pattern

```go
//go:build chaos
package transport_chaos_test

import (
    "context"
    "testing"
    "time"

    toxiproxy "github.com/Shopify/toxiproxy/v2/client"
)

func TestQUICRecoveryUnderPartition(t *testing.T) {
    toxi := toxiproxy.NewClient("localhost:8474")
    proxy, err := toxi.CreateProxy("vault", "0.0.0.0:18200", "vault:8200")
    if err != nil {
        t.Fatalf("create proxy: %v", err)
    }
    t.Cleanup(func() { _ = proxy.Delete() })

    // Baseline: no chaos. Sustain a session.
    sess := startSession(t, "localhost:18200")
    runFor(sess, 5*time.Second)
    if !sess.Healthy() {
        t.Fatal("baseline session unhealthy before chaos")
    }

    // Inject 5-second partition.
    tox, err := proxy.AddToxic("partition", "timeout", "downstream",
        1.0, toxiproxy.Attributes{"timeout": 5000})
    if err != nil {
        t.Fatalf("add toxic: %v", err)
    }
    t.Cleanup(func() { _ = proxy.RemoveToxic(tox.Name) })

    // Verify session detects the partition + queues locally.
    eventually(t, 1*time.Second, func() bool {
        return sess.LocalQueueGrowing()
    })

    // End partition.
    _ = proxy.RemoveToxic(tox.Name)

    // Verify QUIC migration: session resumes within 1 second.
    eventually(t, 1*time.Second, func() bool {
        return sess.Healthy() && sess.LocalQueueDrained()
    })
}
```

Properties:
- `t.Cleanup` removes the toxic regardless of test outcome (no contamination).
- `eventually` polls for the actual condition rather than `time.Sleep`.
- Baseline-state assertion before injection prevents false negatives (test passes only because pre-injection was already broken).

## 8b. chaos-mesh CRD Catalogue

For Kubernetes-orchestrated chaos, the canonical CRDs:

| CRD                  | What it does                                                         | Use case (HelixPlay)                                              |
|----------------------|----------------------------------------------------------------------|---------------------------------------------------------------------|
| `NetworkChaos`       | Drop / delay / corrupt / duplicate packets between pods             | helix-transport partition recovery                                  |
| `PodChaos`           | Kill a pod (`pod-kill`) or freeze (`pod-failure`)                   | helix-pipeline survives a co-located submodule's container restart |
| `IOChaos`            | Inject IO latency / errors at the syscall layer                      | helix-record fMP4 mux under disk-IO pressure                       |
| `TimeChaos`          | Skew the wall-clock                                                  | session token expiry + clock-jump scenarios                        |
| `StressChaos`        | CPU / memory pressure                                                | helix-thermal scaling triggers under sustained load                |
| `DNSChaos`           | DNS resolution chaos                                                 | OAuth provider outage simulation                                    |
| `HTTPChaos`          | Inject HTTP-level faults                                             | gRPC + REST middleware reaction                                    |
| `KernelChaos`        | Kernel-layer fault injection                                         | rare; reserved for canary cadence's kernel-bypass tests            |

Each CRD specification lives at `tests/chaos/chaos-mesh/<failure-mode>.yaml` per the directory layout in [Constitution §6](../01_Constitution.md#6-testing-discipline-r-11-r-12-r-13).

## 9. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-T07-A            | chaos-mesh adoption gate — do all chaos tests need k8s, or can Toxiproxy-only suffice?                        | `08_Operations/01_Container_CI_CD.md`              |
| OQ-T07-B            | Per-PR opt-in mechanism — PR-label vs CI matrix variable?                                                      | T10 Full Automation chapter                         |
| OQ-T07-C            | Time-chaos (clock skew) — applicable for which submodules?                                                     | T07 next revision                                   |

---

## 10. References & Anti-Bluff Verification

### 10.1 Internal

- [`01_Test_Matrix.md`](01_Test_Matrix.md) §6.6.
- [`../06_Submodules/per-submodule/`](../06_Submodules/per-submodule/) §9.3 (per-submodule failure-mode tables).
- [`../06_Submodules/02_Containers_Submodule.md`](../06_Submodules/02_Containers_Submodule.md) §3.4.
- [`../06_Submodules/03_Challenges_Submodule.md`](../06_Submodules/03_Challenges_Submodule.md) §5.4.

### 10.2 External (web)

- Toxiproxy: https://github.com/Shopify/toxiproxy (accessed 2026-04-30).
- chaos-mesh: https://chaos-mesh.org/ (accessed 2026-04-30).
- Principles of Chaos Engineering: https://principlesofchaos.org/ (accessed 2026-04-30).

### 10.3 Anti-Bluff Verification

- Coverage: chapter exceeds the 300-line floor.
- Forbidden patterns: clean. Three §9 open questions named with deferred resolution chapters.
- §5 anti-pattern catalogue + §8 blast-radius containment operationalise the Constitution §13 *Exceptions* boundary.

### 10.4 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30.
- Reviewed by: pending operator review.

End of `07_Testing/07_Chaos.md` — 2026-04-30.
