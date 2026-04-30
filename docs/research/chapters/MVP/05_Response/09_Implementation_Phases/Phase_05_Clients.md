# Phase_05 — Clients

> **Source dimensions:** [`Phase_04_Streaming_Pipeline.md`](Phase_04_Streaming_Pipeline.md), [`../06_Submodules/per-submodule/helix-tv-input.md`](../06_Submodules/per-submodule/helix-tv-input.md), [`../06_Submodules/per-submodule/helix-input.md`](../06_Submodules/per-submodule/helix-input.md), [`../06_Submodules/per-submodule/helix-display.md`](../06_Submodules/per-submodule/helix-display.md), [`../03_Architecture/04_Go_Client_Ecosystem.md`](../03_Architecture/04_Go_Client_Ecosystem.md) (C05), [`../03_Architecture/11_TV_UX.md`](../03_Architecture/11_TV_UX.md) (C12), [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row P05.
> **Source line count:** chapter floor 500 lines per Master Plan §7.2 row P05.
> **Phase targets:** R-01 + R-09 (allocation-free hot-path on the client side too) + R-13.
> **Cross-links:** [`Phase_06_Host_Agent.md`](Phase_06_Host_Agent.md), [`Phase_08_Audio_Surround.md`](Phase_08_Audio_Surround.md).
> **Status:** Draft v1 specification.
> **Last updated:** 2026-04-30.

---

## 1. Phase Scope

Phase_05 ships the **HelixPlay clients** across the three primary client surfaces:

1. **Wails desktop client** — Linux + macOS + Windows desktop variant (Go + JS frontend).
2. **Compose-for-TV client** — Android-TV variant (Kotlin Compose-for-TV with `androidx.tv.material3` 1.0 GA + 1.1.0-rc01 per [C12 §6 MC-05 closure](../03_Architecture/11_TV_UX.md)).
3. **Steam Deck game-mode client** — Linux Wayland with Gamescope variant.

Each client integrates **helix-tv-input** (TV input dispatch + WCAG 2.2 SC 2.5.8 64-dp focus targets), **helix-input** (controller-side echo + Reflex round-trip), and **helix-display** (frame pacing + VRR + ALLM negotiation).

After Phase_05, an end-user with a real client (laptop / TV / Steam Deck) can connect to a Phase_04 host-agent + see actual gameplay rendered on their display. This is the **first user-facing milestone**.

---

## 2. Prerequisites

- Phase_04 complete + signed off — host-side streaming pipeline operational.
- helix-tv-input, helix-input, helix-display, helix-tenant, helix-grpc-frame at v1.0.0.
- Client-side build infrastructure: Wails ≥ v2.10, Android Studio Hedgehog+, Steam Deck SteamOS-compatible toolchain.

---

## 3. Tasks Catalogue

| Task ID    | Task                                                          | Subtasks |
|------------|---------------------------------------------------------------|---------:|
| P05.T01   | Wails desktop client scaffold + Go-side wiring                | 6        |
| P05.T02   | Wails JS frontend (catalog browse + session control)          | 6        |
| P05.T03   | Wails OAuth + Vault token acquisition                         | 4        |
| P05.T04   | Wails Pion WebRTC consumer + helix-display integration        | 5        |
| P05.T05   | Compose-for-TV scaffold + JNI shim from helix-tv-input         | 5        |
| P05.T06   | Compose-for-TV catalog UI + reference user journey             | 6        |
| P05.T07   | Compose-for-TV WebRTC consumer + ExoPlayer integration         | 5        |
| P05.T08   | Steam Deck game-mode client (game-scope-friendly Wails variant)| 5        |
| P05.T09   | Cross-platform input handling — controllers + remotes + keyboards| 5    |
| P05.T10   | Per-client smoke test — first user journey end-to-end         | 4        |
| P05.T11   | Client signing for distribution (operator + per-platform)      | 4        |
| P05.T12   | Client crash-reporting integration (Sentry-equivalent)         | 3        |
| P05.T13   | Phase_05 acceptance review                                     | 2        |

13 tasks; ~60 subtasks.

---

## 4. Task Details

### 4.1 P05.T01 — Wails desktop scaffold

Per [C05 §6](../03_Architecture/04_Go_Client_Ecosystem.md):

- Wails v2.10 init + custom title-bar (Apple HIG + Microsoft Fluent + GNOME HIG conformant).
- Go-side service binding for catalog + session + tenant configuration.
- Auto-update mechanism via Wails-supported approach + cosign verify on update artefacts.
- Vendored Go modules (no Go-runtime deps at distribution).
- Distroless-equivalent for the Wails runtime (note: Wails embeds a webview; not strictly distroless but operator-acceptable).

### 4.2 P05.T02 — Wails JS frontend

- React 18 + TypeScript 5.6.
- Catalog browse with helix-tenant.Theme application.
- Session control (start / pause / resume / end).
- Translation system (every UI string in i18n bundles).
- WCAG 2.2 SC compliance for keyboard navigation.

### 4.3 P05.T03 — OAuth + Vault

- OAuth2 PKCE flow against the operator's IdP.
- Vault token acquisition + refresh.
- Per-tenant token scoping.
- Token revocation on logout.

### 4.4 P05.T04 — Pion WebRTC consumer

- Pion WebRTC v4 consumer side.
- ICE / SRTP integration via helix-transport's Connection interface.
- Frame decode via libavcodec (cgo bindings).
- helix-display.Pacer for frame presentation.

### 4.5 P05.T05 — Compose-for-TV scaffold

- Android Studio Hedgehog project.
- `androidx.tv.material3` 1.0 GA (per [C12 §6 MC-05 closure](../03_Architecture/11_TV_UX.md)).
- `gomobile bind` integration of helix-tv-input's JNI shim.
- Leanback fallback path (deprecated post-2026-08-31 64-bit Play Store mandate).

### 4.6 P05.T06 — Compose-for-TV catalog UI

- Horizontal-shelf paradigm per Insight #6 + scènes-à-faire grounding.
- 64 dp minimum focus targets per WCAG 2.2 SC 2.5.8.
- Trailer auto-play with 2 s focus dwell + 7 s auto-advance + reduced-motion override.
- Per-tenant theme application via helix-tenant.Theme.ApplyToClient.

### 4.7 P05.T07 — Compose-for-TV WebRTC consumer

- ExoPlayer integration with Pion-Android bridge.
- helix-display VRR negotiation (HDMI 2.1 VRR + ALLM).

### 4.8 P05.T08 — Steam Deck game-mode

- Wails-derived variant + Gamescope-friendly window management.
- SteamInput SDK integration vs helix-input choice (per OQ-tv-input-B).
- Steam Deck-specific button mapping.

### 4.9 P05.T09 — Cross-platform input

- Xbox 360, Xbox One, DualShock 4, DualSense, Stadia, Joy-Con, Steam Controller — full controller compatibility.
- Per-platform input-device permission UX.
- Reflex round-trip echo via helix-input.

### 4.10 P05.T10 — Per-client smoke

For each of the 3 client variants:

1. User logs in via OAuth.
2. Browses catalog (lists 5 titles).
3. Selects a title; session begins.
4. Receives 1080p60 stream from Phase_04 host-agent.
5. Issues controller input; verify host-side action.
6. Disconnects gracefully.

p999 controller round-trip ≤ 25 ms (Phase_05 relaxed).

### 4.11 P05.T11 — Client distribution signing

- Linux .AppImage signed with cosign.
- Windows .msi signed with operator's EV cert.
- macOS .dmg notarised + signed with Apple Developer ID.
- Android .apk + .aab signed with operator's Play upload key.
- Steam Deck Flatpak signed.

### 4.12 P05.T12 — Crash reporting

Sentry (or operator's equivalent) integrated into every client; per Constitution §10 observability.

### 4.13 P05.T13 — Acceptance review

Operator + UX review per Constitution §16. UX review is non-trivial — per [C12 §6](../03_Architecture/11_TV_UX.md), operator validates the Reference User Journey (per [System Overview §3](../02_System_Overview.md#3-reference-user-journey)) on every client variant.

---

## 5. Subtask Catalogue

60 subtasks across 13 tasks; per-task discrete `[P05.Tyy.Szz]` tickets per [O05](../08_Operations/05_Tracking_GitHub_GitLab.md). Sub-categories: Wails desktop client (T01..T05); Compose-for-TV client (T06..T09); Steam Deck client (T10..T12); cross-cutting acceptance (T13). Each subtask carries: function entry-point + verification probe + observability emission + Challenges scenario reference. Detailed enumeration tracks alongside per-task PRs in operator's project boards.

---

## 6. Exit Criteria

- [ ] Wails desktop client builds + runs on Linux + macOS + Windows.
- [ ] Compose-for-TV client builds + runs on Android-TV emulator + at least 1 real Android-TV device.
- [ ] Steam Deck client builds + runs on real Steam Deck.
- [ ] Per-client smoke test passes for all 3.
- [ ] Cross-platform input device coverage verified.
- [ ] OAuth + Vault token flow end-to-end.
- [ ] Per-client distribution signing complete.
- [ ] Crash-reporting integrated.
- [ ] WCAG 2.2 SC 2.5.8 64-dp focus targets verified on TV variant.
- [ ] Operator signoff per Constitution §16.

---

## 7. Risk Register

| ID       | Risk                                                                | Mitigation                                                            |
|----------|----------------------------------------------------------------------|------------------------------------------------------------------------|
| RP05-01  | Apple notarisation rejection                                          | Phase_00 P00.T03 operator Apple Developer ID set up; tested before Phase_05. |
| RP05-02  | androidx.tv.material3 1.0 GA upstream regression                      | Pin to a specific minor version; Renovate-driven bump policy.         |
| RP05-03  | Steam Deck SteamOS update breaks Wails Gamescope integration         | Operator monitors SteamOS release notes; pinned to a specific SteamOS version + tested before bumps. |
| RP05-04  | Pion-Android bridge instability on legacy Android-TV devices         | Fallback to ExoPlayer's native WebRTC (less feature-rich but stable). |
| RP05-05  | ICE NAT traversal fails in operator's specific tenant network        | Phase_03 P03.T05 TURN server is the documented fallback.              |

---

## 8. Cross-Family Dependencies

- C05 §6 (Go Client Ecosystem) + C12 §6 (TV UX) architecturally specify the patterns.
- helix-tv-input + helix-input + helix-display (Phase_02) provide the primitives.
- helix-pipeline + helix-transport (Phase_04) provide the host-side backend.
- helix-tenant (Phase_02) provides theming.

---

## 9. The Phase_05 Calendar

| Week | Activity                                                         |
|------|------------------------------------------------------------------|
| 1–2  | T01–T04 (Wails desktop)                                           |
| 3–4  | T05–T07 (Compose-for-TV)                                          |
| 5    | T08 (Steam Deck)                                                  |
| 5    | T09 (cross-platform input)                                         |
| 6    | T10 (per-client smoke)                                            |
| 6    | T11–T12 (signing + crash reporting)                                |
| 6    | T13 (acceptance)                                                   |

Operator-side capacity: 4–6 engineers covering Go + TS/React + Kotlin + Wayland + Apple ecosystems. ~ 6 weeks.

---

## 10. Acceptance Criteria

Constitution §16 signoff + §6 exit criteria.

---

## 10a. Per-Phase Detailed Task Acceptance Criteria

### 10a.1 Wails desktop client acceptance (P05.T01..T05)

- Wails v2 application builds on Linux (x86_64 + arm64), macOS (x86_64 + arm64), Windows (x86_64).
- Application bundle size ≤ 80 MB per-platform after distroless-base + tree-shake.
- First-frame render within 3 s p99 from launch on reference hardware (Intel UHD 770 / Apple M1 / AMD Radeon).
- Per-OS code-signing verified: Apple notarization, Microsoft Authenticode, Linux Sigstore-cosign.
- Auto-update mechanism cosign-verified per [Phase_06 P06.T07](Phase_06_Host_Agent.md#47-p06t07--auto-update).
- Crash-rate < 1 / 1,000 sessions verified via Sentry-mirrored telemetry.

### 10a.2 Compose-for-TV client acceptance (P05.T06..T09)

- Compose-for-TV `androidx.tv.material3` 1.0 GA + 1.1.0-rc01 build verified.
- Focus-target 64 dp minimum size (WCAG 2.2 SC 2.5.8 + per [C12 Z-3](../03_Architecture/11_TV_UX.md)).
- Focus-target action response p99 ≤ 100 ms.
- Trailer auto-play: 2 s focus dwell + 7 s auto-advance + reduced-motion override per C12 Z-2.
- Play Store + Amazon Appstore submission accepted (operator-side commercial agreement).
- Leanback deprecation path: Compose-for-TV is the only Android-TV path post 2026-08-31.

### 10a.3 Steam Deck client acceptance (P05.T10..T12)

- Steam Deck native build (arm64 SteamOS).
- Sustained 25 W operation with thermal throttling profile.
- Quick Access Menu integration verified.
- Per-game thermal profile auto-calibrated.
- Steam OAuth + ownership-API integration verified (T13).

### 10a.4 Cross-cutting acceptance (P05.T13)

- Per-launcher OAuth (Steam + GOG + Epic + Battle.net) flow completion ≥ 99% per [Phase_10 P10.T13](Phase_10_Monetization_and_Auth.md#413-p10t13--per-game-launcher-entitlement-check).
- Per-client codec capability negotiation (H.264 + HEVC + AV1) verified.
- Per-client mDNS discovery + first-session smoke green per [Phase_06 P06.T11](Phase_06_Host_Agent.md#411-p06t11--end-to-end-smoke).

---

## 11. Per-Phase Observability Catalogue

### 11.1 Prometheus metrics

| Metric | Type | Labels | SLO Target |
|--------|------|--------|------------|
| `helix_client_session_start_seconds` | histogram | client_type, region | p99 ≤ 5 s |
| `helix_client_first_frame_seconds` | histogram | client_type, region | p99 ≤ 3 s |
| `helix_client_input_to_photons_ms` | histogram | client_type, tenant | p999 per Phase_07 |
| `helix_client_codec_negotiation_total` | counter | client_type, codec | per-session |
| `helix_client_reconnect_total` | counter | client_type, reason | < 1 / hour |
| `helix_client_crash_total` | counter | client_type, os | < 1 / 1000 sessions |
| `helix_client_audio_loss_total` | counter | client_type | 0 (audit) |

### 11.2 Grafana dashboards

- **Per-client-type Health** — Wails / Compose-for-TV / Steam Deck per-version session metrics.
- **Per-region Latency by Client** — p999 input-to-photons stratified by client type.
- **Per-codec adoption** — H.264 / HEVC / AV1 distribution per client type.
- **Per-launcher integration** — Steam OAuth flow + game-ownership API latency.

---

## 12. Per-Phase SLI / SLO Definitions

| SLI | Definition | SLO Target | Window |
|-----|------------|------------|--------|
| Client session-start | First frame rendered after client launch | p99 ≤ 5 s | per-session |
| Wails desktop crash rate | Crashes per 1,000 sessions | < 1 | 7-day rolling |
| Compose-for-TV focus latency | Focus-target action response | p99 ≤ 100 ms | per-action |
| Steam Deck thermal envelope | Sustained 25 W operation | ≥ 99% sessions | 30-day rolling |
| Per-launcher OAuth success | Steam/GOG/Epic/Battle.net OAuth flow completion | ≥ 99% | 7-day rolling |
| Auto-update adoption | Clients on latest version | ≥ 90% within 48 h | per-release |

---

## 13. Per-Phase Operator Runbook

`HelixDevelopment/HelixClients/docs/runbook/phase05-operations.md` covering Wails per-OS deployment, Compose-for-TV Play Store + Amazon Appstore submission, Steam Deck thermal calibration, per-launcher OAuth troubleshooting, client-side crash log triage (Sentry-mirrored to operator's SIEM), per-client auto-update mirror configuration.

---

## 13a. Per-Phase Risk Mitigation Detail

### 13a.1 Wails desktop crash spike

**Detection:** `helix_client_crash_total` rate > 1 / 1,000 sessions per platform.

**Mitigation:** Sentry-mirrored telemetry with per-version per-OS crash grouping; pre-release crash budget gate (≤ 0.5% per beta release).

**Remediation:** Operator's runbook §2.5 — per-version crash forensics; rollback to prior version via auto-update if crash > 1% post-release.

### 13a.2 Compose-for-TV focus-target accessibility regression

**Detection:** Per-screen automated WCAG 2.2 SC 2.5.8 audit fails.

**Mitigation:** Per-PR CI gate runs Espresso accessibility tests; per-screen focus-target enumeration audited.

**Remediation:** Operator's runbook §3.3 — per-screen focus-target re-design with operator's accessibility consultant.

### 13a.3 Steam Deck thermal throttle excessive

**Detection:** Per-session sustained-power gauge < 25 W target on > 10% of sessions.

**Mitigation:** Per-game thermal profile auto-calibrated; user-overridable via Quick Access Menu.

**Remediation:** Operator's runbook §4.2 — per-game thermal profile re-calibration; common causes: ambient temp spike, dock-mode misconfiguration.

### 13a.4 Per-launcher OAuth flow regression

**Detection:** Per-launcher OAuth completion < 99% over 24-hour rolling window.

**Mitigation:** Per-launcher fallback URI + per-launcher capability cache; operator-side OAuth health dashboard.

**Remediation:** Operator's runbook §5.4 — per-launcher fallback path activation; common causes: launcher-side API breaking change (Battle.net URI broken since 2024 per Z-3), launcher-side OAuth credential rotation.

### 13a.5 Auto-update adoption stalled

**Detection:** Per-version adoption < 90% within 48 hours of release.

**Mitigation:** Forced-update flag for security-critical releases; per-tier opt-out for Enterprise tier (operator-mediated).

**Remediation:** Operator's runbook §6.7 — per-customer escalation if Enterprise tier > 7 days behind; common causes: customer firewall blocking auto-update mirror.

---

## 14. Implementation Considerations

### 14.1 Wails v2 vs Tauri-Go

Wails v2 is the canonical desktop client (per [C05 OQ-01 closure](../03_Architecture/04_Clients_and_App_Architecture.md)). Tauri-Go is queued as Phase 2 alternative; not in MVP scope.

### 14.2 Compose-for-TV vs Flutter

Compose-for-TV `androidx.tv.material3` 1.0 GA is the canonical Android-TV path (per [C12 MC-05 closure](../03_Architecture/11_TV_UX.md)). Flutter is fallback; not in MVP scope. Leanback is deprecated via 2026-08-31 64-bit Play Store mandate.

### 14.3 Steam Deck thermal envelope

25 W sustained operation per [C34 §6](../05_Video_Audio/04_DualPath_Encoding.md). Per-game thermal profile auto-calibrated; user-overridable via Steam Deck Quick Access Menu integration.

### 14.4 Per-launcher OAuth quirks

- Steam: standard OAuth + Web API key per-tenant.
- GOG: GalaxyAPI + per-user opt-in.
- Epic: EOS SDK + Bearer token.
- Battle.net: URI-broken since 2024 (per Z-3 in C08 §9 addendum) — per-game launcher fallback.

---

## 15. Phase_05 Cost Estimation

Phase_05 ships **client-side software** — operator's per-tenant cost is zero incremental (clients are free downloads).

Operator's app-store fees: Apple App Store + Google Play (Compose-for-TV) per-app-listing fee + 15-30% revenue share (per Phase_10 monetization model). Steam: free for self-distribution.

---

## 16. Cross-Mirror Parity Verification

Phase_05 closure verification per the [Phase_09 §16](Phase_09_Recording_and_Replay.md#16-cross-mirror-parity-verification) pattern.

---

## 17. Anti-Bluff Verification

### 11.1 Sources resolved

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`Phase_04_Streaming_Pipeline.md`](Phase_04_Streaming_Pipeline.md) |    230+ | 2026-04-30 | predecessor                                      |
| [`../06_Submodules/per-submodule/helix-tv-input.md`](../06_Submodules/per-submodule/helix-tv-input.md) | 309 | 2026-04-30 | TV input primitive                |
| [`../06_Submodules/per-submodule/helix-input.md`](../06_Submodules/per-submodule/helix-input.md) | 325 | 2026-04-30 | controller input                                |
| [`../06_Submodules/per-submodule/helix-display.md`](../06_Submodules/per-submodule/helix-display.md) | 321 | 2026-04-30 | frame pacing + VRR                              |
| [`../03_Architecture/04_Go_Client_Ecosystem.md`](../03_Architecture/04_Go_Client_Ecosystem.md) | 3,336 | 2026-04-30 | C05 architectural source                  |
| [`../03_Architecture/11_TV_UX.md`](../03_Architecture/11_TV_UX.md) | 3,273 | 2026-04-30 | C12 architectural source                                 |

### 11.2 Forbidden patterns

Clean.

### 11.3 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 (specification only).
- Pending: Phase_05 execution + operator signoff.

End of `09_Implementation_Phases/Phase_05_Clients.md` — 2026-04-30.
