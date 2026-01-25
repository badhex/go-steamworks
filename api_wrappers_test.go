// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The go-steamworks Authors

package steamworks

import (
	"testing"
	"unsafe"

	"github.com/jupiterrider/ffi"
)

const testAppID AppId_t = 480

func TestAPIWrappers(t *testing.T) {
	t.Helper()

	setInputOverrides(t)
	setSteamAppsStubs()
	setSteamFriendsStubs()
	setSteamMatchmakingStubs()
	setSteamHTTPStubs()
	setSteamUGCStubs()
	setSteamInventoryStubs()
	setSteamInputStubs()
	setSteamRemoteStorageStubs()
	setSteamUserStubs()
	setSteamUserStatsStubs()
	setSteamUtilsStubs()
	setSteamNetworkingUtilsStubs()
	setSteamGameServerStubs()
	setSteamNetworkingMessagesStubs()
	setSteamNetworkingSocketsStubs()

	apps := steamApps(0x1)
	appID, available, name, ok := apps.BGetDLCDataByIndex(1)
	if appID != testAppID || !available || name != "SpaceWar" || !ok {
		t.Fatalf("BGetDLCDataByIndex = (%v,%v,%q,%v)", appID, available, name, ok)
	}
	if !apps.BIsSubscribed() || apps.BIsLowViolence() || apps.BIsCybercafe() || apps.BIsVACBanned() {
		t.Fatalf("unexpected flags from steamApps")
	}
	if !apps.BIsDlcInstalled(testAppID) || !apps.BIsSubscribedApp(testAppID) {
		t.Fatalf("expected DLC/subscribed app to be true")
	}
	if !apps.BIsSubscribedFromFreeWeekend() || apps.BIsSubscribedFromFamilySharing() {
		t.Fatalf("unexpected subscription weekend/family values")
	}
	allowed, played, trialOK := apps.BIsTimedTrial()
	if !trialOK || allowed != 120 || played != 30 {
		t.Fatalf("BIsTimedTrial = (%d,%d,%v)", allowed, played, trialOK)
	}
	if !apps.BIsAppInstalled(testAppID) {
		t.Fatalf("expected app to be installed")
	}
	if apps.GetAvailableGameLanguages() != "english" {
		t.Fatalf("unexpected available languages")
	}
	if apps.GetEarliestPurchaseUnixTime(testAppID) != 123 {
		t.Fatalf("unexpected earliest purchase time")
	}
	if apps.GetAppInstallDir(testAppID) != "/tmp/spacewar" {
		t.Fatalf("unexpected install dir")
	}
	if apps.GetCurrentGameLanguage() != "english" {
		t.Fatalf("unexpected current game language")
	}
	if apps.GetDLCCount() != 2 {
		t.Fatalf("unexpected dlc count")
	}
	betaName, betaOK := apps.GetCurrentBetaName()
	if !betaOK || betaName != "beta" {
		t.Fatalf("unexpected beta name")
	}
	depots := apps.GetInstalledDepots(testAppID)
	if len(depots) != 2 || depots[0] != 11 || depots[1] != 22 {
		t.Fatalf("unexpected depots: %v", depots)
	}
	if apps.GetAppOwner() != 77 {
		t.Fatalf("unexpected app owner")
	}
	if apps.GetLaunchQueryParam("mode") != "test" {
		t.Fatalf("unexpected launch query param")
	}
	downloaded, total, progressOK := apps.GetDlcDownloadProgress(testAppID)
	if !progressOK || downloaded != 10 || total != 20 {
		t.Fatalf("unexpected dlc progress")
	}
	if apps.GetAppBuildId() != 1001 {
		t.Fatalf("unexpected build id")
	}
	if apps.GetFileDetails("file") != 42 {
		t.Fatalf("unexpected file details call")
	}
	if apps.GetLaunchCommandLine(0) != "-dev" {
		t.Fatalf("unexpected launch command line")
	}
	totalBetas, availableBetas, privateBetas := apps.GetNumBetas()
	if totalBetas != 2 || availableBetas != 1 || privateBetas != 1 {
		t.Fatalf("unexpected beta counts")
	}
	flags, buildID, betaName, betaDesc, betaInfoOK := apps.GetBetaInfo(0)
	if !betaInfoOK || flags != 9 || buildID != 10 || betaName != "beta" || betaDesc != "desc" {
		t.Fatalf("unexpected beta info")
	}
	apps.InstallDLC(testAppID)
	apps.UninstallDLC(testAppID)
	apps.RequestAppProofOfPurchaseKey(testAppID)
	apps.RequestAllProofOfPurchaseKeys()
	if !apps.MarkContentCorrupt(true) {
		t.Fatalf("expected MarkContentCorrupt to succeed")
	}
	if !apps.SetDlcContext(testAppID) {
		t.Fatalf("expected SetDlcContext to succeed")
	}
	if !apps.SetActiveBeta("beta") {
		t.Fatalf("expected SetActiveBeta to succeed")
	}

	friends := steamFriends(0x2)
	if friends.GetPersonaName() != "tester" {
		t.Fatalf("unexpected persona name")
	}
	if friends.GetPersonaState() != EPersonaStateOnline {
		t.Fatalf("unexpected persona state")
	}
	if friends.GetFriendCount(0) != 2 {
		t.Fatalf("unexpected friend count")
	}
	if friends.GetFriendByIndex(0, 0) != 101 {
		t.Fatalf("unexpected friend by index")
	}
	if friends.GetFriendRelationship(101) != EFriendRelationshipFriend {
		t.Fatalf("unexpected friend relationship")
	}
	if friends.GetFriendPersonaState(101) != EPersonaStateOffline {
		t.Fatalf("unexpected friend persona state")
	}
	if friends.GetFriendPersonaName(101) != "buddy" {
		t.Fatalf("unexpected friend persona name")
	}
	if friends.GetFriendPersonaNameHistory(101, 0) != "buddy-old" {
		t.Fatalf("unexpected friend persona name history")
	}
	if friends.GetFriendSteamLevel(101) != 5 {
		t.Fatalf("unexpected friend steam level")
	}
	if friends.GetSmallFriendAvatar(101) != 1 || friends.GetMediumFriendAvatar(101) != 2 || friends.GetLargeFriendAvatar(101) != 3 {
		t.Fatalf("unexpected friend avatars")
	}
	if !friends.SetRichPresence("k", "v") {
		t.Fatalf("expected rich presence to succeed")
	}
	if _, ok := friends.GetFriendGamePlayed(101); !ok {
		t.Fatalf("expected friend game played")
	}
	if !friends.InviteUserToGame(101, "connect") {
		t.Fatalf("expected invite user to game to succeed")
	}
	friends.ActivateGameOverlay("dialog")
	friends.ActivateGameOverlayToUser("dialog", 101)
	friends.ActivateGameOverlayToWebPage("url", EActivateGameOverlayToWebPageMode_Modal)
	friends.ActivateGameOverlayToStore(testAppID, EOverlayToStoreFlag_AddToCart)
	friends.ActivateGameOverlayInviteDialog(101)
	friends.ActivateGameOverlayInviteDialogConnectString("connect")

	matchmaking := steamMatchmaking(0x3)
	if matchmaking.RequestLobbyList() != 7 {
		t.Fatalf("unexpected lobby list call")
	}
	if matchmaking.GetLobbyByIndex(0) != 111 {
		t.Fatalf("unexpected lobby id")
	}
	if matchmaking.CreateLobby(ELobbyType_Public, 4) != 8 {
		t.Fatalf("unexpected create lobby")
	}
	if matchmaking.JoinLobby(111) != 9 {
		t.Fatalf("unexpected join lobby")
	}
	matchmaking.LeaveLobby(111)
	if !matchmaking.InviteUserToLobby(111, 101) {
		t.Fatalf("expected invite user to lobby")
	}
	if matchmaking.GetNumLobbyMembers(111) != 3 {
		t.Fatalf("unexpected lobby member count")
	}
	if matchmaking.GetLobbyMemberByIndex(111, 0) != 101 {
		t.Fatalf("unexpected lobby member by index")
	}
	if matchmaking.GetLobbyData(111, "key") != "value" {
		t.Fatalf("unexpected lobby data")
	}
	if !matchmaking.SetLobbyData(111, "key", "value") {
		t.Fatalf("expected set lobby data")
	}
	if matchmaking.GetLobbyOwner(111) != 101 {
		t.Fatalf("unexpected lobby owner")
	}
	if !matchmaking.SetLobbyOwner(111, 102) {
		t.Fatalf("expected set lobby owner")
	}
	matchmaking.SetLobbyGameServer(111, 127001, 27015, 103)
	ip, port, server, ok := matchmaking.GetLobbyGameServer(111)
	if !ok || ip != 127001 || port != 27015 || server != 103 {
		t.Fatalf("unexpected lobby game server")
	}
	if !matchmaking.SetLobbyJoinable(111, true) || !matchmaking.SetLobbyMemberLimit(111, 4) {
		t.Fatalf("unexpected lobby settings")
	}
	if !matchmaking.SetLobbyType(111, ELobbyType_Private) {
		t.Fatalf("unexpected lobby type")
	}

	http := steamHTTP(0x4)
	if http.CreateHTTPRequest(EHTTPMethodGET, "url") != 200 {
		t.Fatalf("unexpected http request handle")
	}
	if !http.SetHTTPRequestHeaderValue(200, "h", "v") {
		t.Fatalf("expected header set")
	}
	call, ok := http.SendHTTPRequest(200)
	if !ok || call != 33 {
		t.Fatalf("unexpected send http request")
	}
	size, ok := http.GetHTTPResponseBodySize(200)
	if !ok || size != 3 {
		t.Fatalf("unexpected http response size")
	}
	buf := make([]byte, 3)
	if !http.GetHTTPResponseBodyData(200, buf) || string(buf) != "hey" {
		t.Fatalf("unexpected http response body data")
	}
	if !http.ReleaseHTTPRequest(200) {
		t.Fatalf("unexpected release request")
	}

	ugc := steamUGC(0x5)
	if ugc.GetNumSubscribedItems(false) != 2 {
		t.Fatalf("unexpected ugc count")
	}
	items := ugc.GetSubscribedItems(false)
	if len(items) != 2 || items[0] != 401 || items[1] != 402 {
		t.Fatalf("unexpected ugc items")
	}

	inventory := steamInventory(0x6)
	if inventory.GetResultStatus(10) != EResultOK {
		t.Fatalf("unexpected inventory result status")
	}
	out := make([]SteamItemDetails, 1)
	amount, ok := inventory.GetResultItems(10, out)
	if !ok || amount != 1 || out[0].Definition != 501 {
		t.Fatalf("unexpected inventory result items")
	}
	inventory.DestroyResult(10)

	input := steamInput(0x7)
	handles := input.GetConnectedControllers()
	if len(handles) != 2 || handles[0] != 601 || handles[1] != 602 {
		t.Fatalf("unexpected connected controllers")
	}
	if input.GetInputTypeForHandle(handles[0]) != ESteamInputType_SteamController {
		t.Fatalf("unexpected input type")
	}
	if !input.Init(true) {
		t.Fatalf("expected input init")
	}
	input.Shutdown()
	input.RunFrame()
	input.EnableDeviceCallbacks()
	if input.GetActionSetHandle("action") != 701 {
		t.Fatalf("unexpected action set handle")
	}
	input.ActivateActionSet(handles[0], 701)
	if input.GetCurrentActionSet(handles[0]) != 702 {
		t.Fatalf("unexpected current action set handle")
	}
	input.ActivateActionSetLayer(handles[0], 703)
	input.DeactivateActionSetLayer(handles[0], 703)
	input.DeactivateAllActionSetLayers(handles[0])
	actionSets := make([]InputActionSetHandle_t, 4)
	if input.GetActiveActionSetLayers(handles[0], actionSets) != 2 {
		t.Fatalf("unexpected active action set layers")
	}
	if input.GetDigitalActionHandle("jump") != 801 {
		t.Fatalf("unexpected digital action handle")
	}
	if data := input.GetDigitalActionData(handles[0], 801); !data.State || !data.Active {
		t.Fatalf("unexpected digital action data")
	}
	origins := make([]EInputActionOrigin, 4)
	if input.GetDigitalActionOrigins(handles[0], 701, 801, origins) != 2 {
		t.Fatalf("unexpected digital action origins")
	}
	if input.GetAnalogActionHandle("move") != 802 {
		t.Fatalf("unexpected analog action handle")
	}
	if data := input.GetAnalogActionData(handles[0], 802); data.X != 1 || data.Y != 2 || !data.Active {
		t.Fatalf("unexpected analog action data")
	}
	if input.GetAnalogActionOrigins(handles[0], 701, 802, origins) != 2 {
		t.Fatalf("unexpected analog action origins")
	}
	input.StopAnalogActionMomentum(handles[0], 802)
	if data := input.GetMotionData(handles[0]); data.RotVelX != 3 {
		t.Fatalf("unexpected motion data")
	}
	input.TriggerVibration(handles[0], 1, 2)
	input.TriggerVibrationExtended(handles[0], 1, 2, 3, 4)
	input.TriggerSimpleHapticEvent(handles[0], ESteamControllerPad_Left, 1, 2, 3)
	input.SetLEDColor(handles[0], 1, 2, 3, ESteamInputLEDFlag_SetColor)
	if !input.ShowBindingPanel(handles[0]) {
		t.Fatalf("expected binding panel to open")
	}
	if input.GetControllerForGamepadIndex(0) != 601 {
		t.Fatalf("unexpected controller for gamepad index")
	}
	if input.GetGamepadIndexForController(handles[0]) != 0 {
		t.Fatalf("unexpected gamepad index for controller")
	}
	if input.GetStringForActionOrigin(EInputActionOrigin_SteamController_A) != "A" {
		t.Fatalf("unexpected string for action origin")
	}
	if input.GetGlyphForActionOrigin(EInputActionOrigin_SteamController_A) != "glyph" {
		t.Fatalf("unexpected glyph for action origin")
	}
	if input.GetRemotePlaySessionID(handles[0]) != 901 {
		t.Fatalf("unexpected remote play session id")
	}

	storage := steamRemoteStorage(0x8)
	if !storage.FileWrite("file", []byte("data")) {
		t.Fatalf("expected file write")
	}
	read := make([]byte, 3)
	if storage.FileRead("file", read) != 3 || string(read) != "hey" {
		t.Fatalf("unexpected file read")
	}
	if !storage.FileDelete("file") {
		t.Fatalf("expected file delete")
	}
	if storage.GetFileSize("file") != 3 {
		t.Fatalf("unexpected file size")
	}

	user := steamUser(0x9)
	if user.GetSteamID() != 1001 {
		t.Fatalf("unexpected steam id")
	}

	stats := steamUserStats(0x10)
	if achieved, ok := stats.GetAchievement("ach"); !ok || !achieved {
		t.Fatalf("unexpected achievement")
	}
	if !stats.SetAchievement("ach") || !stats.ClearAchievement("ach") || !stats.StoreStats() {
		t.Fatalf("unexpected user stats")
	}

	utils := steamUtils(0x11)
	if utils.GetSecondsSinceAppActive() != 5 || utils.GetSecondsSinceComputerActive() != 6 {
		t.Fatalf("unexpected seconds since")
	}
	if utils.GetConnectedUniverse() != EUniversePublic {
		t.Fatalf("unexpected universe")
	}
	if utils.GetServerRealTime() != 100 {
		t.Fatalf("unexpected server real time")
	}
	if utils.GetIPCountry() != "US" {
		t.Fatalf("unexpected ip country")
	}
	width, height, ok := utils.GetImageSize(1)
	if !ok || width != 2 || height != 2 {
		t.Fatalf("unexpected image size")
	}
	image := make([]byte, 4)
	if !utils.GetImageRGBA(1, image) || string(image) != "RGBA" {
		t.Fatalf("unexpected image rgba")
	}
	if utils.GetCurrentBatteryPower() != 88 {
		t.Fatalf("unexpected battery power")
	}
	if utils.GetAppID() != 480 {
		t.Fatalf("unexpected app id")
	}
	utils.SetOverlayNotificationPosition(ENotificationPositionTopLeft)
	utils.SetOverlayNotificationInset(1, 2)
	if failed, ok := utils.IsAPICallCompleted(12); !ok || failed {
		t.Fatalf("unexpected api call completed")
	}
	if utils.GetAPICallFailureReason(12) != ESteamAPICallFailureNone {
		t.Fatalf("unexpected api call failure reason")
	}
	if failed, ok := utils.GetAPICallResult(12, 0, 0, 0); !ok || failed {
		t.Fatalf("unexpected api call result")
	}
	if utils.GetIPCCallCount() != 7 {
		t.Fatalf("unexpected ipc call count")
	}
	if !utils.IsOverlayEnabled() || utils.BOverlayNeedsPresent() || utils.IsSteamRunningOnSteamDeck() {
		t.Fatalf("unexpected overlay/steam deck state")
	}
	if !utils.ShowFloatingGamepadTextInput(EFloatingGamepadTextInputMode_ModeSingleLine, 1, 2, 3, 4) {
		t.Fatalf("unexpected floating gamepad input")
	}

	netUtils := steamNetworkingUtils(0x12)
	if msg := netUtils.AllocateMessage(64); msg == nil || msg.MessageNumber != 99 {
		t.Fatalf("unexpected allocated message")
	}
	netUtils.InitRelayNetworkAccess()
	if netUtils.GetLocalTimestamp() != 1000 {
		t.Fatalf("unexpected local timestamp")
	}

	gameServer := steamGameServer(0x13)
	gameServer.SetProduct("product")
	gameServer.SetGameDescription("desc")
	gameServer.LogOnAnonymous()
	gameServer.LogOff()
	if !gameServer.BLoggedOn() {
		t.Fatalf("unexpected logged on")
	}
	if gameServer.GetSteamID() != 2001 {
		t.Fatalf("unexpected game server steam id")
	}

	netMessages := steamNetworkingMessages(0x14)
	identity := &SteamNetworkingIdentity{}
	if netMessages.SendMessageToUser(identity, []byte("ping"), SteamNetworkingSend_Reliable, 1) != EResultOK {
		t.Fatalf("unexpected send message result")
	}
	if msgs := netMessages.ReceiveMessagesOnChannel(1, 2); len(msgs) != 2 {
		t.Fatalf("unexpected receive messages count")
	}
	if !netMessages.AcceptSessionWithUser(identity) || !netMessages.CloseSessionWithUser(identity) || !netMessages.CloseChannelWithUser(identity, 1) {
		t.Fatalf("unexpected session operations")
	}

	netSockets := steamNetworkingSockets(0x15)
	if netSockets.CreateListenSocketIP(&SteamNetworkingIPAddr{}, nil) != 3001 {
		t.Fatalf("unexpected listen socket ip")
	}
	if netSockets.CreateListenSocketP2P(1, nil) != 3002 {
		t.Fatalf("unexpected listen socket p2p")
	}
	if netSockets.ConnectByIPAddress(&SteamNetworkingIPAddr{}, nil) != 3003 {
		t.Fatalf("unexpected connect by ip")
	}
	if netSockets.ConnectP2P(identity, 1, nil) != 3004 {
		t.Fatalf("unexpected connect p2p")
	}
	if netSockets.AcceptConnection(4001) != EResultOK {
		t.Fatalf("unexpected accept connection")
	}
	if !netSockets.CloseConnection(4001, 1, "bye", true) {
		t.Fatalf("unexpected close connection")
	}
	if !netSockets.CloseListenSocket(3001) {
		t.Fatalf("unexpected close listen socket")
	}
	if result, messageNumber := netSockets.SendMessageToConnection(4001, []byte("ping"), SteamNetworkingSend_Reliable); result != EResultOK || messageNumber != 123 {
		t.Fatalf("unexpected send message to connection")
	}
	if msgs := netSockets.ReceiveMessagesOnConnection(4001, 2); len(msgs) != 2 {
		t.Fatalf("unexpected receive messages on connection")
	}
	if netSockets.CreatePollGroup() != 5001 {
		t.Fatalf("unexpected create poll group")
	}
	if !netSockets.DestroyPollGroup(5001) {
		t.Fatalf("unexpected destroy poll group")
	}
	if !netSockets.SetConnectionPollGroup(4001, 5001) {
		t.Fatalf("unexpected set connection poll group")
	}
	if msgs := netSockets.ReceiveMessagesOnPollGroup(5001, 2); len(msgs) != 2 {
		t.Fatalf("unexpected receive messages on poll group")
	}
}

