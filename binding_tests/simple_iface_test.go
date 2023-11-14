/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"testing"

	. "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/uniffi_simple_iface"

	"github.com/stretchr/testify/assert"
)

func TestSimpleIface(t *testing.T) {
	obj := MakeObject(9000)
	assert.Equal(t, obj.GetInner(), int32(9000))
	_ = obj.SomeMethod()
}
