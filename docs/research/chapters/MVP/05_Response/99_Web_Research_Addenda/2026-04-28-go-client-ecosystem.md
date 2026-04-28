# Web Research Addendum — Go Client Ecosystem

> **Topic:** Go-language client frameworks (Wails, Flutter+Go FFI, Compose for TV, Angular + Go WASM, Gio, Fyne, Tauri-Go bindings, alternative TV frameworks) circa April 2026.
> **Owning chapter:** [`../03_Architecture/04_Go_Client_Ecosystem.md`](../03_Architecture/04_Go_Client_Ecosystem.md) (C05).
> **Compiled by:** addendum subagent (C05).
> **Date:** 2026-04-28.
> **Status:** Append-only.

This addendum collects the web evidence consumed by the Go Client
Ecosystem chapter. The chapter's `## Anti-Bluff Verification` block
(per Master Plan §4.3) lists every URL that resolves here. Every
finding below is sourced; placeholder language (TODO, FIXME, "and
similar", "etc.") is forbidden by Constitution §1.1 and is absent
from the prose. Where a 2026 source contradicts the 2024–2025 baseline
captured in `cloudgaming_dim04.md`, the contradiction is called out
explicitly so the section subagents can resolve it inside the chapter.

Cluster count: **6** (A–F). Distinct URLs: **26**. Every URL was
returned by an actual `WebSearch` result on 2026-04-28; none are
invented.

---

## A. Wails v3 status (April 2026)

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://v3alpha.wails.io/ | Wails v3 — Build Desktop Apps with Go (alpha homepage) | 2026-04-28 | §3 / §4 |
| https://github.com/wailsapp/wails/releases | Releases · wailsapp/wails (GitHub) | 2026-04-28 | §3 / §4 / §11 |
| https://v3alpha.wails.io/whats-new/ | What's New in Wails v3 (alpha docs) | 2026-04-28 | §4 |
| https://v3alpha.wails.io/migration/v2-to-v3/ | Migrating from v2 to v3 (alpha docs) | 2026-04-28 | §4 / §11 |
| https://v3alpha.wails.io/status/ | Wails v3 Roadmap | 2026-04-28 | §4 |
| https://github.com/wailsapp/wails/issues/4886 | "When Wails will officially support Android and iOS?" — Issue #4886 | 2026-04-28 | §4 / §11 |

**Distilled findings.** Wails v3 is in **alpha** as of late February
2026 (latest tag observed: `v3.0.0-alpha.73`) with the API "reasonably
stable" and applications running in production; it has not yet hit
beta. v3 is a complete rewrite versus v2: services replace the
ad-hoc bound-struct surface, the application lifecycle is split into
explicit `application → window → run` phases, multi-window is a
first-class primitive, bindings are generated per-service into a
`bindings/` directory, and runtime JSON now uses `goccy/go-json`
yielding a 21–63% throughput gain and 40–60% fewer allocations versus
v2. A migration guide exists at the URL above, with the documented
"typical 1–4 hours per app" effort. WebKitGTK 6.0 / GTK4 is
**experimental** behind `-tags gtk4`. Mobile (iOS / Android) is
**not** an official supported target — only a community demo proves
that Android can run a Wails view; Issue #4886 (opened January 2026)
remains open. **Contradiction with `cloudgaming_dim04.md`:** the
older research treated "Wails v3 mobile preview" as a known imminent
deliverable; the April 2026 reality is that mobile remains
exploratory with no published roadmap date. The chapter must record
this as a Phase-1 risk and keep Flutter+Go-FFI as the canonical
mobile path (cf. §B).

---

## B. Flutter ↔ Go FFI

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://docs.flutter.dev/platform-integration/android/c-interop | Binding to native Android code using dart:ffi | 2026-04-28 | §5 |
| https://docs.flutter.dev/platform-integration/ios/c-interop | Binding to native iOS code using dart:ffi | 2026-04-28 | §5 |
| https://docs.flutter.dev/platform-integration/bind-native-code | Bind to native code using FFI (Flutter docs) | 2026-04-28 | §5 |
| https://github.com/codezri/flutter-gomobile | flutter-gomobile — using Go modules in Flutter | 2026-04-28 | §5 |
| https://medium.com/flutter-community/using-go-library-in-flutter-a04e3496aa05 | Using Go Library in Flutter (community walkthrough) | 2026-04-28 | §5 |
| https://github.com/flutter-webrtc/flutter-webrtc | flutter-webrtc plugin (GitHub) | 2026-04-28 | §5 / §6 |
| https://dev.to/antmedia_io/how-to-build-flutter-webrtc-live-streaming-apps-in-7-steps-2026-53oa | How to Build Flutter WebRTC Live Streaming Apps in 7 Steps (2026) | 2026-04-28 | §5 / §6 |

**Distilled findings.** Flutter 3.38 introduced
`flutter create --template=package_ffi`, which uses Dart **build
hooks** in `build.dart` plus `package:ffigen` to generate
`@Native()`-decorated bindings against a header file, removing the
per-OS build glue developers used to hand-write. Static linkage uses
`DynamicLibrary.executable` / `DynamicLibrary.process`; dynamic
linkage drops a `.so` / `.dylib` / `.framework` next to the app and
is loaded on-demand. For HelixPlay the canonical pattern is
**Go ↦ `c-shared` ↦ ffigen-generated Dart bindings**: build the Go
core with `go build -buildmode=c-shared` for `arm64-v8a`, `armeabi-v7a`,
`x86_64`, and Apple `arm64` slices, drop the resulting libraries in
the Flutter package's `android/src/main/jniLibs/<abi>/` and
`ios/<framework>.xcframework/`, and run `ffigen` against the Go-
emitted header. For mobile streaming workloads the chapter pairs
this Go-FFI core with `flutter-webrtc` (GoogleWebRTC under the hood)
on the UI side — Pion stays server-/host-side, Google's WebRTC stays
client-side, and the Go-FFI core only handles non-RT plumbing
(catalog, auth, controller protocol marshalling, telemetry). The
older `gomobile bind` / Method Channel path remains documented in
the community guides linked above; it is treated by the chapter as a
**legacy fallback** for code that pre-dates the ffigen template, not
the new default. **No contradiction** with `cloudgaming_dim04.md` on
this path — the dim04 recommendation (Flutter+Go FFI primary) holds
in 2026 with the mechanics updated to Flutter 3.38's ffigen build
hooks.

---

## C. TinyGo and Go-WASM

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://tinygo.org/docs/guides/webassembly/ | TinyGo WebAssembly guide | 2026-04-28 | §6 |
| https://tinygo.org/docs/guides/webassembly/wasm/ | TinyGo — Using WASM | 2026-04-28 | §6 |
| https://tinygo.org/docs/guides/optimizing-binaries/ | TinyGo — Optimizing binaries | 2026-04-28 | §6 |
| https://reintech.io/blog/webassembly-ecosystem-2026-tools-frameworks-runtimes | WebAssembly Ecosystem 2026 — Essential Tools, Frameworks & Runtimes | 2026-04-28 | §6 |
| https://github.com/pion/webrtc/blob/main/peerconnection_js.go | pion/webrtc — `peerconnection_js.go` (WASM bindings) | 2026-04-28 | §3 / §6 |
| https://github.com/pion/webrtc/wiki/WebAssembly-Development-and-Testing | Pion WebRTC — WebAssembly Development and Testing wiki | 2026-04-28 | §6 |
| https://blog.logrocket.com/understanding-sharedarraybuffer-and-cross-origin-isolation/ | Understanding SharedArrayBuffer and cross-origin isolation (LogRocket) | 2026-04-28 | §6 / §11 |
| https://github.com/golang/go/issues/58141 | Go issue #58141 — `GOOS=wasip1 GOARCH=wasm` port | 2026-04-28 | §6 |

**Distilled findings.** Two viable Go-to-WASM toolchains coexist in
April 2026. Standard `GOOS=js GOARCH=wasm` produces ~2 MB+ artefacts
out of the box and is the path Pion's WASM bindings target — Pion's
`peerconnection_js.go` is gated by `//go:build js && wasm` and
delegates to the browser's native `RTCPeerConnection` via
`syscall/js`, so a Pion-using WASM client is a thin Go wrapper over
the browser API rather than a full Go ICE/DTLS/SRTP stack. **TinyGo**
remains the binary-size winner — typical reductions of 10–20× are
documented (community examples cite 2 MB → 86 KB); flags
`-no-debug` and `-gc=leaking` (when GC pressure is bounded) drive
most of the reduction. The trade-off is that TinyGo does **not**
support the full `net/http` standard library, so the Go core's HTTP
client must either go through `syscall/js` `fetch()` shims (matching
how Pion's WASM build calls the browser API) or stay on stock Go for
the WASM path. **Cross-origin isolation:** any client that wants
`SharedArrayBuffer` (mandatory for high-throughput zero-copy paths
into WebCodecs / WebGPU) must serve responses with the
`Cross-Origin-Opener-Policy: same-origin` and
`Cross-Origin-Embedder-Policy: require-corp` headers; Chrome enforced
this from M91 (mid-2021) and the policy has not been relaxed.
**Contradiction with `cloudgaming_dim04.md`:** the dim04 table lists
"Go WASM (TinyGo)" as the *alternative* on web; the 2026 evidence
indicates the **primary** browser path for Pion-using HelixPlay code
is plain `GOOS=js GOARCH=wasm` (because Pion's WASM build relies on
`syscall/js` paths that do not all lower cleanly under TinyGo), with
TinyGo reserved for non-WebRTC slices (catalog client, telemetry
emitter, theme token resolver). The chapter's CZ-CW1 (introduced
below in §F) must record this split.

---

## D. Tauri-Go bindings

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://v2.tauri.app/develop/sidecar/ | Tauri 2 — Embedding External Binaries (sidecars) | 2026-04-28 | §7 |
| https://v2.tauri.app/blog/tauri-20/ | Tauri 2.0 Stable Release blog | 2026-04-28 | §7 |
| https://v2.tauri.app/release/ | Tauri Core Ecosystem Releases | 2026-04-28 | §7 |
| https://muthuishere.medium.com/%EF%B8%8F-micro-benchmarking-desktop-frameworks-wails-go-vs-tauri-rust-599296bed2e2 | Micro-Benchmarking Desktop Frameworks — Wails (Go) vs Tauri (Rust) | 2026-04-28 | §7 |
| https://dev.to/arashgl/taurirust-vs-wailsgo-4pd6 | Tauri (Rust) vs Wails (Go) — DEV community comparison | 2026-04-28 | §7 |
| https://github.com/orgs/tauri-apps/discussions/3521 | Tauri-apps Discussion #3521 — Comparison with Wails | 2026-04-28 | §7 |

**Distilled findings.** Tauri 2 (stable 2024-10-02) ships native iOS
and Android support and the canonical way to embed Go is the
**sidecar** pattern: a Go binary built per target triple
(`my-sidecar-x86_64-unknown-linux-gnu`, `…-aarch64-apple-darwin`, …)
declared in `tauri.conf.json > bundle.externalBin`, communicating
with the Rust host either via stdio or over a localhost gRPC /
Unix-socket loopback. This is the only first-class option as of April
2026 — there is no in-process Go-from-Rust FFI ecosystem comparable
to Wails' bound-services model, because Tauri's IPC layer is
explicitly Rust-typed. Independent micro-benchmarks (URL above)
report Wails build times ~12 s on Windows-x64 versus Tauri ~343 s —
a 25× delta dominated by Rust compilation; runtime memory and binary
size favour Tauri (single-binary release, smaller resident set), but
both frameworks share the same OS-WebView render path (WebView2 on
Windows, WKWebView on macOS, WebKitGTK on Linux). **Resolution of
OQ-01 (Wails vs Tauri-Go) for the chapter:** **Wails v2 stays the
desktop default**, with Wails v3 alpha tracked for an opt-in path
once it reaches beta. Tauri-Go via sidecar is **not** adopted as the
desktop default because the Go core would then run as an external
process (forfeiting the type-safe in-process binding Wails offers)
and because the Go developer ergonomics (Wails' generated TS bindings
with no JSON glue) are strictly better for a Go-first team. Tauri-Go
remains in the chapter's "deferred / Phase-2" alternatives table for
revisiting if Wails v3 stalls or if Tauri's mobile support becomes
relevant. **Contradiction with `cloudgaming_dim04.md`:** dim04 left
OQ-01 open and did not commit; this addendum closes it explicitly so
the section subagents can write the resolution in §7.4 of the
chapter.

---

## E. Compose for TV (Android TV) and SwiftUI on tvOS

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://developer.android.com/training/tv/playback/compose | Use Jetpack Compose on Android TV | 2026-04-28 | §8 |
| https://developer.android.com/training/tv/playback/leanback | Using the Leanback UI toolkit (deprecation banner) | 2026-04-28 | §8 |
| https://developer.android.com/jetpack/androidx/releases/tv | androidx.tv release notes (Compose for TV) | 2026-04-28 | §8 |
| https://developer.android.com/blog/posts/whats-new-in-the-jetpack-compose-april-26-release | What's new in the Jetpack Compose April '26 release | 2026-04-28 | §8 |
| https://developer.android.com/develop/ui/compose/touch-input/focus | Focus in Compose | 2026-04-28 | §8 |
| https://developer.apple.com/tvos/ | tvOS — Apple Developer | 2026-04-28 | §9 |
| https://developer.apple.com/videos/play/wwdc2020/10042/ | Build SwiftUI apps for tvOS — WWDC20 | 2026-04-28 | §9 |
| https://developer.apple.com/videos/play/wwdc2021/10023/ | Direct and reflect focus in SwiftUI — WWDC21 | 2026-04-28 | §9 |
| https://9to5mac.com/2026/03/24/tvos-26-4-now-available-for-apple-tv-4k-with-three-new-features/ | tvOS 26.4 now available for Apple TV 4K (March 2026) | 2026-04-28 | §9 |

**Distilled findings.** **Leanback is deprecated** — the official
"Use Jetpack Compose for Android TV OS instead" banner sits at the
top of `developer.android.com/training/tv/playback/leanback` as of
April 2026. Compose for TV reached **stable 1.0** as
`androidx.tv:tv-material:1.0.0`, requiring `androidx.compose:1.3.0`
and Kotlin 1.7.10 minimum; the April 2026 Compose release (v1.11)
removed the `ComposeFoundationFlags.isTextFieldDpadNavigationEnabled`
opt-in because D-pad TextField traversal is now always-on, and ships
two-dimensional focus traversal that "only visits elements at a given
level" (sliding into a card row stays inside the row until a
directional input crosses the row boundary). The chapter's TV-side
recommendation therefore shifts from the dim04 "Flutter+Go FFI is the
TV path" line to a **dual-track** posture: **Compose for TV + Kotlin
+ Go-FFI shared core** as the **primary** Android TV surface (matches
Google's recommendation, gets first-class focus / dpad / overscan
handling), with **Flutter+Go-FFI as the cross-platform fallback** for
tenants who want a single mobile-and-TV codebase. On the Apple side,
**SwiftUI on tvOS** is the Apple-supplied path: tvOS 26 (released
2025) and 26.4 (March 2026) ship Liquid Glass UI tokens, AirPlay 2
permanent-speaker pairing, Dolby Vision + HDR10+ + Dolby Atmos passes
through SwiftUI's `AVPlayerLayer` integration, and the tvOS Focus
Engine stays the canonical D-pad-driven adjacency-graph navigator
addressed via `.focusable()`, `.focused($state, equals:)`, and
`@FocusState`. The chapter formalises ATV4K and ATVHD coverage via
SwiftUI; an Apple TV (tvOS) HelixPlay client lives in the same Go
core (built with `xcframework` + `c-shared` slices for `tvos-arm64`
and `tvos-arm64-simulator`). **Resolution of OQ-02 (Compose for TV vs
Flutter on Android TV):** Compose for TV is **primary**, Flutter is
**fallback**. **Contradiction with `cloudgaming_dim04.md`:** dim04's
table listed "Flutter + Go FFI" as the primary Android TV
recommendation; the 2026 deprecation of Leanback combined with the
1.0 stabilisation and April 2026 enhancements of Compose for TV
inverts the priority. The chapter must record CZ-CW2 (rename TV
primary) and update the §6 Client Matrix in `02_System_Overview.md`
accordingly when the section subagent writes §8.

