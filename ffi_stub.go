// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The go-steamworks Authors

//go:build !amd64

package steamworks

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
	return ffi_InputDigitalActionData{}
}

func callInputAnalogActionData(fn uintptr, self uintptr, inputHandle uint64, actionHandle uint64) ffi_InputAnalogActionData {
	return ffi_InputAnalogActionData{}
}

func callInputMotionData(fn uintptr, self uintptr, inputHandle uint64) ffi_InputMotionData {
	return ffi_InputMotionData{}
}