func setInputOverrides(t *testing.T) {
	t.Helper()
	ffi.SetInputDigitalActionDataOverride(func(fn uintptr, self uintptr, inputHandle uint64, actionHandle uint64) ffi.InputDigitalActionData {
		return ffi.InputDigitalActionData{State: true, Active: true}
	})
	ffi.SetInputAnalogActionDataOverride(func(fn uintptr, self uintptr, inputHandle uint64, actionHandle uint64) ffi.InputAnalogActionData {
		return ffi.InputAnalogActionData{Mode: int32(EInputSourceMode_AbsoluteMouse), X: 1, Y: 2, Active: true}
	})
	ffi.SetInputMotionDataOverride(func(fn uintptr, self uintptr, inputHandle uint64) ffi.InputMotionData {
		return ffi.InputMotionData{RotVelX: 3}
	})
	t.Cleanup(func() {
		ffi.SetInputDigitalActionDataOverride(nil)
		ffi.SetInputAnalogActionDataOverride(nil)
		ffi.SetInputMotionDataOverride(nil)
	})
}

func setSteamAppsStubs() {
	ptrAPI_ISteamApps_BGetDLCDataByIndex = func(self uintptr, idx int32, appIDPtr, availablePtr, namePtr uintptr, nameLen int32) bool {
		*(*AppId_t)(unsafe.Pointer(appIDPtr)) = testAppID
		*(*bool)(unsafe.Pointer(availablePtr)) = true
		writeCString(namePtr, nameLen, "SpaceWar")
		return true
	}
	ptrAPI_ISteamApps_BIsSubscribed = func(uintptr) bool { return true }
	ptrAPI_ISteamApps_BIsLowViolence = func(uintptr) bool { return false }
	ptrAPI_ISteamApps_BIsCybercafe = func(uintptr) bool { return false }
	ptrAPI_ISteamApps_BIsVACBanned = func(uintptr) bool { return false }
	ptrAPI_ISteamApps_BIsDlcInstalled = func(uintptr, AppId_t) bool { return true }
	ptrAPI_ISteamApps_BIsSubscribedApp = func(uintptr, AppId_t) bool { return true }
	ptrAPI_ISteamApps_BIsSubscribedFromFreeWeekend = func(uintptr) bool { return true }
	ptrAPI_ISteamApps_BIsSubscribedFromFamilySharing = func(uintptr) bool { return false }
	ptrAPI_ISteamApps_BIsTimedTrial = func(self uintptr, allowedPtr, playedPtr uintptr) bool {
		*(*uint32)(unsafe.Pointer(allowedPtr)) = 120
		*(*uint32)(unsafe.Pointer(playedPtr)) = 30
		return true
	}
	ptrAPI_ISteamApps_BIsAppInstalled = func(uintptr, AppId_t) bool { return true }
	ptrAPI_ISteamApps_GetAvailableGameLanguages = func(uintptr) string { return "english" }
	ptrAPI_ISteamApps_GetEarliestPurchaseUnixTime = func(uintptr, AppId_t) uint32 { return 123 }
	ptrAPI_ISteamApps_GetAppInstallDir = func(self uintptr, appID AppId_t, pathPtr uintptr, pathLen int32) int32 {
		return writeCString(pathPtr, pathLen, "/tmp/spacewar")
	}
	ptrAPI_ISteamApps_GetCurrentGameLanguage = func(uintptr) string { return "english" }
	ptrAPI_ISteamApps_GetDLCCount = func(uintptr) int32 { return 2 }
	ptrAPI_ISteamApps_GetCurrentBetaName = func(self uintptr, namePtr uintptr, nameLen int32) bool {
		writeCString(namePtr, nameLen, "beta")
		return true
	}
	ptrAPI_ISteamApps_GetInstalledDepots = func(self uintptr, appID AppId_t, depotsPtr uintptr, max uint32) uint32 {
		depots := unsafe.Slice((*DepotId_t)(unsafe.Pointer(depotsPtr)), int(max))
		depots[0] = 11
		depots[1] = 22
		return 2
	}
	ptrAPI_ISteamApps_GetAppOwner = func(uintptr) CSteamID { return 77 }
	ptrAPI_ISteamApps_GetLaunchQueryParam = func(uintptr, string) string { return "test" }
	ptrAPI_ISteamApps_GetDlcDownloadProgress = func(self uintptr, appID AppId_t, downloadedPtr, totalPtr uintptr) bool {
		*(*uint64)(unsafe.Pointer(downloadedPtr)) = 10
		*(*uint64)(unsafe.Pointer(totalPtr)) = 20
		return true
	}
	ptrAPI_ISteamApps_GetAppBuildId = func(uintptr) int32 { return 1001 }
	ptrAPI_ISteamApps_GetFileDetails = func(uintptr, string) SteamAPICall_t { return 42 }
	ptrAPI_ISteamApps_GetLaunchCommandLine = func(self uintptr, bufPtr uintptr, bufLen int32) int32 {
		return writeString(bufPtr, bufLen, "-dev")
	}
	ptrAPI_ISteamApps_GetNumBetas = func(self uintptr, availablePtr, privatePtr uintptr) int32 {
		*(*int32)(unsafe.Pointer(availablePtr)) = 1
		*(*int32)(unsafe.Pointer(privatePtr)) = 1
		return 2
	}
	ptrAPI_ISteamApps_GetBetaInfo = func(self uintptr, idx int32, flagsPtr, buildIDPtr, namePtr uintptr, nameLen int32, descPtr uintptr, descLen int32) bool {
		*(*uint32)(unsafe.Pointer(flagsPtr)) = 9
		*(*uint32)(unsafe.Pointer(buildIDPtr)) = 10
		writeCString(namePtr, nameLen, "beta")
		writeCString(descPtr, descLen, "desc")
		return true
	}
	ptrAPI_ISteamApps_InstallDLC = func(uintptr, AppId_t) {}
	ptrAPI_ISteamApps_UninstallDLC = func(uintptr, AppId_t) {}
	ptrAPI_ISteamApps_RequestAppProofOfPurchaseKey = func(uintptr, AppId_t) {}
	ptrAPI_ISteamApps_RequestAllProofOfPurchaseKeys = func(uintptr) {}
	ptrAPI_ISteamApps_MarkContentCorrupt = func(uintptr, bool) bool { return true }
	ptrAPI_ISteamApps_SetDlcContext = func(uintptr, AppId_t) bool { return true }
	ptrAPI_ISteamApps_SetActiveBeta = func(uintptr, string) bool { return true }
}

