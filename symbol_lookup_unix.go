// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

//go:build !windows

package steamworks

import "github.com/ebitengine/purego"

func lookupSymbolAddr(lib uintptr, name string) (uintptr, error) {
	return purego.Dlsym(lib, name)
}
