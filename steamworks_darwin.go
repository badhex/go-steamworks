// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021 The go-steamworks Authors

//go:build steamworks_embedded && darwin

package steamworks

import (
	_ "embed"
)

//go:embed libsteam_api.dylib
var libSteamAPI []byte
