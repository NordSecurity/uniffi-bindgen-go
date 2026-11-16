/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/futures"

	"github.com/stretchr/testify/assert"
)

// ms cycles through 0, 1 and 2 so the inline Ready, Waked and stored-callback scheduler paths all
// run, and each call holds a TimerFuture thread so more waiters are live than the free list holds.
func TestFuturesManyConcurrentCalls(t *testing.T) {
	const calls = 10_000
	var wg sync.WaitGroup
	wg.Add(calls)
	for i := 0; i < calls; i++ {
		go func(i int) {
			defer wg.Done()
			assert.Equal(t, fmt.Sprintf("Hello, %d!", i), futures.SayAfter(uint16(i%3), strconv.Itoa(i)))
		}(i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(60 * time.Second):
		t.Fatal("concurrent SayAfter calls did not finish within 60s")
	}
}
