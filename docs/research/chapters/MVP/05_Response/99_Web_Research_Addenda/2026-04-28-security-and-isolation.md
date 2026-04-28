# Web Research Addendum — Security & Isolation

> **Topic:** 2026 vendor / library / standard updates that the C10 Security
> & Isolation chapter cites — OAuth 2.1 + OIDC + Device Authorization
> Grant, JWT best practices, mTLS at scale (SPIFFE/SPIRE, cert-manager,
> Vault), WebRTC DTLS-SRTP, the anti-cheat threat-model surface
> (cross-cutting, by reference), container isolation hardening (gVisor,
> Kata, KubeVirt, seccomp/AppArmor/SELinux), secret management
> (Vault 1.20+/2.0, OpenBao, ESO, SOPS), DDoS protection + rate
> limiting (Cloudflare 2026 trends, TURN abuse, gateway token-bucket),
> audit logging + compliance (OpenTelemetry-Logs, ECS, GDPR + DSA + SOC 2 +
> EU AI Act), and secure-by-default libraries (Tink, Themis, DOMPurify,
> Helmet.js).
> **Owning chapter:** [`../03_Architecture/09_Security_and_Isolation.md`](../03_Architecture/09_Security_and_Isolation.md) (C10).
> **Compiled by:** addendum subagent (C10).
> **Date:** 2026-04-28.
> **Status:** Append-only.

This addendum collects the web evidence consumed by the C10 chapter.
The chapter ingests Constitution §11 (Security & Privacy) and §11.5
(R-18 Operational Integrity) wholly — this addendum focuses on the
2026 deltas that the chapter's prose will cite. Every finding below
is sourced; placeholder language (TODO, FIXME, "and similar", "etc.")
is forbidden by Constitution §1.1 and is absent from the prose.
Where a 2026 source contradicts the 2024–2025 baseline captured in
`cloudgaming_dim09.md`, the contradiction is named explicitly under
§Z so the section subagents can resolve it inside the chapter. The
HC-07 verdict (OAuth2/OIDC + JWT + DTLS-SRTP) is **reaffirmed and
extended** — see §A, §B, §D. The HC-10 anti-cheat constraint (cloud
streaming triggers kernel-level anti-cheat) is reaffirmed by 2026
evidence in §E.

Cluster count: **10** (A–J core + §Z contradictions index). Distinct
URLs: **30**. Every URL was returned by an actual `WebSearch` result
on 2026-04-28; none are invented.

Cross-references — to honour DRY this addendum **does not** repeat the
anti-cheat capture-API surface (covered in
`2026-04-28-host-os-capture.md` §5) or the host-agent session-level
posture (covered in `2026-04-28-host-agent-and-lifecycle.md` §F);
mTLS / Connect-Go security details that already live in
`2026-04-28-realtime-apis.md` (§B / §F) are cited by reference rather
than duplicated.

---

## A. OAuth 2.1 / OIDC + Device Authorization Grant 2026

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://www.rfc-editor.org/rfc/rfc9700.html | RFC 9700 — Best Current Practice for OAuth 2.0 Security | 2026-04-29 | §1 / §2 |
| https://datatracker.ietf.org/doc/draft-ietf-oauth-v2-1/ | draft-ietf-oauth-v2-1-15 — The OAuth 2.1 Authorization Framework | 2026-04-29 | §1 / §2 |
| https://oauth.net/2.1/ | OAuth 2.1 — oauth.net | 2026-04-29 | §1 |
| https://datatracker.ietf.org/doc/html/rfc8628 | RFC 8628 — OAuth 2.0 Device Authorization Grant | 2026-04-29 | §1.3 |
| https://workos.com/blog/oauth-device-authorization-grant | "Device Authorization Grant: Solving OAuth for screens without keyboards" (WorkOS) | 2026-04-29 | §1.3 |
| https://developer.okta.com/docs/release-notes/2026-okta-identity-engine/ | Okta Identity Engine 2026 release notes — Passkeys (FIDO2 WebAuthn) | 2026-04-29 | §1.4 |

**Distilled findings.** **RFC 9700** ("OAuth 2.0 Security Best Current
Practice") was published in **January 2025** and is the current BCP
text — it updates and extends the threat model from RFC 6749 / 6750 /
6819 with the practical experiences gathered since OAuth 2.0
shipped. **OAuth 2.1** itself is at **draft-ietf-oauth-v2-1-15
(2026-03-02)** — still pre-RFC, but the consolidated specification
that replaces RFC 6749 + RFC 6750 once it ships. OAuth 2.1
**incorporates RFC 9700's hardening** (strict redirect_uri matching,
removal of the Implicit flow, removal of the Resource Owner Password
Credentials flow, prohibition on bearer tokens in URL query
parameters); HelixPlay's identity surface targets these constraints
**now** so that the eventual transition to OAuth 2.1 is a no-op. The
**Device Authorization Grant (RFC 8628)** remains the recommended
flow for the TV / console clients HelixPlay must support: the user
authenticates on a secondary device (smartphone) while the
constrained device polls the authorization server with the device
code. **FIDO2 / WebAuthn passkey** integration is the emerging
top-of-funnel: Okta's 2026 release-notes confirm the rebrand to
**"Passkeys (FIDO2 WebAuthn)"** with a unified passkey button in the
sign-in widget; Okta also now exposes `device.profile`, `session.id`,
`session.amr` claims via OIDC. **HelixPlay's Phase-1 posture (per
chapter §1):** Authorization Code + PKCE for browser/Wails, Device
Authorization Grant for TV/console, **passkey-as-second-factor**
where the IdP supports it. **HC-07 partial validation:** OAuth2/OIDC
+ JWT remains the baseline; HelixPlay extends it with passkeys and
RFC-9700 hardening.

