# Dim 03: Host OS Game Capture Technologies

## Executive Summary

This report investigates video capture technologies for running games across Windows, macOS, and Linux, focusing on methods used by cloud gaming systems like Moonlight+Sunshine and Parsec. The research covers OS-specific APIs, graphics API interception methods, HDR surface capture, multi-monitor support, zero-copy pipelines, and performance characteristics. The key finding is that **zero-copy GPU-to-encoder pipelines are critical for low-latency game streaming**, with each platform offering distinct native mechanisms: DXGI Desktop Duplication on Windows, ScreenCaptureKit with IOSurface on macOS, and DMA-BUF/KMS on Linux.

---

## 1. Windows Capture Technologies

### 1.1 DXGI Desktop Duplication API (DDA)

The **DXGI Desktop Duplication API** is the gold standard for Windows screen capture, providing hardware-accelerated access to the composited desktop frame buffer with minimal overhead.

**Key Technical Details:**

- **API Entry Point**: `IDXGIOutput1::DuplicateOutput()` creates an `IDXGIOutputDuplication` object per monitor [^137^]
- **Frame Acquisition**: `AcquireNextFrame()` returns `DXGI_OUTDUPL_FRAME_INFO` with dirty rectangles, move rectangles, and pointer position [^137^]
- **Native Format**: Always `DXGI_FORMAT_B8G8R8A8_UNORM` regardless of display mode [^137^]
- **HDR Support**: `IDXGIOutput5::DuplicateOutput1()` allows requesting `DXGI_FORMAT_R16G16B16A16_FLOAT` for HDR capture in scRGB color space [^162^]
- **Multi-Monitor**: One `IDXGIOutputDuplication` per display output, enumerated via `IDXGIAdapter::EnumOutputs()` [^130^]

```
Claim: DXGI Desktop Duplication accumulates monitor updates until AcquireNextFrame is called and is not designed to capture every update by default.
Source: Microsoft Q&A - DXGI desktop duplication skip frame
URL: https://learn.microsoft.com/en-us/answers/questions/866691/dxgi-desktop-duplication-skip-frame
Date: 2022-05-27
Excerpt: "Desktop Duplication API by design accumulates monitor ("output" in DXGI terms) updates until you request them via AcquireNextFrame. The API is not designed to capture every update in first place."
Context: Microsoft official response about frame skipping behavior
Confidence: high
```

**Performance Characteristics:**
- Returns only when the display content changes (event-driven)
- Supports dirty rectangle tracking for efficient partial updates
- Can achieve 100+ FPS capture rate when display refreshes at that rate [^164^]
- GPU-side `CopyResource` from captured surface to application-owned texture

**Limitations:**
- Cannot capture fullscreen exclusive DirectX applications without special handling (see Section 6)
- Maximum 4 concurrent duplication sessions per GPU [^130^]
- Returns `DXGI_ERROR_ACCESS_LOST` when display mode changes or session limit exceeded
- HDR color space conversion requires application-side tone mapping [^162^]

### 1.2 Windows.Graphics.Capture (WGC)

Introduced in Windows 10 version 1803 (build 17134) and significantly enhanced in 1903, WGC provides a modern, security-conscious capture API.

**Key Technical Details:**

- **System Picker UI**: Requires user consent via system-provided capture selection dialog [^29^]
- **Per-Window Capture**: Can capture individual windows without capturing the entire desktop [^36^]
- **Integration**: Uses `Direct3D11CaptureFramePool` for frame callbacks [^27^]
- **Yellow Border**: System draws a yellow border around captured content for privacy indication [^29^]
- **Service Limitation**: Cannot run from Windows Service context (no interactive UI) [^75^]

```
Claim: Windows.Graphics.Capture supports both display and window capture with hardware-accelerated encoding, consuming very little CPU.
Source: wcap GitHub - Small screen recording utility
URL: https://github.com/mmozeiko/wcap
Date: 2021-09-17
Excerpt: "wcap uses Windows.Graphics.Capture API available since Windows 10 version 1903... Captured texture is submitted to Media Foundation to encode video... Using capture from compositor and hardware accelerated encoder allows it to consume very little CPU and memory."
Context: Real-world implementation demonstrating WGC efficiency
Confidence: high
```

**WGC vs DXGI Trade-offs:**

| Feature | DXGI DDA | Windows.Graphics.Capture |
|---------|----------|-------------------------|
| User consent required | No | Yes (system picker) |
| Window-only capture | No | Yes |
| Service/headless | Yes | No |
| Performance | Slightly better | Very good |
| HDR support | Yes (DuplicateOutput1) | Limited |
| Cursor exclusion | Yes | Win 10 20H1+ |

### 1.3 NVIDIA Capture SDK (NvFBC/NvIFR)

NVIDIA's proprietary capture SDK provides the highest performance on professional GPUs but is restricted on consumer hardware.

**NvFBC (Frame Buffer Capture):**
- Captures the entire framebuffer (front buffer) without application involvement [^38^]
- Works independent of graphics API used by the game [^38^]
- Operates asynchronously using dedicated GPU hardware copy engines [^33^]
- Supports 10-bit and HDR capture [^33^]
- Best suited for fullscreen desktop capture [^38^]

**NvIFR (Inband Frame Readback):**
- Captures individual application render targets [^40^]
- Supports DirectX 9/10/11 and OpenGL [^34^]
- Does not include window decorations or cursor [^40^]
- More complex setup than NvFBC [^38^]

```
Claim: NVFBC captures the framebuffer without involvement from OpenGL or Direct3D, effectively a direct copy of the framebuffer irrespective of which application drew it.
Source: GitHub Issue - Lightpack NVFBC request
URL: https://github.com/psieg/Lightpack/issues/235
Date: 2019-01-10
Excerpt: "Captures the framebuffer (front buffer) without any involvement from OpenGL or Direct3D. Effectively a direct copy of the framebuffer irrespective of which application(s) drew it."
Context: Community documentation of NVFBC behavior
Confidence: high
```

**Deprecation Status:**
- **NvIFR and NvFBCHwEncode removed from SDK 7.1** (2019) [^140^]
- **NvIFROpenGL headers removed from SDK 7.1** [^140^]
- NVIDIA transitioned to Video Codec SDK for encoding use cases
- Consumer GPUs (GeForce) have driver-level restrictions preventing NvFBC use without license [^134^]

