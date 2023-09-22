/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"testing"
	"time"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/coverall"
	"github.com/stretchr/testify/assert"
)

func TestCoverall(t *testing.T) {
	d := coverall.CreateSomeDict()
	defer d.Destroy()

	assert.Equal(t, "text", d.Text)
	assert.Equal(t, "maybe_text", *d.MaybeText)
	assert.True(t, d.ABool)
	assert.False(t, *d.MaybeABool)
	assert.Equal(t, uint8(1), d.Unsigned8)
	assert.Equal(t, uint8(2), *d.MaybeUnsigned8)

	assert.Equal(t, uint16(3), d.Unsigned16)
	assert.Equal(t, uint16(4), *d.MaybeUnsigned16)
	assert.Equal(t, uint64(18446744073709551615), d.Unsigned64)
	assert.Equal(t, uint64(0), *d.MaybeUnsigned64)
	assert.Equal(t, int8(8), d.Signed8)
	assert.Equal(t, int8(0), *d.MaybeSigned8)
	assert.Equal(t, int64(9223372036854775807), d.Signed64)
	assert.Equal(t, int64(0), *d.MaybeSigned64)

	assert.Equal(t, float32(1.2345), d.Float32)
	assert.Equal(t, float32(22.0/7.0), *d.MaybeFloat32)
	assert.Equal(t, 0.0, d.Float64)
	assert.Equal(t, 1.0, *d.MaybeFloat64)

	assert.Equal(t, "some_dict", (*d.Coveralls).GetName())
}

func TestCoverallArcs(t *testing.T) {
	coveralls := coverall.NewCoveralls("test_arcs")

	assert.Equal(t, uint64(1), coverall.GetNumAlive())
	// One ref held by the foreign-language code, one created for this method call.
	assert.Equal(t, uint64(2), coveralls.StrongCount())
	assert.Nil(t, coveralls.GetOther())
	coveralls.TakeOther(&coveralls)
	// Should now be a new strong ref, held by the object's reference to itself.
	assert.Equal(t, uint64(3), coveralls.StrongCount())
	assert.Equal(t, uint64(1), coverall.GetNumAlive())

	other := *coveralls.GetOther()
	assert.Equal(t, "test_arcs", (*other).GetName())
	other.Destroy()

	assert.ErrorIs(t, coveralls.TakeOtherFallible(), coverall.ErrCoverallErrorTooManyHoles)

	assert.PanicsWithError(t, "expected panic: with an arc!", func() {
		coveralls.TakeOtherPanic("expected panic: with an arc!")
	})

	assert.PanicsWithError(t, "Expected panic in a fallible function!", func() {
		coveralls.FalliblePanic("Expected panic in a fallible function!")
	})

	coveralls.TakeOther(nil)
	assert.Equal(t, uint64(2), coveralls.StrongCount())

	coveralls.Destroy()
	assert.Equal(t, uint64(0), coverall.GetNumAlive())
}

func TestCoverallReturnObjects(t *testing.T) {
	coveralls := coverall.NewCoveralls("test_return_objects")
	assert.Equal(t, uint64(1), coverall.GetNumAlive())
	assert.Equal(t, uint64(2), coveralls.StrongCount())

	c2 := coveralls.CloneMe()
	assert.Equal(t, coveralls.GetName(), c2.GetName())
	assert.Equal(t, uint64(2), coverall.GetNumAlive())
	assert.Equal(t, uint64(2), c2.StrongCount())

	coveralls.TakeOther(&c2)
	// same number alive but `c2` has an additional ref count.
	assert.Equal(t, uint64(2), coverall.GetNumAlive())
	assert.Equal(t, uint64(2), coveralls.StrongCount())
	assert.Equal(t, uint64(3), c2.StrongCount())

	c2.Destroy()
	// Here we've dropped C# reference to `c2`, but the rust struct will not
	// be dropped as coveralls hold an `Arc<>` to it.
	assert.Equal(t, uint64(2), coverall.GetNumAlive())

	coveralls.Destroy()
	assert.Equal(t, uint64(0), coverall.GetNumAlive())
}

func TestCoverallSimpleErrors(t *testing.T) {
	coveralls := coverall.NewCoveralls("test_simple_errors")
	defer coveralls.Destroy()

	_, err := coveralls.MaybeThrow(true)
	assert.ErrorIs(t, err, coverall.ErrCoverallErrorTooManyHoles)

	_, err = coveralls.MaybeThrowInto(true)
	assert.ErrorIs(t, err, coverall.ErrCoverallErrorTooManyHoles)

	assert.PanicsWithError(t, "oops", func() {
		coveralls.Panic("oops")
	})
}

func TestCoverallComplexErrors(t *testing.T) {
	coveralls := coverall.NewCoveralls("test_complex_errors")
	defer coveralls.Destroy()

	{
		v, err := coveralls.MaybeThrowComplex(0)
		if assert.NoError(t, err) {
			assert.True(t, v)
		}
	}

	{
		_, err := coveralls.MaybeThrowComplex(1)
		var osErr *coverall.ComplexErrorOsError
		if assert.ErrorAs(t, err, &osErr) {
			assert.Equal(t, int16(10), osErr.Code)
			assert.Equal(t, int16(20), osErr.ExtendedCode)
		}
	}

	{
		_, err := coveralls.MaybeThrowComplex(2)
		var permissionErr *coverall.ComplexErrorPermissionDenied
		if assert.ErrorAs(t, err, &permissionErr) {
			assert.Equal(t, "Forbidden", permissionErr.Reason)
		}
	}

	{
		_, err := coveralls.MaybeThrowComplex(3)
		var unErr *coverall.ComplexErrorUnknownError
		assert.ErrorAs(t, err, &unErr)
	}

	assert.PanicsWithError(t, "Invalid input", func() {
		coveralls.MaybeThrowComplex(4)
	})
}

func TestCoverallInterfacesInDicts(t *testing.T) {
	coveralls := coverall.NewCoveralls("test_interface_in_dicts")
	defer coveralls.Destroy()

	coveralls.AddPatch(coverall.NewPatch(coverall.ColorRed))
	coveralls.AddRepair(coverall.Repair{time.Now(), coverall.NewPatch(coverall.ColorBlue)})
	assert.Equal(t, 2, len(coveralls.GetRepairs()))
}

func TestCoverallMultiThreadedCallsWork(t *testing.T) {
	// Make sure that there is no blocking during concurrent FFI calls.

	counter := coverall.NewThreadsafeCounter()
	defer counter.Destroy()

	const waitMillis = 10

	finished1 := make(chan struct{})
	go func() {
		// block the thread
		counter.BusyWait(waitMillis)
		finished1 <- struct{}{}
	}()

	count := int32(0)
	finished2 := make(chan struct{})
	go func() {
		for i := 0; i < waitMillis; i++ {
			// `count` is only incremented if another thread is blocking the counter.
			// This ensures that both calls are running concurrently.
			count = counter.IncrementIfBusy()
			time.Sleep(time.Millisecond)
		}
		finished2 <- struct{}{}
	}()

	<-finished1
	<-finished2
	assert.True(t, count > 0)
}
