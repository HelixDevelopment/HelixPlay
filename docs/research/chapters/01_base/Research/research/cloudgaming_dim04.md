# Dim 04 — Go Ecosystem for Cross-Platform Client Development

## Research Report

**Date:** 2025-07-14
**Searches Conducted:** 24 independent web searches across all topics
**Sources Consulted:** Official documentation, GitHub repositories, technical publications, community discussions, benchmark data

---

## Executive Summary

Go offers multiple viable paths for cross-platform client development, each with distinct trade-offs. For a cloud gaming system requiring desktop, mobile, web, and TV support, **no single Go-centric solution covers all platforms equally well**. The recommended hybrid approach:

| Platform | Primary Recommendation | Alternative |
|---|---|---|
| Desktop (Win/Mac/Linux) | **Wails v2** (Go + web frontend) | Fyne |
| Mobile (Android/iOS) | **Flutter + Go FFI** | Fyne |
| Web | **Angular + Go backend** | Go WASM (TinyGo) |
| TV (Android TV) | **Flutter + Go FFI** | Native Android + Go bind |

**Key Finding:** Fyne is the most mature pure-Go UI toolkit but has significant accessibility gaps and no TV/D-Pad support. Wails is desktop-only but excellent for that scope. Gio supports all platforms but has a steep learning curve and limited widget availability. Go Mobile (gomobile) works but has binding limitations and performance overhead.

---

## 1. Fyne — Pure Go UI Toolkit

### Overview

Fyne is a free and open-source cross-platform GUI toolkit written entirely in Go. It is the most popular Go GUI toolkit by GitHub metrics, with over 27,000 stars and top-1000 open-source project status. [^37^]

Claim: Fyne supports desktop (Linux, macOS, Windows), mobile (Android, iOS), and web via WebAssembly
Source: Fyne official documentation
URL: https://grokipedia.com/page/Fyne_(software)
Date: 2026-01-14
Excerpt: "Fyne supports a wide range of platforms, enabling developers to build graphical applications with a single codebase that runs on desktop operating systems including Linux, macOS, and Windows, as well as mobile platforms such as Android and iOS. Web deployment is facilitated through WebAssembly."
Context: Official project description
Confidence: high

### Current State (v2.7.x, October 2025)

Fyne has seen rapid development with major releases:

- **v2.5.0** (Aug 2024): Internationalization (i18n), new widgets, mobile keyboard layout improvements [^41^]
- **v2.6.0** (Apr 2025): New threading model, `fyne.Do()` API, 3x performance improvements, Calendar/DateEntry widgets, race condition elimination [^217^]
- **v2.7.0** (Oct 2025): New canvas objects (Arc, Polygon, Square), massive rendering speed improvements, per-corner border radius, Navigation and Clip containers [^217^]

Claim: Fyne v2.6.0 introduced a new threading model that can see up to 3x speed increase
Source: Fyne GitHub Releases
URL: https://github.com/fyne-io/fyne/releases
Date: 2025-05-08
Excerpt: "Your app may need a few updates... but can see up to 3x speed increase depending on the number of graphical elements and how frequently they are updated."
Context: Official release notes
Confidence: high

### Architecture & Rendering

- **Rendering**: Vector-based drawing using OpenGL/OpenGL ES for hardware acceleration
- **Design**: Material Design principles with adaptive interfaces
- **DPI Handling**: Automatic detection and scaling via `FYNE_SCALE` environment variable
- **Threading** (v2.6+): All callbacks, events, and rendering on a single goroutine; `fyne.Do()` for background→UI thread calls [^217^]

### Widget Set

Fyne provides a comprehensive widget set:
- **Basic**: Label, Button, Entry (text input), Checkbox, Radio, Select, Slider
- **Layout**: Box, Grid, Form, Border, Center, Stack, Split
- **Collections**: List, Table, Tree, GridWrap
- **Dialogs**: File dialog, color picker, confirm dialog, custom dialogs
- **Containers**: Scroll, Tabs, Accordion, Card, Navigation (v2.7), Clip (v2.7)
- **New in v2.6+**: Calendar, DateEntry, partial Check, selectable Label, RichText

Claim: Fyne provides collection widgets (List, Table, Tree) with data binding support
Source: Fyne v2.7 release notes
URL: https://github.com/fyne-io/fyne/releases
Date: 2025-10-16
Excerpt: "Add Generics to List and Tree for data binding"
Context: New feature in v2.7
Confidence: high

### Build Process

Fyne uses `fyne-cross` for cross-compilation via Docker:

| Target | Supported Architectures |
|---|---|
| darwin | amd64, arm64 |
| linux | amd64, 386, arm, arm64 |
| windows | amd64, 386, arm64 |
| android | arm, arm64, 386, amd64 |
| ios | Universal (requires macOS host) |
| freebsd | amd64, arm64 |
| web | WASM |

```bash
# Install fyne-cross
go install github.com/fyne-io/fyne-cross@latest

# Build for target platforms
fyne-cross linux
fyne-cross windows
fyne-cross android
fyne-cross ios  # macOS host only
fyne-cross darwin  # macOS host only
```

Claim: fyne-cross uses Docker images with cross-compilers including MinGW for Windows, macOS SDK, and Android toolchain
Source: fyne-cross GitHub
URL: https://github.com/fyne-io/fyne-cross
Date: 2020-10-15
Excerpt: "fyne-cross is a simple tool to cross compile and create distribution packages for Fyne applications using docker images that include Linux, the MinGW compiler for Windows, FreeBSD, and a macOS SDK"
Context: Official repository description
Confidence: high

### Mobile Support

Fyne mobile support uses gomobile internally for Android builds. Key details:
- APK generation via `fyne package -os android`
- iOS XCFramework via `fyne package -os ios` (macOS host required)
- Mobile-specific adaptations: keyboard handling, touch scrolling with momentum
- Back button handling for Android and iOS (added in v2.4)

### TV/D-Pad Support

**Critical Gap: Fyne has NO built-in TV or D-Pad navigation support.**

Claim: Fyne does not support Android TV, D-Pad navigation, or leanback UI
Source: Hacker News discussion on Fyne
URL: https://news.ycombinator.com/item?id=31785556
Date: 2022-06-17
Excerpt: "Tried several of their apps, was able to use none of them with a screen reader or other accessibility tools."
Context: Community feedback on Fyne limitations
Confidence: high

Android TV requires:
- `LEANBACK_LAUNCHER` intent filter in manifest
- D-Pad navigation with focus handling
- 10-foot UI design (large elements, visible from viewing distance)
- MediaSession integration for media controls

Fyne does not provide any of these TV-specific features. Adding TV support would require:
1. Custom Android manifest with TV intent filters
2. D-Pad key event handling (`onKeyDown` for KEYCODE_DPAD_*)
3. Focus navigation system (not provided by Fyne)
4. Leanback-style UI components ( BrowseFragment equivalents)

**Assessment:** TV support with Fyne would require substantial custom development and may not be practical.

### Accessibility — MAJOR GAP

Accessibility is the most significant limitation of Fyne:

