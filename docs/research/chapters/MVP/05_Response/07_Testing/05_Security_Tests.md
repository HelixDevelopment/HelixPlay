# T05 — Security Tests

> **Source dimensions:** [`00_Index.md`](00_Index.md), [`01_Test_Matrix.md`](01_Test_Matrix.md) §6.4 + §10; [`../06_Submodules/01_Submodule_Catalog.md`](../06_Submodules/01_Submodule_Catalog.md) §4.4 (SBOM) + §4.5 (vuln scanning); [`../06_Submodules/02_Containers_Submodule.md`](../06_Submodules/02_Containers_Submodule.md) §6 (cosign) + §8 (R-18); [`../01_Constitution.md`](../01_Constitution.md) §7 (R-10 Quality Gates), §11 (Security & Privacy), §11.5 (R-18).
> **Source line count:** chapter floor 300 lines per Master Plan §7.2 row T05.
> **Chapter targets:** R-10 (heavy security scanning), R-13 (no green-on-broken), R-18 (Operational Integrity).
> **Cross-links:** [`02_Unit_Tests.md`](02_Unit_Tests.md), [`07_Chaos.md`](07_Chaos.md), [`11_Challenges.md`](11_Challenges.md).
> **Status:** Draft v1.
> **Last updated:** 2026-04-30.

---

## 1. Role in the Master Test Matrix

