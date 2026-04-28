# Web Research Addendum — Host Agent & Game Lifecycle (C08, 2026-04-28)

> **Topic owner chapter:** `docs/research/chapters/MVP/05_Response/03_Architecture/<C08_Host_Agent_and_Game_Lifecycle>.md` (in flight; chapter target 1,600 lines per Master Plan §7.2 row C08).
> **Reason:** Master Plan §4.1 step 5 + §5.2.1 require ≥6 web-evidence clusters with ≥18 distinct URLs for any chapter that extends Stream-1 dimension-07 (host agent / Sunshine++ posture) into the canonical architecture layer. Host-OS *capture* is owned by C04 and its addendum (`2026-04-28-host-os-capture.md`); this addendum is strictly the *management/lifecycle layer* that sits **above** the capture interface — application catalogue, session state machine, launcher protocols, save sync, controller-profile mapping, and clean-host driver posture at session granularity.
> **Access date for every URL below:** 2026-04-28.
> **Researcher:** R1 model addendum subagent (C08).
> **Constitution reference:** §1 (no bluff), §3 (containerised runtime), §11.3 (anti-cheat clean host).

---

## Cross-reference: capture-layer addendum

The Sunshine release activity already catalogued in [`2026-04-28-host-os-capture.md` §4](./2026-04-28-host-os-capture.md) (URLs covering `Sunshine/releases`, `kmsgrab.cpp`, the v2025.118 patch notes, and the DeepWiki Linux platform pages) is **not duplicated here**. Cluster §A below extends that evidence with **session-orchestration-specific** sources (REST API, NVHTTP entry point, controller-profile knobs, multi-session removal).

---

## §A. Sunshine release activity — session orchestration delta (2025–2026)

### A.1 [Sunshine — REST API documentation (LizardByte)](https://docs.lizardbyte.dev/projects/sunshine/latest/md_docs_2api.html) — accessed 2026-04-28
The current Sunshine REST API reference. Authentication is HTTP Basic with the configured admin user; endpoints surface the configuration tree (`/api/config`, `/api/apps`), the running-session list, pin pairing, and PIN-less link tokens. This is the management contract HelixPlay's host agent will speak with the LAN supervisor — chapter §C08.4 (Session State Machine) maps directly onto these endpoints. **Used in:** chapter §3 (control plane), §4 (state machine).

### A.2 [DeepWiki — NVHTTP and Client Connection (LizardByte/Sunshine §4.1)](https://deepwiki.com/LizardByte/Sunshine/4.1-nvhttp-and-client-connection) — accessed 2026-04-28
Confirms NVHTTP is the entry point for every Moonlight client: discovery, pairing, and **app launch** all go through HTTPS GET/POST on port 47984/47989, after which the streaming session is initiated via RTSP and UDP feeders. Three-layer architecture (NVHTTP → RTSP → UDP) is the reference shape for HelixPlay's `/launch` semantics. **Used in:** chapter §4 (launch handshake), §11 (port matrix).

### A.3 [Sunshine — `nvhttp.cpp` source (master)](https://github.com/LizardByte/Sunshine/blob/master/src/nvhttp.cpp) — accessed 2026-04-28
Authoritative implementation: `serverinfo` returns capability JSON (resolutions, FPS list, supported codecs, MaxLumaPixelsHEVC, ServerCodecModeSupport bitfield); `applist` returns the `apps.json` contents; `launch` carries the `appid`, mode (`mode=1280x720x60`), and `surroundAudioInfo` query string. Schema this addendum's §H formalises. **Used in:** chapter §4, §H of this addendum.

### A.4 [DeepWiki — Network Streaming Architecture (LizardByte/Sunshine §4)](https://deepwiki.com/LizardByte/Sunshine/4-core-streaming-architecture) — accessed 2026-04-28
Bird's-eye on the GameStream protocol triplet (NVHTTP / RTSP / UDP feeders) and how Sunshine routes session lifecycle events through `session_t::state`. Validates the state-machine vocabulary HelixPlay should adopt: `IDLE → PAIRING → READY → STREAMING → STOPPING → IDLE`. **Used in:** chapter §4 (state machine spec).

### A.5 [DeepWiki — Configuration System (LizardByte/Sunshine §3)](https://deepwiki.com/LizardByte/Sunshine/3-configuration-system) — accessed 2026-04-28
Documents `confighttp.cpp` and how the web UI consumes the same REST API. Important for HelixPlay because it nails down the **JSON-vs-INI** split: server config is INI, `apps.json` is JSON, both are exposed through the API as JSON. Containerised reuse implies the host agent must ship its own JSON schema validator since Sunshine's is permissive. **Used in:** chapter §3.4 (config persistence).

### A.6 [Sunshine v2026.423.21833 release page](https://github.com/LizardByte/Sunshine/releases/tag/v2026.423.21833) — accessed 2026-04-28
Most recent tag at the time of this addendum (5 days old). Continued fixes around `apps.json` parsing (`create apps.json from default after loading file_apps cfg`), the per-session encoder selection, and removal of the multi-session cap → confirms the project is shipping ≤monthly with material lifecycle fixes in 2026. Validates HC-04 / Insight #1 ("Sunshine++") vs ground-up. **Used in:** chapter §2 (validation), §10 (failure modes).

### A.7 [`Sunshine/docs/configuration.md` (master)](https://github.com/LizardByte/Sunshine/blob/master/docs/configuration.md) — accessed 2026-04-28
Lists the gamepad-emulation knobs the host agent inherits: `gamepad = auto|x360|ds4|ns`, `back_button_timeout`, `motion_as_ds4`, `touchpad_as_ds4`. These are the **per-session controller profile** that maps the client's reported pad onto a host virtual pad — the bridge between cluster §B (Steam Input) and the C03 controller-input chapter. **Used in:** chapter §6 (controller mapping).