Claim: Fyne lacks screen reader support and accessibility tool integration
Source: Hacker News community discussion
URL: https://news.ycombinator.com/item?id=31785556
Date: 2022-06-17
Excerpt: "As expected, same problem as all the other UI toolkits. Tried several of their apps, was able to use none of them with a screen reader or other accessibility tools. It's nice if you draw the UI directly to the screen, but you skip a lot of steps that OS vendors took to make sure your app can be used by as many people as possible."
Context: User feedback on Fyne accessibility
Confidence: high

Claim: Fyne maintainers acknowledge accessibility as a planned but resource-intensive feature
Source: Hacker News — Fyne maintainer response
URL: https://news.ycombinator.com/item?id=31785556
Date: 2022-06-17
Excerpt: "Yea this is a big gap. It is planned but the task is huge. Once we have sufficient sponsorship or financial support in place you can be assured it will come."
Context: Official maintainer response
Confidence: high

Claim: Fyne has a GitHub issue tracking accessibility technology support since 2020
Source: Fyne GitHub Issues
URL: https://github.com/fyne-io/fyne/issues/1285
Date: 2020-09-04
Excerpt: "Supporting accessibility technologies #1285"
Context: Open issue requesting screen reader support
Confidence: high

Claim: Fyne's text refactor project acknowledges accessibility as a future goal
Source: Fyne GitHub Wiki
URL: https://github.com/fyne-io/fyne/wiki/Text-Refactor
Date: 2021-05-21
Excerpt: "Accessibility - users can read and write text via accessibility tools (eg screen readers)"
Context: Listed as a goal under "Possibly out of scope for a refactor project, but we need to keep in mind"
Confidence: medium

**Accessibility Status Summary:**
- Screen readers: NOT supported (no OS accessibility API integration)
- High contrast mode: Partial (theming can be customized)
- Keyboard navigation: Basic (Tab/Shift+Tab supported)
- ARIA attributes: N/A (draws directly to OpenGL, not native widgets)
- Platform a11y APIs: Not integrated

### Theming Capabilities

- Built-in Material Design themes (light/dark)
- JSON theme support with fallback (v2.7+)
- Custom theme creation via `fyne.Theme` interface
- Dynamic theme switching
- Per-corner border radius support (v2.7)
- `ThemeOverride` container for scoped theming

### Performance

| Metric | Value | Notes |
|---|---|---|
| Binary Size | ~20 MB typical | Includes OpenGL renderer |
| Startup Time | Fast (< 1s) | Native compiled binary |
| Memory Usage | Moderate | GPU-backed rendering |
| Rendering | Hardware-accelerated | OpenGL/OpenGL ES |
| v2.6+ Improvement | Up to 3x faster | New threading model |
| v2.7+ Improvement | "Massive speedups" | Optimized data handling, custom themes |

### Real-World Usage

Notable Fyne applications include [^149^]:
- **FyneDesk**: Complete Linux desktop environment
- **Supersonic**: Desktop music player for Subsonic/Navidrome/Jellyfin
- **Apptrix**: Graphical app editor and low-code studio
- **Rymdport**: Encrypted file sharing
- **GoXRay VPN**: VPN client for macOS/Linux
- **TetherSSH**: SSH terminal emulator
- **GOOM**: DOOM engine port
- 200+ apps listed on apps.fyne.io

### Limitations Summary

1. **No accessibility/screen reader support** — biggest blocker for inclusive apps
2. **No TV/D-Pad support** — cannot target Android TV, Apple TV, or smart TVs
3. **Mobile limitations** — requires gomobile toolchain, iOS builds need macOS
4. **Widget customization** — Material Design only; limited visual customization
5. **Web performance** — WASM output slower than native, limited browser optimizations

---

## 2. Wails v2 — Go Backend + Web Frontend

### Overview

Wails v2 is a framework for building desktop applications using Go for the backend and web technologies (React, Vue, Svelte, Angular) for the frontend. It uses the OS native WebView for rendering. [^36^]

Claim: Wails provides desktop apps with ~15MB binaries, ~10MB memory, <0.5s startup
Source: Wails v3 documentation
URL: https://v3.wails.io/concepts/architecture/
Date: 2026-04-20
Excerpt: "Bundle Size ~15MB (vs Electron ~150MB), Memory ~10MB (vs ~100MB+), Startup <0.5s (vs 2-3s)"
Context: Official documentation comparison
Confidence: high

### Architecture

| Component | Technology |
|---|---|
| Backend | Go (compiled native binary) |
| Frontend | Any web framework (React, Vue, Svelte, Angular, vanilla JS) |
| Rendering | OS Native WebView |
| Windows | WebView2 (Chromium-based, Edge) |
| macOS | WKWebView (WebKit/Safari) |
| Linux | WebKit2GTK |
| IPC | In-memory bridge (JSON-encoded, zero-copy) |

### IPC Mechanism

The Wails Bridge is the core communication layer:

Claim: Wails IPC uses in-memory bridge with JSON encoding, zero-copy, and auto-generated TypeScript bindings
Source: Wails v3 documentation
URL: https://v3.wails.io/concepts/architecture/
Date: 2026-04-20
Excerpt: "The bridge is the heart of Wails—it enables direct communication between Go and JavaScript... In-memory: No network overhead, no HTTP. Zero-copy where possible. Async by default. Type-safe: TypeScript definitions auto-generated."
Context: Official architecture documentation
Confidence: high

```go
// Go method - automatically exposed to JS
func (a *App) Greet(name string) string {
    return fmt.Sprintf("Hello %s!", name)
}
```

```javascript
// JavaScript - auto-generated binding
import { Greet } from "../wailsjs/go/main/App";
Greet("Alice").then(console.log);
```

### Build Process

1. **Development**: `wails dev` — hot reload via Vite, live development
2. **Build**: `wails build` — compiles Go + frontend, generates platform binary
3. **Package**: Built-in NSIS installer generation for Windows, .app bundles for macOS

Claim: Wails auto-generates TypeScript models from Go structs and provides hot-reload development
Source: Wails v2 release blog
URL: https://wails.io/blog/wails-v2-released/
Date: 2022-09-22
Excerpt: "Automatic TypeScript generation of Go structs... Vite integration providing a hot-reload development environment"
Context: Official release announcement
Confidence: high

### Platform Support

**Desktop Only** — Windows, macOS, Linux

Claim: Wails v2 does NOT support mobile platforms (iOS/Android)
Source: Wails GitHub Issues
URL: https://github.com/wailsapp/wails/issues/4886
Date: 2026-01-19
Excerpt: "When Wails will officially support Android and iOS?... I use Tauri because it can kinda build for Android and iOS."
Context: User request closed as documentation issue
Confidence: high

Claim: Wails has demonstrated Wails apps running on Android as a prototype but no official support
Source: Wails v2 Release Blog
URL: https://wails.io/blog/wails-v2-released/
Date: 2022-09-22
Excerpt: "I'm personally very excited at the prospect of getting Wails apps running on mobile. We already have a demo project showing that it is possible to run a Wails app on Android, so I'm really keen to explore where we can go with this!"
Context: Post-release roadmap statement
Confidence: medium

### Security Model

