//go:build ignore

// TODO(pna): fix me

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"errors"
	. "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/name_case"
	"testing"
)

type MyNameCaseCallback struct{}

func (*MyNameCaseCallback) Test() {}

func TestNameCaseCompiles(t *testing.T) {
	_ = EnumTestVariantOne
	_ = AssociatedEnumTestVariantTest{0}

	var expectedError *ErrorTestVariantOne
	errors.As(NewErrorTestVariantOne(), &expectedError)

	var expectedAssociatedError *AssociatedErrorTestVariantTest
	errors.As(NewAssociatedErrorTestVariantTest(0), &expectedAssociatedError)

	var _ *ObjectTest = NewObjectTest()
	_ = RecordTest{0}
	var _ CallbackTest = &MyNameCaseCallback{}
}