func setSteamFriendsStubs() {
	ptrAPI_ISteamFriends_GetPersonaName = func(uintptr) string { return "tester" }
	ptrAPI_ISteamFriends_GetPersonaState = func(uintptr) int32 { return int32(EPersonaStateOnline) }
	ptrAPI_ISteamFriends_GetFriendCount = func(uintptr, int32) int32 { return 2 }
	ptrAPI_ISteamFriends_GetFriendByIndex = func(uintptr, int32, int32) CSteamID { return 101 }
	ptrAPI_ISteamFriends_GetFriendRelationship = func(uintptr, CSteamID) int32 { return int32(EFriendRelationshipFriend) }
	ptrAPI_ISteamFriends_GetFriendPersonaState = func(uintptr, CSteamID) int32 { return int32(EPersonaStateOffline) }
	ptrAPI_ISteamFriends_GetFriendPersonaName = func(uintptr, CSteamID) string { return "buddy" }
	ptrAPI_ISteamFriends_GetFriendPersonaNameHistory = func(uintptr, CSteamID, int32) string { return "buddy-old" }
	ptrAPI_ISteamFriends_GetFriendSteamLevel = func(uintptr, CSteamID) int32 { return 5 }
	ptrAPI_ISteamFriends_GetSmallFriendAvatar = func(uintptr, CSteamID) int32 { return 1 }
	ptrAPI_ISteamFriends_GetMediumFriendAvatar = func(uintptr, CSteamID) int32 { return 2 }
	ptrAPI_ISteamFriends_GetLargeFriendAvatar = func(uintptr, CSteamID) int32 { return 3 }
	ptrAPI_ISteamFriends_SetRichPresence = func(uintptr, string, string) bool { return true }
	ptrAPI_ISteamFriends_GetFriendGamePlayed = func(self uintptr, friend CSteamID, gameInfoPtr uintptr) bool {
		info := (*FriendGameInfo)(unsafe.Pointer(gameInfoPtr))
		info.GameID = 200
		return true
	}
	ptrAPI_ISteamFriends_InviteUserToGame = func(uintptr, CSteamID, string) bool { return true }
	ptrAPI_ISteamFriends_ActivateGameOverlay = func(uintptr, string) {}
	ptrAPI_ISteamFriends_ActivateGameOverlayToUser = func(uintptr, string, CSteamID) {}
	ptrAPI_ISteamFriends_ActivateGameOverlayToWebPage = func(uintptr, string, EActivateGameOverlayToWebPageMode) {}
	ptrAPI_ISteamFriends_ActivateGameOverlayToStore = func(uintptr, AppId_t, EOverlayToStoreFlag) {}
	ptrAPI_ISteamFriends_ActivateGameOverlayInviteDialog = func(uintptr, CSteamID) {}
	ptrAPI_ISteamFriends_ActivateGameOverlayInviteDialogConnectString = func(uintptr, string) {}
}

