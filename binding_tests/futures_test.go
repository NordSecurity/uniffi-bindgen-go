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

	. "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/futures"

	"github.com/stretchr/testify/assert"
)

func assertInstantExecution(t *testing.T, start time.Time) {
	t1 := time.Now()
	elapsed := t1.Sub(start)

	fmt.Printf("elapsed %s\n", elapsed)
	assert.True(t, elapsed < 10*time.Millisecond)
	assert.True(t, elapsed > 0)
}

func assertDelayedExecution(t *testing.T, start time.Time, delay time.Duration) {
	t1 := time.Now()
	elapsed := t1.Sub(start)

	fmt.Printf("elapsed %s\n", elapsed)
	assert.True(t, elapsed < delay+50*time.Millisecond)
	assert.True(t, elapsed > delay)
}

func TestFuturesAlwaysReady(t *testing.T) {
	// Test `alwaysReady`
	t0 := time.Now()
	result := AlwaysReady()
	assertInstantExecution(t, t0)
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
	assertInstantExecution(t, t0)
}

func TestFuturesSleep(t *testing.T) {
	// Test `Sleep`
	t0 := time.Now()
	result := Sleep(200)

	assertDelayedExecution(t, t0, 200*time.Millisecond)
	assert.True(t, result)
}

func TestFuturesSequential(t *testing.T) {
	// Test sequential futures.
	t0 := time.Now()
	resultAlice := SayAfter(100, "Alice")
	resultBob := SayAfter(200, "Bob")

	assertDelayedExecution(t, t0, 300*time.Millisecond)
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

	assertDelayedExecution(t, t0, 200*time.Millisecond)
}

func TestFuturesAsyncMethods(t *testing.T) {
	// Test async methods
	megaphone := NewMegaphone()

	t0 := time.Now()
	resultAlice := megaphone.SayAfter(200, "Alice")

	assertDelayedExecution(t, t0, 200*time.Millisecond)
	assert.Equal(t, resultAlice, "HELLO, ALICE!")
}

func TestFuturesAsyncConstructor(t *testing.T) {
	// Async constructors are supported
	megaphone := NewMegaphone()
	assert.NotNil(t, megaphone)

	megaphone = MegaphoneSecondary()
	t0 := time.Now()
	msg := megaphone.SayAfter(20, "Alice")
	assertDelayedExecution(t, t0, 20*time.Millisecond)
	assert.Equal(t, "HELLO, ALICE!", msg)

	udl_megaphone := MegaphoneSecondary()
	t0 = time.Now()
	msg = udl_megaphone.SayAfter(25, "udl")
	assertDelayedExecution(t, t0, 25*time.Millisecond)
	assert.Equal(t, "HELLO, UDL!", msg)
}

func TestFuturesAsyncTraitMethods(t *testing.T) {
	traits := GetSayAfterTraits()
	t0 := time.Now()

	res1 := traits[0].SayAfter(100, "Alice")
	res2 := traits[1].SayAfter(100, "Bob")

	assertDelayedExecution(t, t0, 200*time.Millisecond)
	assert.Equal(t, "Hello, Alice!", res1)
	assert.Equal(t, "Hello, Bob!", res2)
}

func TestFuturesAsyncUdlTraitMethods(t *testing.T) {
	traits := GetSayAfterUdlTraits()
	t0 := time.Now()

	res1 := traits[0].SayAfter(100, "Alice")
	res2 := traits[1].SayAfter(100, "Bob")

	assertDelayedExecution(t, t0, 200*time.Millisecond)
	assert.Equal(t, "Hello, Alice!", res1)
	assert.Equal(t, "Hello, Bob!", res2)
}

type goAsyncParser struct {
	completedDelays uint
}

func (gap *goAsyncParser) AsString(delayMs int32, value int32) string {
	time.Sleep(time.Duration(delayMs) * time.Millisecond)
	return fmt.Sprintf("%d", value)
}

func (gap *goAsyncParser) TryFromString(delayMs int32, value string) (int32, *ParserError) {
	time.Sleep(time.Duration(delayMs) * time.Millisecond)
	if value == "force-unexpected-exception" {
		return 0, NewParserErrorUnexpectedError()
	}
	val, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return int32(val), NewParserErrorNotAnInt()
	}
	return int32(val), nil
}

func (gap *goAsyncParser) Delay(delayMs int32) {
	time.Sleep(time.Duration(delayMs) * time.Millisecond)
	gap.completedDelays += 1
}

func (gap *goAsyncParser) TryDelay(delayMs string) *ParserError {
	ms, err := strconv.ParseInt(delayMs, 10, 32)
	if err != nil {
		return NewParserErrorNotAnInt()
	}

	time.Sleep(time.Duration(ms) * time.Millisecond)
	gap.completedDelays += 1
	return nil
}

func TestFuturesForeignAsyncTraitMethods(t *testing.T) {
	traitObj := &goAsyncParser{}

	t0 := time.Now()
	assert.Equal(t, "42", AsStringUsingTrait(traitObj, 10, 42))
	val, err := TryFromStringUsingTrait(traitObj, 10, "42")
	if assert.Nil(t, err) {
		assert.Equal(t, int32(42), val)
	}
	assertDelayedExecution(t, t0, 20*time.Millisecond)
	val, err = TryFromStringUsingTrait(traitObj, 0, "fourty-two")
	assert.ErrorIs(t, err, ErrParserErrorNotAnInt)
	val, err = TryFromStringUsingTrait(traitObj, 0, "force-unexpected-exception")
	assert.ErrorIs(t, err, ErrParserErrorUnexpectedError)

	t0 = time.Now()
	DelayUsingTrait(traitObj, 15)
	TryDelayUsingTrait(traitObj, "15")
	assertDelayedExecution(t, t0, 30*time.Millisecond)
	assert.Equal(t, uint(2), traitObj.completedDelays)

	CancelDelayUsingTrait(traitObj, 10)
	// Sleep a bit longer then a delay to confirm that task was acutaly cancled
	time.Sleep(20)
	assert.Equal(t, uint(2), traitObj.completedDelays)
}

func TestFuturesAsyncObjectParam(t *testing.T) {
	megaphone := NewMegaphone()
	t0 := time.Now()
	resultAlice := SayAfterWithMegaphone(megaphone, 200, "Alice")

	assertDelayedExecution(t, t0, 200*time.Millisecond)
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

	assertDelayedExecution(t, t0, 200*time.Millisecond)
	assert.Equal(t, resultAlice, "Hello, Alice (with Tokio)!")
}

func TestFuturesFallibleNoThrow(t *testing.T) {
	// Test fallible function/method…
	// … which doesn't throw.
	{
		t0 := time.Now()
		result, err := FallibleMe(false)

		assertInstantExecution(t, t0)
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
		assert.Nil(t, err)

		assertInstantExecution(t, t0)
		assert.Equal(t, result, uint8(42))
	}
}

func TestFuturesFallibleThrows(t *testing.T) {
	// … which does throw.
	{
		t0 := time.Now()

		_, err := FallibleMe(true)
		assert.EqualError(t, err, "MyError: Foo")
		assertInstantExecution(t, t0)

		_, err = FallibleStruct(true)
		assert.EqualError(t, err, "MyError: Foo")
	}
	{
		megaphone := NewMegaphone()

		t0 := time.Now()
		_, err := megaphone.FallibleMe(true)

		assertInstantExecution(t, t0)
		assert.EqualError(t, err, "MyError: Foo")
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