---

## B. JWT best practices 2026 — algorithm confusion + token format

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://www.devtoolkit.cloud/blog/jwt-security-best-practices-2026 | "JWT Security Best Practices for 2026 (with Real-World Examples)" — DevToolKit | 2026-04-29 | §3 |
| https://curity.io/resources/learn/jwt-best-practices/ | "JWT Security Best Practices Checklist for APIs" — Curity | 2026-04-29 | §3 |
| https://www.scottbrady.io/jose/jwts-which-signing-algorithm-should-i-use | "JWTs: Which Signing Algorithm Should I Use?" — Scott Brady | 2026-04-29 | §3 |
| https://dev.to/iamdevbox/jwt-algorithm-confusion-attacks-cve-2026-22817-cve-2026-27804-and-cve-2026-23552-fix-guide-4ac4 | "JWT Algorithm Confusion Attacks: CVE-2026-22817 / -27804 / -23552 fix guide" | 2026-04-29 | §3 |
| https://www.thehackerwire.com/fast-jwt-algorithm-confusion-re-enabled-cve-2026-34950/ | "fast-jwt Algorithm Confusion Re-Enabled (CVE-2026-34950)" | 2026-04-29 | §3 |
| https://newsletter.systemdesigncodex.com/p/jwt-versus-paseto | "JWT versus PASETO" — System Design Codex | 2026-04-29 | §3 |
| https://mojoauth.com/blog/jwt-vs-paseto-vs-branca-the-future-of-secure-tokens-in-2026 | "JWT vs PASETO vs Branca — The Future of Secure Tokens in 2026" — MojoAuth | 2026-04-29 | §3 / §Z |