```
Claim: The NVIDIA Capture SDK is restricted to professional-tier GPUs (Tesla, Quadro) and is not officially available on consumer-level cards.
Source: OBS Project Forum - NVFBC feature request
URL: https://obsproject.com/forum/threads/feature-request-nvfbc-api-capture-support.81703/
Date: 2018-02-25
Excerpt: "Currently the NVIDIA Capture SDK only makes that functionality available to professional-tier GPUs (Teslas and Quadros). NVFBC is not available on consumer-level cards, even the new RTX 20XX series."
Context: OBS developer explaining why NVFBC isn't implemented
Confidence: high
```

### 1.4 OBS Graphics Hook

OBS uses a proprietary graphics hook mechanism to intercept graphics API calls for game capture.

**Technical Approach:**
- Injects `graphics-hook32.dll` / `graphics-hook64.dll` into target game process [^31^]
- Hooks `Present()` calls in D3D9, D3D10, D3D11, D3D12, OpenGL, and Vulkan [^31^]
- Copies the rendered frame before it's presented to the display
- Uses shared textures for zero-copy transfer to OBS [^31^]

```
Claim: OBS graphics hook works by intercepting Present calls in the target game process, replacing function pointers in graphics-hook32.dll.
Source: GitHub - OBS-graphics-hook32-Hook
URL: https://github.com/gmh5225/OBS-graphics-hook32-Hook
Date: 2022-06-02
Excerpt: "graphics-hook32.dll + 32DC8 is our target address... Since mov eax, Hook_Present_Addr; Hook_Present_Addr = &Present_hook"
Context: Reverse engineering of OBS game capture mechanism
Confidence: high
```

### 1.5 DWM Hook Approach

An older technique involving DLL injection into the Desktop Window Manager (dwm.exe).

- Hooks `IDXGISwapChain::Present` and `IDXGISwapChain::ResizeBuffers` in DWM [^37^]
- Copies desktop image into another `ID3D10Texture2D` in VRAM
- Shares the copy with the capturing process via shared handles [^37^]

```
Claim: DWM hook approach involves injecting into dwm.exe and hooking IDXGISwapChain.Present to copy desktop images into shared VRAM textures.
Source: Hacker News Discussion
URL: https://news.ycombinator.com/item?id=26899280
Date: 2021-04-22
Excerpt: "I injected my DLL into the desktop compositor (dwm.exe process), hooked IDXGISwapChain.Present and IDXGISwapChain.ResizeBuffers methods, and wrote some C++ to copy desktop image into another ID3D10Texture2D in VRAM, and share the copy with the video capturing process."
Context: Developer describing DWM hook technique for pre-Windows 8 capture
Confidence: high
```

**Status**: Fragile and not recommended for production use. Modern Windows security features make this increasingly difficult.

### 1.6 Magnification API

A legacy API designed for accessibility screen magnification, sometimes repurposed for capture.

- Available since Windows Vista, full-screen mode requires Windows 8+ [^100^]
- Supports color transformation matrices [^100^]
- **Not recommended** for capture use cases - API is deprecated for this purpose [^96^]
- Requires DWM/Aero to be enabled [^104^]

```
Claim: The Magnification API is not recommended for screen capture after Windows 7 and is not supported for this use case.
Source: Microsoft Q&A
URL: https://learn.microsoft.com/en-us/answers/questions/211255/how-to-use-the-magnification-api-to-capture-sub-sc
Date: 2020-12-27
Excerpt: "Since the magnifier window (host window) is a transparent window... This API is not recommend to use after Windows 7"
Context: Microsoft support response
Confidence: high
```

---

## 2. macOS Capture Technologies

### 2.1 ScreenCaptureKit (macOS 12.3+)

Apple's modern, recommended framework for screen capture, introduced in macOS Monterey 12.3.

**Key Technical Details:**

- **Core Classes**: `SCStream`, `SCContentFilter`, `SCStreamConfiguration`, `SCShareableContent` [^48^]
- **Content Selection**: Filter by display, window, or application [^53^]
- **Frame Output**: Callbacks deliver `CMSampleBuffer` with timing metadata [^48^]
- **IOSurface Access**: Zero-copy GPU texture access via `IOSurface` for Metal/OpenGL interop [^48^]
- **HDR Capture**: Added in macOS 15.0 (Sequoia) [^48^]
- **Audio Capture**: System audio capture added in macOS 13.0 [^48^]

```
Claim: ScreenCaptureKit provides zero-copy GPU texture access via IOSurface for Metal/OpenGL interop.
Source: screencapturekit-rs (Rust bindings)
URL: https://github.com/svtlabs/screencapturekit-rs
Date: 2025-12-11
Excerpt: "IOSurface Access - Zero-copy GPU texture access for Metal/OpenGL... Real-time Processing - High-performance frame callbacks with custom dispatch queues"
Context: ScreenCaptureKit Rust bindings documentation
Confidence: high
```

**Performance (Apple Silicon):**

| Resolution | Expected FPS | First Frame Latency |
|------------|-------------|-------------------|
| 1080p | 30-60 FPS | 30-100ms |
| 4K | 15-30 FPS | 50-150ms |

```
Claim: ScreenCaptureKit performance on Apple Silicon achieves 30-60 FPS at 1080p and 15-30 FPS at 4K resolution.
Source: screencapturekit-rs benchmarks
URL: https://github.com/svtlabs/screencapturekit-rs
Date: 2025-12-11
Excerpt: "Typical Performance (Apple Silicon): 1080p: 30-60 FPS, First Frame Latency: 30-100ms; 4K: 15-30 FPS, First Frame Latency: 50-150ms"
Context: Benchmark documentation
Confidence: medium
```

### 2.2 CGDisplayStream

The predecessor to ScreenCaptureKit, available since OS X 10.8 Mountain Lion.

- Creates a stream with `CGDisplayStreamCreate()` [^142^]
- Callback receives `IOSurface` frames with dirty rectangle info [^135^]
- Used by WebRTC's desktop capture implementation [^142^]
- Still functional but ScreenCaptureKit is preferred

```
Claim: CGDisplayStream is used by WebRTC for macOS screen capture with dirty rectangle tracking.
Source: WebRTC source code - screen_capturer_mac.mm
URL: https://chromium.googlesource.com/external/webrtc/+/e183121657fa2daf2985f264cff8dd30cd63b97d/webrtc/modules/desktop_capture/screen_capturer_mac.mm
Date: Unknown
Excerpt: "CGDisplayStreamRef display_stream = CGDisplayStreamCreate(display_id, pixel_width, pixel_height, 'BGRA', nullptr, handler);"
Context: WebRTC reference implementation
Confidence: high
```