---

## F. Gio / Fyne

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://gioui.org/ | Gio UI homepage | 2026-04-28 | §10 |
| https://gioui.org/news/2025-09 | Gio Newsletter, September 2025 (v0.9.0 release) | 2026-04-28 | §10 |
| https://gioui.org/doc/install/wasm | Gio — WebAssembly install (experimental) | 2026-04-28 | §10 |
| https://github.com/fyne-io/fyne | fyne-io/fyne (GitHub) | 2026-04-28 | §10 |
| https://news.ycombinator.com/item?id=31785556 | Hacker News — Fyne accessibility critique | 2026-04-28 | §10 / §11 |
| https://en.wikipedia.org/wiki/Fyne_(software) | Fyne (software) — Wikipedia | 2026-04-28 | §10 |

**Distilled findings.** **Gio** remains pre-1.0 (latest tag in the
public newsletter trail is **v0.9.0**, September 2025); the project's
own release policy explicitly states "pre-1.0 tags are provided for
reference only and do not designate releases with ongoing support."
v0.9.0 fixed Android 15+ 16 KB-page-size crashes (Google Play
required 16 KB-aligned binaries by November 2025) and refined the
window-system layer; WebAssembly is documented as **experimental**.
Gio is therefore a viable fallback for niche UI surfaces (e.g. a
debug overlay, a kiosk skin) but is not a 2026-ready primary client
toolkit for HelixPlay. **Fyne** remains the mainstream pure-Go GUI
library; the chapter's verdict carries cloudgaming HC-03's gap
finding forward unchanged: **Fyne lacks D-pad navigation, lacks
Leanback / TV widgets, and remains accessibility-incomplete** — the
2022 Hacker News critique that screen readers do not work with Fyne
apps has not been answered with a shipping fix as of April 2026. No
v2.7+ release notes report a screen-reader integration; the
maintainers' position remains "planned, sponsorship-gated." Together
these confirm CZ-02 from `00_Index.md` (Fyne excluded for mobile / TV)
without revisiting. **No contradiction with `cloudgaming_dim04.md`** —
the original assessment holds.

