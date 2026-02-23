// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The go-steamworks Authors

package steamworks

import (
	"unsafe"

	"github.com/ebitengine/purego"
)

// LookupSymbol returns the raw address of a Steamworks SDK symbol for advanced usage.
// Callers are responsible for providing the correct argument and return types.
func LookupSymbol(name string) (uintptr, error) {
	l, err := ensureLoaded()
	if err != nil {
		return 0, err
	}
	theLib = l
	return lookupSymbolAddr(theLib.lib, name)
}

// CallSymbol looks up a Steamworks SDK symbol and invokes it with the provided arguments.
// Callers are responsible for providing the correct argument and return types.
func CallSymbol(name string, args ...uintptr) (uintptr, error) {
	ptr, err := LookupSymbol(name)
	if err != nil {
		return 0, err
	}
	return CallSymbolPtr(ptr, args...), nil
}

// CallSymbolPtr invokes a Steamworks SDK function pointer with the provided arguments.
// Callers are responsible for providing the correct argument and return types.
func CallSymbolPtr(ptr uintptr, args ...uintptr) uintptr {
	r1, _, _ := purego.SyscallN(ptr, args...)
	return r1
}

// SteamEncryptedAppTicketBDecryptTicket decrypts an encrypted app ticket.
func SteamEncryptedAppTicketBDecryptTicket(ticket []byte, decrypted []byte, key []byte) (decryptedSize uint32, ok bool) {
	if len(ticket) == 0 || len(decrypted) == 0 || len(key) == 0 {
		return 0, false
	}
	ptr, err := LookupSymbol("SteamEncryptedAppTicket_BDecryptTicket")
	if err != nil {
		return 0, false
	}
	decryptedSize = uint32(len(decrypted))
	r1, _, _ := purego.SyscallN(
		ptr,
		uintptr(unsafe.Pointer(&ticket[0])),
		uintptr(uint32(len(ticket))),
		uintptr(unsafe.Pointer(&decrypted[0])),
		uintptr(unsafe.Pointer(&decryptedSize)),
		uintptr(unsafe.Pointer(&key[0])),
		uintptr(uint32(len(key))),
	)
	return decryptedSize, r1 != 0
}

// SteamEncryptedAppTicketBIsTicketForApp checks whether a decrypted ticket matches an app ID.
func SteamEncryptedAppTicketBIsTicketForApp(decryptedTicket []byte, appID AppId_t) bool {
	if len(decryptedTicket) == 0 {
		return false
	}
	ptr, err := LookupSymbol("SteamEncryptedAppTicket_BIsTicketForApp")
	if err != nil {
		return false
	}
	r1, _, _ := purego.SyscallN(
		ptr,
		uintptr(unsafe.Pointer(&decryptedTicket[0])),
		uintptr(uint32(len(decryptedTicket))),
		uintptr(appID),
	)
	return r1 != 0
}

// SteamEncryptedAppTicketGetTicketIssueTime returns the issue timestamp from a decrypted ticket.
func SteamEncryptedAppTicketGetTicketIssueTime(decryptedTicket []byte) uint32 {
	if len(decryptedTicket) == 0 {
		return 0
	}
	ptr, err := LookupSymbol("SteamEncryptedAppTicket_GetTicketIssueTime")
	if err != nil {
		return 0
	}
	r1, _, _ := purego.SyscallN(
		ptr,
		uintptr(unsafe.Pointer(&decryptedTicket[0])),
		uintptr(uint32(len(decryptedTicket))),
	)
	return uint32(r1)
}

// SteamEncryptedAppTicketGetTicketSteamID returns the ticket owner SteamID from a decrypted ticket.
func SteamEncryptedAppTicketGetTicketSteamID(decryptedTicket []byte) (CSteamID, bool) {
	if len(decryptedTicket) == 0 {
		return 0, false
	}
	ptr, err := LookupSymbol("SteamEncryptedAppTicket_GetTicketSteamID")
	if err != nil {
		return 0, false
	}
	var id CSteamID
	r1, _, _ := purego.SyscallN(
		ptr,
		uintptr(unsafe.Pointer(&decryptedTicket[0])),
		uintptr(uint32(len(decryptedTicket))),
		uintptr(unsafe.Pointer(&id)),
	)
	return id, r1 != 0
}