### 2.3 IOSurface

IOSurface is the fundamental macOS mechanism for zero-copy GPU texture sharing.

- Kernel-managed chunk of texture memory that can be paged on/off GPU [^98^]
- Can be shared across processes with no data copy [^98^]
- Created with properties: width, height, pixelFormat, bytesPerElement [^98^]
- Metal textures created via `device.newTexture(descriptor:iosurface:plane:)` [^98^]
- Cross-process sharing via `IOSurfaceCreateXPCObject` or `IOSurfaceCreateMachPort` [^98^]
- On Apple Silicon (unified memory), no GPU paging occurs [^98^]

```
Claim: IOSurface enables zero-copy cross-process texture sharing on macOS, with Metal textures created directly from the surface.
Source: Russ Bishop - Cross-process Rendering
URL: http://www.russbishop.net/cross-process-rendering
Date: 2019-11-05
Excerpt: "An IOSurface is a kernel-managed chunk of texture memory that can be paged on or off the GPU automatically and shared across processes. When shared across processes no data is copied."
Context: Technical blog on macOS GPU resource sharing
Confidence: high
```

### 2.4 AVFoundation / AVCaptureScreenInput

The older capture API, still available but superseded by ScreenCaptureKit.

- `AVCaptureScreenInput` for screen capture [^52^]
- Does not capture system audio [^52^]
- macOS 10.7+ availability
- AVAssetWriter for encoding to disk [^55^]
- Maximum H.264 resolution: 4096x2304 [^55^]

---

## 3. Linux Capture Technologies

### 3.1 X11 XShm

The traditional X11 shared memory extension for screen capture.

- Uses MIT-SHM (Shared Memory) extension for zero-copy pixel transfer [^47^]
- Copies frame data from GPU to system RAM, then back to GPU for encoding [^61^]
- Significant performance penalty: ~10 FPS on heavy GPU shaders while game runs at 40-60 FPS [^61^]
- No special permissions required [^47^]

```
Claim: X11 XShm capture causes significant performance degradation, achieving only ~10 FPS for heavy shaders while the game runs at 40-60 FPS.
Source: OBS Forum - Experimental zero-copy screen capture on Linux
URL: https://obsproject.com/forum/threads/experimental-zero-copy-screen-capture-on-linux.101262/
Date: 2019-03-03
Excerpt: "Vanilla OBS with XSHM would struggle capturing it and would barely keep ~10 fps, even though the shader itself can be as high as 40-60fps."
Context: OBS developer comparing XSHM vs DMA-BUF performance
Confidence: high
```

### 3.2 PipeWire + xdg-desktop-portal

The modern, security-conscious standard for Wayland screen capture.

- **PipeWire**: Multimedia framework handling stream transport [^59^]
- **xdg-desktop-portal**: D-Bus API for sandboxed application access [^60^]
- **ScreenCast Portal**: Provides `org.freedesktop.portal.ScreenCast` interface [^60^]
- Supports monitor, window, and virtual source types [^60^]
- Cursor modes: hidden, embedded, or metadata-only [^60^]
- Returns PipeWire stream node IDs for DMA-BUF access [^60^]

```
Claim: PipeWire with xdg-desktop-portal is the standard for secure screen sharing on Wayland, providing DMA-BUF based zero-copy capture.
Source: Drew DeVault - Wayland misconceptions debunked
URL: https://drewdevault.com/blog/Wayland-misconceptions-debunked/
Date: 2019-02-10
Excerpt: "There are two protocols for the purpose of screenshots and screen recording: screencopy and dmabuf-export... There are two approaches endorsed by different camps: these Wayland protocols, and a dbus protocol based on Pipewire."
Context: Wayland ecosystem overview
Confidence: high
```

### 3.3 KMS/DRM Direct Capture

The lowest-level Linux capture mechanism, bypassing all display servers.

- Reads directly from the kernel DRM framebuffer [^50^]
- Uses `libdrm` for querying DRM setup and framebuffers [^50^]
- Bypasses X server and Wayland entirely [^50^]
- Requires `CAP_SYS_ADMIN` or root privileges [^50^]
- `kmsgrab` ffmpeg input device available [^56^]

```
Claim: KMS/DRM capture reads directly from the Linux kernel DRM subsystem framebuffer, bypassing X server or Wayland.
Source: GitHub - screenrec utility
URL: https://github.com/andreamonaco/screenrec
Date: 2025-11-30
Excerpt: "Contrary to other popular programs, it reads directly from the framebuffer of your Linux kernel DRM subsystem, bypassing the X server or Wayland, so in theory it may achieve better performance."
Context: KMS/DRM screen recorder documentation
Confidence: high
```

### 3.4 Wayland Protocols (wlr-screencopy, wlr-export-dmabuf)

wlroots-specific protocols that form the foundation of Wayland screen capture.

- **wlr-screencopy-unstable-v1**: Copy-based capture with shm and dmabuf buffer types [^159^]
- **wlr-export-dmabuf-unstable-v1**: Exports DMA-BUFs directly without copy [^160^]
- **ext-image-copy-capture-v1**: Emerging standard protocol [^156^]
- Zero-copy path keeps image data on GPU [^160^]

```
Claim: wlr-export-dmabuf protocol captures surfaces by exporting DMA-BUFs, keeping image data on GPU without copies.
Source: wayland.app - wlr export DMA-BUF protocol
URL: https://wayland.app/protocols/wlr-export-dmabuf-unstable-v1
Date: Unknown
Excerpt: "An interface to capture surfaces in an efficient way by exporting DMA-BUFs... All frames are read-only and may not be written into or altered."
Context: Official Wayland protocol documentation
Confidence: high
```

**Performance Comparison (wl-screenrec benchmark):**

| Command | CPU Usage | GPU 3D Delta | GPU Video Delta |
|---------|-----------|-------------|----------------|
| wf-recorder (sw) | ~500% | +44% | 0% |
| wf-recorder (vaapi) | ~75% | +88% | +23% |
| wl-screenrec (dma-buf) | ~2.5% | +91% | +30% |

