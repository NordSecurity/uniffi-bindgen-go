//go:build ignore

// TODO(pna): fix async functionality

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"testing"
	"time"

	. "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/fixture_foreign_executor"

	"github.com/stretchr/testify/assert"
)

func runTest(tester *ForeignExecutorTester, delay uint32) *TestResult {
	tester.ScheduleTest(delay)
	time.Sleep(time.Duration(delay+10) * time.Millisecond)
	return tester.GetLastResult()
}

func TestForeignExecutor(t *testing.T) {
	// Test scheduling with no delay
	result := runTest(
		NewForeignExecutorTester(UniFfiForeignExecutor{}),
		0,
	)
	assert.NotNil(t, result)
	assert.True(t, result.CallHappenedInDifferentThread)
	assert.True(t, result.DelayMs <= 10)

	// Test scheduling with delay and an executor created from a list
	result2 := runTest(
		ForeignExecutorTesterNewFromSequence(
			[]UniFfiForeignExecutor{UniFfiForeignExecutor{}},
		),
		100,
	)
	assert.NotNil(t, result2)
	assert.True(t, result2.CallHappenedInDifferentThread)
	assert.True(t, result2.DelayMs >= 90)
	assert.True(t, result2.DelayMs <= 110)
}
