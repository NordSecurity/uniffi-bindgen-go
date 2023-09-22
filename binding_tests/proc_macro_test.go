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

	traitImpl := obj.GetTrait(nil)
	assert.Equal(t, traitImpl.Name(), "TraitImpl")
	assert.Equal(t, obj.GetTrait(&traitImpl).Name(), "TraitImpl")
	assert.Equal(t, GetTraitNameByRef(traitImpl), "TraitImpl")

	assert.Equal(t, EnumIdentity(MaybeBoolTrue), MaybeBoolTrue)

	// just make sure this works / doesn't crash
	_ = Three{Obj: obj}

	assert.Equal(t, MakeZero().Inner, "ZERO")
	assert.Equal(t, MakeRecordWithBytes().SomeBytes, []byte{
		0, 1, 2, 3, 4})

	assert.EqualError(t, AlwaysFails(), "BasicError: OsError")
	assert.Nil(t, obj.DoStuff(5))
	assert.EqualError(t, obj.DoStuff(0), "FlatError: InvalidInput: Invalid input")

	CallCallbackInterface(GoTestCallbackInterface{})
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
