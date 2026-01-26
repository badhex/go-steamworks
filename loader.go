// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The go-steamworks Authors

package steamworks

import "sync"

var (
	theLib       *lib
	ensureLoaded = sync.OnceValues(func() (*lib, error) {
		l, err := loadLib()
		if err != nil {
			return nil, err
		}
		registerFunctions(l)
		return &lib{lib: l}, nil
	})
)

// Load initializes the Steamworks shared library and registers function pointers.
// It is safe to call Load multiple times.
func Load() error {
	l, err := ensureLoaded()
	if err != nil {
		return err
	}
	theLib = l
	return nil
}

func mustLoad() {
	l, err := ensureLoaded()
	if err != nil {
		panic(err)
	}
	theLib = l
}