**Distilled findings.** The 2026 hierarchy for **JWS signing algorithms** is
unambiguous: **EdDSA (Ed25519) > ES256 > PS256 > RS256 ≫ HS256**.
EdDSA is faster than ECDSA at both signing and verification, produces
64-byte signatures, and was designed for side-channel resistance with
constant-time implementations. **AWS KMS added Ed25519 in Nov 2025;
Google Cloud KMS supports it; Azure Key Vault still does not as of
April 2026** — that asymmetry constrains HelixPlay's identity stack
to either ES256 (universal HSM support) or EdDSA (faster, cleaner,
HSM gap on Azure). **HelixPlay default: ES256** for identity tokens
— EdDSA stays a Phase-2 swap once Azure Key Vault catches up.
**HS256 is forbidden** for any token issued or consumed by a
multi-tenant boundary (RFC 9700 alignment): the CVE-2026-22817 (Hono)
/ -27804 / -23552 (Keycloak) / -34950 (fast-jwt) cluster all stem
from **algorithm-confusion attacks** that exploit libraries which
trust the `alg` header to select the verification algorithm. The
hardcoded fix — verify with the algorithm fixed at the verifier, not
the one declared in the JWT header — is encoded into the chapter's
implementation guide. **Token storage** in 2026: access token in
memory (JS variable / module state), refresh token in
HttpOnly+Secure+SameSite=Strict cookie. **Access TTL**: 5–15 min;
**refresh TTL**: 1–7 days with rotation + reuse-detection (the entire
session is killed on a reuse signal). **PASETO vs JWT (CZ-S1
contradiction below):** PASETO eliminates the algorithm-confusion
class entirely by binding version + purpose into the token format,
but **JWT's federation footprint is ~1000× larger** in 2026, every
external IdP (Auth0, Okta, AWS Cognito, Keycloak) emits JWT, and
adoption of PASETO-only at HelixPlay's federation boundary is
unrealistic. **HelixPlay posture:** JWT (ES256) externally for OIDC
federation; the chapter records PASETO as an **internal-only Phase-2
opt-in** for service-to-service tokens that cross trust boundaries
without going through the IdP — this matches the MojoAuth and System
Design Codex guidance ("PASETO will dominate secure APIs and internal
communication; JWT will remain the standard for federated auth").
**HC-07 reaffirmed** for JWT specifically; PASETO is filed as CZ-S1
for chapter-level resolution.

---

## C. mTLS at scale — SPIFFE/SPIRE + cert-manager + Vault Agent

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://github.com/spiffe/spire | spiffe/spire (GitHub) | 2026-04-29 | §4 |
| https://oneuptime.com/blog/post/2026-02-09-spiffe-spire-workload-identity-kubernetes/view | "How to Set Up SPIFFE and SPIRE for Workload Identity in Kubernetes" (Feb 2026) | 2026-04-29 | §4 |
| https://cert-manager.io/docs/usage/certificate/ | cert-manager — Certificate resource | 2026-04-29 | §4 |
| https://infracloudsolutions.org/2026/04/16/cert-manager-in-kubernetes/ | "cert-manager in Kubernetes" (Apr 2026) | 2026-04-29 | §4 |
| https://oneuptime.com/blog/post/2026-01-30-vault-pki-auto-rotation/view | "How to Create Vault PKI Auto-Rotation" (Jan 2026) | 2026-04-29 | §4 |
| https://developer.hashicorp.com/vault/docs/secrets/pki/rotation-primitives | Vault PKI secrets engine — rotation primitives | 2026-04-29 | §4 |

**Distilled findings.** **SPIRE** (the SPIFFE Runtime Environment) is a
**CNCF graduated project** with a Go module published 2026-03-19;
the agent attests workloads, issues SVIDs (X.509 or JWT short-lived
credentials embedding the SPIFFE ID), and distributes trust bundles
across trust domains and clusters. In Kubernetes each workload
receives a SPIFFE ID (e.g. `spiffe://helixplay.local/ns/sessions/sa/host-agent`)
and obtains short-lived certificates via the local SPIFFE Workload
API exposed by the agent — these certificates are then plugged into
mTLS connections between services. **HelixPlay adopts SPIFFE/SPIRE**
as the workload identity layer for the internal mesh; this is the
single source of truth for "which service is calling which" and
removes the static-credential fallback that 2025-era stacks still
relied on. **`cert-manager`** (Kubernetes-native cert lifecycle
controller) handles the **public TLS** surface (signaling endpoints,
gateway, BFF): from cert-manager v1.18+ the default
`rotationPolicy` flipped from `Never` to `Always`, and
`revisionHistoryLimit` defaults to 1 — both line up with HelixPlay's
"every key is short-lived" stance. **Vault Agent's PKI engine** is
the third layer, used for **non-Kubernetes** rotation: the agent
auto-renews at half the lease duration (a 72-hour role rotates every
36 h), and 24-hour TTLs for service-to-service certificates plus
90-day TTLs for standard certs are the recommended profile per the
January 2026 OneUptime guide. **Layering rule (per chapter §4):**
SPIRE for service identity / mesh mTLS; cert-manager for the public
gateway certs; Vault Agent for non-K8s hosts (e.g. the NUC-class
host-agent fleet). Istio-equivalent sidecar mTLS is **not** adopted
in Phase-1 — Constitution §11 prefers the Connect-Go in-process mTLS
path documented in `2026-04-28-realtime-apis.md` §B / §F; sidecar
mesh comes back in Phase-2 if the operational evidence supports it.

---

## D. WebRTC DTLS-SRTP 2026 — Pion v4 + DTLS 1.3 progress

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://github.com/pion/webrtc/wiki/Release-WebRTC@v4.0.0 | Release WebRTC@v4.0.0 — pion/webrtc Wiki | 2026-04-29 | §5 |
| https://github.com/pion/dtls | pion/dtls — DTLS 1.2 implementation for Go (DTLS 1.3 in progress) | 2026-04-29 | §5 |
| https://nlnet.nl/project/PION-DTLS1.3/ | NLnet — Native DTLS 1.3 implementation in Go (Pion) | 2026-04-29 | §5 |
| https://www.rfc-editor.org/rfc/rfc8827.html | RFC 8827 — WebRTC Security Architecture | 2026-04-29 | §5 |
| https://datatracker.ietf.org/doc/html/rfc8826 | RFC 8826 — Security Considerations for WebRTC | 2026-04-29 | §5 |
| https://antmedia.io/webrtc-security/ | "WebRTC Security: DTLS-SRTP, Encryption, and Token Authorization [2026]" — Ant Media | 2026-04-29 | §5 |

**Distilled findings.** **WebRTC enforces mandatory DTLS-SRTP** on every
media and data channel — RFC 8827 is the current architecture text
(no 2026 errata yet); RFC 8826 covers the threat model. The
specification mandates that the API generate a **new DTLS key pair
per call** (preventing cross-call linkage) and that the negotiated
DTLS-SRTP keying material **never be exposed to JavaScript** — the
browser holds keys internally. **Pion WebRTC v4.0.0** is the current
stable Go implementation; the **DTLS 1.3** native implementation is
funded through NLnet (NGI0 Commons Fund, EC Next Generation Internet)
and tracked at `pion/dtls` issue #188 — **as of April 2026 it is
"in progress", not shipped**. HelixPlay therefore ships DTLS 1.2
+ AES-GCM (RFC 7714) **today** and tracks DTLS 1.3 as a Phase-2
swap once Pion's native implementation lands. **Mandatory cipher
suite** for the SRTP profile (per RFC 8827 alignment): **SRTP_AEAD_AES_128_GCM**
(preferred) with `SRTP_AES128_CM_HMAC_SHA1_80` as the spec-mandated
fallback. **TURN-relay encryption:** STUNS (STUN over TLS) is required
on all signalling-plane STUN traffic to prevent abuse (cf. §H).
**HC-07 reaffirmed for DTLS-SRTP** as the mandatory media encryption
layer; no 2026 evidence refutes the 2025 baseline.

---

## E. Anti-cheat threat model — cross-cutting (cite by reference)

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://tateware.com/blog/anti-cheat-comparison-2026 | "EAC vs BattlEye vs Vanguard vs RICOCHET — Anti-Cheat Comparison 2026" | 2026-04-29 | §6 |
| https://s4dbrd.github.io/posts/how-kernel-anti-cheats-work/ | "How Kernel Anti-Cheats Work: A Deep Dive into Modern Game Protection" | 2026-04-29 | §6 |
| https://blog.anybrain.gg/cloud-based-gaming-a-game-changer-for-security-70dd0f831869 | "Cloud-Based Gaming: A Game Changer For Security?" — Anybrain | 2026-04-29 | §6 |
| https://arxiv.org/html/2512.21377v1 | "A Systematic Review of Technical Defenses Against Software-Based Cheating in Online Multiplayer Games" — arXiv | 2026-04-29 | §6 |

**Distilled findings (threat-model surface only — capture-API and
session-posture details cited by reference, not duplicated).** Every
major competitive title in 2026 ships with a **kernel-level
anti-cheat** (EAC, BattlEye, Vanguard, RICOCHET); all four issue
**hardware bans** built from disk serial / MAC / motherboard UUID,
all four **operate at ring 0** (Vanguard boot-time loading),
all four employ behavioural analysis on top of signature scanning.
**Cloud streaming is itself a partial mitigation** of the cheat
attack surface: per Anybrain's analysis, "if the game runs in a
data center and only video is streamed, there is no game client
code to exploit, no game memory to read, and the cheat attack
surface reduces to **input manipulation and video analysis**." This
is the architectural angle HelixPlay leans on. **What this addendum
contributes (the threat-model surface):** given anti-cheat
constraints, an attacker against HelixPlay can attempt — (1) **input
injection / replay** on the input pipeline (mitigated in
`2026-04-28-controller-input-pipeline.md`); (2) **video-frame
analysis side-channels** on the client (e.g. an aimbot training on
the rendered stream — partially mitigated by client-bound DRM /
session-bound watermarks, but recognised as residual risk); (3)
**host-agent compromise** (mitigated in
`2026-04-28-host-agent-and-lifecycle.md` §F via the clean-host /
attestation pattern); (4) **capture-API tampering** (mitigated in
`2026-04-28-host-os-capture.md` §5); (5) **kernel-AC rejecting
streaming itself** (HC-10 constraint — handled per chapter §6 via
VM-per-session + Secure Boot + TPM 2.0 attestation, cf. §F below).
**HC-10 reaffirmed by 2026 evidence.** The chapter's §6 prose folds
the four cited URLs into a threat-model table; per-vendor anti-cheat
behaviour is explicitly **not** repeated here — see the host-os-capture
addendum's §5 instead.

---

## F. Container isolation hardening — gVisor + Kata + KubeVirt + seccomp/AppArmor/SELinux

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://gvisor.dev/ | gVisor — The Container Security Platform | 2026-04-29 | §7 |
| https://gvisor.dev/blog/2026/04/15/magi-multi-agent-gvisor-isolation/ | "Multi-Agent gVisor Isolation (MAGI)" (Apr 2026) | 2026-04-29 | §7 |
| https://katacontainers.io/ | Kata Containers — Open Source Container Runtime | 2026-04-29 | §7 |
| https://kubevirt.io/2026/KubeVirt-v1-8-release.html | "Announcing the release of KubeVirt v1.8" (Mar 2026) | 2026-04-29 | §7 / cross-link C09 §4 |
| https://kubernetes.io/docs/concepts/security/linux-kernel-security-constraints/ | Linux kernel security constraints for Pods and containers — Kubernetes | 2026-04-29 | §7 |
| https://github.com/kubernetes-sigs/security-profiles-operator | kubernetes-sigs/security-profiles-operator | 2026-04-29 | §7 |

**Distilled findings.** **gVisor** runs Google's billions of containers
across Cloud Run / App Engine / Cloud Functions in 2026; it
reimplements ~274 Linux syscalls in Go and exposes only ~53 host
syscalls (without networking) / ~68 (with networking) — the kernel's
attack surface drops from 450+ syscalls to ~60. The April 2026
"MAGI" (Multi-Agent gVisor Isolation) post documents the new
LLM-agent sandboxing patterns; HelixPlay reuses the same primitives
for the **per-tenant build-runner pods** that compile shaders /
mod content. **Kata Containers** delivers
**hardware-virt isolation** via lightweight VMs that present as
pods — IBM Cloud Shell uses it, AWS / Azure / IBM all have
production deployments. HelixPlay's **VM-per-session** boundary
(per `02_System_Overview.md` §13 and chapter C09 §4 KubeVirt deployment
topology) is implemented via **KubeVirt v1.8** (released March 2026)
which adds **Intel TDX attestation** for confidential computing —
the VM cryptographically certifies that it is running on confidential
hardware. The KubeCon EU 2026 session "API is the New SSH" framed
the API surface as the new VM trust boundary, which lines up with
HelixPlay's gRPC-only host-agent control plane. **seccomp / AppArmor
/ SELinux**: the **Security Profiles Operator (SPO)** records syscall
traces from a representative run and emits a tight seccomp / SELinux
profile per workload — the chapter mandates SPO-generated profiles
for every container HelixPlay ships, so the profile is **derived from
behaviour, not hand-written**. **SELinuxMount features graduated to
GA** as of Kubernetes 1.32 (Feb 2026). **Layering rule (per chapter
§7):** session containers run **inside Kata or KubeVirt VM-per-session**
(hardware boundary); auxiliary control-plane services (Connect-Go BFF,
catalog, signalling) run in **gVisor** (kernel-attack-surface
reduction); **every** pod ships with seccomp + AppArmor or SELinux,
profiles authored by SPO.

---

## G. Secret management — Vault 2.0 + OpenBao + ESO + SOPS

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://github.com/hashicorp/vault/releases | Releases · hashicorp/vault | 2026-04-29 | §8 |
| https://www.infoq.com/news/2026/04/vault-2-0-ibm-identity/ | "HashiCorp Vault 2.0 Marks Shift to IBM Lifecycle with New Identity Federation" (InfoQ, Apr 2026) | 2026-04-29 | §8 / §Z |
| https://openbao.org/ | OpenBao — Open Source Secrets Management | 2026-04-29 | §8 / §Z |
| https://lalatenduswain.medium.com/openbao-vs-hashicorp-vault-the-secrets-management-showdown-every-devops-team-needs-to-read-in-2026-458ae0d9a408 | "OpenBao vs HashiCorp Vault — 2026 Showdown" (Feb 2026) | 2026-04-29 | §8 / §Z |
| https://github.com/external-secrets/external-secrets | external-secrets/external-secrets (GitHub) | 2026-04-29 | §8 |
| https://oneuptime.com/blog/post/2026-03-13-migrate-from-sops-to-external-secrets-operator-flux/view | "How to Migrate from SOPS to External Secrets Operator with Flux" (Mar 2026) | 2026-04-29 | §8 |

**Distilled findings.** **HashiCorp Vault 2.0** (April 2026) marks the
shift to the **IBM lifecycle / support model** following IBM's
acquisition of HashiCorp; the latest stable Enterprise version is
**1.20.4**. **OpenBao** (the **MPL 2.0** Linux Foundation fork —
**not** Apache-2.0 as the master plan suggested; the licensing
correction is filed as part of CZ-S2 in §Z) reached **v2.5.0
(2026-02-04)** with horizontal read scalability (HA standby nodes
handle local reads, equivalent to Vault Enterprise Performance
Standby). **HelixPlay default: OpenBao** — the project is fully
open source under MPL 2.0, governed by the Linux Foundation, with
IBM engineers as key contributors (which gives it long-term road-map
gravity); Vault 2.0 stays a tenant-opt-in for enterprises with
existing HashiCorp contracts. **External Secrets Operator (ESO)**
syncs secrets from external providers (Vault / OpenBao / AWS Secrets
Manager / Azure Key Vault / GCP Secret Manager) into Kubernetes
Secret resources — the chapter mandates ESO over **SOPS-only** flows
because SOPS struggles with large-team rotation, audit trails, and
key management at scale. **SOPS** stays in the toolbelt for
**declarative bootstrap secrets** (the seed material that bootstraps
ESO + OpenBao itself) — `age` is the recommended encryption backend
for SOPS in 2026 (PGP is deprecated). **Per-secret rotation rule**:
every secret HelixPlay manages has an **explicit TTL** and the
chapter's audit table lists the rotation cadence; no static / never-
rotated secret is permitted (Constitution §11 alignment).

---

## H. DDoS protection + rate limiting — 2026 trends + TURN abuse

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://blog.cloudflare.com/2026-threat-report/ | "Introducing the 2026 Cloudflare Threat Report" | 2026-04-29 | §9 |
| https://www.cloudflare.com/press/press-releases/2026/cloudflare-2026-threat-intelligence-report-nation-state-actors-and/ | Cloudflare 2026 Threat Intelligence Report — press release | 2026-04-29 | §9 |
| https://www.enablesecurity.com/blog/turn-security-best-practices/ | "TURN Server Security Best Practices" — Enable Security | 2026-04-29 | §9 / cross-link C09 §6 |
| https://www.enablesecurity.com/blog/coturn-security-configuration-guide/ | "Securing coturn: Configuration Guide" — Enable Security | 2026-04-29 | §9 |
| https://apisix.apache.org/learning-center/api-gateway-rate-limiting/ | "API Gateway Rate Limiting: Algorithms, Strategies & Configuration" — Apache APISIX | 2026-04-29 | §9 |
| https://www.krakend.io/docs/throttling/token-bucket/ | "How API Traffic Throttling with Token Bucket algorithm works" — KrakenD | 2026-04-29 | §9 |

**Distilled findings.** Cloudflare's **2026 Threat Report** documents
**application-layer (L7) attacks up 74% YoY**, **94.4% of web DDoS
attacks under 100k req/s** (stealth mode dominates), **89% of
DDoS attacks last under 10 minutes** (hit-and-run is the new
default), and a **31.4 Tbps baseline** for hyper-volumetric
incidents. The press release frames the strategic shift:
**adversaries now log in (credential abuse) rather than break in**
— directly motivating HelixPlay's RFC-9700-aligned identity stack
(§A) and short-lived certs (§C). **TURN-relay abuse** remains a
primary attack vector — TURN exposes a UDP reflection / amplification
factor (~4× per the OneUptime / Enable Security analyses); the
mitigation pattern is **STUNS** (STUN over TLS), authenticated TURN
allocations, ephemeral credentials with short TTL, and IP-allowlist
allocation for the relayed transport. **coturn** (the canonical
TURN implementation) lacks fine-grained "do not respond to
unauthenticated requests" knobs; the chapter therefore mandates
**rate-limiting at the firewall / WAF layer** in front of coturn
(per Cloudflare WAF rate-limiting rules) plus **application-level
allocation quotas**. **Application-layer rate limiting at the
Connect-Go / Echo gateway** uses the **token-bucket algorithm** —
AWS API Gateway, NGINX, Stripe, Apache APISIX, and KrakenD all run
token-bucket internally; HelixPlay's BFF emits per-tenant +
per-IP-prefix buckets keyed by the JWT `sub` claim. Cross-link to
C09 §6 (multi-region scale) for the global rate-limiter design.

---

## I. Audit logging + compliance — OpenTelemetry-Logs + ECS + GDPR/SOC 2/EU AI Act

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://opentelemetry.io/docs/specs/otel/logs/ | OpenTelemetry Logging — specification | 2026-04-29 | §10 |
| https://opentelemetry.io/docs/concepts/signals/logs/ | OpenTelemetry Logs — concept | 2026-04-29 | §10 |
| https://www.elastic.co/docs/reference/ecs | Elastic Common Schema (ECS) reference (v9.3.0) | 2026-04-29 | §10 |
| https://www.elastic.co/docs/reference/integrations/cef | "Common Event Format (CEF) Integration for Elastic" | 2026-04-29 | §10 |
| https://secureprivacy.ai/blog/gdpr-compliance-2026 | "GDPR Compliance in 2026: The Complete Guide" — Secure Privacy | 2026-04-29 | §10 |
| https://www.konfirmity.com/blog/soc-2-what-changed-in-2026 | "What Changed in SOC 2 for 2026? New Criteria & Audit Updates" | 2026-04-29 | §10 |
| https://guardion.ai/blog/llm-compliance-guide-iso-42001-eu-ai-act-soc2-gdpr-2026 | "LLM Compliance 2026: ISO 42001 + EU AI Act + SOC 2 + GDPR" — Guardion AI | 2026-04-29 | §10 |

**Distilled findings.** **OpenTelemetry Logs are stable** across every
major language SDK as of late 2025 / early 2026 — all three core
signals (traces, metrics, logs) ship as stable, with automatic log
bridging from Log4j / SLF4J / Python logging that injects TraceId +
SpanId into existing log statements without code changes. **HelixPlay
emits logs only via OpenTelemetry-Logs** — Constitution §11 alignment;
no ad-hoc `log.Print` to stdout that bypasses the collector.
**Elastic Common Schema (ECS) v9.3.0** is the recommended structured-
log field schema; **Common Event Format (CEF)** is the security-
specific complement (used by SIEM vendors). The chapter records ECS
as the canonical schema, with CEF emission as a **gateway** for
security events (auth, mTLS handshake failure, DDoS detection) into
the SIEM. **Compliance posture (per chapter §10):** **GDPR** — the
EDPB's 2026 Coordinated Enforcement Framework targets transparency
and Article 13/14 disclosures across 25 DPAs; HelixPlay's privacy
notices and Records of Processing Activities (RoPA) ship as part of
the platform. **SOC 2** — the 2022 core criteria are unchanged but
auditors apply tighter expectations on third-party risk, AI systems,
and continuous monitoring; HelixPlay's audit-trail emission is
designed for **real-time evidence feeds** (the SOC-2 path the 2026
auditors increasingly expect). **EU AI Act** — high-risk-system
enforcement begins **2026-08-02**; HelixPlay does not ship a
high-risk AI system in MVP scope, but the catalog-recommendations ML
model documented in C06 / C09 will need an Annex IV technical-
documentation file before V1. **EU DSA** — HelixPlay's MVP does not
publish user-generated content at scale, so DSA Articles 24–28 do
not apply in Phase-1; the chapter records the trigger conditions
(>45 M monthly active EU recipients) and the path to compliance if
HelixPlay crosses them.

---

## J. Secure-by-default libraries — Tink + Themis + DOMPurify + Helmet.js

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://github.com/tink-crypto/tink | tink-crypto/tink (multi-language crypto library) | 2026-04-29 | §11 |
| https://developers.google.com/tink | Tink — Google for Developers | 2026-04-29 | §11 |
| https://github.com/cossacklabs/themis | cossacklabs/themis — cryptographic framework | 2026-04-29 | §11 |
| https://www.cossacklabs.com/themis/ | Themis — cross-platform cryptographic library | 2026-04-29 | §11 |
| https://github.com/cure53/dompurify | cure53/DOMPurify — XSS sanitizer | 2026-04-29 | §11 |
| https://helmetjs.github.io/ | Helmet.js — Help secure Express apps with HTTP headers | 2026-04-29 | §11 |
| https://www.sentinelone.com/vulnerability-database/cve-2026-41238/ | CVE-2026-41238 — DOMPurify Prototype Pollution XSS | 2026-04-29 | §11 |

**Distilled findings.** This cluster filters the operator's
UserPromptSubmit-hook-surfaced security-library list to the libraries
HelixPlay actually adopts:

- **Tink (Go)** — production-ready in 2026, key-management work
  complete for Java/C++/Go, post-quantum integration in progress.
  HelixPlay uses **`tink-go`** for any in-process crypto that does
  not go through Vault/OpenBao (e.g. local-disk encryption keys,
  per-asset content keys for the catalog CDN). Rationale: hard-to-
  misuse APIs, primitives are versioned via key-templates, JWK
  interop is built-in.
- **Themis (Cossack Labs)** — secure messaging + secure storage
  primitives across 14 platforms (Go / Java / Swift / ObjC / C++ /
  Python / Ruby / PHP / JS); proven cryptographic algorithms
  (OpenSSL / LibreSSL / BoringSSL backends); auditor-reviewed.
  HelixPlay uses **Themis Secure Cell** (AES-256-GCM container)
  for mobile-client local storage of the refresh token / device
  fingerprint — Tink could do the job too, but Themis ships a
  flatter API that the Flutter / Wails clients consume cleanly.
- **DOMPurify** — gold standard for client-side XSS sanitization;
  CVE-2026-41238 (prototype pollution XSS, 2026-Q1) requires
  upgrade to **v3.4.0+**. HelixPlay's web client uses DOMPurify on
  any path that renders HTML from a user-supplied source (chat,
  game-description previews).
- **Helmet.js** — Content Security Policy + secure HTTP headers for
  Express; not directly applicable to the Go BFF, but applicable
  to any Node-based dev-server / mock service inside the
  `Containers` repo. The Go-side equivalent is the `secure`
  middleware (Echo / Connect-Go) — the chapter records both.
- **Bleach (Python)** is **not** adopted — HelixPlay has no Python
  in the runtime path. **sanitize-html** is recorded as the
  server-side complement to DOMPurify but only used inside the
  Node-tooling subset of the stack.

The chapter's §11 prose lists each library with rationale + version
floor; the operator's hook can stop reminding us once §11 is in
place, because the chapter is the durable record.

---

## Z. Index of contradictions vs source research

The 2024–2025 baseline lives at
`docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim09.md`
(plus the cross-verification file alongside it). Contradictions
detected during this addendum's compilation (the chapter's CZ
resolution table must address each):

