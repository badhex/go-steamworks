// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The go-steamworks Authors

//go:build !windows

package steamworks

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ebitengine/purego"
	"github.com/jupiterrider/ffi"
)

func sdkLibraryPath() (string, error) {
	envPath := os.Getenv(steamworksLibEnv)
	if envPath == "" {
		return "", fmt.Errorf("%s must be set to the Steamworks SDK library path", steamworksLibEnv)
	}
	if isRemoteLocation(envPath) {
		return downloadLibrary(envPath)
	}
	if _, err := os.Stat(envPath); err != nil {
		return "", fmt.Errorf("%s points to missing file: %w", steamworksLibEnv, err)
	}
	return envPath, nil
}

func isRemoteLocation(path string) bool {
	parsed, err := url.Parse(path)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func downloadLibrary(location string) (string, error) {
	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Get(location)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: %s", resp.Status)
	}

	tmpDir, err := os.MkdirTemp("", "steamworks-sdk-*")
	if err != nil {
		return "", err
	}
	zipPath := filepath.Join(tmpDir, "steamworks_sdk.zip")
	zipFile, err := os.OpenFile(zipPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(zipFile, resp.Body); err != nil {
		zipFile.Close()
		return "", err
	}
	if err := zipFile.Close(); err != nil {
		return "", err
	}

	entryName, err := sdkLibraryEntry()
	if err != nil {
		return "", err
	}
	return extractZipFile(zipPath, entryName, tmpDir)
}

func sdkLibraryEntry() (string, error) {
	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return "sdk/redistributable_bin/linux64/libsteam_api.so", nil
		case "386":
			return "sdk/redistributable_bin/linux32/libsteam_api.so", nil
		}
	case "darwin":
		return "sdk/redistributable_bin/osx/libsteam_api.dylib", nil
	}
	return "", fmt.Errorf("unsupported platform %s/%s", runtime.GOOS, runtime.GOARCH)
}

func extractZipFile(zipPath, entryName, destDir string) (string, error) {
	zipFile, err := os.Open(zipPath)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	info, err := zipFile.Stat()
	if err != nil {
		return "", err
	}

	reader, err := zip.NewReader(zipFile, info.Size())
	if err != nil {
		return "", err
	}

	for _, file := range reader.File {
		if file.Name != entryName {
			continue
		}

		src, err := file.Open()
		if err != nil {
			return "", err
		}
		defer src.Close()

		filename := filepath.Base(entryName)
		if filename == "." || filename == "/" || filename == "" {
			filename = "libsteam_api"
		}
		filename = strings.TrimSuffix(filename, filepath.Ext(filename))
		outPath := filepath.Join(destDir, filename)
		outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return "", err
		}

		if _, err := io.Copy(outFile, src); err != nil {
			outFile.Close()
			return "", err
		}

		if err := outFile.Close(); err != nil {
			return "", err
		}

		return outPath, nil
	}

	return "", fmt.Errorf("sdk library %s not found in archive", entryName)
}

func TestSDKSymbolResolution(t *testing.T) {
	lib := loadSDKLibrary(t)
	expectedMissing := map[string]struct{}{
		flatAPI_ISteamInput_GetGlyphForActionOrigin: {},
	}
	for _, symbol := range allFlatAPISymbols() {
		ptr, err := purego.Dlsym(lib, symbol)
		if err != nil {
			if _, ok := expectedMissing[symbol]; ok {
				continue
			}
			t.Fatalf("Dlsym(%s): %v", symbol, err)
		}
		if _, ok := expectedMissing[symbol]; ok {
			t.Fatalf("expected missing symbol %s, but it was present", symbol)
		}
		if ptr == 0 {
			t.Fatalf("Dlsym(%s) returned 0", symbol)
		}
	}
}

func TestSDKCallSymbol(t *testing.T) {
	lib := loadSDKLibrary(t)
	ptr, err := purego.Dlsym(lib, flatAPI_IsSteamRunning)
	if err != nil {
		t.Fatalf("Dlsym(%s): %v", flatAPI_IsSteamRunning, err)
	}
	result := CallSymbolPtr(ptr)
	if result != 0 && result != 1 {
		t.Fatalf("SteamAPI_IsSteamRunning returned %d, want 0 or 1", result)
	}
}

func TestSDKFunctionSignatures(t *testing.T) {
	actuals := make(map[string]interface{})
	for _, item := range allRegisteredFunctions() {
		actuals[item.name] = item.value
	}

	for _, expectation := range signatureExpectations() {
		actual, ok := actuals[expectation.name]
		if !ok {
			t.Fatalf("missing registered function %s", expectation.name)
		}
		assertSignature(t, expectation.name, actual, expectation.expected)
	}
}

func TestSDKInputStructReturns(t *testing.T) {
	lib := loadSDKLibrary(t)
	registerInputStructReturns(lib)

	ptrs := []struct {
		name string
		ptr  uintptr
	}{
		{name: flatAPI_ISteamInput_GetDigitalActionData, ptr: ptrAPI_ISteamInput_GetDigitalActionData},
		{name: flatAPI_ISteamInput_GetAnalogActionData, ptr: ptrAPI_ISteamInput_GetAnalogActionData},
		{name: flatAPI_ISteamInput_GetMotionData, ptr: ptrAPI_ISteamInput_GetMotionData},
	}

	for _, item := range ptrs {
		if item.ptr == 0 {
			t.Fatalf("%s not registered for struct returns", item.name)
		}
	}

	_ = ffi.CallInputDigitalActionData(ptrAPI_ISteamInput_GetDigitalActionData, 0, 0, 0)
	_ = ffi.CallInputAnalogActionData(ptrAPI_ISteamInput_GetAnalogActionData, 0, 0, 0)
	_ = ffi.CallInputMotionData(ptrAPI_ISteamInput_GetMotionData, 0, 0)
}

type signatureExpectation struct {
	name     string
	expected interface{}
}

type registeredFunction struct {
	name  string
	value interface{}
}

