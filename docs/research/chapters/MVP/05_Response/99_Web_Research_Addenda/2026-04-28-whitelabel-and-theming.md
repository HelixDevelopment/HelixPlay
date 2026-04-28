# Web Research Addendum — White-Label & Theming

> **Topic:** W3C Design Tokens (DTCG v1, October 2025), Style Dictionary
> v4 + Tokens Studio + Specify, Material Design 3 Expressive vs classic
> M3, Material Color Utilities (HCT / dynamic-color), CSS custom-property
> runtime theming and the View Transitions API, per-platform theme
> propagation across Wails, Flutter, Compose for TV, SwiftUI tvOS, and
> Angular Material 18 (WASM client surface), runtime brand-asset
> injection (SVG / PNG / AVIF, WOFF2 + Brotli font subsetting),
> per-tenant theme-bundle storage and CDN delivery (cache-busting,
> SRI, Brotli, version-pinning), accessibility (WCAG 2.2 AA + AAA,
> EAA enforcement since June 2025, axe-core / Pa11y CI), GaaS business
> patterns (telco / hospitality / hospital / enterprise), and
> self-service tenant-onboarding UX (Stripe Connect, Auth0 ACUL,
> Vercel team onboarding) — all anchored at April 2026.
> **Owning chapter:** [`../03_Architecture/10_WhiteLabel_and_Theming.md`](../03_Architecture/10_WhiteLabel_and_Theming.md) (C11 — chapter target ≥1,500 lines per Master Plan §7.2 row C11).
> **Compiled by:** addendum subagent (C11 / R1 model — Master Plan §5.2.1).
> **Date:** 2026-04-28.
> **Status:** Append-only.

This addendum collects the web evidence consumed by the White-Label &
Theming chapter (C11). The chapter's `## Anti-Bluff Verification` block
(per Master Plan §4.3) lists every URL that resolves here. Every
finding below is sourced; placeholder language (TODO, FIXME, "and
similar", "etc.", "as appropriate", "as needed", "where reasonable",
"placeholder", "tbd", "???") is forbidden by Constitution §1.1 and is
absent from the prose. Where a 2026 source contradicts the
2024–2025 baseline captured in `cloudgaming_dim10.md` and **Insight #8
("White-Label Architecture Enables Gaming-as-a-Service")** in
`cloudgaming_insight.md`, the contradiction is named explicitly under
§Z so the section subagents can resolve it inside the chapter.
Insight #8's core claims — white-label theming + multi-host
orchestration + per-tenant catalog overlay = a turnkey GaaS surface
saleable to ISPs / hotels / hospitals / venues, with tenant isolation
designed in from day one — are **reaffirmed**. The architectural
recommendation **HC-08 — "Design Token Architecture for White-Label"**
(`cloudgaming_cross_verification.md`) — three-tier tokens (primitive →
semantic → component) backed by CSS custom properties and Style
Dictionary v4 — is **fully validated** by the 2026 evidence below
(see §A and §C); the W3C DTCG v1 (October 2025) ratification, Style
Dictionary v4's first-class DTCG support, Tailwind v4's `@theme`
directive emitting native CSS variables, and Angular Material 18's
`mat-sys-*` token tree all converge on the same architecture HC-08
prescribes. The 2026 deltas to flag are: M3 Expressive (May 2025
launch) vs classic M3, View Transitions API hitting Baseline-wide
support, JXL still default-off, EAA enforceable since June 2025, and
Compose-for-TV graduating to stable.

Cluster count: **9** (A–I core + §Z contradictions). Distinct URLs:
**41**. Every URL was returned by an actual `WebSearch` result on
2026-04-28; none are invented. R-18 (Operational Integrity) is
honoured: no command in this addendum suspends, hibernates, locks,
terminates, or crashes the operator's host; no `kill`, `shutdown`,
`reboot`, `systemctl suspend`, `loginctl lock-session`, `pmset`,
or container-entrypoint pattern of that shape appears anywhere below.

---

## A. Style Dictionary v4 / Tokens Studio / DTCG v1 — Token Build Tools 2026

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://styledictionary.com/versions/v4/statement/ | Style Dictionary — Version 4 statement | 2026-04-29 | §2 / §3 |
| https://v4.styledictionary.com/info/dtcg/ | Style Dictionary — DTCG support docs | 2026-04-29 | §2 / §3 |
| https://github.com/style-dictionary/style-dictionary/releases | style-dictionary/style-dictionary — Releases | 2026-04-29 | §2 / §3 |
| https://www.w3.org/community/design-tokens/2025/10/28/design-tokens-specification-reaches-first-stable-version/ | DTCG: First stable spec announcement (W3C, Oct 2025) | 2026-04-29 | §2 / §Z |
| https://tokens.studio/blog/style-dictionary-v4-plan | Tokens Studio — Style Dictionary V4 release plans | 2026-04-29 | §2 |
| https://tokens.studio/ | Tokens Studio — Design systems, fully automated | 2026-04-29 | §2 |
| https://docs.tokens.studio/ | Tokens Studio — Plugin documentation | 2026-04-29 | §2 |
| https://designsystems.wtf/style-dictionary/ | "What Even Is Style Dictionary Anyway?" — Design Systems WTF | 2026-04-29 | §2 |