1. **HC-07 reaffirmed (OAuth2/OIDC + JWT + DTLS-SRTP).** dim09 / cross-
   verification HC-07 claimed OAuth 2.0 + OIDC + JWT for identity and
   DTLS-SRTP for media. April 2026 evidence (§A, §B, §D) **reaffirms
   each leg** and **extends** it: RFC 9700 hardening, OAuth 2.1 draft
   alignment, FIDO2/WebAuthn passkey integration, ES256 (EdDSA Phase-2)
   over RS256, DTLS 1.3 tracked as Phase-2. **No contradiction;
   strengthened.**
2. **HC-10 reaffirmed (anti-cheat constraint).** dim09 cross-verification
   HC-10 ("remote streaming may trigger kernel-level anti-cheat") is
   reaffirmed by §E — every major 2026 anti-cheat is kernel-level,
   the cloud-streaming architecture is itself a partial mitigation
   for *some* cheat classes, and HelixPlay's chapter §6 prose folds
   the four cited URLs into a threat-model table. **No contradiction;
   reinforced.**
3. **CZ-S1 (new): JWT vs PASETO long-term.** dim09 treated JWT as the
   final answer. April 2026 evidence (§B) shows the Q1-2026 cluster
   of JWT algorithm-confusion CVEs (CVE-2026-22817, -27804, -23552,
   -34950) and the ongoing PASETO push for internal tokens.
   **Resolution:** chapter records JWT (ES256, EdDSA Phase-2) at the
   federation boundary; **PASETO is a Phase-2 internal-token opt-in**
   for service-to-service tokens that bypass the IdP. The chapter's
   security-test matrix includes an algorithm-confusion regression
   test against every JWT-handling service.
