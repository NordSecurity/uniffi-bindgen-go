/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	goerrors "errors"
	"fmt"
	"testing"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/errors"

	"github.com/stretchr/testify/assert"
)

func TestReturnErrorWithVoid(t *testing.T) {
	err := errors.TryVoid(true)
	assert.ErrorIs(t, err, errors.ErrBoobyTrapErrorIceSlip)
}

func TestNoReturnErrorWithVoid(t *testing.T) {
	err := errors.TryVoid(false)
	assert.NoError(t, err)
}

func TestReturnErrorWithValue(t *testing.T) {
	value, err := errors.TryString(true)
	assert.Equal(t, "", value)
	assert.ErrorIs(t, err, errors.ErrBoobyTrapErrorIceSlip)
}

func TestNoReturnErrorWithValue(t *testing.T) {
	value, err := errors.TryString(false)
	if assert.NoError(t, err) {
		assert.Equal(t, "hello world", value)
	}
}

func TestFlatErrorIs(t *testing.T) {
	err := errors.TryVoid(true)
	assert.ErrorIs(t, err, errors.ErrBoobyTrapErrorIceSlip)
	assert.NotErrorIs(t, err, errors.ErrBoobyTrapErrorHotDoorKnob)
}

func TestFlatErrorAs(t *testing.T) {
	err := errors.TryVoid(true)

	{
		var expectedError *errors.BoobyTrapError
		assert.ErrorAs(t, err, &expectedError)
	}

	{
		var expectedError *errors.BoobyTrapErrorIceSlip
		assert.ErrorAs(t, err, &expectedError)
	}

	{
		var expectedError *errors.BoobyTrapErrorHotDoorKnob
		assert.False(t, goerrors.As(err, &expectedError))
		assert.Nil(t, expectedError)
	}
}

func TestComplexErrorMessage(t *testing.T) {
	err := errors.ValidateMessage(100, "")
	assert.EqualError(t, err, "ValidationError: InvalidUser: UserId=100")

	err = errors.ValidateMessage(0, "byebye")
	assert.EqualError(t, err, "ValidationError: InvalidMessage: Message=byebye")

	err = errors.ValidateMessage(100, "byebye")
	assert.EqualError(t, err, "ValidationError: InvalidUserAndMessage: UserId=100, Message=byebye")

	err = errors.GetComplexError("struct")
	assert.EqualError(t, err, "ComplexError: Struct: PositionA={1 1}, PositionB={2 2}")

	err = errors.GetComplexError("list")
	assert.EqualError(t, err, "ComplexError: List: List=[{1 1} {2 2}]")

	err = errors.GetComplexError("map")
	assert.EqualError(t, err, "ComplexError: Map: Map=map[0:{1 1} 1:{2 2}]")

	{
		err = errors.GetComplexError("option")
		var optionError *errors.ComplexErrorOption
		assert.ErrorAs(t, err, &optionError)
		// `Option` is a pointer value, its not dereferenced to display the actual value :(
		assert.EqualError(t, err, fmt.Sprintf("ComplexError: Option: IdA=%v, IdB=<nil>", optionError.IdA))
	}
}

func TestComplexErrorIs(t *testing.T) {
	err := errors.ValidateMessage(100, "")
	assert.ErrorIs(t, err, errors.ErrValidationErrorInvalidUser)
	assert.NotErrorIs(t, err, errors.ErrValidationErrorInvalidMessage)
	assert.NotErrorIs(t, err, errors.ErrValidationErrorInvalidUserAndMessage)

	err = errors.ValidateMessage(0, "byebye")
	assert.NotErrorIs(t, err, errors.ErrValidationErrorInvalidUser)
	assert.ErrorIs(t, err, errors.ErrValidationErrorInvalidMessage)
	assert.NotErrorIs(t, err, errors.ErrValidationErrorInvalidUserAndMessage)

	err = errors.ValidateMessage(100, "byebye")
	assert.NotErrorIs(t, err, errors.ErrValidationErrorInvalidUser)
	assert.NotErrorIs(t, err, errors.ErrValidationErrorInvalidMessage)
	assert.ErrorIs(t, err, errors.ErrValidationErrorInvalidUserAndMessage)
}

func TestComplexErrorAs(t *testing.T) {
	{
		err := errors.ValidateMessage(100, "")
		var expectedError *errors.ValidationErrorInvalidUser
		if assert.ErrorAs(t, err, &expectedError) {
			assert.Equal(t, int32(100), expectedError.UserId)
		}
	}

	{
		err := errors.ValidateMessage(0, "byebye")
		var expectedError *errors.ValidationErrorInvalidMessage
		if assert.ErrorAs(t, err, &expectedError) {
			assert.Equal(t, "byebye", expectedError.Message)
		}
	}

	{
		err := errors.ValidateMessage(100, "byebye")
		var expectedError *errors.ValidationErrorInvalidUserAndMessage
		if assert.ErrorAs(t, err, &expectedError) {
			assert.Equal(t, int32(100), expectedError.UserId)
			assert.Equal(t, "byebye", expectedError.Message)
		}
	}
}

func TestComplexErrorAsBase(t *testing.T) {
	err := errors.ValidateMessage(100, "byebye")
	var expectedError *errors.ValidationError
	assert.ErrorAs(t, err, &expectedError)
	assert.EqualError(t, expectedError, "ValidationError: InvalidUserAndMessage: UserId=100, Message=byebye")
}

func TestErrorNamedError(t *testing.T) {
	// this test exists to ensure ErrorNamedError is not removed from the UDL without causing test failures.
	// The purpose of ErrorNamedError is to ensure that the generated bindings produce compilable Go code,
	// so there isn't really anything to actually test at runtime.
	err := errors.NewErrorNamedErrorError("it's an error")
	var expectedError *errors.ErrorNamedError
	assert.ErrorAs(t, err, &expectedError)
	assert.Equal(t, "it's an error", expectedError.Unwrap().(*errors.ErrorNamedErrorError).Error_)
}

func TestNestedError(t *testing.T) {
	assert.Equal(t, nil, errors.TryNested(false))
	err := errors.TryNested(true)
	var expectedError *errors.NestedError
	assert.ErrorAs(t, err, &expectedError)
	var expectedNestedError *errors.NestedErrorNested
	assert.ErrorAs(t, expectedError.Unwrap(), &expectedNestedError)
	assert.Equal(t, "ValidationError: UnknownError", expectedNestedError.Source.Error())
}
