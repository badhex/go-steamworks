// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The go-steamworks Authors

package ffi

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

const ffiTypeStruct = 13

type ffiStatus uint32

const (
	ffiOK         ffiStatus = 0
	ffiBadTypedef ffiStatus = 1
	ffiBadABI     ffiStatus = 2
)

type ffiType struct {
	size      uintptr
	alignment uint16
	ffiType   uint16
	elements  **ffiType
}

type ffiCif struct {
	abi        uint32
	nargs      uint32
	argTypes   **ffiType
	rtype      *ffiType
	bytes      uint32
	flags      uint32
	nfixedargs uint32
}

type callSignature struct {
	once     sync.Once
	err      error
	cif      ffiCif
	argTypes []*ffiType
}

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

var (
	libffiOnce     sync.Once
	libffiErr      error
	libffi         uintptr
	ffiPrepCif     uintptr
	ffiCall        uintptr
	ffiTypeVoid    *ffiType
	ffiTypeUint8   *ffiType
	ffiTypeSint32  *ffiType
	ffiTypeUint64  *ffiType
	ffiTypeFloat   *ffiType
	ffiTypePointer *ffiType
	defaultABI     uint32

	structTypesOnce      sync.Once
	inputDigitalType     ffiType
	inputDigitalElements []*ffiType
	inputAnalogType      ffiType
	inputAnalogElements  []*ffiType
	inputMotionType      ffiType
	inputMotionElements  []*ffiType

	digitalSig callSignature
	analogSig  callSignature
	motionSig  callSignature

	callInputDigitalOverride func(fn uintptr, self uintptr, inputHandle uint64, actionHandle uint64) InputDigitalActionData
	callInputAnalogOverride  func(fn uintptr, self uintptr, inputHandle uint64, actionHandle uint64) InputAnalogActionData
	callInputMotionOverride  func(fn uintptr, self uintptr, inputHandle uint64) InputMotionData
)

func CallInputDigitalActionData(fn uintptr, self uintptr, inputHandle uint64, actionHandle uint64) InputDigitalActionData {
	if callInputDigitalOverride != nil {
		return callInputDigitalOverride(fn, self, inputHandle, actionHandle)
	}
	if err := ensureSignature(&digitalSig, &inputDigitalType, []*ffiType{ffiTypePointer, ffiTypeUint64, ffiTypeUint64}); err != nil {
		panic(err)
	}
	var result InputDigitalActionData
	selfArg := unsafe.Pointer(self)
	inputArg := inputHandle
	actionArg := actionHandle
	args := []unsafe.Pointer{
		unsafe.Pointer(&selfArg),
		unsafe.Pointer(&inputArg),
		unsafe.Pointer(&actionArg),
	}
	callWithArgs(&digitalSig.cif, fn, unsafe.Pointer(&result), args)
	runtime.KeepAlive(args)
	return result
}

func CallInputAnalogActionData(fn uintptr, self uintptr, inputHandle uint64, actionHandle uint64) InputAnalogActionData {
	if callInputAnalogOverride != nil {
		return callInputAnalogOverride(fn, self, inputHandle, actionHandle)
	}
	if err := ensureSignature(&analogSig, &inputAnalogType, []*ffiType{ffiTypePointer, ffiTypeUint64, ffiTypeUint64}); err != nil {
		panic(err)
	}
	var result InputAnalogActionData
	selfArg := unsafe.Pointer(self)
	inputArg := inputHandle
	actionArg := actionHandle
	args := []unsafe.Pointer{
		unsafe.Pointer(&selfArg),
		unsafe.Pointer(&inputArg),
		unsafe.Pointer(&actionArg),
	}
	callWithArgs(&analogSig.cif, fn, unsafe.Pointer(&result), args)
	runtime.KeepAlive(args)
	return result
}

func CallInputMotionData(fn uintptr, self uintptr, inputHandle uint64) InputMotionData {
	if callInputMotionOverride != nil {
		return callInputMotionOverride(fn, self, inputHandle)
	}
	if err := ensureSignature(&motionSig, &inputMotionType, []*ffiType{ffiTypePointer, ffiTypeUint64}); err != nil {
		panic(err)
	}
	var result InputMotionData
	selfArg := unsafe.Pointer(self)
	inputArg := inputHandle
	args := []unsafe.Pointer{
		unsafe.Pointer(&selfArg),
		unsafe.Pointer(&inputArg),
	}
	callWithArgs(&motionSig.cif, fn, unsafe.Pointer(&result), args)
	runtime.KeepAlive(args)
	return result
}

// SetInputDigitalActionDataOverride swaps the libffi-backed implementation for tests.
// Passing nil restores the default behavior.
func SetInputDigitalActionDataOverride(fn func(fn uintptr, self uintptr, inputHandle uint64, actionHandle uint64) InputDigitalActionData) {
	callInputDigitalOverride = fn
}

// SetInputAnalogActionDataOverride swaps the libffi-backed implementation for tests.
// Passing nil restores the default behavior.
func SetInputAnalogActionDataOverride(fn func(fn uintptr, self uintptr, inputHandle uint64, actionHandle uint64) InputAnalogActionData) {
	callInputAnalogOverride = fn
}

// SetInputMotionDataOverride swaps the libffi-backed implementation for tests.
// Passing nil restores the default behavior.
func SetInputMotionDataOverride(fn func(fn uintptr, self uintptr, inputHandle uint64) InputMotionData) {
	callInputMotionOverride = fn
}

