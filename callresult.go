// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The go-steamworks Authors

package steamworks

import (
	"context"
	"errors"
	"math"
	"time"
	"unsafe"
)

var (
	ErrAPICallNotReady = errors.New("steamworks: api call result not ready")
	ErrAPICallTooLarge = errors.New("steamworks: api call result size exceeds int32")
)

// CallResult wraps a SteamAPICall_t with typed result helpers.
type CallResult[T any] struct {
	call             SteamAPICall_t
	expectedCallback int32
}

// NewCallResult constructs a typed call result helper for a SteamAPICall_t.
func NewCallResult[T any](call SteamAPICall_t, expectedCallback int32) *CallResult[T] {
	return &CallResult[T]{
		call:             call,
		expectedCallback: expectedCallback,
	}
}

// IsComplete reports whether the call has completed and whether it failed.
func (c *CallResult[T]) IsComplete() (failed bool, ok bool) {
	return SteamUtils().IsAPICallCompleted(c.call)
}

// Result returns the typed call result if it is ready.
func (c *CallResult[T]) Result() (result T, failed bool, err error) {
	var zero T
	if c.call == 0 {
		return zero, false, ErrAPICallNotReady
	}
	size := unsafe.Sizeof(result)
	if size > math.MaxInt32 {
		return zero, false, ErrAPICallTooLarge
	}
	failed, ok := SteamUtils().GetAPICallResult(c.call, uintptr(unsafe.Pointer(&result)), int32(size), c.expectedCallback)
	if !ok {
		return zero, false, ErrAPICallNotReady
	}
	return result, failed, nil
}

// Wait blocks until the call completes or the context is done.
func (c *CallResult[T]) Wait(ctx context.Context, pollInterval time.Duration) (result T, failed bool, err error) {
	if pollInterval <= 0 {
		pollInterval = 50 * time.Millisecond
	}
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return result, false, ctx.Err()
		case <-ticker.C:
			if _, ok := c.IsComplete(); ok {
				return c.Result()
			}
		}
	}
}

// WaitAndDispatch waits for completion and invokes handler with the typed result.
func (c *CallResult[T]) WaitAndDispatch(ctx context.Context, pollInterval time.Duration, handler func(T, bool)) error {
	result, failed, err := c.Wait(ctx, pollInterval)
	if err != nil {
		return err
	}
	handler(result, failed)
	return nil
}
