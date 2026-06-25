/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"runtime"
	"testing"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/issue83"
)

type noopStringReceiver struct{}

func (noopStringReceiver) Receive(string) {}

const stackSweepDepth = 1024

//go:noinline
func invokeAtStackDepth(depth int) {
	if depth <= 0 {
		issue83.InvokeStringReceiver(noopStringReceiver{})
		return
	}
	var pad [16]byte
	pad[0] = byte(depth)
	invokeAtStackDepth(depth - 1)
	runtime.KeepAlive(&pad)
}

// Ensure dangling pointers do not cause panic when stack is scanned.
// https://github.com/NordSecurity/uniffi-bindgen-go/issues/83
func TestIssue83(t *testing.T) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for depth := 0; depth < stackSweepDepth; depth++ {
			runtime.GC()
			invokeAtStackDepth(depth)
		}
	}()
	<-done
}