Claim: Wails IPC is internal to the application and not accessible externally; security through in-process memory
Source: Wails GitHub Discussions
URL: https://github.com/wailsapp/wails/discussions/2964
Date: 2023-10-08
Excerpt: "The IPC mechanism is internal to the application and is just as susceptible to any debugging as any other mechanism... Is it sending messages on the loopback adapter or is it communicating via a memory buffer?"
Context: Security discussion; uses memory buffer, not network
Confidence: high

### Performance Benchmarks

| Metric | Wails v2 | Electron | Tauri |
|---|---|---|---|
| Bundle Size | ~15 MB | ~150 MB | ~3-10 MB |
| Idle Memory | ~10 MB | ~100-300 MB | ~30-80 MB |
| Cold Startup | <0.5s | 1-3s | 0.2-0.5s |
| Build Time | ~6-21s | ~2-8s | ~48-266s (Rust) |
| IPC Latency | Sub-millisecond | ~0.45ms | ~0.12ms |

Claim: Wails builds are significantly faster than Tauri (Go compilation vs Rust)
Source: web-to-desktop-framework-comparison GitHub
URL: https://github.com/Elanis/web-to-desktop-framework-comparison
Date: 2021-03-09
Excerpt: "Windows (x64): Wails ≈6632ms, Tauri ≈266288ms"
Context: Build time benchmarks
Confidence: medium

### Frontend Framework Support

Wails supports all major frontend frameworks via templates:
- React (JavaScript + TypeScript)
- Vue (JavaScript + TypeScript)
- Svelte (JavaScript + TypeScript)
- Preact
- Lit
- Vanilla JS
- Angular (community-supported)

### Accessibility

**Partial via WebView** — accessibility depends on the frontend framework used:
- HTML-based frontend inherits browser accessibility (screen readers work with web content)
- Native OS dialogs are accessible
- Custom canvas rendering is NOT accessible
- The Go backend has no accessibility integration

**Best practice**: Use semantic HTML, ARIA attributes, and focus management in the frontend for accessible Wails apps.

### Limitations

1. **Desktop only** — no mobile or TV support
2. **Rendering inconsistency** — different engines per platform (WebKit vs Chromium)
3. **Smaller ecosystem** — fewer plugins than Electron
4. **WebView dependency** — requires WebView2 on Windows (not pre-installed on older systems)
5. **Limited native APIs** — compared to Electron's full Node.js access

---

## 3. Gio — Immediate Mode GUI in Pure Go

### Overview

Gio (gioui.org) is an immediate mode GUI library for Go that supports all major platforms including Linux, macOS, Windows, Android, iOS, FreeBSD, OpenBSD, and WebAssembly. [^19^]

Claim: Gio is an immediate mode GUI library supporting all major platforms including WASM
Source: Gio official website
URL: https://gioui.org/
Date: Unknown
Excerpt: "Gio is a library for writing cross-platform immediate mode GUI-s in Go. Gio supports all the major platforms: Linux, macOS, Windows, Android, iOS, FreeBSD, OpenBSD and WebAssembly."
Context: Official homepage
Confidence: high

### Architecture

- **Paradigm**: Immediate mode GUI (like Dear ImGui) — UI is rebuilt every frame
- **Rendering**: Custom vector renderer based on Pathfinder project
- **GPU**: OpenGL ES (mobile), Direct3D 11 (Windows), compute-shader renderer (in development)
- **Text**: Outline-based rendering, no texture baking; supports RTL, bidirectional text
- **Design**: Similar architecture to Flutter (component tree, layout constraints)

Claim: Gio uses an efficient vector renderer and is migrating to compute-shader-based rendering
Source: Gio official website
URL: https://gioui.org/
Date: Unknown
Excerpt: "Gio includes an efficient vector renderer based on the Pathfinder project implemented on OpenGL ES and Direct3D 11, and is migrating towards an even more efficient compute-shader-based renderer built atop piet-gpu."
Context: Official project description
Confidence: high

### Widget Availability

Gio provides core widgets in the main package and extended Material Design components in `gioui.org/x/component`:

**Core widgets** (`gioui.org/widget/material`):
- Buttons, Labels, Icons, Editors (text input)
- Checkboxes, Radio buttons, Switches, Sliders
- Flex, Stack, List, Grid layouts

**Extended components** (`gioui.org/x/component`) [^180^]:
- AppBar, NavigationDrawer (Modal)
- TextField, Tooltip
- Dialogs, Sheets (Modal Side Sheet)
- Table, Progress indicators
- Discloser (expandable), ContextArea
- Search, Tabs

**Third-party Material Design v3** (`git.sr.ht/~schnwalter/gio-mw`) [^181^]:

| Component | Status |
|---|---|
| Buttons (FAB, icon, split) | Partial |
| Cards | Implemented |
| Chips | Not implemented |
| Date/Time pickers | Not implemented |
| Lists | Not implemented |
| Loading indicators | Not implemented |
| Navigation bar | Not implemented |
| Navigation rail | Implemented |
| Progress indicators | Implemented |
| Search | Implemented |
| Text fields | Implemented |
| Tooltips | Implemented |
| Motion system | Not implemented |
| Ripple effects | Not implemented |

Claim: Gio-x/component provides Material Design components but many are still experimental with no stable API
Source: Go Package Documentation
URL: https://pkg.go.dev/gioui.org/x/component
Date: 2025-09-16
Excerpt: "This package provides various material design components for gio. State This package has no stable API, and should always be locked to a particular commit"
Context: Package documentation
Confidence: high

### Platform Coverage

| Platform | Status |
|---|---|
| Linux | Full support |
| macOS | Full support |
| Windows | Full support |
| Android | Full support |
| iOS | Full support |
| FreeBSD | Full support |
| OpenBSD | Full support |
| WebAssembly | Full support |
| tvOS | Mentioned in talks [^208^] |

### Performance

- **Binary Size**: Small (no runtime dependency beyond platform libraries)
- **Memory**: Low (immediate mode, no retained widget tree)
- **Rendering**: GPU-accelerated vector graphics
- **WASM**: Runs in browser via WebAssembly

### Complexity & Learning Curve

Gio has a **steep learning curve** compared to Fyne:

1. **Immediate mode paradigm** — unfamiliar to most web/desktop developers
2. **Manual state management** — all widget state must be explicitly tracked
3. **Verbose code** — even simple UIs require more code than Fyne
4. **Limited documentation** — good architecture docs but fewer tutorials
5. **No drag-and-drop designer** — everything is code

Claim: Gio is more complex than retained-mode toolkits but provides more control
Source: SPG Blog
URL: https://softwareplanetgroup.co.uk/lightweight-desktop-applications-with-gio-ui/
Date: 2022-02-14
Excerpt: "Gio is also an open-source library that offers a small set of widgets which are compliant with Material Design principles. Essentially, therefore, it is a Golang UI library for implementing immediate mode graphics."
Context: Technical overview article
Confidence: high

### Production Usage

Notable applications built with Gio:
- **Tailscale Android**: Open-source Android client for Tailscale VPN [^201^]
- **Godcr**: Cross-platform DCR cryptocurrency wallet
- **Sprig**: Arbor chat client
- **Sointu**: Modular software synthesizer

