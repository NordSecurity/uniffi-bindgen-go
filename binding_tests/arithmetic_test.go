/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"testing"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/uniffi/arithmetic"
	"github.com/stretchr/testify/assert"
)

func TestArithmetic(t *testing.T) {
	value, err := arithmetic.Add(2, 4)
	if assert.NoError(t, err) {
		assert.Equal(t, uint64(6), value)
	}

	value, err = arithmetic.Add(4, 8)
	if assert.NoError(t, err) {
		assert.Equal(t, uint64(12), value)
	}

	value, err = arithmetic.Sub(0, 2)
	assert.ErrorIs(t, err, arithmetic.ErrArithmeticErrorIntegerOverflow)

	value, err = arithmetic.Sub(4, 2)
	if assert.NoError(t, err) {
		assert.Equal(t, uint64(2), value)
	}

	value, err = arithmetic.Sub(8, 4)
	if assert.NoError(t, err) {
		assert.Equal(t, uint64(4), value)
	}

	assert.Equal(t, uint64(2), arithmetic.Div(8, 4))

	// TODO: assert that a panic happens
	// value, err = arithmetic.Div(8, 0)
	// assert.EqualError(t, err, "Integer overflow on an operation with 0 and 2")

	assert.True(t, arithmetic.Equal(2, 2))
	assert.True(t, arithmetic.Equal(4, 4))

	assert.False(t, arithmetic.Equal(2, 4))
	assert.False(t, arithmetic.Equal(4, 8))
}
