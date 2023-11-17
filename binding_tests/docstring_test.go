/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"testing"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/uniffi_docstring"
)

func TestDocstring(_ *testing.T) {
	_ = uniffi_docstring.EnumTestOne
	_ = uniffi_docstring.AssociatedEnumTestTest{}
	_ = uniffi_docstring.NewErrorTestOne()

	_ = uniffi_docstring.NewObjectTest()
	obj2 := uniffi_docstring.ObjectTestNewAlternate()
	obj2.Test()

	rec := uniffi_docstring.RecordTest { Test: 123 }
	_ = rec.Test

	// class CallbackImpls: CallbackTest {
	// 	func test() {}
	// }
}