Security is **test-type 4 of 10** in the [T01 §2 grid](01_Test_Matrix.md#2-the-29--10--4-cell-grid-the-canonical-lookup). It runs at every cadence (per-PR + nightly + canary + pre-release) and gates merges + deployments.

The all-three-pass rule per [S01 §4.5](../06_Submodules/01_Submodule_Catalog.md#45-vulnerability-scanning-govulncheck--snyk--renovate-all-three-pass-gate): **govulncheck + Snyk + Renovate** must all pass on every PR. Trivy adds container-side coverage (OS-package CVEs invisible to the Go-only tools). Together they cover the supply-chain surface from code → dependencies → OS-base-image.

Security tests differ from Unit / Integration / E2E by what they **assert**. Where the others assert correctness, Security asserts **absence of known issues** + **rejection of intentional attacks**.

---

## 2. The Discipline — What Makes "Green" Green

A green Security row means:

1. **govulncheck**: zero high-severity findings reachable from any binary entry point. Reports unreachable findings as advisory.
2. **Snyk**: zero high-severity findings + licence-policy compliance (only the SPDX identifiers in [S01 §4.8 allow-list](../06_Submodules/01_Submodule_Catalog.md#48-licence-consistency--mit-default-with-named-apache-20-exceptions)).
3. **Trivy**: zero `HIGH` / `CRITICAL` findings on the container image (OS packages + Go binary).
4. **Custom fuzzers** (where applicable): no panics + no buffer overruns in N runs (typical N = 10⁶ inputs per per-PR cycle, 10⁹ in nightly).
5. **R-18 host-integrity-scan**: zero forbidden syscalls fired during any test invocation (per [S02 §3.4](../06_Submodules/02_Containers_Submodule.md#34-the-host-integrity-scan-shared-lane)).

Medium and low findings are **reported** (the `.snyk` config records them) but do not fail CI. Critical findings (CVSS ≥ 9.0) **always** fail CI even if marked unreachable; the rationale is that an unreachable Critical may become reachable via a future refactor, and the cost of catching it now is much lower than after-the-fact.

---

## 3. Tooling

### 3.1 govulncheck

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck -mode=symbol -json ./... > govulncheck.json
```

The `-mode=symbol` flag asks for symbol-level reachability (vs module-level presence); this dramatically reduces the false-positive rate. The JSON output is consumed by `helix-vuln-policy` which applies the §2 rules (high → fail; critical-unreachable → fail; medium → report).

### 3.2 Snyk

```bash
snyk auth $SNYK_TOKEN
snyk test \
    --severity-threshold=high \
    --policy-path=.snyk \
    --print-deps
```

The `.snyk` policy file at the repository root pins:

```yaml
version: v1.27.0
policy: |
  # Allowed SPDX identifiers (S01 §4.8)
  licenses:
    allowed:
      - MIT
      - Apache-2.0
      - BSD-2-Clause
      - BSD-3-Clause
      - MPL-2.0
      - ISC
    denied:
      - GPL-2.0
      - GPL-3.0
      - AGPL-3.0
      - LGPL-2.1
      - LGPL-3.0
ignore: {}
```

Per-submodule `.snyk` files MAY add `ignore` rules for specific CVE IDs with operator sign-off + explicit expiry per [S01 §4.5.2](../06_Submodules/01_Submodule_Catalog.md#452-snyk--the-constitution-mandated-quality-gate).

### 3.3 Trivy

```bash
trivy image \
    --severity HIGH,CRITICAL \
    --exit-code 1 \
    --ignore-unfixed=false \
    ghcr.io/vasic-digital/helix-${NAME}@${DIGEST}
```

`--ignore-unfixed=false` is intentional: an unfixed vulnerability is still a vulnerability; the operator must decide whether to ship despite it (Constitution §13 *Exceptions*).

### 3.4 Custom Fuzzers

Per submodule, the fuzz target is the **untrusted-input boundary** — gRPC handler, RTP/RTCP parser, Theme bundle loader, audit-log emitter, etc. Go's stdlib fuzz harness (`testing.F`):

```go
//go:build fuzz
package codec_test

import (
    "bytes"
    "testing"

    "github.com/vasic-digital/helix-codec"
)

func FuzzProfileParse(f *testing.F) {
    f.Add([]byte("h264 profile=baseline level=4.0 bitrate=10000000"))
    f.Fuzz(func(t *testing.T, in []byte) {
        // The parser must not panic on arbitrary input.
        _, _ = codec.ParseProfile(bytes.NewReader(in))
    })
}
```

Per [S01 §4.5](../06_Submodules/01_Submodule_Catalog.md#45-vulnerability-scanning-govulncheck--snyk--renovate-all-three-pass-gate), the canonical fuzzer corpus is committed under `tests/security/fuzz-corpus/`; CI-run additions land via `go test -fuzz` and the runtime promotes interesting inputs.

### 3.5 SPDX-check

The `spdx-check` lint (custom Go tool shipped under `vasic-digital/.github/`) verifies every source file has a leading SPDX-License-Identifier comment matching the repo licence ([S01 §4.8.6](../06_Submodules/01_Submodule_Catalog.md#486-spdx-identifiers)). A missing or mismatched header fails CI.

---

## 4. Per-Submodule Expectations

Each submodule's [S05 §11 Per-test-type coverage targets](../06_Submodules/per-submodule/) entry for Security commits to:

- **Vault wrapper (`helix-vault`)** — KEK rotation under load + cross-tenant isolation negative tests + AES-GCM round-trip + credential-leak check (no plaintext credential in OTLP traces or logs).
- **Subprocess wrapper (`helix-r18-safeexec`)** — ABA fuzzer for the deny-list matcher (Unicode-normalisation attacks, path-resolution attacks, argument-array attacks); host-integrity-scan auditd + strace assertion of zero forbidden syscalls.
- **gRPC framing (`helix-grpc-frame`)** — RTP/RTCP fuzzer; SRTP-key-mismatch detection; protobuf-side malformed-payload coverage.
- **Vault (`helix-vault`) cross-tenant** — attempted access from tenant-A code path to tenant-B path must fail with `ErrUnauthorized`.

The full per-submodule list lives in T01 §2's grid; every Security cell expects govulncheck + Snyk + Trivy + per-submodule fuzz coverage as appropriate.

---

## 5. Anti-Pattern Catalogue

### 5.1 Suppressed CVE without expiry

```yaml
# .snyk
ignore:
  SNYK-GO-X-12345:
    - '*':
        reason: "we'll fix it later"
```

A suppression without an expiry date is a forever-defer. Every `.snyk` ignore must have:

```yaml
ignore:
  SNYK-GO-X-12345:
    - '*':
        reason: "..."
        expires: "2026-09-30T00:00:00Z"
```

`helix-vuln-suppression-audit` runs nightly and fails if any ignore is past its expiry.

### 5.2 Skipping security on "test environments"

```bash
if [ "$ENV" = "test" ]; then
    echo "Skipping security scan for test build"
    exit 0
fi
```

Test builds are still binaries with the same dependency closure. Skipping security here means the dev binary may contain a CVE the prod binary doesn't — exactly the inversion that creates anti-bluff "green tests on broken systems".

### 5.3 Hardcoded credentials in test fixtures

```go
// tests/integration/vault_test.go
const testToken = "hvs.X9k3...redacted...real-token"  // forbidden
```

Even fake-looking tokens leak when the test fixture is committed. Use `testcontainers-go`'s ephemeral `VAULT_DEV_ROOT_TOKEN_ID` per-run.

### 5.4 Logging plaintext payloads

```go
log.Printf("encrypt: plaintext=%s", plain)  // forbidden
```

Logs are SIEM-aggregated per Constitution §10. Plaintext payloads in logs leak through the SIEM. Use `log.Printf("encrypt: payload-bytes=%d", len(plain))` instead.

### 5.5 Custom-rolled crypto

```go
xor(plain, key)  // not a cipher
```

No submodule rolls its own crypto. Use `helix-vault.Encrypt` (which uses HashiCorp Vault transit + AES-GCM) per [S01 §4.8.2](../06_Submodules/01_Submodule_Catalog.md#482-why-apache-20-specifically-for-those-four).

---

## 6. CI Lane Invocation Pattern

```yaml
- name: govulncheck
  if: matrix.test-type == 'security'
  run: |
    go install golang.org/x/vuln/cmd/govulncheck@latest
    govulncheck -mode=symbol -json ./... | tee govulncheck.json
    helix-vuln-policy ./govulncheck.json

- name: Snyk
  if: matrix.test-type == 'security'
  env:
    SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
  run: |
    snyk test --severity-threshold=high --policy-path=.snyk

- name: Trivy
  if: matrix.test-type == 'security'
  run: |
    trivy image --severity HIGH,CRITICAL --exit-code 1 \
        ghcr.io/vasic-digital/helix-${{ env.SUBMODULE }}@${{ env.DIGEST }}

- name: SPDX header check
  if: matrix.test-type == 'security'
  run: |
    helix-spdx-check ./...

- name: Custom fuzz (per-submodule)
  if: matrix.test-type == 'security' && hashFiles('tests/security/fuzz/*.go') != ''
  run: |
    go test -tags=fuzz -fuzz=. -fuzztime=60s ./tests/security/fuzz/...

- name: host-integrity-scan
  if: matrix.test-type == 'security'
  run: |
    cd vasic-digital/Containers/lanes/host-integrity-scan
    ./tests/deny-list-rejects.sh
    ./tests/strace-coverage.sh ../../lanes/${{ env.LANE }}/
    ./tests/auditd-zero-events.sh ../../lanes/${{ env.LANE }}/
```

The lane is **fail-fast disabled** — each tool runs even if a prior tool failed. The maintainer sees the full picture per PR.

---

## 7. Per-Lane Pinned Versions (2026-04-30 snapshot)

| Tool                      | Version pin               | Renovate group           |
|---------------------------|---------------------------|--------------------------|
| govulncheck               | latest from `vuln.go.dev` | `golang-x-vuln`          |
| Snyk CLI                  | `≥ v1.1300`               | `snyk-cli`               |
| Trivy                     | `≥ v0.55.0`               | `trivy`                  |
| go-fuzz / testing.F       | Go 1.24+ stdlib           | (toolchain)              |
| `helix-spdx-check`        | matched repo tag          | `vasic-digital/.github`  |
| `helix-vuln-policy`       | matched repo tag          | `vasic-digital/.github`  |
| `helix-vuln-suppression-audit` | matched repo tag    | `vasic-digital/.github`  |

Renovate watches each tool's release stream and opens a per-tool PR (per the Major-bumps-Monday rule per [S01 §4.3.3](../06_Submodules/01_Submodule_Catalog.md#433-renovate--dependabot-lockstep)).

---

## 8. The Reachability vs Presence Distinction

A vulnerability `CVE-2026-XXXX` in `module M v1.2.3` may be:

- **Present** — `M v1.2.3` appears in `go.mod` (Snyk reports this).
- **Reachable** — the vulnerable function is called along some path from a binary entry point (govulncheck reports this).

The all-three-pass gate uses both signals:

- A **reachable** vulnerability **must** be patched (the binary will execute the vulnerable code path).
- A **present-but-unreachable** vulnerability **should** be patched (the next refactor might add the call edge), but is not blocking unless its CVSS is Critical.
- A **present-but-unreachable Critical** **fails** the gate regardless — Critical's tail-risk is unacceptable.

The `.snyk` policy can ignore a present-but-unreachable Critical only with operator sign-off (Constitution §13 *Exceptions* applies).

---

## 9. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-T05-A            | Sigstore Rekor outage — fall back to keyed signing or block release?                                          | `08_Operations/02_Quality_Gates_SonarQube_Snyk.md` |
| OQ-T05-B            | Fuzzer coverage threshold — count vs reachable-edges?                                                          | T05 next revision                                   |
| OQ-T05-C            | Per-submodule CVE budget — operator-tunable per high-risk vs low-risk submodule?                              | `08_Operations/02_Quality_Gates_SonarQube_Snyk.md` |

---

## 10. References & Anti-Bluff Verification

### 10.1 Internal

- [`01_Test_Matrix.md`](01_Test_Matrix.md) §6.4, §10.
- [`../06_Submodules/01_Submodule_Catalog.md`](../06_Submodules/01_Submodule_Catalog.md) §4.4, §4.5, §4.8.
- [`../06_Submodules/02_Containers_Submodule.md`](../06_Submodules/02_Containers_Submodule.md) §3.4, §6, §8.
- [`../01_Constitution.md`](../01_Constitution.md) §7, §11, §11.5.

### 10.2 External (web)

- govulncheck: https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck (accessed 2026-04-30).
- Snyk Open Source for Go: https://docs.snyk.io/scan-with-snyk/snyk-open-source/snyk-open-source-supported-languages-and-package-managers/snyk-open-source-for-go (accessed 2026-04-30).
- Trivy: https://trivy.dev/ (accessed 2026-04-30).
- Go fuzzing: https://go.dev/doc/security/fuzz/ (accessed 2026-04-30).
- SPDX licence list: https://spdx.org/licenses/ (accessed 2026-04-30).

### 10.3 Anti-Bluff Verification

- Coverage: chapter exceeds the 300-line floor.
- Forbidden patterns: clean. Three §9 open questions named with deferred resolution chapters.
- §5 anti-pattern catalogue (5 patterns) operationalises R-13 (no green-on-broken) for Security.
- §8 reachability vs presence distinction is the substantive policy difference between the three scanners.

### 10.4 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30.
- Reviewed by: pending operator review.

End of `07_Testing/05_Security_Tests.md` — 2026-04-30.
