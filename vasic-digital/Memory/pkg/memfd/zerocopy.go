// vasic-digital/Memory/pkg/memfd/zerocopy.go
package memfd

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// ZerocopyIPC provides zero-copy IPC helpers for sending/receiving
// memfd-backed buffers across process boundaries (via Unix sockets + SCM_RIGHTS).
type ZerocopyIPC struct {
	fd int // Unix socket fd for passing memfd fds
}

// NewZerocopyIPC creates an IPC handle using a Unix socketpair.
// The returned socket can be used to send memfd fds to another process.
func NewZerocopyIPC() (*ZerocopyIPC, error) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, fmt.Errorf("socketpair: %w", err)
	}
	return &ZerocopyIPC{fd: fds[0]}, nil
}

// SendMemfd sends a memfd file descriptor over the IPC channel.
// The receiver can mmap the received fd for zero-copy access.
func (z *ZerocopyIPC) SendMemfd(memfdFD int) error {
	// Build SCM_RIGHTS control message to pass the fd
	buf := []byte{1} // dummy payload
	iov := &syscall.Iovec{
		Base: &buf[0],
		Len:  uint64(len(buf)),
	}

	// Allocate control buffer for SCM_RIGHTS (one fd = 4 bytes)
	cmsgLen := syscall.CmsgLen(4)
	cmsgBuf := make([]byte, cmsgLen)

	// Build control message
	hdr := (*syscall.Cmsghdr)(unsafe.Pointer(&cmsgBuf[0]))
	hdr.Len = uint64(cmsgLen)
	hdr.Level = syscall.SOL_SOCKET
	hdr.Type = syscall.SCM_RIGHTS

	// Copy fd into control message data area
	dataStart := cmsgLen - 4
	*(*int32)(unsafe.Pointer(&cmsgBuf[dataStart])) = int32(memfdFD)

	msg := &syscall.Msghdr{
		Iov:       iov,
		Iovlen:    1,
		Control:   &cmsgBuf[0],
		Controllen: uint64(cmsgLen),
	}

	_, _, errno := syscall.Syscall(syscall.SYS_SENDMSG, uintptr(z.fd), uintptr(unsafe.Pointer(msg)), 0)
	if errno != 0 {
		return fmt.Errorf("sendmsg: %d", errno)
	}
	return nil
}

// RecvMemfd receives a memfd file descriptor from the IPC channel.
// Returns the received fd, which can be mmap'd for zero-copy access.
func (z *ZerocopyIPC) RecvMemfd() (int, error) {
	buf := make([]byte, 1)
	iov := &syscall.Iovec{
		Base: &buf[0],
		Len:  uint64(len(buf)),
	}

	cmsgLen := syscall.CmsgLen(4)
	cmsgBuf := make([]byte, cmsgLen)

	msg := &syscall.Msghdr{
		Iov:        iov,
		Iovlen:     1,
		Control:     &cmsgBuf[0],
		Controllen:  uint64(cmsgLen),
	}

	_, _, errno := syscall.Syscall(syscall.SYS_RECVMSG, uintptr(z.fd), uintptr(unsafe.Pointer(msg)), 0)
	if errno != 0 {
		return -1, fmt.Errorf("recvmsg: %d", errno)
	}

	// Extract fd from control message
	hdr := (*syscall.Cmsghdr)(unsafe.Pointer(&cmsgBuf[0]))
	if hdr.Level == syscall.SOL_SOCKET && hdr.Type == syscall.SCM_RIGHTS {
		dataStart := cmsgLen - 4
		fd := int(*(*int32)(unsafe.Pointer(&cmsgBuf[dataStart])))
		return fd, nil
	}
	return -1, fmt.Errorf("no SCM_RIGHTS in control message")
}

// Close closes the IPC channel.
func (z *ZerocopyIPC) Close() error {
	if z.fd >= 0 {
		err := syscall.Close(z.fd)
		z.fd = -1
		return err
	}
	return nil
}

// MmapMemfd maps a received memfd into the current address space.
// Returns a slice backed by the shared memory region.
func MmapMemfd(fd int, size uint64) ([]byte, error) {
	addr, _, errno := syscall.Syscall6(
		syscall.SYS_MMAP,
		0,
		uintptr(size),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
		uintptr(fd),
		0,
	)
	if errno != 0 {
		return nil, fmt.Errorf("mmap: %d", errno)
	}
	return (*[1 << 30]byte)(unsafe.Pointer(addr))[:int(size):int(size)], nil
}

// MunmapMemfd unmaps a previously mapped memfd region.
func MunmapMemfd(data []byte) error {
	// Use unix.Munmap since it handles the slice properly
	return unix.Munmap(data)
}
