/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"testing"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/geometry"

	"github.com/stretchr/testify/assert"
)

func TestGeometry(t *testing.T) {
	line1 := geometry.Line{geometry.Point{0, 0}, geometry.Point{1, 2}}
	line2 := geometry.Line{geometry.Point{1, 1}, geometry.Point{2, 2}}

	assert.Equal(t, float64(2), geometry.Gradient(line1))
	assert.Equal(t, float64(1), geometry.Gradient(line2))

	assert.Equal(t, &geometry.Point{0, 0}, geometry.Intersection(line1, line2))
	assert.Nil(t, geometry.Intersection(line1, line1))
}