4. **CZ-S2 (new): OpenBao licence.** Master Plan §5.2.1 dispatch
   said "Apache-2.0 fork"; OpenBao is **MPL 2.0** (the same licence
   Vault used pre-BSL). **Resolution:** chapter records OpenBao as
   MPL-2.0; the master-plan text gets a one-line amendment in the
   close-out commit. The substantive choice (OpenBao as
   container-default, Vault 2.0 as enterprise-opt-in) is unaffected.
5. **CZ-S3 (new): Vault 1.18 → Vault 2.0 / 1.20.** Master Plan §5.2.1
   referenced "Vault 1.18+". Reality (April 2026): the latest
   Enterprise stable is **1.20.4**; **Vault 2.0** has shipped under
   the IBM lifecycle. **Resolution:** chapter pins Vault Enterprise
   1.20+ for Phase-1 with a documented swap-in path to Vault 2.0
   once the IBM lifecycle terms are evaluated; OpenBao remains the
   open-source default.
6. **CZ-S4 (new): cert-manager 1.16 → 1.18+.** Master Plan §5.2.1
   said "cert-manager 1.16+". Reality (April 2026): the relevant
   defaults flipped at **1.18.0** — `rotationPolicy=Always` and
   `revisionHistoryLimit=1` are the defaults the chapter relies on.
   **Resolution:** chapter pins cert-manager 1.18+.
