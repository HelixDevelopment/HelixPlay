// vasic-digital/Memory/pkg/memfd/ringbuffer_test.go
package memfd_test

import (
    "testing"
    "vasic-digital/Memory/pkg/memfd"
)

func TestRingBufferWriteRead(t *testing.T) {
    rb, err := memfd.NewPSC(1024) // 1KB buffer
    if err != nil {
        t.Fatalf("Failed to create ring buffer: %v", err)
    }
    defer rb.Close()

    data := []byte{1, 2, 3, 4, 5}
    n, err := rb.Write(data)
    if err != nil || n != len(data) {
        t.Fatalf("Write failed: %v, n=%d", err, n)
    }

    buf := make([]byte, len(data))
    n, err = rb.Read(buf)
    if err != nil || n != len(data) {
        t.Fatalf("Read failed: %v, n=%d", err, n)
    }
}
