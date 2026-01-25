# go-steamworks

Go bindings for a subset of the Steamworks SDK.

> [!WARNING]
> 32-bit OSes are not supported.

## Steamworks SDK version

163

> [!NOTE]
> If newer Steamworks SDK releases add or update symbols that are not yet in these bindings, use the [raw symbol access](#raw-symbol-access) method to call them directly.

## Getting started

### Requirements

Before using this library, make sure Steam's redistributable binaries are
available on your runtime machine. This repository no longer ships the
precompiled Steamworks shared libraries; provide them alongside your
application at runtime (for example, next to your executable).

Common locations and filenames:

* Linux (64-bit): `libsteam_api.so`
* macOS: `libsteam_api.dylib`
* Windows (64-bit): `steam_api64.dll`

On Windows, copy the DLL into the working directory:

* `steam_api64.dll` (copy from `redistribution_bin\\win64\\steam_api64.dll` in the SDK)

For local development, ensure `steam_appid.txt` is available next to the
executable (or run Steam with your app ID configured).

### Initialization

The Steamworks client must be running and the API must be initialized before
calling most interfaces. `Load` is optional, but allows you to surface missing
redistributables early.

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/badhex/go-steamworks"
)

const appID = 480 // Replace with your own App ID.

func main() {
	if steamworks.RestartAppIfNecessary(appID) {
		os.Exit(0)
	}
	if err := steamworks.Load(); err != nil {
		log.Fatalf("failed to load steamworks: %v", err)
	}
	if err := steamworks.Init(); err != nil {
		log.Fatalf("steamworks.Init failed: %v", err)
	}

	fmt.Printf("SteamID: %v\n", steamworks.SteamUser().GetSteamID())
}
```

### Callback pump

Steamworks expects you to poll callbacks regularly on your main thread.

```go
for running {
	steamworks.RunCallbacks()
	// ...your game loop...
}
```

### Example: language selection

```go
package steamapi

import (
	"github.com/badhex/go-steamworks"
	"golang.org/x/text/language"
)

func SystemLang() language.Tag {
	switch steamworks.SteamApps().GetCurrentGameLanguage() {
	case "english":
		return language.English
	case "japanese":
		return language.Japanese
	}
	return language.Und
}
```

### Example: achievements

```go
if achieved, ok := steamworks.SteamUserStats().GetAchievement("FIRST_WIN"); ok && !achieved {
	steamworks.SteamUserStats().SetAchievement("FIRST_WIN")
	steamworks.SteamUserStats().StoreStats()
}
```

### Example: async call results

```go
call := steamworks.SteamHTTP().CreateHTTPRequest(steamworks.EHTTPMethodGET, "https://example.com")
callHandle, ok := steamworks.SteamHTTP().SendHTTPRequest(call)
if !ok {
	// handle request creation failure
}

type HTTPRequestCompleted struct {
	Request steamworks.HTTPRequestHandle
	Context uint64
	Status  int32
}

// Define the struct to mirror the Steamworks callback payload you expect.
// Use the SDK's callback ID for the expected payload.
result := steamworks.NewCallResult[HTTPRequestCompleted](callHandle, 2101)

