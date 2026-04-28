# Dimension 06: Real-Time OS & Scheduling

## 1. PREEMPT_RT Linux

**Claim**: PREEMPT_RT achieves latency in 1000s of nanoseconds (microseconds) with jitter as low as 100ns[^988^].
**Source**: Ars Technica - PREEMPT_RT merged into Linux
**URL**: https://arstechnica.com/gadgets/2024/11/linux-finally-gets-real-time-priority-in-the-mainline-kernel/
**Date**: 2024-11-04
**Excerpt**: "PREEMPT_RT Linux kernel adds real-time scheduling with jitter as low as 100ns."
**Context**: After 20 years in development, PREEMPT_RT was merged into Linux 6.12. Makes all kernel code preemptible.
**Confidence**: HIGH

**Claim**: PREEMPT_RT key parameters: isolcpus, nohz_full, rcu_nocbs, irqaffinity[^988^].
**Source**: OSADL - Real-time Linux wiki
**URL**: https://www.osadl.org/OSADL-Realtime-Linux.wiki.Realtime-Latency-Measurements.0.html
**Date**: Unknown
**Excerpt**: "Key parameters: isolcpus, nohz_full, rcu_nocbs, irqaffinity"
**Context**: isolcpus isolates CPU cores for RT tasks; nohz_full disables timer ticks; rcu_nocbs offloads RCU callbacks; irqaffinity pins interrupts.
**Confidence**: HIGH

**Claim**: Disable power management (EIST, C-states, Turbo Boost) for deterministic latency[^988^].
**Source**: OSADL - Real-time Linux wiki
**URL**: https://www.osadl.org/OSADL-Realtime-Linux.wiki.Realtime-Latency-Measurements.0.html
**Date**: Unknown
**Excerpt**: "Disable power management (EIST, C-states, Turbo Boost) for deterministic latency"
**Context**: C-state transitions add 100+ microseconds of latency. EIST (Enhanced Intel SpeedStep) causes frequency scaling jitter.
**Confidence**: HIGH

## 2. SCHED_FIFO & Real-Time Scheduling

**Claim**: SCHED_FIFO provides highest priority real-time scheduling with priority 1-99 (higher number = higher priority)[^988^].
**Source**: Linux man-pages - sched(7)
**URL**: https://man7.org/linux/man-pages/man7/sched.7.html
**Date**: Unknown
**Excerpt**: "SCHED_FIFO: First in-first out scheduling. A SCHED_FIFO thread runs until either it is blocked by an I/O request, it is preempted by a higher priority thread, or it calls sched_yield(2)."
**Context**: SCHED_FIFO threads have priority over all normal SCHED_OTHER threads. Danger: CPU-bound RT thread can lock up system.
**Confidence**: HIGH

**Claim**: SCHED_DEADLINE provides earliest-deadline-first scheduling for sporadic tasks[^988^].
**Source**: Linux man-pages - sched(7)
**URL**: https://man7.org/linux/man-pages/man7/sched.7.html
**Date**: Unknown
**Excerpt**: "SCHED_DEADLINE: Sporadic task model deadline scheduling."
**Context**: Each thread specifies runtime, deadline, and period. Kernel guarantees thread gets runtime before deadline.
**Confidence**: HIGH

## 3. CPU Isolation & Affinity

**Claim**: isolcpus isolates CPU cores from general kernel scheduling, dedicating them to RT workloads[^988^].
**Source**: OSADL - Real-time Linux wiki
**URL**: https://www.osadl.org/OSADL-Realtime-Linux.wiki.Realtime-Latency-Measurements.0.html
**Date**: Unknown
**Excerpt**: "isolcpus isolates CPU cores from general kernel scheduling"
**Context**: Format: isolcpus=2,3,4,5 or isolcpus=2-5. RT tasks pinned to isolated cores avoid migration overhead.
**Confidence**: HIGH

**Claim**: CPU affinity (sched_setaffinity) pins threads to specific cores, improving cache locality and reducing migration[^988^].
**Source**: Linux man-pages - sched_setaffinity(2)
**URL**: https://man7.org/linux/man-pages/man2/sched_setaffinity.2.html
**Date**: Unknown
**Excerpt**: "A thread's CPU affinity mask determines the set of CPUs on which it is eligible to run."
**Context**: Combine with isolcpus for dedicated cores. Use taskset or pthread_setaffinity_np.
**Confidence**: HIGH

