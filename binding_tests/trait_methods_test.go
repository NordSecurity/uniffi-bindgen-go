/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"testing"

	. "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/trait_methods"

	"github.com/stretchr/testify/assert"
)

func TestTraitMethods(t *testing.T) {
	m1 := NewTraitMethods("yo")
	m2 := NewTraitMethods("yo")
	m3 := NewTraitMethods("yah")

	assert.Equal(t, "TraitMethods(yo)", m1.String())
	assert.Equal(t, "TraitMethods(yah)", m3.String())

	assert.Equal(t, "TraitMethods { val: \"yo\" }", m1.DebugString())
	assert.Equal(t, "TraitMethods { val: \"yah\" }", m3.DebugString())

	assert.Equal(t, uint64(0x90dbae7908cd1bd), m1.Hash())
	assert.Equal(t, uint64(0x90dbae7908cd1bd), m2.Hash())
	assert.Equal(t, uint64(0x40fb01c5911b5fe1), m3.Hash())

	assert.True(t, m1.Eq(m1))
	assert.True(t, m1.Eq(m2))
	assert.False(t, m1.Eq(m3))
	assert.False(t, m1.Ne(m1))
	assert.False(t, m1.Ne(m2))
	assert.True(t, m1.Ne(m3))

	assert.False(t, m3.Eq(m1))
	assert.False(t, m3.Eq(m2))
	assert.True(t, m3.Eq(m3))
	assert.True(t, m3.Ne(m1))
	assert.True(t, m3.Ne(m2))
	assert.False(t, m3.Ne(m3))
}