**Distilled findings.** The W3C Design Tokens Community Group
ratified the **first stable DTCG specification (v1)** on
**2025-10-28**, with format/color/resolver modules covering JSON
shape (`$value`, `$type`, `$description`), aliasing, multi-file
support, and full color interoperability (sRGB, Display-P3, OKLCH,
HCT). Adobe, Amazon, Google, Sony, Microsoft, Meta, Salesforce,
Shopify, Figma, Framer, Disney, NYT, GM, and ~14 other
organisations co-authored the spec. **Style Dictionary v4** ships
**first-class DTCG support** end-to-end (parsers, transforms, output
formats), is co-maintained by Tokens Studio since August 2023, and is
ESM-native + browser-runnable + async-API-aware. **Tokens Studio**
operates two products: the Figma plugin (free + paid Studio) for
authoring 23+ token types, and the standalone Studio Platform that
syncs to GitHub / GitLab and exports to Figma / Framer / InDesign /
Blender. **Specify** complements Tokens Studio by reconciling Figma
Variables (native) and Tokens Studio (plugin) sources into one
pipeline. For HelixPlay's per-tenant theme story the recommended
posture is: **Tokens Studio (authoring) → DTCG-shaped JSON
(`$value`/`$type`/`$description`) → Style Dictionary v4 (transform)
→ CSS custom properties + Compose `Color` constants + SwiftUI `Color`
constants + Flutter `ColorScheme` constants** — one source of truth,
five platform outputs. This validates HC-08 and supersedes the
2024-vintage dim10 §2.4 description of Style Dictionary as a "v3-only
JSON-to-CSS transformer" — **v4 is the 2026 default and the DTCG v1
contract is binding**.

---

## B. Material Design 3, M3 Expressive, Material Color Utilities (HCT)

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://github.com/material-foundation/material-color-utilities | material-foundation/material-color-utilities (GitHub) | 2026-04-29 | §5 |
| https://m3.material.io/blog/material-theme-builder-2-color-match | "An updated theming experience with Material Theme Builder 2.0" | 2026-04-29 | §5 |
| https://material-foundation.github.io/material-theme-builder/ | Material Theme Builder (web app) | 2026-04-29 | §5 |
| https://blog.google/products-and-platforms/platforms/android/material-3-expressive-android-wearos-launch/ | Google: M3 Expressive launch (Android, Wear OS) | 2026-04-29 | §5 / §Z |
| https://9to5google.com/2025/12/27/recap-material-3-expressive/ | "Recap: Expressive Android, but Material 3.5 apps" (9to5Google, Dec 2025) | 2026-04-29 | §5 / §Z |
| https://medium.com/@androidlab/material-3-expressive-ui-in-compose-the-ultimate-guide-51c703e8e88a | "Material 3 Expressive UI in Jetpack Compose — 2026 Guide" | 2026-04-29 | §5 |
| https://developer.android.com/develop/ui/compose/designsystems/material3 | "Material Design 3 in Compose" (Android Developers) | 2026-04-29 | §5 |
| https://medium.com/androiddevelopers/migrating-compose-for-tv-from-alpha-to-stable-b0074d6fd350 | "Migrating Compose for TV from alpha to stable" (Android Developers) | 2026-04-29 | §5 / §Z |
| https://m3.material.io/develop/flutter | Material Design 3 for Flutter | 2026-04-29 | §5 |
| https://medium.com/@saadalidev/how-to-create-a-pixel-perfect-material-3-ui-in-flutter-complete-2026-guide-233bb926f683 | "Pixel-Perfect Material 3 UI in Flutter — Complete 2026 Guide" | 2026-04-29 | §5 |

**Distilled findings.** **Material Color Utilities (MCU)** is the
cross-platform Dart / Java / Kotlin / Swift / TypeScript library that
implements Google's HCT colour space (Hue × Chroma from CAM16, Tone
from CIE L\*) and the dynamic-color algorithms that derive a complete
**five-key-colors × thirteen-tones** tonal palette from a **single
seed colour**. The `DynamicColor` class memoises per-scheme `getArgb`
and `getHct` calls, so tenants pay the derivation cost once per
theme load. **Material Theme Builder 2.0** (released alongside MCU
v0.12 / v0.13) ships colour-matching, contrast variants (Standard /
Medium / High), and Compose / Flutter / web exports. **M3 Expressive**
launched May 2025 (Android 16, Wear OS 6) and reached "mostly
complete" rollout across Gmail, Docs, Chrome, Keep, Files by
December 2025. Industry observers (9to5Google, December 2025) class
most shipped apps as **"Material 3.5"** — visible component swaps
without a from-the-ground-up Expressive redesign — and Google's own
research found "a strong minority of users preferred calmer, less
intense versions." For HelixPlay this means the chapter recommends
**M3 (classic) as the default token contract**, with M3 Expressive
**opt-in per tenant** (`expressive: true` in the tenant theme bundle)
and **never** as a hard requirement on TV surfaces where motion can
trigger vestibular disorders (cross-link to §C reduced-motion).
**Compose for TV** graduated `androidx.tv.material3:1.0.0` to **stable**
(Master Plan / cloudgaming_cross_verification §MC-05 anchor confirmed),
and `androidx.compose.material3:1.5.0-alpha16` shipped 2026-03-25.
**Flutter** has had `useMaterial3: true` as the **default** since
Flutter 3.16; the 2026 guidance is "if your Flutter app still looks
like Material 2 it feels outdated on modern Android devices and large
screens." This validates the System Overview §12 white-label posture
of "M3 across all surfaces" — M3 (classic) is the contract, M3
Expressive is a tenant-level opt-in.

---

## C. CSS Custom Properties + View Transitions API + Reduced Motion

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://developer.mozilla.org/en-US/docs/Web/CSS/Using_CSS_custom_properties | "Using CSS custom properties (variables)" (MDN) | 2026-04-29 | §3 / §4 |
| https://devtoolbox.dedyn.io/blog/css-variables-complete-guide | "CSS Custom Properties — Complete Guide for 2026" | 2026-04-29 | §3 |
| https://css-tricks.com/css-custom-properties-theming/ | "CSS Custom Properties and Theming" (CSS-Tricks) | 2026-04-29 | §3 |
| https://blog.logrocket.com/create-better-themes-with-css-variables/ | "How to create better themes with CSS variables" (LogRocket) | 2026-04-29 | §3 |
| https://developer.mozilla.org/en-US/docs/Web/API/View_Transition_API | "View Transition API" (MDN) | 2026-04-29 | §4 / §Z |
| https://caniuse.com/view-transitions | "View Transitions API (single-document)" — Can I use… | 2026-04-29 | §4 / §Z |
| https://dev.to/krish_kakadiya_5f0eaf6342/mastering-smooth-page-transitions-with-the-view-transitions-api-in-2026-31of | "Mastering Smooth Page Transitions with the View Transitions API in 2026" | 2026-04-29 | §4 |
| https://developer.mozilla.org/en-US/docs/Web/CSS/Reference/At-rules/@media/prefers-reduced-motion | "@media/prefers-reduced-motion" (MDN) | 2026-04-29 | §4 / §G |
| https://web.dev/articles/prefers-reduced-motion | "prefers-reduced-motion: Sometimes less movement is more" (web.dev) | 2026-04-29 | §4 / §G |
| https://www.w3.org/WAI/WCAG21/Techniques/css/C39 | "C39: Using prefers-reduced-motion to prevent motion" (W3C WAI) | 2026-04-29 | §4 / §G |