if _, failed, err := result.Wait(context.Background(), 0); err == nil && !failed {
	// process response
}
```

## SDK-aligned helpers

This repository ships typed helpers for async call results and manual callback
dispatch, plus additional interface accessors to align with common Steamworks
flows.

* Use `NewCallResult` to await async call results with typed payloads.
* Use `NewCallbackDispatcher` + `RegisterCallback` for manual callback registration and dispatch.
* Use versioned accessors such as `SteamAppsV008()` when you need explicit
  interface versions.

## Build tags and runtime loading

By default, the package expects Steam redistributables to be available on the
runtime library path. You can also opt into embedding redistributables with a
build tag:

* Runtime loading (default): rely on `libsteam_api.so` / `libsteam_api.dylib`
  being in the dynamic linker path or alongside your executable.
* Embedded loading: build with `-tags steamworks_embedded` to embed the SDK
  redistributables and load them from a temporary file at runtime.

Use `STEAMWORKS_LIB_PATH` to point at a custom shared library location when
runtime loading.

## Repository layout

* `gen.go` — code generator for parsing the SDK and building bindings.
* `examples/` — runnable samples for common startup flows.

### Supported APIs and methods

This binding exposes a subset of the Steamworks SDK via Go interfaces. The
implemented methods include:

**General**

* `RestartAppIfNecessary(appID uint32) bool`
* `Init() error`
* `RunCallbacks()`
* `Shutdown()`
* `IsSteamRunning() bool`
* `GetSteamInstallPath() string`
* `ReleaseCurrentThreadMemory()`

**ISteamApps** (`steamworks.SteamApps()`)

* `BGetDLCDataByIndex(iDLC int) (appID AppId_t, available bool, name string, success bool)`
* `BIsSubscribed() bool`
* `BIsLowViolence() bool`
* `BIsCybercafe() bool`
* `BIsVACBanned() bool`
* `BIsDlcInstalled(appID AppId_t) bool`
* `BIsSubscribedApp(appID AppId_t) bool`
* `BIsSubscribedFromFreeWeekend() bool`
* `BIsSubscribedFromFamilySharing() bool`
* `BIsTimedTrial() (allowedSeconds, playedSeconds uint32, ok bool)`
* `BIsAppInstalled(appID AppId_t) bool`
* `GetAvailableGameLanguages() string`
* `GetEarliestPurchaseUnixTime(appID AppId_t) uint32`
* `GetAppInstallDir(appID AppId_t) string`
* `GetCurrentGameLanguage() string`
* `GetDLCCount() int32`
* `GetCurrentBetaName() (string, bool)`
* `GetInstalledDepots(appID AppId_t) []DepotId_t`
* `GetAppOwner() CSteamID`
* `GetLaunchQueryParam(key string) string`
* `GetDlcDownloadProgress(appID AppId_t) (downloaded, total uint64, ok bool)`
* `GetAppBuildId() int32`
* `GetFileDetails(filename string) SteamAPICall_t`
* `GetLaunchCommandLine(bufferSize int) string`
* `GetNumBetas() (total int, available int, private int)`
* `GetBetaInfo(index int) (flags uint32, buildID uint32, name string, description string, ok bool)`
* `InstallDLC(appID AppId_t)`
* `UninstallDLC(appID AppId_t)`
* `RequestAppProofOfPurchaseKey(appID AppId_t)`
* `RequestAllProofOfPurchaseKeys()`
* `MarkContentCorrupt(missingFilesOnly bool) bool`
* `SetDlcContext(appID AppId_t) bool`
* `SetActiveBeta(name string) bool`

**ISteamFriends** (`steamworks.SteamFriends()`)

* `GetPersonaName() string`
* `GetPersonaState() EPersonaState`
* `GetFriendCount(flags EFriendFlags) int`
* `GetFriendByIndex(index int, flags EFriendFlags) CSteamID`
* `GetFriendRelationship(friend CSteamID) EFriendRelationship`
* `GetFriendPersonaState(friend CSteamID) EPersonaState`
* `GetFriendPersonaName(friend CSteamID) string`
* `GetFriendPersonaNameHistory(friend CSteamID, index int) string`
* `GetFriendSteamLevel(friend CSteamID) int`
* `GetSmallFriendAvatar(friend CSteamID) int32`
* `GetMediumFriendAvatar(friend CSteamID) int32`
* `GetLargeFriendAvatar(friend CSteamID) int32`
* `SetRichPresence(key, value string) bool`
* `GetFriendGamePlayed(friend CSteamID) (FriendGameInfo, bool)`
* `InviteUserToGame(friend CSteamID, connectString string) bool`
* `ActivateGameOverlay(dialog string)`
* `ActivateGameOverlayToUser(dialog string, steamID CSteamID)`
* `ActivateGameOverlayToWebPage(url string, mode EActivateGameOverlayToWebPageMode)`
* `ActivateGameOverlayToStore(appID AppId_t, flag EOverlayToStoreFlag)`
* `ActivateGameOverlayInviteDialog(lobbyID CSteamID)`
* `ActivateGameOverlayInviteDialogConnectString(connectString string)`

**ISteamInput** (`steamworks.SteamInput()`)

* `GetConnectedControllers() []InputHandle_t`
* `GetInputTypeForHandle(inputHandle InputHandle_t) ESteamInputType`
* `Init(bExplicitlyCallRunFrame bool) bool`
* `Shutdown()`
* `RunFrame()`
* `EnableDeviceCallbacks()`
* `DisableDeviceCallbacks()`
* `GetActionSetHandle(actionSetName string) InputActionSetHandle_t`
* `ActivateActionSet(inputHandle InputHandle_t, actionSetHandle InputActionSetHandle_t)`
* `GetCurrentActionSet(inputHandle InputHandle_t) InputActionSetHandle_t`
* `ActivateActionSetLayer(inputHandle InputHandle_t, actionSetHandle InputActionSetHandle_t)`
* `DeactivateActionSetLayer(inputHandle InputHandle_t, actionSetHandle InputActionSetHandle_t)`
* `DeactivateAllActionSetLayers(inputHandle InputHandle_t)`
* `GetActiveActionSetLayers(inputHandle InputHandle_t, handles []InputActionSetHandle_t) int`
* `GetDigitalActionHandle(actionName string) InputDigitalActionHandle_t`
* `GetDigitalActionData(inputHandle InputHandle_t, actionHandle InputDigitalActionHandle_t) InputDigitalActionData`
* `GetDigitalActionOrigins(inputHandle InputHandle_t, actionSetHandle InputActionSetHandle_t, actionHandle InputDigitalActionHandle_t, origins []EInputActionOrigin) int`
* `GetAnalogActionHandle(actionName string) InputAnalogActionHandle_t`
* `GetAnalogActionData(inputHandle InputHandle_t, actionHandle InputAnalogActionHandle_t) InputAnalogActionData`
* `GetAnalogActionOrigins(inputHandle InputHandle_t, actionSetHandle InputActionSetHandle_t, actionHandle InputAnalogActionHandle_t, origins []EInputActionOrigin) int`
* `StopAnalogActionMomentum(inputHandle InputHandle_t, actionHandle InputAnalogActionHandle_t)`
* `GetMotionData(inputHandle InputHandle_t) InputMotionData`
* `TriggerVibration(inputHandle InputHandle_t, leftSpeed, rightSpeed uint16)`
* `TriggerVibrationExtended(inputHandle InputHandle_t, leftSpeed, rightSpeed, leftTriggerSpeed, rightTriggerSpeed uint16)`
* `TriggerSimpleHapticEvent(inputHandle InputHandle_t, pad ESteamControllerPad, durationMicroSec, offMicroSec, repeat uint16)`
* `SetLEDColor(inputHandle InputHandle_t, red, green, blue uint8, flags ESteamInputLEDFlag)`
* `ShowBindingPanel(inputHandle InputHandle_t) bool`
* `GetControllerForGamepadIndex(index int) InputHandle_t`
* `GetGamepadIndexForController(inputHandle InputHandle_t) int`
* `GetStringForActionOrigin(origin EInputActionOrigin) string`
* `GetGlyphForActionOrigin(origin EInputActionOrigin) string`
* `GetRemotePlaySessionID(inputHandle InputHandle_t) uint32`

**ISteamMatchmaking** (`steamworks.SteamMatchmaking()`)

* `RequestLobbyList() SteamAPICall_t`
* `GetLobbyByIndex(index int) CSteamID`
* `CreateLobby(lobbyType ELobbyType, maxMembers int) SteamAPICall_t`
* `JoinLobby(lobbyID CSteamID) SteamAPICall_t`
* `LeaveLobby(lobbyID CSteamID)`
* `InviteUserToLobby(lobbyID, invitee CSteamID) bool`
* `GetNumLobbyMembers(lobbyID CSteamID) int`
* `GetLobbyMemberByIndex(lobbyID CSteamID, memberIndex int) CSteamID`
* `GetLobbyData(lobbyID CSteamID, key string) string`
* `SetLobbyData(lobbyID CSteamID, key, value string) bool`
* `GetLobbyOwner(lobbyID CSteamID) CSteamID`
* `SetLobbyOwner(lobbyID, owner CSteamID) bool`
* `SetLobbyGameServer(lobbyID CSteamID, ip uint32, port uint16, server CSteamID)`
* `GetLobbyGameServer(lobbyID CSteamID) (ip uint32, port uint16, server CSteamID, ok bool)`
* `SetLobbyJoinable(lobbyID CSteamID, joinable bool) bool`
* `SetLobbyMemberLimit(lobbyID CSteamID, maxMembers int) bool`
* `SetLobbyType(lobbyID CSteamID, lobbyType ELobbyType) bool`

**ISteamHTTP** (`steamworks.SteamHTTP()`)

* `CreateHTTPRequest(method EHTTPMethod, absoluteURL string) HTTPRequestHandle`
* `SetHTTPRequestHeaderValue(request HTTPRequestHandle, headerName, headerValue string) bool`
* `SendHTTPRequest(request HTTPRequestHandle) (SteamAPICall_t, bool)`
* `GetHTTPResponseBodySize(request HTTPRequestHandle) (uint32, bool)`
* `GetHTTPResponseBodyData(request HTTPRequestHandle, buffer []byte) bool`
* `ReleaseHTTPRequest(request HTTPRequestHandle) bool`

**ISteamUGC** (`steamworks.SteamUGC()`)

* `GetNumSubscribedItems(includeLocallyDisabled bool) uint32`
* `GetSubscribedItems(includeLocallyDisabled bool) []PublishedFileId_t`

**ISteamInventory** (`steamworks.SteamInventory()`)

* `GetResultStatus(result SteamInventoryResult_t) EResult`
* `GetResultItems(result SteamInventoryResult_t, outItems []SteamItemDetails) (int, bool)`
* `DestroyResult(result SteamInventoryResult_t)`

**ISteamNetworkingMessages** (`steamworks.SteamNetworkingMessages()`)

* `SendMessageToUser(identity *SteamNetworkingIdentity, data []byte, sendFlags SteamNetworkingSendFlags, remoteChannel int) EResult`
* `ReceiveMessagesOnChannel(channel int, maxMessages int) []*SteamNetworkingMessage`
* `AcceptSessionWithUser(identity *SteamNetworkingIdentity) bool`
* `CloseSessionWithUser(identity *SteamNetworkingIdentity) bool`
* `CloseChannelWithUser(identity *SteamNetworkingIdentity, channel int) bool`

**ISteamNetworkingUtils** (`steamworks.SteamNetworkingUtils()`)

* `AllocateMessage(size int) *SteamNetworkingMessage`
* `InitRelayNetworkAccess()`
* `GetLocalTimestamp() SteamNetworkingMicroseconds`

**ISteamNetworkingSockets** (`steamworks.SteamNetworkingSockets()`)

* `CreateListenSocketIP(localAddress *SteamNetworkingIPAddr, options []SteamNetworkingConfigValue) HSteamListenSocket`
* `CreateListenSocketP2P(localVirtualPort int, options []SteamNetworkingConfigValue) HSteamListenSocket`
* `ConnectByIPAddress(address *SteamNetworkingIPAddr, options []SteamNetworkingConfigValue) HSteamNetConnection`
* `ConnectP2P(identity *SteamNetworkingIdentity, remoteVirtualPort int, options []SteamNetworkingConfigValue) HSteamNetConnection`
* `AcceptConnection(connection HSteamNetConnection) EResult`
* `CloseConnection(connection HSteamNetConnection, reason int, debug string, enableLinger bool) bool`
* `CloseListenSocket(socket HSteamListenSocket) bool`
* `SendMessageToConnection(connection HSteamNetConnection, data []byte, sendFlags SteamNetworkingSendFlags) (EResult, int64)`
* `ReceiveMessagesOnConnection(connection HSteamNetConnection, maxMessages int) []*SteamNetworkingMessage`
* `CreatePollGroup() HSteamNetPollGroup`
* `DestroyPollGroup(group HSteamNetPollGroup) bool`
* `SetConnectionPollGroup(connection HSteamNetConnection, group HSteamNetPollGroup) bool`
* `ReceiveMessagesOnPollGroup(group HSteamNetPollGroup, maxMessages int) []*SteamNetworkingMessage`

**ISteamGameServer** (`steamworks.SteamGameServer()`)

* `SetProduct(product string)`
* `SetGameDescription(description string)`
* `LogOnAnonymous()`
* `LogOff()`
* `BLoggedOn() bool`
* `GetSteamID() CSteamID`

**ISteamRemoteStorage** (`steamworks.SteamRemoteStorage()`)

* `FileWrite(file string, data []byte) bool`
* `FileRead(file string, data []byte) int32`
* `FileDelete(file string) bool`
* `GetFileSize(file string) int32`

**ISteamUser** (`steamworks.SteamUser()`)

* `GetSteamID() CSteamID`

**ISteamUserStats** (`steamworks.SteamUserStats()`)

* `GetAchievement(name string) (achieved, success bool)`
* `SetAchievement(name string) bool`
* `ClearAchievement(name string) bool`
* `StoreStats() bool`

**ISteamUtils** (`steamworks.SteamUtils()`)

* `GetSecondsSinceAppActive() uint32`
* `GetSecondsSinceComputerActive() uint32`
* `GetConnectedUniverse() EUniverse`
* `GetServerRealTime() uint32`
* `GetIPCountry() string`
* `GetImageSize(image int) (width, height uint32, ok bool)`
* `GetImageRGBA(image int, dest []byte) bool`
* `GetCurrentBatteryPower() uint8`
* `GetAppID() uint32`
* `IsOverlayEnabled() bool`
* `BOverlayNeedsPresent() bool`
* `IsSteamRunningOnSteamDeck() bool`
* `SetOverlayNotificationPosition(position ENotificationPosition)`
* `SetOverlayNotificationInset(horizontal, vertical int32)`
* `IsAPICallCompleted(call SteamAPICall_t) (failed bool, ok bool)`
* `GetAPICallFailureReason(call SteamAPICall_t) ESteamAPICallFailure`
* `GetAPICallResult(call SteamAPICall_t, callback uintptr, callbackSize int32, expectedCallback int32) (failed bool, ok bool)`
* `GetIPCCallCount() uint32`
* `ShowFloatingGamepadTextInput(...) bool`

### Raw symbol access

To access newer or unsupported Steamworks SDK methods, you can call raw symbols
directly:

```go
// Look up a symbol and call it directly (advanced usage).
ptr, err := steamworks.LookupSymbol("SteamAPI_ISteamFriends_GetPersonaName")
if err != nil {
	panic(err)
}
result := steamworks.CallSymbolPtr(ptr)
_ = result
```

Or use `CallSymbol` to combine lookup + call:

```go
result, err := steamworks.CallSymbol("SteamAPI_SteamApps_v008")
if err != nil {
	panic(err)
}
_ = result
```

## License

All the source code files are licensed under Apache License 2.0.

## Resources

 * [Steamworks SDK](https://partner.steamgames.com/doc/sdk)
