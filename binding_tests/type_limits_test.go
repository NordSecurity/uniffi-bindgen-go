/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"testing"

	. "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/uniffi_type_limits"

	"github.com/stretchr/testify/assert"
)

func TestTypeLimits(t *testing.T) {
	// strings cannot contain surrogates, "\u{d800}" gives an error.
	// assert.PanicsWithError(t, "Exception", func() { TakeString("\ud800") })
	assert.Equal(t, TakeString(""), "")
	assert.Equal(t, TakeString("愛"), "愛")
	assert.Equal(t, TakeString("💖"), "💖")
}
