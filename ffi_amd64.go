// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The go-steamworks Authors

//go:build amd64

package steamworks

import (
	"math"
	"unsafe"

	"github.com/ebitengine/purego"
)

type ffi_InputDigitalActionData struct {
	State  bool
	Active bool
}

type ffi_InputAnalogActionData struct {
	Mode   int32
	X      float32
	Y      float32
	Active bool
}

type ffi_InputMotionData struct {
	RotQuatX  float32
	RotQuatY  float32
	RotQuatZ  float32
	RotQuatW  float32
	PosAccelX float32
	PosAccelY float32
	PosAccelZ float32
	RotVelX   float32
	RotVelY   float32
	RotVelZ   float32
}

func callInputDigitalActionData(fn uintptr, self uintptr, inputHandle uint64, actionHandle uint64) ffi_InputDigitalActionData {
	if fn == 0 || self == 0 {
		return ffi_InputDigitalActionData{}
	}
	r1, _, _ := purego.SyscallN(fn, self, uintptr(inputHandle), uintptr(actionHandle))
	return ffi_InputDigitalActionData{
		State:  byte(r1&0xFF) != 0,
		Active: byte((r1>>8)&0xFF) != 0,
	}
}

func callInputAnalogActionData(fn uintptr, self uintptr, inputHandle uint64, actionHandle uint64) ffi_InputAnalogActionData {
	if fn == 0 || self == 0 {
		return ffi_InputAnalogActionData{}
	}
	r1, r2, _ := purego.SyscallN(fn, self, uintptr(inputHandle), uintptr(actionHandle))
	return ffi_InputAnalogActionData{
		Mode:   int32(uint32(r1)),
		X:      math.Float32frombits(uint32(r1 >> 32)),
		Y:      math.Float32frombits(uint32(r2)),
		Active: byte((r2>>32)&0xFF) != 0,
	}
}

func callInputMotionData(fn uintptr, self uintptr, inputHandle uint64) ffi_InputMotionData {
	if fn == 0 || self == 0 {
		return ffi_InputMotionData{}
	}
	// InputMotionData_t is larger than register return sizes on amd64 ABIs, so pass
	// an explicit output buffer as the hidden struct-return argument.
	var out ffi_InputMotionData
	purego.SyscallN(fn, uintptr(unsafe.Pointer(&out)), self, uintptr(inputHandle))
	return out
}