func allRegisteredFunctions() []registeredFunction {
	return []registeredFunction{
		{name: "ptrAPI_RestartAppIfNecessary", value: ptrAPI_RestartAppIfNecessary},
		{name: "ptrAPI_InitFlat", value: ptrAPI_InitFlat},
		{name: "ptrAPI_RunCallbacks", value: ptrAPI_RunCallbacks},
		{name: "ptrAPI_Shutdown", value: ptrAPI_Shutdown},
		{name: "ptrAPI_IsSteamRunning", value: ptrAPI_IsSteamRunning},
		{name: "ptrAPI_GetSteamInstallPath", value: ptrAPI_GetSteamInstallPath},
		{name: "ptrAPI_ReleaseCurrentThreadMemory", value: ptrAPI_ReleaseCurrentThreadMemory},

		{name: "ptrAPI_SteamApps", value: ptrAPI_SteamApps},
		{name: "ptrAPI_ISteamApps_BIsSubscribed", value: ptrAPI_ISteamApps_BIsSubscribed},
		{name: "ptrAPI_ISteamApps_BIsLowViolence", value: ptrAPI_ISteamApps_BIsLowViolence},
		{name: "ptrAPI_ISteamApps_BIsCybercafe", value: ptrAPI_ISteamApps_BIsCybercafe},
		{name: "ptrAPI_ISteamApps_BIsVACBanned", value: ptrAPI_ISteamApps_BIsVACBanned},
		{name: "ptrAPI_ISteamApps_BGetDLCDataByIndex", value: ptrAPI_ISteamApps_BGetDLCDataByIndex},
		{name: "ptrAPI_ISteamApps_BIsDlcInstalled", value: ptrAPI_ISteamApps_BIsDlcInstalled},
		{name: "ptrAPI_ISteamApps_GetAvailableGameLanguages", value: ptrAPI_ISteamApps_GetAvailableGameLanguages},
		{name: "ptrAPI_ISteamApps_BIsSubscribedApp", value: ptrAPI_ISteamApps_BIsSubscribedApp},
		{name: "ptrAPI_ISteamApps_GetEarliestPurchaseUnixTime", value: ptrAPI_ISteamApps_GetEarliestPurchaseUnixTime},
		{name: "ptrAPI_ISteamApps_BIsSubscribedFromFreeWeekend", value: ptrAPI_ISteamApps_BIsSubscribedFromFreeWeekend},
		{name: "ptrAPI_ISteamApps_GetAppInstallDir", value: ptrAPI_ISteamApps_GetAppInstallDir},
		{name: "ptrAPI_ISteamApps_GetCurrentGameLanguage", value: ptrAPI_ISteamApps_GetCurrentGameLanguage},
		{name: "ptrAPI_ISteamApps_GetDLCCount", value: ptrAPI_ISteamApps_GetDLCCount},
		{name: "ptrAPI_ISteamApps_InstallDLC", value: ptrAPI_ISteamApps_InstallDLC},
		{name: "ptrAPI_ISteamApps_UninstallDLC", value: ptrAPI_ISteamApps_UninstallDLC},
		{name: "ptrAPI_ISteamApps_RequestAppProofOfPurchaseKey", value: ptrAPI_ISteamApps_RequestAppProofOfPurchaseKey},
		{name: "ptrAPI_ISteamApps_GetCurrentBetaName", value: ptrAPI_ISteamApps_GetCurrentBetaName},
		{name: "ptrAPI_ISteamApps_MarkContentCorrupt", value: ptrAPI_ISteamApps_MarkContentCorrupt},
		{name: "ptrAPI_ISteamApps_GetInstalledDepots", value: ptrAPI_ISteamApps_GetInstalledDepots},
		{name: "ptrAPI_ISteamApps_BIsAppInstalled", value: ptrAPI_ISteamApps_BIsAppInstalled},
		{name: "ptrAPI_ISteamApps_GetAppOwner", value: ptrAPI_ISteamApps_GetAppOwner},
		{name: "ptrAPI_ISteamApps_GetLaunchQueryParam", value: ptrAPI_ISteamApps_GetLaunchQueryParam},
		{name: "ptrAPI_ISteamApps_GetDlcDownloadProgress", value: ptrAPI_ISteamApps_GetDlcDownloadProgress},
		{name: "ptrAPI_ISteamApps_GetAppBuildId", value: ptrAPI_ISteamApps_GetAppBuildId},
		{name: "ptrAPI_ISteamApps_RequestAllProofOfPurchaseKeys", value: ptrAPI_ISteamApps_RequestAllProofOfPurchaseKeys},
		{name: "ptrAPI_ISteamApps_GetFileDetails", value: ptrAPI_ISteamApps_GetFileDetails},
		{name: "ptrAPI_ISteamApps_GetLaunchCommandLine", value: ptrAPI_ISteamApps_GetLaunchCommandLine},
		{name: "ptrAPI_ISteamApps_BIsSubscribedFromFamilySharing", value: ptrAPI_ISteamApps_BIsSubscribedFromFamilySharing},
		{name: "ptrAPI_ISteamApps_BIsTimedTrial", value: ptrAPI_ISteamApps_BIsTimedTrial},
		{name: "ptrAPI_ISteamApps_SetDlcContext", value: ptrAPI_ISteamApps_SetDlcContext},
		{name: "ptrAPI_ISteamApps_GetNumBetas", value: ptrAPI_ISteamApps_GetNumBetas},
		{name: "ptrAPI_ISteamApps_GetBetaInfo", value: ptrAPI_ISteamApps_GetBetaInfo},
		{name: "ptrAPI_ISteamApps_SetActiveBeta", value: ptrAPI_ISteamApps_SetActiveBeta},

		{name: "ptrAPI_SteamFriends", value: ptrAPI_SteamFriends},
		{name: "ptrAPI_ISteamFriends_GetPersonaName", value: ptrAPI_ISteamFriends_GetPersonaName},
		{name: "ptrAPI_ISteamFriends_GetPersonaState", value: ptrAPI_ISteamFriends_GetPersonaState},
		{name: "ptrAPI_ISteamFriends_GetFriendCount", value: ptrAPI_ISteamFriends_GetFriendCount},
		{name: "ptrAPI_ISteamFriends_GetFriendByIndex", value: ptrAPI_ISteamFriends_GetFriendByIndex},
		{name: "ptrAPI_ISteamFriends_GetFriendRelationship", value: ptrAPI_ISteamFriends_GetFriendRelationship},
		{name: "ptrAPI_ISteamFriends_GetFriendPersonaState", value: ptrAPI_ISteamFriends_GetFriendPersonaState},
		{name: "ptrAPI_ISteamFriends_GetFriendPersonaName", value: ptrAPI_ISteamFriends_GetFriendPersonaName},
		{name: "ptrAPI_ISteamFriends_GetFriendPersonaNameHistory", value: ptrAPI_ISteamFriends_GetFriendPersonaNameHistory},
		{name: "ptrAPI_ISteamFriends_GetFriendSteamLevel", value: ptrAPI_ISteamFriends_GetFriendSteamLevel},
		{name: "ptrAPI_ISteamFriends_GetSmallFriendAvatar", value: ptrAPI_ISteamFriends_GetSmallFriendAvatar},
		{name: "ptrAPI_ISteamFriends_GetMediumFriendAvatar", value: ptrAPI_ISteamFriends_GetMediumFriendAvatar},
		{name: "ptrAPI_ISteamFriends_GetLargeFriendAvatar", value: ptrAPI_ISteamFriends_GetLargeFriendAvatar},
		{name: "ptrAPI_ISteamFriends_SetRichPresence", value: ptrAPI_ISteamFriends_SetRichPresence},
		{name: "ptrAPI_ISteamFriends_GetFriendGamePlayed", value: ptrAPI_ISteamFriends_GetFriendGamePlayed},
		{name: "ptrAPI_ISteamFriends_InviteUserToGame", value: ptrAPI_ISteamFriends_InviteUserToGame},
		{name: "ptrAPI_ISteamFriends_ActivateGameOverlay", value: ptrAPI_ISteamFriends_ActivateGameOverlay},
		{name: "ptrAPI_ISteamFriends_ActivateGameOverlayToUser", value: ptrAPI_ISteamFriends_ActivateGameOverlayToUser},
		{name: "ptrAPI_ISteamFriends_ActivateGameOverlayToWebPage", value: ptrAPI_ISteamFriends_ActivateGameOverlayToWebPage},
		{name: "ptrAPI_ISteamFriends_ActivateGameOverlayToStore", value: ptrAPI_ISteamFriends_ActivateGameOverlayToStore},
		{name: "ptrAPI_ISteamFriends_ActivateGameOverlayInviteDialog", value: ptrAPI_ISteamFriends_ActivateGameOverlayInviteDialog},
		{name: "ptrAPI_ISteamFriends_ActivateGameOverlayInviteDialogConnectString", value: ptrAPI_ISteamFriends_ActivateGameOverlayInviteDialogConnectString},

		{name: "ptrAPI_SteamMatchmaking", value: ptrAPI_SteamMatchmaking},
		{name: "ptrAPI_ISteamMatchmaking_RequestLobbyList", value: ptrAPI_ISteamMatchmaking_RequestLobbyList},
		{name: "ptrAPI_ISteamMatchmaking_GetLobbyByIndex", value: ptrAPI_ISteamMatchmaking_GetLobbyByIndex},
		{name: "ptrAPI_ISteamMatchmaking_CreateLobby", value: ptrAPI_ISteamMatchmaking_CreateLobby},
		{name: "ptrAPI_ISteamMatchmaking_JoinLobby", value: ptrAPI_ISteamMatchmaking_JoinLobby},
		{name: "ptrAPI_ISteamMatchmaking_LeaveLobby", value: ptrAPI_ISteamMatchmaking_LeaveLobby},
		{name: "ptrAPI_ISteamMatchmaking_InviteUserToLobby", value: ptrAPI_ISteamMatchmaking_InviteUserToLobby},
		{name: "ptrAPI_ISteamMatchmaking_GetNumLobbyMembers", value: ptrAPI_ISteamMatchmaking_GetNumLobbyMembers},
		{name: "ptrAPI_ISteamMatchmaking_GetLobbyMemberByIndex", value: ptrAPI_ISteamMatchmaking_GetLobbyMemberByIndex},
		{name: "ptrAPI_ISteamMatchmaking_GetLobbyData", value: ptrAPI_ISteamMatchmaking_GetLobbyData},
		{name: "ptrAPI_ISteamMatchmaking_SetLobbyData", value: ptrAPI_ISteamMatchmaking_SetLobbyData},
		{name: "ptrAPI_ISteamMatchmaking_GetLobbyOwner", value: ptrAPI_ISteamMatchmaking_GetLobbyOwner},
		{name: "ptrAPI_ISteamMatchmaking_SetLobbyOwner", value: ptrAPI_ISteamMatchmaking_SetLobbyOwner},
		{name: "ptrAPI_ISteamMatchmaking_SetLobbyGameServer", value: ptrAPI_ISteamMatchmaking_SetLobbyGameServer},
		{name: "ptrAPI_ISteamMatchmaking_GetLobbyGameServer", value: ptrAPI_ISteamMatchmaking_GetLobbyGameServer},
		{name: "ptrAPI_ISteamMatchmaking_SetLobbyJoinable", value: ptrAPI_ISteamMatchmaking_SetLobbyJoinable},
		{name: "ptrAPI_ISteamMatchmaking_SetLobbyMemberLimit", value: ptrAPI_ISteamMatchmaking_SetLobbyMemberLimit},
		{name: "ptrAPI_ISteamMatchmaking_SetLobbyType", value: ptrAPI_ISteamMatchmaking_SetLobbyType},

		{name: "ptrAPI_SteamHTTP", value: ptrAPI_SteamHTTP},
		{name: "ptrAPI_ISteamHTTP_CreateHTTPRequest", value: ptrAPI_ISteamHTTP_CreateHTTPRequest},
		{name: "ptrAPI_ISteamHTTP_SetHTTPRequestHeaderValue", value: ptrAPI_ISteamHTTP_SetHTTPRequestHeaderValue},
		{name: "ptrAPI_ISteamHTTP_SendHTTPRequest", value: ptrAPI_ISteamHTTP_SendHTTPRequest},
		{name: "ptrAPI_ISteamHTTP_GetHTTPResponseBodySize", value: ptrAPI_ISteamHTTP_GetHTTPResponseBodySize},
		{name: "ptrAPI_ISteamHTTP_GetHTTPResponseBodyData", value: ptrAPI_ISteamHTTP_GetHTTPResponseBodyData},
		{name: "ptrAPI_ISteamHTTP_ReleaseHTTPRequest", value: ptrAPI_ISteamHTTP_ReleaseHTTPRequest},

		{name: "ptrAPI_SteamUGC", value: ptrAPI_SteamUGC},
		{name: "ptrAPI_ISteamUGC_GetNumSubscribedItems", value: ptrAPI_ISteamUGC_GetNumSubscribedItems},
		{name: "ptrAPI_ISteamUGC_GetSubscribedItems", value: ptrAPI_ISteamUGC_GetSubscribedItems},

		{name: "ptrAPI_SteamInventory", value: ptrAPI_SteamInventory},
		{name: "ptrAPI_ISteamInventory_GetResultStatus", value: ptrAPI_ISteamInventory_GetResultStatus},
		{name: "ptrAPI_ISteamInventory_GetResultItems", value: ptrAPI_ISteamInventory_GetResultItems},
		{name: "ptrAPI_ISteamInventory_DestroyResult", value: ptrAPI_ISteamInventory_DestroyResult},

		{name: "ptrAPI_SteamInput", value: ptrAPI_SteamInput},
		{name: "ptrAPI_ISteamInput_GetConnectedControllers", value: ptrAPI_ISteamInput_GetConnectedControllers},
		{name: "ptrAPI_ISteamInput_GetInputTypeForHandle", value: ptrAPI_ISteamInput_GetInputTypeForHandle},
		{name: "ptrAPI_ISteamInput_Init", value: ptrAPI_ISteamInput_Init},
		{name: "ptrAPI_ISteamInput_Shutdown", value: ptrAPI_ISteamInput_Shutdown},
		{name: "ptrAPI_ISteamInput_RunFrame", value: ptrAPI_ISteamInput_RunFrame},
		{name: "ptrAPI_ISteamInput_EnableDeviceCallbacks", value: ptrAPI_ISteamInput_EnableDeviceCallbacks},
		{name: "ptrAPI_ISteamInput_GetActionSetHandle", value: ptrAPI_ISteamInput_GetActionSetHandle},
		{name: "ptrAPI_ISteamInput_ActivateActionSet", value: ptrAPI_ISteamInput_ActivateActionSet},
		{name: "ptrAPI_ISteamInput_GetCurrentActionSet", value: ptrAPI_ISteamInput_GetCurrentActionSet},
		{name: "ptrAPI_ISteamInput_ActivateActionSetLayer", value: ptrAPI_ISteamInput_ActivateActionSetLayer},
		{name: "ptrAPI_ISteamInput_DeactivateActionSetLayer", value: ptrAPI_ISteamInput_DeactivateActionSetLayer},
		{name: "ptrAPI_ISteamInput_DeactivateAllActionSetLayers", value: ptrAPI_ISteamInput_DeactivateAllActionSetLayers},
		{name: "ptrAPI_ISteamInput_GetActiveActionSetLayers", value: ptrAPI_ISteamInput_GetActiveActionSetLayers},
		{name: "ptrAPI_ISteamInput_GetDigitalActionHandle", value: ptrAPI_ISteamInput_GetDigitalActionHandle},
		{name: "ptrAPI_ISteamInput_GetDigitalActionData", value: ptrAPI_ISteamInput_GetDigitalActionData},
		{name: "ptrAPI_ISteamInput_GetDigitalActionOrigins", value: ptrAPI_ISteamInput_GetDigitalActionOrigins},
		{name: "ptrAPI_ISteamInput_GetAnalogActionHandle", value: ptrAPI_ISteamInput_GetAnalogActionHandle},
		{name: "ptrAPI_ISteamInput_GetAnalogActionData", value: ptrAPI_ISteamInput_GetAnalogActionData},
		{name: "ptrAPI_ISteamInput_GetAnalogActionOrigins", value: ptrAPI_ISteamInput_GetAnalogActionOrigins},
		{name: "ptrAPI_ISteamInput_StopAnalogActionMomentum", value: ptrAPI_ISteamInput_StopAnalogActionMomentum},
		{name: "ptrAPI_ISteamInput_GetMotionData", value: ptrAPI_ISteamInput_GetMotionData},
		{name: "ptrAPI_ISteamInput_TriggerVibration", value: ptrAPI_ISteamInput_TriggerVibration},
		{name: "ptrAPI_ISteamInput_TriggerVibrationExtended", value: ptrAPI_ISteamInput_TriggerVibrationExtended},
		{name: "ptrAPI_ISteamInput_TriggerSimpleHapticEvent", value: ptrAPI_ISteamInput_TriggerSimpleHapticEvent},
		{name: "ptrAPI_ISteamInput_SetLEDColor", value: ptrAPI_ISteamInput_SetLEDColor},
		{name: "ptrAPI_ISteamInput_ShowBindingPanel", value: ptrAPI_ISteamInput_ShowBindingPanel},
		{name: "ptrAPI_ISteamInput_GetControllerForGamepadIndex", value: ptrAPI_ISteamInput_GetControllerForGamepadIndex},
		{name: "ptrAPI_ISteamInput_GetGamepadIndexForController", value: ptrAPI_ISteamInput_GetGamepadIndexForController},
		{name: "ptrAPI_ISteamInput_GetStringForActionOrigin", value: ptrAPI_ISteamInput_GetStringForActionOrigin},
		{name: "ptrAPI_ISteamInput_GetGlyphForActionOrigin", value: ptrAPI_ISteamInput_GetGlyphForActionOrigin},
		{name: "ptrAPI_ISteamInput_GetRemotePlaySessionID", value: ptrAPI_ISteamInput_GetRemotePlaySessionID},

		{name: "ptrAPI_SteamRemoteStorage", value: ptrAPI_SteamRemoteStorage},
		{name: "ptrAPI_ISteamRemoteStorage_FileWrite", value: ptrAPI_ISteamRemoteStorage_FileWrite},
		{name: "ptrAPI_ISteamRemoteStorage_FileRead", value: ptrAPI_ISteamRemoteStorage_FileRead},
		{name: "ptrAPI_ISteamRemoteStorage_FileDelete", value: ptrAPI_ISteamRemoteStorage_FileDelete},
		{name: "ptrAPI_ISteamRemoteStorage_GetFileSize", value: ptrAPI_ISteamRemoteStorage_GetFileSize},

		{name: "ptrAPI_SteamUser", value: ptrAPI_SteamUser},
		{name: "ptrAPI_ISteamUser_GetSteamID", value: ptrAPI_ISteamUser_GetSteamID},

		{name: "ptrAPI_SteamUserStats", value: ptrAPI_SteamUserStats},
		{name: "ptrAPI_ISteamUserStats_GetAchievement", value: ptrAPI_ISteamUserStats_GetAchievement},
		{name: "ptrAPI_ISteamUserStats_SetAchievement", value: ptrAPI_ISteamUserStats_SetAchievement},
		{name: "ptrAPI_ISteamUserStats_ClearAchievement", value: ptrAPI_ISteamUserStats_ClearAchievement},
		{name: "ptrAPI_ISteamUserStats_StoreStats", value: ptrAPI_ISteamUserStats_StoreStats},

		{name: "ptrAPI_SteamUtils", value: ptrAPI_SteamUtils},
		{name: "ptrAPI_ISteamUtils_GetSecondsSinceAppActive", value: ptrAPI_ISteamUtils_GetSecondsSinceAppActive},
		{name: "ptrAPI_ISteamUtils_GetSecondsSinceComputerActive", value: ptrAPI_ISteamUtils_GetSecondsSinceComputerActive},
		{name: "ptrAPI_ISteamUtils_GetConnectedUniverse", value: ptrAPI_ISteamUtils_GetConnectedUniverse},
		{name: "ptrAPI_ISteamUtils_GetServerRealTime", value: ptrAPI_ISteamUtils_GetServerRealTime},
		{name: "ptrAPI_ISteamUtils_GetIPCountry", value: ptrAPI_ISteamUtils_GetIPCountry},
		{name: "ptrAPI_ISteamUtils_GetImageSize", value: ptrAPI_ISteamUtils_GetImageSize},
		{name: "ptrAPI_ISteamUtils_GetImageRGBA", value: ptrAPI_ISteamUtils_GetImageRGBA},
		{name: "ptrAPI_ISteamUtils_GetCurrentBatteryPower", value: ptrAPI_ISteamUtils_GetCurrentBatteryPower},
		{name: "ptrAPI_ISteamUtils_GetAppID", value: ptrAPI_ISteamUtils_GetAppID},
		{name: "ptrAPI_ISteamUtils_SetOverlayNotificationPosition", value: ptrAPI_ISteamUtils_SetOverlayNotificationPosition},
		{name: "ptrAPI_ISteamUtils_IsAPICallCompleted", value: ptrAPI_ISteamUtils_IsAPICallCompleted},
		{name: "ptrAPI_ISteamUtils_GetAPICallFailureReason", value: ptrAPI_ISteamUtils_GetAPICallFailureReason},
		{name: "ptrAPI_ISteamUtils_GetAPICallResult", value: ptrAPI_ISteamUtils_GetAPICallResult},
		{name: "ptrAPI_ISteamUtils_GetIPCCallCount", value: ptrAPI_ISteamUtils_GetIPCCallCount},
		{name: "ptrAPI_ISteamUtils_IsOverlayEnabled", value: ptrAPI_ISteamUtils_IsOverlayEnabled},
		{name: "ptrAPI_ISteamUtils_BOverlayNeedsPresent", value: ptrAPI_ISteamUtils_BOverlayNeedsPresent},
		{name: "ptrAPI_ISteamUtils_IsSteamRunningOnSteamDeck", value: ptrAPI_ISteamUtils_IsSteamRunningOnSteamDeck},
		{name: "ptrAPI_ISteamUtils_ShowFloatingGamepadTextInput", value: ptrAPI_ISteamUtils_ShowFloatingGamepadTextInput},
		{name: "ptrAPI_ISteamUtils_SetOverlayNotificationInset", value: ptrAPI_ISteamUtils_SetOverlayNotificationInset},

		{name: "ptrAPI_SteamNetworkingUtils", value: ptrAPI_SteamNetworkingUtils},
		{name: "ptrAPI_ISteamNetworkingUtils_AllocateMessage", value: ptrAPI_ISteamNetworkingUtils_AllocateMessage},
		{name: "ptrAPI_ISteamNetworkingUtils_InitRelayNetworkAccess", value: ptrAPI_ISteamNetworkingUtils_InitRelayNetworkAccess},
		{name: "ptrAPI_ISteamNetworkingUtils_GetLocalTimestamp", value: ptrAPI_ISteamNetworkingUtils_GetLocalTimestamp},

		{name: "ptrAPI_SteamGameServer", value: ptrAPI_SteamGameServer},
		{name: "ptrAPI_ISteamGameServer_SetProduct", value: ptrAPI_ISteamGameServer_SetProduct},
		{name: "ptrAPI_ISteamGameServer_SetGameDescription", value: ptrAPI_ISteamGameServer_SetGameDescription},
		{name: "ptrAPI_ISteamGameServer_LogOnAnonymous", value: ptrAPI_ISteamGameServer_LogOnAnonymous},
		{name: "ptrAPI_ISteamGameServer_LogOff", value: ptrAPI_ISteamGameServer_LogOff},
		{name: "ptrAPI_ISteamGameServer_BLoggedOn", value: ptrAPI_ISteamGameServer_BLoggedOn},
		{name: "ptrAPI_ISteamGameServer_GetSteamID", value: ptrAPI_ISteamGameServer_GetSteamID},

		{name: "ptrAPI_SteamNetworkingMessages", value: ptrAPI_SteamNetworkingMessages},
		{name: "ptrAPI_ISteamNetworkingMessages_SendMessageToUser", value: ptrAPI_ISteamNetworkingMessages_SendMessageToUser},
		{name: "ptrAPI_ISteamNetworkingMessages_ReceiveMessagesOnChannel", value: ptrAPI_ISteamNetworkingMessages_ReceiveMessagesOnChannel},
		{name: "ptrAPI_ISteamNetworkingMessages_AcceptSessionWithUser", value: ptrAPI_ISteamNetworkingMessages_AcceptSessionWithUser},
		{name: "ptrAPI_ISteamNetworkingMessages_CloseSessionWithUser", value: ptrAPI_ISteamNetworkingMessages_CloseSessionWithUser},
		{name: "ptrAPI_ISteamNetworkingMessages_CloseChannelWithUser", value: ptrAPI_ISteamNetworkingMessages_CloseChannelWithUser},

		{name: "ptrAPI_SteamNetworkingSockets", value: ptrAPI_SteamNetworkingSockets},
		{name: "ptrAPI_ISteamNetworkingSockets_CreateListenSocketIP", value: ptrAPI_ISteamNetworkingSockets_CreateListenSocketIP},
		{name: "ptrAPI_ISteamNetworkingSockets_CreateListenSocketP2P", value: ptrAPI_ISteamNetworkingSockets_CreateListenSocketP2P},
		{name: "ptrAPI_ISteamNetworkingSockets_ConnectByIPAddress", value: ptrAPI_ISteamNetworkingSockets_ConnectByIPAddress},
		{name: "ptrAPI_ISteamNetworkingSockets_ConnectP2P", value: ptrAPI_ISteamNetworkingSockets_ConnectP2P},
		{name: "ptrAPI_ISteamNetworkingSockets_AcceptConnection", value: ptrAPI_ISteamNetworkingSockets_AcceptConnection},
		{name: "ptrAPI_ISteamNetworkingSockets_CloseConnection", value: ptrAPI_ISteamNetworkingSockets_CloseConnection},
		{name: "ptrAPI_ISteamNetworkingSockets_CloseListenSocket", value: ptrAPI_ISteamNetworkingSockets_CloseListenSocket},
		{name: "ptrAPI_ISteamNetworkingSockets_SendMessageToConnection", value: ptrAPI_ISteamNetworkingSockets_SendMessageToConnection},
		{name: "ptrAPI_ISteamNetworkingSockets_ReceiveMessagesOnConnection", value: ptrAPI_ISteamNetworkingSockets_ReceiveMessagesOnConnection},
		{name: "ptrAPI_ISteamNetworkingSockets_CreatePollGroup", value: ptrAPI_ISteamNetworkingSockets_CreatePollGroup},
		{name: "ptrAPI_ISteamNetworkingSockets_DestroyPollGroup", value: ptrAPI_ISteamNetworkingSockets_DestroyPollGroup},
		{name: "ptrAPI_ISteamNetworkingSockets_SetConnectionPollGroup", value: ptrAPI_ISteamNetworkingSockets_SetConnectionPollGroup},
		{name: "ptrAPI_ISteamNetworkingSockets_ReceiveMessagesOnPollGroup", value: ptrAPI_ISteamNetworkingSockets_ReceiveMessagesOnPollGroup},
	}
}