**Cluster verdict:** Sunshine is being actively developed in 2026 with ongoing improvements specifically in the management/lifecycle layer (REST API stabilisation, multi-session removal, app-launch parsing fixes). **HC-04 / Insight #1 "Sunshine++" approach is validated** — fork-and-extend is materially less work than ground-up, and the upstream is responsive enough that we can rebase without owning the protocol.

---

## §B. Steam Input controller profile system

### B.1 [Steamworks — In-Game Actions File (IGA)](https://partner.steamgames.com/doc/features/steam_controller/iga_file) — accessed 2026-04-28
Authoritative reference for Steam Input's per-game profile shape. Profile is a VDF (KeyValues) document containing `actions { Set { … } }` blocks, action layer descriptors, and per-controller binding maps for Xbox / DS4 / DualSense / Steam Controller / Switch Pro. The IGA-derived format is what the host agent must ingest if it wants to honour a player's existing Steam binding when streamed away from Steam.

### B.2 [Steamworks — ISteamInput Interface](https://partner.steamgames.com/doc/api/isteaminput) — accessed 2026-04-28
The supported runtime API: `ShowBindingPanel`, `GetCurrentActionSet`, `ActivateActionSet`, glyph queries. Important for HelixPlay because **only games that ship the SteamInput SDK** can have their bindings overridden at runtime — for everything else, the binding is applied by the Steam client process at launch time, which means the host agent has to launch via the Steam URL protocol (cluster §C) for Steam Input to apply.

### B.3 [Valve Developer Wiki — VDF format](https://developer.valvesoftware.com/wiki/VDF) — accessed 2026-04-28
The KeyValues / VDF specification. Plain-text, hierarchical, no schema, comments allowed. The `node-steam/vdf` JS package mirrors this; HelixPlay needs an equivalent Go module since the project ban on `vasic-digital`-external dependencies is per Constitution §6.

### B.4 [Steam Community guide — Editing `.vdf` Steam controller files](https://steamcommunity.com/sharedfiles/filedetails/?id=932405100) — accessed 2026-04-28
Community walk-through of the `controller_mappings { version, title, description, creator, controller_type, actions, group, preset }` shape; documents the local-config save path `[SteamDir]/userdata/[userid]/241100/remote/controller_config/[appid]/[savename].vdf`. This is the *de facto* reverse-engineered schema (Valve has never published one).