func setSteamMatchmakingStubs() {
	ptrAPI_ISteamMatchmaking_RequestLobbyList = func(uintptr) SteamAPICall_t { return 7 }
	ptrAPI_ISteamMatchmaking_GetLobbyByIndex = func(uintptr, int32) CSteamID { return 111 }
	ptrAPI_ISteamMatchmaking_CreateLobby = func(uintptr, ELobbyType, int32) SteamAPICall_t { return 8 }
	ptrAPI_ISteamMatchmaking_JoinLobby = func(uintptr, CSteamID) SteamAPICall_t { return 9 }
	ptrAPI_ISteamMatchmaking_LeaveLobby = func(uintptr, CSteamID) {}
	ptrAPI_ISteamMatchmaking_InviteUserToLobby = func(uintptr, CSteamID, CSteamID) bool { return true }
	ptrAPI_ISteamMatchmaking_GetNumLobbyMembers = func(uintptr, CSteamID) int32 { return 3 }
	ptrAPI_ISteamMatchmaking_GetLobbyMemberByIndex = func(uintptr, CSteamID, int32) CSteamID { return 101 }
	ptrAPI_ISteamMatchmaking_GetLobbyData = func(uintptr, CSteamID, string) string { return "value" }
	ptrAPI_ISteamMatchmaking_SetLobbyData = func(uintptr, CSteamID, string, string) bool { return true }
	ptrAPI_ISteamMatchmaking_GetLobbyOwner = func(uintptr, CSteamID) CSteamID { return 101 }
	ptrAPI_ISteamMatchmaking_SetLobbyOwner = func(uintptr, CSteamID, CSteamID) bool { return true }
	ptrAPI_ISteamMatchmaking_SetLobbyGameServer = func(uintptr, CSteamID, uint32, uint16, CSteamID) {}
	ptrAPI_ISteamMatchmaking_GetLobbyGameServer = func(self uintptr, lobbyID CSteamID, ipPtr, portPtr, serverPtr uintptr) bool {
		*(*uint32)(unsafe.Pointer(ipPtr)) = 127001
		*(*uint16)(unsafe.Pointer(portPtr)) = 27015
		*(*CSteamID)(unsafe.Pointer(serverPtr)) = 103
		return true
	}
	ptrAPI_ISteamMatchmaking_SetLobbyJoinable = func(uintptr, CSteamID, bool) bool { return true }
	ptrAPI_ISteamMatchmaking_SetLobbyMemberLimit = func(uintptr, CSteamID, int32) bool { return true }
	ptrAPI_ISteamMatchmaking_SetLobbyType = func(uintptr, CSteamID, ELobbyType) bool { return true }
}

