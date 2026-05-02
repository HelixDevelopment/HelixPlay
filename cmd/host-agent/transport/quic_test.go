package transport_test

import (
	"testing"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/transport"
)

func TestQUICStartReturnsError(t *testing.T) {
	q, err := transport.NewQUIC("localhost:4242")
	if err != nil {
		t.Fatalf("NewQUIC failed: %v", err)
	}
	// QUIC transport is not implemented; Start must return an error
	err = q.Start()
	if err == nil {
		t.Fatal("Expected Start to return error for unimplemented QUIC transport")
	}
	if q.IsConnected() {
		t.Error("Expected QUIC to not be connected after failed Start")
	}
}

func TestQUICSendWithoutConnection(t *testing.T) {
	q, _ := transport.NewQUIC("localhost:4242")
	// Do not call Start (it would fail anyway)
	err := q.SendDatagram([]byte("hello"))
	if err == nil {
		t.Fatal("Expected SendDatagram to fail when not connected")
	}
}

func TestQUICReceiveWithoutConnection(t *testing.T) {
	q, _ := transport.NewQUIC("localhost:4242")
	_, err := q.ReceiveDatagram()
	if err == nil {
		t.Fatal("Expected ReceiveDatagram to fail when not connected")
	}
}