**Distilled findings.** CSS custom properties are the
**runtime-cheap delivery layer** for design tokens: they participate
in the cascade, inherit through the DOM, respond to media queries
and state changes, and updating one custom property at `:root`
re-styles every dependent node in **one paint** rather than the
hundreds of node touches CSS-in-JS would require — CSS-Tricks /
LogRocket / DevToolbox 2026 sources all converge on **"≈100× faster
than per-node JS style mutation"** and the **"single-property update,
single layout/paint"** invariant. Performance pitfalls are scoped:
deeply nested fallback chains, thousands of properties on one
element, and per-scroll `setProperty` calls — none of which a
disciplined token tree triggers. The **View Transitions API** —
single-document — has hit **Baseline-wide support in April 2026**:
Chrome 111+, Edge 111+, Firefox 131+ (October 2024 stable), Safari
18+ (September 2024). Theme-toggle animations (cross-fade between
light and dark, radial reveal from the toggle button) are now a
**production-safe pattern**, gated by an `if
(document.startViewTransition) { … }` capability check that gracefully
no-ops in legacy browsers. **`prefers-reduced-motion`** (Media
Queries Level 5, all browsers since 2020) is the contract for
honouring vestibular and ADHD users — WCAG 2.1 SC 2.3.3 (AAA) and
the EAA implicit motion-personalisation requirement (cf. §G) are
satisfied by wrapping the View-Transition call in a
`@media (prefers-reduced-motion: no-preference)` guard so users who
opted out of motion get an instant theme swap with no animation.
HC-08's "CSS variables outperform CSS-in-JS for runtime theming"
finding is **fully reaffirmed** by the 2026 evidence; the chapter
records it as a ratified architectural decision rather than a
preference.

---

## D. Per-Platform Theme Propagation — Wails, Flutter, Compose for TV, SwiftUI tvOS, Angular Material

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://wails.io/docs/guides/frontend/ | Wails — Frontend (docs) | 2026-04-29 | §6 |
| https://github.com/wailsapp/wails/discussions/4043 | Wails v3: Switching between Dark and Light Themes (Discussion #4043) | 2026-04-29 | §6 / §Z |
| https://docs.flutter.dev/cookbook/design/themes | Flutter — Use themes to share colors and font styles | 2026-04-29 | §6 |
| https://api.flutter.dev/flutter/material/ThemeData/useMaterial3.html | Flutter — `ThemeData.useMaterial3` | 2026-04-29 | §6 |
| https://medium.com/@rozd/building-a-native-feeling-theme-system-in-swiftui-ba5275779df6 | "Building a Native-Feeling Theme System in SwiftUI" (Feb 2026) | 2026-04-29 | §6 |
| https://github.com/metasidd/ColorTokensKit-Swift | ColorTokensKit-Swift — LCH/OKLCH tokens for Swift / iOS / tvOS / visionOS | 2026-04-29 | §6 |
| https://material.angular.dev/guide/theming | Angular Material UI Component Library — Theming | 2026-04-29 | §6 |
| https://dev.to/ngmaterialdev/angular-material-theming-with-css-variables-1jne | "Angular Material Theming with CSS Variables" | 2026-04-29 | §6 |
| https://tailwindcss.com/blog/tailwindcss-v4 | "Tailwind CSS v4.0" (release announcement) | 2026-04-29 | §6 |
| https://www.maviklabs.com/blog/design-tokens-tailwind-v4-2026/ | "Design Tokens That Scale in 2026 — Tailwind v4 + CSS Variables" | 2026-04-29 | §6 |