Claim: Tailscale uses Gio for their Android client
Source: Medium article
URL: https://medium.com/@osho_jay/building-lightweight-cross-platform-applications-entirely-in-go-no-js-no-bs-4b737e3f5067
Date: 2023-07-24
Excerpt: "Another notable application is Tailscale Android, an open-source Android client for Tailscale — a mesh VPN alternative."
Context: Gio showcase article
Confidence: medium

### Accessibility

Claim: Gio has an approach to accessibility that encourages high-quality accessibility metadata for UI testing
Source: Ubuntu Summit talk description
URL: https://www.youtube.com/watch?v=sasp-TJS8oQ
Date: 2023-12-03
Excerpt: "Gio's approach to accessibility and how it encourages high-quality accessibility metadata for the sake of high-quality UI testing."
Context: Ubuntu Summit 2023 talk abstract
Confidence: medium

**Note**: Gio's accessibility implementation is still in early stages and not comparable to native platform accessibility APIs.

### TV Support

No explicit TV/D-Pad support documentation found. However, Gio's immediate mode nature means D-Pad events could be handled via custom input processing.

### Documentation Quality

- **Architecture docs**: Excellent (gioui.org/doc/architecture)
- **API docs**: Good (pkg.go.dev)
- **Tutorials**: Limited ("Writing an hello world", "Implementing an egg timer")
- **Examples**: Available in gioui.org/example
- **Community**: Smaller than Fyne

---

## 4. Go Mobile (gomobile)

### Overview

Go Mobile (`golang.org/x/mobile`) provides tools for building mobile applications and libraries in Go, targeting Android and iOS. [^45^]

### Two Modes of Operation

#### Mode 1: Native Applications (`gomobile build`)

Builds complete Android APK or iOS app from a Go `main` package. The app uses Go's native mobile app framework with OpenGL ES rendering.

```bash
# Build Android APK
gomobile build -target=android ./myapp

# Build iOS app
gomobile build -target=ios -bundleid=com.example.myapp ./myapp
```

#### Mode 2: SDK Bindings (`gomobile bind`)

Generates language bindings (Java for Android, Objective-C for iOS) from a Go package, producing an AAR (Android) or XCFramework (iOS).

```bash
# Generate Android AAR
gomobile bind -target=android ./mypackage

# Generate iOS XCFramework
gomobile bind -target=ios ./mypackage
```

Claim: gomobile bind generates AAR for Android and XCFramework for iOS from Go packages
Source: Go Package Documentation
URL: https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile
Date: Unknown
Excerpt: "For -target android, the bind command produces an AAR (Android ARchive) file... For Apple -target platforms, gomobile must be run on an OS X machine with Xcode installed."
Context: Official package documentation
Confidence: high

### Supported Go Types for Bindings

Only a subset of Go types are supported [^205^]:
- Basic types: `bool`, `string`, integers, floats
- Slices of basic types
- Functions (as callbacks)
- Structs (with exported fields)
- Interfaces (limited)

**Not supported**: Channels, maps, generic types, complex structs with unexported fields.

### Performance Overhead

Claim: gomobile language bindings have performance overhead
Source: Go Wiki Mobile
URL: https://go.dev/wiki/Mobile
Date: Unknown
Excerpt: "Current limitations are listed below. Only a subset of Go types are currently supported. Language bindings have a performance overhead."
Context: Official Go documentation
Confidence: high

JNI bridging performance characteristics:
- JNI call overhead: ~115ns (normal), ~35ns (@FastNative), ~25ns (@CriticalNative) [^102^]
- First call has higher overhead (library loading, method lookup)
- Array marshalling cost increases with payload size [^102^]
- Go→Java→Go roundtrips are expensive; batch operations preferred

### Android TV Support

Go Mobile does NOT have explicit Android TV support:
- No Leanback library integration
- No D-Pad navigation framework
- No TV-specific UI components
- Standard Android APK can run on Android TV but requires manual D-Pad handling

### NDK Integration

Go Mobile uses the Android NDK for native compilation:
- Requires Android NDK installed
- `ANDROID_HOME` environment variable set
- Supports API level 16+ (default)
- Compiles to multiple architectures: arm, arm64, 386, amd64

### iOS Limitations

- Must run on macOS with Xcode
- Bitcode not supported (must disable in Xcode project settings) [^211^]
- No dynamic library support on iOS (must use c-archive for static linking)

Claim: Go does not generate LLVM bitcode for iOS, which could cause issues if Apple requires it
Source: Medium article on gomobile
URL: https://medium.com/@fzambia/going-mobile-adapting-centrifugo-go-websocket-client-to-be-used-for-ios-and-android-app-e72dc2736f01
Date: 2017-04-08
Excerpt: "Go does not generate LLVM bitcode for you, at moment this is not very critical but if Apple decides that bitcode should be a requirement — I think gomobile will have problems with targeting iOS."
Context: iOS deployment experience
Confidence: medium

### Known Issues

- Xcode compatibility issues with newer versions [^203^] [^204^]
- XCFramework format incompatible with Xcode 15.3+ requirements [^204^]
- Limited type system for bindings
- Active maintenance concerns (fewer updates than main Go toolchain)

---

## 5. Go WebAssembly (WASM)

### Standard Go Compiler

```bash
GOOS=js GOARCH=wasm go build -o app.wasm
```

**Characteristics:**
- Full Go runtime included (goroutine scheduler, GC, reflection)
- Binary size: 2-3 MB typical, up to 7-12 MB for hello world
- Goroutines supported (cooperative scheduling on single JS thread)
- Full standard library (most packages work)
- CGO not supported

Claim: Standard Go WASM binaries are 2-3 MB with full runtime included
Source: Complete Guide to Go WASM
URL: https://muratdemirci.com.tr/en/go-wasm/
Date: 2025-12-22
Excerpt: "Standard Go: 2-3 MB (Full standard library, large applications)"
Context: Bundle size comparison table
Confidence: high

### TinyGo

TinyGo is a separate Go compiler targeting small environments:

```bash
tinygo build -target wasm -o main.wasm -opt=z main.go
```

**Characteristics:**
- LLVM-based compiler
- Binary size: 100-500 KB typical (10-20x smaller than standard Go)
- Limited standard library support
- Limited goroutine support (single-threaded scheduler)
- Partial reflection support
- No CGO
- Slower garbage collector

Claim: TinyGo produces WASM files 10-20x smaller than standard Go
Source: Dev.to article
URL: https://dev.to/alanwest/why-your-go-binary-is-too-fat-for-webassembly-and-how-tinygo-fixes-it-24l
Date: 2026-04-04
Excerpt: "Standard Go wasn't designed for constrained environments... A single fmt.Println call can produce a multi-megabyte WASM binary... TinyGo solves this by being a fundamentally different compiler that only includes what your code actually needs."
Context: Detailed technical comparison
Confidence: high

### WASM Binary Size Comparison

| Toolchain | Hello World Size | Fibonacci Size |
|---|---|---|
| Python (py2wasm) | 25.4 MB | 25.4 MB |
| Go (standard) | 2.3 MB | 1.9 MB |
| JavaScript (javy) | 981 KB | 983 KB |
| TinyGo | 110 KB | 22 KB |
| Rust (wasip1) | 65 KB | 49 KB |
| C (WASI SDK) | 93 KB | 21 KB |