func ensureSignature(sig *callSignature, rtype *ffiType, argTypes []*ffiType) error {
	if err := ensureLibffi(); err != nil {
		return err
	}
	sig.once.Do(func() {
		sig.argTypes = append([]*ffiType(nil), argTypes...)
		var argsPtr **ffiType
		if len(sig.argTypes) > 0 {
			argsPtr = &sig.argTypes[0]
		}
		status, err := prepCif(&sig.cif, rtype, argsPtr, uint32(len(sig.argTypes)))
		if err != nil {
			sig.err = err
			return
		}
		switch status {
		case ffiOK:
			return
		case ffiBadTypedef:
			sig.err = fmt.Errorf("ffi: bad typedef for signature")
		case ffiBadABI:
			sig.err = fmt.Errorf("ffi: unsupported ABI")
		default:
			sig.err = fmt.Errorf("ffi: unsupported status %d", status)
		}
	})
	return sig.err
}

func callWithArgs(cif *ffiCif, fn uintptr, result unsafe.Pointer, args []unsafe.Pointer) {
	var argsPtr unsafe.Pointer
	if len(args) > 0 {
		argsPtr = unsafe.Pointer(&args[0])
	}
	purego.SyscallN(
		ffiCall,
		uintptr(unsafe.Pointer(cif)),
		fn,
		uintptr(result),
		uintptr(argsPtr),
	)
}

func ensureLibffi() error {
	libffiOnce.Do(func() {
		var err error
		libffi, err = loadLibffi()
		if err != nil {
			libffiErr = err
			return
		}
		ffiPrepCif, err = purego.Dlsym(libffi, "ffi_prep_cif")
		if err != nil {
			libffiErr = err
			return
		}
		ffiCall, err = purego.Dlsym(libffi, "ffi_call")
		if err != nil {
			libffiErr = err
			return
		}
		ffiTypeVoid, err = loadFFIType("ffi_type_void")
		if err != nil {
			libffiErr = err
			return
		}
		ffiTypeUint8, err = loadFFIType("ffi_type_uint8")
		if err != nil {
			libffiErr = err
			return
		}
		ffiTypeSint32, err = loadFFIType("ffi_type_sint32")
		if err != nil {
			libffiErr = err
			return
		}
		ffiTypeUint64, err = loadFFIType("ffi_type_uint64")
		if err != nil {
			libffiErr = err
			return
		}
		ffiTypeFloat, err = loadFFIType("ffi_type_float")
		if err != nil {
			libffiErr = err
			return
		}
		ffiTypePointer, err = loadFFIType("ffi_type_pointer")
		if err != nil {
			libffiErr = err
			return
		}
		defaultABI, err = detectABI()
		if err != nil {
			libffiErr = err
			return
		}
		structTypesOnce.Do(initStructTypes)
	})
	return libffiErr
}

func loadLibffi() (uintptr, error) {
	var names []string
	switch runtime.GOOS {
	case "darwin":
		names = []string{"/usr/lib/libffi.dylib", "libffi.dylib"}
	case "windows":
		names = []string{"libffi-8.dll", "libffi-7.dll", "libffi-6.dll"}
	default:
		names = []string{"libffi.so.8", "libffi.so.7", "libffi.so.6", "libffi.so"}
	}
	for _, name := range names {
		lib, err := purego.Dlopen(name, purego.RTLD_LAZY|purego.RTLD_LOCAL)
		if err == nil {
			return lib, nil
		}
	}
	return 0, fmt.Errorf("ffi: unable to load libffi")
}

func loadFFIType(name string) (*ffiType, error) {
	ptr, err := purego.Dlsym(libffi, name)
	if err != nil {
		return nil, err
	}
	return (*ffiType)(unsafe.Pointer(ptr)), nil
}

func detectABI() (uint32, error) {
	var cif ffiCif
	for abi := uint32(0); abi <= 15; abi++ {
		status, err := prepCifWithABI(&cif, abi, ffiTypeVoid, nil, 0)
		if err != nil {
			return 0, err
		}
		if status == ffiOK {
			return abi, nil
		}
	}
	return 0, fmt.Errorf("ffi: unable to detect ABI")
}

func prepCif(cif *ffiCif, rtype *ffiType, args **ffiType, nargs uint32) (ffiStatus, error) {
	return prepCifWithABI(cif, defaultABI, rtype, args, nargs)
}

func prepCifWithABI(cif *ffiCif, abi uint32, rtype *ffiType, args **ffiType, nargs uint32) (ffiStatus, error) {
	if ffiPrepCif == 0 {
		return ffiBadABI, fmt.Errorf("ffi: libffi not loaded")
	}
	ret, _, err := purego.SyscallN(
		ffiPrepCif,
		uintptr(unsafe.Pointer(cif)),
		uintptr(abi),
		uintptr(nargs),
		uintptr(unsafe.Pointer(rtype)),
		uintptr(unsafe.Pointer(args)),
	)
	if err != 0 {
		return ffiBadABI, fmt.Errorf("ffi: prep_cif failed: %v", err)
	}
	return ffiStatus(ret), nil
}

func initStructTypes() {
	inputDigitalElements = []*ffiType{ffiTypeUint8, ffiTypeUint8, nil}
	inputDigitalType = ffiType{ffiType: ffiTypeStruct, elements: &inputDigitalElements[0]}

	inputAnalogElements = []*ffiType{ffiTypeSint32, ffiTypeFloat, ffiTypeFloat, ffiTypeUint8, nil}
	inputAnalogType = ffiType{ffiType: ffiTypeStruct, elements: &inputAnalogElements[0]}

	inputMotionElements = []*ffiType{
		ffiTypeFloat,
		ffiTypeFloat,
		ffiTypeFloat,
		ffiTypeFloat,
		ffiTypeFloat,
		ffiTypeFloat,
		ffiTypeFloat,
		ffiTypeFloat,
		ffiTypeFloat,
		ffiTypeFloat,
		nil,
	}
	inputMotionType = ffiType{ffiType: ffiTypeStruct, elements: &inputMotionElements[0]}
}