func setSteamHTTPStubs() {
	ptrAPI_ISteamHTTP_CreateHTTPRequest = func(uintptr, int32, string) HTTPRequestHandle { return 200 }
	ptrAPI_ISteamHTTP_SetHTTPRequestHeaderValue = func(uintptr, HTTPRequestHandle, string, string) bool { return true }
	ptrAPI_ISteamHTTP_SendHTTPRequest = func(self uintptr, request HTTPRequestHandle, callPtr uintptr) bool {
		*(*SteamAPICall_t)(unsafe.Pointer(callPtr)) = 33
		return true
	}
	ptrAPI_ISteamHTTP_GetHTTPResponseBodySize = func(self uintptr, request HTTPRequestHandle, sizePtr uintptr) bool {
		*(*uint32)(unsafe.Pointer(sizePtr)) = 3
		return true
	}
	ptrAPI_ISteamHTTP_GetHTTPResponseBodyData = func(self uintptr, request HTTPRequestHandle, bufPtr uintptr, bufLen uint32) bool {
		copy(unsafe.Slice((*byte)(unsafe.Pointer(bufPtr)), int(bufLen)), []byte("hey"))
		return true
	}
	ptrAPI_ISteamHTTP_ReleaseHTTPRequest = func(uintptr, HTTPRequestHandle) bool { return true }
}