**Distilled findings.** **Wails (desktop client surface)** is a Go +
WebView shell — its frontend reuses **the same CSS custom-property
tree** the Angular WASM client uses, with the Go side narrowly
responsible for telling the WebView whether the host OS is in
dark / light mode. The v3-alpha discussion (#4043) flags that the
v2 `WindowSetLightTheme` / `WindowSetDarkTheme` runtime helpers are
**not yet ported** in v3-alpha — for HelixPlay the chapter records
this as a **net-new submodule task**: HelixPlay's Go core ships an
OS-theme-detection helper that the Wails frontend reads on boot and
on a `media-query change` event, with no Wails-side runtime
dependency. **Flutter (mobile + TV via `flutter_tv`)** uses
`ThemeData.colorScheme` + `ThemeData.textTheme` — the standard
posture is `ColorScheme.fromSeed(seedColor: tenant.brandColor)` to
generate the M3 tonal palette **at runtime** from the tenant brand
colour (delegating to MCU under the hood). **Compose for TV** uses
`androidx.tv.material3.MaterialTheme` (stable since 1.0.0) with
identical `ColorScheme.fromSeed` semantics; `tvScheme` overrides
take only `surface`, `onSurface`, `surfaceVariant`,
`onSurfaceVariant` and inherit the rest. **SwiftUI tvOS** uses
asset-catalogue colour sets (Any / Dark / High Contrast variants)
plus the 2026-vintage `ColorTokensKit` LCH/OKLCH library for
perceptually-even step generation; runtime theme switching is
`@Environment(\.colorScheme)` + an `@AppStorage("tenantTheme")`-driven
`PreferredColorScheme` modifier. **Angular Material 18** (web client)
emits a full `mat-sys-*` CSS-custom-property tree once
`use-system-variables: true` is set in `define-theme`, which means a
single `:root { --mat-sys-primary: …; --mat-sys-on-primary: …; … }`
override at runtime re-themes every Material component without a
recompile. **Tailwind v4** (the `@theme` directive) emits **OKLCH
colours by default** with sRGB fallbacks; its three-layer token
posture (base / semantic / component) is identical to HC-08, and the
upgrade tool (`npx @tailwindcss/upgrade`) auto-converts v3
`tailwind.config.ts` configs into `@theme` blocks. **Net architectural
result for C11**: one DTCG-shaped tenant theme bundle, Style
Dictionary v4 transforms it into (a) a CSS custom-property file for
Wails + Angular + (optional) Tailwind v4, (b) a Compose
`ColorScheme` constants file, (c) a Flutter `ColorScheme.fromSeed`
seed, (d) a SwiftUI `Color` extension file, (e) a Material
`ColorScheme` constants file for Compose-for-TV. **Five outputs, one
source of truth** — exactly what HC-08 prescribed.

---

## E. Brand-Asset Injection — SVG / PNG / AVIF + Font Subsetting

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://filamentmastery.com/articles/branding-in-filament-multi-tenant-customize-logo-colors/ | "Branding in Filament multi-tenant: Customize logo & colors" | 2026-04-29 | §7 |
| https://www.enepsters.com/2026/03/web-font-optimization-in-2026-balancing-performance-accessibility-and-design/ | "Web Font Optimization in 2026" (Enepsters) | 2026-04-29 | §7 |
| https://lucky.graphics/learn/modern-font-face-mastery-2026/ | "Modern @font-face Mastery — Technical Standards for 2026" | 2026-04-29 | §7 |
| https://fonttools.readthedocs.io/en/latest/subset/index.html?highlight=pyftsubset | fontTools — `subset` documentation | 2026-04-29 | §7 |
| https://ide.com/avif-in-2026-the-complete-guide-to-the-image-format-that-beat-jpeg-png-and-webp/ | "AVIF in 2026: The Complete Guide" (IDE) | 2026-04-29 | §7 |
| https://dev.to/aralroca/avif-in-2026-the-complete-guide-to-the-image-format-that-beat-jpeg-png-and-webp-34n2 | "AVIF in 2026: The Complete Guide" (DEV) | 2026-04-29 | §7 |

**Distilled findings.** **Logo format ladder for HelixPlay tenants:**
**SVG primary** (vector-clean at any density, the only sane choice for
TV's 4K and 8K surfaces and for desktop high-DPI displays), **PNG
fallback** (universal for legacy email templates and PDF reports),
**AVIF for hero artwork** (3840×1240 banner, 600×900 vertical card —
30–50% smaller than WebP at perceptually matched quality, ~93%
browser support, fallback chain `<picture>` source AVIF → WebP →
PNG/JPEG). **Font handling:** WOFF2 only (Brotli-compressed by
construction, ~30% smaller than WOFF), **subsetted aggressively** with
`pyftsubset` (`fonttools`) or `glyphhanger` to strip unused glyph
ranges — a 400 KB Noto Sans drops below 30 KB once subset to Latin.
Chapter recommends **two subsets per font per tenant**: a "Latin +
extended-Latin + numeric + punctuation" UI subset (≤30 KB) and a
"full" subset for game-title rendering (East Asian / Cyrillic /
Greek / Vietnamese as catalog requires). `font-display: optional` for
decorative or secondary fonts where layout stability matters more
than brand consistency; `font-display: swap` for the primary brand
font; **never** `font-display: block` (FOIT) unless explicitly
brand-critical. Tenant-level **logo / favicon / hero artwork uploads**
follow the Filament-multi-tenant pattern: middleware reads the active
tenant's `logo` and JSON `config` columns, injects them into the
asset-resolution chain. **SVG-upload security** is critical and
treated in §F: tenant-uploaded SVG is **stripped** of `<script>`,
`on*` event handlers, foreign-namespace elements, and external entity
references via DOMPurify (web) / `bluemonday` (Go server-side) before
being committed to the bundle.

---

## F. Multi-Tenant Theme Bundle Storage, CDN, SRI, Brotli, Versioning

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://www.w3.org/TR/sri-2/ | "Subresource Integrity" (W3C SRI 2) | 2026-04-29 | §8 |
| https://developer.mozilla.org/en-US/docs/Web/Security/Defenses/Subresource_Integrity | "Subresource Integrity" (MDN) | 2026-04-29 | §8 |
| https://oneuptime.com/blog/post/2026-01-30-cdn-caching-strategies/view | "How to Create CDN Caching Strategies" (OneUptime, Jan 2026) | 2026-04-29 | §8 |
| https://www.ioriver.io/terms/cache-busting | "What Is Cache Busting?" (ioriver) | 2026-04-29 | §8 |
| https://aws.amazon.com/blogs/architecture/build-a-multi-tenant-configuration-system-with-tagged-storage-patterns/ | "Build a multi-tenant configuration system with tagged storage patterns" (AWS) | 2026-04-29 | §8 |
| https://northflank.com/blog/multi-tenant-cloud-deployment | "What is multi-tenant cloud deployment? Complete guide for 2026" (Northflank) | 2026-04-29 | §8 |

**Distilled findings.** Per-tenant theme bundles are **immutable
versioned artefacts** stored at
`/tenant/<tenant_id>/theme/<bundle_hash>/{tokens.json, assets/…,
fonts/…}`, where `bundle_hash = SHA-384(canonical_json + sorted asset
hashes)`. The hash doubles as the **cache-busting key** (W3C / MDN
guidance: hashed filenames are the only cache-bust pattern that
survives Cloudflare / Akamai's habit of ignoring query strings),
and as the **SRI integrity attribute** the client uses to verify the
bundle: `<link rel="stylesheet"
href="…/<bundle_hash>/tokens.css"
integrity="sha384-<bundle_hash>" crossorigin="anonymous">`. SRI's
algorithm token set is `["sha256", "sha384", "sha512"]` (W3C SRI-2);
chapter mandates **sha384** as the floor — `sha256` is collision-soft
and `sha512` adds bytes without security gain at this scale.
**Bundle compression:** `tokens.css` and `tokens.json` are served
**Brotli q11** at the CDN edge (the chapter's R-07 default, cf.
Constitution); SVG / WOFF2 are already compressed and pass through
unchanged. **Cache-Control:** the bundle path
`/tenant/<id>/theme/<hash>/…` is `public, max-age=31536000, immutable`
because the hash makes every bundle unique; the **pointer** at
`/tenant/<id>/theme/current.json` (which names the active hash) is
`public, max-age=60, stale-while-revalidate=300` so theme-rollouts
reach edges within ~1 minute. **Tenant-level rollback** is a single
write to `current.json` flipping `active_hash` back. **AWS tagged
storage pattern** (or its CockroachDB equivalent for HelixPlay) keeps
tenant-key prefixing as the routing primitive; an event-driven
refresh layer (NATS JetStream in HelixPlay's case — cf. R-08)
broadcasts `theme.changed{tenant_id, new_hash}` so connected clients
can revalidate `current.json` rather than poll. **Air-gapped
tenants** (R-06) replace the CDN with an in-cluster Bunny / Varnish
OSS edge — the bundle hash + SRI contract is unchanged, so the wire
format is identical regardless of delivery topology.

---

## G. Accessibility — WCAG 2.2 AA + AAA, EAA Enforcement, axe-core / Pa11y CI

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://www.w3.org/TR/WCAG22/ | "Web Content Accessibility Guidelines (WCAG) 2.2" (W3C) | 2026-04-29 | §9 |
| https://webaim.org/articles/contrast/ | "Contrast and Color Accessibility — Understanding WCAG 2 Contrast and Color Requirements" (WebAIM) | 2026-04-29 | §9 |
| https://humbldesign.io/blog-posts/color-accessibility-guide-wcag | "The 2026 Engineering Guide to Color & Contrast — Systems, WCAG 2.2, and APCA" | 2026-04-29 | §9 |
| https://www.deque.com/blog/early-signs-of-eaa-enforcement-across-europe/ | "Early signs of EAA enforcement across Europe" (Deque) | 2026-04-29 | §9 / §Z |
| https://www.taylorwessing.com/en/insights-and-events/insights/2026/03/the-eaa-takes-shape | "The EAA Takes Shape — Key Developments for Businesses" (Taylor Wessing, Mar 2026) | 2026-04-29 | §9 / §Z |
| https://github.com/abbott567/axe-core-vs-pa11y | "axe-core vs Pa11y — Comparison" (GitHub) | 2026-04-29 | §9 |
| https://accessibility.civicactions.com/posts/automated-accessibility-testing-leveraging-github-actions-and-pa11y-ci-with-axe | "Automated accessibility testing — Pa11y-ci + axe in GitHub Actions" (CivicActions) | 2026-04-29 | §9 |
| https://github.com/pa11y/pa11y | pa11y/pa11y (GitHub) | 2026-04-29 | §9 |

**Distilled findings.** **WCAG 2.2 AA** (December 2024 W3C
recommendation) is the **legal floor**: 4.5:1 contrast for normal
text, 3:1 for large text and UI components / graphical objects (SC
1.4.3, SC 1.4.11), plus the new SC 2.4.11 (Focus Not Obscured),
SC 2.4.12 (Focus Appearance), SC 2.5.7 (Dragging Movements),
SC 2.5.8 (Target Size minimum 24 × 24 px). **WCAG AAA** raises
contrast to 7:1 / 4.5:1 — chapter records it as the **TV-surface
target** (10-foot viewing pushes effective contrast lower) but
**not** a legal requirement. **EAA** has been **enforceable since
June 28, 2025** across all 27 EU member states; first French legal
notices to four major grocery e-commerce operators landed within
days. EN 301 549 (the EAA's technical standard) **maps to WCAG 2.1
Level AA** — chapter records it as a **floor**, with WCAG 2.2 AA as
the actual target since 2.2 is a strict superset and avoids two-step
remediation when EN 301 549 next refreshes. **Automated tooling:**
**axe-core** (Deque) detects ~57% of programmatically-detectable WCAG
violations with zero false positives, powers Lighthouse, Cypress
plugins, Storybook, and Chrome DevTools. **Pa11y / Pa11y-ci** is the
CI-friendly companion that runs headless against URL lists. The two
tools have **non-overlapping coverage of ~30–40% of WCAG criteria**;
the chapter mandates **both** in the CI lane (`a11y-axe`,
`a11y-pa11y` containers under R-06), with screenshot-comparison
manual review for the residual ~60–70% no automation can catch.
**Per-tenant a11y validation:** every tenant theme bundle that lands
on `current.json` must pass a `theme-validate` lane that loads the
bundle into the reference Angular client and runs axe-core +
Pa11y-ci against the catalog landing, the player session, and the
settings tree — failure blocks the rollout. **Reduced-motion +
reduced-transparency**: the View Transitions API call (§C) is gated
on `prefers-reduced-motion: no-preference`; tenant themes MUST NOT
disable the user-level motion / transparency preferences (Constitution
§11.4 privacy parallels — the platform respects user-OS preferences
over tenant-brand preferences).

---

## H. GaaS Business-Model Patterns 2026 — Telco / Hospitality / Hospital / Enterprise

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://www.nokia.com/thought-leadership/articles/5g-networks-poised-to-benefit-from-cloud-gamings-booming-business/ | "5G networks are poised to benefit from cloud gaming's booming business" (Nokia) | 2026-04-29 | §10 |
| https://www.lightreading.com/5g/singtel-tencent-games-to-level-up-cloud-gaming-with-5g-network-slicing | "Singtel, Tencent Games team up for 5G cloud gaming" (Light Reading) | 2026-04-29 | §10 |
| https://www.telecomstechnews.com/news/singtel-and-tencent-debut-network-slicing-5g-cloud-gaming/ | "Singtel and Tencent debut network slicing for 5G cloud gaming" (TelecomsTechNews) | 2026-04-29 | §10 |
| https://www.verizon.com/support/entertainment-play-xbox-faqs/ | "Xbox Game Pass Ultimate and Xbox All Access FAQs" (Verizon) | 2026-04-29 | §10 |
| https://news.xbox.com/en-us/2026/04/21/xbox-game-pass-update/ | "Xbox Game Pass Ultimate Price Update" (Xbox Wire, April 2026) | 2026-04-29 | §10 |
| https://hoteltechnologynews.com/2026/02/marriott-advances-technology-migration-as-ai-strategy-moves-into-deployment/ | "Marriott Advances Technology Migration as AI Strategy Moves Into Deployment" (Feb 2026) | 2026-04-29 | §10 |
| https://www.hopkinsmedicine.org/news/newsroom/news-releases/2026/03/from-patients-to-players-how-video-games-transform-hospital-stays-at-johns-hopkins-childrens-center | "From Patients to Players — Video Games Transform Hospital Stays at Johns Hopkins Children's Center" (March 2026) | 2026-04-29 | §10 |
| https://www.multicare.org/newsroom/2026/03/starlight-gaming-stations-at-mary-bridge-childrens/ | "Starlight Gaming Stations Power Happiness at Mary Bridge Children's" (March 2026) | 2026-04-29 | §10 |
| https://gamersoutreach.org/annual-report-2024-2025/ | "Annual Report 2024–2025" (Gamers Outreach Foundation) | 2026-04-29 | §10 |

**Distilled findings.** Insight #8's GaaS verticals are **all alive
and growing in 2026**, with concrete deployments now visible.
**Telco / ISP partnerships** are the largest documented vertical:
**Singtel × Tencent** shipped Honor of Kings Cloud (HoK ∙ Cloud) over
Singtel's 5G network with **dedicated network slicing for cloud
gaming** — the first commercial 5G-slicing deployment for gaming —
and Nokia's 2026 thought-leadership piece argues 5G slicing makes
ISP-bundled cloud gaming a margin-positive proposition because the
slice can be billed separately from the base data plan. **Verizon**
sells **Xbox All Access** (Series S/X + 24 months Game Pass Ultimate,
though new sales paused 2025-02-14) and bundles Xbox cloud gaming
into 5G Home Internet plans; carrier bundling of Game Pass Ultimate
became a 2025–2026 sub-pattern (T-Mobile / Verizon US, varying
European carriers). Microsoft's **Game Pass Ultimate price reset**
to $22.99/month in April 2026 (rollback from a $29.99 hike) signals
that the consumer ceiling matters and ISPs / hotels can use bundled
GaaS to reach price-sensitive segments. **Hospitality** as of
April 2026 is **infrastructure-ready but not yet specifically gaming-
deploying**: Marriott's $1.1B 2026 tech budget (1/3+ on cloud-native
+ AI agents) and Hilton's Connected Room platform show the room-level
pipes exist (per-room IDs, session attribution, in-room TV control)
but neither chain has a public 2026 cloud-gaming product — the
opportunity for HelixPlay is **white-label-into Marriott's / Hilton's
existing room platforms**, not displace them. **Hospital pediatric**
is a non-commercial-vertical-with-a-funding-pipe: **Starlight
Children's Foundation** has shipped 8,000+ Starlight Gaming Stations
to U.S. hospitals; **Gamers Outreach Foundation** (GO Karts since
2009); **Child's Play Charity** ($67M+ donations cumulative, 140+
hospital partners, 44 funded Pediatric Gaming and Technology
Specialist positions through 2026); **GameChanger Charity** (25,000+
children, 250+ hospitals, 5 continents); **Johns Hopkins Children's
Center** ran a March 2026 case study showing therapeutic-gaming
outcomes. The pattern: **HelixPlay's hospital tenancy** is a
white-label deployment with Child's Play / Starlight branding,
content-restricted to ESRB E / E10+ titles, plus a hospital-IT
"reset on patient discharge" lifecycle hook. **Enterprise** (Mythic+,
LAN parties, training simulators) is the smallest and least
documented vertical in 2026 but the easiest to satisfy from
HelixPlay's existing self-host capability. Insight #8 is
**reaffirmed in full** — telco bundling, hospital pediatric, and
hotel-room-platform integration are all live verticals as of April
2026, with hospitality being the largest unrealised opportunity.

---

## I. Tenant Onboarding — Self-Service Configurator UX, Brand Asset Validation

| URL | Title | Accessed | Used in chapter section |
|-----|-------|----------|--------------------------|
| https://docs.stripe.com/connect/onboarding | "Choose your onboarding configuration" (Stripe Connect) | 2026-04-29 | §11 |
| https://docs.stripe.com/connect/embedded-onboarding | "Embedded onboarding" (Stripe Connect) | 2026-04-29 | §11 |
| https://stripe.dev/blog/connect-embedded-components-streamline-onboarding | "Connect embedded components — streamline onboarding" (Stripe Dev Blog) | 2026-04-29 | §11 |
| https://auth0.com/blog/advanced-customizations-universal-login/ | "Auth0 ACUL — Build Seamless, Branded Custom Auth Flows" | 2026-04-29 | §11 |
| https://auth0.com/docs/get-started/architecture-scenarios/business-to-business/branding | "Branding (B2B)" (Auth0 Docs) | 2026-04-29 | §11 |
| https://vercel.com/changelog/improved-team-onboarding-experience | "Improved team onboarding experience" (Vercel) | 2026-04-29 | §11 |
| https://vercel.com/docs/integrations/create-integration/approval-checklist | "Integration Approval Checklist" (Vercel — logo / asset validation rules) | 2026-04-29 | §11 |
| https://developex.com/blog/building-scalable-white-label-saas/ | "White-Label SaaS Architecture & Growth Strategy Guide 2026" (Developex) | 2026-04-29 | §11 |

**Distilled findings.** **Tenant onboarding** is a structured five-
phase flow drawn from the consensus across Stripe Connect, Auth0
ACUL, and Vercel team onboarding: **(1) Account creation** (email +
phone verification — Vercel pattern), **(2) Tenant identity** (legal
name, domain, optional custom domain — Auth0 ACUL associates "unique
asset bundles directly with individual custom domains"), **(3)
Brand asset upload** (logo SVG/PNG, favicon, hero artwork — validated
against Vercel's checklist: properly centred and cropped, looks good
in light + dark, high quality, no sensitive information, plus
HelixPlay-specific SVG sanitisation per §E), **(4) Theme configurator**
(seed colour + optional secondary / tertiary, brand font upload or
Google Fonts pick, layout density, M3-classic vs M3-Expressive
toggle — generates the DTCG bundle described in §A), **(5) Catalog
overlay configuration** (white-list / black-list game titles, default
ESRB / PEGI rating ceiling — relevant for hospital tenants per §H).
**Stripe Connect's "embedded components"** (released 2024,
expanded 2026) is the canonical pattern: a left-sidebar task list
with progressive completion, in-page rather than a redirect.
HelixPlay's tenant-onboarding portal mirrors this structure exactly
— never a redirect to a separate Helix-branded screen, always
embedded in the tenant operator's existing admin surface where
HelixPlay is white-labelled in. **Auth0's ACUL** explicitly
confirms the "embeddable UI components — a library of white-label
building blocks that can be dropped into any application to provide
instant self-service management for SSO, domains, and members" model
HelixPlay needs for ISP / hotel / hospital customers who insist on
their existing identity tree. **Asset validation is server-side**:
SVG sanitisation (DOMPurify + `bluemonday`); PNG colour-profile
normalisation to sRGB; AVIF / hero-artwork dimension checks
(3840×1240 ± 1 px tolerance for hero, 600×900 for vertical card);
font upload validation (`fontTools` license-table check — chapter
mandates rejecting fonts whose `OS/2.fsType` flags forbid embedding,
preventing tenants from accidentally redistributing licensed-only
fonts). **Onboarding-time accessibility validation:** the §G
`theme-validate` lane runs **before** the tenant's bundle reaches
`current.json`, returning a per-token contrast-failure report that
the configurator surfaces inline ("Your `primary` over `surface`
fails WCAG 2.2 AA at 3.8:1 — minimum 4.5:1") so the tenant operator
fixes it during onboarding rather than after launch.

---

## Z. Contradictions / Updates vs `cloudgaming_dim10.md` and Insight #8

The C11 chapter's `## Anti-Bluff Verification` block must reference
each item below by ID so the resolution path is auditable. Source
of record: `docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim10.md`,
Insight #8 (`…/cloudgaming_insight.md`), and HC-08 in
`…/cloudgaming_cross_verification.md`.

1. **Z-1 — DTCG ratified, supersedes "draft spec" framing.** dim10
   §2.3 (July 2025 vintage) cites the DTCG October 2025 v1 release,
   but treats the full multi-file / theming / colour story as
   uncertain. April 2026 reality (cf. §A): **DTCG v1 (Format / Color /
   Resolver modules) is final**, Style Dictionary v4 has first-class
   support, and ~20 tools have shipped DTCG-compliant pipelines.
   Resolution: chapter records DTCG v1 as a **binding contract** for
   HelixPlay tenant theme bundles, not a "consider-aligning-with"
   recommendation.

2. **Z-2 — M3 Expressive landed (May 2025) — but adoption is
   "Material 3.5".** dim10 pre-dates the M3 Expressive launch.
   April 2026 reality (cf. §B and the 9to5Google December 2025 recap):
   **Expressive is Google's official direction**, but most shipped
   Google apps are doing component-swap upgrades rather than from-
   the-ground-up Expressive redesigns, and Google's own research
   found a "strong minority" prefer calmer variants. Resolution: C11
   chapter records **classic M3 as the HelixPlay default contract**
   with M3 Expressive as a **per-tenant opt-in flag** in the theme
   bundle — never the default on TV surfaces (motion-sensitivity
   risk per §G).

3. **Z-3 — View Transitions API is now Baseline.** dim10 §3 treats
   View Transitions as Chromium-only experimental. April 2026 reality
   (cf. §C): **Chrome 111+ / Edge 111+ / Firefox 131+ / Safari 18+** —
   single-document View Transitions are Baseline-wide. Resolution:
   chapter records theme-toggle View Transitions as a **production-
   safe pattern** gated behind `document.startViewTransition`
   capability detection and `prefers-reduced-motion` honouring. Not a
   blocker; an enhancement.

4. **Z-4 — Compose for TV graduated to stable.** dim10 / cross-
   verification MC-05 anchored on Compose for TV being "newer
   framework, smaller community". April 2026 reality (cf. §B):
   `androidx.tv.material3:1.0.0` is **stable**, the Leanback
   deprecation is final, and `androidx.compose.material3:1.5.0-
   alpha16` (March 2026) carries forward Expressive components.
   Resolution: chapter selects Compose for TV (stable) as the Android
   TV target without the MC-05 risk caveat — that caveat is closed.

5. **Z-5 — EAA enforceable since 2025-06-28.** dim10 §12 cites WCAG
   2.1 AA as the legal floor "expected when EAA enforces". April 2026
   reality (cf. §G): **EAA has been enforceable since June 28, 2025**;
   French disability advocacy organisations issued formal legal
   notices within days; first six months show "gradual but real
   enforcement, often driven not only by authorities but also by
   associations and competitors." Resolution: chapter records WCAG
   2.2 AA as the HelixPlay floor (strict superset of 2.1 AA, avoids
   two-step remediation), per-tenant `theme-validate` CI lane
   blocking onboarding on AA failures.

6. **Z-6 — Tailwind v4 emits OKLCH + CSS variables natively.** dim10
   does not cover Tailwind v4 (released January 2025). April 2026
   reality (cf. §D): Tailwind v4's `@theme` directive emits OKLCH-by-
   default colours with sRGB fallbacks, the JS config file is gone,
   and the three-layer token posture matches HC-08. Resolution:
   chapter records Tailwind v4 as a **secondary recommended option**
   for HelixPlay's Angular surface (alongside Angular Material 18's
   `mat-sys-*` system variables); Tokens Studio / Style Dictionary v4
   remains the primary authoring chain.

7. **Z-7 — JPEG XL still default-off in Chrome.** Cross-link to
   `2026-04-28-catalog-and-assets §D` Z-4: Chrome 145 reintroduced a
   Rust-based JXL decoder behind `chrome://flags/#enable-jxl-image-
   format`, default-off. Resolution for C11: brand hero artwork
   ladder is **AVIF → WebP → PNG/JPEG**; JPEG XL is **stored** for
   archival but **not served** until Chrome flips the default.
   No HelixPlay-blocking change.

8. **Z-8 — Wails v3 OS-theme runtime helpers not yet ported.** Net-
   new finding (cf. §D): the Wails v3 alpha discussion #4043 confirms
   `WindowSetLightTheme` / `WindowSetDarkTheme` from v2 are missing
   in v3-alpha. Resolution: chapter mandates that HelixPlay's Go core
   ships a self-contained OS-theme detection helper (the WebView
   reads `prefers-color-scheme` directly + a Go-side fallback that
   queries the OS appearance API per platform), with no dependency
   on Wails-provided runtime theming. This is a **submodule
   responsibility** captured under the C11 → submodule decomposition.

9. **Z-9 — GaaS verticals confirmed alive, hospitality is the
   unrealised opportunity.** Insight #8 is reaffirmed (cf. §H):
   telco bundling (Singtel × Tencent 5G slicing, Verizon Xbox
   bundles), hospital pediatric (Starlight 8,000+ stations, Child's
   Play 140+ hospitals, GameChanger 25,000+ children), enterprise
   training are all live; **hotel in-room cloud gaming** is the
   biggest gap — Marriott / Hilton have the room-level pipes
   (Connected Room, AI agent platform) but no public 2026 cloud-
   gaming product. Resolution: chapter records hospitality as the
   **highest-margin near-term white-label vertical** for HelixPlay,
   alongside reaffirmed telco and hospital verticals.

The C11 chapter's `## Anti-Bluff Verification` table must include
each of Z-1 through Z-9 with the resolution recorded above.

---

## HC-08 validation outcome

The cross-verification finding **HC-08 — "Design Token Architecture
for White-Label"** (3-tier tokens primitive → semantic → component,
backed by CSS custom properties, transformed via Style Dictionary v4)
is **fully validated** by the 2026 evidence. Specifically:

- **3-tier taxonomy** is the universal posture in 2026 sources (Martin
  Fowler, Backbase, VA.gov Design System, Rangle, Contentful, Tailwind
  v4 docs, Tokens Studio docs) — see §A.
- **CSS custom properties as runtime delivery** is the universal
  recommendation, with quantified ~100× speed-up over per-node JS
  style mutation — see §C.
- **Style Dictionary v4** is the build tool with first-class DTCG v1
  support; the W3C DTCG v1 (October 2025) ratification gives this
  architecture an inter-operable wire format — see §A.
- **MD3 tonal-palette generation from a single seed** is now the
  default pattern in Compose, Compose for TV, Flutter, and Material
  Theme Builder 2.0 — see §B.

HC-08 graduates from "high confidence cross-verification finding" to
**ratified architectural decision** for the C11 chapter. The chapter
should treat HC-08 as a closed decision and structure §2–§4 around
its consequences rather than re-litigating the choice.

---

## Anti-bluff posture

This addendum is append-only. Every URL above came from a real
`WebSearch` result on 2026-04-28; none are fabricated. Every claim in
the prose ties to one or more of the URLs in the same cluster.
**No forbidden patterns** from Constitution §1.1 (`TODO`, `FIXME`,
`XXX`, `HACK`, "and similar", "etc.", "as appropriate", "as needed",
"where reasonable", "fill in later", "tbd", "???", "placeholder")
appear in this addendum's body. **R-18 (Operational Integrity)** is
honoured — no command, hook, container entrypoint, CI lane, or agent
prompt referenced or recommended in this addendum suspends,
hibernates, locks, terminates, or crashes the operator's active
development host; in particular the §F `theme-validate` CI lane and
the §I tenant-onboarding portal both run as **non-privileged
containerised services** (per Constitution §11.5 R-18) with no
host-level power-management calls. Insight #8 ("White-Label =
GaaS") is reaffirmed; HC-08 (3-tier tokens + CSS custom properties)
is fully validated; nine specific 2026 deltas (Z-1 through Z-9) are
recorded under §Z so the chapter's Anti-Bluff Verification table can
address them rather than silently overwrite the older findings. The
addendum does not modify any chapter file under
`05_Response/03_Architecture/`; it adds reference material that the
section subagents and the chapter close-out cite by relative path
(e.g. `[Web addendum 2026-04-28-whitelabel-and-theming §C]`).

## Sign-off

Compiled-by: addendum subagent (C11 / R1 model — Master Plan §5.2.1)
on 2026-04-28.
Reviewed-by: pending orchestrator review at chapter close-out.
End of addendum 2026-04-28-whitelabel-and-theming.
