/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"testing"

	usf "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/uniffi_simple_fns/uniffi_simple_fns"
	"github.com/stretchr/testify/assert"
)

func TestSimpleFns(t *testing.T) {
	assert.Equal(t, usf.GetString(), "String created by Rust")
	assert.Equal(t, usf.GetInt(), int32(1289))
	assert.Equal(t, usf.StringIdentity("String created by Kotlin"), "String created by Kotlin")
	assert.Equal(t, usf.ByteToU32(255), uint32(255))

	aSet := usf.NewSet()
	usf.AddToSet(aSet, "foo")
	usf.AddToSet(aSet, "bar")
	assert.True(t, usf.SetContains(aSet, "foo"))
	assert.True(t, usf.SetContains(aSet, "bar"))
	assert.False(t, usf.SetContains(aSet, "baz"))
}
