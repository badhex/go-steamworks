// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021 The go-steamworks Authors

//go:build steamworks_embedded && !windows

package steamworks

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ebitengine/purego"
)

func loadLib() (uintptr, error) {
	ext := ".so"
	if runtime.GOOS == "darwin" {
		ext = ".dylib"
	}
	file, err := os.CreateTemp("", "go-steamworks-*"+ext)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	path := filepath.Clean(file.Name())
	if err := os.WriteFile(path, libSteamAPI, 0644); err != nil {
		return 0, err
	}

	lib, err := purego.Dlopen(path, purego.RTLD_LAZY|purego.RTLD_LOCAL)
	if err != nil {
		return 0, fmt.Errorf("steamworks: dlopen failed: %w", err)
	}

	_ = os.Remove(path)

	return lib, nil
}
