// vasic-digital/Memory/pkg/memfd/ringbuffer.go
package memfd

import (
	"fmt"
	"sync/atomic"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Cache-line size for padding (128-byte rule for modern CPUs)
const cacheLineSize = 128

// pscHeader sits at the start of the memfd region.
// All fields are cache-line padded to avoid false sharing.
type pscHeader struct {
	writePos uint64
	_pad0    [cacheLineSize - 8]byte
	readPos  uint64
	_pad1    [cacheLineSize - 8]byte
	size      uint64
	_pad2    [cacheLineSize - 8]byte
}

// PSC is a lock-free producer-single-consumer ring buffer backed by memfd.
// The producer and consumer must run on different goroutines.
// Only one goroutine may call Write (producer) and only one may call Read (consumer).
type PSC struct {
	fd   int
	addr uintptr
	hdr  *pscHeader
	data []byte // slice into the mmap'd region after the header
}

// NewPSC creates a lock-free PSC ring buffer of the given size (rounded up to page boundary).
func NewPSC(size uint64) (*PSC, error) {
	// Align size to page boundary
	pageSize := uint64(4096)
	if size%pageSize != 0 {
		size = ((size / pageSize) + 1) * pageSize
	}

	// memfd_create (Linux-specific)
	fd, err := unix.MemfdCreate("helixplay-ringbuf", 0)
	if err != nil {
		return nil, fmt.Errorf("memfd_create: %w", err)
	}

	// Total mmap region = header + data
	totalSize := uint64(unsafe.Sizeof(pscHeader{})) + size

	// Set size
	if err := syscall.Ftruncate(fd, int64(totalSize)); err != nil {
		syscall.Close(fd)
		return nil, fmt.Errorf("ftruncate: %w", err)
	}

	// mmap shared (so zero-copy IPC works across fork)
	addr, _, errno := syscall.Syscall6(
		syscall.SYS_MMAP,
		0,
		uintptr(totalSize),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
		uintptr(fd),
		0,
	)
	if errno != 0 {
		syscall.Close(fd)
		return nil, fmt.Errorf("mmap: %d", errno)
	}

	hdr := (*pscHeader)(unsafe.Pointer(addr))
	hdr.size = size

	// data slice starts after the header
	dataAddr := addr + uintptr(unsafe.Sizeof(pscHeader{}))
	data := (*[1 << 30]byte)(unsafe.Pointer(dataAddr))[:int(size):int(size)]

	return &PSC{
		fd:   fd,
		addr: addr,
		hdr:  hdr,
		data: data,
	}, nil
}

// Write writes data into the ring buffer. Returns number of bytes written.
// This is the producer side (lock-free, single producer assumed).
func (rb *PSC) Write(p []byte) (int, error) {
	written := 0
	for len(p) > 0 {
		wp := atomic.LoadUint64(&rb.hdr.writePos)
		rp := atomic.LoadUint64(&rb.hdr.readPos)

		// Available space (handling wrap-around)
		var avail uint64
		if wp >= rp {
			avail = rb.hdr.size - (wp - rp) - 1
		} else {
			avail = rp - wp - 1
		}
		if avail == 0 {
			break // buffer full
		}

		toWrite := uint64(len(p))
		if toWrite > avail {
			toWrite = avail
		}
		if toWrite > rb.hdr.size-wp {
			toWrite = rb.hdr.size - wp
		}

		// Copy into ring buffer
		copy(rb.data[wp:wp+toWrite], p[:toWrite])
		p = p[toWrite:]
		written += int(toWrite)

		// Update writePos with atomic store
		wp = (wp + toWrite) % rb.hdr.size
		atomic.StoreUint64(&rb.hdr.writePos, wp)
	}

	if written == 0 && len(p) > 0 {
		return 0, fmt.Errorf("ring buffer full")
	}
	return written, nil
}

// Read reads data from the ring buffer. Returns number of bytes read.
// This is the consumer side (lock-free, single consumer assumed).
func (rb *PSC) Read(p []byte) (int, error) {
	read := 0
	for len(p) > 0 {
		wp := atomic.LoadUint64(&rb.hdr.writePos)
		rp := atomic.LoadUint64(&rb.hdr.readPos)

		if rp == wp {
			break // buffer empty
		}

		// Available data
		var avail uint64
		if wp >= rp {
			avail = wp - rp
		} else {
			avail = rb.hdr.size - rp
		}
		if avail == 0 {
			break
		}

		toRead := uint64(len(p))
		if toRead > avail {
			toRead = avail
		}

		copy(p[:toRead], rb.data[rp:rp+toRead])
		p = p[toRead:]
		read += int(toRead)

		// Update readPos
		rp = (rp + toRead) % rb.hdr.size
		atomic.StoreUint64(&rb.hdr.readPos, rp)
	}

	if read == 0 {
		return 0, fmt.Errorf("ring buffer empty")
	}
	return read, nil
}

// Close unmaps the region and closes the memfd.
func (rb *PSC) Close() error {
	var firstErr error
	if rb.addr != 0 {
		totalSize := uintptr(rb.hdr.size + uint64(unsafe.Sizeof(pscHeader{})))
		if _, _, errno := syscall.Syscall(syscall.SYS_MUNMAP, rb.addr, totalSize, 0); errno != 0 {
			firstErr = fmt.Errorf("munmap: %d", errno)
		}
		rb.addr = 0
	}
	if rb.fd >= 0 {
		err := syscall.Close(rb.fd)
		rb.fd = -1
		if firstErr == nil {
			return err
		}
		return firstErr
	}
	return firstErr
}

// FD returns the underlying memfd file descriptor (for zero-copy IPC via fork/sendmsg).
func (rb *PSC) FD() int {
	return rb.fd
}

// Size returns the usable buffer size (excluding header).
func (rb *PSC) Size() uint64 {
	return rb.hdr.size
}
