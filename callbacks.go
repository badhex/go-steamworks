// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The go-steamworks Authors

package steamworks

import (
	"sync"
	"unsafe"
)

// CallbackID represents a Steam callback identifier.
type CallbackID int32

type callbackHandler struct {
	size uintptr
	fn   func(unsafe.Pointer)
}

// CallbackDispatcher stores typed callback handlers keyed by callback ID.
// It is intended for use with manual dispatch flows.
type CallbackDispatcher struct {
	mu       sync.RWMutex
	handlers map[CallbackID]callbackHandler
}

// NewCallbackDispatcher constructs a new dispatcher.
func NewCallbackDispatcher() *CallbackDispatcher {
	return &CallbackDispatcher{
		handlers: make(map[CallbackID]callbackHandler),
	}
}

// RegisterCallback registers a typed handler for a callback ID.
func RegisterCallback[T any](d *CallbackDispatcher, id CallbackID, handler func(T)) {
	var zero T
	d.mu.Lock()
	d.handlers[id] = callbackHandler{
		size: unsafe.Sizeof(zero),
		fn: func(ptr unsafe.Pointer) {
			handler(*(*T)(ptr))
		},
	}
	d.mu.Unlock()
}

// Dispatch invokes the registered handler for the callback ID, if any.
func (d *CallbackDispatcher) Dispatch(id CallbackID, data unsafe.Pointer) bool {
	d.mu.RLock()
	handler, ok := d.handlers[id]
	d.mu.RUnlock()
	if !ok {
		return false
	}
	handler.fn(data)
	return true
}

// ExpectedSize returns the expected payload size for the callback ID, if known.
func (d *CallbackDispatcher) ExpectedSize(id CallbackID) (uintptr, bool) {
	d.mu.RLock()
	handler, ok := d.handlers[id]
	d.mu.RUnlock()
	if !ok {
		return 0, false
	}
	return handler.size, true
}
