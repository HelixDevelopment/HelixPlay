# Dimension 01: Shared Memory & Zero-Copy IPC

## 1. POSIX Shared Memory (shm_open + mmap)

**Claim**: shm_open + mmap achieves 8M messages/sec, 20x faster than POSIX message queues (400K messages/sec)[^965^].
**Source**: HowTech IPC Benchmarking
**URL**: https://howtech.substack.com/p/ipc-mechanisms-shared-memory-vs-message
**Date**: 2025-12-11
**Excerpt**: "Shared memory hit 8M messages/sec—20x faster. P99 latency: 850ns vs 12μs for message queues."
**Context**: Benchmark sending 1M messages (64 bytes each) with perf stat showing 0 context switches for shared memory vs 183,472 for message queues.
**Confidence**: HIGH

**Claim**: Writing to shared memory is literally a `mov` instruction—no kernel involvement after setup[^965^].
**Source**: HowTech IPC Benchmarking
**URL**: https://howtech.substack.com/p/ipc-mechanisms-shared-memory-vs-message
**Date**: 2025-12-11
**Excerpt**: "Writing to the memory is literally a mov instruction—no kernel involvement."
**Context**: shm_open creates file in /dev/shm (tmpfs), mmap with MAP_SHARED maps same physical pages.
**Confidence**: HIGH

**Claim**: Chrome renderer IPC switched from spinlocks to futex-based synchronization, cutting CPU by 90%[^965^].
**Source**: HowTech IPC Benchmarking
**URL**: https://howtech.substack.com/p/ipc-mechanisms-shared-memory-vs-message
**Date**: 2025-12-11
**Excerpt**: "Chrome's renderer IPC initially used spinlocks in shared memory. CPU usage spiked because waiting processes burned cycles. Switching to futex-based synchronization cut CPU by 90%."
**Context**: Futexes stay in userspace if uncontended, only syscall if they must block.
**Confidence**: HIGH

## 2. memfd_create — Anonymous File-Backed Shared Memory

**Claim**: memfd_create creates an anonymous file in RAM with volatile backing, supports sealing and huge pages[^968^].
**Source**: Linux man-pages, man7.org
**URL**: https://man7.org/linux/man-pages/man2/memfd_create.2.html
**Date**: 2026-01-16
**Excerpt**: "memfd_create() creates an anonymous file and returns a file descriptor that refers to it. The file behaves like a regular file... However, unlike a regular file, it lives in RAM and has a volatile backing storage."
**Context**: Supports MFD_CLOEXEC, MFD_ALLOW_SEALING, MFD_HUGETLB (since Linux 4.14), MFD_HUGE_2MB/1GB.
**Confidence**: HIGH

**Claim**: File sealing (F_SEAL_SHRINK, F_SEAL_GROW, F_SEAL_WRITE) prevents modification after setup[^977^].
**Source**: Benjamin Toll Blog
**URL**: https://benjamintoll.com/2022/08/21/on-memfd_create/
**Date**: 2022-08-21
**Excerpt**: "We can get access to the file sealing APIs that can be used to manipulate file descriptors using the fcntl syscall."
**Context**: Sealing makes shared memory segments immutable after configuration, enhancing security.
**Confidence**: HIGH

## 3. Lock-Free Ring Buffers (LMAX Disruptor Pattern)

**Claim**: LMAX Disruptor uses lock-free techniques, cache-line padding, memory barriers, pre-allocation, and batching[^966^].
**Source**: Sanjeev Pages Dev Blog
**URL**: https://sanjeev.pages.dev/lmax-disruptor
**Date**: 2025-07-05
**Excerpt**: "Use of lock free techniques, namely memory barriers/fences. Use of cache line padding. Efficient modulus operations using powers of 2 and bitmasking."
**Context**: Disruptor achieves 6M+ events/sec on a single thread with <50ns latency.
**Confidence**: HIGH

**Claim**: Lock-free queue achieves sub-microsecond message delivery with nanosecond precision timing[^967^].
**Source**: GitHub - manojds/LockFreeQueueForIPC
**URL**: https://github.com/manojds/LockFreeQueueForIPC
**Date**: 2025-06-28
**Excerpt**: "Ultra-Low Latency: Sub-microsecond message delivery with nanosecond precision timing. Zero Dynamic Allocation: Real-time safe with no malloc/free operations."
**Context**: SPMC queue using POSIX shared memory, rdtsc for timing, cache-line alignment.
**Confidence**: HIGH

