/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"fmt"
	"sync"
	"testing"
	"time"

	. "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/futures"

	"github.com/stretchr/testify/assert"
)

func TestFuturesAlwaysReady(t *testing.T) {
	// Test `alwaysReady`

	t0 := time.Now()
	result := AlwaysReady()
	t1 := time.Now()

	assert.True(t, t1.Sub(t0) < 1*time.Millisecond)
	assert.True(t, result)
}

func TestFuturesRecord(t *testing.T) {
	// Test record.

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
func TestFuturesSleep(t *testing.T) {
	// Test `Sleep`
	t0 := time.Now()
	result := Sleep(200)
	t1 := time.Now()

	elapsed := t1.Sub(t0)
	fmt.Printf("elapsed %s\n", elapsed)
	assert.True(t, elapsed < 250*time.Millisecond)
	assert.True(t, elapsed > 200*time.Millisecond)
	assert.True(t, result)
}

func TestFuturesSequential(t *testing.T) {
	// Test sequential futures.
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

func TestFuturesConcurrent(t *testing.T) {
	// Test concurrent futures.
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

func TestFuturesAsyncMethods(t *testing.T) {
	// Test async methods
	megaphone := NewMegaphone()

	t0 := time.Now()
	resultAlice := megaphone.SayAfter(200, "Alice")
	t1 := time.Now()

	elapsed := t1.Sub(t0)
	fmt.Printf("elapsed %s\n", elapsed)
	assert.True(t, elapsed < 250*time.Millisecond)
	assert.True(t, elapsed > 200*time.Millisecond)
	assert.Equal(t, resultAlice, "HELLO, ALICE!")
}

func TestFuturesAsyncFunctionRetObject(t *testing.T) {
	// Test async function returning an object
	megaphone := AsyncNewMegaphone()

	result, err := megaphone.FallibleMe(false)
	assert.Nil(t, err)
	assert.Equal(t, result, uint8(42))
}

func TestFuturesTokio(t *testing.T) {
	// Test with the Tokio runtime.
	t0 := time.Now()
	resultAlice := SayAfterWithTokio(200, "Alice")
	t1 := time.Now()

	elapsed := t1.Sub(t0)
	fmt.Printf("elapsed %s\n", elapsed)
	assert.True(t, elapsed < 250*time.Millisecond)
	assert.True(t, elapsed > 200*time.Millisecond)
	assert.Equal(t, resultAlice, "Hello, Alice (with Tokio)!")
}

func TestFuturesFallibleNoThrow(t *testing.T) {
	// Test fallible function/method…
	// … which doesn't throw.
	{
		t0 := time.Now()
		result, err := FallibleMe(false)
		t1 := time.Now()

		elapsed := t1.Sub(t0)
		fmt.Printf("elapsed %s\n", elapsed)
		assert.True(t, elapsed < 1*time.Millisecond)
		assert.True(t, elapsed > 0*time.Millisecond)

		assert.Nil(t, err)
		assert.Equal(t, result, uint8(42))

	}
	{
		m, err := FallibleStruct(false)
		assert.Nil(t, err)
		result, err := m.FallibleMe(false)
		assert.Nil(t, err)
		assert.Equal(t, result, uint8(42))
	}
	{
		megaphone := NewMegaphone()

		t0 := time.Now()
		result, err := megaphone.FallibleMe(false)
		t1 := time.Now()
		assert.Nil(t, err)

		elapsed := t1.Sub(t0)
		fmt.Printf("elapsed %s\n", elapsed)
		assert.True(t, elapsed < 1*time.Millisecond)
		assert.True(t, elapsed > 0*time.Millisecond)
		assert.Equal(t, result, uint8(42))
	}
}

func TestFuturesFallibleThrows(t *testing.T) {
	// … which does throw.
	{
		t0 := time.Now()

		_, err := FallibleMe(true)
		assert.EqualError(t, err, "MyError: Foo")
		t1 := time.Now()

		elapsed := t1.Sub(t0)
		fmt.Printf("elapsed %s\n", elapsed)
		assert.True(t, elapsed < 1*time.Millisecond)
		assert.True(t, elapsed > 0*time.Millisecond)

		_, err = FallibleStruct(true)
		assert.EqualError(t, err, "MyError: Foo")
	}
	{

		megaphone := NewMegaphone()

		t0 := time.Now()
		_, err := megaphone.FallibleMe(true)
		t1 := time.Now()
		assert.EqualError(t, err, "MyError: Foo")

		elapsed := t1.Sub(t0)
		fmt.Printf("elapsed %s\n", elapsed)
		assert.True(t, elapsed < 1*time.Millisecond)
		assert.True(t, elapsed > 0*time.Millisecond)
	}
}

func TestFuturesLockAndCancel(t *testing.T) {
	// Test a future that uses a lock and that is cancelled.

	cancel := make(chan struct{})
	go func() {
		done := make(chan struct{})
		go func() {
			UseSharedResource(SharedResourceOptions{ReleaseAfterMs: 100, TimeoutMs: 1000})
			done <- struct{}{}
		}()

		select {
		case <-done:
			fmt.Printf("Task finished\n")
			return
		case <-cancel:
			fmt.Printf("Task canceled\n")
			return
		}
	}()

	// Wait some time to ensure the task has locked the shared resource
	time.Sleep(50 * time.Millisecond)

	// Cancel the job task the shared resource has been released.
	cancel <- struct{}{}

	// Try accessing the shared resource again.  The initial task should release the shared resource
	// before the timeout expires.
	UseSharedResource(SharedResourceOptions{ReleaseAfterMs: 0, TimeoutMs: 1000})
}

func TestFuturesLockAndNotCancel(t *testing.T) {
	// Test a future that uses a lock and that is not cancelled.
	go func() {
		UseSharedResource(SharedResourceOptions{ReleaseAfterMs: 100, TimeoutMs: 1000})
	}()

	UseSharedResource(SharedResourceOptions{ReleaseAfterMs: 0, TimeoutMs: 1000})
}
