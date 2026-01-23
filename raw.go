// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The go-steamworks Authors

package steamworks

import "github.com/ebitengine/purego"

// LookupSymbol returns the raw address of a Steamworks SDK symbol for advanced usage.
// Callers are responsible for providing the correct argument and return types.
func LookupSymbol(name string) (uintptr, error) {
	if err := ensureLoaded(); err != nil {
		return 0, err
	}
	return purego.Dlsym(theLib.lib, name)
}

// CallSymbol looks up a Steamworks SDK symbol and invokes it with the provided arguments.
// Callers are responsible for providing the correct argument and return types.
func CallSymbol(name string, args ...uintptr) (uintptr, error) {
	ptr, err := LookupSymbol(name)
	if err != nil {
		return 0, err
	}
	return CallSymbolPtr(ptr, args...), nil
}

// CallSymbolPtr invokes a Steamworks SDK function pointer with the provided arguments.
// Callers are responsible for providing the correct argument and return types.
func CallSymbolPtr(ptr uintptr, args ...uintptr) uintptr {
	r1, _, _ := purego.SyscallN(ptr, args...)
	return r1
}
