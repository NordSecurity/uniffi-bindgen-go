/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"strconv"
	"testing"

	. "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/proc_macro"

	"github.com/stretchr/testify/assert"
)

func TestProcMacro(t *testing.T) {
	one := MakeOne(123)
	assert.Equal(t, one.Inner, int32(123))
	assert.Equal(t, OneInnerByRef(one), int32(123))

	two := Two{A: "a"}
	assert.Equal(t, TakeTwo(two), "a")

	rwb := RecordWithBytes{SomeBytes: []byte{1, 2, 3}}
	assert.Equal(t, TakeRecordWithBytes(rwb), []byte{1, 2, 3})

	obj := NewObject()
	obj = ObjectNamedCtor(1)
	assert.Equal(t, obj.IsHeavy(), MaybeBoolUncertain)
	obj2 := NewObject()
	assert.Equal(t, obj.IsOtherHeavy(obj2), MaybeBoolUncertain)

	robj := NewRenamed()
	assert.True(t, robj.Func())
	assert.True(t, RenameTest())

	traitImpl := obj.GetTrait(nil)
	assert.Equal(t, traitImpl.ConcatStrings("foo", "bar"), "foobar")
	assert.Equal(t, obj.GetTrait(&traitImpl).ConcatStrings("foo", "bar"), "foobar")
	assert.Equal(t, ConcatStringsByRef(traitImpl, "foo", "bar"), "foobar")

	traitImpl2 := obj.GetTraitWithForeign(nil)
	assert.Equal(t, traitImpl2.Name(), "RustTraitImpl")
	assert.Equal(t, obj.GetTraitWithForeign(&traitImpl2).Name(), "RustTraitImpl")

	assert.Equal(t, EnumIdentity(MaybeBoolTrue), MaybeBoolTrue)

	// just make sure this works / doesn't crash
	_ = Three{Obj: obj}

	assert.Equal(t, MakeZero().Inner, "ZERO")
	assert.Equal(t, MakeRecordWithBytes().SomeBytes, []byte{0, 1, 2, 3, 4})

	assert.Equal(t, MakeHashmap(1, 2), map[int8]uint64{1: 2})
	d := map[int8]uint64{1: 2}
	assert.Equal(t, ReturnHashmap(d), d)

	assert.Equal(t, Join([]string{"a", "b", "c"}, ":"), "a:b:c")

	assert.EqualError(t, AlwaysFails(), "BasicError: OsError")
	assert.Nil(t, obj.DoStuff(5))
	assert.EqualError(t, obj.DoStuff(0), "FlatError: InvalidInput: Invalid input")

	// Defaults not supported

	CallCallbackInterface(GoTestCallbackInterface{})

	// # udl exposed functions with procmacro types.
	assert.Equal(t, GetOne(nil).Inner, int32(0))
	assert.Equal(t, GetBool(nil), MaybeBoolUncertain)
	assert.Equal(t, GetObject(nil).IsHeavy(), MaybeBoolUncertain)
	assert.Equal(t, GetTraitWithForeign(nil).Name(), "RustTraitImpl")
	assert.Nil(t, GetExternals(nil).One)

	assert.Equal(t, uint(MaybeBoolTrue), uint(1))
	assert.Equal(t, uint(MaybeBoolFalse), uint(2))
	assert.Equal(t, uint(MaybeBoolUncertain), uint(3))

	assert.Equal(t, uint8(ReprU8One), uint8(1))
	assert.Equal(t, uint8(ReprU8Three), uint8(3))

	assert.Equal(t, GetMixedEnum(nil), MixedEnumInt{Field0: 1})
	var me MixedEnum = MixedEnumNone{}
	assert.Equal(t, GetMixedEnum(&me), MixedEnumNone{})
	me = MixedEnumString{Field0: "hello"}
	assert.Equal(t, GetMixedEnum(&me), MixedEnumString{Field0: "hello"})
	me = MixedEnumAll{
		S: "string",
		I: 2,
	}
	assert.Equal(t, GetMixedEnum(&me), MixedEnumAll{S: "string", I: 2})
}

type GoTestCallbackInterface struct{}

func (c GoTestCallbackInterface) DoNothing() {}

func (c GoTestCallbackInterface) Add(a, b uint32) uint32 {
	return a + b
}

func (c GoTestCallbackInterface) Optional(a *uint32) uint32 {
	if a == nil {
		return 0
	}
	return *a
}

func (c GoTestCallbackInterface) WithBytes(rwb RecordWithBytes) []byte {
	return rwb.SomeBytes
}

func (c GoTestCallbackInterface) TryParseInt(value string) (uint32, *BasicError) {
	if value == "force-unexpected-error" {
		// raise an error that's not expected
		return 0, NewBasicErrorUnexpectedError("some other error")
	}
	parsed, ok := strconv.ParseUint(value, 10, 64)
	if ok != nil {
		return 0, NewBasicErrorInvalidInput()
	}

	return uint32(parsed), nil
}

func (c GoTestCallbackInterface) CallbackHandler(h *Object) uint32 {
	v := h.TakeError(NewBasicErrorInvalidInput())
	return v
}

func (c GoTestCallbackInterface) GetOtherCallbackInterface() OtherCallbackInterface {
	return GoTestCallbackInterface2{}
}

type GoTestCallbackInterface2 struct{}

func (c GoTestCallbackInterface2) Multiply(a, b uint32) uint32 {
	return a * b
}