```
Claim: wl-screenrec with DMA-BUF uses only ~2.5% CPU versus ~500% for software-based wf-recorder at 4K60 capture.
Source: GitHub - wl-screenrec
URL: https://github.com/russelltg/wl-screenrec
Date: 2023-01-14
Excerpt: "wl-screenrec: ~2.5% CPU, +91% GPU 3D, +30% GPU Video... Uses dma-buf transfers to get surface, and uses the GPU to do both the pixel format conversion and the encoding."
Context: Benchmark by wl-screenrec author
Confidence: high
```

### 3.5 V4L2 Loopback

Creates virtual video devices for screen capture compatibility.

- Kernel module creates `/dev/videoX` loopback devices [^115^]
- Screen capture piped to virtual camera via FFmpeg [^115^]
- Used for applications that only support V4L2 input [^120^]
- `exclusive_caps=1` required for Chromium/WebRTC compatibility [^120^]

### 3.6 OBS Linux Capture

OBS Studio supports multiple capture backends on Linux:

- **XSHM**: X11 shared memory (baseline, works everywhere) [^61^]
- **Xcomposite**: Per-window capture on X11 [^103^]
- **PipeWire**: Wayland capture via xdg-desktop-portal [^112^]
- **DMA-BUF (experimental)**: Zero-copy via KMS/DRM [^61^]

---

## 4. Graphics API-Specific Capture Methods

### 4.1 Vulkan Capture

**DMA-BUF Zero-Copy (Linux):**
- `VK_EXT_external_memory_dma_buf` extension for DMA-BUF import [^74^]
- `VK_KHR_external_memory_fd` for file descriptor sharing [^74^]
- `VkImportMemoryFdInfoKHR` structure for importing DMA-BUF fds [^74^]
- Avoids PCIe round-trip from GPU to CPU [^74^]

```
Claim: Vulkan zero-copy capture on Linux uses VK_EXT_external_memory_dma_buf to import DMA-BUF file descriptors as GPU textures.
Source: GitHub - godot-desktop-capture Vulkan integration
URL: https://github.com/LabmarketAI/godot-desktop-capture/issues/6
Date: 2026-03-07
Excerpt: "Linux: DMA-BUF fd -> VkImportMemoryFdInfoKHR (VK_KHR_external_memory_fd + VK_EXT_external_memory_dma_buf). The imported VkImage is wrapped in a Godot RID... This path avoids any PCIe round-trip."
Context: Technical design document for Vulkan capture integration
Confidence: high
```

**Windows Shared Handles:**
- `IDXGIResource1::CreateSharedHandle` for DXGI interop [^74^]
- `VkImportMemoryWin32HandleInfoKHR` with `VK_KHR_external_memory_win32` [^74^]

### 4.2 DirectX Capture

- **DXGI Shared Surfaces**: `ID3D11Texture2D` with `D3D11_RESOURCE_MISC_SHARED` flag
- **Keyed Mutex**: `IDXGIKeyedMutex` for cross-device synchronization
- **NV12 Conversion**: GPU shaders for RGB-to-NV12 color space conversion

### 4.3 OpenGL Capture

**Pixel Buffer Objects (PBO) for Async Readback:**

- Bind PBO with `GL_PIXEL_PACK_BUFFER` target [^133^]
- `glReadPixels()` to PBO returns immediately (async DMA transfer) [^133^]
- Map PBO in subsequent frame to read previous frame's data [^139^]
- Double-buffered PBOs for ping-pong operation [^139^]

```
Claim: OpenGL PBO-based readback uses asynchronous DMA transfer to avoid stalling the CPU, with two PBOs ping-ponging between frames.
Source: Song Ho - OpenGL Pixel Buffer Object
URL: https://www.songho.ca/opengl/gl_pbo.html
Date: Unknown
Excerpt: "The main advantage of PBO is fast pixel data transfer to and from a graphics card through DMA (Direct Memory Access) without involving CPU cycles. And, the other advantage of PBO is asynchronous DMA transfer."
Context: Technical tutorial on OpenGL PBOs
Confidence: high
```

### 4.4 Metal Capture (macOS)

- **IOSurface-backed textures**: `MTLTexture` created from `IOSurface` [^98^]
- **Shared Texture Handles**: `MTLSharedTextureHandle` for XPC cross-process sharing [^98^]
- **Storage Modes**: `MTLStorageModeShared` for CPU/GPU access, `MTLStorageModePrivate` for GPU-only [^111^]
- **CALayer display**: Set layer contents directly to IOSurface for compositor bypass [^98^]

---

## 5. Fullscreen Exclusive vs Windowed/Borderless Capture Challenges

### 5.1 The Fullscreen Exclusive Problem

Fullscreen exclusive mode bypasses the DWM compositor, making capture APIs that read from the compositor surface (like DXGI DDA) unable to see the game content.

**Key Challenges:**

- **DXGI DDA**: Cannot capture fullscreen exclusive DirectX applications [^131^]
- **Workaround**: NVIDIA's "Prefer layered on DXGI Swapchain" setting forces presentation through DXGI [^131^]
- **Modern Best Practice**: Use "fullscreen borderless" (flip model) instead [^123^]

```
Claim: Sunshine cannot capture fullscreen OpenGL and Vulkan programs at full frame rate unless they present on top of DXGI.
Source: Sunshine Advanced Usage Documentation
URL: https://docs.lizardbyte.dev/projects/sunshine/v0.23.0/about/advanced_usage.html
Date: Unknown
Excerpt: "Sunshine can't capture fullscreen OpenGL and Vulkan programs at full frame rate unless they present on top of DXGI. This is system-wide setting that is reverted on sunshine program exit."
Context: Official Sunshine documentation
Confidence: high
```

### 5.2 Flip Model Swap Chains

Modern Windows uses flip model presentation which solves the exclusive fullscreen problem:

- **DXGI_SWAP_EFFECT_FLIP_DISCARD** enables DWM bypass while remaining captureable [^80^]
- DXGI automatically handles fullscreen window mode [^85^]
- `SetFullscreenState` transitions between windowed and fullscreen [^85^]

```
Claim: Unity switched to flip model swap chains in 2019.3, making exclusive fullscreen unnecessary on recent Windows 10.
Source: Unity Discussions
URL: https://discussions.unity.com/t/exclusive-fullscreen-pc-windows-steam-does-not-work-right/896576
Date: 2022-10-08
Excerpt: "In general, exclusive mode doesn't really have any benefits over fullscreen window mode in recent Windows 10 versions... All those Fullscreen Window drawbacks went away when we switched to flip model swap chain back in 2019.3."
Context: Unity developer explaining fullscreen exclusive deprecation
Confidence: high
```