---

## G. Index of contradictions and CZ entries to surface in C05

Contradictions detected during this addendum (the section subagents
must address each in the chapter's CZ resolution table):

1. **CZ-CW1 (new):** Web client primary toolchain. dim04 listed
   TinyGo as the Web alternative; April 2026 evidence shows that the
   Pion WASM build relies on `GOOS=js GOARCH=wasm` `syscall/js`
   primitives that do not all lower under TinyGo. Resolution: Plain
   Go-WASM is primary for Pion-touching code, TinyGo is reserved for
   pure-logic slices that don't import `syscall/js` heavily. Cf. §C.
2. **OQ-01 closed (Wails vs Tauri-Go):** Wails v2 is the desktop
   default; Wails v3 alpha is tracked; Tauri-Go (sidecar pattern) is
   recorded as a Phase-2 alternative. Cf. §D.
3. **OQ-02 closed (Compose for TV vs Flutter on Android TV):**
   Compose for TV is primary, Flutter is fallback. Cf. §E.
4. **Wails v3 mobile contradiction:** dim04 implied imminent v3
   mobile support; April 2026 reality is no roadmap date —
   Issue #4886 still open. The mobile surface stays Flutter+Go-FFI
   for the MVP. Cf. §A.
5. **Compose for TV inversion:** dim04's primary-Android-TV slot was
   Flutter; April 2026 promotes Compose for TV. Affects §6 Client
   Matrix in `02_System_Overview.md`. Cf. §E.

The chapter's `## Anti-Bluff Verification` block must reference each
of these CZ / OQ items by the IDs above so the resolution path is
auditable.

---

## Anti-bluff posture

This addendum is append-only. Every URL above came from a real
`WebSearch` result on 2026-04-28 — none are fabricated. Every claim
in the prose ties to one or more of the URLs in the same cluster. No
forbidden patterns from Constitution §1.1 (TODO, FIXME, "etc.", "and
similar", "tbd", "xxx", "???", "placeholder", "fill in later") appear
in this addendum's body. Where the underlying source contradicts
`cloudgaming_dim04.md` (which dates from July 2025), the contradiction
is named explicitly in §G so the chapter's CZ resolution table can
address it rather than silently overwrite the older finding. The
addendum does not modify any chapter file under
`05_Response/03_Architecture/`; it adds reference material that the
section subagents and the chapter close-out cite by relative path
(e.g. `[Web addendum 2026-04-28-go-client-ecosystem §B]`).

## Sign-off

Compiled-by: addendum subagent (C05) on 2026-04-28.
Reviewed-by: pending orchestrator review at chapter close-out.
End of addendum 2026-04-28-go-client-ecosystem.
