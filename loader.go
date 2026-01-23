// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The go-steamworks Authors

package steamworks

import "sync"

var (
	theLib   *lib
	loadOnce sync.Once
	loadErr  error
)

// Load initializes the Steamworks shared library and registers function pointers.
// It is safe to call Load multiple times.
func Load() error {
	return ensureLoaded()
}

func ensureLoaded() error {
	loadOnce.Do(func() {
		l, err := loadLib()
		if err != nil {
			loadErr = err
			return
		}
		registerFunctions(l)
		theLib = &lib{
			lib: l,
		}
	})
	return loadErr
}

func mustLoad() {
	if err := ensureLoaded(); err != nil {
		panic(err)
	}
}
