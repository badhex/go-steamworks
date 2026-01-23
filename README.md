# go-steamworks

A Steamworks SDK binding for Go

> [!WARNING]
> 32bit OSes are not supported.

## Steamworks SDK version

163

## How to use

Before using this library, make sure Steam's redistributable binaries are
available on your runtime machine. This repository no longer ships the
precompiled Steamworks shared libraries; provide them alongside your
application at runtime (for example, next to your executable).

Common locations and filenames:

* Linux (64-bit): `libsteam_api.so`
* macOS: `libsteam_api.dylib`
* Windows (64-bit): `steam_api64.dll`

On Windows, copy one of these files on the working directory:

 * `steam_api64.dll` (For 64bit. Copy `redistribution_bin\win64\steam_api64.dll` in the SDK)

```go
package steamapi

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/go-steamworks"
	"golang.org/x/text/language"
)

const appID = 480 // Rewrite this

func init() {
	if steamworks.RestartAppIfNecessary(appID) {
		os.Exit(1)
	}
	if err := steamworks.Init(); err != nil {
		panic(fmt.Sprintf("steamworks.Init failed: %v", err))
	}
}

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

### Supported APIs and methods

This binding exposes a subset of the Steamworks SDK via Go interfaces. The
implemented methods include:

**General**

* `RestartAppIfNecessary(appID uint32) bool`
* `Init() error`
* `RunCallbacks()`

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
* `RunFrame()`

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

**ISteamNetworkingMessages** (`steamworks.SteamNetworkingMessages()`)

* `SendMessageToUser(identity *SteamNetworkingIdentity, data []byte, sendFlags SteamNetworkingSendFlags, remoteChannel int) EResult`
* `ReceiveMessagesOnChannel(channel int, maxMessages int) []*SteamNetworkingMessage`
* `AcceptSessionWithUser(identity *SteamNetworkingIdentity) bool`
* `CloseSessionWithUser(identity *SteamNetworkingIdentity) bool`
* `CloseChannelWithUser(identity *SteamNetworkingIdentity, channel int) bool`

**ISteamNetworkingSockets** (`steamworks.SteamNetworkingSockets()`)

* `CreateListenSocketP2P(localVirtualPort int, options []SteamNetworkingConfigValue) HSteamListenSocket`
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

This repository no longer ships Steamworks redistribution binaries. You must
obtain the appropriate shared libraries from the Steamworks SDK and distribute
them with your application in accordance with the
[Valve Corporation Steamworks SDK Access Agreement](https://partner.steamgames.com/documentation/sdk_access_agreement).

## Resources

 * [Steamworks SDK](https://partner.steamgames.com/doc/sdk)