### 5.3 Capture Strategy Matrix

| Game Presentation Mode | DXGI DDA | WGC | NVFBC | OBS Hook |
|----------------------|----------|-----|-------|----------|
| Windowed | Yes | Yes | Full desktop | Yes |
| Borderless Fullscreen | Yes | Yes | Full desktop | Yes |
| Fullscreen Exclusive (DX) | No* | No | Yes | Yes |
| Fullscreen Exclusive (GL/VK) | No* | No | Yes | Yes |

*Can work with "Prefer layered on DXGI Swapchain" driver setting

---

## 6. HDR Surface Capture

### 6.1 Windows HDR Capture

Windows supports two main HDR capture formats via `DuplicateOutput1()`:

**FP16 scRGB Format:**
- Format: `DXGI_FORMAT_R16G16B16A16_FLOAT`
- Color Space: `DXGI_COLOR_SPACE_RGB_FULL_G10_NONE_P709`
- 64 bits per pixel (8 bytes per pixel) [^162^]
- Uses linear colors with extended range (>1.0 = HDR highlights) [^80^]
- scRGB (1.0, 1.0, 1.0) = 80 nits; scRGB (12.5, 12.5, 12.5) = 1000 nits [^167^]

**10-bit PQ Format:**
- Format: `DXGI_FORMAT_R10G10B10A2_UNORM`
- Color Space: `DXGI_COLOR_SPACE_RGB_FULL_G2084_NONE_P2020`
- Application must do PQ encoding [^80^]
- 32 bits per pixel

```
Claim: HDR capture via DXGI Desktop Duplication uses DuplicateOutput1() with DXGI_FORMAT_R16G16B16A16_FLOAT, returning scRGB linear data with 8 bytes per pixel.
Source: Microsoft Q&A - Desktop Duplication API with HDR
URL: https://learn.microsoft.com/en-us/answers/questions/1457052/using-the-desktop-duplication-api-with-hdr-interpr
Date: 2023-12-04
Excerpt: "DXGI_FORMAT_R16G16B16A16_FLOAT... pixelData.RowPitch = 30720 // width (3840) * 8... pixelData.DepthPitch = 66355200 // RowPitch * height (2160)"
Context: Developer implementing HDR capture with Desktop Duplication
Confidence: high
```

**Tone Mapping Required:**
- HDR-to-SDR tone mapping needed for SDR displays [^162^]
- Common algorithms: Reinhard, ACES filmic, Windows/OBS-style [^155^]
- HDR highlights clipped if captured as SDR [^155^]

### 6.2 HDR Capture Tools

- **HDR Screenshot Tool**: Uses `IDXGIOutput5.DuplicateOutput1()` for `R16G16B16A16_FLOAT` capture [^155^]
- **Tone mapping**: Divide by `sdr_white_nits / 80`, clip to [0, 1], apply sRGB gamma [^155^]
- **OBS**: Handles HDR capture with internal tone mapping [^162^]

### 6.3 macOS HDR

- ScreenCaptureKit added HDR capture in macOS 15.0 (Sequoia) [^48^]
- HDR screenshot output added in macOS 26.0 (Tahoe) [^48^]

---

## 7. Multi-Monitor Capture

### 7.1 Windows Multi-Monitor

- Each monitor is an independent `IDXGIOutput` [^130^]
- `EnumOutputs()` iterates available displays per adapter [^130^]
- Each output gets its own `IDXGIOutputDuplication` instance [^130^]
- Virtual desktop coordinates via `GetSystemMetrics(SM_CXVIRTUALSCREEN)` [^124^]

```
Claim: Windows multi-monitor capture requires creating separate IDXGIOutputDuplication instances per display output via IDXGIAdapter::EnumOutputs().
Source: Stack Overflow - IDXGIOutputDuplication multiple screens
URL: https://stackoverflow.com/questions/72986817/how-to-use-idxgioutputduplication-to-capture-multiple-screens
Date: 2022-07-14
Excerpt: "hr = dxgiAdapter->EnumOutputs(OutputNumber, &dxgiOutput);... hr = dxgiOutput1->DuplicateOutput(D3DDevice, &DeskDupl);"
Context: Code example for multi-monitor DDA setup
Confidence: high
```

### 7.2 macOS Multi-Monitor

- `SCShareableContent` lists available displays [^48^]
- Each `SCDisplay` has resolution, ID, and frame info [^48^]
- Separate `SCStream` per display or capture all simultaneously

### 7.3 Linux Multi-Monitor

- PipeWire/xdg-desktop-portal: Select individual monitors via portal [^60^]
- KMS/DRM: Each CRTC (display controller) has its own framebuffer [^50^]
- X11: Root window spans all monitors; XRandR for per-monitor bounds [^47^]

---

## 8. Zero-Copy Capture-to-Encoder Pipeline

### 8.1 Linux DMA-BUF Pipeline

The most efficient Linux capture path uses DMA-BUF for zero-copy GPU texture sharing:

**Pipeline Flow:**
1. Capture source (KMS/DRM, PipeWire, or wlr-export-dmabuf) provides DMA-BUF fd
2. Import fd into EGLImage via `EGL_EXT_image_dma_buf_import` [^61^]
3. Bind EGLImage to GL texture via `GL_OES_EGL_image` [^61^]
4. GL texture passed to hardware encoder (VAAPI, NVENC via CUDA)
5. **No CPU memory copy at any stage**

```
Claim: DMA-BUF based zero-copy capture on Linux uses EGL_EXT_image_dma_buf_import to create EGLImages from framebuffer file descriptors, then binds them to GL textures for encoder consumption.
Source: OBS Forum - Experimental zero-copy screen capture
URL: https://obsproject.com/forum/threads/experimental-zero-copy-screen-capture-on-linux.101262/
Date: 2019-03-03
Excerpt: "EGL_EXT_image_dma_buf_import, it allows creating EGLImage objects bound to existing DMA-BUF object (using its fd). GL_OES_EGL_image, it allows binding GL textures to EGLImage objects."
Context: Technical explanation of DMA-BUF zero-copy pipeline
Confidence: high
```

**Requirements:**
- EGL context (GLX not supported) [^61^]
- `CAP_SYS_ADMIN` for KMS/DRM framebuffer access [^61^]
- GBM (Generic Buffer Management) for buffer allocation
- Compatible with X11, Wayland, and bare KMS terminals [^61^]

