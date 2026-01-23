// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The go-steamworks Authors

package steamworks

import (
	"os"
	"strings"
	"testing"
)

func TestSDKHeadersContainExpectedInterfaces(t *testing.T) {
	data, err := os.ReadFile("steam/steam_api_flat.h")
	if err != nil {
		t.Fatalf("read steam_api_flat.h: %v", err)
	}
	text := string(data)
	expected := []string{
		"SteamAPI_SteamApps_v008",
		"SteamAPI_SteamFriends_v018",
		"SteamAPI_SteamHTTP_v003",
		"SteamAPI_SteamUGC_v021",
		"SteamAPI_SteamInventory_v003",
		"SteamAPI_SteamNetworkingUtils_SteamAPI_v004",
		"SteamAPI_SteamNetworkingMessages_SteamAPI_v002",
		"SteamAPI_SteamNetworkingSockets_SteamAPI_v012",
		"SteamAPI_SteamGameServer_v015",
	}
	for _, needle := range expected {
		if !strings.Contains(text, needle) {
			t.Fatalf("steam_api_flat.h missing %q", needle)
		}
	}
}

func TestSDKVersionListedInReadme(t *testing.T) {
	data, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	if !strings.Contains(string(data), SDKVersion) {
		t.Fatalf("README.md does not mention SDK version %s", SDKVersion)
	}
}
