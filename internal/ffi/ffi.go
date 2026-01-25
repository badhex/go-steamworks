// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The go-steamworks Authors

package ffi

type InputDigitalActionData struct {
	State  bool
	Active bool
}

type InputAnalogActionData struct {
	Mode   int32
	X      float32
	Y      float32
	Active bool
}

type InputMotionData struct {
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

func CallInputDigitalActionData(fn uintptr, self uintptr, inputHandle uint64, actionHandle uint64) InputDigitalActionData {
	return InputDigitalActionData{}
}

func CallInputAnalogActionData(fn uintptr, self uintptr, inputHandle uint64, actionHandle uint64) InputAnalogActionData {
	return InputAnalogActionData{}
}

func CallInputMotionData(fn uintptr, self uintptr, inputHandle uint64) InputMotionData {
	return InputMotionData{}
}
