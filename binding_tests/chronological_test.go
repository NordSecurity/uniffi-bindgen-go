/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"math"
	"testing"
	"time"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/chronological"
	"github.com/stretchr/testify/assert"
)

func parseDuration(s string) time.Duration {
	duration, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return duration
}

func TestTimestampMinMax(t *testing.T) {
	min := time.Unix(math.MinInt64, math.MinInt32)
	value, err := chronological.ReturnTimestamp(min)
	if assert.Nil(t, err) {
		assert.Equal(t, min, value)
	}

	max := time.Unix(math.MaxInt64, math.MaxInt32)
	value, err = chronological.ReturnTimestamp(max)
	if assert.Nil(t, err) {
		assert.Equal(t, max, value)
	}
}

func TestDurationMax(t *testing.T) {
	// Rust does not allow negative timespan, so only maximum value is tested.

	max := time.Duration(math.MaxInt64)
	value, err := chronological.ReturnDuration(max)
	if assert.Nil(t, err) {
		assert.Equal(t, max, value)
	}
}

func TestChronologicalWorks(t *testing.T) {
	assert.Equal(
		t,
		"2022-12-16T14:13:12.123456789Z",

		chronological.ToStringTimestamp(
			time.Date(2022, 12, 16, 14, 13, 12, 123456789, time.UTC)))

	// Test passing timestamp and duration while returning timestamp
	value1, err := chronological.Add(time.Unix(100, 1), parseDuration("1s1ns"))
	if assert.Nil(t, err) {
		assert.Equal(t, time.Unix(101, 2), value1)
	}

	// Test passing timestamp while returning duration
	value2, err := chronological.Diff(time.Unix(101, 2), time.Unix(100, 1))
	if assert.Nil(t, err) {
		assert.Equal(t, parseDuration("1s1ns"), value2)
	}

	_, err = chronological.Diff(time.Unix(100, 0), time.Unix(101, 0))
	assert.ErrorIs(t, err, chronological.ErrChronologicalErrorTimeDiffError)
}

func TestPreEpochTimestampsSerializesCorrectly(t *testing.T) {
	assert.Equal(
		t,
		"1969-12-12T00:00:00.000000000Z",
		chronological.ToStringTimestamp(
			time.Date(1969, 12, 12, 0, 0, 0, 0, time.UTC)))

	// [-999_999_999; 0) is unrepresentable
	// https://github.com/mozilla/uniffi-rs/issues/1433
	// assert.Equal(
	// 	t,
	// 	"1969-12-31T23:59:59.999999999Z",
	// 	chronological.ToStringTimestamp(
	// 		time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC)))

	assert.Equal(
		t,
		"1969-12-31T23:59:58.999999999Z",
		chronological.ToStringTimestamp(
			time.Date(1969, 12, 31, 23, 59, 58, 999999999, time.UTC)))

	assert.Equal(
		t,
		time.Date(1969, 12, 31, 23, 59, 58, 999000000, time.UTC),
		chronological.GetPreEpochTimestamp().UTC())

	value, err := chronological.Add(time.Date(1955, 11, 5, 0, 6, 0, 283000001, time.UTC), parseDuration("1s1ns"))
	if assert.Nil(t, err) {
		assert.Equal(t, time.Date(1955, 11, 5, 0, 6, 1, 283000002, time.UTC), value.UTC())
	}
}

func TestTimeWorksLikeRustSystemTime(t *testing.T) {
	// Sleep inbetween to make sure that the clock has enough resolution
	before := time.Now()
	time.Sleep(time.Millisecond)
	now := chronological.Now()
	time.Sleep(time.Millisecond)
	after := time.Now()
	assert.True(t, before.Before(now))
	assert.True(t, now.Before(after))
}

func TestTimeAndDurationOptionals(t *testing.T) {
	now := time.Now()
	duration := parseDuration("0s")

	assert.True(t, chronological.Optional(&now, &duration))
	assert.False(t, chronological.Optional(nil, &duration))
	assert.False(t, chronological.Optional(&now, nil))
}
