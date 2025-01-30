/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"testing"

	. "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/traits"

	"github.com/stretchr/testify/assert"
)

type goButton struct{}

func (btn goButton) Name() string {
	return "GoButton"
}

func TestTraits(t *testing.T) {
	for _, button := range GetButtons() {
		name := button.Name()
		// Check that the name is one of the expected values
		assert.Contains(t, []string{"go", "stop"}, name)
		// Check that we can round-trip the button through Rust
		assert.Equal(t, name, Press(button).Name())
	}

	assert.Equal(t, "GoButton", Press(goButton{}).Name())
}
