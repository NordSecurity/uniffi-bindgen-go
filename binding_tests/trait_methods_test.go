/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"testing"

	. "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/trait_methods/trait_methods"
	"github.com/stretchr/testify/assert"
)

func TestTraitMethods(t *testing.T) {
	m := NewTraitMethods("yo")
	assert.Equal(t, m.String(), "TraitMethods(yo)")

	// Not implemented yet are Debug, Eq, Hash
}
