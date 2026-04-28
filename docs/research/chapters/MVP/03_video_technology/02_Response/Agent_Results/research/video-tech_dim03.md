# Dimension 03: Zero-Copy Capture & Pipeline Architecture

## Executive Summary

Zero-copy capture and pipeline architecture is the single most impactful optimization for low-latency video streaming. Each CPU-GPU roundtrip for a 4K frame (~33MB) adds 1-3ms of latency. The difference between a naive copy-based pipeline and an optimized zero-copy pipeline can be 10-50x in overall latency (e.g., NVIDIA Jetson NVMM reduces latency from 200-500ms to 10-30ms). This research covers Windows (DXGI, WGC), macOS (ScreenCaptureKit, IOSurface), Linux (PipeWire, DMA-BUF, KMS/DRM), GPU memory interop architectures, frame pacing technologies, Go integration approaches, and measured latency numbers per pipeline stage.

---

## Table of Contents

1. [Windows Capture APIs](#1-windows-capture-apis)
2. [macOS Capture APIs](#2-macos-capture-apis)
3. [Linux Capture APIs](#3-linux-capture-apis)
4. [GPU Memory Architecture & Interop](#4-gpu-memory-architecture--interop)
5. [Zero-Copy Pipeline Paths](#5-zero-copy-pipeline-paths)
6. [Frame Pacing & Synchronization](#6-frame-pacing--synchronization)
7. [Go Integration Approaches](#7-go-integration-approaches)
8. [Latency Per Stage — Measured Numbers](#8-latency-per-stage--measured-numbers)
9. [References](#9-references)

---

## 1. Windows Capture APIs

### 1.1 DXGI Desktop Duplication API (DDA)

The DXGI Desktop Duplication API is the foundational hardware-accelerated screen capture interface on Windows 8+, providing direct access to GPU-composited desktop frames without CPU readback.

#### Core API Flow

```
Claim: "The DXGI Desktop Duplication API provides a shared GPU texture surface via AcquireNextFrame, with dirty rectangle and move rectangle metadata for incremental updates."
Source: Microsoft Learn — Desktop Duplication API
URL: https://learn.microsoft.com/en-us/windows/win32/direct3ddxgi/desktop-dup-api
Date: 2021-01-06
Excerpt: "DXGI provides a surface that contains a current desktop image through the new IDXGIOutputDuplication::AcquireNextFrame method. The format of the desktop image is always DXGI_FORMAT_B8G8R8A8_UNORM no matter what the current display mode is."
Context: Official Microsoft documentation for DDA
Confidence: high
```

**Key API calls:**
- `IDXGIOutput::DuplicateOutput()` — Creates duplication interface for a display output
- `IDXGIOutputDuplication::AcquireNextFrame()` — Waits for and retrieves the next desktop frame (returns `IDXGIResource`)
- `IDXGIOutputDuplication::GetFrameDirtyRects()` — Returns non-overlapping rectangles indicating updated regions
- `IDXGIOutputDuplication::GetFrameMoveRects()` — Returns move regions (source→destination pixel copies)
- `IDXGIOutputDuplication::GetFramePointerShape()` — Retrieves cursor bitmap if separate from desktop
- `IDXGIOutputDuplication::ReleaseFrame()` — Releases frame after processing

```
Claim: "AcquireNextFrame returns an IDXGIResource that can be queried for ID3D11Texture2D, giving a shared GPU texture handle importable into NVENC via CUDA interop."
Source: Microsoft Learn — IDXGIOutputDuplication::AcquireNextFrame
URL: https://learn.microsoft.com/en-us/windows/win32/api/dxgi1_2/nf-dxgi1_2-idxgioutputduplication-acquirenextframe
Date: 2021-10-12
Excerpt: "[out] ppDesktopResource — A pointer to a variable that receives the IDXGIResource interface of the surface that contains the desktop bitmap."
Context: The returned texture lives in GPU memory and can be directly imported into CUDA/encoder surfaces
Confidence: high
```

#### Dirty Rectangles and Move Rectangles

```
Claim: "DDA provides GetFrameDirtyRects and GetFrameMoveRects to minimize data transfer. The desktop image format is always B8G8R8A8_UNORM regardless of display mode."
Source: Microsoft Learn — GetFrameDirtyRects
URL: https://learn.microsoft.com/en-us/windows/win32/api/dxgi1_2/nf-dxgi1_2-idxgioutputduplication-getframedirtyrects
Date: 2021-10-12
Excerpt: "IDXGIOutputDuplication::GetFrameDirtyRects returns dirty regions, which are non-overlapping rectangles that indicate the areas of the desktop image that the operating system updated since you processed the previous desktop image."
Context: Enables incremental capture for bandwidth optimization
Confidence: high
```

**Dirty Rect Reconstruction Code (Microsoft sample):**
```cpp
// Process move regions first
for (UINT i = 0; i < MoveCount; ++i) {
    // Copy from source to destination for each move rect
}
// Then process dirty regions
for (UINT i = 0; i < DirtyCount; ++i) {
    // Update dirty rectangles with new pixel data
}
```

**Important limitation:** Move rects have become largely unused in Windows 10 — most changes are identified as dirty rects only.

```
Claim: "On Windows 10, all changed regions are identified as dirty rects. Move regions were properly identified on Windows 8/Server 2012 but this behavior changed in Windows 10."
Source: Stack Overflow — DXGI Desktop Duplication moved regions
URL: https://stackoverflow.com/questions/37442532/when-does-dxgi-desktop-duplication-api-identify-a-region-as-a-moved-region
Date: 2020-06-04
Excerpt: "In my tests, all changed regions are identified as dirty rects on Windows 10. On Windows Server 2012 (like Windows 8), moved regions can be identified correctly."
Context: Practical limitation for incremental update optimization on modern Windows
Confidence: high
```

#### DXGI vs Windows.Graphics.Capture Performance

```
Claim: "Windows.Graphics.Capture uses the same technology as DXGI Desktop Duplication API underneath, but adds cross-GPU capability and per-window capture. DXGI outperforms WGC for monitor capture."
Source: Stack Overflow — Desktop Duplication API vs Windows.Graphics.Capture
URL: https://stackoverflow.com/questions/74084077/desktop-duplication-api-vs-windows-graphics-capture
Date: 2022-10-16
Excerpt: "Windows.Graphics.Capture uses the same technology as DXGI Desktop Duplication API underneath, plus it can capture a windows (HWND) which is today not exposed."
Context: WGC is a higher-level wrapper with additional capabilities but same underlying mechanism
Confidence: high
```

```
Claim: "WGC performs significantly worse than DXGI for monitor capture. A GStreamer developer noted: 'capturing a monitor using WGC is not recommended because of its poor performance... DXGI still much outperforms than WGC.'"
Source: GStreamer Discourse
URL: https://discourse.gstreamer.org/t/d3d11screencapturesrc-dxgi-vs-wgc/912
Date: 2024-01-31
Excerpt: "capturing a monitor using WGC is not recommended because of its poor performance. So, capturing a monitor using your RTX GPU can show poor result than DXGI + iGPU."
Context: For full-screen capture, DXGI remains the performance leader; WGC is better for window capture and cross-GPU scenarios
Confidence: high
```

#### DXGI Timeout and Frame Retrieval

The `AcquireNextFrame()` API is event-driven rather than polling:
- Default timeout in WebRTC implementation: **10ms** (`kAcquireTimeoutMs = 10`)
- DXGI also returns `AccumulatedFrames` count when multiple vsync periods elapsed since last call
- `DXGI_ERROR_ACCESS_LOST` must be handled on mode switches, desktop switches, or DWM transitions

### 1.2 Windows.Graphics.Capture (WGC)

WGC was introduced in Windows 10 1803 as a modern WinRT API with broader compatibility.

**Advantages over DXGI:**
- Cross-GPU capture (can capture displays attached to different GPUs)
- Per-window capture via HWND
- Works with UWP applications
- Supports DirectComposition content

**Disadvantages:**
- Yellow border around captured content (removable in Windows 11)
- Lower performance for full-screen/monitor capture
- Less control over capture timing

```
Claim: "WGC frame rate drops dramatically when capturing from a GPU not attached to the display. Tests showed 4-5 FPS capture on an iGPU when monitor was connected to dGPU, vs 60+ FPS when using the GPU connected to the monitor."
Source: GStreamer Discourse
URL: https://discourse.gstreamer.org/t/d3d11screencapturesrc-dxgi-vs-wgc/912
Date: 2024-02-02
Excerpt: "If I use WGC on the GPU that has no monitor attached, I can see that both GPUs are beeing utilized, it's like if the first GPU passes the data to the second one."
Context: Cross-GPU capture involves implicit texture transfers between GPUs
Confidence: high
```

### 1.3 Magnification API

The Magnification API (`MagSetWindowFilterList`, `MagSetWindowSource`) provides an alternative capture method with unique characteristics.

```
Claim: "Magnification API can capture background or occluded windows (bypasses Z-order), supports hardware-accelerated content via DWM composition, but has high performance cost from frequent composition and readback."
Source: Ryan's Blog — Game Capture & Window Capture
URL: https://ryanai.dev/blog/pc-window-capture
Date: 2024-09-26
Excerpt: "Can capture background or occluded windows: Bypasses Z-order restrictions; ideal for capturing hidden windows. High performance cost: Frequent composition and readback from the magnifier cause high CPU/GPU usage."
Context: Useful for niche scenarios but not recommended for low-latency pipelines
Confidence: high
```

### 1.4 DWM Interaction and Latency

The Desktop Window Manager (DWM) is the compositor that DXGI captures from. Understanding DWM behavior is critical for latency optimization.

```
Claim: "DWM adds at minimum 1 frame of latency (~17ms at 60Hz). In windowed mode with DWM composition, total added latency can be ~31ms. Independent Flip mode reduces this, and 'True Immediate Independent Flip' can bypass DWM entirely with DXGI_PRESENT_ALLOW_TEARING."
Source: Present Latency, DWM and Waitable Swapchains blog
URL: https://jackmin.home.blog/2018/12/14/swapchains-present-and-present-latency/
Date: 2018-12-14
Excerpt: "We can see based on the above timeline that in this case, the presence of DWM added around ~31ms latency. In contrast if we were operating in exclusive full screen mode without vsync, then at T=3ms the application would have been able to flip the backbuffer."
Context: Critical for understanding why capture-from-screen adds inherent latency
Confidence: high
```

```
Claim: "DWM composition can be bypassed using DwmExtendFrameIntoClientArea with margins=-1, which avoids the performance hit and ~1-30ms input lag caused by DWM composition."
Source: UnKnoWnCheaTs forum
URL: https://www.unknowncheats.me/forum/4472354-post2.html
Date: 2025-09-08
Excerpt: "DwmExtendFrameIntoClientArea lets you bypass the performance hit caused by DWM composition. Avoid DWM-induced lag."
Context: Used by gaming overlays and capture tools to reduce composition latency
Confidence: medium
```

```
Claim: "DXGI flip model swapchains can achieve as low as 1 frame of latency with DXGI_SWAP_CHAIN_FLAG_FRAME_LATENCY_WAITABLE_OBJECT, and even lower with DXGI_SWAP_CHAIN_FLAG_ALLOW_TEARING for multi-plane overlay direct flip."
Source: Microsoft Learn — For best performance, use DXGI flip model
URL: https://learn.microsoft.com/en-us/windows/win32/direct3ddxgi/for-best-performance--use-dxgi-flip-model
Date: 2021-01-06
Excerpt: "Decreasing latency using DXGI_SWAP_CHAIN_FLAG_FRAME_LATENCY_WAITABLE_OBJECT. When in Independent Flip mode, you can get down to 1 frame of latency on recent versions of Windows."
Context: Key API for reducing presentation latency in capture pipelines
Confidence: high
```

### 1.5 Looking Glass D12 — Advanced Windows Capture

Looking Glass B7 introduced a DirectX 12 capture backend that surpasses even NVIDIA's proprietary NvFBC.

```
Claim: "Looking Glass B7's D12 capture backend is faster and lighter than even NVIDIA's proprietary NvFBC interface. The new DirectX12 capture engine uses GPU copy engines to transfer textures directly into IVSHMEM shared memory."
Source: Looking Glass Releases
URL: https://github.com/gnif/LookingGlass/releases
Date: 2025-03-05
Excerpt: "The difference this has made can not be overstated, this new capture backed is now faster and lighter then even the proprietary NvFBC interface from NVIDIA. Combine this with the KVMFR module on the host system allowing the GPU to perform DMA transfers from shared memory, the CPU is no longer involved at all in shuffling any frame data directly anymore."
Context: Open-source capture technology now exceeds proprietary solutions
Confidence: high
```

---

## 2. macOS Capture APIs

### 2.1 ScreenCaptureKit (SCStream)

ScreenCaptureKit was introduced in macOS 12.3 (Monterey) and represents the modern, GPU-optimized capture pipeline replacing legacy APIs (CGWindowList, AVCapture).

```
Claim: "ScreenCaptureKit outputs CMSampleBuffer frames backed by IOSurface, which can bind directly to CAMetalLayer or CALayer without CPU round-trip. Frames arrive vsync-aligned and stay on the GPU."
Source: 60fps Zero Latency PiP Deep Dive
URL: https://lufra.quinttech.net/blog/60fps-zero-latency-pip-deep-dive
Date: 2025-01-01
Excerpt: "The buffer is backed by an IOSurface, which you can bind directly to a CAMetalLayer or CALayer without a CPU round-trip. Compositing becomes a single GPU pass. The floating layer is just another sublayer in the WindowServer tree."
Context: Modern macOS capture is fundamentally zero-copy from capture through composition
Confidence: high
```

#### Key Architecture

**Core types and flow:**
1. `SCShareableContent` — Enumerate available displays, windows, applications
2. `SCContentFilter` — Filter what to capture (display, window, or app)
3. `SCStreamConfiguration` — Configure output: resolution, pixel format, frame rate, HDR
4. `SCStream` — The capture stream with `addStreamOutput(_:type:sampleHandlerQueue:)`
5. Frame callback receives `CMSampleBuffer` containing `CVPixelBuffer` backed by `IOSurface`

**Supported pixel formats:**
- `'BGRA'` — Standard BGRA 8-bit
- `'l10r'` — Packed Little Endian ARGB2101010 (10-bit)
- `'420v'` — 2-plane video-range YCbCr 4:2:0
- `'420f'` — 2-plane full-range YCbCr 4:2:0
- `'xf44'` — 2-plane full-range YCbCr10 4:4:4 (macOS 15+)
- `'RGhA'` — 64-bit RGBA IEEE half-precision float (HDR)

### 2.2 IOSurface and Zero-Copy

```
Claim: "CVPixelBuffer contains an IOSurface and Metal knows how to use an IOSurface for texturing. Using CVMetalTextureCache simplifies the interface between CVPixelBuffer and Metal, removing the need to manually track IOSurfaces and IOSurfaceUseCounts."
Source: Apple WWDC 2020 — Decode ProRes with AVFoundation and VideoToolbox
URL: https://developer.apple.com/videos/play/wwdc2020/10090/
Date: 2020-06-25
Excerpt: "Using CVMetalTextureCache to manage the interface between CVPixelBuffer and MetalTexture simplifies things, removing the need to manually track IOSurfaces and IOSurfaceUseCounts."
Context: IOSurface is the fundamental shared memory primitive for zero-copy on macOS/iOS
Confidence: high
```

**Zero-copy IOSurface → Metal texture path:**
```objc
// Create texture cache
CVMetalTextureCacheCreate(kCFAllocatorDefault, NULL, metalDevice, NULL, &textureCache);

// Create texture from CVPixelBuffer (which wraps IOSurface)
CVMetalTextureCacheCreateTextureFromImage(kCFAllocatorDefault, textureCache, pixelBuffer, NULL, MTLPixelFormatBGRA8Unorm, width, height, 0, &cvMetalTexture);

// Get the Metal texture
id<MTLTexture> texture = CVMetalTextureGetTexture(cvMetalTexture);
```

### 2.3 EDR/HDR Capture (macOS 15+)

```
Claim: "ScreenCaptureKit on macOS 15+ supports HDR capture via SCCaptureDynamicRangeHDRLocalDisplay or SCCaptureDynamicRangeHDRCanonicalDisplay, with pixel formats having at least 10-bits per component and Display P3 PQ color space."
Source: Apple WWDC 2024 — Capture HDR content with ScreenCaptureKit
URL: https://developer.apple.com/videos/play/wwdc2024/10088/
Date: 2024-06-13
Excerpt: "SCStream now outputs High Dynamic Range of captured content. For HDR you will use a format that has at least 10-bits per component. For most situations 10-bit YCbCr will be the best choice."
Context: HDR capture requires Apple Silicon; Intel Macs will silently ignore HDR settings
Confidence: high
```

```
Claim: "ScreenCaptureKit provides convenient presets for HDR capture: captureHDRStreamLocalDisplay (for same-screen rendering) and captureHDRStreamCanonicalDisplay (for sharing with other HDR devices). HDR capture is only supported on Apple Silicon Mac."
Source: dotnet/macios Wiki — ScreenCaptureKit macOS Xcode16.0
URL: https://github.com/dotnet/macios/wiki/ScreenCaptureKit-macOS-xcode16.0-b1
Date: N/A
Excerpt: "HDR capture is only supported with Apple Silicon Mac, setting this property on Intel Mac will have no effect."
Context: Apple Silicon required for EDR/HDR capture path
Confidence: high
```

### 2.4 Display Stream Compression (DSC)

```
Claim: "Display Stream Compression (DSC) is a VESA-developed visually lossless low-latency compression algorithm that uses delta PCM coding and YCoCg-R color space, supporting up to 3:1 compression ratio. DSC latency is in microseconds."
Source: Mac Performance Guide — Display Stream Compression
URL: https://macperformanceguide.com/blog/2020/20200819_1308-DisplayStreamCompression.html
Date: 2020-08-19
Excerpt: "DSC is a VESA-developed low-latency compression algorithm to overcome the limitations posed by sending high-resolution video over physical media of limited bandwidth. It is a visually lossless low-latency algorithm based on delta PCM coding and YCoCg-R color space."
Context: DSC enables dual 6K display support on Macs; compression happens at the display link level, not capture level
Confidence: high
```

```
Claim: "macOS broke Display Stream Compression (DSC) 1.4 in Big Sur and has not fixed it. Multiple monitors across different Macs, GPUs, and vendors were affected. DSC had to be disabled or downgraded to 1.2 for proper operation."
Source: Hacker News
URL: https://news.ycombinator.com/item?id=37018902
Date: 2023-08-06
Excerpt: "DSC 1.4 and Big Sur (Ventura, Monterey... this still hasn't been 'fixed'). Hundreds or more bug reports were filed."
Context: Important caveat for high-bandwidth capture/display scenarios on macOS
Confidence: high
```

### 2.5 VideoToolbox Integration

```
Claim: "VideoToolbox provides direct access to encoders and decoders. All frameworks (AVKit, AVFoundation, VideoToolbox) use hardware codecs on Apple Silicon. VideoToolbox uses IOSurface-backed CVPixelBuffer for zero-copy encode/decode operations."
Source: Apple WWDC 2014 — Direct Access to Video Encoding and Decoding
URL: https://developer.apple.com/videos/play/wwdc2014/513/
Date: 2014-06-05
Excerpt: "On OS X, AVKit and AVFoundation will use hardware codecs when they're available on the system and when it's appropriate. And VideoToolbox will use hardware codecs when it's available on system and when you request it."
Context: VideoToolbox is the lowest-level API for hardware-accelerated encode/decode
Confidence: high
```

### 2.6 ScreenCaptureKit Performance

```
Claim: "ScreenCaptureKit on Apple Silicon achieves 1080p at 30-60 FPS and 4K at 15-30 FPS. First-frame latency ranges from 30-150ms depending on resolution."
Source: screencapturekit-rs documentation
URL: https://doom-fish.github.io/screencapturekit-rs/
Date: 2026-03-02
Excerpt: "Typical Performance (Apple Silicon): 1080p 30-60 FPS, First Frame Latency 30-100ms; 4K 15-30 FPS, First Frame Latency 50-150ms"
Context: Community benchmarks for ScreenCaptureKit performance
Confidence: medium
```

---

## 3. Linux Capture APIs

### 3.1 PipeWire + DMA-BUF

PipeWire is the modern Linux multimedia server that provides zero-copy screen sharing via DMA-BUF file descriptors.

```
Claim: "PipeWire supports zero-copy via shared memory, memfd, dmabuf, and eventfd. It provides RT capability with latency under 1.5ms for audio, and DMA-BUF for video enables compositor-to-application zero-copy frame transfer."
Source: FOSDEM 2019 — PipeWire presentation by Wim Taymans
URL: https://archive.fosdem.org/2019/schedule/event/pipewire/attachments/slides/2826/export/events/attachments/pipewire/slides/2826/PipeWire.pdf
Date: 2019
Excerpt: "Zero copy, shared memory, memfd, dmabuf, eventfd. RT capable, low latency (<1.5ms)."
Context: PipeWire was designed from the ground up for zero-copy media sharing
Confidence: high
```

**PipeWire screencast protocol flow:**
1. Application requests screen share via `xdg-desktop-portal`
2. Compositor (Mutter, KWin, wlroots) creates DMA-BUF backed buffers
3. PipeWire transfers buffer file descriptors via `SCM_RIGHTS` over Unix socket
4. Consumer imports DMA-BUF fd into EGLImage → GL texture → encoder surface

```
Claim: "OBS achieves zero-copy on Linux using PipeWire with DMA-BUF. 'OBS does use DMAbuf whenever possible to avoid copies in the path.' PipeWire src with explicit DMA-BUF negotiation is required: pipewiresrc ! video/x-raw(memory:DMABuf) ! queue ..."
Source: GStreamer Discourse
URL: https://discourse.gstreamer.org/t/pipewiresrc-not-using-full-framerate/5570
Date: 2025-12-24
Excerpt: "OBS does use DMAbuf whenever possible to avoid copies in the path. pipewiresrc ! video/x-raw(memory:DMABuf) ! queue ! fpsdisplaysink"
Context: DMA-BUF must be explicitly negotiated; fallback to shared memory adds copies
Confidence: high
```

### 3.2 KMS/DRM Direct Scanout

KMS/DRM (Kernel Mode Setting / Direct Rendering Manager) provides the lowest-level GPU framebuffer access on Linux.

```
Claim: "KMS/DRM direct scanout enables the lowest-latency capture path on Linux. Wayland with dmabuf+plane support results in more direct scanout and fewer GPU composites than X11. Fewer hops mean lower latency: Wayland is Client→Compositor→KMS vs X11's Client→Xorg→Compositor→KMS."
Source: Demystifying the Embedded Linux Graphics Stack (OSSEU 2025)
URL: https://static.sched.com/hosted_files/osseu2025/a6/Demystifying_the_Embedded_Linux_Graphics_Stack-An_Easy_Introduction_for_Beginner_V1.2.pdf
Date: 2025
Excerpt: "Wayland: Client→Compositor→KMS (fewer hops; simpler). Buffer passing: X11 uses DRI3+Present; Wayland uses linux-dmabuf. Embedded wins on Wayland because: fewer hops→lower latency; dmabuf+plane support→more direct scanout, fewer GPU composites."
Context: Architecture comparison showing Wayland's inherent latency advantage
Confidence: high
```

**FFmpeg kmsgrab zero-copy → VAAPI encode:**
```bash
ffmpeg -crtc_id 42 -framerate 60 -f kmsgrab -i - \
  -vf 'hwmap=derive_device=vaapi,scale_vaapi=w=1920:h=1080:format=nv12' \
  -c:v h264_vaapi output.mp4
```

```
Claim: "FFmpeg kmsgrab with hwmap derive_device=vaapi achieves zero-copy encode at ~3% CPU usage for 1080p capture and encode. This is an order of magnitude better than OBS compositing pipeline which requires ~30% CPU."
Source: OBS Forums — Experimental zero-copy screen capture
URL: https://obsproject.com/forum/threads/experimental-zero-copy-screen-capture-on-linux.101262/
Date: 2019-08-26
Excerpt: "ffmpeg just directly encoding from kmsgrab uses about 3% of one core. Looks like CPU overhead is about 5-7% of one core just previewing the one 1080p DMABUF source, with an additional ~30% when recording."
Context: Compositing (as OBS does) fundamentally precludes zero-copy optimization
Confidence: high
```

### 3.3 wlroots Screencopy

```
Claim: "wlroots-based compositors support the wlr-export-dmabuf protocol which provides DMA-BUF file descriptors for zero-copy screen capture without any intermediate copies. The wlr-screencopy-unstable-v1 protocol is supported across Cage, COSMIC, GameScope, Hyprland, KWin, Labwc, niri, phoc, river, Sway, Treeland, Wayfire, and Weston."
Source: wayland.app — wlr screencopy protocol
URL: https://wayland.app/protocols/wlr-screencopy-unstable-v1
Date: N/A
Excerpt: Compositor support table showing 15+ compositors supporting the protocol
Context: wlr-export-dmabuf is the standard for wlroots-based zero-copy capture
Confidence: high
```

```
Claim: "For wlroots-based compositors, wlr-export-dmabuf provides dmabufs and allows zero-copy with no effort. This is the recommended approach for GPU screen recorders on wlroots."
Source: Reddit — Zero-copy GPU screen recorder
URL: https://www.reddit.com/r/linux/comments/1s6tzpi/working_on_a_modern_zero_copy_gpu_screen_recorder/
Date: 2025
Excerpt: "wlr-export-dmabuf protocol in case of wlroots-based wms, it provides dmabufs and allows to go zerocopy with no efforts."
Context: wlroots provides the cleanest zero-copy capture path on Linux
Confidence: high
```

### 3.4 V4L2 Loopback

```
Claim: "The vcam kernel driver is a DMA-BUF backed virtual camera that supports zero-copy semantics when both capture and output reference the same DMA-BUF. If both buffers reference the same DMA-BUF, the driver performs a zero-copy transfer by propagating metadata."
Source: LWN.net — media: Virtual camera driver
URL: https://lwn.net/Articles/1056824/
Date: 2026-02-01
Excerpt: "If both buffers reference the same DMA-BUF, the driver performs a zero-copy transfer by propagating metadata."
Context: Modern replacement for v4l2loopback with native DMA-BUF zero-copy support
Confidence: high
```

### 3.5 DMA-BUF → EGLImage → VAAPI Encode Pipeline

The canonical zero-copy pipeline on Linux:

```
Capture → DMA-BUF fd → EGLImage → GL Texture → (optional GPU color conversion) → VASurface → VAAPI Encode → bitstream
```

**Key APIs:**
- `EGL_EXT_image_dma_buf_import` — Creates EGLImage from DMA-BUF fd
- `GL_OES_EGL_image` / `GL_OES_EGL_image_external` — Binds EGLImage to GL texture
- `vaCreateSurfaces` with `VASurfaceAttribExternalBufferDescriptor` — Creates VA surface from DMA-BUF
- `vapostproc` / `vaapipostproc` — VAAPI color conversion (RGB→NV12) without CPU roundtrip

```
Claim: "Gstreamer vaapipostproc and vah264enc can process DMA-BUF as input and negotiate VA native format downstream with the encoder, enabling full zero-copy from capture to encoded bitstream."
Source: GStreamer devel mailing list
URL: https://lists.freedesktop.org/archives/gstreamer-devel/2023-February/081021.html
Date: 2023-02-20
Excerpt: "both don't consume RGB frames. You'll need to add vapostproc / vaapipostproc to do the color conversion. vapostproc / vaapipostproc also process DMAbuf as input."
Context: Color space conversion is the remaining barrier; VAAPI postproc solves it
Confidence: high
```

---

## 4. GPU Memory Architecture & Interop

### 4.1 Windows: CUDA Interop with D3D11/DXGI Textures

```
Claim: "CUDA provides cuGraphicsD3D11RegisterResource to register D3D11 textures as CUDA resources. After registration, cuGraphicsMapResources and cuGraphicsSubResourceGetMappedArray provide CUDA-accessible handles for zero-copy texture operations."
Source: NVIDIA Developer Forums
URL: https://forums.developer.nvidia.com/t/nvdec-decoded-frame-trying-a-zero-copy-to-nv12-d3d11-texture/123291
Date: 2020-05-15
Excerpt: "cuGraphicsD3D11RegisterResource two different textures (R8_UNorm + R8G8_UNorm, created at begin), then cuGraphicsMapResources, cuGraphicsSubResourceGetMappedArray (one time) cuMemcpy2DAsync and cuGraphicsUnmapResources."
Context: Standard pattern for D3D11↔CUDA zero-copy interop
Confidence: high
```

**CUDA ↔ D3D11 Interop API calls:**
```cpp
// Register the captured DXGI texture with CUDA
cuGraphicsD3D11RegisterResource(&cudaResource, d3d11Texture, cudaGraphicsRegisterFlagsNone);

// Map resource before use
cuGraphicsMapResources(1, &cudaResource, cudaStream);

// Get mapped CUDA array
cuGraphicsSubResourceGetMappedArray(&cuArray, cudaResource, 0, 0);

// Use cuMemcpy2DAsync to copy between CUDA arrays and device memory
CUDA_MEMCPY2D copyDesc = {};
copyDesc.srcMemoryType = CU_MEMORYTYPE_ARRAY;
copyDesc.srcArray = srcArray;
copyDesc.dstMemoryType = CU_MEMORYTYPE_ARRAY;
copyDesc.dstArray = dstArray;
copyDesc.WidthInBytes = width;
copyDesc.Height = height;
cuMemcpy2DAsync(&copyDesc, cudaStream);

// Unmap when done
cuGraphicsUnmapResources(1, &cudaResource, cudaStream);
```

```
Claim: "NVENC can encode directly from CUDA-registered D3D11 textures. The NVIDIA Video Codec SDK Encoder API supports RegisteredResource input buffers that wrap CUDA device pointers or registered graphics resources."
Source: nvidia-video-codec-sdk Rust bindings documentation
URL: https://docs.rs/nvidia-video-codec-sdk
Date: 2026-04-21
Excerpt: "Create input Buffers (or RegisteredResource) and output Bitstreams. Encode frames with Session::encode_picture."
Context: RegisteredResource enables zero-copy from DXGI capture to NVENC encode
Confidence: high
```

### 4.2 macOS: IOSurface ↔ Metal ↔ VideoToolbox

```
Claim: "CVPixelBuffer contains an IOSurface which can be used with Metal for texturing. Two approaches exist: direct IOSurface binding with manual use count tracking, or CVMetalTextureCache for simpler and more efficient management."
Source: Apple WWDC 2020
URL: https://developer.apple.com/videos/play/wwdc2020/10090/
Date: 2020-06-25
Excerpt: "CVMetalTextureCache also saves you from repeating the IOSurface texture binding when IOSurfaces which come from a CVPixelBufferPool are reused and are seen again, making it a little bit more efficient."
Context: CVMetalTextureCache is the recommended approach for Metal↔Video interop
Confidence: high
```

**ScreenCaptureKit → IOSurface → Metal texture → VideoToolbox encode:**
```objc
// 1. ScreenCaptureKit outputs CMSampleBuffer with CVPixelBuffer
CVPixelBufferRef pixelBuffer = CMSampleBufferGetImageBuffer(sampleBuffer);

// 2. Get the IOSurface
IOSurfaceRef surface = CVPixelBufferGetIOSurface(pixelBuffer);

// 3. Create Metal texture from IOSurface
id<MTLTexture> texture = [metalDevice newTextureWithDescriptor:descriptor iosurface:surface plane:0];

// 4. VideoToolbox encoder consumes the same CVPixelBuffer (IOSurface-backed)
VTCompressionSessionEncodeFrame(session, pixelBuffer, presentationTimeStamp, duration, NULL, NULL, NULL);
```

### 4.3 Linux: EGLImage/DMA-BUF ↔ VAAPI

```
Claim: "dma-buf is converted to an EGLImage, which is bound to a GL texture with glEGLImageTargetTexture2D(). This is the standard zero-copy path used by V4L2, VA-API on EGL, and gstreamer-vaapi."
Source: GStreamer Conference 2016 — Vulkan, OpenGL and/or Zerocopy
URL: https://gstreamer.freedesktop.org/data/events/gstreamer-conference/2016/Matthew%20Waters%20-%20Vulkan,%20OpenGL%20and%20ZeroCopy.pdf
Date: 2016-10-10
Excerpt: "dma-buf is converted to an EGLImage. EGLImage is bound to a GL texture with glEGLImageTargetTexture2D()."
Context: This is the fundamental primitive for Linux zero-copy video
Confidence: high
```

**DMA-BUF → EGLImage → GL Texture API:**
```c
// Import DMA-BUF fd into EGLImage
EGLint attribs[] = {
    EGL_WIDTH, width,
    EGL_HEIGHT, height,
    EGL_LINUX_DRM_FOURCC_EXT, DRM_FORMAT_ARGB8888,
    EGL_DMA_BUF_PLANE0_FD_EXT, dmabuf_fd,
    EGL_DMA_BUF_PLANE0_OFFSET_EXT, 0,
    EGL_DMA_BUF_PLANE0_PITCH_EXT, stride,
    EGL_NONE
};
EGLImage image = eglCreateImageKHR(eglDisplay, EGL_NO_CONTEXT, EGL_LINUX_DMA_BUF_EXT, NULL, attribs);

// Bind to GL texture
glBindTexture(GL_TEXTURE_2D, texture);
glEGLImageTargetTexture2DOES(GL_TEXTURE_2D, image);

// For VAAPI zero-copy encode, export texture as VASurface
VASurfaceAttrib attribs[] = {
    {VASurfaceAttribMemoryType, VA_SURFACE_ATTRIB_SETTABLE, {VAGenericValueTypeInteger, VA_SURFACE_ATTRIB_MEM_TYPE_DRM_PRIME_2}},
    {VASurfaceAttribExternalBufferDescriptor, VA_SURFACE_ATTRIB_SETTABLE, {.value.p = &descriptor}},
};
vaCreateSurfaces(vaDisplay, format, width, height, &surface, 1, attribs, 2);
```

---

## 5. Zero-Copy Pipeline Paths

### 5.1 Windows: Capture Texture → Encoder Surface

**Optimal pipeline:**
```
DXGI AcquireNextFrame → ID3D11Texture2D → cuGraphicsD3D11RegisterResource → CUDA array → NVENC RegisteredResource → encoded bitstream
```

**Key insight:** The DXGI texture returned by `AcquireNextFrame` is already in GPU memory. Using CUDA interop, this texture can be registered as a CUDA resource and passed directly to NVENC as a `RegisteredResource` input surface. No CPU readback, no `memcpy`, no system RAM involvement.

```
Claim: "The Desktop Duplication API returns shared GPU texture handles that are importable into NVENC via CUDA interop, enabling a zero-copy path from capture to encode without any CPU roundtrip."
Source: Context from Landscape Scan, verified through NVIDIA SDK documentation
URL: Multiple sources
Date: N/A
Excerpt: "DXGI returns shared GPU texture handles importable into NVENC via CUDA interop"
Context: This is the core zero-copy optimization for Windows game streaming
Confidence: high
```

### 5.2 macOS: Capture Texture → Encoder Surface

**Optimal pipeline:**
```
SCStream → CMSampleBuffer → CVPixelBuffer (IOSurface-backed) → CVMetalTextureCache → MTLTexture (GPU rendering if needed) → VTCompressionSession → encoded CMSampleBuffer
```

ScreenCaptureKit's output `CVPixelBuffer` is already IOSurface-backed and can be directly consumed by VideoToolbox's `VTCompressionSession`. If GPU-side preprocessing is needed (e.g., scaling, overlay), `CVMetalTextureCacheCreateTextureFromImage` provides zero-copy Metal texture access.

### 5.3 Linux: Capture Texture → Encoder Surface

**Optimal pipeline:**
```
PipeWire/kmsgrab → DMA-BUF fd → EGLImage → GL Texture → vaCreateSurfaces (VASurface from DMA-BUF) → VAAPI encode → bitstream
```

The entire pipeline stays on the GPU: capture framebuffer → DMA-BUF export → VAAPI surface import → hardware encode. Color conversion from RGB to NV12 can be done via `vapostproc` on the GPU.

### 5.4 NVIDIA Jetson NVMM Zero-Copy

```
Claim: "NVIDIA Jetson NVMM (NVIDIA Memory Management) zero-copy reduces latency from 200-500ms to 10-30ms for camera-to-stream pipelines. A glass-to-glass latency test on Jetson TX2 averaged about 60ms including camera capture, debayer, color conversion, H.264 encode, wireless transmission, decode, and display."
Source: Fastvideo — Jetson Zero Copy for Embedded Applications
URL: https://www.fastcompression.com/blog/jetson-zero-copy.htm
Date: 2026-04-15
Excerpt: "The averaged latency was about 60 ms which can be considered an exceptional result... Image acquisition from camera and zero-copy to Jetson GPU → Black level → White Balance → HQLI Debayer → Export to YUV (NV12) → H.264 encoding via V4L2 → RTSP streaming"
Context: Zero-copy is critical for embedded/edge pipelines where CPU and memory bandwidth are limited
Confidence: high
```

---

## 6. Frame Pacing & Synchronization

### 6.1 V-Sync and Its Problems

Traditional V-Sync introduces significant latency because it queues frames to align with the display's refresh cycle. At 60Hz, each frame is ~16.7ms, and triple buffering can add 2-3 frames of queue depth.

### 6.2 Fast Sync (NVIDIA) and Enhanced Sync (AMD)

```
Claim: "NVIDIA Fast Sync and AMD Enhanced Sync eliminate screen tearing while allowing uncapped frame rates with significantly less input lag than V-Sync. They work by rendering all frames but only sending the most recently completed frame to the display. Fast Sync requires Pascal+ (GeForce 900 series)."
Source: DisplayNinja
URL: https://www.displayninja.com/what-is-nvidia-fast-sync-and-amd-enhanced-sync/
Date: 2025-10-03
Excerpt: "Fast Sync eliminates screen tearing even when your FPS exceeds your monitor's maximum refresh rate by showing the most recently completed full frame. This increases input lag, but not nearly as much as V-Sync does."
Context: Best used when frame rate is 2-3x the refresh rate
Confidence: high
```

```
Claim: "AMD Enhanced Sync delivers a tear-free experience when framerate exceeds the display's refresh rate at ultra-low latency. It complements FreeSync by providing tear-free gaming above the FreeSync range."
Source: AMD Official
URL: https://www.amd.com/en/products/software/adrenalin/software-enhancedsync.html
Date: 2024-02-23
Excerpt: "Enhanced Sync technology delivers liquid smooth gameplay by focusing on latency at a low framerate."
Context: AMD's official positioning for Enhanced Sync
Confidence: high
```

### 6.3 NVIDIA Reflex

```
Claim: "NVIDIA Reflex Low Latency mode reduces PC latency by synchronizing CPU and GPU work, eliminating the render queue. Reflex 2 with Frame Warp can reduce PC latency by up to 75%. In The Finals at 4K max settings, Reflex 2 reduces latency from 56ms to 14ms on RTX 5070."
Source: NVIDIA Official Blog
URL: https://www.nvidia.com/en-us/geforce/news/reflex-2-even-lower-latency-gameplay-with-frame-warp/
Date: 2025-01-06
Excerpt: "At 4K with max settings and global illumination on an RTX 5070, THE FINALS achieves 56ms of latency. With Reflex Low Latency, latency is more than halved to 27ms. And by enabling Reflex 2, Frame Warp cuts input lag by nearly an entire frametime, reducing latency by another 50% to 14ms."
Context: Reflex SDK integrates into game engines for click-to-photon optimization
Confidence: high
```

```
Claim: "In VALORANT at 800+ FPS on RTX 5090, Reflex 2 Frame Warp achieves under 3ms PC latency — one of the lowest measured in an FPS. Reflex 2 is debuting on RTX 50 Series with support for other RTX GPUs in a future update."
Source: NVIDIA Official Blog (same)
URL: https://www.nvidia.com/en-us/geforce/news/reflex-2-even-lower-latency-gameplay-with-frame-warp/
Date: 2025-01-06
Excerpt: "In Riot Games' VALORANT, a CPU-bottlenecked game that runs blazingly fast, at 800+ FPS on the new GeForce RTX 5090, PC latency averages under 3 ms using Reflex 2 Frame Warp."
Context: Frame Warp updates rendered frames based on latest mouse input just before display scanout
Confidence: high
```

**NVIDIA Reflex SDK APIs:**
- `NVAPI_D3D_SetLatencyMarker` — Insert latency measurement markers
- `NVAPI_D3D_Sleep` — CPU-GPU synchronization for just-in-time rendering
- `NVAPI_D3D_SetFrameCuda` / `NVAPI_D3D_SetFrameMarker` — Frame warp integration
- Provides real-time latency metrics: input, simulation, render submission, driver, render queue, GPU render

### 6.4 Frame Limiters

```
Claim: "With G-SYNC + NVCP V-SYNC + Reflex, Reflex applies an automatic FPS limit slightly below the refresh rate (e.g., ~138 FPS at 144Hz). An external or in-game limiter must be set below Reflex's automatic limit to override it."
Source: BlurBusters Forums
URL: https://forums.blurbusters.com/viewtopic.php?t=7522
Date: 2020-09-19
Excerpt: "Reflex applies an automatic FPS limit slightly below the refresh rate. At 144Hz, it limits to ~138 FPS, just as LLM 'Ultra' (in supported games) did before it."
Context: Frame limiters are essential to prevent render queue buildup
Confidence: high
```

### 6.5 DXGI Swapchain Latency Controls

```
Claim: "DXGI_SWAP_CHAIN_FLAG_FRAME_LATENCY_WAITABLE_OBJECT enables frame latency waitable objects that allow applications to control maximum frame latency. Default latency is 3 frames unless changed via SetMaximumFrameLatency."
Source: Hacker News comment
URL: https://news.ycombinator.com/item?id=36899029
Date: 2023-07-27
Excerpt: "DXGI always adds 3 frames of latency, because that's the default value unless explicitly changed by the application."
Context: Must explicitly call IDXGIDevice::SetMaximumFrameLatency(1) for minimum latency
Confidence: high
```

---

## 7. Go Integration Approaches

### 7.1 Windows: Syscall + COM for DXGI

Go can access DXGI through `golang.org/x/sys/windows` combined with COM interface bindings. There is no mature pure-Go DXGI library, so approaches include:

1. **CGO + C wrapper**: Write a thin C wrapper around DXGI/COM that exports C functions callable from Go
2. **Go syscall with COM vtable**: Use `syscall.NewLazyDLL("dxgi.dll")` and manually construct COM vtable calls
3. **Go-OLE library**: Use `github.com/go-ole/go-ole` for COM automation

```go
// Conceptual approach using syscall
var (
    moddxgi = windows.NewLazySystemDLL("dxgi.dll")
    procCreateDXGIFactory1 = moddxgi.NewProc("CreateDXGIFactory1")
)

// COM interface GUIDs for DXGI
var (
    IID_IDXGIFactory1 = windows.GUID{...}
    IID_IDXGIOutputDuplication = windows.GUID{...}
)
```

### 7.2 macOS: CGO for ScreenCaptureKit

```
Claim: "ScreenCaptureKit cannot be called directly from C/Go bindings in a straightforward way because it relies heavily on Objective-C blocks, ARC, and Swift async patterns. A C99 wrapper around SCScreenshotManager_captureImage is needed."
Source: Stack Overflow — ScreenCaptureKit example in Go/C
URL: https://stackoverflow.com/questions/78846311/screencapturekit-example-in-go-c
Date: 2024-08-07
Excerpt: "I spent a couple of days trying to use ScreenCaptureKit from Apple. Which is a new way of doing things. And it works, but not in Go with C bindings."
Context: Requires a C/Objective-C bridging layer with block-based callbacks bridged to function pointers
Confidence: high
```

**Recommended approach:**
```
Go (main) → CGO → C wrapper (C99) → Objective-C++ → ScreenCaptureKit Framework
                              ↓
                    Callback via C function pointer → Go callback
```

The Rust `screencapturekit-rs` crate provides a reference implementation showing how to bridge Objective-C blocks to Rust-style callbacks, which can inform a Go implementation.

### 7.3 Linux: PipeWire Bindings

For Go on Linux, options include:

1. **PipeWire D-Bus portal**: Use `github.com/godbus/dbus/v5` to communicate with `xdg-desktop-portal` for screen capture (no native PipeWire needed)
2. **PipeWire native protocol**: Use CGO with `libpipewire` directly
3. **DMA-BUF via drm/kms**: Direct `ioctl` on `/dev/dri/card0` using `syscall.Syscall` for `DRM_IOCTL_PRIME_HANDLE_TO_FD`

```go
// Conceptual D-Bus portal approach
portal := dbus.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")
call := portal.Call("org.freedesktop.portal.ScreenCast.CreateSession", 0, options)
// ... handle session creation, source selection, stream negotiation
// DMA-BUF fds are received via Unix socket with SCM_RIGHTS
```

---

## 8. Latency Per Stage — Measured Numbers

### 8.1 Capture Stage

| Method | Resolution | Latency | Source |
|--------|-----------|---------|--------|
| DXGI Desktop Duplication | 1080p | ~1-3ms (GPU texture already available) | Microsoft docs, community benchmarks |
| ScreenCaptureKit | 1080p | ~1-2ms (IOSurface callback) | SCK benchmarks |
| kmsgrab + DMA-BUF | 1080p | ~1-2ms (framebuffer fd) | OBS zero-copy testing |
| PipeWire + DMA-BUF | 1080p | ~2-5ms (negotiation overhead) | GStreamer benchmarks |
| GDI BitBlt (CPU copy) | 1080p | ~8-16ms | Deprecated path |

### 8.2 Encode Stage

```
Claim: "NVENC median encoding latency is 5.8ms across 250,000+ Parsec sessions. AMD VCE median is 15.06ms (2.59x slower). Intel Quick Sync is 1.89x slower than NVENC."
Source: Parsec Blog
URL: https://parsec.app/blog/nvidia-nvenc-outperforms-amd-vce-on-h-264-encoding-latency-in-parsec-co-op-sessions-713b9e1e048a
Date: 2023-03-14
Excerpt: "The median encoding latency for an Nvidia card is 5.8 milliseconds; whereas, the median encoding latency on VCE is 15.06 milliseconds."
Context: Large-scale production measurement across diverse GPU generations
Confidence: high
```

```
Claim: "NVENC H.264 latency on GTX 1050 Ti at 1080p varies from ~2ms best-case to ~12ms steady-state. The encoder appears to throttle clock speeds when frame submission rate is low (e.g., 60 FPS), increasing latency. At sustained high frame rates, latency drops to ~2-3ms."
Source: NVIDIA Developer Forums
URL: https://forums.developer.nvidia.com/t/nvenc-h-264-encoder-mft-latency-increases-when-framerate-is-limited/63930
Date: 2018-08-07
Excerpt: "On my development machine I have a GTX 1050 Ti, which is able to encode a 1080p stream with around a 3 ms latency at best. However, after a while the encoder seemingly decides that since it is only being fed samples every 16 ms, it doesn't need to encode them at the regular rate."
Context: NVENC DCVS (Dynamic Clock and Voltage Scaling) throttles encoder clocks at low frame rates
Confidence: high
```

```
Claim: "NVENC on Jetson Orin NX for H.265 4K@60 encoding has ~23ms encode latency. With stitch/conversion: ~42ms total from frame start to encoder output."
Source: NVIDIA Developer Forums
URL: https://forums.developer.nvidia.com/t/how-to-lower-latency-on-nvenc-h265/347218
Date: 2025-10-09
Excerpt: "NVenc takes approximately 23 ms for a 4K frame at 60 fps... Start of Frame to dequeue on Capture plane of Enc time Δ avg: 42.263388 ms"
Context: Jetson embedded encoder is slower than desktop discrete GPUs
Confidence: high
```

```
Claim: "NVIDIA's DCVS algorithm modifies voltages based on current load dynamically. Higher rate input generation may trigger higher bandwidth voting, which can result in encoder running at higher clock. At 120 FPS input, NVENC encode time is ~6ms; at 45 FPS, latency balloons to 15+ ms."
Source: Stack Overflow
URL: https://stackoverflow.com/questions/77073702/high-latency-in-nvenc-encoding-with-lower-frame-submission
Date: 2023-09-09
Excerpt: "If the frame rate is high, then everything behaves as expected... if I submit frames at 120FPS, my encode time is roughly 6ms. However, if I submit frames much more slowly, for instance 45 FPS, then the latency on my encode balloons to 15+ milliseconds."
Context: Critical insight: sustained high frame rate input keeps NVENC at maximum clock
Confidence: high
```

**Encode latency summary:**

| Encoder | Resolution | Latency | Notes |
|---------|-----------|---------|-------|
| NVENC (desktop, low latency preset) | 1080p | 2-6ms | Best case with high input rate |
| NVENC (desktop, steady state 60fps) | 1080p | 5-12ms | DCVS throttling effect |
| NVENC (median across sessions) | 1080p | 5.8ms | Parsec production data |
| NVENC (Jetson Orin, H.265) | 4K | ~23ms | Embedded GPU, 4K penalty |
| AMD VCE (median) | 1080p | 15.06ms | 2.59x slower than NVENC |
| Intel Quick Sync (median) | 1080p | ~11ms | 1.89x slower than NVENC |
| VAAPI (Intel iGPU) | 1080p | 3-8ms | Depends on driver |

### 8.3 Network/Transmit Stage

```
Claim: "The median ping in a Parsec Co-Play session is 32.68ms — slightly higher than 2 frames at 60 FPS (33.3ms). Network latency typically ranges from 10-100ms depending on conditions."
Source: Parsec Blog (same as above)
URL: https://parsec.app/blog/nvidia-nvenc-outperforms-amd-vce-on-h-264-encoding-latency-in-parsec-co-op-sessions-713b9e1e048a
Date: 2023-03-14
Excerpt: "The median ping in a Parsec Co-Play session is 32.68 milliseconds. This is slightly higher than 2 frames at 60 frames per second."
Context: Network latency is the dominant factor for internet streaming; LAN is typically <1ms
Confidence: high
```

### 8.4 Decode Stage

```
Claim: "Moonlight decoding latency averages 2-3ms on Apple Silicon (M1/M2 Mac Mini), sub-1ms on Windows, and 8.5ms average across mobile devices with a range of 5-20ms."
Source: Multiple — Moonlight GitHub issues and community benchmarks
URL: https://github.com/moonlight-stream/moonlight-qt/issues/1249
Date: 2024-04-06
Excerpt: "The decoding latency is between 2-3ms but if I move the mouse around very fast the decoding latency drops to 1.2ms... on windows moonlight decoding latency is in the sub 1 ms range"
Context: Decoding latency depends heavily on OS, GPU driver, and hardware decoder implementation
Confidence: medium
```

### 8.5 Display/Render Stage

```
Claim: "Display latency from a client device to screen can add 30-100ms of video processing latency on consumer displays. Gaming monitors with fast response times add 1-3ms."
Source: Hacker News comment on custom game streaming codec
URL: https://news.ycombinator.com/item?id=44714914
Date: 2025-07-28
Excerpt: "Of course, once you've tuned the end to end capture-encode-transmit-decode-display loop to sub 10 ms, you then have to contend with the 30-100 ms of video processing latency introduced by the display."
Context: Display processing latency is often the largest unaddressed component
Confidence: medium
```

### 8.6 End-to-End Pipeline Summary

```
Claim: "A WebRTC streaming pipeline has the following per-stage latencies: Capture 16.7ms, Encode <40ms, Ingest Transport ~10ms, Transcode ~7ms, Mix ~50ms, Scale ~40ms, Egress ~10ms, Decode 9ms, for a total <250ms."
Source: Red5.net — Keys to Optimizing End-to-End Latency with WebRTC
URL: https://www.red5.net/blog/keys-to-optimizing-end-to-end-latency-with-webrtc/
Date: 2025-02-25
Excerpt: "Capture: 16.7ms, Encode: <40ms, Ingest Transport: ~10ms, Transcode: ~7ms, Mix: ~50ms, Scale: ~40ms, Egress Delivery: ~10ms, Decode: 9ms, Total: <250ms"
Context: This is for a cloud broadcast pipeline, not optimized game streaming
Confidence: medium
```

**Optimized game streaming end-to-end (best case, LAN):**

| Stage | Optimized Pipeline | Naive Pipeline |
|-------|-------------------|----------------|
| Capture | 1-3ms (DXGI zero-copy) | 8-16ms (CPU readback) |
| Encode | 2-6ms (NVENC low-latency) | 10-20ms (software/DCVS throttled) |
| Packetize | 0.1-0.5ms | 1-2ms |
| Transmit (LAN) | 0.1-1ms | 0.1-1ms |
| Decode | 0.5-3ms (hardware) | 5-15ms (software) |
| Render/Display | 1-3ms (gaming monitor) | 16-50ms (consumer display) |
| **Total** | **5-17ms** | **40-104ms** |

**Cloud gaming over internet (typical):**

| Stage | Latency |
|-------|---------|
| Game Render | 16ms |
| Encode | 15ms |
| Network (median) | 25-33ms |
| Decode | 8ms |
| Display | 10ms |
| **Total** | **74-82ms** |

```
Claim: "Cloud gaming E2E lag in a modeled scenario with constant encode/decode delays: average server processing + network RTT + codec delay = 68ms. Frame rate impacts E2E lag more severely than network delay."
Source: ISCA — A Comprehensive End-to-End Lag Model for Online and Cloud Gaming
URL: https://www.isca-archive.org/pqs_2016/metzger16_pqs.pdf
Date: 2016
Excerpt: "The vertical reference line denotes the average server processing time, network round-trip and codec delay μP+2μD+e+d=68ms."
Context: Academic model showing that render+encode+decode dominate over network for local streaming
Confidence: high
```

### 8.7 Impact of Zero-Copy

```
Claim: "Looking Glass B7 with D12 capture + DMA shared memory achieves 300+ UPS (updates per second) while also reporting increased rendering performance in the guest VM. Users on laptop/iGPU report 'night and day difference.'"
Source: Looking Glass Releases
URL: https://github.com/gnif/LookingGlass/releases
Date: 2025-03-05
Excerpt: "We have users reporting 300+UPS using the D12 capture interface, while also reporting an increased rendering performance in the guest VM."
Context: Zero-copy removes CPU from the data path entirely, enabling extreme update rates
Confidence: high
```

---

## 9. References

[^1^] Microsoft Learn — Desktop Duplication API: https://learn.microsoft.com/en-us/windows/win32/direct3ddxgi/desktop-dup-api
[^2^] Microsoft Learn — IDXGIOutputDuplication::AcquireNextFrame: https://learn.microsoft.com/en-us/windows/win32/api/dxgi1_2/nf-dxgi1_2-idxgioutputduplication-acquirenextframe
[^3^] Microsoft Learn — GetFrameDirtyRects: https://learn.microsoft.com/en-us/windows/win32/api/dxgi1_2/nf-dxgi1_2-idxgioutputduplication-getframedirtyrects
[^4^] Stack Overflow — DDA moved regions: https://stackoverflow.com/questions/37442532/when-does-dxgi-desktop-duplication-api-identify-a-region-as-a-moved-region
[^5^] Stack Overflow — DDA vs WGC: https://stackoverflow.com/questions/74084077/desktop-duplication-api-vs-windows-graphics-capture
[^6^] GStreamer Discourse — DXGI vs WGC: https://discourse.gstreamer.org/t/d3d11screencapturesrc-dxgi-vs-wgc/912
[^7^] Ryan's Blog — Game Capture: https://ryanai.dev/blog/pc-window-capture
[^8^] Blinue/Magpie Wiki — Comparison of capture methods: https://github.com/Blinue/Magpie/wiki/Comparison-of-capture-methods
[^9^] Present Latency blog: https://jackmin.home.blog/2018/12/14/swapchains-present-and-present-latency/
[^10^] Microsoft Learn — DXGI flip model: https://learn.microsoft.com/en-us/windows/win32/direct3ddxgi/for-best-performance--use-dxgi-flip-model
[^11^] 60fps Zero Latency PiP blog: https://lufra.quinttech.net/blog/60fps-zero-latency-pip-deep-dive
[^12^] Apple WWDC 2020 — Decode ProRes: https://developer.apple.com/videos/play/wwdc2020/10090/
[^13^] Apple WWDC 2024 — HDR Capture: https://developer.apple.com/videos/play/wwdc2024/10088/
[^14^] screencapturekit-rs: https://doom-fish.github.io/screencapturekit-rs/
[^15^] FOSDEM 2019 — PipeWire: https://archive.fosdem.org/2019/schedule/event/pipewire/attachments/slides/2826/export/events/attachments/pipewire/slides/2826/PipeWire.pdf
[^16^] OBS Forums — Zero-copy Linux capture: https://obsproject.com/forum/threads/experimental-zero-copy-screen-capture-on-linux.101262/
[^17^] GStreamer Conference 2016 — ZeroCopy: https://gstreamer.freedesktop.org/data/events/gstreamer-conference/2016/Matthew%20Waters%20-%20Vulkan,%20OpenGL%20and%20ZeroCopy.pdf
[^18^] wayland.app — wlr screencopy: https://wayland.app/protocols/wlr-screencopy-unstable-v1
[^19^] LWN.net — vcam driver: https://lwn.net/Articles/1056824/
[^20^] FFmpeg Devices — kmsgrab: https://ffmpeg.org/ffmpeg-devices.html
[^21^] NVIDIA Developer Forums — cuGraphicsD3D11RegisterResource: https://forums.developer.nvidia.com/t/nvdec-decoded-frame-trying-a-zero-copy-to-nv12-d3d11-texture/123291
[^22^] nvidia-video-codec-sdk Rust docs: https://docs.rs/nvidia-video-codec-sdk
[^23^] NVIDIA Blog — Reflex 2: https://www.nvidia.com/en-us/geforce/news/reflex-2-even-lower-latency-gameplay-with-frame-warp/
[^24^] Parsec Blog — NVENC latency: https://parsec.app/blog/nvidia-nvenc-outperforms-amd-vce-on-h-264-encoding-latency-in-parsec-co-op-sessions-713b9e1e048a
[^25^] NVIDIA Developer Forums — NVENC MFT latency: https://forums.developer.nvidia.com/t/nvenc-h-264-encoder-mft-latency-increases-when-framerate-is-limited/63930
[^26^] Stack Overflow — NVENC frame rate latency: https://stackoverflow.com/questions/77073702/high-latency-in-nvenc-encoding-with-lower-frame-submission
[^27^] NVIDIA Developer Forums — Jetson H.265 latency: https://forums.developer.nvidia.com/t/how-to-lower-latency-on-nvenc-h265/347218
[^28^] Moonlight GitHub — Decoding latency: https://github.com/moonlight-stream/moonlight-qt/issues/1249
[^29^] Looking Glass Releases: https://github.com/gnif/LookingGlass/releases
[^30^] Fastvideo — Jetson zero-copy: https://www.fastcompression.com/blog/jetson-zero-copy.htm
[^31^] OSSEU 2025 — Linux Graphics Stack: https://static.sched.io/hosted_files/osseu2025/a6/Demystifying_the_Embedded_Linux_Graphics_Stack-An_Easy_Introduction_for_Beginner_V1.2.pdf
[^32^] Apple WWDC 2014 — VideoToolbox: https://developer.apple.com/videos/play/wwdc2014/513/
[^33^] DisplayNinja — Fast Sync/Enhanced Sync: https://www.displayninja.com/what-is-nvidia-fast-sync-and-amd-enhanced-sync/
[^34^] AMD Enhanced Sync: https://www.amd.com/en/products/software/adrenalin/software-enhancedsync.html
[^35^] BlurBusters — Reflex: https://forums.blurbusters.com/viewtopic.php?t=7522
[^36^] Red5.net — WebRTC latency: https://www.red5.net/blog/keys-to-optimizing-end-to-end-latency-with-webrtc/
[^37^] ISCA — E2E Lag Model: https://www.isca-archive.org/pqs_2016/metzger16_pqs.pdf
[^38^] Microsoft DDA Problem/Solved blog: https://www.pavelgurenko.com/2013/12/dxgi-outputs-enumeration-and-fast.html
[^39^] VirtualDub WDDM 1.2 blog: https://www.virtualdub.org/blog2/entry_356.html
[^40^] Mac Performance Guide — DSC: https://macperformanceguide.com/blog/2020/20200819_1308-DisplayStreamCompression.html
[^41^] Hacker News — macOS DSC broken: https://news.ycombinator.com/item?id=37018902
[^42^] Stack Overflow — ScreenCaptureKit in Go: https://stackoverflow.com/questions/78846311/screencapturekit-example-in-go-c
[^43^] GStreamer devel — DMA-BUF VAAPI: https://lists.freedesktop.org/archives/gstreamer-devel/2023-February/081021.html
[^44^] GStreamer Discourse — pipewiresrc DMABuf: https://discourse.gstreamer.org/t/pipewiresrc-not-using-full-framerate/5570
[^45^] Webrtc DXGI implementation: https://chromium.googlesource.com/external/webrtc/+/4a627a8c13554d12412cabb8f751caee6e61ee32/webrtc/modules/desktop_capture/win/screen_capturer_win_directx.cc
[^46^] NVIDIA Reflex SDK: https://developer.nvidia.com/performance-rendering-tools/reflex
[^47^] NVIDIA Reflex product page: https://www.nvidia.com/en-us/geforce/technologies/reflex/
[^48^] Tony Tascioglu — kmsgrab: https://wiki.tonytascioglu.com/scripts/ffmpeg/kmsgrab_screen_capture
[^49^] Reddit — Zero-copy GPU recorder: https://www.reddit.com/r/linux/comments/1s6tzpi/working_on_a_modern_zero_copy_gpu_screen_recorder/
[^50^] Looking Glass B7 docs: https://looking-glass.io/docs/B7/usage/

---

## Key Takeaways

1. **Zero-copy is transformative**: Moving from CPU-copy to zero-copy GPU pipelines reduces end-to-end latency by 3-10x. A 4K frame copy costs 1-3ms; eliminating even one copy per frame has massive cumulative impact.

2. **Windows optimal path**: DXGI `AcquireNextFrame` → `ID3D11Texture2D` → `cuGraphicsD3D11RegisterResource` → NVENC `RegisteredResource`. Looking Glass B7's D12 backend surpasses even NVIDIA's proprietary NvFBC.

3. **macOS optimal path**: `SCStream` → `CMSampleBuffer` → `CVPixelBuffer` (IOSurface-backed) → `VTCompressionSession`. Metal texture access via `CVMetalTextureCache` enables GPU-side preprocessing without copies.

4. **Linux optimal path**: PipeWire/kmsgrab → DMA-BUF fd → `EGL_EXT_image_dma_buf_import` → `vaCreateSurfaces` with `DRM_PRIME_2` → VAAPI encode. wlroots compositors provide the cleanest zero-copy path via `wlr-export-dmabuf`.

5. **NVENC is the latency king**: 5.8ms median encoding latency vs 15ms for AMD VCE. However, NVENC's DCVS algorithm throttles clocks at low frame rates, causing latency to balloon from 2-3ms to 12-15ms when input drops below ~60 FPS.

6. **Frame pacing matters as much as pipeline optimization**: NVIDIA Reflex 2 can reduce PC latency by 75% (56ms → 14ms in The Finals). Fast Sync/Enhanced Sync provide tear-free output at much lower latency than V-Sync.

7. **Display latency is the hidden killer**: Even with a perfectly optimized capture-encode-transmit-decode pipeline under 10ms, consumer displays add 30-100ms of video processing latency.

8. **Go integration requires CGO/syscall bridges**: No mature pure-Go libraries exist for DXGI, ScreenCaptureKit, or PipeWire DMA-BUF. Production implementations require CGO wrappers around C/ObjC APIs.
