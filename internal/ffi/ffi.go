// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The go-steamworks Authors

package ffi

/*
#include <stdbool.h>
#include <stdint.h>

typedef struct {
	bool state;
	bool active;
} InputDigitalActionData;

typedef struct {
	int32_t mode;
	float x;
	float y;
	bool active;
} InputAnalogActionData;

typedef struct {
	float rotQuatX;
	float rotQuatY;
	float rotQuatZ;
	float rotQuatW;
	float posAccelX;
	float posAccelY;
	float posAccelZ;
	float rotVelX;
	float rotVelY;
	float rotVelZ;
} InputMotionData;

typedef InputDigitalActionData (*InputDigitalActionDataFn)(void*, uint64_t, uint64_t);
typedef InputAnalogActionData (*InputAnalogActionDataFn)(void*, uint64_t, uint64_t);
typedef InputMotionData (*InputMotionDataFn)(void*, uint64_t);

static inline InputDigitalActionData callInputDigitalActionData(void* fn, void* self, uint64_t inputHandle, uint64_t actionHandle) {
	return ((InputDigitalActionDataFn)fn)(self, inputHandle, actionHandle);
}

static inline InputAnalogActionData callInputAnalogActionData(void* fn, void* self, uint64_t inputHandle, uint64_t actionHandle) {
	return ((InputAnalogActionDataFn)fn)(self, inputHandle, actionHandle);
}

static inline InputMotionData callInputMotionData(void* fn, void* self, uint64_t inputHandle) {
	return ((InputMotionDataFn)fn)(self, inputHandle);
}
*/
import "C"

import "unsafe"

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
	result := C.callInputDigitalActionData(
		unsafe.Pointer(fn),
		unsafe.Pointer(self),
		C.uint64_t(inputHandle),
		C.uint64_t(actionHandle),
	)
	return InputDigitalActionData{
		State:  result.state != C.bool(false),
		Active: result.active != C.bool(false),
	}
}

func CallInputAnalogActionData(fn uintptr, self uintptr, inputHandle uint64, actionHandle uint64) InputAnalogActionData {
	result := C.callInputAnalogActionData(
		unsafe.Pointer(fn),
		unsafe.Pointer(self),
		C.uint64_t(inputHandle),
		C.uint64_t(actionHandle),
	)
	return InputAnalogActionData{
		Mode:   int32(result.mode),
		X:      float32(result.x),
		Y:      float32(result.y),
		Active: result.active != C.bool(false),
	}
}

func CallInputMotionData(fn uintptr, self uintptr, inputHandle uint64) InputMotionData {
	result := C.callInputMotionData(
		unsafe.Pointer(fn),
		unsafe.Pointer(self),
		C.uint64_t(inputHandle),
	)
	return InputMotionData{
		RotQuatX:  float32(result.rotQuatX),
		RotQuatY:  float32(result.rotQuatY),
		RotQuatZ:  float32(result.rotQuatZ),
		RotQuatW:  float32(result.rotQuatW),
		PosAccelX: float32(result.posAccelX),
		PosAccelY: float32(result.posAccelY),
		PosAccelZ: float32(result.posAccelZ),
		RotVelX:   float32(result.rotVelX),
		RotVelY:   float32(result.rotVelY),
		RotVelZ:   float32(result.rotVelZ),
	}
}