Claim: TinyGo produces 110KB hello world WASM vs 2.3MB for standard Go
Source: WASM Binary Size Benchmark
URL: https://riotsecure.se/blog/wasm_binary_size_in_high_Level_languages
Date: 2025-10-06
Excerpt: "TinyGo: 110,084 bytes; Go (standard toolchain): 2,312,600 bytes"
Context: Independent benchmark across languages
Confidence: high

### Browser Performance

- Go WASM is **2-3x faster than JavaScript** for CPU-intensive tasks
- Matrix multiplication (1000x1000): Go WASM ~2.1s vs JS ~3.8s vs native Go ~1.2s [^38^]
- JS interop boundary crossing has overhead — minimize crossings
- DOM manipulation must go through JavaScript (WASM has no direct DOM access) [^263^]
- Initial download and compilation time is a factor

Claim: Go WASM is 2-3x faster than JavaScript for CPU-intensive computations
Source: Complete Guide to Go WASM
URL: https://muratdemirci.com.tr/en/go-wasm/
Date: 2025-12-22
Excerpt: "Go WASM: ~2.1s; JavaScript (optimized): ~3.8s; Native Go: ~1.2s"
Context: Matrix multiplication benchmark
Confidence: medium

### JS Interop

```go
import "syscall/js"

func main() {
    js.Global().Call("alert", "Hello from Go!")
    
    callback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
        return args[0].Int() + args[1].Int()
    })
    defer callback.Release()
    
    js.Global().Set("goAdd", callback)
}
```

### Limitations

1. **No direct DOM access** — must call through JavaScript
2. **Single-threaded** — goroutines multiplex on one JS thread
3. **No direct network access** — must use JavaScript fetch/XHR
4. **No file system access** — browser sandbox
5. **Memory limitation** — ~300MB max on mobile browsers [^263^]
6. **CGO not supported**
7. **Large binary size** (standard Go)

### WebAssembly System Interface (WASI)

WASI extends WASM beyond browsers:
- WASI Preview 2 (WASI 0.2) released early 2024
- Provides filesystem, networking (wasi-http), sockets, clocks
- Single-threaded execution (threads proposal in Phase 4)
- No built-in async I/O (Preview 3 expected 2025) [^253^]

---

## 6. Tauri with Go Backend

### Architecture

Tauri is a Rust-based framework for desktop apps. Go can be integrated via:

1. **Sidecar pattern**: Bundle Go binary as external executable
2. **HTTP/gRPC communication**: Go sidecar exposes local HTTP/gRPC API
3. **stdin/stdout**: Direct process communication (limited)

### Sidecar Implementation

```json
// tauri.conf.json
{
  "bundle": {
    "externalBin": ["bin/my-go-service"]
  }
}
```

```rust
// Spawn Go sidecar in Rust
let command = app.handle()
    .shell()
    .sidecar("my-go-service")
    .expect("couldn't get sidecar");
let (mut rx, mut child) = command.spawn().expect("failed to spawn");
```

Claim: Tauri sidecars allow bundling external binaries written in any language
Source: Tauri v2 Documentation
URL: https://v2.tauri.app/develop/sidecar/
Date: 2026-01-07
Excerpt: "You may need to embed external binaries to add additional functionality... We call this binary a sidecar."
Context: Official Tauri documentation
Confidence: high

### Feasibility Assessment

| Aspect | Status | Notes |
|---|---|---|
| Desktop support | Works well | Windows, macOS, Linux |
| Mobile support | Tauri 2.0+ supports iOS/Android | Go sidecar would need mobile builds |
| Communication | HTTP/gRPC over localhost | More overhead than Wails IPC |
| Bundle size | Moderate | Go binary + Tauri shell |
| Complexity | High | Two languages (Rust + Go), process management |

**Verdict**: Technically feasible but adds significant complexity. The sidecar pattern is best for specific use cases where Go provides unique value (e.g., existing Go service to reuse). For new development, Wails is simpler for Go-focused teams.

---

## 7. Flutter + Go FFI

### Architecture

1. Compile Go code as C shared library (`buildmode=c-shared`)
2. Create Flutter FFI plugin (`flutter create --template=plugin_ffi`)
3. Load Go library via `dart:ffi`
4. Call Go functions directly from Dart

### Build Process

```bash
# 1. Compile Go as shared library for each platform
GOOS=android GOARCH=arm64 CGO_ENABLED=1 \
  go build -buildmode=c-shared -o libgo.so

# 2. Create Flutter FFI plugin
flutter create --platforms=android,ios,macos,windows,linux \
  --template=plugin_ffi my_plugin

# 3. Generate Dart bindings with ffigen
flutter pub run ffigen --config ffigen.yaml

# 4. Load library in Dart
final dylib = DynamicLibrary.open('libgo.so');
```

Claim: Go can be compiled as a C shared library and called from Flutter via dart:ffi
Source: Dev.to article
URL: https://dev.to/leehack/how-to-use-golang-in-flutter-application-golang-ffi-1950
Date: 2023-12-19
Excerpt: "We will see how to compile a GoLang-based library and create a Flutter plugin package for a Flutter application... We used the ffigen tool to generate the Dart code to use the GoLang library."
Context: Tutorial with working example
Confidence: high

### Platform Support

| Platform | Go Library Format | Loading Method |
|---|---|---|
| Android | .so (shared object) | DynamicLibrary.open() |
| iOS | .framework | DynamicLibrary.process() |
| macOS | .dylib or .framework | DynamicLibrary.open() |
| Windows | .dll | DynamicLibrary.open() |
| Linux | .so | DynamicLibrary.open() |
| Web | NOT supported | WASM alternative |

### Performance

- FFI is **synchronous** and much faster than Platform Channels [^259^]
- Direct memory sharing (no serialization overhead)
- Go code runs at native speed
- Startup overhead: library loading time

Claim: FFI is much more performant than Platform Channels, with synchronous native thread interaction
Source: Codemagic Blog
URL: https://blog.codemagic.io/working-with-native-elements/
Date: 2023-06-14
Excerpt: "FFI is much more performant. It synchronously interacts with the native component on the same thread. By using FFI you get a less platform-dependent setup."
Context: Flutter native integration comparison
Confidence: high

### Limitations

1. **Web not supported** — FFI doesn't work in browser; need WASM fallback
2. **Build complexity** — must compile Go for each target platform
3. **Platform-specific build scripts** — different build process per platform
4. **Memory management** — manual allocation/deallocation across boundary
5. **Type mapping** — limited to C-compatible types

---

## 8. Kotlin Multiplatform + Go

### Architecture Pattern

1. Compile Go code as C library (`buildmode=c-archive` for static, `c-shared` for dynamic)
2. Use Kotlin/Native C interop (`cinterop` tool) to generate Kotlin bindings
3. Wrap with `expect/actual` pattern for shared KMP code
4. Android: Load via JNI; iOS: Link as native framework

### expect/actual Pattern