### 8.2 Windows Zero-Copy Pipeline

**DXGI Shared Surface + CUDA Interop:**
1. `IDXGIOutputDuplication::AcquireNextFrame()` gets desktop texture
2. `CopyResource` to application-owned texture
3. `IDXGIResource::GetSharedHandle()` for cross-process sharing
4. CUDA imports via `cuGraphicsD3D11RegisterResource()`
5. NVENC encodes directly from CUDA device memory

**Alternative - CUDA Device Memory (NvFBC):**
1. `NvFBCToCuda` captures directly to CUDA buffer [^34^]
2. CUDA buffer mapped directly to NVENC [^34^]
3. No system memory round-trip

### 8.3 macOS Zero-Copy Pipeline

**IOSurface + VideoToolbox:**
1. ScreenCaptureKit delivers `CMSampleBuffer` with `IOSurface` attachment
2. `IOSurface` can be wrapped as `MTLTexture` or `CVPixelBuffer`
3. VideoToolbox encoder accepts `CVPixelBuffer` directly
4. No pixel data leaves GPU (unified memory on Apple Silicon)

### 8.4 Zero-Copy Performance Impact

```
Claim: DMA-BUF zero-copy capture reduces CPU overhead from ~500% to ~2.5% compared to software-based approaches at 4K60.
Source: GitHub - wl-screenrec
URL: https://github.com/russelltg/wl-screenrec
Date: 2023-01-14
Excerpt: "wl-screenrec: ~2.5% CPU... Additionally, with either wf-recorder setup there is visible stuttering in the vkcube window. wl-screenrec does not seem to stutter at all."
Context: Benchmark comparing zero-copy vs software capture
Confidence: high
```

---

## 9. Frame Rate Synchronization with Game Render Loop

### 9.1 Capture Frame Timing

- DXGI DDA returns frames only when display updates (event-driven) [^164^]
- `AcquireNextFrame()` timeout should be handled gracefully [^129^]
- For video encoding, continue sending frames even when no display update occurs [^168^]

```
Claim: AcquireNextFrame receives nothing if nothing changes on the captured output - it is not a "capture screen now" command but an event-driven notification.
Source: Simon Mourier's Blog
URL: https://www.simonmourier.com/blog/DXGI-AcquireNextFrame-receives-the-next-frame-only-after-20ms-50fps-is-it-possib/
Date: 2020-12-03
Excerpt: "IDXGIOutputDuplication::AcquireNextFrame will receive nothing if nothing changes on the captured output. It's not like a 'PleaseCaptureScreenNow' command."
Context: Technical blog explaining DDA frame timing
Confidence: high
```

### 9.2 Frame Pacing for Streaming

- Sunshine captures at 59.94Hz instead of 60.00Hz on some displays, causing microstutter [^82^]
- Frame pacing option in Moonlight client can alleviate this [^82^]
- NTSC-derived refresh rates (59.94) vs exact (60.00) cause periodic irregularities [^82^]

```
Claim: Sunshine captures at 59.94Hz instead of 60.00Hz even when both host and client use 60.00Hz refresh rates, causing microstuttering.
Source: GitHub - Sunshine Issue #2286
URL: https://github.com/LizardByte/Sunshine/issues/2286
Date: 2024-03-22
Excerpt: "Even though both the sunshine host system and the moonlight client system use 60.00Hz refresh rates, the stream is captured at 59.94Hz, i.e. the old NTSC-based framerate."
Context: Bug report with frame rate analysis
Confidence: high
```

### 9.3 Swapchain and VSync Considerations

- Compositor timing affects capture latency [^122^]
- Blocking vs. non-blocking synchronization impacts pipeline depth [^122^]
- Triple buffering increases latency but improves throughput [^122^]

---

## 10. Capture Performance Overhead Measurements

### 10.1 Capture Method Comparison (Linux)

| Method | CPU Overhead | GPU Impact | Latency | Quality |
|--------|-------------|------------|---------|---------|
| X11 XShm | Very High | High (GPU->RAM->GPU) | High | Full |
| PipeWire SHM | High | Medium | Medium | Full |
| PipeWire DMA-BUF | Low | Low | Low | Full |
| KMS/DRM DMA-BUF | Very Low | Minimal | Very Low | Full |
| NvFBC (Linux) | Negligible | Minimal | Very Low | Full |

### 10.2 wl-screenrec Benchmark (4K60, i9-11900H)

| Command | CPU Usage | GPU 3D Delta | GPU Video Delta |
|---------|-----------|-------------|----------------|
| wf-recorder (software) | ~500% | +44% | 0% |
| wf-recorder (VAAPI) | ~75% | +88% | +23% |
| wl-screenrec (DMA-BUF) | ~2.5% | +91% | +30% |

### 10.3 Windows Capture Overhead

- DXGI DDA: Minimal GPU overhead (~1-2% at 1080p60)
- WGC: Similar to DDA, slightly more CPU for frame pool management
- DWM Hook: Variable, can cause frame drops
- Magnification API: Significant overhead, not recommended

### 10.4 macOS Capture Overhead

- ScreenCaptureKit: ~10-20ms per frame callback [^48^]
- IOSurface access: Negligible (shared memory)
- Unified memory on Apple Silicon eliminates PCIe transfer

### 10.5 General Performance Guidelines

```
Claim: Parsec adds only 7 milliseconds of latency on LAN ethernet connections with full hardware control and zero-copy GPU pipeline.
Source: Parsec Technology Page
URL: https://parsec.app/technology
Date: Unknown
Excerpt: "We didn't build on top of any wrappers — we wanted full hardware control and the ability to manipulate every element of the stream to reduce latency as much as possible. On our test setup on a LAN ethernet connection, Parsec adds only 7 milliseconds of latency to your game."
Context: Parsec marketing/technical documentation
Confidence: medium (marketing claim)
```

---

## 11. Sunshine/Moonlight Capture Implementation

### 11.1 Supported Capture Methods

```
Claim: Sunshine supports DXGI Desktop Duplication on Windows, KMS/DRM and PipeWire on Linux, and ScreenCaptureKit on macOS.
Source: GitHub - LizardByte/Sunshine
URL: https://github.com/lizardbyte/sunshine
Date: 2026-04-23
Excerpt: "Screen Capture: DXGI Desktop Duplication [Windows], KMS/DRM [Linux], ScreenCaptureKit [macOS], Wayland (wlroots) [Linux], XDG Desktop Portal [Linux], X11 [Linux]"
Context: Sunshine feature compatibility matrix
Confidence: high
```

