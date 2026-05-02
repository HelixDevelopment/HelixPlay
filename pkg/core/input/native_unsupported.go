//go:build !(linux || windows)

package input

import "fmt"

// NativeCapture is unavailable on this platform.
type NativeCapture struct{}

// OpenNativeCapture returns an error on unsupported platforms.
func OpenNativeCapture(id int) (*NativeCapture, error) {
	return nil, fmt.Errorf("native capture unsupported on this platform")
}

// Read always returns an error.
func (nc *NativeCapture) Read() (ControllerState, error) {
	return ControllerState{}, fmt.Errorf("native capture unsupported on this platform")
}

// Close is a no-op.
func (nc *NativeCapture) Close() error {
	return nil
}
