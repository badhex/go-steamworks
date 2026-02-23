// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

//go:build windows

package steamworks

import (
	"fmt"
	"syscall"
)

func lookupSymbolAddr(lib uintptr, name string) (uintptr, error) {
	ptr, err := syscall.GetProcAddress(syscall.Handle(lib), name)
	if err != nil {
		return 0, fmt.Errorf("steamworks: GetProcAddress failed for %s: %w", name, err)
	}
	return uintptr(ptr), nil
}
