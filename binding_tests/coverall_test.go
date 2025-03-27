/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/coverall"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	defer EnsureNoOneIsAlive()
	m.Run()
}

func EnsureNoOneIsAlive() {
	if num := coverall.GetNumAlive(); num != 0 {
		panic(fmt.Errorf("There are still living objects!! n=%d", num))
	}
}

func TestCoverallDicts(t *testing.T) {
	some := coverall.CreateSomeDict()
	defer some.Destroy()

	assert.Equal(t, "text", some.Text)
	assert.Equal(t, "maybe_text", *some.MaybeText)
	assert.True(t, some.ABool)
	assert.False(t, *some.MaybeABool)
	assert.Equal(t, uint8(1), some.Unsigned8)
	assert.Equal(t, uint8(2), *some.MaybeUnsigned8)

	assert.Equal(t, uint16(3), some.Unsigned16)
	assert.Equal(t, uint16(4), *some.MaybeUnsigned16)
	assert.Equal(t, uint64(18446744073709551615), some.Unsigned64)
	assert.Equal(t, uint64(0), *some.MaybeUnsigned64)
	assert.Equal(t, int8(8), some.Signed8)
	assert.Equal(t, int8(0), *some.MaybeSigned8)
	assert.Equal(t, int64(9223372036854775807), some.Signed64)
	assert.Equal(t, int64(0), *some.MaybeSigned64)

	assert.Equal(t, float32(1.2345), some.Float32)
	assert.Equal(t, float32(22.0/7.0), *some.MaybeFloat32)
	assert.Equal(t, 0.0, some.Float64)
	assert.Equal(t, 1.0, *some.MaybeFloat64)

	assert.Equal(t, "some_dict", (*some.Coveralls).GetName())
	assert.Equal(t, "node-2", (*some.TestTrait).Name())

	noneDict := coverall.CreateNoneDict()
	defer noneDict.Destroy()

	assert.Equal(t, "text", noneDict.Text)
	assert.Nil(t, noneDict.MaybeText)
	assert.Equal(t, []byte("some_bytes"), noneDict.SomeBytes)
	assert.Nil(t, noneDict.MaybeSomeBytes)
	assert.True(t, noneDict.ABool)
	assert.Nil(t, noneDict.MaybeABool)
	assert.Equal(t, uint8(1), noneDict.Unsigned8)
	assert.Nil(t, noneDict.MaybeUnsigned8)
	assert.Equal(t, uint16(3), noneDict.Unsigned16)
	assert.Nil(t, noneDict.MaybeUnsigned16)
	assert.Equal(t, uint64(18446744073709551615), noneDict.Unsigned64)
	assert.Nil(t, noneDict.MaybeUnsigned64)
	assert.Equal(t, int8(8), noneDict.Signed8)
	assert.Nil(t, noneDict.MaybeSigned8)
	assert.Equal(t, int64(9223372036854775807), noneDict.Signed64)
	assert.Nil(t, noneDict.MaybeSigned64)
	assert.Equal(t, float32(1.2345), noneDict.Float32)
	assert.Nil(t, noneDict.MaybeFloat32)
	assert.Equal(t, float64(0.0), noneDict.Float64)
	assert.Nil(t, noneDict.MaybeFloat64)
	assert.Nil(t, noneDict.Coveralls)
	assert.Nil(t, noneDict.TestTrait)
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

	mErr := coverall.ThrowMacroError()
	assert.ErrorIs(t, mErr, coverall.ErrCoverallMacroErrorTooManyMacros)

	_, err = coveralls.MaybeThrowInto(true)
	assert.ErrorIs(t, err, coverall.ErrCoverallErrorTooManyHoles)

	assert.PanicsWithError(t, "oops", func() {
		coveralls.Panic("oops")
	})
}

func TestCoverallFlatErrors(t *testing.T) {
	fe := coverall.ThrowFlatError()
	assert.ErrorIs(t, fe, coverall.ErrCoverallFlatErrorTooManyVariants)

	me := coverall.ThrowFlatMacroError()
	assert.ErrorIs(t, me, coverall.ErrCoverallFlatMacroErrorTooManyVariants)

	re := coverall.ThrowRichErrorNoVariantData()
	assert.ErrorIs(t, re, coverall.ErrCoverallRichErrorNoVariantDataTooManyPlainVariants)
}

