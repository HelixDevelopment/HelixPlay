# Phase_11 — Hardening & Security

> **Source dimensions:** [`Phase_10_Monetization_and_Auth.md`](Phase_10_Monetization_and_Auth.md), [`../06_Submodules/per-submodule/helix-r18-safeexec.md`](../06_Submodules/per-submodule/helix-r18-safeexec.md), [`../06_Submodules/per-submodule/helix-vault.md`](../06_Submodules/per-submodule/helix-vault.md), [`../03_Architecture/10_Security_and_Hardening.md`](../03_Architecture/10_Security_and_Hardening.md) (C11), [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row P11.
> **Source line count:** chapter floor 500 lines per Master Plan §7.2 row P11.
> **Phase targets:** R-13 + R-18 (Operational Integrity) + R-12 (Security testing) + Constitution §11.5 forbidden-commands list zero-tolerance.
> **Cross-links:** [`Phase_12_Beta_Launch.md`](Phase_12_Beta_Launch.md), [`Phase_10_Monetization_and_Auth.md`](Phase_10_Monetization_and_Auth.md).
> **Status:** Draft v1 specification.
> **Last updated:** 2026-04-30.

---

## 1. Phase Scope

Phase_11 is the **security-hardening pass** before beta. Every prior phase optimized for getting things working; Phase_11 audits + locks down + verifies + signs off the system as **production-grade**:

- **R-18 audit:** Every subprocess invocation across the full codebase wraps in helix-r18-safeexec; the helix-r18-safeexec-vet lint runs in CI; zero `os/exec` direct calls survive.
- **Constitution §11.5 forbidden-commands sweep:** Trivy / Semgrep / custom regex scans confirm no forbidden command (shutdown / poweroff / reboot / suspend / hibernate / lock / kill -9 / etc.) exists in any source file, container image, or runtime path.
- **Penetration test:** Operator-contracted external pentest firm runs against the full stack; findings tracked + remediated + re-tested.
- **OWASP ASVS Level 2 conformance:** Application Security Verification Standard checks pass.
- **CIS Benchmarks:** Every container image conforms to CIS Docker Benchmark + CIS Kubernetes Benchmark.
- **STIG / FIPS 140-3:** Optional operator-pinned compliance profiles for government / regulated-industry deployments.
- **Cosign + SLSA Level 3:** Every released artifact is cosign-signed + SLSA L3-attested + recorded in the operator's Rekor instance.
- **Secret rotation:** Per-tenant secrets + per-service secrets + cluster-bootstrap secrets all rotated; rotation cadence documented + enforced via helix-vault.
- **Network segmentation:** Cilium NetworkPolicy enforced at L3/L4 + L7; tenant-isolation verified.
- **Vulnerability remediation:** Snyk + govulncheck + Trivy + grype scans run; CRITICAL + HIGH CVEs all remediated; MEDIUM tracked + remediated within 30 days.

After Phase_11, the operator can present an **external auditor** with a complete security dossier: pentest report, ASVS conformance attestation, CIS benchmark scores, SLSA L3 attestations, secret-rotation certificates, network-segmentation proof.

---

## 2. Prerequisites

- Phase_10 complete + signed off — auth + billing operational.
- helix-r18-safeexec at v1.0.0 + helix-r18-safeexec-vet linter integrated.
- Operator-provisioned: Snyk subscription, SonarQube cloud or self-hosted, external pentest firm contracted.
- Cosign + Rekor + SLSA verifier deployed per O01.
- helix-vault with KEK rotation policy active.

---

## 3. Tasks Catalogue

| Task ID    | Task                                                                | Subtasks |
|------------|---------------------------------------------------------------------|---------:|
| P11.T01   | helix-r18-safeexec full-codebase audit + zero-tolerance lint enforce | 5        |
| P11.T02   | Constitution §11.5 forbidden-commands sweep (source + image + runtime) | 5      |
| P11.T03   | OWASP ASVS Level 2 conformance audit                                 | 6        |
| P11.T04   | CIS Docker Benchmark + CIS Kubernetes Benchmark                      | 5        |
| P11.T05   | External pentest engagement — full stack                             | 5        |
| P11.T06   | Snyk + govulncheck + Trivy + grype CVE remediation                   | 5        |
| P11.T07   | SonarQube quality gate — A grade across all submodules               | 4        |
| P11.T08   | Cosign + SLSA L3 attestation for every released artifact             | 5        |
| P11.T09   | Per-tenant + per-service + cluster secrets rotation drill            | 4        |
| P11.T10   | Cilium NetworkPolicy L3/L4 + L7 enforcement + tenant-isolation       | 5        |
| P11.T11   | TLS 1.3-only + ECH + post-quantum hybrid (Kyber + X25519) verification | 4      |
| P11.T12   | DDoS resilience — operator-tunable rate limits + Cloudflare/Coturn integration | 4 |
| P11.T13   | Container image hardening — distroless + non-root + read-only rootfs | 4        |
| P11.T14   | RBAC + workload-identity audit (Kubernetes ServiceAccount review)    | 4        |
| P11.T15   | Audit log integrity verification — chain-of-custody check            | 4        |
| P11.T16   | Incident response runbook — on-call playbook + tabletop drill        | 5        |
| P11.T17   | STIG / FIPS 140-3 optional compliance profile activation             | 4        |
| P11.T18   | End-to-end security smoke — full pentest re-run after remediation    | 5        |
| P11.T19   | Phase_11 acceptance review                                            | 2        |

19 tasks; ~85 subtasks.

---

## 4. Task Details

### 4.1 P11.T01 — helix-r18-safeexec full-codebase audit

helix-r18-safeexec-vet linter scans every Go file in every submodule + the host agent + the clients. Zero tolerance: any direct `os/exec.Command`, `exec.CommandContext`, `syscall.Exec`, `syscall.ForkExec` call → CI fail. Exception list documented + cosign-signed by operator's compliance officer.

**Subtasks:**

- T01.S01 — Linter rules per [helix-r18-safeexec §3](../06_Submodules/per-submodule/helix-r18-safeexec.md).
- T01.S02 — Full-codebase scan; remediate every direct exec call.
- T01.S03 — CI lane integration (T07.G fail-closed on direct-exec match).
- T01.S04 — Exception register (operator-mediated; compliance-officer-signed).
- T01.S05 — Re-scan after every release for regression detection.

### 4.2 P11.T02 — Forbidden-commands sweep

Constitution §11.5 lists forbidden commands (shutdown / poweroff / reboot / suspend / hibernate / lock / kill -9 / etc.). Sweep:

- Source code (Semgrep custom regex rules).
- Container images (Trivy custom rules).
- Runtime processes (host-side eBPF probe via helix-rtos).

**Subtasks:**

- T02.S01 — Source-code Semgrep rules (operator-managed rule pack).
- T02.S02 — Container-image Trivy rules.
- T02.S03 — Runtime eBPF probe (logs every `execve` syscall against the forbidden list).
- T02.S04 — Audit log integration with operator's SIEM.
- T02.S05 — Quarterly sweep + remediation cycle.

### 4.3 P11.T03 — OWASP ASVS Level 2

OWASP Application Security Verification Standard Level 2 — 286 checks across V1-V14 chapters. Operator's security engineer documents per-check conformance.

**Subtasks:**

- T03.S01 — V1 Architecture, Design, and Threat Modeling — threat model document published.
- T03.S02 — V2 Authentication — OAuth 2.1 + MFA conformance.
- T03.S03 — V3 Session Management — session lifecycle conformance.
- T03.S04 — V4-V7 Access Control + Validation + Cryptography + Errors — case-by-case.
- T03.S05 — V8-V14 Data + Comm + Malicious + Business Logic + Files + API + Config — case-by-case.
- T03.S06 — Per-check evidence + remediation + sign-off.

### 4.4 P11.T04 — CIS Docker + CIS Kubernetes Benchmark

CIS Docker Benchmark v1.6.0 + CIS Kubernetes Benchmark v1.9.0 — every container image + every Kubernetes manifest scored.

**Subtasks:**

- T04.S01 — kube-bench scan against every cluster.
- T04.S02 — docker-bench-security against every container image.
- T04.S03 — Per-finding remediation (CIS findings categorized: must-fix vs. operator-exempt).
- T04.S04 — CIS score export to operator's compliance dashboard.
- T04.S05 — Quarterly re-scan.

### 4.5 P11.T05 — External pentest

Operator contracts an external pentest firm (recommended: Trail of Bits, Doyensec, NCC Group, or equivalent — operator's choice). Pentest scope: full stack including host agent, clients, backend services, payment gateway integration.

**Subtasks:**

- T05.S01 — Scoping + rules of engagement document.
- T05.S02 — Pentest execution (≥ 4 weeks).
- T05.S03 — Findings triage — CRITICAL + HIGH must-fix; MEDIUM track + remediate; LOW operator-discretion.
- T05.S04 — Remediation — engineering team fixes each must-fix finding.
- T05.S05 — Re-test — pentest firm verifies remediation.

### 4.6 P11.T06 — CVE remediation

Snyk + govulncheck + Trivy + grype scans run nightly + on every PR. CRITICAL + HIGH must-fix; MEDIUM track + remediate within 30 days; LOW operator-discretion.

**Subtasks:**

- T06.S01 — Snyk integration (per-PR + nightly).
- T06.S02 — govulncheck integration (Go-specific; nightly).
- T06.S03 — Trivy integration (container images; per-build).
- T06.S04 — grype integration (SBOM-based; per-release).
- T06.S05 — CVE dashboard — operator-visible burndown.

### 4.7 P11.T07 — SonarQube quality gate

SonarQube cloud or self-hosted; quality gate: A grade (no CRITICAL bugs, no security hotspots, ≤ 3% code duplication, ≥ 80% test coverage).

**Subtasks:**

- T07.S01 — SonarQube project per submodule.
- T07.S02 — Quality gate configuration (operator-managed).
- T07.S03 — Per-PR scan + fail-closed on quality-gate failure.
- T07.S04 — Quality-gate trend reporting (monthly).

### 4.8 P11.T08 — Cosign + SLSA L3

Every released artifact (binary + container image + Helm chart + recording) is cosign-signed + SLSA L3-attested + recorded in operator's Rekor.

**Subtasks:**

- T08.S01 — cosign sign + cosign verify integration in CI.
- T08.S02 — SLSA L3 build-environment isolation (per O01 §10).
- T08.S03 — Rekor transparency log integration.
- T08.S04 — Verification tooling for downstream consumers.
- T08.S05 — Key rotation + revocation workflows.

### 4.9 P11.T09 — Secret rotation drill

helix-vault.RotationService rotates every secret on a documented cadence:

- KEK: annual.
- DEK: per-recording-session.
- TLS certs: 90 days (Let's Encrypt) or operator-pinned CA cadence.
- API keys (per-tenant + per-service): 90 days.
- Database passwords: 60 days.

Rotation drill — operator runs the rotation in a staging environment + verifies no service interruption.

**Subtasks:**

- T09.S01 — Rotation schedule documented per [helix-vault §4](../06_Submodules/per-submodule/helix-vault.md).
- T09.S02 — Per-secret rotation runbook.
- T09.S03 — Staging-environment rotation drill.
- T09.S04 — Production rotation cadence enforced.

### 4.10 P11.T10 — Cilium NetworkPolicy

Cilium L3/L4 + L7 NetworkPolicy enforced. Per-tenant network namespace isolation. Cross-tenant traffic denied at the policy layer + verified via traffic capture.

**Subtasks:**

- T10.S01 — Per-tenant NetworkPolicy generated from helix-tenant config.
- T10.S02 — Default-deny baseline (only allow-listed flows pass).
- T10.S03 — L7 HTTP policy (per-route + per-method).
- T10.S04 — Hubble flow-log integration.
- T10.S05 — Cross-tenant traffic denial Challenges scenario.

### 4.11 P11.T11 — TLS 1.3-only + ECH + PQ-hybrid

Every TLS endpoint enforces TLS 1.3-only (TLS 1.2 disabled). ECH (Encrypted Client Hello) for SNI privacy. Post-quantum hybrid key exchange (Kyber768 + X25519) per [C11 §6.3](../03_Architecture/10_Security_and_Hardening.md).

**Subtasks:**

- T11.S01 — TLS 1.3-only enforcement at all endpoints.
- T11.S02 — ECH configuration (operator's authoritative DNS publishes ECH config).
- T11.S03 — Kyber768 + X25519 hybrid via helix-tls primitive.
- T11.S04 — TLS configuration audit via testssl.sh + nmap-ssl-enum-ciphers.

### 4.12 P11.T12 — DDoS resilience

Operator-tunable rate limits (per-IP + per-tenant + per-method). Coturn for STUN/TURN traffic isolation. Optional Cloudflare integration for L4 + L7 DDoS scrubbing.

**Subtasks:**

- T12.S01 — Per-tier rate limits (anonymous → authenticated → tenant.user → tenant.admin).
- T12.S02 — Coturn deployment behind operator's firewall.
- T12.S03 — Cloudflare integration (operator-mediated; requires DNS to point to Cloudflare).
- T12.S04 — DDoS Challenges scenario — synthetic L7 flood + verify rate-limit rejection.

### 4.13 P11.T13 — Container image hardening

Every container image — distroless base + non-root user + read-only rootfs + dropped Linux capabilities. Per O01 §6.

**Subtasks:**

- T13.S01 — Distroless base (gcr.io/distroless/static or operator-mirrored equivalent).
- T13.S02 — Non-root user (UID 65534 default; operator-overridable).
- T13.S03 — Read-only rootfs + writable tmpfs for ephemeral state.
- T13.S04 — Linux capabilities dropped to bare minimum (NET_BIND_SERVICE only for services binding < 1024).

### 4.14 P11.T14 — Kubernetes RBAC + workload identity

Per-ServiceAccount review — every Kubernetes ServiceAccount has minimum-privilege RBAC. Workload identity (Kubernetes ServiceAccount tokens projected into pods + bound to vault per [helix-vault §5](../06_Submodules/per-submodule/helix-vault.md)).

**Subtasks:**

- T14.S01 — Per-ServiceAccount RBAC review + remediation.
- T14.S02 — Workload identity bound to helix-vault.
- T14.S03 — ServiceAccount token rotation (per Kubernetes BoundServiceAccountTokenVolume).
- T14.S04 — Audit log of every ServiceAccount-mediated API call.

### 4.15 P11.T15 — Audit log integrity

helix-billing.AuditLog + helix-tenant.AuditLog + helix-vault.AuditLog — every audit log has cosign-signed per-event entries + Merkle-chained sequence. Verify integrity end-to-end.

**Subtasks:**

- T15.S01 — Cosign-verify every audit log entry.
- T15.S02 — Chain-of-custody check (every event references prior).
- T15.S03 — Daily Merkle-root publication + operator-counter-signed.
- T15.S04 — Tamper detection alarm.

### 4.16 P11.T16 — Incident response runbook + tabletop drill

`HelixDevelopment/HelixOps/docs/runbook/incident-response.md` — on-call playbook covering: secret leak, container compromise, payment-gateway breach, GDPR data leak, R-18 violation. Tabletop drill once per quarter.

**Subtasks:**

- T16.S01 — Secret-leak playbook.
- T16.S02 — Container-compromise playbook.
- T16.S03 — Payment-gateway-breach playbook.
- T16.S04 — GDPR data-leak playbook.
- T16.S05 — Quarterly tabletop drill — operator's on-call team runs scripted scenarios.

### 4.17 P11.T17 — STIG / FIPS 140-3 (optional)

Operator-pinned compliance profile for government / regulated-industry deployments. STIG (Security Technical Implementation Guide) profile + FIPS 140-3 cryptographic module conformance.

**Subtasks:**

- T17.S01 — STIG-hardened OS image (operator's choice — RHEL STIG, Ubuntu STIG, etc.).
- T17.S02 — FIPS 140-3 cryptographic module (Boring Crypto or operator-pinned validated module).
- T17.S03 — FIPS-mode runtime verification (every TLS handshake uses validated cipher).
- T17.S04 — STIG / FIPS attestation document.

### 4.18 P11.T18 — End-to-end security smoke

After all remediation lands, the external pentest firm runs a re-test. Findings remediated; re-test green; final security report cosign-signed by pentest firm.

**Subtasks:**

- T18.S01 — Re-test scoping document.
- T18.S02 — Re-test execution.
- T18.S03 — Findings remediation (any new findings).
- T18.S04 — Final report.
- T18.S05 — Cosign-signed final report archived.

### 4.19 P11.T19 — Phase_11 acceptance review

Operator + compliance officer + CISO signoff per Constitution §16 + the §6 exit criteria.

---

## 5. Subtask Catalogue

85 subtasks across 19 tasks; bulk-imported.

---

## 6. Exit Criteria

- [ ] helix-r18-safeexec full-codebase audit clean.
- [ ] Constitution §11.5 forbidden-commands sweep clean.
- [ ] OWASP ASVS Level 2 conformance attested.
- [ ] CIS Docker + CIS Kubernetes Benchmark scores documented.
- [ ] External pentest report — all CRITICAL + HIGH remediated + re-tested.
- [ ] Snyk + govulncheck + Trivy + grype CVE burndown — all CRITICAL + HIGH remediated.
- [ ] SonarQube quality gate A grade across all submodules.
- [ ] Cosign + SLSA L3 attestation on every released artifact.
- [ ] Secret rotation drill executed in staging.
- [ ] Cilium NetworkPolicy + tenant isolation verified.
- [ ] TLS 1.3-only + ECH + PQ-hybrid verified.
- [ ] DDoS resilience tested.
- [ ] Container images distroless + non-root + read-only rootfs.
- [ ] Kubernetes RBAC + workload identity audited.
- [ ] Audit log integrity verified end-to-end.
- [ ] Incident response runbook + tabletop drill completed.
- [ ] STIG / FIPS 140-3 (operator-optional) attested if pinned.
- [ ] End-to-end security smoke passes.
- [ ] Operator + compliance officer + CISO signoff per Constitution §16.

---

## 7. Risk Register

| ID       | Risk                                                                | Mitigation                                                            |
|----------|----------------------------------------------------------------------|------------------------------------------------------------------------|
| RP11-01  | External pentest reveals CRITICAL finding too late for beta launch  | Pentest scheduled 8 weeks before beta to allow remediation buffer.    |
| RP11-02  | helix-r18-safeexec linter false-positive blocks legitimate code     | Exception register + compliance-officer sign-off process.             |
| RP11-03  | Forbidden-command found in dependency (third-party Go module)       | Module vendoring + cosign-signed allow-list per [O02](../08_Operations/02_Build_System.md). |
| RP11-04  | CVE remediation breaks API compatibility                            | Per-PR API-contract test (T07.B) catches regressions.                 |
| RP11-05  | Cosign key compromise                                               | Operator's key rotation policy + Sigstore-Fulcio short-lived certs.   |
| RP11-06  | Cilium NetworkPolicy mis-configuration causes service outage         | Staged rollout + per-namespace canary.                                |
| RP11-07  | Kyber768 + X25519 hybrid not interoperable with all clients         | Fallback to X25519-only after grace period; per-client capability negotiation. |
| RP11-08  | DDoS scrubbing introduces unacceptable latency                      | Per-region scrubbing decision + bypass for Phase_07-grade latency.     |

---

## 8. Cross-Family Dependencies

- C11 §6 (Security & Hardening) — architecturally specifies the patterns.
- helix-r18-safeexec, helix-vault, helix-tenant — primitives for security enforcement.
- O02 (Build System) — cosign + SLSA integration.
- O04 (Observability) — audit log shipping.
- T07 (Security testing) — CI integration of Snyk + govulncheck + Trivy + grype.

---

## 9. Acceptance Criteria

Constitution §16 signoff (operator + compliance officer + CISO) + §6 exit criteria.

---

## 10. The Phase_11 Calendar

~ 8 weeks. Operator-side capacity: 5 engineers (2× security engineers for pentest engagement + remediation; 2× backend engineers for hardening pass; 1× ops engineer for runbook + drill).

The Phase_11 calendar is **non-compressible** — external pentest engagement timing is operator-side commercial agreement-driven.

---

## 11. Anti-Bluff Verification

### 11.1 Sources resolved

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`Phase_10_Monetization_and_Auth.md`](Phase_10_Monetization_and_Auth.md) | 380+ | 2026-04-30 | predecessor                                      |
| [`../03_Architecture/10_Security_and_Hardening.md`](../03_Architecture/10_Security_and_Hardening.md) | 2,200+ | 2026-04-30 | C11 architectural source            |
| [`../06_Submodules/per-submodule/helix-r18-safeexec.md`](../06_Submodules/per-submodule/helix-r18-safeexec.md) | 343 | 2026-04-30 | R-18 wrapper       |
| [`../06_Submodules/per-submodule/helix-vault.md`](../06_Submodules/per-submodule/helix-vault.md) | 290+ | 2026-04-30 | secrets primitive   |
| [`../01_Constitution.md`](../01_Constitution.md)                  | 1,400+ | 2026-04-30 | §11 + §11.5 + §16                              |

### 11.2 Forbidden patterns

Clean.

### 11.3 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 (specification only).
- Pending: Phase_11 execution + operator + compliance officer + CISO signoff.

End of `09_Implementation_Phases/Phase_11_Hardening_and_Security.md` — 2026-04-30.