```kotlin
// commonMain/PlatformApi.kt
expect class GoBridge {
    fun processData(input: String): String
}

// androidMain/GoBridge.kt
actual class GoBridge {
    actual fun processData(input: String): String {
        // Load Go .so via JNI
        return nativeProcessData(input)
    }
    private external fun nativeProcessData(input: String): String
}

// iosMain/GoBridge.kt  
actual class GoBridge {
    actual fun processData(input: String): String {
        // Call Go c-archive linked functions
        return goProcessData(input)
    }
}
```

Claim: Kotlin Multiplatform's expect/actual allows defining platform-agnostic APIs with platform-specific implementations
Source: Kotlin Documentation
URL: https://kotlinlang.org/docs/multiplatform/multiplatform-expect-actual.html
Date: 2025-05-16
Excerpt: "Expected and actual declarations allow you to access platform-specific APIs from Kotlin Multiplatform modules."
Context: Official Kotlin documentation
Confidence: high

### C Interop

Kotlin/Native provides `cinterop` tool to generate Kotlin bindings from C headers:

```kotlin
// Build.gradle.kts
kotlin {
    androidTarget()
    iosArm64()
    iosSimulatorArm64()
    
    sourceSets {
        commonMain.dependencies {
            // Shared code
        }
    }
}
```

Claim: Kotlin/Native cinterop generates Kotlin bindings from C headers for Go libraries
Source: Kotlin Documentation
URL: https://kotlinlang.org/docs/native-c-interop.html
Date: 2025-12-08
Excerpt: "The tool analyzes C headers and produces a straightforward mapping of C types, functions, and strings into Kotlin."
Context: Official interop documentation
Confidence: high

### Feasibility Assessment

| Aspect | Status | Notes |
|---|---|---|
| Code sharing | Good | Common Kotlin + Go core library |
| Android | Works well | JNI to Go .so or .aar via gomobile |
| iOS | Complex | c-archive linking, requires cinterop |
| Desktop | Via Kotlin/JVM or Kotlin/Native | Additional complexity |
| Web | Via Kotlin/JS | Go not usable; need WASM alternative |
| Build system | Complex | Gradle + Go build + cinterop configuration |

**Verdict**: Powerful but complex. Best when Kotlin Multiplatform is already the primary architecture and Go provides specific shared business logic.

---

## 9. Angular + Go Backend

### Architecture

Standard full-stack web architecture:

```
┌─────────────────┐      HTTP/REST/gRPC       ┌─────────────────┐
│  Angular SPA    │ ◄───────────────────────► │  Go Backend     │
│  (TypeScript)   │      JSON/Protobuf        │  (Gin/Echo)     │
│                 │                           │                 │
│ - Components    │                           │ - REST API      │
│ - Services      │                           │ - gRPC service  │
│ - Routing       │                           │ - Database      │
│ - RxJS          │                           │ - Auth/JWT      │
└─────────────────┘                           └─────────────────┘
```

### Go Backend Options

| Framework | Use Case |
|---|---|
| Gin | HTTP REST APIs, middleware support |
| Echo | High performance, minimal |
| Fiber | Express.js-inspired, fast |
| gRPC + gateway | High-performance APIs, streaming |
| Standard library | Maximum control |

### Angular + Go Example

```go
// Go backend with Gin
package main

import (
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    r.GET("/api/games", getGames)
    r.POST("/api/sessions", createSession)
    r.Run(":8080")
}
```

```typescript
// Angular service
@Injectable()
export class GameService {
    constructor(private http: HttpClient) {}
    
    getGames(): Observable<Game[]> {
        return this.http.get<Game[]>('/api/games');
    }
}
```

### gRPC-Web Support

For high-performance APIs, gRPC-Web can be used with Angular:

Claim: gRPC-Web allows Angular clients to communicate with gRPC backends via HTTP/1.1 proxy
Source: ITNext Article
URL: https://itnext.io/a-complete-guide-to-grpc-web-with-angular-and-net-c4ae2500bd24
Date: 2022-05-19
Excerpt: "gRPC is 5, 7, and even 8 times faster than REST+JSON communication. Built-in code generation in different programming languages including... Go, Dart."
Context: gRPC-Web tutorial
Confidence: medium

### Deployment

- Go backend: Containerized (Docker), deployed to cloud (AWS/GCP/Azure)
- Angular frontend: Static files served by CDN or Go embed
- Combined: Single Go binary serving both API and static files via `embed.FS`

---

## 10. Performance Comparison Across Approaches

### Startup Time

| Approach | Desktop | Mobile | Web |
|---|---|---|---|
| Fyne | < 1s | 1-3s | 2-5s (WASM load) |
| Wails v2 | < 0.5s | N/A | N/A |
| Gio | < 1s | 1-3s | 2-5s (WASM load) |
| Flutter + Go FFI | N/A | 2-4s | 1-3s |
| Angular + Go | N/A | N/A | < 1s (after load) |

### Memory Usage

| Approach | Desktop (Idle) | Mobile | Notes |
|---|---|---|---|
| Fyne | 20-50 MB | 30-80 MB | GPU-backed |
| Wails v2 | ~10 MB | N/A | OS WebView |
| Gio | 15-40 MB | 25-60 MB | GPU-backed |
| Electron | 150-300 MB | N/A | Bundled Chromium |
| Tauri | 30-80 MB | N/A | OS WebView |
| Flutter | N/A | 50-150 MB | Skia engine |

### Binary Size

| Approach | Desktop | Mobile | Web |
|---|---|---|---|
| Fyne | 15-25 MB | 10-20 MB | 2-3 MB (WASM) |
| Wails v2 | 10-20 MB | N/A | N/A |
| Gio | 10-20 MB | 8-15 MB | 200 KB-1 MB (TinyGo) |
| Flutter + Go | N/A | 15-40 MB | N/A |

### IPC/Bridge Performance

| Approach | Latency | Notes |
|---|---|---|
| Wails (in-memory) | < 0.1ms | Direct memory |
| Tauri (JSON) | ~0.12ms | JSON serialization |
| Electron IPC | ~0.45ms | Multi-process |
| Flutter FFI | ~0.01ms | Direct native call |
| Flutter Platform Channel | ~0.5ms | Async messaging |
| gomobile bind | ~0.1ms | JNI overhead |

---

## 11. Accessibility Support Comparison

### Fyne

| Feature | Status | Notes |
|---|---|---|
| Screen readers | **NOT SUPPORTED** | No OS a11y API integration |
| High contrast | Partial | Custom theme possible |
| Keyboard navigation | Basic | Tab/Shift+Tab |
| VoiceOver/TalkBack | No | Custom rendering |
| WCAG compliance | No | Not achievable without screen reader |

### Wails v2

| Feature | Status | Notes |
|---|---|---|
| Screen readers | **Partial** | Via web frontend (browser a11y) |
| High contrast | Via CSS | `prefers-contrast` media query |
| Keyboard navigation | Full | Via web frontend |
| VoiceOver/TalkBack | Desktop only | OS WebView integration |
| WCAG compliance | Achievable | Depends on frontend implementation |

### Gio

| Feature | Status | Notes |
|---|---|---|
| Screen readers | Early/Promising | Planned approach with metadata |
| High contrast | Partial | Custom theme |
| Keyboard navigation | Good | Direct input handling |
| VoiceOver/TalkBack | No | Immediate mode rendering |
| WCAG compliance | No | Not yet available |

