# Web Research Addendum — Controller Input Pipeline (C03)

> **Date assembled:** 2026-04-28
> **Owning chapter:** [`../03_Architecture/02_Controller_Input_Pipeline.md`](../03_Architecture/02_Controller_Input_Pipeline.md)
> **Author:** subagent C03
> **Purpose:** record the web evidence consulted while extending the
> Stream-1 dimension-02 research into the canonical Architecture
> chapter for controller capture, transport, and host-side injection.
> All sources below were reached via the project's `WebSearch` tool on
> the date listed and are cited inline in the chapter body.

---

## Index of sources

| # | URL | Title | Accessed | Used in chapter section |
|---|-----|-------|----------|-------------------------|
| W-01 | <https://github.com/nefarius/ViGEmBus> | nefarius/ViGEmBus — Windows kernel-mode driver emulating well-known USB game controllers (retired tag 1.22.0) | 2026-04-28 | §5.1 (Windows injection successor) |
| W-02 | <https://github.com/nefarius/ViGEmBus/releases> | ViGEmBus releases page — 1.22.0 final signed binaries | 2026-04-28 | §5.1, §10 (failure modes) |
| W-03 | <https://docs.nefarius.at/Downloads/> | Nefarius project documentation — downloads & Virtual Pad commercial successor mention | 2026-04-28 | §5.1, §12 (open question on Virtual Pad licensing) |
| W-04 | <https://developer.apple.com/documentation/hiddriverkit> | Apple Developer — HIDDriverKit framework reference | 2026-04-28 | §5.2 (macOS DriverKit virtual HID) |
| W-05 | <https://developer.apple.com/forums/thread/812774> | Apple Developer Forums — Virtual Controllers on Mac via the Game Controller framework | 2026-04-28 | §5.2, §6 (DualSense parity), §12 |
| W-06 | <https://developer.apple.com/forums/thread/811634> | Apple Developer Forums — HidHide on macOS thread (DriverKit/IOHIDUserDevice limits) | 2026-04-28 | §5.2, §10 |
| W-07 | <https://github.com/pqrs-org/Karabiner-DriverKit-VirtualHIDDevice> | Karabiner DriverKit Virtual HID Device — production-shipping notarised macOS DriverKit virtual HID precedent | 2026-04-28 | §5.2 (signing, notarisation) |
| W-08 | <https://github.com/Ohjurot/DualSense-Windows> | DualSense-Windows — Windows API for DualSense raw HID, output reports for haptics & adaptive triggers | 2026-04-28 | §3 (protocol), §6 |
| W-09 | <https://github.com/rafaelvaloto/Gamepad-Core> | Gamepad-Core — modern C++ library for advanced DualSense/DS4 features (engine-agnostic) | 2026-04-28 | §6 (parity matrix), §9 |
| W-10 | <https://www.pcgamingwiki.com/wiki/Controller:DualSense> | PCGamingWiki — DualSense feature support matrix on PC (USB vs Bluetooth) | 2026-04-28 | §6, §7 (BT tradeoffs) |
| W-11 | <https://github.com/streetpea/chiaki-ng/issues/403> | chiaki-ng issue #403 — DualSense Edge haptic/trigger regressions over BT | 2026-04-28 | §6, §10 |
| W-12 | <https://v1993.github.io/cemuhook-protocol/> | cemuhook-protocol — community canonical CemuhookUDP (DSU) specification | 2026-04-28 | §3.4 (motion frame), §5 |
| W-13 | <https://cemuhook.sshnuke.net/padudpserver.html> | UDP Pad motion data provider setup — original Cemuhook documentation, port 26760, units (g, deg/s) | 2026-04-28 | §3.4 |
| W-14 | <https://github.com/kmicki/SteamDeckGyroDSU> | SteamDeckGyroDSU — DSU/CemuhookUDP server for Steam Deck (active 2024–2026) | 2026-04-28 | §3.4 (adoption note) |
| W-15 | <https://hidusbf.com/2025/08/25/what-is-hidusbf-complete-beginners-guide-to-usb-polling-rate-overclocking-in-windows-10-and-11/> | hidusbf 2025 guide — Windows 10/11 polling overclock; Secure Boot interaction | 2026-04-28 | §8 (1000 Hz) |
| W-16 | <https://hidusbf.com/2025/08/25/hidusbf-troubleshooting-guide-fix-polling-rate-issues-secure-boot-errors-and-driver-conflicts/> | hidusbf troubleshooting — driver conflicts, Secure Boot blockers | 2026-04-28 | §8, §10 |
| W-17 | <https://wiki.archlinux.org/title/Mouse_polling_rate> | ArchWiki — Mouse polling rate (kernel parameters, USB HS limits) | 2026-04-28 | §8 |
| W-18 | <https://github.com/p0358/usb_oc-dkms> | usb_oc-dkms — Linux kernel module equivalent of hidusbf for USB overclocking | 2026-04-28 | §8 |
| W-19 | <https://github.com/MiSTer-devel/Main_MiSTer/issues/139> | MiSTer issue #139 — `usbhid.jspoll=1` not always taking effect on USB-3 hubs | 2026-04-28 | §8, §10 |
| W-20 | <https://www.kernel.org/doc/html/v4.12/input/uinput.html> | Linux kernel — uinput module documentation (creation, capabilities, security) | 2026-04-28 | §5.3 (Linux injection) |
| W-21 | <https://kernel.org/doc/html/latest/input/input_uapi.html> | Linux kernel — Input Subsystem userspace API | 2026-04-28 | §5.3, §9 |
| W-22 | <https://github.com/KarsMulder/evsieve> | evsieve — userspace evdev → uinput remapper (model for HelixPlay's host adapter) | 2026-04-28 | §9 (Go pseudocode references), §11 |
| W-23 | <https://en.wikipedia.org/wiki/Evdev> | Wikipedia — evdev (input event interface) | 2026-04-28 | §2 (Linux row) |

