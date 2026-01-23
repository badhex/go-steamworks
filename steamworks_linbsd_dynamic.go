// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The go-steamworks Authors

//go:build !windows && !steamworks_embedded

package steamworks

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ebitengine/purego"
)

const steamworksLibEnv = "STEAMWORKS_LIB_PATH"

func loadLib() (uintptr, error) {
	libName := "libsteam_api.so"
	if runtime.GOOS == "darwin" {
		libName = "libsteam_api.dylib"
	}

	if customPath := os.Getenv(steamworksLibEnv); customPath != "" {
		path := filepath.Clean(customPath)
		lib, err := purego.Dlopen(path, purego.RTLD_LAZY|purego.RTLD_LOCAL)
		if err != nil {
			return 0, fmt.Errorf("steamworks: dlopen failed for %s: %w", path, err)
		}
		return lib, nil
	}

	lib, err := purego.Dlopen(libName, purego.RTLD_LAZY|purego.RTLD_LOCAL)
	if err != nil {
		return 0, fmt.Errorf("steamworks: dlopen failed for %s: %w", libName, err)
	}
	return lib, nil
}