func setSteamUGCStubs() {
	ptrAPI_ISteamUGC_GetNumSubscribedItems = func(uintptr, bool) uint32 { return 2 }
	ptrAPI_ISteamUGC_GetSubscribedItems = func(self uintptr, itemsPtr uintptr, count uint32, include bool) uint32 {
		items := unsafe.Slice((*PublishedFileId_t)(unsafe.Pointer(itemsPtr)), int(count))
		items[0] = 401
		items[1] = 402
		return 2
	}
}

func setSteamInventoryStubs() {
	ptrAPI_ISteamInventory_GetResultStatus = func(uintptr, SteamInventoryResult_t) int32 { return int32(EResultOK) }
	ptrAPI_ISteamInventory_GetResultItems = func(self uintptr, result SteamInventoryResult_t, itemsPtr, sizePtr uintptr) bool {
		*(*uint32)(unsafe.Pointer(sizePtr)) = 1
		items := unsafe.Slice((*SteamItemDetails)(unsafe.Pointer(itemsPtr)), 1)
		items[0] = SteamItemDetails{Definition: 501}
		return true
	}
	ptrAPI_ISteamInventory_DestroyResult = func(uintptr, SteamInventoryResult_t) {}
}

func setSteamInputStubs() {
	ptrAPI_ISteamInput_GetConnectedControllers = func(self uintptr, handlesPtr uintptr) int32 {
		handles := unsafe.Slice((*InputHandle_t)(unsafe.Pointer(handlesPtr)), _STEAM_INPUT_MAX_COUNT)
		handles[0] = 601
		handles[1] = 602
		return 2
	}
	ptrAPI_ISteamInput_GetInputTypeForHandle = func(uintptr, InputHandle_t) int32 { return int32(ESteamInputType_SteamController) }
	ptrAPI_ISteamInput_Init = func(uintptr, bool) bool { return true }
	ptrAPI_ISteamInput_Shutdown = func(uintptr) {}
	ptrAPI_ISteamInput_RunFrame = func(uintptr, bool) {}
	ptrAPI_ISteamInput_EnableDeviceCallbacks = func(uintptr) {}
	ptrAPI_ISteamInput_GetActionSetHandle = func(uintptr, string) InputActionSetHandle_t { return 701 }
	ptrAPI_ISteamInput_ActivateActionSet = func(uintptr, InputHandle_t, InputActionSetHandle_t) {}
	ptrAPI_ISteamInput_GetCurrentActionSet = func(uintptr, InputHandle_t) InputActionSetHandle_t { return 702 }
	ptrAPI_ISteamInput_ActivateActionSetLayer = func(uintptr, InputHandle_t, InputActionSetHandle_t) {}
	ptrAPI_ISteamInput_DeactivateActionSetLayer = func(uintptr, InputHandle_t, InputActionSetHandle_t) {}
	ptrAPI_ISteamInput_DeactivateAllActionSetLayers = func(uintptr, InputHandle_t) {}
	ptrAPI_ISteamInput_GetActiveActionSetLayers = func(self uintptr, inputHandle InputHandle_t, handlesPtr uintptr) int32 {
		handles := unsafe.Slice((*InputActionSetHandle_t)(unsafe.Pointer(handlesPtr)), 4)
		handles[0] = 701
		handles[1] = 703
		return 2
	}
	ptrAPI_ISteamInput_GetDigitalActionHandle = func(uintptr, string) InputDigitalActionHandle_t { return 801 }
	ptrAPI_ISteamInput_GetDigitalActionOrigins = func(self uintptr, inputHandle InputHandle_t, actionSetHandle InputActionSetHandle_t, actionHandle InputDigitalActionHandle_t, originsPtr uintptr) int32 {
		origins := unsafe.Slice((*EInputActionOrigin)(unsafe.Pointer(originsPtr)), 4)
		origins[0] = EInputActionOrigin_SteamController_A
		origins[1] = EInputActionOrigin_SteamController_B
		return 2
	}
	ptrAPI_ISteamInput_GetAnalogActionHandle = func(uintptr, string) InputAnalogActionHandle_t { return 802 }
	ptrAPI_ISteamInput_GetAnalogActionOrigins = func(self uintptr, inputHandle InputHandle_t, actionSetHandle InputActionSetHandle_t, actionHandle InputAnalogActionHandle_t, originsPtr uintptr) int32 {
		origins := unsafe.Slice((*EInputActionOrigin)(unsafe.Pointer(originsPtr)), 4)
		origins[0] = EInputActionOrigin_SteamController_X
		origins[1] = EInputActionOrigin_SteamController_Y
		return 2
	}
	ptrAPI_ISteamInput_StopAnalogActionMomentum = func(uintptr, InputHandle_t, InputAnalogActionHandle_t) {}
	ptrAPI_ISteamInput_TriggerVibration = func(uintptr, InputHandle_t, uint16, uint16) {}
	ptrAPI_ISteamInput_TriggerVibrationExtended = func(uintptr, InputHandle_t, uint16, uint16, uint16, uint16) {}
	ptrAPI_ISteamInput_TriggerSimpleHapticEvent = func(uintptr, InputHandle_t, ESteamControllerPad, uint16, uint16, uint16) {}
	ptrAPI_ISteamInput_SetLEDColor = func(uintptr, InputHandle_t, uint8, uint8, uint8, ESteamInputLEDFlag) {}
	ptrAPI_ISteamInput_ShowBindingPanel = func(uintptr, InputHandle_t) bool { return true }
	ptrAPI_ISteamInput_GetControllerForGamepadIndex = func(uintptr, int32) InputHandle_t { return 601 }
	ptrAPI_ISteamInput_GetGamepadIndexForController = func(uintptr, InputHandle_t) int32 { return 0 }
	ptrAPI_ISteamInput_GetStringForActionOrigin = func(uintptr, EInputActionOrigin) string { return "A" }
	ptrAPI_ISteamInput_GetGlyphForActionOrigin = func(uintptr, EInputActionOrigin) string { return "glyph" }
	ptrAPI_ISteamInput_GetRemotePlaySessionID = func(uintptr, InputHandle_t) uint32 { return 901 }
}