## Notes on each cluster

### Windows virtual gamepad after ViGEmBus

Sources W-01 through W-03 confirm that ViGEmBus 1.22.0 (released
December 2023) is the **final** open-source signed build; Nefarius
has rebranded the active line as **"Virtual Pad"**, available only
to commercial business partners at the time of writing. For
HelixPlay this means: (a) the existing 1.22.0 binary remains usable
on Windows 10 and Windows 11 with Secure Boot enabled, (b) any
future on-driver work needs either a commercial Virtual Pad licence
or a second-source kernel-mode driver, and (c) the Constitution's
§11.3 anti-cheat clean-host rule requires the driver be both signed
and routinely re-evaluated against EAC/BattlEye/Vanguard whitelists.
The chapter must record this as an open question (§12) and a
failure-mode row (§10) until Nefarius publishes a community
licensing arrangement.

### macOS DriverKit virtual HID

W-04 through W-07 confirm Apple's HIDDriverKit framework is
**explicitly oriented toward physical HIDs** (keyboards, pointing
devices, digitizers); the Game Controller framework deliberately
ignores virtual HIDs to avoid feedback loops. The Karabiner project
(W-07) demonstrates that a notarised DriverKit driver can present a
virtual HID device, but Apple's developer-account permission for
DriverKit signing requires an application beyond the standard tier.
HelixPlay's macOS path therefore stays on `IOHIDUserDevice` /
DriverKit with Karabiner-style notarisation as the reference, with
foohid (kext) explicitly deprecated as a fallback only on legacy
Intel Macs with SIP disabled.

### DualSense feature parity

W-08 through W-11 anchor the per-feature parity matrix in §6. PC
support is mature for output reports (rumble, lightbar, adaptive
triggers, player LEDs) over USB; haptic-audio output and adaptive
triggers are **disabled over Bluetooth** at the protocol level. The
Chiaki-NG bug thread (W-11) further documents a regression on
DualSense Edge over BT in mid-2025, which remains the standing
reason HelixPlay's competitive tier requires USB or 2.4 GHz dongle.

### CemuhookUDP / DSU

W-12 through W-14 confirm that the protocol is still the
de-facto-standard for motion forwarding outside Sony's first-party
software, with active Steam Deck adoption. There is **no Sony
official adoption**; the project must either implement DSU as a
parallel stream (the 1,000 Hz IMU path) or fold the IMU bytes into
the primary 24-byte frame. The chapter's recommendation is the
parallel stream pattern (§3.4) so DSU consumers (Cemu, Yuzu,
RPCS3, Dolphin) can ride alongside HelixPlay's primary protocol.

### 1000 Hz polling under Secure Boot

W-15 through W-19 describe the practical regression: Windows 11 with
Secure Boot blocks unsigned filter drivers, so `hidusbf` requires
**either** Secure Boot disabled **or** test-signing **or** a signed
fork (none of which ship by default). On Linux, the
`usbhid.jspoll=1` parameter is unreliable on USB-3 hubs and on
devices that bind a non-default driver (xpad). The chapter must
therefore avoid promising a silver-bullet "1000 Hz everywhere" and
instead document the per-OS recipe + measurement methodology so the
client can surface the real attained polling rate.

### Linux uinput security

W-20 through W-23 confirm uinput remains the canonical kernel-mediated
path; access requires write to `/dev/uinput` (typical solution: udev
rule granting the `input` or a dedicated `uinput` group). No
`CAP_SYS_ADMIN` is needed for normal use, but `CAP_NET_ADMIN`
adjacent capabilities may be required for some sandboxed runtimes.
The Go pseudocode in §9 follows the evsieve pattern (W-22) for
userspace forwarding semantics.

---

## Anti-bluff posture

- All URLs above were retrieved via the official `WebSearch` tool on
  2026-04-28; no source is cited from training memory.
- Where the WebSearch returned multiple candidates for the same
  factual claim (e.g. ViGEmBus retirement), the canonical source
  (the Nefarius repository or the Apple Developer Forum thread) is
  the one quoted in the chapter; secondary sources are listed here
  for triangulation.
- No claim from these sources is rephrased to remove a constraint
  the source explicitly states (per Constitution §1.1, no dodging
  language).

End of addendum — 2026-04-28.
