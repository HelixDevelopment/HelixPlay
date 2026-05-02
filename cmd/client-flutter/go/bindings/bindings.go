package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"sync"
	"time"
	"unsafe"

	"github.com/HelixDevelopment/HelixPlay/pkg/core/catalog"
	"github.com/HelixDevelopment/HelixPlay/pkg/core/discovery"
	"github.com/HelixDevelopment/HelixPlay/pkg/core/input"
	"github.com/HelixDevelopment/HelixPlay/pkg/core/streaming"
)

var (
	bindingsMu        sync.Mutex
	discoveryClients  = make(map[int]*discovery.Client)
	streamControllers = make(map[int]*streaming.SessionController)
	inputManager      = input.NewManager()
	catalogClients    = make(map[int]*catalog.Client)
	nextHandle        int
)

func getNextHandle() int {
	bindingsMu.Lock()
	defer bindingsMu.Unlock()
	nextHandle++
	return nextHandle
}

//export HelixVersion
func HelixVersion() *C.char {
	return C.CString("1.0.0")
}

//export HelixFreeString
func HelixFreeString(s *C.char) {
	C.free(unsafe.Pointer(s))
}

//export HelixDiscoveryStart
func HelixDiscoveryStart(rendezvousAddr, tenantID, region *C.char) int {
	handle := getNextHandle()
	addr := C.GoString(rendezvousAddr)
	tID := C.GoString(tenantID)
	reg := C.GoString(region)

	dc := discovery.NewClient(addr, tID, reg)
	ctx, cancel := context.WithCancel(context.Background())
	if err := dc.Start(ctx); err != nil {
		cancel()
		return -1
	}

	bindingsMu.Lock()
	discoveryClients[handle] = dc
	bindingsMu.Unlock()

	go func() {
		<-ctx.Done()
	}()
	return handle
}

//export HelixDiscoveryStop
func HelixDiscoveryStop(handle int) {
	bindingsMu.Lock()
	dc, ok := discoveryClients[handle]
	delete(discoveryClients, handle)
	bindingsMu.Unlock()
	if ok {
		dc.Stop()
	}
}

//export HelixDiscoveryHostCount
func HelixDiscoveryHostCount(handle int) int {
	bindingsMu.Lock()
	dc, ok := discoveryClients[handle]
	bindingsMu.Unlock()
	if !ok {
		return 0
	}
	return len(dc.Hosts())
}

//export HelixStreamingConnect
func HelixStreamingConnect(hostAddress, clientID, userID, authToken, gameID *C.char) int {
	handle := getNextHandle()
	cfg := streaming.Config{
		HostAddress:     C.GoString(hostAddress),
		ClientID:        C.GoString(clientID),
		UserID:          C.GoString(userID),
		AuthToken:       C.GoString(authToken),
		RequestedGameID: C.GoString(gameID),
	}
	sc := streaming.NewSessionController(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := sc.Connect(ctx); err != nil {
		return -1
	}

	bindingsMu.Lock()
	streamControllers[handle] = sc
	bindingsMu.Unlock()
	return handle
}

//export HelixStreamingDisconnect
func HelixStreamingDisconnect(handle int) {
	bindingsMu.Lock()
	sc, ok := streamControllers[handle]
	delete(streamControllers, handle)
	bindingsMu.Unlock()
	if ok {
		sc.Disconnect()
	}
}

//export HelixStreamingState
func HelixStreamingState(handle int) *C.char {
	bindingsMu.Lock()
	sc, ok := streamControllers[handle]
	bindingsMu.Unlock()
	if !ok {
		return C.CString("idle")
	}
	return C.CString(string(sc.State()))
}

//export HelixInputUpdate
func HelixInputUpdate(id int, connected int, buttonMask uint32, leftStickX, leftStickY, rightStickX, rightStickY int16) {
	st := input.ControllerState{
		ID:          id,
		Connected:   connected != 0,
		ButtonMask:  buttonMask,
		LeftStickX:  leftStickX,
		LeftStickY:  leftStickY,
		RightStickX: rightStickX,
		RightStickY: rightStickY,
	}
	inputManager.UpdateState(st)
}

//export HelixCatalogOpen
func HelixCatalogOpen(catalogAddr *C.char) int {
	handle := getNextHandle()
	addr := C.GoString(catalogAddr)
	cc, err := catalog.NewClient(addr)
	if err != nil {
		return -1
	}
	bindingsMu.Lock()
	catalogClients[handle] = cc
	bindingsMu.Unlock()
	return handle
}

//export HelixCatalogClose
func HelixCatalogClose(handle int) {
	bindingsMu.Lock()
	cc, ok := catalogClients[handle]
	delete(catalogClients, handle)
	bindingsMu.Unlock()
	if ok {
		cc.Close()
	}
}

func main() {}