func setSteamRemoteStorageStubs() {
	ptrAPI_ISteamRemoteStorage_FileWrite = func(uintptr, string, uintptr, int32) bool { return true }
	ptrAPI_ISteamRemoteStorage_FileRead = func(self uintptr, name string, dataPtr uintptr, dataLen int32) int32 {
		copy(unsafe.Slice((*byte)(unsafe.Pointer(dataPtr)), int(dataLen)), []byte("hey"))
		return 3
	}
	ptrAPI_ISteamRemoteStorage_FileDelete = func(uintptr, string) bool { return true }
	ptrAPI_ISteamRemoteStorage_GetFileSize = func(uintptr, string) int32 { return 3 }
}

func setSteamUserStubs() {
	ptrAPI_ISteamUser_GetSteamID = func(uintptr) CSteamID { return 1001 }
}

func setSteamUserStatsStubs() {
	ptrAPI_ISteamUserStats_GetAchievement = func(self uintptr, name string, achievedPtr uintptr) bool {
		*(*bool)(unsafe.Pointer(achievedPtr)) = true
		return true
	}
	ptrAPI_ISteamUserStats_SetAchievement = func(uintptr, string) bool { return true }
	ptrAPI_ISteamUserStats_ClearAchievement = func(uintptr, string) bool { return true }
	ptrAPI_ISteamUserStats_StoreStats = func(uintptr) bool { return true }
}

func setSteamUtilsStubs() {
	ptrAPI_ISteamUtils_GetSecondsSinceAppActive = func(uintptr) uint32 { return 5 }
	ptrAPI_ISteamUtils_GetSecondsSinceComputerActive = func(uintptr) uint32 { return 6 }
	ptrAPI_ISteamUtils_GetConnectedUniverse = func(uintptr) int32 { return int32(EUniversePublic) }
	ptrAPI_ISteamUtils_GetServerRealTime = func(uintptr) uint32 { return 100 }
	ptrAPI_ISteamUtils_GetIPCountry = func(uintptr) string { return "US" }
	ptrAPI_ISteamUtils_GetImageSize = func(self uintptr, image int32, widthPtr, heightPtr uintptr) bool {
		*(*uint32)(unsafe.Pointer(widthPtr)) = 2
		*(*uint32)(unsafe.Pointer(heightPtr)) = 2
		return true
	}
	ptrAPI_ISteamUtils_GetImageRGBA = func(self uintptr, image int32, destPtr uintptr, destLen int32) bool {
		copy(unsafe.Slice((*byte)(unsafe.Pointer(destPtr)), int(destLen)), []byte("RGBA"))
		return true
	}
	ptrAPI_ISteamUtils_GetCurrentBatteryPower = func(uintptr) uint8 { return 88 }
	ptrAPI_ISteamUtils_GetAppID = func(uintptr) uint32 { return 480 }
	ptrAPI_ISteamUtils_SetOverlayNotificationPosition = func(uintptr, ENotificationPosition) {}
	ptrAPI_ISteamUtils_SetOverlayNotificationInset = func(uintptr, int32, int32) {}
	ptrAPI_ISteamUtils_IsAPICallCompleted = func(self uintptr, call SteamAPICall_t, failedPtr uintptr) bool {
		*(*bool)(unsafe.Pointer(failedPtr)) = false
		return true
	}
	ptrAPI_ISteamUtils_GetAPICallFailureReason = func(uintptr, SteamAPICall_t) int32 { return int32(ESteamAPICallFailureNone) }
	ptrAPI_ISteamUtils_GetAPICallResult = func(self uintptr, call SteamAPICall_t, callbackPtr uintptr, callbackSize int32, expectedCallback int32, failedPtr uintptr) bool {
		*(*bool)(unsafe.Pointer(failedPtr)) = false
		return true
	}
	ptrAPI_ISteamUtils_GetIPCCallCount = func(uintptr) uint32 { return 7 }
	ptrAPI_ISteamUtils_IsOverlayEnabled = func(uintptr) bool { return true }
	ptrAPI_ISteamUtils_BOverlayNeedsPresent = func(uintptr) bool { return false }
	ptrAPI_ISteamUtils_IsSteamRunningOnSteamDeck = func(uintptr) bool { return false }
	ptrAPI_ISteamUtils_ShowFloatingGamepadTextInput = func(uintptr, EFloatingGamepadTextInputMode, int32, int32, int32, int32) bool {
		return true
	}
}