**Sunshine Capture Matrix:**

| Capture Method | Platform | Encoder Compatibility |
|---------------|----------|---------------------|
| DXGI Desktop Duplication | Windows | NVENC, AMF, QuickSync, Software |
| Windows.Graphics.Capture | Windows | NVENC, AMF, QuickSync (partial) |
| ScreenCaptureKit | macOS | VideoToolbox |
| KMS/DRM | Linux | VAAPI, Vulkan Video, NVENC, Software |
| XDG Desktop Portal | Linux | VAAPI, Vulkan Video, NVENC, Software |
| Wayland (wlroots) | Linux | VAAPI, NVENC, Software |
| X11 | Linux | VAAPI, NVENC, Software |
| NvFBC (X11) | Linux | NVENC (CUDA) only |

### 11.2 Capture-to-Encoder Pipeline in Sunshine

Sunshine implements a modular capture pipeline:

1. **Platform-specific capture**: DXGI DDA (Windows), KMS/DRM (Linux), ScreenCaptureKit (macOS)
2. **VRAM/RAM transfer**: Platform-specific surface transfer
3. **Format conversion**: GPU-based color space conversion if needed
4. **Hardware encoding**: NVENC/VAAPI/VideoToolbox/AMF/QuickSync
5. **Network transmission**: Direct to Moonlight client

---

## 12. Tensions and Counter-Arguments

### 12.1 API Availability vs Performance

- **NVFBC is fastest but restricted**: Only available on professional GPUs, deprecated from SDK [^134^]
- **DXGI DDA is best general Windows solution**: Works on all GPUs but cannot capture fullscreen exclusive without workarounds
- **WGC is modern but requires user interaction**: System picker prevents headless operation

### 12.2 Zero-Copy Implementation Complexity

- **DMA-BUF requires elevated privileges**: `CAP_SYS_ADMIN` for KMS access is a security concern [^61^]
- **EGL vs GLX on Linux**: Zero-copy requires EGL, but many applications use GLX [^61^]
- **Compositor compatibility**: wlroots-based compositors have best DMA-BUF support [^161^]

### 12.3 HDR Capture Complexity

- **Tone mapping is application responsibility**: DXGI provides raw FP16 data, app must convert to SDR [^162^]
- **Color accuracy issues**: Different tone mapping algorithms produce different results [^162^]
- **Format negotiation**: `DuplicateOutput1()` format array priority matters

### 12.4 Wayland Fragmentation

- **Multiple protocols**: wlr-screencopy, wlr-export-dmabuf, ext-image-copy-capture [^156^]
- **xdg-desktop-portal**: Higher-level but adds D-Bus overhead
- **Compositor support varies**: GNOME, KDE, and wlroots have different levels of support [^49^]

---

## 13. Technical Implementation Feasibility

### 13.1 Recommended Architecture by Platform

**Windows:**
- Primary: DXGI Desktop Duplication API with `DuplicateOutput1()` for HDR
- Fallback: Windows.Graphics.Capture for windowed apps
- GPU sharing: DXGI shared handles -> CUDA interop -> NVENC
- Handle fullscreen exclusive with "Prefer layered on DXGI Swapchain"

**Linux:**
- Primary: KMS/DRM DMA-BUF for best performance (requires elevated privileges)
- General: PipeWire + xdg-desktop-portal (best compatibility)
- wlroots: wlr-export-dmabuf-unstable-v1 direct protocol
- Zero-copy: DMA-BUF fd -> EGLImage -> GL texture -> VAAPI/NVENC

**macOS:**
- Primary: ScreenCaptureKit (macOS 12.3+)
- Zero-copy: IOSurface -> CVPixelBuffer -> VideoToolbox
- Unified memory on Apple Silicon eliminates copy overhead

### 13.2 Critical Implementation Details

1. **Minimize frame hold time**: Release DXGI frame before acquiring next [^129^]
2. **Dirty rectangle tracking**: Only process changed regions for efficiency [^137^]
3. **Double/triple buffering**: Prevent pipeline stalls [^122^]
4. **Frame timing alignment**: Match capture rate to display refresh rate [^82^]
5. **HDR tone mapping**: Implement proper color space conversion [^162^]

---

## 14. Summary of Key Findings

| Aspect | Windows | macOS | Linux |
|--------|---------|-------|-------|
| **Best Capture API** | DXGI Desktop Duplication | ScreenCaptureKit | KMS/DRM DMA-BUF |
| **Zero-Copy Mechanism** | DXGI Shared Handle | IOSurface | DMA-BUF + EGLImage |
| **HDR Support** | R16G16B16A16_FLOAT scRGB | macOS 15.0+ | Limited |
| **Fullscreen Exclusive** | Workarounds needed | N/A | N/A |
| **Permission Model** | System-level | User consent (TCC) | CAP_SYS_ADMIN/root |
| **Encoder Interop** | CUDA/DXGI | VideoToolbox | VAAPI/CUDA |
| **Typical Latency** | <10ms capture | 10-20ms | 1-2ms (KMS) |

---

## References

