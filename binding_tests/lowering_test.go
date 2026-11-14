/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"strings"
	"testing"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/coverall"
	usf "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/uniffi_simple_fns"

	"github.com/stretchr/testify/assert"
)

func patternedBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i*7 + i/251)
	}
	return b
}

func reversed(b []byte) []byte {
	r := make([]byte, len(b))
	for i, v := range b {
		r[len(b)-1-i] = v
	}
	return r
}

func TestLowerBytesRoundTrip(t *testing.T) {
	coveralls := coverall.NewCoveralls("lowering")
	defer coveralls.Destroy()

	cases := map[string][]byte{
		"empty":    {},
		"one byte": {0xab},
		"1KiB":     patternedBytes(1 << 10),
		"1MiB+7":   patternedBytes(1<<20 + 7),
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, reversed(input), coveralls.Reverse(input))
		})
	}
}

func TestLowerStringRoundTrip(t *testing.T) {
	cases := map[string]string{
		"empty":     "",
		"ascii":     "hello, world",
		"accents":   "crème brûlée, naïve façade",
		"cjk":       "日本語のテキスト",
		"emoji":     "🦀🐹 ok",
		"nul":       "before\x00after",
		"over 1MiB": strings.Repeat("ñ日🦀", 1<<20/9+1),
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, input, usf.StringIdentity(input))
		})
	}
}
