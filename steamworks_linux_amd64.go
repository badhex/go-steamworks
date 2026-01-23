// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021 The go-steamworks Authors

//go:build steamworks_embedded && linux && amd64

package steamworks

import (
	_ "embed"
)

//go:embed libsteam_api64.so
var libSteamAPI []byte