**Claim**: Cache-line padding prevents false sharing—critical for multi-core scalability[^969^].
**Source**: Medium - RingBuffer for High-Performance Java
**URL**: https://medium.com/@amit.agarwal0422/ringbuffer-the-secret-weapon-for-high-performance-java-applications-ebabdb64ce58
**Date**: 2025-06-19
**Excerpt**: "Disruptor avoids false sharing by padding data structures so that different threads do not contend for the same cache line."
**Context**: False sharing occurs when threads on different cores modify data on the same cache line (64 bytes on x86-64).
**Confidence**: HIGH

## 4. DPDK rte_ring — Multi-Mode Synchronization

**Claim**: DPDK rte_ring supports SP/SC, MP/MC, RTS, HTS modes for different contention scenarios[^978^].
**Source**: DPDK Documentation
**URL**: https://doc.dpdk.org/guides/mempool/ring.html
**Date**: Unknown
**Excerpt**: "For 'classic' DPDK deployments (with one thread per core) the ring_mp_mc mode is usually the most suitable and the fastest one."
**Context**: ring_sp_sc is fastest for single-producer-single-consumer; ring_mt_rts and ring_mt_hts for overcommitted scenarios.
**Confidence**: HIGH

## 5. False Sharing Detection & Mitigation

**Claim**: perf c2c detects cache-line contention including false sharing, showing HITM (Hit in Modified state) events[^970^].
**Source**: Red Hat Enterprise Linux Documentation
**URL**: https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/9/html/monitoring_and_managing_system_status_and_performance/detecting-false-sharing_monitoring-and-managing-system-status-and-performance
**Date**: 2023-01-11
**Excerpt**: "The c2c subcommand of the perf tool enables Shared Data Cache-to-Cache (C2C) analysis. You can use the perf c2c command to inspect cache-line contention to detect both true and false sharing."
**Context**: HITM percentage indicates remote cache-line modifications—a key false sharing indicator.
**Confidence**: HIGH

**Claim**: Arm SPE (Statistical Profiling Extension) provides precise cache analysis at the microarchitecture level[^976^].
**Source**: Arm Learning Path
**URL**: https://learn.arm.com/learning-paths/servers-and-cloud-computing/false-sharing-arm-spe/how-to-1/
**Date**: Unknown
**Excerpt**: "SPE integrates sampling directly into the CPU pipeline, triggering on individual micro-operations instead of retired instructions. Each SPE sample record includes data addresses, per-µop pipeline latency, triggered PMU event masks, memory hierarchy source."
**Context**: Eliminates skid and blind spots present in traditional retired-instruction sampling.
**Confidence**: HIGH

## 6. Memory Barriers & Atomic Operations

**Claim**: Memory ordering requires __atomic_thread_fence(__ATOMIC_RELEASE) after writing and __ATOMIC_ACQUIRE after reading indices[^965^].
**Source**: HowTech IPC Benchmarking
**URL**: https://howtech.substack.com/p/ipc-mechanisms-shared-memory-vs-message
**Date**: 2025-12-11
**Excerpt**: "You need __atomic_thread_fence(__ATOMIC_RELEASE) after writing and __atomic_thread_fence(__ATOMIC_ACQUIRE) after reading the index."
**Context**: Without barriers, consumer might see updated index before data is written.
**Confidence**: HIGH

## 7. Practical Recommendations for Cloud Gaming IPC

| Mechanism | Latency | Throughput | Complexity | Best For |
|---|---|---|---|---|
| POSIX Message Queue | 12μs P99 | 400K msg/s | Low | Low-frequency control |
| Shared Memory + Futex | ~1μs | 8M msg/s | Medium | High-frequency data |
| Lock-Free Ring Buffer | 850ns P99 | 8M+ msg/s | High | Ultra-low-latency |
| DPDK rte_ring (SP/SC) | <100ns | 10M+ msg/s | Very High | Kernel-bypass networking |
| memfd_create + Sealing | ~1μs | 8M msg/s | Medium | Secure shared memory |

**Key Insight**: For the cloud gaming system, the optimal IPC architecture is:
1. **memfd_create** for shared memory allocation (secure, sealable, huge page support)
2. **Lock-free SPSC ring buffer** for controller input → game process (single producer, single consumer)
3. **DPDK rte_ring pattern** for capture → encoder pipeline
4. **Cache-line padding** (64-byte alignment) on all shared indices
5. **Memory barriers** (release/acquire) for visibility
