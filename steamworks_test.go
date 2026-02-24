// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The go-steamworks Authors

package steamworks

import (
	"bytes"
	"testing"
	"unsafe"
)

func TestCStringToGo(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		want string
	}{
		{name: "empty", in: []byte{0}, want: ""},
		{name: "no-null", in: []byte("steam"), want: "steam"},
		{name: "with-null", in: []byte{'s', 't', 'e', 'a', 'm', 0, 'x'}, want: "steam"},
	}
	for _, tt := range tests {
		if got := cStringToGo(tt.in); got != tt.want {
			t.Fatalf("%s: cStringToGo(%v)=%q, want %q", tt.name, tt.in, got, tt.want)
		}
	}
}

func TestPutUint32(t *testing.T) {
	var buf [4]byte
	putUint32(buf[:], 0x11223344)
	want := []byte{0x44, 0x33, 0x22, 0x11}
	if !bytes.Equal(buf[:], want) {
		t.Fatalf("putUint32 wrote %v, want %v", buf, want)
	}
}

func TestPutUint64(t *testing.T) {
	var buf [8]byte
	putUint64(buf[:], 0x1122334455667788)
	want := []byte{0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11}
	if !bytes.Equal(buf[:], want) {
		t.Fatalf("putUint64 wrote %v, want %v", buf, want)
	}
}

func TestSteamNetworkingIdentitySetSteamID64(t *testing.T) {
	var id SteamNetworkingIdentity
	id.SetSteamID64(0x0102030405060708)
	wantType := []byte{0x10, 0x00, 0x00, 0x00}
	wantSize := []byte{0x08, 0x00, 0x00, 0x00}
	if !bytes.Equal(id.data[0:4], wantType) {
		t.Fatalf("identity type = %v, want %v", id.data[0:4], wantType)
	}
	if !bytes.Equal(id.data[4:8], wantSize) {
		t.Fatalf("identity size = %v, want %v", id.data[4:8], wantSize)
	}
	wantID := []byte{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01}
	if !bytes.Equal(id.data[8:16], wantID) {
		t.Fatalf("identity steamID bytes = %v, want %v", id.data[8:16], wantID)
	}
}

func TestSteamNetworkingIdentitySetIPv4Addr(t *testing.T) {
	var id SteamNetworkingIdentity
	id.SetIPv4Addr(0x7f000001, 27015)
	wantType := []byte{0x01, 0x00, 0x00, 0x00}
	wantSize := []byte{0x12, 0x00, 0x00, 0x00}
	if !bytes.Equal(id.data[0:4], wantType) {
		t.Fatalf("identity type = %v, want %v", id.data[0:4], wantType)
	}
	if !bytes.Equal(id.data[4:8], wantSize) {
		t.Fatalf("identity size = %v, want %v", id.data[4:8], wantSize)
	}
	wantAddr := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 0x7f, 0x00, 0x00, 0x01, 0x87, 0x69}
	if !bytes.Equal(id.data[8:26], wantAddr) {
		t.Fatalf("identity addr = %v, want %v", id.data[8:26], wantAddr)
	}
}

func TestSteamNetworkingIPAddrSetIPv4(t *testing.T) {
	var addr SteamNetworkingIPAddr
	addr.SetIPv4(0x7f000001, 27015)
	want := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 0x7f, 0x00, 0x00, 0x01, 0x87, 0x69}
	if !bytes.Equal(addr.data[:], want) {
		t.Fatalf("addr data = %v, want %v", addr.data, want)
	}
}

func TestSteamNetworkingIPAddrSetIPv6(t *testing.T) {
	var addr SteamNetworkingIPAddr
	ip := [16]byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	addr.SetIPv6(ip, 8080)
	want := append(ip[:], 0x90, 0x1f)
	if !bytes.Equal(addr.data[:], want) {
		t.Fatalf("addr data = %v, want %v", addr.data, want)
	}
}

func TestOptionsPtr(t *testing.T) {
	if ptr := optionsPtr(nil); ptr != 0 {
		t.Fatalf("optionsPtr(nil)=%d, want 0", ptr)
	}
	values := []SteamNetworkingConfigValue{{Value: 1}}
	if ptr := optionsPtr(values); ptr == 0 {
		t.Fatalf("optionsPtr(values)=0, want non-zero")
	}
}

func TestConnectedControllersIterator(t *testing.T) {
	// This is a smoke test to ensure the iterator is correctly defined.
	// Since we don't have a real Steam environment, we can't test actual functionality.
	var s steamInput
	_ = s.ConnectedControllers()
}

func TestFriendsIterator(t *testing.T) {
	var s steamFriends
	_ = s.Friends(EFriendFlagImmediate)
}

func TestLobbyMembersIterator(t *testing.T) {
	var s steamMatchmaking
	var lobbyID CSteamID
	_ = s.LobbyMembers(lobbyID)
}

func TestLobbyCallbackPayloadSizes(t *testing.T) {
	var (
		dataUpdate LobbyDataUpdate
		chatUpdate LobbyChatUpdate
		chatMsg    LobbyChatMsg
	)

	if got, want := unsafe.Sizeof(dataUpdate), uintptr(24); got != want {
		t.Fatalf("LobbyDataUpdate size=%d, want %d", got, want)
	}
	if got, want := unsafe.Sizeof(chatUpdate), uintptr(32); got != want {
		t.Fatalf("LobbyChatUpdate size=%d, want %d", got, want)
	}
	if got, want := unsafe.Sizeof(chatMsg), uintptr(24); got != want {
		t.Fatalf("LobbyChatMsg size=%d, want %d", got, want)
	}
}

func TestLobbyChatMsgLayout(t *testing.T) {
	var msg LobbyChatMsg

	if got, want := unsafe.Offsetof(msg.ChatEntryType), uintptr(16); got != want {
		t.Fatalf("LobbyChatMsg.ChatEntryType offset=%d, want %d", got, want)
	}
	if got, want := unsafe.Offsetof(msg.ChatID), uintptr(20); got != want {
		t.Fatalf("LobbyChatMsg.ChatID offset=%d, want %d", got, want)
	}

	msg.ChatEntryType = uint8(EChatEntryTypeWasKicked)
	if got, want := msg.EntryType(), EChatEntryTypeWasKicked; got != want {
		t.Fatalf("LobbyChatMsg.EntryType()=%v, want %v", got, want)
	}
}