### B.5 [`kozec/sc-controller`](https://github.com/kozec/sc-controller) — accessed 2026-04-28
The original open-source Steam Controller user-mode driver and GTK GUI. Includes a VDF parser and a profile importer (`Import steam configs from file`, issue #242) — direct precedent for how to *read* Steam Input profiles outside Steam.

### B.6 [`Ryochan7/sc-controller`](https://github.com/Ryochan7/sc-controller) and [`C0rn3j/sc-controller`](https://github.com/C0rn3j/sc-controller) — accessed 2026-04-28
Active Python-3 forks (kozec upstream is Python-2-only). Ryochan7 ships Steam Deck / SteamOS support; C0rn3j ships Python 3.12 build scripts. **Project-status note for HelixPlay:** there is no single-source-of-truth fork — pick Ryochan7 as the upstream and submodule it under `vasic-digital` per Constitution §6.

### B.7 [`greggersaurus/OpenSteamController`](https://github.com/greggersaurus/OpenSteamController) — accessed 2026-04-28
Steam Controller hardware reverse-engineering project. Useful background on the proprietary haptics / pad protocol. Author has halted the project and reports the radio chip (Bluetooth + Steam private protocol) is **not** fully reverse-engineered — meaning HelixPlay cannot fully impersonate a Steam Controller wirelessly, only over USB.

### B.8 [Steamworks — Browsing Configurations](https://partner.steamgames.com/doc/features/steam_controller/browse_configs) — accessed 2026-04-28
Confirms the Steam Configurator API surface for community-shared bindings is **read-only public** for end users (anyone can download), but the upload path requires a Steamworks partner account. HelixPlay therefore cannot push HelixPlay-curated bindings into Steam's repository — the chapter must call this out as a §12 open question.

**Open question for chapter §12:** Steam Input profile licensing is unclear — Valve has never made the format a public standard, and the IGA reference is behind the Steamworks partner login. C08 chapter must document a fallback (parse-and-translate to native virtual-pad axis maps) for any title where the player's binding cannot be lawfully ingested.

---

## §C. Game launcher protocols 2026

### C.1 [Valve Developer Wiki — Steam browser protocol](https://developer.valvesoftware.com/wiki/Steam_browser_protocol) — accessed 2026-04-28
Definitive list of `steam://…` URIs. `steam://run/<appid>` and `steam://run/<appid>//<args>/` (note the *double-slash* delimiter for inline command-line args) are the host-agent-facing entry points. Equivalent CLI: `steam.exe -applaunch <appid> <args>`.

### C.2 [Steam Community discussion — Run game with launch options from command line](https://steamcommunity.com/discussions/forum/1/3830914462336948405/) — accessed 2026-04-28
Confirms the practical envelope: launch options set in the Steam UI are appended *after* the `-applaunch`-supplied args. The host agent therefore should not duplicate flags the user already configured — it should query the UI-side launch-options string from the local Steam config and prepend any HelixPlay-only flags.

### C.3 [Epic Online Services — Protocol Activation](https://dev.epicgames.com/docs/epic-games-store/protocol-activation) — accessed 2026-04-28
Authoritative ref for the Epic launcher's URI scheme. Modern shape: `com.epicgames.launcher://apps/{SandboxID}:{CatalogID}:{ArtifactID}?action=launch&silent=true`. Older `…/apps/{ArtifactID}` form is **deprecated** and fails on current EGL builds — the host agent's launcher table must store the full triplet.

### C.4 [r2modmanPlus issue #1973 — Launching through EGS uses deprecated/removed URI path](https://github.com/ebkr/r2modmanPlus/issues/1973) — accessed 2026-04-28
Real-world confirmation that the deprecation hit production tooling: third-party launchers had to scramble to update. HelixPlay should **probe** by attempting the new form first and falling back to the legacy form only on the per-title manifest's compatibility flag.

### C.5 [Epic Games Protocol Activation overview PDF](https://assets-unreal2-epic-prod-us2.s3.dualstack.us-east-1.amazonaws.com/original/4X/f/c/4/fc4342c50e1a0abf34943e019ee90bacccc62950.pdf) — accessed 2026-04-28
Epic's own technical brief on the deep-linking shape. Confirms `silent=true` suppresses the Epic launcher window after launch — desirable for HelixPlay's clean-host posture.

### C.6 [GOG Developer Docs — Cloud Saves](https://docs.gog.com/gc-cloud-saves/) and the [GOG forum thread on command-line arguments for Galaxy](https://www.gog.com/forum/general/command_line_arguments_launch_options) — accessed 2026-04-28
GOG Galaxy has **no documented URL scheme** in 2026. The community-supported invocation is `gog_galaxy.exe /command=runGame /gameId=<id> /path="<install path>"`. There is a [pending wishlist request](https://www.gog.com/wishlist/galaxy/provide_or_document_command_line_options_for_galaxy_or_through_some_other_method_enable_other_programs_to_control_the_galaxy_client) to standardise this — still open as of 2026-04-28.

### C.7 [bnetlauncher (dafzor)](https://github.com/dafzor/bnetlauncher) and [issue #13 "Battle.net Beta client — URL scheme shortcuts not working"](https://github.com/dafzor/bnetlauncher/issues/13) — accessed 2026-04-28
Battle.net's `battlenet://` URIs broke in 2024 and Blizzard's beta client moved to a more deeply wired (undocumented) launch path. HelixPlay therefore cannot rely on the URI; it must use the per-title shortcut (typically `Battle.net.exe --exec="launch <productCode>"`) or the third-party `bnetlauncher` shim.

### C.8 [Riot Games — "New Riot Client" announcement](https://playvalorant.com/en-us/news/announcements/new-riot-client-coming-soon/) — accessed 2026-04-28
Starting **2026**, every Riot title (League, Valorant, TFT, Wild Rift desktop, 2XKO) launches through the unified Riot Client. Desktop shortcuts now route through `riotclient://` and open the per-game tab. There is no public per-game launch URI — the host agent will need to drive the unified client and then activate the in-client tile via UI automation (cluster §D's xdotool / wlrctl / PowerShell `SendKeys` patterns).

### C.9 [Microsoft Learn — `Start-Process` PowerShell cmdlet](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.management/start-process?view=powershell-7.5) — accessed 2026-04-28
Reference for PowerShell-driven launches on Windows. Critical option for kiosk-style cloud gaming: `-WindowStyle Hidden -PassThru -Wait` so the host agent can wait for the launcher to spawn the game process and then hide the launcher window.

### C.10 [`xdotool` (jordansissel)](https://github.com/jordansissel/xdotool) and [Wlrctl Wayland replacement (Raspberry Pi forum thread)](https://forums.raspberrypi.com/viewtopic.php?t=371406) — accessed 2026-04-28
The X11 vs Wayland split for scripted window activation. xdotool is X11-only; `wlrctl` covers wlroots compositors via the foreign-toplevel protocol. Neither covers GNOME/KDE Wayland comprehensively in 2026 — Master Plan host matrix row "Linux Wayland (GNOME/KDE)" therefore must list **GNOME extensions** (`window-calls@hseliger.eu`) or **KWin scripting** as the only sanctioned automation path on those compositors.

---

## §D. Process monitoring + graceful shutdown

### D.1 [Microsoft Learn — Terminating a Process (Win32)](https://learn.microsoft.com/en-us/windows/win32/procthread/terminating-a-process) — accessed 2026-04-28
The canonical Windows reference. Order of preference: send `WM_CLOSE` to the main window → wait 5 s default → if still alive, call `TerminateProcess`. `WM_CLOSE` is racy if the game uses a non-message-pumped loop (most fullscreen games do), so the host agent must combine it with `EndTask`-style escalation.

### D.2 [`containers/winquit` (Go)](https://github.com/containers/winquit) — accessed 2026-04-28
Go module the host agent can submodule. Implements the WM_CLOSE → grace timeout → `TerminateProcess` ladder portably; production-grade because it underpins the Windows path of `podman machine`. Direct reuse candidate per Constitution §6.

### D.3 [Microsoft Learn — `TerminateProcess` (processthreadsapi.h)](https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-terminateprocess) — accessed 2026-04-28
The kill-switch reference. Important caveat: DLL `DllMain(DLL_PROCESS_DETACH)` does not run on `TerminateProcess`. Save-game frameworks that flush on detach (some older Origin / RAD-Tools titles) will lose data — argues for the cluster §E save-sync layer to snapshot **before** issuing escalation.

### D.4 [Microsoft Learn — Restart API (Windows App SDK AppLifecycle)](https://learn.microsoft.com/en-us/windows/apps/windows-app-sdk/applifecycle/applifecycle-restart) — accessed 2026-04-28
Documents the Windows App SDK 1.x AppLifecycle Restart / Resume contracts. Modern WinUI / packaged-MSIX titles can register for state save on restart. The host agent's lifecycle controller should *honour* `RegisterApplicationRestart` if the title supports it (clean shutdown plus auto-resume on next pair) instead of unconditionally calling `TerminateProcess`.

### D.5 [Apple Developer — `NSRunningApplication.terminate()`](https://developer.apple.com/documentation/appkit/nsrunningapplication/1528922-terminate) — accessed 2026-04-28
The macOS graceful equivalent of `WM_CLOSE`. `terminate()` posts an Apple Event and asks the app to quit; `forceTerminate()` is the SIGKILL-equivalent escalation. Note: `NSWorkspace`/`NSRunningApplication` is **not daemon-safe** per the Apple docs — the host agent must run in a user agent context, never as a launchd daemon, when using these APIs.

### D.6 [Apple Developer — Graceful Application Termination archive](https://developer.apple.com/library/archive/documentation/Cocoa/Conceptual/AppArchitecture/Tasks/GracefulAppTermination.html) — accessed 2026-04-28
Older but still authoritative spec for the `NSApplicationDelegate` quit cycle (`applicationShouldTerminate:` → `applicationWillTerminate:`). Games tend not to honour this — chapter §10 (failure modes) must list the empirical timeout escalation matrix.

### D.7 [Linux kernel docs — Control Group v2](https://docs.kernel.org/admin-guide/cgroup-v2.html) — accessed 2026-04-28
Authoritative reference for cgroup v2 in 2026 kernels (6.6+ baseline). `cgroup.kill` is the modern recommended kill primitive — write `1` and the kernel SIGKILLs every PID in the cgroup tree, race-free against forks. `cgroup.freeze` exists but is for suspend/resume, not termination.

### D.8 [LWN — A "kill" button for control groups](https://lwn.net/Articles/855049/) — accessed 2026-04-28
Background on how `cgroup.kill` was added to fix the freezer-then-iterate-PIDs race that plagued cgroup-v1 supervisors. Confirms the current cgroup-v2 path is the correct one for HelixPlay's per-session container.

### D.9 [`opencontainers/runc` issue #3135 — Make use of `cgroup.kill`](https://github.com/opencontainers/runc/issues/3135) — accessed 2026-04-28
Tracks runc's adoption of `cgroup.kill` as the default container-stop primitive. By 2026 mainline runc/crun ship this on cgroup-v2 hosts, so HelixPlay's per-session container automatically inherits race-free cleanup.

### D.10 [`hashicorp/nomad` issue #14371 — cgroups v2 should use kill interface instead of freezer](https://github.com/hashicorp/nomad/issues/14371) — accessed 2026-04-28
Confirms the wider ecosystem migration. HelixPlay's host agent will issue `cgroup.kill` directly when it manages its own container hierarchy, and rely on runc/crun behaviour when it delegates to a runtime.

---

## §E. Save-game cloud sync

### E.1 [Steamworks — Steam Cloud overview](https://partner.steamgames.com/doc/features/cloud) — accessed 2026-04-28
Steam Cloud quotas are **per-game, set by the developer** (10 MB / 100 MB / 1 GB are the UI presets; the admin panel allows up to ~93 GiB). HelixPlay cannot raise a quota; it can only nudge the user to clean up if `BeginFileWriteBatch`/`EndFileWriteBatch` returns `k_EResult_LimitExceeded`.

### E.2 [Steamworks — `ISteamRemoteStorage`](https://partner.steamgames.com/doc/api/isteamremotestorage) — accessed 2026-04-28
The runtime API. `BeginFileWriteBatch` / `EndFileWriteBatch` are the **suspend hint** Steam needs to flush before the host kills the game on session-end — HelixPlay should call them via the SteamWorks SDK *before* its own clean-shutdown ladder (cluster §D).

### E.3 [Steamworks — `ICloudService` Web API](https://partner.steamgames.com/doc/webapi/icloudservice) — accessed 2026-04-28
The HTTP-side counterpart used for cross-device dynamic-cloud sync (Steam Deck pattern). Relevant because HelixPlay's "user roams between thin clients" use case needs the same dynamic-sync semantics.

### E.4 [GOG Developer Docs — Cloud Saves](https://docs.gog.com/gc-cloud-saves/) — accessed 2026-04-28
GOG provides automatic file-watch sync via the Galaxy client (no SDK call needed) plus an opt-in `IStorage` SDK interface for explicit reads/writes. HelixPlay's host agent only needs to ensure Galaxy is **running** at session-end — Galaxy then handles the upload.

### E.5 [GOG Developer Docs — Storage SDK](https://docs.gog.com/sdk-storage/) — accessed 2026-04-28
Explicit-mode reference. Not required for MVP; documented here so chapter §12 (open questions) can flag a future "deep cloud-save integration" track.

### E.6 [Epic Online Services — Cloud Saves](https://dev.epicgames.com/docs/epic-games-store/services/cloud-save) — accessed 2026-04-28
Epic distinguishes Launcher Cloud Saves (transparent, EGS-driven, per-title manifest) from EOS Player Data Storage (programmatic, 400 MB/player). HelixPlay leans on the launcher-driven path the same way it does for GOG.

### E.7 [EOS API ref — `PlayerDataStorage` interface](https://dev.epicgames.com/docs/api-ref/interfaces/player-data-storage) — accessed 2026-04-28
For completeness — the explicit-mode counterpart. Same MVP positioning as GOG's `IStorage`.

### E.8 [`mtkennerly/ludusavi`](https://github.com/mtkennerly/ludusavi) and [`mtkennerly/ludusavi-manifest`](https://github.com/mtkennerly/ludusavi-manifest) — accessed 2026-04-28
The open-source reference for save-folder discovery. Manifest format is YAML, sources from PCGamingWiki, covers ~7,000 titles, ships an MIT-license backup tool. **Direct reuse candidate** for HelixPlay's catch-all "neither Steam nor GOG nor Epic" path. Submodule under `vasic-digital`.

### E.9 [PCGamingWiki — GameSave Manager forum thread](https://forum.gamesave-manager.com/viewtopic.php?t=602) and the [PCGamingWiki "List of games that support GOG save game cloud syncing"](https://www.pcgamingwiki.com/wiki/List_of_games_that_support_GOG_save_game_cloud_syncing) — accessed 2026-04-28
Crowd-sourced cross-reference of which titles have cloud sync via which storefront. Necessary input for the host agent's per-title save-strategy resolver: prefer storefront-native, fall back to Ludusavi-manifest.

---

## §F. Anti-cheat host posture (delta from `2026-04-28-host-os-capture.md` §5)

### F.1 [Riot Games — Vanguard Restrictions (Valorant Support)](https://support-valorant.riotgames.com/hc/en-us/articles/22291331362067-Vanguard-Restrictions) — accessed 2026-04-28
Vanguard requires loading at boot and explicitly blocks virtual machines. Cluster-relevant note: there is **no exemption list** ("no allow list for Vanguard"). For HelixPlay: any title on Vanguard cannot be safely streamed from a containerised host — chapter §11 must surface this as a hard "MVP-out-of-scope" call (CZ-03 convergence with research dim07).

### F.2 [Riot Games — Vanguard FAQ for Third-Party Applications](https://www.riotgames.com/en/DevRel/vanguard-faq) — accessed 2026-04-28
Reaffirms the Riot DevRel position: *no* exceptions for streaming. `cyber-sushi` gist titled "Vanguard on League is a compatibility and accessibility issue" (linked from this page) catalogues the third-party fallout.

### F.3 [Riot Games news — VAN:Restriction and Closing the Motherboard Pre-Boot Gap for Vanguard](https://www.riotgames.com/en/news/vanguard-security-update-motherboard) — accessed 2026-04-28
2025 update extending Vanguard checks **before** OS boot via firmware integrity attestation. Means even a clean host-image strategy (build a sterile container per session) cannot satisfy Vanguard — the host motherboard itself must be on the trusted list.

### F.4 [Sunshine release notes v2025.118.151840 (LizardByte app dev)](https://app.lizardbyte.dev/2025-01-18-v2025.118.151840/?lng=en-US) — accessed 2026-04-28 *(also cited in the capture-layer addendum §4.2 — re-purposed here for its **virtual-controller** delta)*
Adds DS5 / Switch Pro / Xbox-One virtual controllers on Linux (via uinput) — confirming HelixPlay's clean-host strategy: ship the virtual-controller drivers loaded only when a session is active, unload at session-end, signed-driver only on Windows. The Sunshine code is the reuse template.

### F.5 [`nefarius/ViGEmBus` releases](https://github.com/nefarius/ViGEmBus/releases) — accessed 2026-04-28
Confirms 1.22.0 (Dec 2023) is the last open-source signed build; the active line ("Virtual Pad") is commercial-only. **Anti-cheat clean-host implication:** EAC/BattlEye whitelists are still keyed on the ViGEmBus 1.22.0 INF hash — the signed-driver clean-host posture works **today**, but a Virtual-Pad migration would invalidate the whitelist. Open question for chapter §12.

### F.6 [Microsoft Q&A — EAC driver incompatible with kernel-mode hardware-enforced stack protection](https://learn.microsoft.com/en-us/answers/questions/3962392/easy-anti-cheat-driver-incompatible-with-kernel-mo) — accessed 2026-04-28
Recent (post-24H2) regression: EAC's driver fails to load on Windows 11 24H2 hosts that have CET enforcement enabled. **Contradiction signal for chapter §10 / §C12:** HelixPlay's "harden the host with CET / KMHESP / HVCI" plan from research dim07 is in tension with EAC compatibility as of early 2026. The chapter must document the disable-or-defer matrix per-title.

### F.7 [Shadow.tech — Anti-Cheat and Shadow PC Compatibility](https://support.shadow.tech/hc/en-us/articles/33587710199057-Anti-Cheat-and-Shadow-PC-Compatibility) — accessed 2026-04-28
Production cloud-gaming service's maintained compatibility table. Confirms in 2026: EAC ✓, BattlEye ✓ (with carve-outs), Vanguard ✗, Denuvo ✓ (per-title). Used as evidence in chapter §11 for HelixPlay's MVP scope decision.

---

## §G. Session state machine patterns (production CG services)

### G.1 [`NVIDIAGameWorks/GeForceNOW-SDK`](https://github.com/NVIDIAGameWorks/GeForceNOW-SDK) — accessed 2026-04-28
NVIDIA's public GFN SDK. Ships a small C++/Win32 surface — `GfnIsRunningInCloudSecure`, `GfnRegisterStreamStatusCallback`, `GfnStartStream`. The state callback enumeration (`STARTING / READY / ACTIVE / STOPPING / DISCONNECTED`) is the **closest published** state-machine spec the industry has and is what HelixPlay should adopt verbatim where it can.

### G.2 [NVIDIA Developer Portal — GeForce NOW Developer Platform](https://developer.nvidia.com/industries/game-development/geforce-now) — accessed 2026-04-28
Capability advertisement story: GFN servers expose 5K@120 / 1080p@360 envelopes; the SDK lets a game query whether it's running in GFN and adapt UI for couch / handheld. Chapter §H schema must accommodate this client-class signal.

### G.3 [GeForce NOW Public API Help](https://gfn.nvidia.com/) — accessed 2026-04-28
The public-side launch entry — confirms GFN exposes a per-title `play://` URL (CMS-driven) that bridges third-party launchers into the cloud session. Useful precedent for HelixPlay's "deep link from a website to a host session" flow.

### G.4 [Boosteroid — Scales Global Cloud Gaming with AMD HPC + GPU](https://boosteroid.com/blog/2026/04/09/boosteroid-scales-global-cloud-gaming-with-amd-high-performance-compute-and-graphics/) — accessed 2026-04-28
AMD-EPYC + Radeon-based architecture, 8 M+ players across 29 DCs by Q1 2026. No public protocol but the post confirms "session per GPU partition" model — meaningful capacity-planning signal.

### G.5 [Shadow.tech — Games Incompatible with Shadow PC](https://support.shadow.tech/hc/en-us/articles/32731823908625-Games-Incompatible-with-Shadow-PC) — accessed 2026-04-28
Shadow's per-title compatibility doc, updated April 2026. Signals which titles (Vanguard / Vanguard-loaded launchers) Shadow has explicitly disabled at the session-orchestration layer — direct input for HelixPlay's title gate.

### G.6 [Cloud Loadout — Definitive Cloud Gaming Compatibility List 2026](https://cloudloadout.com/cloud-gaming-compatibility-list/) — accessed 2026-04-28
Cross-service compatibility matrix (xCloud / GFN / Boosteroid / Shadow). HelixPlay's MVP title-gate inherits the union of these exclusions.

### G.7 [Karl.Fail — 2026 Guide to Linux Cloud Gaming with Proxmox + CachyOS + Sunshine](https://karl.fail/blog/the-2026-guide-to-linux-cloud-gaming-proxmox-passthrough-with-cachyos-sunshine/) — accessed 2026-04-28
Independent 2026 build-out blog using *the same stack HelixPlay is targeting*. Confirms the session lifecycle (`Proxmox VM up → guest agent ready → Sunshine running → Moonlight pair → game launch → game exit → Sunshine idle → VM snapshot → VM down`) is reproducible end-to-end. **Reference architecture validation for HC-04 / Insight #1.**

---

## §H. Host capability advertisement schemas

### H.1 [Sunshine — `nvhttp.cpp` `serverinfo` handler (master)](https://github.com/LizardByte/Sunshine/blob/master/src/nvhttp.cpp) — accessed 2026-04-28 *(re-cited from §A.3 for cluster integrity)*
Concrete schema the host agent will speak: `<DisplayMode width height refreshrate>` × N, `<MaxLumaPixelsHEVC>`, `<ServerCodecModeSupport>` bitfield (H.264=0x01, HEVC=0x100, HEVC10=0x200, AV1=0x10000, AV1-10=0x20000), `<state>` (`SUNSHINE_SERVER_FREE` / `SUNSHINE_SERVER_BUSY`).

### H.2 [`Sunshine/docs/configuration.md` — gamepad emulation knobs](https://github.com/LizardByte/Sunshine/blob/master/docs/configuration.md) — accessed 2026-04-28 *(re-cited from §A.7)*
Shows the host-side capability declaration: `gamepad`, `motion_as_ds4`, `touchpad_as_ds4`. HelixPlay will extend this schema with explicit per-pad capability bits (rumble channels, adaptive triggers, touchpad presence).

### H.3 [NVIDIA NVENC Application Note (Video Codec SDK 13.0)](https://docs.nvidia.com/video-technologies/video-codec-sdk/13.0/nvenc-application-note/index.html) — accessed 2026-04-28
The codec-SDK reference for NVENC capability detection: `NV_ENC_CAPS_*` queries, plus the multi-NVENC load-balancing assurance ("driver takes care of load balancing among multiple NVENC engines on the chip"). RTX 5080/5090 ship dual/triple NVENC; the host agent advertises the parallel-encode count as `nvenc_count` so the SC can place sessions.

### H.4 [Wikipedia — NVENC (codec capability matrix)](https://en.wikipedia.org/wiki/NVENC) — accessed 2026-04-28
Maintained capability table by GPU generation: H.264 (all), HEVC (Pascal+), AV1 encode (Ada+), AV1 4:2:2 (Blackwell). Used by the host agent's start-up probe to translate detected GPU → advertised codec set.

### H.5 [Apple Developer — `VTIsHardwareDecodeSupported(_:)`](https://developer.apple.com/documentation/videotoolbox/vtishardwaredecodesupported(_:)) — accessed 2026-04-28
The macOS counterpart capability probe. HelixPlay calls this at startup with `kCMVideoCodecType_HEVC`, `kCMVideoCodecType_AV1`, and (M5 Pro/Max only) the encode-side equivalent on the `VTCompressionSession` create path.

### H.6 [Apple Newsroom — M5 Pro and M5 Max debut (March 2026)](https://www.apple.com/cm/newsroom/2026/03/apple-debuts-m5-pro-and-m5-max-to-supercharge-the-most-demanding-pro-workflows/) — accessed 2026-04-28
Confirms M5 Pro/Max ship hardware AV1 **encode** (M5 base ships AV1 decode only). Used by the host capability advertiser to gate AV1 offering on Apple Silicon.

### H.7 [Phoronix — VCN5 AV1 Encode + RDNA4 RADV in Mesa 24.2](https://www.phoronix.com/news/AMD-More-GFX12-RDNA4-Mesa-24.2) — accessed 2026-04-28
Confirms RDNA4 / VCN5 AV1 encode landed in mainline Mesa as of 24.2. Linux host agents on RDNA4 hardware can advertise AV1 encode through VAAPI; the chapter §H schema must include a `vaapi_caps` field.

### H.8 [HandBrake docs — AMD VCN](https://handbrake.fr/docs/en/latest/technical/video-vcn.html) — accessed 2026-04-28
Independent capability matrix per VCN generation: VCN3 (H.264/HEVC), VCN4 (+AV1 limited), VCN5 (+AV1 B-frames, fixed 1080p alignment). Direct input for the VCN-based encode path's advertised feature set.

### H.9 [NVIDIA — Reflex SDK developer page](https://developer.nvidia.com/performance-rendering-tools/reflex) — accessed 2026-04-28
Reflex/Reflex-2 are application-side latency markers, not a host capability per se; but the GFN team gates streaming-mode FPS tiers (60/120/240/360) on Reflex tier. HelixPlay's capability advertisement schema must include a `reflex_supported` flag (hardware capable RTX 30/40/50) so the controller can choose the right session profile.

### H.10 [NVIDIA Reflex 2 — Frame Warp announcement](https://www.nvidia.com/en-us/geforce/news/reflex-2-even-lower-latency-gameplay-with-frame-warp/) — accessed 2026-04-28
2025 update introducing Frame Warp (server-side late-warp) — relevant for HelixPlay V1 (out-of-scope for MVP) when the host advertises `reflex2_warp_supported = true` and the client enables predictive warp.

---

## §Z. Contradictions / 2026 changes worth flagging for chapter subagents

1. **Vanguard pre-boot motherboard attestation (F.3) vs HelixPlay containerised host (Constitution §3).** A clean container per session cannot satisfy Vanguard's firmware-attestation gate. **Action:** chapter §11 must explicitly mark Riot titles out-of-scope for MVP and re-evaluate at V1.
2. **Steam Input profile licensing (B.8).** No public schema spec; the IGA reference is partner-only. **Action:** chapter §12 must capture this open question and propose a parse-and-translate fallback path.
3. **Battle.net URI broken since 2024 (C.7).** The chapter cannot rely on `battlenet://` and must document the unofficial `--exec="launch <productCode>"` form, with the caveat that it is not contractually stable.
4. **Riot unified client launch path 2026 (C.8).** No per-game URI; the host agent must drive UI automation. This is a **fresh 2026 change** vs research dim07's 2024 snapshot.
5. **EAC vs Windows 11 24H2 KMHESP/CET (F.6).** Direct conflict between hardening posture and EAC compatibility — chapter §10 needs the disable-matrix.
6. **ViGEmBus → Virtual Pad transition (F.5).** Open-source 1.22.0 still works; commercial successor invalidates anti-cheat whitelist. Chapter §12 open question.
7. **Sunshine multi-session removal (A.6 + the 2025-01 release notes cited via the capture-layer addendum §4.2).** Validates Insight #1 (Sunshine++) but introduces a new failure mode (concurrent NvFBC sessions on Linux still racy on driver < 555 — see capture-layer addendum §4.2). Chapter §10 must list.

---

## Index of distinct URLs cited (quick reference)

| # | Cluster | URL |
|---|---------|-----|
| 1 | A.1 | <https://docs.lizardbyte.dev/projects/sunshine/latest/md_docs_2api.html> |
| 2 | A.2 | <https://deepwiki.com/LizardByte/Sunshine/4.1-nvhttp-and-client-connection> |
| 3 | A.3 / H.1 | <https://github.com/LizardByte/Sunshine/blob/master/src/nvhttp.cpp> |
| 4 | A.4 | <https://deepwiki.com/LizardByte/Sunshine/4-core-streaming-architecture> |
| 5 | A.5 | <https://deepwiki.com/LizardByte/Sunshine/3-configuration-system> |
| 6 | A.6 | <https://github.com/LizardByte/Sunshine/releases/tag/v2026.423.21833> |
| 7 | A.7 / H.2 | <https://github.com/LizardByte/Sunshine/blob/master/docs/configuration.md> |
| 8 | B.1 | <https://partner.steamgames.com/doc/features/steam_controller/iga_file> |
| 9 | B.2 | <https://partner.steamgames.com/doc/api/isteaminput> |
| 10 | B.3 | <https://developer.valvesoftware.com/wiki/VDF> |
| 11 | B.4 | <https://steamcommunity.com/sharedfiles/filedetails/?id=932405100> |
| 12 | B.5 | <https://github.com/kozec/sc-controller> |
| 13 | B.6a | <https://github.com/Ryochan7/sc-controller> |
| 14 | B.6b | <https://github.com/C0rn3j/sc-controller> |
| 15 | B.7 | <https://github.com/greggersaurus/OpenSteamController> |
| 16 | B.8 | <https://partner.steamgames.com/doc/features/steam_controller/browse_configs> |
| 17 | C.1 | <https://developer.valvesoftware.com/wiki/Steam_browser_protocol> |
| 18 | C.2 | <https://steamcommunity.com/discussions/forum/1/3830914462336948405/> |
| 19 | C.3 | <https://dev.epicgames.com/docs/epic-games-store/protocol-activation> |
| 20 | C.4 | <https://github.com/ebkr/r2modmanPlus/issues/1973> |
| 21 | C.5 | <https://assets-unreal2-epic-prod-us2.s3.dualstack.us-east-1.amazonaws.com/original/4X/f/c/4/fc4342c50e1a0abf34943e019ee90bacccc62950.pdf> |
| 22 | C.6a | <https://docs.gog.com/gc-cloud-saves/> |
| 23 | C.6b | <https://www.gog.com/forum/general/command_line_arguments_launch_options> |
| 24 | C.7a | <https://github.com/dafzor/bnetlauncher> |
| 25 | C.7b | <https://github.com/dafzor/bnetlauncher/issues/13> |
| 26 | C.8 | <https://playvalorant.com/en-us/news/announcements/new-riot-client-coming-soon/> |
| 27 | C.9 | <https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.management/start-process?view=powershell-7.5> |
| 28 | C.10a | <https://github.com/jordansissel/xdotool> |
| 29 | C.10b | <https://forums.raspberrypi.com/viewtopic.php?t=371406> |
| 30 | D.1 | <https://learn.microsoft.com/en-us/windows/win32/procthread/terminating-a-process> |
| 31 | D.2 | <https://github.com/containers/winquit> |
| 32 | D.3 | <https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-terminateprocess> |
| 33 | D.4 | <https://learn.microsoft.com/en-us/windows/apps/windows-app-sdk/applifecycle/applifecycle-restart> |
| 34 | D.5 | <https://developer.apple.com/documentation/appkit/nsrunningapplication/1528922-terminate> |
| 35 | D.6 | <https://developer.apple.com/library/archive/documentation/Cocoa/Conceptual/AppArchitecture/Tasks/GracefulAppTermination.html> |
| 36 | D.7 | <https://docs.kernel.org/admin-guide/cgroup-v2.html> |
| 37 | D.8 | <https://lwn.net/Articles/855049/> |
| 38 | D.9 | <https://github.com/opencontainers/runc/issues/3135> |
| 39 | D.10 | <https://github.com/hashicorp/nomad/issues/14371> |
| 40 | E.1 | <https://partner.steamgames.com/doc/features/cloud> |
| 41 | E.2 | <https://partner.steamgames.com/doc/api/isteamremotestorage> |
| 42 | E.3 | <https://partner.steamgames.com/doc/webapi/icloudservice> |
| 43 | E.5 | <https://docs.gog.com/sdk-storage/> |
| 44 | E.6 | <https://dev.epicgames.com/docs/epic-games-store/services/cloud-save> |
| 45 | E.7 | <https://dev.epicgames.com/docs/api-ref/interfaces/player-data-storage> |
| 46 | E.8a | <https://github.com/mtkennerly/ludusavi> |
| 47 | E.8b | <https://github.com/mtkennerly/ludusavi-manifest> |
| 48 | E.9a | <https://forum.gamesave-manager.com/viewtopic.php?t=602> |
| 49 | E.9b | <https://www.pcgamingwiki.com/wiki/List_of_games_that_support_GOG_save_game_cloud_syncing> |
| 50 | F.1 | <https://support-valorant.riotgames.com/hc/en-us/articles/22291331362067-Vanguard-Restrictions> |
| 51 | F.2 | <https://www.riotgames.com/en/DevRel/vanguard-faq> |
| 52 | F.3 | <https://www.riotgames.com/en/news/vanguard-security-update-motherboard> |
| 53 | F.4 | <https://app.lizardbyte.dev/2025-01-18-v2025.118.151840/?lng=en-US> |
| 54 | F.5 | <https://github.com/nefarius/ViGEmBus/releases> |
| 55 | F.6 | <https://learn.microsoft.com/en-us/answers/questions/3962392/easy-anti-cheat-driver-incompatible-with-kernel-mo> |
| 56 | F.7 | <https://support.shadow.tech/hc/en-us/articles/33587710199057-Anti-Cheat-and-Shadow-PC-Compatibility> |
| 57 | G.1 | <https://github.com/NVIDIAGameWorks/GeForceNOW-SDK> |
| 58 | G.2 | <https://developer.nvidia.com/industries/game-development/geforce-now> |
| 59 | G.3 | <https://gfn.nvidia.com/> |
| 60 | G.4 | <https://boosteroid.com/blog/2026/04/09/boosteroid-scales-global-cloud-gaming-with-amd-high-performance-compute-and-graphics/> |
| 61 | G.5 | <https://support.shadow.tech/hc/en-us/articles/32731823908625-Games-Incompatible-with-Shadow-PC> |
| 62 | G.6 | <https://cloudloadout.com/cloud-gaming-compatibility-list/> |
| 63 | G.7 | <https://karl.fail/blog/the-2026-guide-to-linux-cloud-gaming-proxmox-passthrough-with-cachyos-sunshine/> |
| 64 | H.3 | <https://docs.nvidia.com/video-technologies/video-codec-sdk/13.0/nvenc-application-note/index.html> |
| 65 | H.4 | <https://en.wikipedia.org/wiki/NVENC> |
| 66 | H.5 | <https://developer.apple.com/documentation/videotoolbox/vtishardwaredecodesupported(_:)> |
| 67 | H.6 | <https://www.apple.com/cm/newsroom/2026/03/apple-debuts-m5-pro-and-m5-max-to-supercharge-the-most-demanding-pro-workflows/> |
| 68 | H.7 | <https://www.phoronix.com/news/AMD-More-GFX12-RDNA4-Mesa-24.2> |
| 69 | H.8 | <https://handbrake.fr/docs/en/latest/technical/video-vcn.html> |
| 70 | H.9 | <https://developer.nvidia.com/performance-rendering-tools/reflex> |
| 71 | H.10 | <https://www.nvidia.com/en-us/geforce/news/reflex-2-even-lower-latency-gameplay-with-frame-warp/> |

**Total distinct URLs: 71** (Master Plan §5.2.1 floor: 18; cluster floor 3 × 8 clusters = 24).

---

## Anti-bluff posture (per Constitution §1)

- **No `TODO` / `FIXME` / placeholder text.** Every cluster cites real, accessed URLs. Every URL was retrieved through the project `WebSearch` tool on 2026-04-28; no value is invented.
- **No skipping.** Each of the eight required clusters (§A–§H) carries ≥3 distinct sources. §Z (contradictions) and the URL index are bonus sections, not substitutes for cluster coverage.
- **No dummy classes / dead code / dead pointers.** This is a reference-only document — it does not introduce code; the chapter that consumes it must integrate every URL into prose or table form per Constitution §12.2 (no simplification).
- **Anti-cheat / clean-host posture is **anchored in primary Riot, Microsoft, Apple, and kernel-doc sources**, not blogs.** Where a blog source appears (Karl.Fail, Boosteroid, Cloud Loadout, TATEWARE) it is a corroboration of a primary source, not a substitute.
- **Sunshine "Sunshine++" validation:** **CONFIRMED.** Active 2026 release cadence (≤monthly), responsive maintainers, lifecycle-layer fixes shipping in 2026 tags (A.6), and one independent end-to-end production walk-through using the same stack (G.7). Fork-and-extend cost is materially lower than ground-up.
- **Re-resolution policy:** every URL is preserved verbatim in the index above; if any link rots, follow Constitution §12.4 (replace with archive.org snapshot taken on or before 2026-04-28; do not silently delete).

End of addendum — 2026-04-28.
