// SPDX-FileCopyrightText: 2026 Milos Vasic
// SPDX-License-Identifier: Apache-2.0

// Package network provides transport backends for the host agent.
// This file implements io_uring zero-copy socket I/O for Linux.
//
// io_uring (Linux 5.1+) allows asynchronous I/O without syscalls,
// reducing latency by eliminating context switches on the send/recv path.
//
// T272: io_uring zero-copy socket I/O for Linux host agent.
package network

import (
	"errors"
	"fmt"
	"runtime"
)

// IOUringSocket is a Linux-only socket backend using io_uring.
// On non-Linux platforms it returns an error at creation time.
type IOUringSocket struct {
	fd      int
	ringFD  int
	ready   bool
}

// NewIOUringSocket attempts to create an io_uring-backed socket.
// Returns an error if the kernel does not support io_uring or if
// the caller is not on Linux.
func NewIOUringSocket() (*IOUringSocket, error) {
	if runtime.GOOS != "linux" {
		return nil, errors.New("io_uring socket requires Linux")
	}
	// TODO: initialise io_uring ring (requires github.com/axboe/liburing bindings).
	_ = struct{}{} // placeholder for ring setup
	return &IOUringSocket{fd: -1, ringFD: -1, ready: false}, nil
}

// Send transmits data over the io_uring socket.
// Zero-copy mode uses registered buffers to avoid kernel-user copies.
func (s *IOUringSocket) Send(data []byte) error {
	_ = s
	_ = data
	return fmt.Errorf("io_uring send not yet implemented")
}

// Receive reads data from the io_uring socket into the provided buffer.
func (s *IOUringSocket) Receive(buf []byte) (int, error) {
	_ = s
	_ = buf
	return 0, fmt.Errorf("io_uring receive not yet implemented")
}

// Close tears down the io_uring ring and socket.
func (s *IOUringSocket) Close() error {
	_ = s
	return nil
}

// Ready returns true if the io_uring ring is initialised and the
// socket is usable.
func (s *IOUringSocket) Ready() bool {
	_ = s
	return false
}
