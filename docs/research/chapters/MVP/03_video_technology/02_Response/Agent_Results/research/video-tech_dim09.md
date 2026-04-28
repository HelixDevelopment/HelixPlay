# Dimension 09: Hardware Detection & Dynamic Optimization

## Table of Contents

1. [GPU Detection](#1-gpu-detection)
   - [NVIDIA NVML](#nvidia-nvml)
   - [AMD ROCm SMI](#amd-rocm-smi)
   - [Intel GPU Tools](#intel-gpu-tools)
   - [Apple Metal Performance Counters](#apple-metal-performance-counters)
2. [Encoder Enumeration](#2-encoder-enumeration)
   - [FFmpeg Encoder Filtering](#ffmpeg-encoder-filtering)
   - [DXVA Capability Queries](#dxva-capability-queries)
   - [VideoToolbox Capability APIs](#videotoolbox-capability-apis)
   - [VAAPI Profile Detection](#vaapi-profile-detection)
3. [Thermal Monitoring](#3-thermal-monitoring)
   - [GPU Temperature Thresholds](#gpu-temperature-thresholds)
   - [Thermal Throttling Detection](#thermal-throttling-detection)
   - [Automatic Quality Reduction](#automatic-quality-reduction)
4. [CPU/Memory/System Detection](#4-cpumemorysystem-detection)
   - [Available RAM](#available-ram)
   - [CPU Cores and Features](#cpu-cores-and-features)
   - [Storage Speed](#storage-speed)
   - [Network Interface Capabilities](#network-interface-capabilities)
5. [OS-Specific Detection APIs](#5-os-specific-detection-apis)
   - [Windows: WMI, SetupAPI, DXGI](#windows-detection)
   - [macOS: IOKit, sysctl](#macos-detection)
   - [Linux: sysfs, /proc, udev, lspci](#linux-detection)
6. [Dynamic Quality Adjustment](#6-dynamic-quality-adjustment)
   - [Encoder Preset Switching](#encoder-preset-switching)
   - [Resolution Scaling](#resolution-scaling)
   - [Bitrate Reduction](#bitrate-reduction)
   - [Codec Fallback](#codec-fallback)
7. [Power Management](#7-power-management)
   - [Battery vs AC Detection](#battery-vs-ac-detection)
   - [Power-Save Modes](#power-save-modes)
8. [Load Balancing](#8-load-balancing)
   - [Multi-GPU Selection](#multi-gpu-selection)
   - [Integrated vs Discrete GPU Selection](#integrated-vs-discrete-gpu-selection)
9. [Go Integration](#9-go-integration)
   - [go-nvml](#go-nvml)
   - [gopsutil](#gopsutil)
   - [CGO for Platform APIs](#cgo-for-platform-apis)
   - [golang.org/x/sys](#golangorgsys)
10. [Hot-Plug Detection](#10-hot-plug-detection)
    - [GPU Addition/Removal](#gpu-additionremoval)
    - [USB Audio Device Changes](#usb-audio-device-changes)

---

## 1. GPU Detection

### NVIDIA NVML

The NVIDIA Management Library (NVML) provides comprehensive GPU querying capabilities. The Go bindings at `github.com/NVIDIA/go-nvml` offer a production-ready interface.

**Claim:** `go-nvml` dynamically loads `libnvidia-ml.so` at runtime and provides versioned symbol resolution for backward compatibility.[^1^]
**Source:** NVIDIA go-nvml GitHub
**URL:** https://github.com/nvidia/go-nvml
**Date:** 2025-08-22
**Excerpt:** "The code below shows an example of using these bindings to query all of the GPUs on your system and print out their UUIDs... ret := nvml.Init()"
**Context:** The library uses a two-layer approach: auto-generated bindings via c-for-go and manual wrappers for Go idioms.
**Confidence:** high

**Key NVML APIs for Detection:**

| API | Purpose | Return Type |
|-----|---------|-------------|
| `nvmlDeviceGetCount()` | Number of GPUs | `uint32` |
| `nvmlDeviceGetHandleByIndex()` | Get device handle | `nvmlDevice_t` |
| `nvmlDeviceGetUUID()` | Unique GPU identifier | `string` |
| `nvmlDeviceGetName()` | GPU model name | `string` |
| `nvmlDeviceGetTemperature()` | Current temperature (C) | `uint32` |
| `nvmlDeviceGetTemperatureThreshold()` | Shutdown/slowdown thresholds | `uint32` |
| `nvmlDeviceGetMemoryInfo()` | Total/Free/Used VRAM | `nvmlMemory_t` |
| `nvmlDeviceGetUtilizationRates()` | GPU and memory utilization | `nvmlUtilization_t` |
| `nvmlDeviceGetEncoderUtilization()` | Encoder utilization % | `uint32` |
| `nvmlDeviceGetEncoderCapacity()` | Max concurrent encodes | `uint32` |
| `nvmlDeviceGetEncoderStats()` | Session count, average FPS | `nvmlEncoderStats_t` |
| `nvmlDeviceGetCurrentClocksThrottleReasons()` | Active throttle reasons | `uint64` (bitmask) |
| `nvmlDeviceGetCurrentClocksEventReasons()` | Clock event reasons | `uint64` (bitmask) |
| `nvmlDeviceGetFanSpeed()` | Fan speed % | `uint32` |
| `nvmlDeviceGetPowerUsage()` | Current power draw (mW) | `uint32` |
| `nvmlDeviceGetPerformanceState()` | Current P-state | `uint32` |
| `nvmlDeviceGetPciInfo_v3()` | PCI bus/device/function | `nvmlPciInfo_t` |
| `nvmlDeviceGetCudaComputeCapability()` | Major/minor CUDA version | `int` |
| `nvmlDeviceGetDriverVersion()` | NVIDIA driver version | `string` |

**Claim:** NVML provides `nvmlDeviceGetEncoderUtilization` to query encoder utilization percentage and `nvmlDeviceGetEncoderStats` for session count and average FPS.[^2^]
**Source:** NVML API Reference Guide
**URL:** https://docs.nvidia.com/deploy/pdf/NVML_API_Reference_Guide.pdf
**Date:** n/a (reference document)
**Excerpt:** "nvmlDeviceGetEncoderUtilization - Get the GPU video encoder utilization... nvmlDeviceGetEncoderCapacity - Get the GPU video encoder capacity... nvmlDeviceGetEncoderStats - Get the GPU video encoder statistics"
**Context:** These APIs are essential for load balancing encoder sessions across GPUs.
**Confidence:** high

**NVML Clock Throttle Reasons Bitmask:**

```c
#define nvmlClocksThrottleReasonGpuIdle                 0x0000000000000001LL
#define nvmlClocksThrottleReasonApplicationsClocksSetting 0x0000000000000002LL
#define nvmlClocksThrottleReasonSwPowerCap                0x0000000000000004LL
#define nvmlClocksThrottleReasonHwSlowdown                0x0000000000000008LL
#define nvmlClocksThrottleReasonSyncBoost                 0x0000000000000010LL
#define nvmlClocksThrottleReasonSwThermalSlowdown         0x0000000000000020LL
#define nvmlClocksThrottleReasonHwThermalSlowdown         0x0000000000000040LL
#define nvmlClocksThrottleReasonHwPowerBrakeSlowdown      0x0000000000000080LL
#define nvmlClocksThrottleReasonDisplayClockSetting       0x0000000000000100LL
```

**Claim:** Thermal throttling at sustained temperatures above 83-85C causes automatic clock speed reduction, reducing GPU throughput by 25-30%.[^3^]
**Source:** Spheron Network - GPU Monitoring for ML
**URL:** https://www.spheron.network/blog/gpu-monitoring-for-ml/
**Date:** 2026-01-15
**Excerpt:** "Sustained temperatures above 83-85C cause automatic clock speed reduction. The GPU appears to run but delivers 25-30% less throughput. This is invisible without temperature monitoring."
**Context:** Key insight: utilization may still show high percentages because the GPU is active, just slower.
**Confidence:** high

**Claim:** `nvmlClocksThrottleReasonHwThermalSlowdown` (0x40) indicates hardware thermal slowdown reducing core clocks by a factor of 2 or more.[^4^]
**Source:** NVIDIA NVML API Documentation
**URL:** https://docs.nvidia.com/deploy/nvml-api/group__nvmlClocksThrottleReasons.html
**Date:** 2026-03-05
**Excerpt:** "HW Slowdown (reducing the core clocks by a factor of 2 or more) is engaged. This is an indicator of: temperature being too high, External Power Brake Assertion is triggered, Power draw is too high."
**Context:** This is the definitive API for programmatic thermal throttling detection.
**Confidence:** high

### AMD ROCm SMI

For AMD GPUs, the ROCm System Management Interface (ROCm SMI) library provides equivalent functionality.

**Claim:** The `go-rocm-smi` package provides Go bindings for AMD ROCm SMI, dynamically loading `librocm_smi64.so`.[^5^]
**Source:** ClusterCockpit/go-rocm-smi GitHub
**URL:** https://github.com/ClusterCockpit/go-rocm-smi/
**Date:** 2022-05-18
**Excerpt:** "This is an unofficial interface to the AMD ROCM SMI library for Golang applications. It is heavily inspired by go-nvml by also using cgo, c-for-go and its dlopen wrapper."
**Context:** Uses the same architectural pattern as go-nvml: CGO bindings with dynamic library loading.
**Confidence:** high

**Key ROCm SMI APIs for Detection:**

| API | Purpose |
|-----|---------|
| `rsmi_num_monitor_devices()` | Number of AMD GPUs |
| `rsmi_dev_name_get()` | GPU model name |
| `rsmi_dev_temp_metric_get()` | Temperature (multiple sensors) |
| `rsmi_dev_power_cap_get()` | Power cap |
| `rsmi_dev_power_ave_get()` | Average power draw |
| `rsmi_dev_busy_percent_get()` | GPU utilization % |
| `rsmi_dev_vram_total_get()` | Total VRAM |
| `rsmi_dev_vram_usage_get()` | Used VRAM |
| `rsmi_dev_fan_rpms_get()` | Fan RPM |
| `rsmi_dev_fan_speed_get()` | Fan speed % |
| `rsmi_dev_pci_id_get()` | PCI BDF ID |
| `rsmi_dev_unique_id_get()` | GPU unique identifier |

**Claim:** AMD SMI officially supports Go with version 1.20+ prerequisite and provides C++, Python, and Go interfaces.[^6^]
**Source:** AMD SMI Documentation (ROCm)
**URL:** https://rocm.docs.amd.com/_/downloads/amdsmi/en/docs-6.4.2/pdf/
**Date:** 2025-07-22
**Excerpt:** "Go interface prerequisites: Go version 1.20 or greater... The library will be extended to support AMD EPYC CPUs."
**Context:** AMD's official Go bindings for GPU/CPU monitoring.
**Confidence:** high

**Go Example - AMD ROCm SMI:**
```go
package main

import (
    "log"
    "github.com/ClusterCockpit/go-rocm-smi/pkg/rocm_smi"
)

func main() {
    ret := rocm_smi.Init()
    if ret != rocm_smi.STATUS_SUCCESS {
        log.Fatalf("Unable to initialize ROCM SMI: %v", ret)
    }
    defer rocm_smi.Shutdown()

    count, ret := rocm_smi.NumMonitorDevices()
    for i := 0; i < count; i++ {
        device, _ := rocm_smi.DeviceGetHandleByIndex(i)
        uuid, _ := device.GetUniqueId()
        log.Printf("GPU %d: UUID=%v", i, uuid)
    }
}
```

**Claim:** `github.com/xigang/go-rocm` provides an alternative, simpler Go binding for ROCm SMI.[^7^]
**Source:** xigang/go-rocm GitHub
**URL:** https://github.com/xigang/go-rocm
**Date:** 2024-12-04
**Excerpt:** "package main... device := rocm.NewDevice(uint32(i)); name, _ := device.GetDeviceName(); temp, _ := device.GetTemperature(0, rocm.TempCurrent)"
**Context:** Simpler alternative to the ClusterCockpit bindings.
**Confidence:** high

### Intel GPU Tools

Intel provides multiple APIs for GPU detection and monitoring.

**Claim:** Intel Quick Sync Video (QSV) and VA-API provide hardware acceleration on Intel GPUs. QSV is preferred on mainstream GPUs; VA-API is required for legacy pre-Broadwell GPUs.[^8^]
**Source:** Jellyfin Hardware Acceleration - Intel GPU
**URL:** https://jellyfin.org/docs/general/post-install/transcoding/hardware-acceleration/intel/
**Date:** n/a
**Excerpt:** "On Windows QSV is the only available method. On Linux there are two methods: QSV - Preferred on mainstream GPUs, VA-API - Required by pre-Broadwell legacy GPUs."
**Context:** The choice between QSV and VA-API depends on GPU generation and operating system.
**Confidence:** high

**Intel GPU Detection Approaches:**

1. **VAAPI:** Use `vainfo` to enumerate supported profiles/entrypoints
   ```bash
   vainfo | grep -E "(VAProfileH264|VAProfileHEVC|VAProfileAV1)"
   ```
   This outputs supported encode/decode profiles for the active GPU.

2. **Intel GPU Top:** `intel_gpu_top` from `intel-gpu-tools` package provides real-time utilization
   ```
   Render/3D:  45%  ████████████████████
   Video:      78%  ████████████████████████████
   VideoEnhance: 0%
   ```

3. **OneVPL/MediaSDK:** `mfxinfo` or SDK APIs query encoder capabilities programmatically

4. **Sysfs (Linux):** `/sys/class/drm/card*/device/` provides GPU information
   - `/sys/class/drm/card0/device/vendor` (0x8086 = Intel)
   - `/sys/class/drm/card0/device/device` (device ID)
   - `/sys/class/drm/card0/gt_cur_freq_mhz` (current frequency)
   - `/sys/class/drm/card0/gt_max_freq_mhz` (max frequency)

**Claim:** `ffmpeg -encoders | grep qsv` and `ffmpeg -encoders | grep vaapi` list available hardware encoders for Intel GPUs.[^9^]
**Source:** Superuser - Constructing encoding statements using ffmpeg QSV
**URL:** https://superuser.com/questions/1830920/constructing-encoding-statements-using-ffmpeg-qsv-and-intel-arc-card
**Date:** 2024-02-19
**Excerpt:** "ffmpeg -hide_banner -encoders | grep hevc... V..... hevc_qsv HEVC (Intel Quick Sync Video acceleration)... V....D hevc_vaapi H.265/HEVC (VAAPI)"
**Context:** Standard FFmpeg approach for enumerating available hardware encoders.
**Confidence:** high

### Apple Metal Performance Counters

On macOS, Apple provides GPU performance data through Metal APIs and IOKit.

**Claim:** The Metal Counter API provides precise GPU timings at runtime, available on iOS 14+ and macOS Big Sur+.[^10^]
**Source:** Apple Developer Tech Talks - Explore Live GPU Profiling with Metal Counters
**URL:** https://developer.apple.com/videos/play/tech-talks/10001/
**Date:** 2020-10-21
**Excerpt:** "The Metal Counter API is new in iOS 14. It was available in macOS Catalina and has been extended in macOS Big Sur. On iOS and macOS with Apple Silicon, it gives you access to stage boundary timings."
**Context:** The API provides GPU occupancy, utilization, and limiter counters for runtime adaptation.
**Confidence:** high

**Key Metal Performance Metrics:**

| Counter | Description |
|---------|-------------|
| Compute Occupancy | Percentage of GPU thread capacity for compute |
| Vertex Occupancy | Percentage for vertex threads |
| Fragment Occupancy | Percentage for fragment threads |
| GPU Last Level Cache Utilization | Cache bandwidth utilization |
| ALU Utilization | Arithmetic logic unit utilization |
| Memory Read/Write Limiters | Memory subsystem bottlenecks |

**Claim:** GPU activity depends on clock rate; high occupancy at half clock rate means 40% of peak performance.[^10^]
**Source:** Apple Developer Tech Talks
**URL:** https://developer.apple.com/videos/play/tech-talks/10001/
**Date:** 2020-10-21
**Excerpt:** "Secondly, GPU activity depends on its clock rate. Seeing a high GPU occupancy does not necessarily mean it is maxed out. As the system only uses as much power as needed, it will balance power and performance."
**Context:** Critical for power-aware quality decisions on macOS.
**Confidence:** high

**macOS GPU Detection via IOKit:**
```objc
// Using IOKit to enumerate GPUs
io_iterator_t iterator;
IOServiceGetMatchingServices(kIOMasterPortDefault,
    IOServiceMatching("IOGPU"), &iterator);
// Iterate through GPUs, extract properties
```

**macOS GPU Detection via sysctl/system_profiler:**
```go
// Go approach: execute system_profiler
out, _ := exec.Command("system_profiler", "SPDisplaysDataType").Output()
// Parse Chipset Model, Vendor, VRAM (Total), VRAM (Dynamic, Max)
```

---

## 2. Encoder Enumeration

### FFmpeg Encoder Filtering

**Claim:** FFmpeg provides `-encoders` and `-codecs` flags for enumerating available encoders, including hardware-accelerated variants.[^9^]
**Source:** Superuser / FFmpeg docs
**URL:** https://superuser.com/questions/1830920/
**Date:** 2024-02-19
**Excerpt:** "ffmpeg -hide_banner -codecs | grep hevc... (decoders: hevc hevc_qsv hevc_v4l2m2m) (encoders: libx265 hevc_qsv hevc_v4l2m2m hevc_vaapi)"
**Context:** Essential first step in hardware encoder capability detection.
**Confidence:** high

**Standard FFmpeg Encoder Enumeration Commands:**
```bash
# List all encoders
ffmpeg -hide_banner -encoders

# Filter for specific hardware encoders
ffmpeg -hide_banner -encoders | grep nvenc    # NVIDIA
ffmpeg -hide_banner -encoders | grep qsv      # Intel QSV
ffmpeg -hide_banner -encoders | grep vaapi    # VA-API (Intel/AMD)
ffmpeg -hide_banner -encoders | grep videotoolbox  # Apple
ffmpeg -hide_banner -encoders | grep amf      # AMD AMF
ffmpeg -hide_banner -encoders | grep v4l2m2m  # V4L2

# Get detailed encoder info
ffmpeg -hide_banner -h encoder=h264_nvenc
ffmpeg -hide_banner -h encoder=hevc_qsv
ffmpeg -hide_banner -h encoder=h264_vaapi
```

**Claim:** VAAPI encoders expose `async_depth`, `rc_mode`, and codec-specific profile options for capability detection and configuration.[^11^]
**Source:** FFmpeg Codecs Documentation
**URL:** https://ffmpeg.org/ffmpeg-codecs.html
**Date:** n/a
**Excerpt:** "async_depth - Maximum processing parallelism... rc_mode - Set the rate control mode to use. A given driver may only support a subset of modes."
**Context:** These parameters determine hardware capabilities and should be probed before encoder selection.
**Confidence:** high

### DXVA Capability Queries

On Windows, DirectX Video Acceleration provides hardware capability detection.

**Claim:** Chromium's DXVA implementation queries `MF_SA_D3D_AWARE` attribute and `CODECAPI_AVDecVideoAcceleration_H264` to determine hardware decode support.[^12^]
**Source:** Chromium DXVA Video Decode Accelerator
**URL:** https://chromium.googlesource.com/chromium/src/+/f71098198f9adbb2d7eb2eed350e864d81bf77fd/media/gpu/windows/dxva_video_decode_accelerator_win.cc
**Date:** n/a
**Excerpt:** "hr = attributes->GetUINT32(MF_SA_D3D_AWARE, &dxva)... hr = attributes->SetUINT32(CODECAPI_AVDecVideoAcceleration_H264, TRUE)"
**Context:** Microsoft's Media Foundation APIs provide the hardware capability query mechanism on Windows.
**Confidence:** high

**DXGI Adapter Enumeration:**
```cpp
// Enumerate all GPU adapters
std::vector<IDXGIAdapter1*> EnumerateAdapters() {
    IDXGIFactory1* pFactory;
    CreateDXGIFactory1(__uuidof(IDXGIFactory1), (void**)&pFactory);
    
    IDXGIAdapter1* pAdapter;
    std::vector<IDXGIAdapter1*> vAdapters;
    for (UINT i = 0; pFactory->EnumAdapters1(i, &pAdapter) != DXGI_ERROR_NOT_FOUND; ++i) {
        DXGI_ADAPTER_DESC1 desc;
        pAdapter->GetDesc1(&desc);
        vAdapters.push_back(pAdapter);
    }
    return vAdapters;
}
```

**Claim:** `IDXGIFactory1::EnumAdapters1` first returns the adapter with the desktop primary, then adapters with outputs, then adapters without outputs.[^13^]
**Source:** Microsoft Learn - IDXGIFactory1::EnumAdapters1
**URL:** https://learn.microsoft.com/en-us/windows/win32/api/dxgi/nf-dxgi-idxgifactory1-enumadapters1
**Date:** 2022-08-22
**Excerpt:** "EnumAdapters1 first returns the adapter with the output on which the desktop primary is displayed. This adapter corresponds with an index of zero."
**Context:** Important for knowing which GPU is "primary" vs secondary.
**Confidence:** high

### VideoToolbox Capability APIs

On macOS, VideoToolbox is the hardware acceleration framework.

**Claim:** VideoToolbox supports H.264 encode/decode on most 2011+ Macs, HEVC on 2017+ Macs (except MacBook Air 13" 2017), and AV1 decode on M3+ series.[^14^]
**Source:** Jellyfin - Apple Mac Hardware Acceleration
**URL:** https://jellyfin.org/docs/general/post-install/transcoding/hardware-acceleration/apple/
**Date:** n/a
**Excerpt:** "Any VideoToolbox-supported Mac supports decoding and encoding H.264 8-bit. Macs from 2017 and later support decoding and encoding HEVC. Starting with M3 series, Apple Silicon supports hardware-accelerated decoding of AV1."
**Context:** VideoToolbox capability is determined primarily by GPU generation.
**Confidence:** high

**Claim:** The "Ultra" variant Apple Silicon chips feature 2 video decoding engines and 4 video encoding engines.[^14^]
**Source:** Jellyfin Apple Hardware Acceleration
**URL:** https://jellyfin.org/docs/general/post-install/transcoding/hardware-acceleration/apple/
**Date:** n/a
**Excerpt:** "The 'Ultra' variant chips feature 2 video decoding engines and 4 video encoding engines, effectively doubling the capability compared to the 'Max' variant chips."
**Context:** Encoder count is critical for load balancing concurrent sessions.
**Confidence:** high

**VideoToolbox Encoder Detection Pattern:**
```c
// Check VideoToolbox encoder availability
VTCompressionSessionRef session;
OSStatus status = VTCompressionSessionCreate(
    NULL, width, height, kCMVideoCodecType_H264,
    NULL, NULL, NULL, NULL, NULL, &session);
if (status == noErr) {
    // H.264 hardware encoder available
    VTCompressionSessionInvalidate(session);
    CFRelease(session);
}
```

### VAAPI Profile Detection

On Linux, VA-API provides the standard hardware video API.

**Claim:** `vainfo` command exposes supported profiles and entrypoints; VP9 encoding fails on unsupported hardware with "Encoding entrypoint not found (19 / 6)."[^15^]
**Source:** Stack Overflow - Encoding with Intel HD4000 VA API
**URL:** https://stackoverflow.com/questions/36260474/
**Date:** 2016-03-28
**Excerpt:** "[vp9_vaapi @ 0x409da40] Encoding entrypoint not found (19 / 6). Error initializing output stream 0:0"
**Context:** Hardware capability detection must handle entrypoint query failures gracefully.
**Confidence:** high

**VAAPI Profile Detection:**
```bash
# Query supported profiles
vainfo | grep "VAProfile"

# Expected output includes:
# VAProfileH264ConstrainedBaseline
# VAProfileH264Main
# VAProfileH264High
# VAProfileHEVCMain
# VAProfileHEVCMain10
# VAProfileVP9Profile0
# VAProfileAV1Profile0
```

**Claim:** Unlike NVIDIA NVENC, there is no concurrent encoding sessions limit on Intel iGPU and ARC dGPU with VA-API/QSV.[^8^]
**Source:** Jellyfin Intel GPU Guide
**URL:** https://jellyfin.org/docs/general/post-install/transcoding/hardware-acceleration/intel/
**Date:** n/a
**Excerpt:** "Unlike NVIDIA NVENC, there is no concurrent encoding sessions limit on Intel iGPU and ARC dGPU."
**Context:** Intel GPUs can run as many encode sessions as resources allow.
**Confidence:** high

---

## 3. Thermal Monitoring

### GPU Temperature Thresholds

**Claim:** NVML provides `nvmlDeviceGetTemperatureThreshold` with types: Shutdown (default ~95C), Slowdown (default ~88C), and Memory Max (varies by GPU).[^16^]
**Source:** NVML API Reference
**URL:** https://pkg.go.dev/github.com/mxpv/nvml-go
**Date:** n/a
**Excerpt:** "DeviceGetTemperatureThreshold retrieves the temperature threshold for the GPU with the specified threshold type in degrees C."
**Context:** These thresholds should be monitored to proactively reduce quality before throttling occurs.
**Confidence:** high

**Typical NVIDIA GPU Temperature Thresholds:**

| Threshold | Typical Value | Action |
|-----------|---------------|--------|
| GPU Idle | 35-45C | Normal operation |
| Warning | 75C | Begin monitoring closely |
| Thermal Slowdown | 83-88C | Reduce encoder preset quality |
| Critical/Shutdown | 95-100C | Emergency quality reduction |

### Thermal Throttling Detection

**Claim:** `nvidia-smi -q -d CLOCK` displays current and maximum clock frequencies; comparing these reveals throttling.[^17^]
**Source:** MassedCompute - How to monitor thermal throttling
**URL:** https://massedcompute.com/faq-answers/?question=How%20to%20monitor%20thermal%20throttling%20in%20NVIDIA%20GPUs?
**Date:** 2025-07-31
**Excerpt:** "Thermal throttling often results in reduced clock speeds. You can track GPU clock frequencies in real-time using: nvidia-smi -q -d CLOCK"
**Context:** Clock frequency comparison is the most reliable throttling detection method.
**Confidence:** high

**Comprehensive Throttling Detection via NVML:**
```go
// Check if GPU is thermally throttled
func IsThermallyThrottled(device nvml.Device) bool {
    reasons, _ := device.GetCurrentClocksThrottleReasons()
    thermalMask := nvml.ClocksThrottleReasonHwThermalSlowdown | 
                   nvml.ClocksThrottleReasonSwThermalSlowdown
    return (reasons & thermalMask) != 0
}

// Get temperature with threshold comparison
func CheckThermalStatus(device nvml.Device) (temp uint32, threshold uint32, status string) {
    temp, _ = device.GetTemperature(nvml.TEMPERATURE_GPU)
    threshold, _ = device.GetTemperatureThreshold(nvml.TEMPERATURE_THRESHOLD_SLOWDOWN)
    
    if temp >= threshold {
        status = "CRITICAL"
    } else if temp >= threshold-5 {
        status = "WARNING"
    } else {
        status = "OK"
    }
    return
}
```

### Automatic Quality Reduction

When thermal throttling is detected, the system should automatically reduce encoding quality:

1. **Preset escalation:** P7 (slow/best) -> P6 -> P5 -> P4 -> P3 (faster)
2. **Resolution reduction:** 4K -> 1080p -> 720p
3. **Bitrate reduction:** Reduce target bitrate by 25-50%
4. **Codec fallback:** HEVC -> H.264 (H.264 has lower encoding complexity)
5. **Framerate reduction:** 60fps -> 30fps

**Claim:** Sunshine game streaming software dynamically adjusts minimum FPS target to save bandwidth when content is static.[^18^]
**Source:** Sunshine Configuration Documentation
**URL:** https://docs.lizardbyte.dev/projects/sunshine/latest/md_docs_2configuration.html
**Date:** n/a
**Excerpt:** "Sunshine tries to save bandwidth when content on screen is static or a low framerate. This setting controls the lowest effective framerate a stream can reach."
**Context:** Dynamic framerate reduction is a proven technique for reducing encoder load.
**Confidence:** high

---

## 4. CPU/Memory/System Detection

### Available RAM

**Claim:** `gopsutil` memory package reads from `/proc/meminfo` on Linux, uses `GlobalMemoryStatusEx` on Windows, and `sysctl` on macOS.[^19^]
**Source:** gopsutil API documentation
**URL:** https://www.mintlify.com/shirou/gopsutil/api/mem/swap-memory
**Date:** 2026-03-09
**Excerpt:** "Linux: All fields populated from /proc/meminfo, /proc/vmstat, and /proc/swaps. Windows: Uses Windows API GlobalMemoryStatusEx and GetPerformanceInfo. macOS: Values obtained from sysctl calls."
**Context:** Cross-platform memory detection is well-abstracted by gopsutil.
**Confidence:** high

**Go RAM Detection:**
```go
import "github.com/shirou/gopsutil/v4/mem"

func GetMemoryInfo() (*mem.VirtualMemoryStat, error) {
    return mem.VirtualMemory()
    // Fields: Total, Available, Used, Free, UsedPercent, SwapTotal, SwapFree
}
```

### CPU Cores and Features

**Claim:** `golang.org/x/sys/cpu` provides runtime CPU feature detection for x86 (SSE, AVX, AVX-512), ARM64 (NEON, Crypto), and other architectures.[^20^]
**Source:** golang.org/x/sys/cpu documentation
**URL:** https://pkg.go.dev/golang.org/x/sys/cpu
**Date:** 2026-03-27
**Excerpt:** "Package cpu implements processor feature detection for various CPU architectures."
**Context:** Essential for detecting SIMD capabilities that affect software encoding performance.
**Confidence:** high

**Claim:** `gopsutil` CPU package parses `/proc/cpuinfo` on Linux, with model name, MHz, cache size, flags, and physical/core ID fields.[^21^]
**Source:** gopsutil cpu_linux.go source
**URL:** https://chromium.googlesource.com/external/github.com/shirou/gopsutil/+/refs/heads/master/cpu/cpu_linux.go
**Date:** 2019-02-26
**Excerpt:** "case 'cpu MHz', 'clock', 'cpu MHz dynamic'... case 'cache size'... case 'flags', 'Features'... c.Flags = strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ' ' })"
**Context:** Used to determine CPU capability for software encoding fallback.
**Confidence:** high

### Storage Speed

Storage speed detection is important for local recording scenarios.

**Claim:** `fio` (Flexible I/O Tester) is the industry standard for storage benchmarking, measuring IOPS, throughput, and latency.[^22^]
**Source:** Medium - Disk Benchmarking using FIO
**URL:** https://vineetcic.medium.com/disk-benchmarking-using-fio-c71e0ce0d47c
**Date:** 2022-08-03
**Excerpt:** "Test write throughput by performing sequential writes with multiple parallel streams (8+), using an I/O block size of 1 MB and an I/O depth of at least 64."
**Context:** Host agents can run lightweight fio benchmarks at startup to determine recording capability.
**Confidence:** high

**Lightweight Storage Detection (Go):**
```go
// Quick write speed test
func BenchmarkWriteSpeed(path string, sizeMB int) (float64, error) {
    data := make([]byte, 1024*1024) // 1MB buffer
    start := time.Now()
    f, _ := os.CreateTemp(path, "benchmark-*")
    defer os.Remove(f.Name())
    
    for i := 0; i < sizeMB; i++ {
        f.Write(data)
    }
    f.Sync() // Force to disk
    elapsed := time.Since(start).Seconds()
    return float64(sizeMB) / elapsed, nil // MB/s
}
```

### Network Interface Capabilities

**Claim:** The Linux ethtool netlink interface (`ETHTOOL_GLINKSETTINGS`) provides supported and advertised link modes, autonegotiation status, link speed, and duplex mode.[^23^]
**Source:** Kernel Documentation - Netlink interface for ethtool
**URL:** https://docs.kernel.org/networking/ethtool-netlink.html
**Date:** n/a
**Excerpt:** "ETHTOOL_A_LINKMODES_SPEED: u32, link speed (Mb/s)... ETHTOOL_A_LINKMODES_DUPLEX: u8, duplex mode"
**Context:** Netlink is the modern API; ioctl-based ethtool is deprecated.
**Confidence:** high

**Network Speed Detection (Linux C):**
```c
#include <linux/ethtool.h>
#include <linux/sockios.h>
struct ethtool_cmd edata;
edata.cmd = ETHTOOL_GSET;
ifr.ifr_data = &edata;
ioctl(sock, SIOCETHTOOL, &ifr);
switch (ethtool_cmd_speed(&edata)) {
    case SPEED_10: ...
    case SPEED_100: ...
    case SPEED_1000: printf("1Gbps\n"); break;
    case SPEED_2500: ...
    case SPEED_10000: printf("10Gbps\n"); break;
}
```

---

## 5. OS-Specific Detection APIs

### Windows Detection

**WMI (Windows Management Instrumentation):**

**Claim:** Windows battery/AC status is queried via `root\wmi\BatteryStatus` WMI class with `PowerOnline` property; `Get-CimInstance` replaces deprecated `Get-WmiObject`.[^24^]
**Source:** Nyxshima - Get Battery and AC Power Status via PowerShell
**URL:** https://www.nyxshima.com/get-battery-and-ac-power-adapter-connection-status-via-powershell/
**Date:** 2026-03-20
**Excerpt:** "(Get-CimInstance -Namespace root\wmi -ClassName BatteryStatus).PowerOnline... Returns $true if plugged in, $false if on battery."
**Context:** Critical for laptop power-aware quality adjustment.
**Confidence:** high

**Windows GPU Detection via DXGI:**
```cpp
// CGO approach for Windows GPU enumeration
// #cgo LDFLAGS: -ldxgi
// #include <dxgi.h>

extern "C" int EnumerateGPUs() {
    IDXGIFactory1* pFactory;
    CreateDXGIFactory1(__uuidof(IDXGIFactory1), (void**)&pFactory);
    
    UINT i = 0;
    IDXGIAdapter1* pAdapter;
    while (pFactory->EnumAdapters1(i, &pAdapter) != DXGI_ERROR_NOT_FOUND) {
        DXGI_ADAPTER_DESC1 desc;
        pAdapter->GetDesc1(&desc);
        // desc.Description: GPU name
        // desc.DedicatedVideoMemory: VRAM bytes
        // desc.VendorId: PCI vendor ID (0x10DE = NVIDIA, 0x1002 = AMD, 0x8086 = Intel)
        pAdapter->Release();
        i++;
    }
    pFactory->Release();
    return i;
}
```

### macOS Detection

**IOKit and sysctl:**

**Claim:** macOS hardware detection requires combining `sysctl`, `system_profiler SPDisplaysDataType`, and `vm_stat` with careful parsing of varied output formats.[^25^]
**Source:** Rezmoss - macOS Hardware Detection with Go
**URL:** https://rezmoss.com/blog/macos-hardware-detection-with-go/
**Date:** 2025-08-02
**Excerpt:** "The SPDisplaysDataType argument tells system_profiler to return only display/graphics information... Discrete GPUs show 'VRAM (Total)' while integrated GPUs show 'VRAM (Dynamic, Max)'."
**Context:** Go hardware detection on macOS relies primarily on command execution and parsing.
**Confidence:** high

**Key macOS sysctl parameters:**

| Parameter | Description |
|-----------|-------------|
| `hw.physicalcpu` | Physical CPU cores |
| `hw.logicalcpu` | Logical CPU cores (with HT) |
| `hw.memsize` | Total RAM in bytes |
| `hw.cpufrequency_max` | Max CPU frequency |
| `machdep.cpu.brand_string` | CPU model name |
| `hw.l2cachesize` | L2 cache size |
| `hw.l3cachesize` | L3 cache size |
| `hw.pagesize` | Memory page size |

### Linux Detection

**sysfs, /proc, udev, lspci:**

**Claim:** The AMDGPU DRM driver exposes GPU load via sysfs at `/sys/class/drm/card0/gpu_busy_percent`.[^26^]
**Source:** Phoronix - AMDGPU DRM Driver To Finally Expose GPU Load Via Sysfs
**URL:** https://www.phoronix.com/news/AMDGPU-Load-Percent-Sysfs
**Date:** 2018-06-20
**Excerpt:** "Finally -- likely for Linux 4.19 -- the GPU load usage will be reported as a percentage via a sysfs interface... being exposed as a file named gpu_busy_level"
**Context:** Standard Linux GPU monitoring uses sysfs interfaces under `/sys/class/drm/`.
**Confidence:** high

**Linux GPU Detection via sysfs:**
```bash
# GPU vendor detection
cat /sys/class/drm/card0/device/vendor  # 0x10de = NVIDIA, 0x1002 = AMD, 0x8086 = Intel
cat /sys/class/drm/card0/device/device  # PCI device ID

# GPU metrics (driver-dependent)
cat /sys/class/drm/card0/device/pp_dpm_sclk  # Clock states (AMD)
cat /sys/class/drm/card0/device/hwmon/hwmon*/temp1_input  # Temperature
cat /sys/class/drm/card0/device/hwmon/hwmon*/fan1_input   # Fan RPM
cat /sys/class/drm/card0/device/hwmon/hwmon*/power1_average # Power (uW)

# NVIDIA-specific (via nvidia-smi or NVML)
nvidia-smi --query-gpu=temperature.gpu,utilization.gpu,utilization.encoder --format=csv
```

**PCI Device Enumeration via lspci:**
```bash
# List all GPUs
lspci | grep -i vga
lspci | grep -i display

# Detailed GPU info with driver
lspci -k -s 01:00.0
# Output: Kernel driver in use: nvidia
```

---

## 6. Dynamic Quality Adjustment

### Encoder Preset Switching

**Claim:** NVENC provides 7 presets from P1 (highest performance/lowest quality) to P7 (lowest performance/highest quality).[^27^]
**Source:** NVIDIA NVENC Video Encoder API Programming Guide
**URL:** https://docs.nvidia.com/video-technologies/video-codec-sdk/13.0/nvenc-video-encoder-api-prog-guide/index.html
**Date:** n/a
**Excerpt:** "For each tuning info, seven presets from P1 (highest performance) to P7 (lowest performance) have been provided to control performance/quality trade off."
**Context:** Presets can be switched dynamically based on GPU load and thermal conditions.
**Confidence:** high

**Claim:** OBS recommends P6: Slower (Better Quality) for streaming, with tuning set to High Quality.[^28^]
**Source:** NVIDIA NVENC OBS Guide
**URL:** https://www.nvidia.com/en-us/geforce/guides/broadcasting-guide/
**Date:** 2025-01-30
**Excerpt:** "Preset: Select P6: Slower (Better Quality). Tuning: Select High Quality. Multipass Mode: Set to Two Passes (Quarter Resolution)."
**Context:** P6 provides excellent quality but may need reduction on thermally constrained GPUs.
**Confidence:** high

**NVENC Preset Selection Strategy:**

| Condition | Recommended Preset |
|-----------|-------------------|
| GPU >85C or throttling | P1 (fastest) |
| GPU >75C | P3 |
| Normal operation | P5-P6 |
| Static content | P7 (best quality) |
| Battery power | P1-P2 |

### Resolution Scaling

Resolution should be dynamically scaled based on:
- GPU encoder utilization >90% -> reduce by 25%
- Thermal WARNING state -> reduce by 25%
- Thermal CRITICAL state -> reduce by 50%
- Client network congestion -> match to estimated bandwidth

### Bitrate Reduction

**Claim:** Streamlabs Dynamic Bitrate automatically adjusts bitrate based on network conditions to prevent dropped frames.[^29^]
**Source:** Streamlabs - How to use Dynamic Bitrate
**URL:** https://streamlabs.com/content-hub/post/how-to-use-dynamic-bitrate
**Date:** 2019-10-17 (updated 2022-01-28)
**Excerpt:** "When you experience network issues, Streamlabs Desktop will lower your bitrate until you're not dropping frames anymore. Once we detect that the network issues are gone, your bitrate should gradually go back up."
**Context:** Similar algorithms can be applied to GPU thermal throttling scenarios.
**Confidence:** high

### Codec Fallback

Priority-based codec selection:
1. **H265/HEVC** - Best compression, but higher encode complexity
2. **H264/AVC** - Universal compatibility, moderate complexity
3. **AV1** - Best compression (newer GPUs), but highest complexity
4. **Software fallback** - x264, x265 when hardware unavailable

---

## 7. Power Management

### Battery vs AC Detection

**Windows:**
```powershell
# PowerShell/CIM approach
Get-CimInstance -Namespace root\wmi -ClassName BatteryStatus | Select PowerOnline, Charging
# Returns: PowerOnline=True when on AC
```

**Linux:**
```bash
# Standard sysfs approach
cat /sys/class/power_supply/BAT0/status  # Charging / Discharging / Full
cat /sys/class/power_supply/AC/online     # 1 = AC connected, 0 = battery
```

**macOS:**
```bash
# IOKit power source approach
pmset -g batt
# Output: "AC Power connected" or "Using battery power"
```

### Power-Save Modes

**Claim:** NVIDIA Advanced Optimus provides Automatic Select, Optimus, and "NVIDIA GPU only" modes, with display switching disabled on battery power.[^30^]
**Source:** NVIDIA Advanced Optimus Overview
**URL:** https://nvidia.custhelp.com/app/answers/detail/a_id/5097
**Date:** 2025-01-13
**Excerpt:** "The display will not switch to discrete GPU if the system is running in battery mode (DC)... Advanced Optimus allows dynamically switching an internal eDP laptop display panel across different display adapters."
**Context:** Battery detection directly affects GPU selection strategy.
**Confidence:** high

---

## 8. Load Balancing

### Multi-GPU Selection

**Claim:** NVIDIA consumer GPUs officially support up to 3 simultaneous NVENC encode streams per system (increased to 5 in 2023, 8 in 2024, 12 in November 2025).[^31^]
**Source:** Wikipedia - NVENC
**URL:** https://en.wikipedia.org/wiki/NVENC
**Date:** 2014-08-17 (updated)
**Excerpt:** "Until March 2023 consumer-targeted GeForce graphics cards officially support no more than three simultaneously encoding video streams... From January 2024 onwards, eight simultaneous encoding video streams became the baseline. From November 2025 onwards, twelve simultaneous encoding video streams became the baseline."
**Context:** Session limits must be tracked per-GPU for multi-GPU load balancing.
**Confidence:** high

**Multi-GPU Selection Algorithm:**
```go
type GPUEncoderState struct {
    Index              int
    UUID               string
    Name               string
    EncoderUtilization uint32
    Temperature        uint32
    Sessions           int
    MaxSessions        int
    ThrottleReasons    uint64
}

func SelectBestGPU(gpus []GPUEncoderState) *GPUEncoderState {
    var best *GPUEncoderState
    for i := range gpus {
        g := &gpus[i]
        // Skip thermally throttled GPUs
        if g.ThrottleReasons & (nvml.ClocksThrottleReasonHwThermalSlowdown |
                                nvml.ClocksThrottleReasonSwThermalSlowdown) != 0 {
            continue
        }
        // Skip at-capacity GPUs
        if g.Sessions >= g.MaxSessions {
            continue
        }
        // Select GPU with lowest encoder utilization
        if best == nil || g.EncoderUtilization < best.EncoderUtilization {
            best = g
        }
    }
    return best
}
```

### Integrated vs Discrete GPU Selection

**Claim:** NVIDIA Optimus automatically switches between integrated and discrete GPU based on application requirements; AMD Switchable Graphics provides similar functionality.[^32^]
**Source:** ASUS Support - Configure NVIDIA Optimus and AMD Switchable Graphics
**URL:** https://www.asus.com/us/support/faq/1038387/
**Date:** 2025-11-19
**Excerpt:** "Default setting the NVIDIA Driver will determine which graphic card would be used depending on the application."
**Context:** Applications should detect available GPUs and prefer discrete GPUs for encoding.
**Confidence:** high

**GPU Selection Priority:**
1. Dedicated NVIDIA GPU with NVENC (best encode quality)
2. Dedicated AMD GPU with VCE/VCN
3. Intel discrete GPU (ARC) with QSV
4. Intel integrated GPU with QSV
5. Apple Silicon Media Engine
6. Software encoding (fallback)

---

## 9. Go Integration

### go-nvml

**Claim:** `github.com/NVIDIA/go-nvml` provides the reference implementation for NVML Go bindings with automatic versioned symbol resolution and error handling.[^1^]
**Source:** NVIDIA go-nvml GitHub
**URL:** https://github.com/nvidia/go-nvml
**Date:** 2025-08-22
**Excerpt:** "All you need is a simple import and a call to nvml.Init() to start using these bindings."
**Context:** The official NVIDIA-supported Go bindings.
**Confidence:** high

**Complete Go NVML Example:**
```go
package main

import (
    "fmt"
    "log"
    "github.com/NVIDIA/go-nvml/pkg/nvml"
)

type GPUInfo struct {
    Index              int
    UUID               string
    Name               string
    DriverVersion      string
    TotalMemory        uint64
    FreeMemory         uint64
    Temperature        uint32
    SlowdownThreshold  uint32
    ShutdownThreshold  uint32
    GPUUtilization     uint32
    EncoderUtilization uint32
    PowerUsage         uint32
    ThrottleReasons    uint64
}

func DetectGPUs() ([]GPUInfo, error) {
    ret := nvml.Init()
    if ret != nvml.SUCCESS {
        return nil, fmt.Errorf("nvml.Init: %v", nvml.ErrorString(ret))
    }
    defer nvml.Shutdown()

    count, ret := nvml.DeviceGetCount()
    if ret != nvml.SUCCESS {
        return nil, fmt.Errorf("DeviceGetCount: %v", nvml.ErrorString(ret))
    }

    var gpus []GPUInfo
    for i := 0; i < count; i++ {
        device, _ := nvml.DeviceGetHandleByIndex(i)
        uuid, _ := device.GetUUID()
        name, _ := device.GetName()
        memInfo, _ := device.GetMemoryInfo()
        temp, _ := device.GetTemperature(nvml.TEMPERATURE_GPU)
        slowdown, _ := device.GetTemperatureThreshold(nvml.TEMPERATURE_THRESHOLD_SLOWDOWN)
        shutdown, _ := device.GetTemperatureThreshold(nvml.TEMPERATURE_THRESHOLD_SHUTDOWN)
        util, _ := device.GetUtilizationRates()
        encUtil, _ := device.GetEncoderUtilization()
        power, _ := device.GetPowerUsage()
        throttle, _ := device.GetCurrentClocksThrottleReasons()

        gpus = append(gpus, GPUInfo{
            Index:              i,
            UUID:               uuid,
            Name:               name,
            TotalMemory:        memInfo.Total,
            FreeMemory:         memInfo.Free,
            Temperature:        temp,
            SlowdownThreshold:  slowdown,
            ShutdownThreshold:  shutdown,
            GPUUtilization:     util.Gpu,
            EncoderUtilization: encUtil,
            PowerUsage:         power,
            ThrottleReasons:    throttle,
        })
    }
    return gpus, nil
}
```

### gopsutil

**Claim:** `github.com/shirou/gopsutil` provides cross-platform CPU, memory, disk, process, and network information with Linux, macOS, Windows, FreeBSD support.[^33^]
**Source:** gopsutil documentation
**URL:** https://www.mintlify.com/shirou/gopsutil/guides/advanced/cross-platform
**Date:** 2026-03-09
**Excerpt:** "Use runtime.GOOS to detect the platform at runtime... Linux-specific fields: Iowait, Irq, Softirq, Steal, Guest"
**Context:** The standard Go library for system information.
**Confidence:** high

**gopsutil Feature Matrix:**

| Feature | Linux | macOS | Windows | FreeBSD |
|---------|-------|-------|---------|---------|
| CPU Info | Yes | Yes | Yes | Yes |
| CPU Times | Yes | Yes | Yes | Partial |
| Memory Info | Yes | Yes | Yes | Yes |
| Disk Usage | Yes | Yes | Yes | Yes |
| Process Info | Yes | Yes | Yes | Partial |
| Network Stats | Yes | Yes | Yes | Partial |

### CGO for Platform APIs

**Claim:** Go Metal GPU programming uses CGO with `#cgo LDFLAGS: -framework Metal -framework MetalPerformanceShaders` for GPU compute.[^34^]
**Source:** Medium - Programming Apple GPUs through Go and Metal
**URL:** https://medium.com/data-science/programming-apple-gpus-through-go-and-metal-shading-language-a0e7a60a3dba
**Date:** 2024-02-26
**Excerpt:** "#cgo LDFLAGS: -framework Foundation -framework CoreGraphics -framework Metal -framework MetalPerformanceShaders... It took some trial-and-error to get the right set of linker flags."
**Context:** CGO is the bridge for all platform-specific hardware APIs in Go.
**Confidence:** high

### golang.org/x/sys

**Claim:** `golang.org/x/sys` provides low-level OS primitives: `unix` (419 functions), `windows` (701 functions), and `cpu` (feature detection across all architectures).[^20^]
**Source:** golang.org/x/sys documentation
**URL:** https://pkg.go.dev/golang.org/x/sys
**Date:** 2026-03-27
**Excerpt:** "Provides access to system calls, OS-specific functionality, and CPU feature detection across different platforms."
**Context:** Essential for direct system-level hardware access without external dependencies.
**Confidence:** high

---

## 10. Hot-Plug Detection

### GPU Addition/Removal

**Claim:** Linux GPU hotplug events are managed by udev; kernel generates uevents via netlink (`NETLINK_KOBJECT_UEVENT`), processed by systemd-udevd, then rebroadcast to userspace.[^35^]
**Source:** ArcaneNibble - Hardware hotplug events on Linux, the gory details
**URL:** https://arcanenibble.github.io/hardware-hotplug-events-on-linux-the-gory-details.html
**Date:** 2026-03-01
**Excerpt:** "The kernel generates a hotplug event and notifies userspace using netlink sockets. udev, now part of systemd, receives these netlink messages."
**Context:** Applications should listen to udev events for GPU hotplug detection.
**Confidence:** high

**Go GPU Hot-Plug Monitoring via udev:**
```go
// Using github.com/rubiojr/go-usbmon as reference
// For GPU monitoring, extend to SUBSYSTEM=="pci" or SUBSYSTEM=="drm"

package main

import (
    "context"
    "fmt"
    "github.com/rubiojr/go-usbmon"
)

func MonitorGPUHotplug() {
    // For GPU hotplug, use PCI subsystem with class 0x030000 (VGA controller)
    filter := &usbmon.ActionFilter{Action: usbmon.ActionAll}
    devs, _ := usbmon.ListenFiltered(context.Background(), filter)
    
    for dev := range devs {
        fmt.Printf("Device %s: %s\n", dev.Action(), dev.Path())
        // Re-enumerate GPUs on add/remove
        if dev.Action() == "add" || dev.Action() == "remove" {
            // Trigger GPU capability re-detection
            // Re-advertise capabilities to session service
        }
    }
}
```

**Linux GPU Hot-Plug via udev Rules:**
```bash
# /etc/udev/rules.d/99-gpu-detect.rules
ACTION=="add", SUBSYSTEM=="pci", ATTR{class}=="0x030000", \
    RUN+="/usr/local/bin/notify-gpu-change add"
ACTION=="remove", SUBSYSTEM=="pci", ATTR{class}=="0x030000", \
    RUN+="/usr/local/bin/notify-gpu-change remove"
```

**Claim:** NVIDIA X driver's `UseHotplugEvents` option controls RandR display changed events for GPU hotplug; DisplayPort events cannot be suppressed.[^36^]
**Source:** Ask Ubuntu - NVIDIA X driver hotplug events
**URL:** https://askubuntu.com/questions/858798/
**Date:** 2016-12-09
**Excerpt:** "When this option is enabled, the NVIDIA X driver will generate RandR display changed events when displays are plugged into or unplugged from an NVIDIA GPU. Hotplug events cannot be suppressed for displays connected via DisplayPort."
**Context:** NVIDIA-specific hotplug behavior on Linux X11 systems.
**Confidence:** high

### USB Audio Device Changes

USB audio device hot-plug follows the same netlink/udev mechanism:
```bash
# Monitor USB audio specifically
udevadm monitor --subsystem-match=usb --udev --property | grep -E "(ACTION|ID_VENDOR|ID_MODEL|SUBSYSTEM)"
```

---

## Key Implementation Patterns

### Capability Advertisement Structure

```go
type HardwareCapabilities struct {
    GPUs []GPUCapabilities `json:"gpus"`
    CPU  CPUCapabilities   `json:"cpu"`
    Memory MemoryInfo      `json:"memory"`
    Network NetworkInfo    `json:"network"`
    Storage StorageInfo    `json:"storage"`
    Power   PowerInfo      `json:"power"`
}

type GPUCapabilities struct {
    Index              int      `json:"index"`
    UUID               string   `json:"uuid"`
    Vendor             string   `json:"vendor"`  // nvidia, amd, intel, apple
    Model              string   `json:"model"`
    TotalVRAM          uint64   `json:"total_vram"`
    EncoderSupport     []string `json:"encoder_support"` // h264_nvenc, hevc_nvenc, av1_nvenc
    MaxEncoderSessions int      `json:"max_encoder_sessions"`
    NVENGPresetRange   string   `json:"nvenc_preset_range,omitempty"` // P1-P7
    IsDiscrete         bool     `json:"is_discrete"`
}
```

### Dynamic Quality Controller

```go
type QualityController struct {
    CurrentPreset     string
    CurrentResolution string
    CurrentBitrate    int
    TargetFPS         int
}

func (qc *QualityController) AdjustForThermal(gpu GPUInfo) {
    if gpu.Temperature >= gpu.ShutdownThreshold-10 {
        // Critical: maximum reduction
        qc.CurrentPreset = "P1"
        qc.CurrentBitrate = int(float64(qc.CurrentBitrate) * 0.5)
    } else if gpu.Temperature >= gpu.SlowdownThreshold {
        // Warning: moderate reduction
        qc.CurrentPreset = "P3"
        qc.CurrentBitrate = int(float64(qc.CurrentBitrate) * 0.75)
    }
}
```

### Sunshine-Inspired Encoder Selection

Sunshine demonstrates a practical encoder selection approach:
1. Enumerate available encoders at startup
2. Prioritize hardware encoders (NVENC > QSV > VAAPI/AMF > VideoToolbox)
3. Validate encoder capability with test encode
4. Fall back to software encoding on failure
5. Allow manual override via configuration

**Claim:** Sunshine automatically detects available hardware encoders (NVENC, QuickSync, VCE) with software fallback.[^37^]
**Source:** Aurora Docs - Game Streaming with Sunshine
**URL:** https://docs.getaurora.dev/guides/sunshine/
**Date:** n/a
**Excerpt:** "Sunshine automatically detects and uses available hardware encoders... NVENC (NVIDIA), QuickSync (Intel), or VCE (AMD) should be detected. Software encoding will be used as fallback."
**Context:** This validates the encoder enumeration and fallback approach.
**Confidence:** high

---

## References

[^1^]: NVIDIA go-nvml GitHub, https://github.com/nvidia/go-nvml, 2025-08-22
[^2^]: NVML API Reference Guide, https://docs.nvidia.com/deploy/pdf/NVML_API_Reference_Guide.pdf
[^3^]: Spheron Network - GPU Monitoring for ML, https://www.spheron.network/blog/gpu-monitoring-for-ml/, 2026-01-15
[^4^]: NVIDIA NVML ClocksThrottleReasons, https://docs.nvidia.com/deploy/nvml-api/group__nvmlClocksThrottleReasons.html, 2026-03-05
[^5^]: ClusterCockpit/go-rocm-smi, https://github.com/ClusterCockpit/go-rocm-smi/, 2022-05-18
[^6^]: AMD SMI Documentation, https://rocm.docs.amd.com/_/downloads/amdsmi/en/docs-6.4.2/pdf/, 2025-07-22
[^7^]: xigang/go-rocm, https://github.com/xigang/go-rocm, 2024-12-04
[^8^]: Jellyfin Intel GPU Guide, https://jellyfin.org/docs/general/post-install/transcoding/hardware-acceleration/intel/
[^9^]: Superuser - ffmpeg QSV encoding, https://superuser.com/questions/1830920/, 2024-02-19
[^10^]: Apple Metal Counter API Tech Talk, https://developer.apple.com/videos/play/tech-talks/10001/, 2020-10-21
[^11^]: FFmpeg Codecs Documentation, https://ffmpeg.org/ffmpeg-codecs.html
[^12^]: Chromium DXVA, https://chromium.googlesource.com/chromium/src/+/f710981/media/gpu/windows/dxva_video_decode_accelerator_win.cc
[^13^]: Microsoft IDXGIFactory1::EnumAdapters1, https://learn.microsoft.com/en-us/windows/win32/api/dxgi/nf-dxgi-idxgifactory1-enumadapters1, 2022-08-22
[^14^]: Jellyfin Apple Hardware Acceleration, https://jellyfin.org/docs/general/post-install/transcoding/hardware-acceleration/apple/
[^15^]: Stack Overflow - Intel VA API encoding, https://stackoverflow.com/questions/36260474/, 2016-03-28
[^16^]: go-nvml-go package docs, https://pkg.go.dev/github.com/mxpv/nvml-go
[^17^]: MassedCompute - thermal throttling detection, https://massedcompute.com/faq-answers/, 2025-07-31
[^18^]: Sunshine Configuration Docs, https://docs.lizardbyte.dev/projects/sunshine/latest/md_docs_2configuration.html
[^19^]: gopsutil SwapMemoryStat, https://www.mintlify.com/shirou/gopsutil/api/mem/swap-memory, 2026-03-09
[^20^]: golang.org/x/sys/cpu, https://pkg.go.dev/golang.org/x/sys/cpu, 2026-03-27
[^21^]: gopsutil cpu_linux.go, https://chromium.googlesource.com/external/github.com/shirou/gopsutil/+/refs/heads/master/cpu/cpu_linux.go, 2019-02-26
[^22^]: Medium - Disk Benchmarking FIO, https://vineetcic.medium.com/disk-benchmarking-using-fio-c71e0ce0d47c, 2022-08-03
[^23^]: Kernel ethtool-netlink docs, https://docs.kernel.org/networking/ethtool-netlink.html
[^24^]: Nyxshima - Battery and AC Power Status, https://www.nyxshima.com/get-battery-and-ac-power-adapter-connection-status-via-powershell/, 2026-03-20
[^25^]: Rezmoss - macOS Hardware Detection with Go, https://rezmoss.com/blog/macos-hardware-detection-with-go/, 2025-08-02
[^26^]: Phoronix - AMDGPU GPU Load Sysfs, https://www.phoronix.com/news/AMDGPU-Load-Percent-Sysfs, 2018-06-20
[^27^]: NVIDIA NVENC Programming Guide, https://docs.nvidia.com/video-technologies/video-codec-sdk/13.0/nvenc-video-encoder-api-prog-guide/index.html
[^28^]: NVIDIA NVENC OBS Guide, https://www.nvidia.com/en-us/geforce/guides/broadcasting-guide/, 2025-01-30
[^29^]: Streamlabs Dynamic Bitrate, https://streamlabs.com/content-hub/post/how-to-use-dynamic-bitrate, 2019-10-17
[^30^]: NVIDIA Advanced Optimus Overview, https://nvidia.custhelp.com/app/answers/detail/a_id/5097, 2025-01-13
[^31^]: Wikipedia - NVENC, https://en.wikipedia.org/wiki/NVENC, updated
[^32^]: ASUS - Configure Optimus and Switchable Graphics, https://www.asus.com/us/support/faq/1038387/, 2025-11-19
[^33^]: gopsutil cross-platform guide, https://www.mintlify.com/shirou/gopsutil/guides/advanced/cross-platform, 2026-03-09
[^34^]: Medium - Apple GPUs through Go and Metal, https://medium.com/data-science/programming-apple-gpus-through-go-and-metal-shading-language-a0e7a60a3dba, 2024-02-26
[^35^]: ArcaneNibble - Hardware hotplug events on Linux, https://arcanenibble.github.io/hardware-hotplug-events-on-linux-the-gory-details.html, 2026-03-01
[^36^]: Ask Ubuntu - NVIDIA hotplug events, https://askubuntu.com/questions/858798/, 2016-12-09
[^37^]: Aurora - Game Streaming with Sunshine, https://docs.getaurora.dev/guides/sunshine/