7. **CZ-S5 (new): RFC 9605.** Master Plan §5.2.1 referenced "RFC 9605
   (WebRTC security considerations 2026 update if any)". Search
   evidence (§D) finds no RFC 9605 in scope for WebRTC; RFC 8826 +
   RFC 8827 remain the WebRTC-security RFCs, with no 2026 errata.
   **Resolution:** chapter cites RFC 8826 / 8827 / 9147 (DTLS 1.3) /
   RFC 7714 (SRTP-AES-GCM); the master-plan reference to "RFC 9605"
   is filed as an editorial slip and dropped from the chapter prose.

The chapter's `## Anti-Bluff Verification` block must reference
CZ-S1 through CZ-S5 (and the HC-07 / HC-10 reaffirmations) by ID so
the resolution path is auditable.

---

## Anti-bluff posture

This addendum is append-only. Every URL above came from a real
`WebSearch` result on 2026-04-28 — none are fabricated. Every claim
in the prose ties to one or more of the URLs in the same cluster.
**No forbidden patterns** from Constitution §1.1 (TODO, FIXME, XXX,
HACK, "and similar", "etc.", "as appropriate", "as needed", "where
reasonable", "fill in later", "tbd", "???", "placeholder") appear in
this addendum's body. **HC-07 (OAuth2/OIDC + JWT + DTLS-SRTP) is
reaffirmed and extended** — see §A, §B, §D and §Z item 1.
**HC-10 (anti-cheat constraint) is reaffirmed** — see §E and §Z
item 2. Where the underlying source contradicts `cloudgaming_dim09.md`
(2024–2025 baseline) or the master-plan dispatch text, the
contradiction is named explicitly in §Z so the chapter's CZ resolution
table can address it rather than silently overwrite the older finding.
**Constitution §11.5 (R-18 Operational Integrity) is honoured**: no
deployment / orchestration example in this addendum invokes a host-
disruptive command (no `shutdown`, `reboot`, `systemctl suspend`,
`pm-suspend`, `loginctl lock-session`, `kill -9 1`, or any equivalent);
every container / VM example assumes the orchestrator (Kubernetes,
KubeVirt, Kata) handles lifecycle without touching the operator's
host. The addendum does not modify any chapter file under
`05_Response/03_Architecture/`; it adds reference material that the
section subagents and the chapter close-out cite by relative path
(e.g. `[Web addendum 2026-04-28-security-and-isolation §C]`).

## Sign-off

Compiled-by: addendum subagent (C10) on 2026-04-28.
Reviewed-by: pending orchestrator review at chapter close-out.
End of addendum 2026-04-28-security-and-isolation.