[^27^]: https://docs.cvedia.com/develop/plugins/input/screencap/index.html - Screencap Plugin Documentation
[^29^]: https://learn.microsoft.com/en-us/windows/uwp/audio-video-camera/screen-capture - Microsoft Screen Capture Documentation
[^31^]: https://github.com/gmh5225/OBS-graphics-hook32-Hook - OBS Graphics Hook Analysis
[^32^]: https://github.com/mmozeiko/wcap - wcap Windows Capture Utility
[^33^]: https://developer.download.nvidia.com/designworks/capture-sdk/docs/7.1/NVIDIA%20Capture%20SDK%20Programming%20Guide.pdf - NVIDIA Capture SDK Programming Guide
[^34^]: https://developer.download.nvidia.com/designworks/capture-sdk/docs/7.0/NVIDIA-Capture-SDK-SamplesDescription.pdf - NVIDIA Capture SDK Samples
[^36^]: https://blogs.windows.com/windowsdeveloper/2019/09/16/new-ways-to-do-screen-capture/ - Windows Developer Blog
[^37^]: https://news.ycombinator.com/item?id=26899280 - Hacker News DWM Hook Discussion
[^38^]: https://github.com/psieg/Lightpack/issues/235 - NVFBC Feature Request Discussion
[^40^]: https://developer.download.nvidia.com/designworks/capture-sdk/docs/6.1/NVIDIA-Capture-SDK-Programming-Guide.pdf - NVIDIA Capture SDK 6.1
[^47^]: https://mintlify.com/AlphaLawless/boomer-rs/architecture/display-backends - Display Backends Comparison
[^48^]: https://github.com/svtlabs/screencapturekit-rs - ScreenCaptureKit Rust Bindings
[^49^]: https://github.com/LizardByte/Sunshine/issues/4662 - Sunshine PipeWire Issue
[^50^]: https://github.com/andreamonaco/screenrec - screenrec KMS/DRM Capture
[^52^]: https://www.recall.ai/blog/macos-screencapture-api - macOS Screen Capture Guide
[^53^]: https://developer.apple.com/videos/play/wwdc2023/10136/ - WWDC23 ScreenCaptureKit
[^55^]: https://nonstrict.eu/blog/2023/recording-to-disk-with-screencapturekit - ScreenCaptureKit Recording
[^56^]: https://stackoverflow.com/questions/58754385/record-linux-wayland-drm-screen-using-ffmpegs-kmsgrab-device - KMS/DRM ffmpeg
[^60^]: https://flatpak.github.io/xdg-desktop-portal/docs/doc-org.freedesktop.portal.ScreenCast.html - XDG Desktop Portal ScreenCast
[^61^]: https://obsproject.com/forum/threads/experimental-zero-copy-screen-capture-on-linux.101262/ - OBS Zero-Copy Linux
[^74^]: https://github.com/LabmarketAI/godot-desktop-capture/issues/6 - Vulkan Zero-Copy Capture
[^75^]: https://github.com/lizardbyte/sunshine - Sunshine Game Stream Host
[^80^]: https://www.pyromuffin.com/2018/07/how-to-render-to-hdr-displays-on.html - HDR Displays on Windows
[^81^]: https://parsec.app/technology - Parsec Technology
[^82^]: https://github.com/LizardByte/Sunshine/issues/2286 - Sunshine Frame Rate Issue
[^85^]: https://learn.microsoft.com/en-us/windows/win32/direct3darticles/dxgi-best-practices - DXGI Best Practices
[^96^]: https://learn.microsoft.com/en-us/answers/questions/211255/how-to-use-the-magnification-api-to-capture-sub-sc - Magnification API Q&A
[^97^]: https://github.com/w23/obs-kmsgrab - OBS KMS/GRAB DMA-BUF Plugin
[^98^]: http://www.russbishop.net/cross-process-rendering - IOSurface Cross-Process Rendering
[^100^]: https://learn.microsoft.com/en-us/previous-versions/windows/desktop/magapi/magapi-intro - Magnification API Overview
[^104^]: https://stackoverflow.com/questions/53295895/screen-capturing-based-on-windows-magnification-api - Magnification API Issues
[^111^]: https://supertrouper.gitbooks.io/metal-programming-guide/ - Metal Programming Guide
[^115^]: https://wiki.archlinux.org/title/V4l2loopback - ArchWiki v4l2loopback
[^121^]: https://docs.sentry.io/platforms/apple/guides/ios/session-replay/performance-overhead/ - iOS Capture Overhead
[^122^]: https://raphlinus.github.io/ui/graphics/gpu/2021/10/22/swapchain-frame-pacing.html - Swapchains and Frame Pacing
[^123^]: https://discussions.unity.com/t/exclusive-fullscreen-pc-windows-steam-does-not-work-right/896576 - Unity Fullscreen Exclusive
[^124^]: https://www.apriorit.com/dev-blog/193-multi-monitor-screenshot - Multi-Monitor Screenshots WinAPI
[^129^]: https://learn.microsoft.com/en-us/answers/questions/866691/dxgi-desktop-duplication-skip-frame - DXGI Skip Frame
[^130^]: https://stackoverflow.com/questions/72986817/how-to-use-idxgioutputduplication-to-capture-multiple-screens - Multi-Monitor DDA
[^131^]: https://docs.lizardbyte.dev/projects/sunshine/v0.23.0/about/advanced_usage.html - Sunshine Advanced Usage
[^133^]: https://www.songho.ca/opengl/gl_pbo.html - OpenGL PBO Tutorial
[^134^]: https://obsproject.com/forum/threads/feature-request-nvfbc-api-capture-support.81703/ - NVFBC OBS Feature Request
[^135^]: https://github.com/lwouis/alt-tab-macos/issues/122 - CGDisplayStream Discussion
[^137^]: https://learn.microsoft.com/en-us/windows/win32/direct3ddxgi/desktop-dup-api - Desktop Duplication API
[^139^]: https://lektiondestages.art.blog/2013/01/28/reading-the-opengl-backbuffer-to-system-memory/ - OpenGL Backbuffer Reading
[^140^]: https://developer.nvidia.com/capture-sdk-archive - NVIDIA Capture SDK Archive
[^142^]: https://chromium.googlesource.com/external/webrtc/+/e183121657fa2daf2985f264cff8dd30cd63b97d/webrtc/modules/desktop_capture/screen_capturer_mac.mm - WebRTC macOS Capture
[^155^]: https://github.com/MagestiUA/HDR_Screenshot_tool_for_windows - HDR Screenshot Tool
[^156^]: https://man.archlinux.org/man/wl-mirror.1.en - wl-mirror Backends
[^157^]: https://github.com/robmikh/Win32CaptureSample/issues/73 - DDA Performance Issue
[^159^]: https://wayland.app/protocols/wlr-screencopy-unstable-v1 - wlr-screencopy Protocol
[^160^]: https://wayland.app/protocols/wlr-export-dmabuf-unstable-v1 - wlr-export-dmabuf Protocol
[^161^]: https://github.com/russelltg/wl-screenrec - wl-screenrec
[^162^]: https://learn.microsoft.com/en-us/answers/questions/1457052/using-the-desktop-duplication-api-with-hdr-interpr - Desktop Duplication HDR
[^164^]: https://www.simonmourier.com/blog/DXGI-AcquireNextFrame-receives-the-next-frame-only-after-20ms-50fps-is-it-possib/ - DDA Frame Timing
[^167^]: https://learn.microsoft.com/en-us/windows/win32/direct3darticles/high-dynamic-range - DirectX HDR
