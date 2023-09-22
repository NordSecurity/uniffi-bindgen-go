/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"sync"
	"fmt"
	"testing"
	"time"

	. "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/futures/futures"
	"github.com/stretchr/testify/assert"
)

func TestFutures(t *testing.T) {
	// Test `alwaysReady`
	{
		t0 := time.Now()
		result := AlwaysReady()
		t1 := time.Now()

		assert.True(t, t1.Sub(t0) < 1*time.Millisecond)
		assert.True(t, result)
	}

	// Test record.
	{
		result := NewMyRecord("foo", 42)
		assert.Equal(t, result.A, "foo")
		assert.Equal(t, result.B, uint32(42))

		// Test `void`
		t0 := time.Now()
		Void()
		t1 := time.Now()
		
		elapsed := t1.Sub(t0)
		fmt.Printf("elapsed %s\n", elapsed)
		assert.True(t, elapsed < 1*time.Millisecond)
	}
	// Test `Sleep`
	{
		t0 := time.Now()
		result := Sleep(200)
		t1 := time.Now()

		elapsed := t1.Sub(t0)
		fmt.Printf("elapsed %s\n", elapsed)
		assert.True(t, elapsed < 250*time.Millisecond)
		assert.True(t, elapsed > 200*time.Millisecond)
		assert.True(t, result)
	}

	// Test sequential futures.
	{
		t0 := time.Now()
		resultAlice := SayAfter(100, "Alice")
		resultBob := SayAfter(200, "Bob")
		t1 := time.Now()

		elapsed := t1.Sub(t0)
		fmt.Printf("elapsed %s\n", elapsed)

		assert.True(t, elapsed < 350*time.Millisecond)
		assert.True(t, elapsed > 300*time.Millisecond)
		assert.Equal(t, resultAlice, "Hello, Alice!")
		assert.Equal(t, resultBob, "Hello, Bob!")
	}

	// Test concurrent futures.
	{
		var wg sync.WaitGroup

		t0 := time.Now()
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := SayAfter(100, "Alice")
			assert.Equal(t, result, "Hello, Alice!")
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			result := SayAfter(200, "Bob")
			assert.Equal(t, result, "Hello, Bob!")
		}()

		wg.Wait()
		t1 := time.Now()
		elapsed := t1.Sub(t0)
		fmt.Printf("elapsed %s\n", elapsed)
		assert.True(t, elapsed < 250*time.Millisecond)
		assert.True(t, elapsed > 200*time.Millisecond)

	}
	// Test async methods
	// 	let megaphone = newMegaphone()

	// 	let t0 = Date()
	// 	let result_alice = await megaphone.sayAfter(ms: 2000, who: "Alice")
	// 	let t1 = Date()

	// 	let tDelta = DateInterval(start: t0, end: t1)
	// 	assert(tDelta.duration > 2 && tDelta.duration < 2.1)
	// 	assert(result_alice == "HELLO, ALICE!")

	// Test async function returning an object
	// 	let megaphone = await asyncNewMegaphone()

	// 	let result = try await megaphone.fallibleMe(doFail: false)
	// 	assert(result == 42)

	// Test with the Tokio runtime.
	// 	let t0 = Date()
	// 	let result_alice = await sayAfterWithTokio(ms: 2000, who: "Alice")
	// 	let t1 = Date()

	// 	let tDelta = DateInterval(start: t0, end: t1)
	// 	assert(tDelta.duration > 2 && tDelta.duration < 2.1)
	// 	assert(result_alice == "Hello, Alice (with Tokio)!")

	// Test fallible function/method…
	// … which doesn't throw.
	// 	let t0 = Date()
	// 	let result = try await fallibleMe(doFail: false)
	// 	let t1 = Date()

	// 	let tDelta = DateInterval(start: t0, end: t1)
	// 	assert(tDelta.duration > 0 && tDelta.duration < 0.1)
	// 	assert(result == 42)

	// 	let m = try await fallibleStruct(doFail: false)
	// 	let result = try await m.fallibleMe(doFail: false)
	// 	assert(result == 42)

	// 	let megaphone = newMegaphone()

	// 	let t0 = Date()
	// 	let result = try await megaphone.fallibleMe(doFail: false)
	// 	let t1 = Date()

	// 	let tDelta = DateInterval(start: t0, end: t1)
	// 	assert(tDelta.duration > 0 && tDelta.duration < 0.1)
	// 	assert(result == 42)

	// … which does throw.
	// 	let t0 = Date()

	// 	do {
	// 		let _ = try await fallibleMe(doFail: true)
	// 	} catch MyError.Foo {
	// 		assert(true)
	// 	} catch {
	// 		assert(false) // should never be reached
	// 	}

	// 	let t1 = Date()

	// 	let tDelta = DateInterval(start: t0, end: t1)
	// 	assert(tDelta.duration > 0 && tDelta.duration < 0.1)

	// 	do {
	// 		let _ = try await fallibleStruct(doFail: true)
	// 	} catch MyError.Foo {
	// 		assert(true)
	// 	} catch {
	// 		assert(false)
	// 	}

	// 	let megaphone = newMegaphone()

	// 	let t0 = Date()

	// 	do {
	// 		let _ = try await megaphone.fallibleMe(doFail: true)
	// 	} catch MyError.Foo {
	// 		assert(true)
	// 	} catch {
	// 		assert(false) // should never be reached
	// 	}

	// 	let t1 = Date()

	// 	let tDelta = DateInterval(start: t0, end: t1)
	// 	assert(tDelta.duration > 0 && tDelta.duration < 0.1)

	// Test a future that uses a lock and that is cancelled.
	// 	let task = Task {
	// 	    try! await useSharedResource(options: SharedResourceOptions(releaseAfterMs: 100, timeoutMs: 1000))
	// 	}

	// 	// Wait some time to ensure the task has locked the shared resource
	// 	try await Task.sleep(nanoseconds: 50_000_000)
	// 	// Cancel the job task the shared resource has been released.
	// 	//
	// 	// FIXME: this test currently passes because `test.cancel()` doesn't actually cancel the
	// 	// operation.  We need to rework the Swift async handling to handle this properly.
	// 	task.cancel()

	// 	// Try accessing the shared resource again.  The initial task should release the shared resource
	// 	// before the timeout expires.
	// 	try! await useSharedResource(options: SharedResourceOptions(releaseAfterMs: 0, timeoutMs: 1000))

	// Test a future that uses a lock and that is not cancelled.
	// 	try! await useSharedResource(options: SharedResourceOptions(releaseAfterMs: 100, timeoutMs: 1000))
	// 	try! await useSharedResource(options: SharedResourceOptions(releaseAfterMs: 0, timeoutMs: 1000))
}
