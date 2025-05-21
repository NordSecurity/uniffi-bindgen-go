package binding_tests

import (
	"errors"
	"testing"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/error_types"

	"github.com/stretchr/testify/assert"
)

func TestNormalCatch(t *testing.T) {
	err := error_types.Oops()
	assert.ErrorContains(t, err, "because uniffi told me so\n\nCaused by:\n    oops")
}

func TestNormalCatchWithImplitArcWrapping(t *testing.T) {
	err := error_types.OopsNowrap()
	assert.ErrorContains(t, err, "because uniffi told me so\n\nCaused by:\n    oops")
}

func TestErrorInterface(t *testing.T) {
	err := error_types.Oops()
	var expectedError *error_types.ErrorInterface
	assert.ErrorAs(t, err, &expectedError)
	assert.Equal(t, []string{"because uniffi told me so", "oops"}, expectedError.Chain())

	link := expectedError.Link(0)
	assert.NotNil(t, link)
	assert.Equal(t, "because uniffi told me so", *link)
}

func TestAsyncErrorInterface(t *testing.T) {
	err := error_types.Aoops()
	assert.ErrorContains(t, err, "async-oops")
}

func TestErrorTrait(t *testing.T) {
	err := error_types.Toops()
	var expectedError *error_types.ErrorTrait
	assert.ErrorAs(t, err, &expectedError)
	assert.Equal(t, "trait-oops", expectedError.Msg())
}

func TestErrorReturn(t *testing.T) {
	err := error_types.GetError("the error")
	assert.NotNil(t, err)
	assert.Equal(t, []string{"the error"}, err.Chain())
	assert.Equal(t, "the error", err.Error())
}

func TestRichError(t *testing.T) {
	err := error_types.ThrowRich("oh no")
	assert.ErrorContains(t, err, "RichError: \"oh no\"")
}

func TestInterfaceErrors(t *testing.T) {
	_, err := error_types.TestInterfaceFallibleNew()
	assert.ErrorContains(t, err, "fallible_new")

	interfaceish := error_types.NewTestInterface()
	err = interfaceish.Oops()
	assert.ErrorContains(t, err, "because the interface told me so\n\nCaused by:\n    oops")

	err = interfaceish.Aoops()
	assert.ErrorContains(t, err, "async-oops")
}

func TestProcmacroInterfaceErrors(t *testing.T) {
	err := error_types.ThrowProcError("eek")
	var expectedError *error_types.ProcErrorInterface
	assert.ErrorAs(t, err, &expectedError)
	assert.Equal(t, "eek", expectedError.Message())
	assert.Equal(t, "ProcErrorInterface(eek)", expectedError.Error())
}

func TestEnumError(t *testing.T) {
	for _, tc := range []struct {
		name     string
		code     uint16
		expected *error_types.Error
	}{
		{"Oops", 0, error_types.NewErrorOops()},
		{"Value", 1, error_types.NewErrorValue("value")},
		{"Int", 2, error_types.NewErrorIntValue(2)},
		{"Inner", 5, error_types.NewErrorInnerError(error_types.NewInnerCaseA("inner"))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := error_types.OopsEnum(tc.code)
			if assert.NotNil(t, err) {
				assert.EqualValues(t, tc.expected, err)
			}
		})
	}
}

func TestEnumErrorFlatInner(t *testing.T) {
	err := error_types.OopsEnum(3)
	if assert.NotNil(t, err) {
		assert.ErrorIs(t, err, error_types.ErrErrorFlatInnerError)
		assert.ErrorContains(t, err, "Error: FlatInnerError: Error_=FlatInner: CaseA: inner")
	}

	err = error_types.OopsEnum(4)
	if assert.NotNil(t, err) {
		assert.ErrorIs(t, err, error_types.ErrErrorFlatInnerError)
		assert.ErrorContains(t, err, "Error: FlatInnerError: Error_=FlatInner: CaseB: NonUniffiTypeValue: value")
	}
}

func TestTupleError(t *testing.T) {
	terr := error_types.GetTuple(nil)
	assert.ErrorContains(t, terr, "oops")

	err1 := error_types.OopsTuple(0)
	if errors.As(err1, &terr) {
		assert.Equal(t, error_types.NewTupleErrorOops("oops"), terr)
	}

	err2 := error_types.OopsTuple(1)
	if errors.As(err2, &terr) {
		assert.Equal(t, error_types.NewTupleErrorValue(1), terr)
	}
}
