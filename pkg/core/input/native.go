//go:build linux || windows

package input

import (
	"fmt"

	"github.com/0xcafed00d/joystick"
)

// NativeCapture polls a local joystick device.
type NativeCapture struct {
	id int
	js joystick.Joystick
}

// OpenNativeCapture opens the joystick with the given id.
func OpenNativeCapture(id int) (*NativeCapture, error) {
	js, err := joystick.Open(id)
	if err != nil {
		return nil, fmt.Errorf("open joystick %d: %w", id, err)
	}
	return &NativeCapture{id: id, js: js}, nil
}

// Read reads the current state from the native device.
func (nc *NativeCapture) Read() (ControllerState, error) {
	st, err := nc.js.Read()
	if err != nil {
		return ControllerState{}, fmt.Errorf("read joystick: %w", err)
	}

	state := ControllerState{
		ID:         nc.id,
		Connected:  true,
		ButtonMask: st.Buttons,
	}

	if len(st.AxisData) > 0 {
		state.LeftStickX = int16(st.AxisData[0])
	}
	if len(st.AxisData) > 1 {
		state.LeftStickY = int16(st.AxisData[1])
	}
	if len(st.AxisData) > 2 {
		state.RightStickX = int16(st.AxisData[2])
	}
	if len(st.AxisData) > 3 {
		state.RightStickY = int16(st.AxisData[3])
	}
	if len(st.AxisData) > 4 {
		state.LeftTrigger = uint8((st.AxisData[4] + 32767) / 256)
	}
	if len(st.AxisData) > 5 {
		state.RightTrigger = uint8((st.AxisData[5] + 32767) / 256)
	}

	return state, nil
}

// Close releases the native device.
func (nc *NativeCapture) Close() error {
	nc.js.Close()
	return nil
}
