/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"testing"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/issue43"
	"github.com/stretchr/testify/assert"
)

// Ensure you can call async functions which return types in a different package.
// See https://github.com/NordSecurity/uniffi-bindgen-go/issues/43

func TestIssue43(t *testing.T) {
	record := issue43.GetAsyncExternalType()
	assert.Equal(t, record.Id, "foo")
	assert.Equal(t, record.Tag, "bar")
}
