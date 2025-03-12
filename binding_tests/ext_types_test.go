//go:build ignore

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	netUrl "net/url"
	"testing"

	itl "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/imported_types_lib"
	uo "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/uniffi_one_ns"

	"github.com/stretchr/testify/assert"
)

func TestExtTypesLib(t *testing.T) {
	ct := itl.GetCombinedType(nil)
	assert.Equal(t, ct.Uot.Sval, "hello")
	assert.Equal(t, ct.Guid, "a-guid")
	assert.Equal(t, ct.Url, *unwrap(netUrl.Parse("http://example.com/")))

	ct2 := itl.GetCombinedType(&ct)
	assert.Equal(t, ct, ct2)

	url := *unwrap(netUrl.Parse("http://example.com/"))
	assert.Equal(t, itl.GetUrl(url), url)
	assert.Equal(t, itl.GetMaybeUrl(&url), &url)
	assert.Nil(t, itl.GetMaybeUrl(nil))
	assert.Equal(t, itl.GetUrls([]netUrl.URL{url}), []netUrl.URL{url})
	assert.Equal(t, itl.GetMaybeUrls([]*netUrl.URL{&url, nil}), []*netUrl.URL{&url, nil})

	assert.Equal(t, itl.GetUniffiOneType(uo.UniffiOneType{Sval: "hello"}).Sval, "hello")
	assert.Equal(t, itl.GetMaybeUniffiOneType(&uo.UniffiOneType{Sval: "hello"}).Sval, "hello")
	assert.Nil(t, itl.GetMaybeUniffiOneType(nil))
	assert.Equal(
		t,
		itl.GetUniffiOneTypes([]uo.UniffiOneType{uo.UniffiOneType{Sval: "hello"}}),
		[]uo.UniffiOneType{uo.UniffiOneType{Sval: "hello"}},
	)
	assert.Equal(
		t,
		itl.GetMaybeUniffiOneTypes([]*uo.UniffiOneType{&uo.UniffiOneType{Sval: "hello"}, nil}),
		[]*uo.UniffiOneType{&uo.UniffiOneType{Sval: "hello"}, nil},
	)

	// Hack around golang not being able to take addresses of a constant
	one := uo.UniffiOneEnumOne

	assert.Equal(t, itl.GetUniffiOneEnum(uo.UniffiOneEnumOne), uo.UniffiOneEnumOne)
	assert.Equal(t, itl.GetMaybeUniffiOneEnum(&one), &one)
	assert.Nil(t, itl.GetMaybeUniffiOneEnum(nil))
	assert.Equal(
		t,
		itl.GetUniffiOneEnums([]uo.UniffiOneEnum{uo.UniffiOneEnumOne}),
		[]uo.UniffiOneEnum{uo.UniffiOneEnumOne},
	)
	assert.Equal(
		t,
		itl.GetMaybeUniffiOneEnums([]*uo.UniffiOneEnum{&one, nil}),
		[]*uo.UniffiOneEnum{&one, nil},
	)

	assert.Equal(t, ct.Ecd.Sval, "ecd")
	assert.Equal(t, itl.GetExternalCrateInterface("foo").Value(), "foo")
}
