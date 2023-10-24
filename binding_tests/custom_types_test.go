/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"net/url"
	"testing"

    "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/custom_types"

    "github.com/stretchr/testify/assert"
)

func unwrap[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func TestCustomTypes(t *testing.T) {
	// Get the custom types and check their data
	demo := custom_types.GetCustomTypesDemo(nil)

	// URL is customized on the bindings side
	assert.Equal(t, *unwrap(url.Parse("http://example.com/")), demo.Url)
	// Handle isn't so it appears as a plain Long
	assert.Equal(t, int64(123), demo.Handle)

	// Change some data and ensure that the round-trip works
	demo.Url = *unwrap(url.Parse("http://new.example.com/"))
	demo.Handle = 456
	assert.Equal(t, demo, custom_types.GetCustomTypesDemo(&demo))
}
