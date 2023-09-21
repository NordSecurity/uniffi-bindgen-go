/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"testing"

	. "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/proc_macro/proc_macro"
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
	_ = Three{ Obj: obj}

	assert.Equal(t, MakeZero().Inner, "ZERO")
	assert.Equal(t, MakeRecordWithBytes().SomeBytes, []byte{
		0, 1, 2, 3, 4})

	// do {
	// 	try alwaysFails()
	// 	fatalError("alwaysFails should have thrown")
	// } catch BasicError.OsError {
	// }

	// try! obj.doStuff(times: 5)

	// do {
	// 	try obj.doStuff(times: 0)
	// 	fatalError("doStuff should throw if its argument is 0")
	// } catch FlatError.InvalidInput {
	// }

	// struct SomeOtherError: Error { }

	// class SwiftTestCallbackInterface : TestCallbackInterface {
	// 	func doNothing() { }

	// 	func add(a: UInt32, b: UInt32) -> UInt32 {
	// 		return a + b;
	// 	}

	// 	func `optional`(a: Optional<UInt32>) -> UInt32 {
	// 		return a ?? 0;
	// 	}

	// 	func withBytes(rwb: RecordWithBytes) -> Data {
	// 		return rwb.someBytes
	// 	}

	// 	func tryParseInt(value: String) throws -> UInt32 {
	// 		if (value == "force-unexpected-error") {
	// 			// raise an error that's not expected
	// 			throw SomeOtherError()
	// 		}
	// 		let parsed = UInt32(value)
	// 		if parsed != nil {
	// 			return parsed!
	// 		} else {
	// 			throw BasicError.InvalidInput
	// 		}
	// 	}

	// 	func callbackHandler(h: Object) -> UInt32 {
	// 		var v = h.takeError(e: BasicError.InvalidInput)
	// 		return v
	// 	}
	// }

	// testCallbackInterface(cb: SwiftTestCallbackInterface())
}