func setSteamNetworkingUtilsStubs() {
	ptrAPI_ISteamNetworkingUtils_AllocateMessage = func(self uintptr, size int32) uintptr {
		msg := &SteamNetworkingMessage{MessageNumber: 99}
		return uintptr(unsafe.Pointer(msg))
	}
	ptrAPI_ISteamNetworkingUtils_InitRelayNetworkAccess = func(uintptr) {}
	ptrAPI_ISteamNetworkingUtils_GetLocalTimestamp = func(uintptr) SteamNetworkingMicroseconds { return 1000 }
}

func setSteamGameServerStubs() {
	ptrAPI_ISteamGameServer_SetProduct = func(uintptr, string) {}
	ptrAPI_ISteamGameServer_SetGameDescription = func(uintptr, string) {}
	ptrAPI_ISteamGameServer_LogOnAnonymous = func(uintptr) {}
	ptrAPI_ISteamGameServer_LogOff = func(uintptr) {}
	ptrAPI_ISteamGameServer_BLoggedOn = func(uintptr) bool { return true }
	ptrAPI_ISteamGameServer_GetSteamID = func(uintptr) CSteamID { return 2001 }
}

func setSteamNetworkingMessagesStubs() {
	ptrAPI_ISteamNetworkingMessages_SendMessageToUser = func(self uintptr, identityPtr, dataPtr uintptr, dataLen uint32, sendFlags int32, remoteChannel int32) EResult {
		return EResultOK
	}
	ptrAPI_ISteamNetworkingMessages_ReceiveMessagesOnChannel = func(self uintptr, channel int32, messagesPtr uintptr, maxMessages int32) int32 {
		slice := unsafe.Slice((**SteamNetworkingMessage)(unsafe.Pointer(messagesPtr)), int(maxMessages))
		slice[0] = &SteamNetworkingMessage{MessageNumber: 1}
		slice[1] = &SteamNetworkingMessage{MessageNumber: 2}
		return 2
	}
	ptrAPI_ISteamNetworkingMessages_AcceptSessionWithUser = func(uintptr, uintptr) bool { return true }
	ptrAPI_ISteamNetworkingMessages_CloseSessionWithUser = func(uintptr, uintptr) bool { return true }
	ptrAPI_ISteamNetworkingMessages_CloseChannelWithUser = func(uintptr, uintptr, int32) bool { return true }
}

func setSteamNetworkingSocketsStubs() {
	ptrAPI_ISteamNetworkingSockets_CreateListenSocketIP = func(uintptr, uintptr, int32, uintptr) HSteamListenSocket { return 3001 }
	ptrAPI_ISteamNetworkingSockets_CreateListenSocketP2P = func(uintptr, int32, int32, uintptr) HSteamListenSocket { return 3002 }
	ptrAPI_ISteamNetworkingSockets_ConnectByIPAddress = func(uintptr, uintptr, int32, uintptr) HSteamNetConnection { return 3003 }
	ptrAPI_ISteamNetworkingSockets_ConnectP2P = func(uintptr, uintptr, int32, int32, uintptr) HSteamNetConnection { return 3004 }
	ptrAPI_ISteamNetworkingSockets_AcceptConnection = func(uintptr, HSteamNetConnection) EResult { return EResultOK }
	ptrAPI_ISteamNetworkingSockets_CloseConnection = func(uintptr, HSteamNetConnection, int32, string, bool) bool { return true }
	ptrAPI_ISteamNetworkingSockets_CloseListenSocket = func(uintptr, HSteamListenSocket) bool { return true }
	ptrAPI_ISteamNetworkingSockets_SendMessageToConnection = func(self uintptr, connection HSteamNetConnection, dataPtr uintptr, dataLen uint32, sendFlags int32, messageNumberPtr uintptr) EResult {
		*(*int64)(unsafe.Pointer(messageNumberPtr)) = 123
		return EResultOK
	}
	ptrAPI_ISteamNetworkingSockets_ReceiveMessagesOnConnection = func(self uintptr, connection HSteamNetConnection, messagesPtr uintptr, maxMessages int32) int32 {
		slice := unsafe.Slice((**SteamNetworkingMessage)(unsafe.Pointer(messagesPtr)), int(maxMessages))
		slice[0] = &SteamNetworkingMessage{MessageNumber: 3}
		slice[1] = &SteamNetworkingMessage{MessageNumber: 4}
		return 2
	}
	ptrAPI_ISteamNetworkingSockets_CreatePollGroup = func(uintptr) HSteamNetPollGroup { return 5001 }
	ptrAPI_ISteamNetworkingSockets_DestroyPollGroup = func(uintptr, HSteamNetPollGroup) bool { return true }
	ptrAPI_ISteamNetworkingSockets_SetConnectionPollGroup = func(uintptr, HSteamNetConnection, HSteamNetPollGroup) bool { return true }
	ptrAPI_ISteamNetworkingSockets_ReceiveMessagesOnPollGroup = func(self uintptr, group HSteamNetPollGroup, messagesPtr uintptr, maxMessages int32) int32 {
		slice := unsafe.Slice((**SteamNetworkingMessage)(unsafe.Pointer(messagesPtr)), int(maxMessages))
		slice[0] = &SteamNetworkingMessage{MessageNumber: 5}
		slice[1] = &SteamNetworkingMessage{MessageNumber: 6}
		return 2
	}
}

func writeCString(ptr uintptr, max int32, value string) int32 {
	buf := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), int(max))
	data := append([]byte(value), 0)
	return int32(copy(buf, data))
}

func writeString(ptr uintptr, max int32, value string) int32 {
	buf := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), int(max))
	return int32(copy(buf, []byte(value)))
}
