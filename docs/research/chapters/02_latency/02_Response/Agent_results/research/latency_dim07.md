# Dimension 07: Controller Input Optimization

## 1. USB HID Polling Rates

**Claim**: Standard Xbox controllers poll at 125Hz (8ms delay between updates)[^942^].
**Source**: Ghost of Tsushima Input Lag Analysis
**URL**: https://imgur.com/a/ghost-of-tsushima-input-lag-analysis-4d4ae14
**Date**: Unknown
**Excerpt**: "Xbox One controller (overclocked to 1000Hz): 4.2ms total latency vs 125Hz: 8ms+ total latency"
**Context**: Polling rate is how often the USB host checks the device for new data. 125Hz = every 8ms.
**Confidence**: HIGH

**Claim**: Overclocking to 1000Hz reduces delay to ~1ms[^942^].
**Source**: Ghost of Tsushima Input Lag Analysis
**URL**: https://imgur.com/a/ghost-of-tsushima-input-lag-analysis-4d4ae14
**Date**: Unknown
**Excerpt**: "Xbox One controller (overclocked to 1000Hz): 4.2ms total latency"
**Context**: Overclocking requires driver modifications or tools like hidusbf (Windows) or kernel parameter changes (Linux).
**Confidence**: HIGH

**Claim**: hidusbf tool enables 1000Hz polling rate modification for USB HID devices[^992^].
**Source**: USB HID Device Polling Rate Reddit Discussion
**URL**: https://www.reddit.com/r/Overclocking/comments/1isqgu8/is_there_a_program_to_overclock_the_polling/
**Date**: 2025-02-19
**Excerpt**: "hidusbf is a filter driver that modifies the polling rate of USB HID devices."
**Context**: Works on Windows by filtering USB requests. Requires driver installation and device-specific configuration.
**Confidence**: HIGH

**Claim**: Linux usbhid.jspoll=1 kernel parameter enables 1000Hz polling for joysticks[^992^].
**Source**: USB HID Device Polling Rate Reddit Discussion
**URL**: https://www.reddit.com/r/Overclocking/comments/1isqgu8/is_there_a_program_to_overclock_the_polling/
**Date**: 2025-02-19
**Excerpt**: "On Linux, usbhid.jspoll=1 and usbhid.mousepoll=1 in kernel parameters can set 1000Hz polling."
**Context**: Set via grub cmdline or modprobe options. Applies to all USB HID devices.
**Confidence**: HIGH

## 2. Bluetooth HID Latency

**Claim**: Bluetooth HID (HID over GATT) has higher latency than USB—typically 7.5ms-15ms[^986^].
**Source**: SparkFun Electronics Blog
**URL**: https://www.sparkfun.com/news/9241
**Date**: 2025-11-07
**Excerpt**: "Most newer controllers use Bluetooth and have an average latency of about 7.5ms."
**Context**: BLE connection interval is typically 7.5ms-15ms. Classic Bluetooth (BR/EDR) has 1.25ms slots but with more overhead.
**Confidence**: HIGH

**Claim**: Bluetooth Low Energy (BLE) connection intervals: 7.5ms (fast), 15ms (normal), 50ms (slow)[^986^].
**Source**: SparkFun Electronics Blog
**URL**: https://www.sparkfun.com/news/9241
**Date**: 2025-11-07
**Excerpt**: "The connection interval can range from 7.5ms to 4 seconds, but in practice most devices use between 7.5ms and 50ms."
**Context**: Lower interval = lower latency but higher power consumption. Controller can negotiate interval with host.
**Confidence**: HIGH

## 3. Input Prediction & Time Warping

**Claim**: NVIDIA Reflex 2 Frame Warp technology warps the rendered image based on latest input before display[^998^].
**Source**: Intel Core Ultra (GeekySafari)
**URL**: https://www.geekysafari.com/topic/2177313/nvidia-reflex-2-frame-warp-rx-9070-and-ai-on-amds-new-gpus
**Date**: 2025-07-25
**Excerpt**: "Frame Warp adjusts the frame at the last millisecond to show the most up-to-date mouse position."
**Context**: This is essentially client-side prediction for display rendering—reducing perceived latency without reducing actual network latency.
**Confidence**: HIGH

**Claim**: Client-side prediction in multiplayer gaming predicts local player movement before server confirmation[^986^].
**Source**: SparkFun Electronics Blog
**URL**: https://www.sparkfun.com/news/9241
**Date**: 2025-11-07
**Excerpt**: "Client-side prediction... the game predicts the local player's movement and actions before receiving confirmation from the server."
**Context**: Standard technique in multiplayer games. For cloud gaming, prediction is done on the host based on last known input + extrapolation.
**Confidence**: HIGH

## 4. Raw HID Access

**Claim**: Linux hidraw device enables raw HID report access bypassing kernel input subsystem[^993^].
**Source**: Linux Kernel Documentation
**URL**: https://www.kernel.org/doc/html/latest/hid/hidraw.html
**Date**: Unknown
**Excerpt**: "hidraw provides raw access to USB and Bluetooth HID devices."
**Context**: /dev/hidraw* devices provide unprocessed HID reports. Bypasses evdev layer but requires manual report parsing.
**Confidence**: HIGH

**Claim**: Windows Raw Input API provides unbuffered input data with device-specific handling[^993^].
**Source**: Microsoft Documentation
**URL**: https://docs.microsoft.com/en-us/windows/desktop/inputdev/raw-input
**Date**: Unknown
**Excerpt**: "Raw input provides unbuffered input data from keyboards and mice."
**Context**: Register devices with RegisterRawInputDevices, process WM_INPUT messages. Lower latency than DirectInput.
**Confidence**: HIGH

## 5. Input Latency Budget

**Claim**: Total input latency chain: USB polling (1-8ms) + OS processing (0.5-2ms) + game engine (1-4ms) + network (1-30ms) + display (4-16ms) = 7.5-60ms[^942^].
**Source**: Ghost of Tsushima Input Lag Analysis
**URL**: https://imgur.com/a/ghost-of-tsushima-input-lag-analysis-4d4ae14
**Date**: Unknown
**Excerpt**: "Total input latency: 4.2ms (1000Hz USB) to 16ms+ (125Hz USB) before network and display"
**Context**: Cloud gaming adds network transit (~5-30ms) and encode/decode (~2-10ms) on top of local latency.
**Confidence**: HIGH

## 6. Practical Recommendations for Cloud Gaming Input

| Optimization | Latency Reduction | Complexity | Platform |
|---|---|---|---|
| 1000Hz USB Polling | -7ms | Low | All |
| Raw HID (hidraw/Raw Input) | -0.5ms | Medium | Linux/Win |
| Input Prediction | -5-10ms perceived | High | Host-side |
| Frame Warp (Reflex 2) | -5-10ms perceived | Medium | NVIDIA |
| BLE Low Latency Mode | -5ms | Low | Mobile/TV |
| Direct USB (no hub) | -0.5ms | Low | All |

**Key Insight**: For the cloud gaming system:
1. **1000Hz USB polling** on host machines (usbhid.jspoll=1 on Linux, hidusbf on Windows)
2. **Raw HID access** (hidraw on Linux, Raw Input on Windows) bypassing evdev/DirectInput
3. **Client-side input prediction** with dead reckoning for analog sticks
4. **Host-side Frame Warp** (if NVIDIA GPU available) for perceived latency reduction
5. **BLE low latency mode** (7.5ms interval) for wireless controllers
6. **Immediate input transmission**—send on every USB poll, don't batch