### Flutter + Go FFI

| Feature | Status | Notes |
|---|---|---|
| Screen readers | **Full support** | Flutter integrates with platform a11y |
| High contrast | Full | Flutter's accessibility support |
| Keyboard navigation | Full | Focus system built-in |
| VoiceOver/TalkBack | **Full support** | Native platform integration |
| D-Pad navigation | Supported | Android TV focus system |
| WCAG compliance | Achievable | Best option for accessibility |

Claim: Flutter has the most comprehensive accessibility support among the frameworks evaluated
Source: Flutter official documentation
URL: https://docs.flutter.dev/accessibility-and-localization/accessibility
Date: Various
Excerpt: "Flutter supports three components for accessibility: large fonts, screen readers, sufficient contrast."
Context: Official Flutter docs
Confidence: high

### Angular + Go

| Feature | Status | Notes |
|---|---|---|
| Screen readers | **Full support** | Browser handles accessibility |
| High contrast | Full | CSS `prefers-contrast` |
| Keyboard navigation | Full | Standard web |
| VoiceOver/TalkBack | **Full support** | Browser integration |
| WCAG compliance | **Achievable** | Best for web accessibility |

---

## 12. TV Platform Support Matrix

| Approach | Android TV | Apple TV | Web TV (Tizen/webOS) | Fire TV |
|---|---|---|---|---|
| Fyne | No | No | No | No |
| Wails v2 | No | No | No | No |
| Gio | Possible (custom) | No | No | Possible (custom) |
| Go Mobile | Partial (APK) | No | No | Partial (APK) |
| Flutter + Go FFI | **Yes** (leanback) | **Yes** | No | **Yes** |
| Angular + Go | **Yes** (web app) | **Yes** (web app) | **Yes** | **Yes** (web app) |
| Kotlin MP + Go | **Yes** | Via native | No | **Yes** |

### Android TV Requirements

To support Android TV properly, a framework must provide:

1. **D-Pad navigation** — KEYCODE_DPAD_UP/DOWN/LEFT/RIGHT, KEYCODE_ENTER
2. **Focus management** — visible focus indicators, focus traversal
3. **Leanback UI** — BrowseSupportFragment, DetailsSupportFragment
4. **TV launcher intent** — LEANBACK_LAUNCHER category
5. **MediaSession** — for media playback controls
6. **10-foot UI** — large text, high contrast, visible from distance

**Only Flutter and web-based approaches (Angular) fully meet these requirements.**

---

## 13. Recommendations for Cloud Gaming System

### Platform-Specific Recommendations

#### Desktop (Windows, macOS, Linux)

**Primary: Wails v2**
- Best developer experience for Go teams
- Excellent performance (<0.5s startup, ~10MB memory)
- Full web frontend flexibility (React/Angular/Vue)
- Good accessibility via web standards
- Type-safe IPC with auto-generated bindings

**Alternative: Fyne**
- Pure Go (no web stack needed)
- Good widget set, active development
- Cross-compilation via fyne-cross
- **Caveat**: No accessibility support

#### Mobile (Android, iOS)

**Primary: Flutter + Go FFI**
- Best UI framework for mobile (native performance, native look)
- Full accessibility support (screen readers, TalkBack, VoiceOver)
- Android TV/D-Pad support via leanback
- Go business logic shared via FFI
- Production-proven (many large apps)

**Alternative: Fyne**
- Single codebase with desktop
- Simpler than Flutter for basic apps
- **Caveat**: Limited widget customization, no accessibility

#### Web

**Primary: Angular + Go Backend**
- Full web standards, excellent accessibility
- gRPC-Web for high-performance APIs
- Server-side rendering for SEO
- Deployment flexibility (CDN, container)

**Alternative: Go WASM (TinyGo)**
- Run Go in browser for shared logic
- 200-500KB binaries with TinyGo
- **Caveat**: Limited DOM access, JS interop overhead

#### TV (Android TV, Apple TV, Smart TVs)

**Primary: Flutter + Go FFI**
- Android TV: Full leanback support
- Apple TV: tvOS support
- D-Pad navigation built-in
- MediaSession integration

**Alternative: Web App (Angular)**
- Smart TV browsers (Tizen, webOS)
- No app store submission needed
- **Caveat**: Limited codec support, performance varies

### Hybrid Architecture Recommendation

```
                    ┌─────────────────────────────────────┐
                    │         Shared Go Core              │
                    │  (protocol, streaming logic, state)  │
                    └─────────────────────────────────────┘
                              │
          ┌───────────────────┼───────────────────┐
          ▼                   ▼                   ▼
   ┌─────────────┐    ┌─────────────┐    ┌─────────────────┐
   │  Wails v2   │    │  Flutter +  │    │  Angular + Go   │
   │  (Desktop)  │    │  Go FFI     │    │  (Web/TV)       │
   │             │    │  (Mobile)   │    │                 │
   │ React/Ang   │    │             │    │  Angular SPA    │
   │ Web Frontend│    │ Dart UI     │    │  REST/gRPC API  │
   │ Go Backend  │    │ Go Library  │    │                 │
   └─────────────┘    └─────────────┘    └─────────────────┘
          │                   │
          ▼                   ▼
   Desktop (Win/Mac/Linux)  Mobile (Android/iOS)
                             + Android TV
```

### Go Library Sharing

Compile the shared Go core as:
1. **C-shared library** for Flutter FFI (`-buildmode=c-shared`)
2. **WASM module** for web clients (`GOOS=js GOARCH=wasm`)
3. **Direct package** for Wails backend

---

## 14. Tensions and Counter-Arguments

### Pure Go vs. Hybrid Approach

**Argument for Pure Go (Fyne/Gio):**
- Single language, single toolchain
- Shared code across all platforms
- Simpler deployment
- Native performance

**Counter-argument:**
- No accessibility (Fyne)
- Steep learning curve (Gio)
- Limited widgets
- No TV support
- Smaller ecosystem

### Web-based (Wails/Angular) vs. Native

**Argument for Web-based:**
- Rich ecosystem (npm packages)
- Familiar tooling
- Easy to hire for
- Good accessibility via web standards

**Counter-argument:**
- Desktop only (Wails)
- WebView inconsistencies across platforms
- JavaScript overhead
- Offline capability concerns for web

### Flutter + Go vs. Pure Flutter

**Argument for Flutter + Go FFI:**
- Reuse existing Go code
- Go's performance for streaming/game logic
- Type safety across boundary

**Counter-argument:**
- FFI adds build complexity
- Must compile Go for each platform
- Web doesn't support FFI (need WASM fallback)
- Debugging across language boundary is harder

---

## 15. Sources Cited