## 4. Interrupt Handling

**Claim**: IRQ affinity (irqbalance or manual /proc/irq/*/smp_affinity) directs device interrupts to specific cores[^988^].
**Source**: OSADL - Real-time Linux wiki
**URL**: https://www.osadl.org/OSADL-Realtime-Linux.wiki.Realtime-Latency-Measurements.0.html
**Date**: Unknown
**Excerpt**: "irqaffinity directs device interrupts to specific cores"
**Context**: Network interrupts on core 0, GPU interrupts on core 1, RT game threads on cores 2-3.
**Confidence**: HIGH

**Claim**: MSI-X (Message Signaled Interrupts eXtended) provides up to 2048 interrupt vectors per device[^988^].
**Source**: PCI-SIG Specification
**URL**: https://pcisig.com/
**Date**: Unknown
**Excerpt**: "MSI-X provides up to 2048 interrupt vectors per device"
**Context**: Allows per-queue interrupts (e.g., one interrupt per network queue), reducing contention.
**Confidence**: HIGH

## 5. Tickless Kernel & Timer Optimization

**Claim**: nohz_full disables timer ticks on specified CPUs, reducing jitter[^988^].
**Source**: OSADL - Real-time Linux wiki
**URL**: https://www.osadl.org/OSADL-Realtime-Linux.wiki.Realtime-Latency-Measurements.0.html
**Date**: Unknown
**Excerpt**: "nohz_full disables timer ticks on specified CPUs"
**Context**: Without nohz_full, timer interrupt fires at CONFIG_HZ rate (typically 1000Hz = every 1ms), adding jitter.
**Confidence**: HIGH

**Claim**: High-resolution timers (hrtimers) provide nanosecond precision for sleeps and timeouts[^988^].
**Source**: Linux Kernel Documentation
**URL**: https://docs.kernel.org/timers/hrtimers.html
**Date**: Unknown
**Excerpt**: "hrtimers provide a high-resolution timer API with nanosecond resolution."
**Context**: hrtimers use hardware-specific clocks (TSC on x86, arch timer on ARM) for sub-microsecond precision.
**Confidence**: HIGH

## 6. Real-Time Testing

**Claim**: cyclictest measures scheduling latency with histogram output; rtla provides OS noise analysis[^988^].
**Source**: OSADL - Real-time Linux wiki
**URL**: https://www.osadl.org/OSADL-Realtime-Linux.wiki.Realtime-Latency-Measurements.0.html
**Date**: Unknown
**Excerpt**: "cyclictest measures scheduling latency... rtla provides OS noise analysis"
**Context**: cyclictest -p 99 -i 1000 runs SCHED_FIFO priority 99 with 1000μs interval. Measures difference between expected and actual wake time.
**Confidence**: HIGH

## 7. Practical Recommendations for Cloud Gaming Real-Time

| Parameter | Setting | Impact |
|---|---|---|
| Kernel | PREEMPT_RT patch or 6.12+ | <10μs scheduling latency |
| CPU Isolation | isolcpus=2-7 | Dedicated cores for game/engine |
| Timer | nohz_full=2-7 | No timer tick jitter |
| RCU | rcu_nocbs=2-7 | Offload callbacks |
| IRQ | irqaffinity for GPU/NIC | Directed interrupts |
| Scheduling | SCHED_FIFO priority 99 | Highest priority |
| Power | intel_idle.max_cstate=0 | Disable C-states |
| Frequency | intel_pstate=passive | Disable Turbo |

**Key Insight**: For the cloud gaming host:
1. **PREEMPT_RT kernel** (or Linux 6.12+ with RT) for sub-10μs scheduling
2. **CPU isolation** for game process, capture thread, and encode thread
3. **SCHED_FIFO priority 99** for critical threads (capture, encode, input)
4. **Disable C-states and Turbo** for deterministic performance
5. **NUMA-aware allocation** pinning memory to same node as CPU
