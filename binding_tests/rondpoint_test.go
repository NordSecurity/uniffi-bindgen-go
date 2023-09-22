/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"math"
	"testing"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/rondpoint"

	"github.com/stretchr/testify/assert"
)

func TestRondpointCopyWorks(t *testing.T) {
	dico := rondpoint.Dictionnaire{rondpoint.EnumerationDeux, true, 0, 123456789}
	copyDico := rondpoint.CopieDictionnaire(dico)
	assert.Equal(t, dico, copyDico)

	assert.Equal(t, rondpoint.EnumerationDeux, rondpoint.CopieEnumeration(rondpoint.EnumerationDeux))

	list := []rondpoint.Enumeration{rondpoint.EnumerationUn, rondpoint.EnumerationUn}
	assert.Equal(t, list, rondpoint.CopieEnumerations(list))

	dict := map[string]rondpoint.EnumerationAvecDonnees{
		"0": rondpoint.EnumerationAvecDonneesZero{},
		"1": rondpoint.EnumerationAvecDonneesUn{1},
		"2": rondpoint.EnumerationAvecDonneesDeux{2, "deux"},
	}
	assert.Equal(t, dict, rondpoint.CopieCarte(dict))

	assert.True(t, rondpoint.Switcheroo(false))
}

func compareEnums(a rondpoint.EnumerationAvecDonnees, b rondpoint.EnumerationAvecDonnees) bool {
	return a == b
}

func TestRondpointComparisonOperatorWorks(t *testing.T) {
	assert.True(t, compareEnums(rondpoint.EnumerationAvecDonneesZero{}, rondpoint.EnumerationAvecDonneesZero{}))
	assert.True(t, compareEnums(rondpoint.EnumerationAvecDonneesUn{1}, rondpoint.EnumerationAvecDonneesUn{1}))
	assert.True(t, compareEnums(rondpoint.EnumerationAvecDonneesDeux{2, "deux"}, rondpoint.EnumerationAvecDonneesDeux{2, "deux"}))

	assert.False(t, compareEnums(rondpoint.EnumerationAvecDonneesZero{}, rondpoint.EnumerationAvecDonneesUn{1}))
	assert.False(t, compareEnums(rondpoint.EnumerationAvecDonneesUn{1}, rondpoint.EnumerationAvecDonneesUn{2}))
	assert.False(t, compareEnums(rondpoint.EnumerationAvecDonneesDeux{2, "un"}, rondpoint.EnumerationAvecDonneesDeux{2, "deux"}))
}

func affirmAllerRetour[T any](t *testing.T, callback func(T) T, list ...T) {
	for _, value := range list {
		assert.Equal(t, value, callback(value))
	}
}

func TestRondpointTestRoundTrip(t *testing.T) {
	rt := rondpoint.NewRetourneur()
	defer rt.Destroy()

	meanValue := 0x1234_5678_9123_4567

	// booleans
	affirmAllerRetour(t, rt.IdentiqueBoolean, true, false)

	// bytes
	affirmAllerRetour(t, rt.IdentiqueI8, math.MinInt8, math.MaxInt8, int8(meanValue))
	affirmAllerRetour(t, rt.IdentiqueU8, math.MaxUint8, uint8(meanValue))

	// shorts
	affirmAllerRetour(t, rt.IdentiqueI16, math.MinInt16, math.MaxInt16, int16(meanValue))
	affirmAllerRetour(t, rt.IdentiqueU16, math.MaxUint16, uint16(meanValue))

	// ints
	affirmAllerRetour(t, rt.IdentiqueI32, math.MinInt32, math.MaxInt32, int32(meanValue))
	affirmAllerRetour(t, rt.IdentiqueU32, math.MaxUint32, uint32(meanValue))

	// longs
	affirmAllerRetour(t, rt.IdentiqueI64, math.MinInt64, math.MaxInt64, int64(meanValue))
	affirmAllerRetour(t, rt.IdentiqueU64, math.MaxUint64, uint64(meanValue))

	// floats
	affirmAllerRetour(t, rt.IdentiqueFloat, math.MaxFloat32, math.SmallestNonzeroFloat32)

	// doubles
	affirmAllerRetour(t, rt.IdentiqueDouble, math.MaxFloat64, math.SmallestNonzeroFloat64)

	// strings
	affirmAllerRetour(t,
		rt.IdentiqueString,
		"",
		"abc",
		"null\u0000byte",
		"été",
		"ښي لاس ته لوستلو لوستل",
		"😻emoji 👨‍👧‍👦multi-emoji, 🇨🇭a flag, a canal, panama")

	// signed record
	nombresSignes := func(v int) rondpoint.DictionnaireNombresSignes {
		return rondpoint.DictionnaireNombresSignes{int8(v), int16(v), int32(v), int64(v)}
	}
	affirmAllerRetour(t, rt.IdentiqueNombresSignes, nombresSignes(-1), nombresSignes(0), nombresSignes(1))

	// unsigned record
	nombres := func(v int) rondpoint.DictionnaireNombres {
		return rondpoint.DictionnaireNombres{uint8(v), uint16(v), uint32(v), uint64(v)}
	}
	affirmAllerRetour(t, rt.IdentiqueNombres, nombres(0), nombres(1))
}