func loadSDKLibrary(t *testing.T) uintptr {
	t.Helper()

	path, err := sdkLibraryPath()
	if err != nil {
		t.Fatalf("sdkLibraryPath: %v", err)
	}
	_ = os.Setenv(steamworksLibEnv, path)

	lib, err := loadLib()
	if err != nil {
		t.Fatalf("loadLib: %v", err)
	}

	return lib
}

func assertSignature(t *testing.T, name string, actual interface{}, expected interface{}) {
	t.Helper()

	if _, ok := expected.(uintptr); ok {
		if _, ok := actual.(uintptr); !ok {
			t.Fatalf("%s has type %T, want uintptr", name, actual)
		}
		return
	}

	actualType := reflect.TypeOf(actual)
	expectedType := reflect.TypeOf(expected)
	if actualType != expectedType {
		t.Fatalf("%s has type %v, want %v", name, actualType, expectedType)
	}
}

func signatureExpectations() []signatureExpectation {
	return []signatureExpectation{
		{name: "ptrAPI_RestartAppIfNecessary", expected: (func(uint32) bool)(nil)},
		{name: "ptrAPI_InitFlat", expected: (func(uintptr) ESteamAPIInitResult)(nil)},
		{name: "ptrAPI_RunCallbacks", expected: (func())(nil)},
		{name: "ptrAPI_Shutdown", expected: (func())(nil)},
		{name: "ptrAPI_IsSteamRunning", expected: (func() bool)(nil)},
		{name: "ptrAPI_GetSteamInstallPath", expected: (func() string)(nil)},
		{name: "ptrAPI_ReleaseCurrentThreadMemory", expected: (func())(nil)},

		{name: "ptrAPI_SteamApps", expected: (func() uintptr)(nil)},
		{name: "ptrAPI_ISteamApps_BIsSubscribed", expected: (func(uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamApps_BIsLowViolence", expected: (func(uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamApps_BIsCybercafe", expected: (func(uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamApps_BIsVACBanned", expected: (func(uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamApps_BGetDLCDataByIndex", expected: (func(uintptr, int32, uintptr, uintptr, uintptr, int32) bool)(nil)},
		{name: "ptrAPI_ISteamApps_BIsDlcInstalled", expected: (func(uintptr, AppId_t) bool)(nil)},
		{name: "ptrAPI_ISteamApps_GetAvailableGameLanguages", expected: (func(uintptr) string)(nil)},
		{name: "ptrAPI_ISteamApps_BIsSubscribedApp", expected: (func(uintptr, AppId_t) bool)(nil)},
		{name: "ptrAPI_ISteamApps_GetEarliestPurchaseUnixTime", expected: (func(uintptr, AppId_t) uint32)(nil)},
		{name: "ptrAPI_ISteamApps_BIsSubscribedFromFreeWeekend", expected: (func(uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamApps_GetAppInstallDir", expected: (func(uintptr, AppId_t, uintptr, int32) int32)(nil)},
		{name: "ptrAPI_ISteamApps_GetCurrentGameLanguage", expected: (func(uintptr) string)(nil)},
		{name: "ptrAPI_ISteamApps_GetDLCCount", expected: (func(uintptr) int32)(nil)},
		{name: "ptrAPI_ISteamApps_InstallDLC", expected: (func(uintptr, AppId_t))(nil)},
		{name: "ptrAPI_ISteamApps_UninstallDLC", expected: (func(uintptr, AppId_t))(nil)},
		{name: "ptrAPI_ISteamApps_RequestAppProofOfPurchaseKey", expected: (func(uintptr, AppId_t))(nil)},
		{name: "ptrAPI_ISteamApps_GetCurrentBetaName", expected: (func(uintptr, uintptr, int32) bool)(nil)},
		{name: "ptrAPI_ISteamApps_MarkContentCorrupt", expected: (func(uintptr, bool) bool)(nil)},
		{name: "ptrAPI_ISteamApps_GetInstalledDepots", expected: (func(uintptr, AppId_t, uintptr, uint32) uint32)(nil)},
		{name: "ptrAPI_ISteamApps_BIsAppInstalled", expected: (func(uintptr, AppId_t) bool)(nil)},
		{name: "ptrAPI_ISteamApps_GetAppOwner", expected: (func(uintptr) CSteamID)(nil)},
		{name: "ptrAPI_ISteamApps_GetLaunchQueryParam", expected: (func(uintptr, string) string)(nil)},
		{name: "ptrAPI_ISteamApps_GetDlcDownloadProgress", expected: (func(uintptr, AppId_t, uintptr, uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamApps_GetAppBuildId", expected: (func(uintptr) int32)(nil)},
		{name: "ptrAPI_ISteamApps_RequestAllProofOfPurchaseKeys", expected: (func(uintptr))(nil)},
		{name: "ptrAPI_ISteamApps_GetFileDetails", expected: (func(uintptr, string) SteamAPICall_t)(nil)},
		{name: "ptrAPI_ISteamApps_GetLaunchCommandLine", expected: (func(uintptr, uintptr, int32) int32)(nil)},
		{name: "ptrAPI_ISteamApps_BIsSubscribedFromFamilySharing", expected: (func(uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamApps_BIsTimedTrial", expected: (func(uintptr, uintptr, uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamApps_SetDlcContext", expected: (func(uintptr, AppId_t) bool)(nil)},
		{name: "ptrAPI_ISteamApps_GetNumBetas", expected: (func(uintptr, uintptr, uintptr) int32)(nil)},
		{name: "ptrAPI_ISteamApps_GetBetaInfo", expected: (func(uintptr, int32, uintptr, uintptr, uintptr, int32, uintptr, int32) bool)(nil)},
		{name: "ptrAPI_ISteamApps_SetActiveBeta", expected: (func(uintptr, string) bool)(nil)},

		{name: "ptrAPI_SteamFriends", expected: (func() uintptr)(nil)},
		{name: "ptrAPI_ISteamFriends_GetPersonaName", expected: (func(uintptr) string)(nil)},
		{name: "ptrAPI_ISteamFriends_GetPersonaState", expected: (func(uintptr) int32)(nil)},
		{name: "ptrAPI_ISteamFriends_GetFriendCount", expected: (func(uintptr, int32) int32)(nil)},
		{name: "ptrAPI_ISteamFriends_GetFriendByIndex", expected: (func(uintptr, int32, int32) CSteamID)(nil)},
		{name: "ptrAPI_ISteamFriends_GetFriendRelationship", expected: (func(uintptr, CSteamID) int32)(nil)},
		{name: "ptrAPI_ISteamFriends_GetFriendPersonaState", expected: (func(uintptr, CSteamID) int32)(nil)},
		{name: "ptrAPI_ISteamFriends_GetFriendPersonaName", expected: (func(uintptr, CSteamID) string)(nil)},
		{name: "ptrAPI_ISteamFriends_GetFriendPersonaNameHistory", expected: (func(uintptr, CSteamID, int32) string)(nil)},
		{name: "ptrAPI_ISteamFriends_GetFriendSteamLevel", expected: (func(uintptr, CSteamID) int32)(nil)},
		{name: "ptrAPI_ISteamFriends_GetSmallFriendAvatar", expected: (func(uintptr, CSteamID) int32)(nil)},
		{name: "ptrAPI_ISteamFriends_GetMediumFriendAvatar", expected: (func(uintptr, CSteamID) int32)(nil)},
		{name: "ptrAPI_ISteamFriends_GetLargeFriendAvatar", expected: (func(uintptr, CSteamID) int32)(nil)},
		{name: "ptrAPI_ISteamFriends_SetRichPresence", expected: (func(uintptr, string, string) bool)(nil)},
		{name: "ptrAPI_ISteamFriends_GetFriendGamePlayed", expected: (func(uintptr, CSteamID, uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamFriends_InviteUserToGame", expected: (func(uintptr, CSteamID, string) bool)(nil)},
		{name: "ptrAPI_ISteamFriends_ActivateGameOverlay", expected: (func(uintptr, string))(nil)},
		{name: "ptrAPI_ISteamFriends_ActivateGameOverlayToUser", expected: (func(uintptr, string, CSteamID))(nil)},
		{name: "ptrAPI_ISteamFriends_ActivateGameOverlayToWebPage", expected: (func(uintptr, string, EActivateGameOverlayToWebPageMode))(nil)},
		{name: "ptrAPI_ISteamFriends_ActivateGameOverlayToStore", expected: (func(uintptr, AppId_t, EOverlayToStoreFlag))(nil)},
		{name: "ptrAPI_ISteamFriends_ActivateGameOverlayInviteDialog", expected: (func(uintptr, CSteamID))(nil)},
		{name: "ptrAPI_ISteamFriends_ActivateGameOverlayInviteDialogConnectString", expected: (func(uintptr, string))(nil)},

		{name: "ptrAPI_SteamMatchmaking", expected: (func() uintptr)(nil)},
		{name: "ptrAPI_ISteamMatchmaking_RequestLobbyList", expected: (func(uintptr) SteamAPICall_t)(nil)},
		{name: "ptrAPI_ISteamMatchmaking_GetLobbyByIndex", expected: (func(uintptr, int32) CSteamID)(nil)},
		{name: "ptrAPI_ISteamMatchmaking_CreateLobby", expected: (func(uintptr, ELobbyType, int32) SteamAPICall_t)(nil)},
		{name: "ptrAPI_ISteamMatchmaking_JoinLobby", expected: (func(uintptr, CSteamID) SteamAPICall_t)(nil)},
		{name: "ptrAPI_ISteamMatchmaking_LeaveLobby", expected: (func(uintptr, CSteamID))(nil)},
		{name: "ptrAPI_ISteamMatchmaking_InviteUserToLobby", expected: (func(uintptr, CSteamID, CSteamID) bool)(nil)},
		{name: "ptrAPI_ISteamMatchmaking_GetNumLobbyMembers", expected: (func(uintptr, CSteamID) int32)(nil)},
		{name: "ptrAPI_ISteamMatchmaking_GetLobbyMemberByIndex", expected: (func(uintptr, CSteamID, int32) CSteamID)(nil)},
		{name: "ptrAPI_ISteamMatchmaking_GetLobbyData", expected: (func(uintptr, CSteamID, string) string)(nil)},
		{name: "ptrAPI_ISteamMatchmaking_SetLobbyData", expected: (func(uintptr, CSteamID, string, string) bool)(nil)},
		{name: "ptrAPI_ISteamMatchmaking_GetLobbyOwner", expected: (func(uintptr, CSteamID) CSteamID)(nil)},
		{name: "ptrAPI_ISteamMatchmaking_SetLobbyOwner", expected: (func(uintptr, CSteamID, CSteamID) bool)(nil)},
		{name: "ptrAPI_ISteamMatchmaking_SetLobbyGameServer", expected: (func(uintptr, CSteamID, uint32, uint16, CSteamID))(nil)},
		{name: "ptrAPI_ISteamMatchmaking_GetLobbyGameServer", expected: (func(uintptr, CSteamID, uintptr, uintptr, uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamMatchmaking_SetLobbyJoinable", expected: (func(uintptr, CSteamID, bool) bool)(nil)},
		{name: "ptrAPI_ISteamMatchmaking_SetLobbyMemberLimit", expected: (func(uintptr, CSteamID, int32) bool)(nil)},
		{name: "ptrAPI_ISteamMatchmaking_SetLobbyType", expected: (func(uintptr, CSteamID, ELobbyType) bool)(nil)},

		{name: "ptrAPI_SteamHTTP", expected: (func() uintptr)(nil)},
		{name: "ptrAPI_ISteamHTTP_CreateHTTPRequest", expected: (func(uintptr, int32, string) HTTPRequestHandle)(nil)},
		{name: "ptrAPI_ISteamHTTP_SetHTTPRequestHeaderValue", expected: (func(uintptr, HTTPRequestHandle, string, string) bool)(nil)},
		{name: "ptrAPI_ISteamHTTP_SendHTTPRequest", expected: (func(uintptr, HTTPRequestHandle, uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamHTTP_GetHTTPResponseBodySize", expected: (func(uintptr, HTTPRequestHandle, uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamHTTP_GetHTTPResponseBodyData", expected: (func(uintptr, HTTPRequestHandle, uintptr, uint32) bool)(nil)},
		{name: "ptrAPI_ISteamHTTP_ReleaseHTTPRequest", expected: (func(uintptr, HTTPRequestHandle) bool)(nil)},

		{name: "ptrAPI_SteamUGC", expected: (func() uintptr)(nil)},
		{name: "ptrAPI_ISteamUGC_GetNumSubscribedItems", expected: (func(uintptr, bool) uint32)(nil)},
		{name: "ptrAPI_ISteamUGC_GetSubscribedItems", expected: (func(uintptr, uintptr, uint32, bool) uint32)(nil)},

		{name: "ptrAPI_SteamInventory", expected: (func() uintptr)(nil)},
		{name: "ptrAPI_ISteamInventory_GetResultStatus", expected: (func(uintptr, SteamInventoryResult_t) int32)(nil)},
		{name: "ptrAPI_ISteamInventory_GetResultItems", expected: (func(uintptr, SteamInventoryResult_t, uintptr, uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamInventory_DestroyResult", expected: (func(uintptr, SteamInventoryResult_t))(nil)},

		{name: "ptrAPI_SteamInput", expected: (func() uintptr)(nil)},
		{name: "ptrAPI_ISteamInput_GetConnectedControllers", expected: (func(uintptr, uintptr) int32)(nil)},
		{name: "ptrAPI_ISteamInput_GetInputTypeForHandle", expected: (func(uintptr, InputHandle_t) int32)(nil)},
		{name: "ptrAPI_ISteamInput_Init", expected: (func(uintptr, bool) bool)(nil)},
		{name: "ptrAPI_ISteamInput_Shutdown", expected: (func(uintptr))(nil)},
		{name: "ptrAPI_ISteamInput_RunFrame", expected: (func(uintptr, bool))(nil)},
		{name: "ptrAPI_ISteamInput_EnableDeviceCallbacks", expected: (func(uintptr))(nil)},
		{name: "ptrAPI_ISteamInput_GetActionSetHandle", expected: (func(uintptr, string) InputActionSetHandle_t)(nil)},
		{name: "ptrAPI_ISteamInput_ActivateActionSet", expected: (func(uintptr, InputHandle_t, InputActionSetHandle_t))(nil)},
		{name: "ptrAPI_ISteamInput_GetCurrentActionSet", expected: (func(uintptr, InputHandle_t) InputActionSetHandle_t)(nil)},
		{name: "ptrAPI_ISteamInput_ActivateActionSetLayer", expected: (func(uintptr, InputHandle_t, InputActionSetHandle_t))(nil)},
		{name: "ptrAPI_ISteamInput_DeactivateActionSetLayer", expected: (func(uintptr, InputHandle_t, InputActionSetHandle_t))(nil)},
		{name: "ptrAPI_ISteamInput_DeactivateAllActionSetLayers", expected: (func(uintptr, InputHandle_t))(nil)},
		{name: "ptrAPI_ISteamInput_GetActiveActionSetLayers", expected: (func(uintptr, InputHandle_t, uintptr) int32)(nil)},
		{name: "ptrAPI_ISteamInput_GetDigitalActionHandle", expected: (func(uintptr, string) InputDigitalActionHandle_t)(nil)},
		{name: "ptrAPI_ISteamInput_GetDigitalActionData", expected: uintptr(0)},
		{name: "ptrAPI_ISteamInput_GetDigitalActionOrigins", expected: (func(uintptr, InputHandle_t, InputActionSetHandle_t, InputDigitalActionHandle_t, uintptr) int32)(nil)},
		{name: "ptrAPI_ISteamInput_GetAnalogActionHandle", expected: (func(uintptr, string) InputAnalogActionHandle_t)(nil)},
		{name: "ptrAPI_ISteamInput_GetAnalogActionData", expected: uintptr(0)},
		{name: "ptrAPI_ISteamInput_GetAnalogActionOrigins", expected: (func(uintptr, InputHandle_t, InputActionSetHandle_t, InputAnalogActionHandle_t, uintptr) int32)(nil)},
		{name: "ptrAPI_ISteamInput_StopAnalogActionMomentum", expected: (func(uintptr, InputHandle_t, InputAnalogActionHandle_t))(nil)},
		{name: "ptrAPI_ISteamInput_GetMotionData", expected: uintptr(0)},
		{name: "ptrAPI_ISteamInput_TriggerVibration", expected: (func(uintptr, InputHandle_t, uint16, uint16))(nil)},
		{name: "ptrAPI_ISteamInput_TriggerVibrationExtended", expected: (func(uintptr, InputHandle_t, uint16, uint16, uint16, uint16))(nil)},
		{name: "ptrAPI_ISteamInput_TriggerSimpleHapticEvent", expected: (func(uintptr, InputHandle_t, ESteamControllerPad, uint16, uint16, uint16))(nil)},
		{name: "ptrAPI_ISteamInput_SetLEDColor", expected: (func(uintptr, InputHandle_t, uint8, uint8, uint8, ESteamInputLEDFlag))(nil)},
		{name: "ptrAPI_ISteamInput_ShowBindingPanel", expected: (func(uintptr, InputHandle_t) bool)(nil)},
		{name: "ptrAPI_ISteamInput_GetControllerForGamepadIndex", expected: (func(uintptr, int32) InputHandle_t)(nil)},
		{name: "ptrAPI_ISteamInput_GetGamepadIndexForController", expected: (func(uintptr, InputHandle_t) int32)(nil)},
		{name: "ptrAPI_ISteamInput_GetStringForActionOrigin", expected: (func(uintptr, EInputActionOrigin) string)(nil)},
		{name: "ptrAPI_ISteamInput_GetGlyphForActionOrigin", expected: (func(uintptr, EInputActionOrigin) string)(nil)},
		{name: "ptrAPI_ISteamInput_GetRemotePlaySessionID", expected: (func(uintptr, InputHandle_t) uint32)(nil)},

		{name: "ptrAPI_SteamRemoteStorage", expected: (func() uintptr)(nil)},
		{name: "ptrAPI_ISteamRemoteStorage_FileWrite", expected: (func(uintptr, string, uintptr, int32) bool)(nil)},
		{name: "ptrAPI_ISteamRemoteStorage_FileRead", expected: (func(uintptr, string, uintptr, int32) int32)(nil)},
		{name: "ptrAPI_ISteamRemoteStorage_FileDelete", expected: (func(uintptr, string) bool)(nil)},
		{name: "ptrAPI_ISteamRemoteStorage_GetFileSize", expected: (func(uintptr, string) int32)(nil)},

		{name: "ptrAPI_SteamUser", expected: (func() uintptr)(nil)},
		{name: "ptrAPI_ISteamUser_GetSteamID", expected: (func(uintptr) CSteamID)(nil)},

		{name: "ptrAPI_SteamUserStats", expected: (func() uintptr)(nil)},
		{name: "ptrAPI_ISteamUserStats_GetAchievement", expected: (func(uintptr, string, uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamUserStats_SetAchievement", expected: (func(uintptr, string) bool)(nil)},
		{name: "ptrAPI_ISteamUserStats_ClearAchievement", expected: (func(uintptr, string) bool)(nil)},
		{name: "ptrAPI_ISteamUserStats_StoreStats", expected: (func(uintptr) bool)(nil)},

		{name: "ptrAPI_SteamUtils", expected: (func() uintptr)(nil)},
		{name: "ptrAPI_ISteamUtils_GetSecondsSinceAppActive", expected: (func(uintptr) uint32)(nil)},
		{name: "ptrAPI_ISteamUtils_GetSecondsSinceComputerActive", expected: (func(uintptr) uint32)(nil)},
		{name: "ptrAPI_ISteamUtils_GetConnectedUniverse", expected: (func(uintptr) int32)(nil)},
		{name: "ptrAPI_ISteamUtils_GetServerRealTime", expected: (func(uintptr) uint32)(nil)},
		{name: "ptrAPI_ISteamUtils_GetIPCountry", expected: (func(uintptr) string)(nil)},
		{name: "ptrAPI_ISteamUtils_GetImageSize", expected: (func(uintptr, int32, uintptr, uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamUtils_GetImageRGBA", expected: (func(uintptr, int32, uintptr, int32) bool)(nil)},
		{name: "ptrAPI_ISteamUtils_GetCurrentBatteryPower", expected: (func(uintptr) uint8)(nil)},
		{name: "ptrAPI_ISteamUtils_GetAppID", expected: (func(uintptr) uint32)(nil)},
		{name: "ptrAPI_ISteamUtils_SetOverlayNotificationPosition", expected: (func(uintptr, ENotificationPosition))(nil)},
		{name: "ptrAPI_ISteamUtils_IsAPICallCompleted", expected: (func(uintptr, SteamAPICall_t, uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamUtils_GetAPICallFailureReason", expected: (func(uintptr, SteamAPICall_t) int32)(nil)},
		{name: "ptrAPI_ISteamUtils_GetAPICallResult", expected: (func(uintptr, SteamAPICall_t, uintptr, int32, int32, uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamUtils_GetIPCCallCount", expected: (func(uintptr) uint32)(nil)},
		{name: "ptrAPI_ISteamUtils_IsOverlayEnabled", expected: (func(uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamUtils_BOverlayNeedsPresent", expected: (func(uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamUtils_IsSteamRunningOnSteamDeck", expected: (func(uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamUtils_ShowFloatingGamepadTextInput", expected: (func(uintptr, EFloatingGamepadTextInputMode, int32, int32, int32, int32) bool)(nil)},
		{name: "ptrAPI_ISteamUtils_SetOverlayNotificationInset", expected: (func(uintptr, int32, int32))(nil)},

		{name: "ptrAPI_SteamNetworkingUtils", expected: (func() uintptr)(nil)},
		{name: "ptrAPI_ISteamNetworkingUtils_AllocateMessage", expected: (func(uintptr, int32) uintptr)(nil)},
		{name: "ptrAPI_ISteamNetworkingUtils_InitRelayNetworkAccess", expected: (func(uintptr))(nil)},
		{name: "ptrAPI_ISteamNetworkingUtils_GetLocalTimestamp", expected: (func(uintptr) SteamNetworkingMicroseconds)(nil)},

		{name: "ptrAPI_SteamGameServer", expected: (func() uintptr)(nil)},
		{name: "ptrAPI_ISteamGameServer_SetProduct", expected: (func(uintptr, string))(nil)},
		{name: "ptrAPI_ISteamGameServer_SetGameDescription", expected: (func(uintptr, string))(nil)},
		{name: "ptrAPI_ISteamGameServer_LogOnAnonymous", expected: (func(uintptr))(nil)},
		{name: "ptrAPI_ISteamGameServer_LogOff", expected: (func(uintptr))(nil)},
		{name: "ptrAPI_ISteamGameServer_BLoggedOn", expected: (func(uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamGameServer_GetSteamID", expected: (func(uintptr) CSteamID)(nil)},

		{name: "ptrAPI_SteamNetworkingMessages", expected: (func() uintptr)(nil)},
		{name: "ptrAPI_ISteamNetworkingMessages_SendMessageToUser", expected: (func(uintptr, uintptr, uintptr, uint32, int32, int32) EResult)(nil)},
		{name: "ptrAPI_ISteamNetworkingMessages_ReceiveMessagesOnChannel", expected: (func(uintptr, int32, uintptr, int32) int32)(nil)},
		{name: "ptrAPI_ISteamNetworkingMessages_AcceptSessionWithUser", expected: (func(uintptr, uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamNetworkingMessages_CloseSessionWithUser", expected: (func(uintptr, uintptr) bool)(nil)},
		{name: "ptrAPI_ISteamNetworkingMessages_CloseChannelWithUser", expected: (func(uintptr, uintptr, int32) bool)(nil)},

		{name: "ptrAPI_SteamNetworkingSockets", expected: (func() uintptr)(nil)},
		{name: "ptrAPI_ISteamNetworkingSockets_CreateListenSocketIP", expected: (func(uintptr, uintptr, int32, uintptr) HSteamListenSocket)(nil)},
		{name: "ptrAPI_ISteamNetworkingSockets_CreateListenSocketP2P", expected: (func(uintptr, int32, int32, uintptr) HSteamListenSocket)(nil)},
		{name: "ptrAPI_ISteamNetworkingSockets_ConnectByIPAddress", expected: (func(uintptr, uintptr, int32, uintptr) HSteamNetConnection)(nil)},
		{name: "ptrAPI_ISteamNetworkingSockets_ConnectP2P", expected: (func(uintptr, uintptr, int32, int32, uintptr) HSteamNetConnection)(nil)},
		{name: "ptrAPI_ISteamNetworkingSockets_AcceptConnection", expected: (func(uintptr, HSteamNetConnection) EResult)(nil)},
		{name: "ptrAPI_ISteamNetworkingSockets_CloseConnection", expected: (func(uintptr, HSteamNetConnection, int32, string, bool) bool)(nil)},
		{name: "ptrAPI_ISteamNetworkingSockets_CloseListenSocket", expected: (func(uintptr, HSteamListenSocket) bool)(nil)},
		{name: "ptrAPI_ISteamNetworkingSockets_SendMessageToConnection", expected: (func(uintptr, HSteamNetConnection, uintptr, uint32, int32, uintptr) EResult)(nil)},
		{name: "ptrAPI_ISteamNetworkingSockets_ReceiveMessagesOnConnection", expected: (func(uintptr, HSteamNetConnection, uintptr, int32) int32)(nil)},
		{name: "ptrAPI_ISteamNetworkingSockets_CreatePollGroup", expected: (func(uintptr) HSteamNetPollGroup)(nil)},
		{name: "ptrAPI_ISteamNetworkingSockets_DestroyPollGroup", expected: (func(uintptr, HSteamNetPollGroup) bool)(nil)},
		{name: "ptrAPI_ISteamNetworkingSockets_SetConnectionPollGroup", expected: (func(uintptr, HSteamNetConnection, HSteamNetPollGroup) bool)(nil)},
		{name: "ptrAPI_ISteamNetworkingSockets_ReceiveMessagesOnPollGroup", expected: (func(uintptr, HSteamNetPollGroup, uintptr, int32) int32)(nil)},
	}
}

func allFlatAPISymbols() []string {
	return []string{
		flatAPI_RestartAppIfNecessary,
		flatAPI_InitFlat,
		flatAPI_RunCallbacks,
		flatAPI_Shutdown,
		flatAPI_IsSteamRunning,
		flatAPI_GetSteamInstallPath,
		flatAPI_ReleaseCurrentThreadMemory,

		flatAPI_SteamApps,
		flatAPI_ISteamApps_BIsSubscribed,
		flatAPI_ISteamApps_BIsLowViolence,
		flatAPI_ISteamApps_BIsCybercafe,
		flatAPI_ISteamApps_BIsVACBanned,
		flatAPI_ISteamApps_BGetDLCDataByIndex,
		flatAPI_ISteamApps_BIsDlcInstalled,
		flatAPI_ISteamApps_GetAvailableGameLanguages,
		flatAPI_ISteamApps_BIsSubscribedApp,
		flatAPI_ISteamApps_GetEarliestPurchaseUnixTime,
		flatAPI_ISteamApps_BIsSubscribedFromFreeWeekend,
		flatAPI_ISteamApps_GetAppInstallDir,
		flatAPI_ISteamApps_GetCurrentGameLanguage,
		flatAPI_ISteamApps_GetDLCCount,
		flatAPI_ISteamApps_InstallDLC,
		flatAPI_ISteamApps_UninstallDLC,
		flatAPI_ISteamApps_RequestAppProofOfPurchaseKey,
		flatAPI_ISteamApps_GetCurrentBetaName,
		flatAPI_ISteamApps_MarkContentCorrupt,
		flatAPI_ISteamApps_GetInstalledDepots,
		flatAPI_ISteamApps_BIsAppInstalled,
		flatAPI_ISteamApps_GetAppOwner,
		flatAPI_ISteamApps_GetLaunchQueryParam,
		flatAPI_ISteamApps_GetDlcDownloadProgress,
		flatAPI_ISteamApps_GetAppBuildId,
		flatAPI_ISteamApps_RequestAllProofOfPurchaseKeys,
		flatAPI_ISteamApps_GetFileDetails,
		flatAPI_ISteamApps_GetLaunchCommandLine,
		flatAPI_ISteamApps_BIsSubscribedFromFamilySharing,
		flatAPI_ISteamApps_BIsTimedTrial,
		flatAPI_ISteamApps_SetDlcContext,
		flatAPI_ISteamApps_GetNumBetas,
		flatAPI_ISteamApps_GetBetaInfo,
		flatAPI_ISteamApps_SetActiveBeta,

		flatAPI_SteamFriends,
		flatAPI_ISteamFriends_GetPersonaName,
		flatAPI_ISteamFriends_GetPersonaState,
		flatAPI_ISteamFriends_GetFriendCount,
		flatAPI_ISteamFriends_GetFriendByIndex,
		flatAPI_ISteamFriends_GetFriendRelationship,
		flatAPI_ISteamFriends_GetFriendPersonaState,
		flatAPI_ISteamFriends_GetFriendPersonaName,
		flatAPI_ISteamFriends_GetFriendPersonaNameHistory,
		flatAPI_ISteamFriends_GetFriendSteamLevel,
		flatAPI_ISteamFriends_GetSmallFriendAvatar,
		flatAPI_ISteamFriends_GetMediumFriendAvatar,
		flatAPI_ISteamFriends_GetLargeFriendAvatar,
		flatAPI_ISteamFriends_SetRichPresence,
		flatAPI_ISteamFriends_GetFriendGamePlayed,
		flatAPI_ISteamFriends_InviteUserToGame,
		flatAPI_ISteamFriends_ActivateGameOverlay,
		flatAPI_ISteamFriends_ActivateGameOverlayToUser,
		flatAPI_ISteamFriends_ActivateGameOverlayToWebPage,
		flatAPI_ISteamFriends_ActivateGameOverlayToStore,
		flatAPI_ISteamFriends_ActivateGameOverlayInviteDialog,
		flatAPI_ISteamFriends_ActivateGameOverlayInviteDialogConnectString,

		flatAPI_SteamMatchmaking,
		flatAPI_ISteamMatchmaking_RequestLobbyList,
		flatAPI_ISteamMatchmaking_GetLobbyByIndex,
		flatAPI_ISteamMatchmaking_CreateLobby,
		flatAPI_ISteamMatchmaking_JoinLobby,
		flatAPI_ISteamMatchmaking_LeaveLobby,
		flatAPI_ISteamMatchmaking_InviteUserToLobby,
		flatAPI_ISteamMatchmaking_GetNumLobbyMembers,
		flatAPI_ISteamMatchmaking_GetLobbyMemberByIndex,
		flatAPI_ISteamMatchmaking_GetLobbyData,
		flatAPI_ISteamMatchmaking_SetLobbyData,
		flatAPI_ISteamMatchmaking_GetLobbyOwner,
		flatAPI_ISteamMatchmaking_SetLobbyOwner,
		flatAPI_ISteamMatchmaking_SetLobbyGameServer,
		flatAPI_ISteamMatchmaking_GetLobbyGameServer,
		flatAPI_ISteamMatchmaking_SetLobbyJoinable,
		flatAPI_ISteamMatchmaking_SetLobbyMemberLimit,
		flatAPI_ISteamMatchmaking_SetLobbyType,

		flatAPI_SteamHTTP,
		flatAPI_ISteamHTTP_CreateHTTPRequest,
		flatAPI_ISteamHTTP_SetHTTPRequestHeaderValue,
		flatAPI_ISteamHTTP_SendHTTPRequest,
		flatAPI_ISteamHTTP_GetHTTPResponseBodySize,
		flatAPI_ISteamHTTP_GetHTTPResponseBodyData,
		flatAPI_ISteamHTTP_ReleaseHTTPRequest,

		flatAPI_SteamUGC,
		flatAPI_ISteamUGC_GetNumSubscribedItems,
		flatAPI_ISteamUGC_GetSubscribedItems,

		flatAPI_SteamInventory,
		flatAPI_ISteamInventory_GetResultStatus,
		flatAPI_ISteamInventory_GetResultItems,
		flatAPI_ISteamInventory_DestroyResult,

		flatAPI_SteamInput,
		flatAPI_ISteamInput_GetConnectedControllers,
		flatAPI_ISteamInput_GetInputTypeForHandle,
		flatAPI_ISteamInput_Init,
		flatAPI_ISteamInput_Shutdown,
		flatAPI_ISteamInput_RunFrame,
		flatAPI_ISteamInput_EnableDeviceCallbacks,
		flatAPI_ISteamInput_GetActionSetHandle,
		flatAPI_ISteamInput_ActivateActionSet,
		flatAPI_ISteamInput_GetCurrentActionSet,
		flatAPI_ISteamInput_ActivateActionSetLayer,
		flatAPI_ISteamInput_DeactivateActionSetLayer,
		flatAPI_ISteamInput_DeactivateAllActionSetLayers,
		flatAPI_ISteamInput_GetActiveActionSetLayers,
		flatAPI_ISteamInput_GetDigitalActionHandle,
		flatAPI_ISteamInput_GetDigitalActionData,
		flatAPI_ISteamInput_GetDigitalActionOrigins,
		flatAPI_ISteamInput_GetAnalogActionHandle,
		flatAPI_ISteamInput_GetAnalogActionData,
		flatAPI_ISteamInput_GetAnalogActionOrigins,
		flatAPI_ISteamInput_StopAnalogActionMomentum,
		flatAPI_ISteamInput_GetMotionData,
		flatAPI_ISteamInput_TriggerVibration,
		flatAPI_ISteamInput_TriggerVibrationExtended,
		flatAPI_ISteamInput_TriggerSimpleHapticEvent,
		flatAPI_ISteamInput_SetLEDColor,
		flatAPI_ISteamInput_ShowBindingPanel,
		flatAPI_ISteamInput_GetControllerForGamepadIndex,
		flatAPI_ISteamInput_GetGamepadIndexForController,
		flatAPI_ISteamInput_GetStringForActionOrigin,
		flatAPI_ISteamInput_GetGlyphForActionOrigin,
		flatAPI_ISteamInput_GetRemotePlaySessionID,

		flatAPI_SteamRemoteStorage,
		flatAPI_ISteamRemoteStorage_FileWrite,
		flatAPI_ISteamRemoteStorage_FileRead,
		flatAPI_ISteamRemoteStorage_FileDelete,
		flatAPI_ISteamRemoteStorage_GetFileSize,

		flatAPI_SteamUser,
		flatAPI_ISteamUser_GetSteamID,

		flatAPI_SteamUserStats,
		flatAPI_ISteamUserStats_GetAchievement,
		flatAPI_ISteamUserStats_SetAchievement,
		flatAPI_ISteamUserStats_ClearAchievement,
		flatAPI_ISteamUserStats_StoreStats,

		flatAPI_SteamUtils,
		flatAPI_ISteamUtils_GetSecondsSinceAppActive,
		flatAPI_ISteamUtils_GetSecondsSinceComputerActive,
		flatAPI_ISteamUtils_GetConnectedUniverse,
		flatAPI_ISteamUtils_GetServerRealTime,
		flatAPI_ISteamUtils_GetIPCountry,
		flatAPI_ISteamUtils_GetImageSize,
		flatAPI_ISteamUtils_GetImageRGBA,
		flatAPI_ISteamUtils_GetCurrentBatteryPower,
		flatAPI_ISteamUtils_GetAppID,
		flatAPI_ISteamUtils_SetOverlayNotificationPosition,
		flatAPI_ISteamUtils_IsAPICallCompleted,
		flatAPI_ISteamUtils_GetAPICallFailureReason,
		flatAPI_ISteamUtils_GetAPICallResult,
		flatAPI_ISteamUtils_GetIPCCallCount,
		flatAPI_ISteamUtils_IsOverlayEnabled,
		flatAPI_ISteamUtils_BOverlayNeedsPresent,
		flatAPI_ISteamUtils_IsSteamRunningOnSteamDeck,
		flatAPI_ISteamUtils_ShowFloatingGamepadTextInput,
		flatAPI_ISteamUtils_SetOverlayNotificationInset,

		flatAPI_SteamNetworkingUtils,
		flatAPI_ISteamNetworkingUtils_AllocateMessage,
		flatAPI_ISteamNetworkingUtils_InitRelayNetworkAccess,
		flatAPI_ISteamNetworkingUtils_GetLocalTimestamp,

		flatAPI_SteamNetworkingMessages,
		flatAPI_ISteamNetworkingMessages_SendMessageToUser,
		flatAPI_ISteamNetworkingMessages_ReceiveMessagesOnChannel,
		flatAPI_ISteamNetworkingMessages_AcceptSessionWithUser,
		flatAPI_ISteamNetworkingMessages_CloseSessionWithUser,
		flatAPI_ISteamNetworkingMessages_CloseChannelWithUser,

		flatAPI_SteamNetworkingSockets,
		flatAPI_ISteamNetworkingSockets_CreateListenSocketIP,
		flatAPI_ISteamNetworkingSockets_CreateListenSocketP2P,
		flatAPI_ISteamNetworkingSockets_ConnectByIPAddress,
		flatAPI_ISteamNetworkingSockets_ConnectP2P,
		flatAPI_ISteamNetworkingSockets_AcceptConnection,
		flatAPI_ISteamNetworkingSockets_CloseConnection,
		flatAPI_ISteamNetworkingSockets_CloseListenSocket,
		flatAPI_ISteamNetworkingSockets_SendMessageToConnection,
		flatAPI_ISteamNetworkingSockets_ReceiveMessagesOnConnection,
		flatAPI_ISteamNetworkingSockets_CreatePollGroup,
		flatAPI_ISteamNetworkingSockets_DestroyPollGroup,
		flatAPI_ISteamNetworkingSockets_SetConnectionPollGroup,
		flatAPI_ISteamNetworkingSockets_ReceiveMessagesOnPollGroup,

		flatAPI_SteamGameServer,
		flatAPI_ISteamGameServer_SetProduct,
		flatAPI_ISteamGameServer_SetGameDescription,
		flatAPI_ISteamGameServer_LogOnAnonymous,
		flatAPI_ISteamGameServer_LogOff,
		flatAPI_ISteamGameServer_BLoggedOn,
		flatAPI_ISteamGameServer_GetSteamID,
	}
}
