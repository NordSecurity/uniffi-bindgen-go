/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"testing"

	itl "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/imported_types_lib/imported_types_lib"
	"github.com/stretchr/testify/assert"
)


func TestExtTypesLib(t *testing.T) {
	ct := itl.GetCombinedType(nil)
	assert.Equal(t, ct.Uot.Sval, "hello")
}