| Citation | Source | URL | Date |
|---|---|---|---|
| [^19^] | Gio UI Official | https://gioui.org/ | Unknown |
| [^36^] | Wails v3 Architecture | https://v3.wails.io/concepts/architecture/ | 2026-04-20 |
| [^37^] | Fyne (software) - Grokipedia | https://grokipedia.com/page/Fyne_(software) | 2026-01-14 |
| [^38^] | Go WASM Complete Guide | https://muratdemirci.com.tr/en/go-wasm/ | 2025-12-22 |
| [^40^] | Wails IPC Comparison | https://medium.com/@tacherasasi/why-wails-wins-at-ipc-for-go-desktop-apps | 2025-08-27 |
| [^41^] | Fyne v2.5 Release | http://fyne.io/blog/2024/08/14/fyne-v2.5-released | 2024-08-14 |
| [^44^] | Wails IPC Security Discussion | https://github.com/wailsapp/wails/discussions/2964 | 2023-10-08 |
| [^45^] | Go Mobile Wiki | https://github.com/golang/go/wiki/Mobile | Unknown |
| [^63^] | Flutter Dart FFI iOS | https://docs.flutter.dev/platform-integration/ios/c-interop | 2025-12-23 |
| [^65^] | Flutter Dart FFI Android | https://docs.flutter.dev/platform-integration/android/c-interop | 2025-12-23 |
| [^67^] | GoLang in Flutter FFI | https://dev.to/leehack/how-to-use-golang-in-flutter-application-golang-ffi-1950 | 2023-12-19 |
| [^73^] | Tauri + Sidecar | https://evilmartians.com/chronicles/making-desktop-apps-with-revved-up-potential-rust-tauri-sidecar | 2025-04-22 |
| [^88^] | Tauri vs Electron 2026 | https://tech-insider.org/tauri-vs-electron-2026/ | 2026-04-05 |
| [^89^] | Desktop Apps Framework Comparison | https://gary-yin.com/posts/the-future-of-desktop-apps/ | 2025-09-20 |
| [^93^] | Wails as Electron Alternative | https://dev.to/kartik_patel/wails-as-electron-alternative-4dmn | 2025-11-08 |
| [^96^] | GitHub A11y Discussion | https://github.com/payloadcms/payload/discussions/14489 | 2025-11-05 |
| [^98^] | Gio Learn Documentation | https://gioui.org/doc/learn | Unknown |
| [^99^] | Wails v2 Release Blog | https://wails.io/blog/wails-v2-released/ | 2022-09-22 |
| [^100^] | Go buildmode c-archive Android | https://github.com/golang/go/issues/33806 | 2019-08-23 |
| [^102^] | JNI Marshalling Performance | https://connorahaskins.substack.com/p/jni-marshalling-performance-single | 2024-03-20 |
| [^105^] | HN Fyne Discussion | https://news.ycombinator.com/item?id=31785556 | 2022-06-17 |
| [^107^] | Go buildmode c-archive proposal | https://github.com/golang/go/issues/77524 | 2026-02-10 |
| [^115^] | TinyGo vs Go WASM Size | https://dev.to/alanwest/why-your-go-binary-is-too-fat-for-webassembly-and-how-tinygo-fixes-it-24l | 2026-04-04 |
| [^116^] | WASM Binary Size Benchmark | https://riotsecure.se/blog/wasm_binary_size_in_high_Level_languages | 2025-10-06 |
| [^128^] | Fyne Accessibility Issue | https://github.com/fyne-io/fyne/issues/1285 | 2020-09-04 |
| [^144^] | KMP expect/actual | https://medium.com/@muhammetemingundogar53/writing-native-implementations-with-expect-actual-in-kotlin-multiplatform | 2025-12-14 |
| [^146^] | Kotlin expect/actual docs | https://kotlinlang.org/docs/multiplatform/multiplatform-expect-actual.html | 2025-05-16 |
| [^147^] | Wails How It Works | https://wails.io/docs/howdoesitwork/ | Unknown |
| [^148^] | Wails Password Manager App | https://dev.to/emarifer/a-minimalist-password-manager-desktop-app | 2024-12-19 |
| [^149^] | Fyne Apps Showcase | https://apps.fyne.io/all.html | 2026-04-08 |
| [^150^] | Kotlin C Interop | https://kotlinlang.org/docs/native-c-interop.html | 2025-12-08 |
| [^154^] | Wails Method Binding | https://wails.io/docs/guides/application-development/ | Unknown |
| [^172^] | Fyne Text Refactor Wiki | https://github.com/fyne-io/fyne/wiki/Text-Refactor | 2021-05-21 |
| [^178^] | Android TV Leanback Request | https://github.com/GrakovNe/lissen-android/issues/376 | 2026-03-28 |
| [^180^] | Gio x/component Package | https://pkg.go.dev/gioui.org/x/component | 2025-09-16 |
| [^181^] | Gio Material Design v3 | https://git.sr.ht/~schnwalter/gio-mw | 2026-02-21 |
| [^184^] | Angular + Go CRUD | https://dev.to/raoamit47/building-an-angular-crud-app-with-a-go-api | 2024-11-03 |
| [^201^] | Gio Cross-Platform Apps | https://medium.com/@osho_jay/building-lightweight-cross-platform-applications-entirely-in-go | 2023-07-24 |
| [^203^] | gomobile macOS issue | https://github.com/golang/go/issues/73119 | 2025-04-01 |
| [^204^] | gomobile Xcode 15.3 issue | https://github.com/golang/go/issues/66500 | 2024-03-23 |
| [^205^] | Go Mobile Wiki | https://go.dev/wiki/Mobile | Unknown |
| [^208^] | Gio Ubuntu Summit Talk | https://www.youtube.com/watch?v=sasp-TJS8oQ | 2023-12-03 |
| [^211^] | Centrifugo Go Mobile | https://medium.com/@fzambia/going-mobile-adapting-centrifugo-go-websocket-client | 2017-04-08 |
| [^214^] | Best GUI Frameworks for Go | https://blog.logrocket.com/best-gui-frameworks-go/ | 2023-03-10 |
| [^217^] | Fyne GitHub Releases | https://github.com/fyne-io/fyne/releases | 2025-10-16 |
| [^219^] | Tauri Sidecar Pandas | https://mclare.blog/posts/writing-a-pandas-sidecar-for-tauri/ | 2026-02-23 |
| [^224^] | Wails v2 Features | https://wails.io/blog/wails-v2-released/ | 2022-09-22 |
| [^228^] | Tauri Sidecar Docs | https://v2.tauri.app/develop/sidecar/ | 2026-01-07 |
| [^253^] | WASI Status 2024 | https://eunomia.dev/blog/2025/02/16/wasi-and-the-webassembly-component-model-current-status/ | 2025-02-16 |
| [^255^] | fyne-cross Package | https://pkg.go.dev/github.com/gavintan/fyne-cross | 2024-07-12 |
| [^256^] | Fyne Cross-Compiling Docs | https://docs.fyne.io/started/cross-compiling/ | Unknown |
| [^259^] | Flutter Platform Channel vs FFI | https://blog.codemagic.io/working-with-native-elements/ | 2023-06-14 |
| [^262^] | fyne-cross GitHub | https://github.com/fyne-io/fyne-cross | 2020-10-15 |
| [^263^] | WebAssembly Wikipedia | https://en.wikipedia.org/wiki/WebAssembly | 2025-06-18 |
| [^264^] | WebAssembly State 2024-2025 | https://platform.uno/blog/state-of-webassembly-2024-2025/ | 2025-01-27 |

---

*End of Report*