func TestCoverallComplexErrors(t *testing.T) {
	coveralls := coverall.NewCoveralls("test_complex_errors")
	defer coveralls.Destroy()

	{
		v, err := coveralls.MaybeThrowComplex(0)
		if assert.Nil(t, err) {
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

type GoGetters struct{}

func (g *GoGetters) GetBool(v bool, arg2 bool) bool {
	return v != arg2
}

func (g *GoGetters) GetString(v string, arg2 bool) (string, *coverall.CoverallError) {
	switch v {
	case "too-many-holes":
		return "", coverall.NewCoverallErrorTooManyHoles()
	case "unexpected-error":
		panic(fmt.Errorf("oh no"))
	}

	if arg2 {
		return strings.ToUpper(v), nil
	}
	return v, nil
}

func (g *GoGetters) GetOption(v string, arg2 bool) (*string, *coverall.ComplexError) {
	switch v {
	case "os-error":
		return nil, coverall.NewComplexErrorOsError(100, 200)
	case "unknown-error":
		return nil, coverall.NewComplexErrorUnknownError()
	}

	if arg2 {
		if v == "" {
			return nil, nil
		}
		upper := strings.ToUpper(v)
		return &upper, nil
	}
	return &v, nil
}

func (g *GoGetters) GetList(v []int32, arg2 bool) []int32 {
	if arg2 {
		return v
	}
	return []int32{}
}

func (g *GoGetters) GetNothing(v string) {}

func (g *GoGetters) RoundTripObject(coveralls *coverall.Coveralls) *coverall.Coveralls {
	return coveralls
}

// Implementation of the GoNode struct
type GoNode struct {
	parent *coverall.NodeTrait
}

func (n *GoNode) Name() string {
	return "node-go"
}

func (n *GoNode) SetParent(parent *coverall.NodeTrait) {
	n.parent = parent
}

func (n *GoNode) GetParent() *coverall.NodeTrait {
	return n.parent
}

func (n *GoNode) StrongCount() uint64 {
	return 0 // Needs proper implementation
}

// Test suite similar to TraitsTest in Python
func TestGoGetters(t *testing.T) {
	assert := assert.New(t)
	goGetters := &GoGetters{}
	assert.PanicsWithError("oh no", func() {
		coverall.TestGetters(goGetters)
	})

	assert.Equal(false, goGetters.GetBool(true, true))
	assert.Equal(true, goGetters.GetBool(true, false))
	assert.Equal(true, goGetters.GetBool(false, true))
	assert.Equal(false, goGetters.GetBool(false, false))

	result, caErr := goGetters.GetString("hello", false)
	assert.Nil(caErr)
	assert.Equal("hello", result)

	result, caErr = goGetters.GetString("hello", true)
	assert.Nil(caErr)
	assert.Equal("HELLO", result)

	optionResult, cxErr := goGetters.GetOption("hello", true)
	assert.Nil(cxErr)
	assert.Equal("HELLO", *optionResult)

	optionResult, cxErr = goGetters.GetOption("hello", false)
	assert.Nil(cxErr)
	assert.Equal("hello", *optionResult)

	optionResult, cxErr = goGetters.GetOption("", true)
	assert.Nil(cxErr)
	assert.Nil(optionResult)

	assert.Equal([]int32{1, 2, 3}, goGetters.GetList([]int32{1, 2, 3}, true))
	assert.Equal([]int32{}, goGetters.GetList([]int32{1, 2, 3}, false))

	goGetters.GetNothing("hello")

	_, caErr = goGetters.GetString("too-many-holes", true)
	assert.ErrorIs(caErr, coverall.ErrCoverallErrorTooManyHoles)

	_, cxErr = goGetters.GetOption("os-error", true)
	assert.ErrorIs(cxErr, coverall.ErrComplexErrorOsError)

	_, cxErr = goGetters.GetOption("unknown-error", true)
	assert.ErrorIs(cxErr, coverall.ErrComplexErrorUnknownError)
}

func TestPath(t *testing.T) {
	{
		assert := assert.New(t)
		traits := coverall.GetTraits()
		assert.Equal("node-1", traits[0].Name())
		assert.Equal(uint64(2), traits[0].StrongCount())

		assert.Equal("node-2", traits[1].Name())
		assert.Equal(uint64(2), traits[1].StrongCount())

		traits[0].SetParent(&traits[1])
		assert.Equal(uint64(2), traits[1].StrongCount())
		assert.ElementsMatch([]string{"node-2"}, coverall.AncestorNames(traits[0]))
		assert.ElementsMatch([]string{}, coverall.AncestorNames(traits[1]))
		assert.Equal("node-2", (*traits[0].GetParent()).Name())

		var node coverall.NodeTrait = &GoNode{}
		traits[1].SetParent(&node)
		assert.ElementsMatch([]string{"node-2", "node-go"}, coverall.AncestorNames(traits[0]))
		assert.ElementsMatch([]string{"node-go"}, coverall.AncestorNames(traits[1]))
		assert.ElementsMatch([]string{}, coverall.AncestorNames(node))

		traits[1].SetParent(nil)
		node.SetParent(&traits[0])
		assert.ElementsMatch([]string{"node-1", "node-2"}, coverall.AncestorNames(node))
		assert.ElementsMatch([]string{"node-2"}, coverall.AncestorNames(traits[0]))
		assert.ElementsMatch([]string{}, coverall.AncestorNames(traits[1]))

		node.SetParent(nil)
		traits[0].SetParent(nil)
	}
	runtime.GC()
	time.Sleep(20 * time.Millisecond)
}

func TestRoundTripping(t *testing.T) {
	{
		rustGetters := coverall.MakeRustGetters()
		coverall.TestRoundTripThroughRust(rustGetters)
		coverall.TestRoundTripThroughForeign(&GoGetters{})
	}
	runtime.GC()
	time.Sleep(20 * time.Millisecond)
}

func TestRustOnlyTraits(t *testing.T) {
	assert := assert.New(t)
	traits := coverall.GetStringUtilTraits()
	assert.Equal("cowboy", traits[0].Concat("cow", "boy"))
	assert.Equal("cowboy", traits[1].Concat("cow", "boy"))
}
