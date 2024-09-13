package binding_tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/fixture_callbacks"

	"github.com/stretchr/testify/assert"
)

type getters struct{}

func (getters) GetBool(v bool, arg2 bool) (bool, *fixture_callbacks.SimpleError) {
	if arg2 {
		return false, fixture_callbacks.NewSimpleErrorBadArgument()
	}
	return v, nil
}
func (getters) GetString(v string, arg2 bool) (string, *fixture_callbacks.SimpleError) {
	if arg2 {
		return "", fixture_callbacks.NewSimpleErrorUnexpectedError()
	}
	return v, nil
}

func (getters) GetOption(v *string, arg2 bool) (*string, *fixture_callbacks.ComplexError) {
	if arg2 {
		errMsg := "nil input"
		if v != nil {
			errMsg = *v
		}
		return nil, fixture_callbacks.NewComplexErrorUnexpectedErrorWithReason(errMsg)
	}
	return v, nil
}

func (getters) GetList(v []int32, arg2 bool) ([]int32, *fixture_callbacks.SimpleError) {
	if arg2 {
		return nil, fixture_callbacks.NewSimpleErrorBadArgument()
	}
	return v, nil
}

func (getters) GetNothing(v string) *fixture_callbacks.SimpleError {
	if v == "bad-argument" {
		return fixture_callbacks.NewSimpleErrorBadArgument()
	}
	if v == "unexpected-error" {
		return fixture_callbacks.NewSimpleErrorUnexpectedError()
	}
	return nil
}

type invalidGetters struct{}

func (invalidGetters) GetBool(v bool, arg2 bool) (bool, *fixture_callbacks.SimpleError) {
	return false, &fixture_callbacks.SimpleError{}
}

func (invalidGetters) GetString(v string, arg2 bool) (string, *fixture_callbacks.SimpleError) {
	return "", &fixture_callbacks.SimpleError{}
}

func (invalidGetters) GetOption(v *string, arg2 bool) (*string, *fixture_callbacks.ComplexError) {
	return nil, &fixture_callbacks.ComplexError{}
}

func (invalidGetters) GetList(v []int32, arg2 bool) ([]int32, *fixture_callbacks.SimpleError) {
	return nil, &fixture_callbacks.SimpleError{}
}

func (invalidGetters) GetNothing(v string) *fixture_callbacks.SimpleError {
	return fixture_callbacks.NewSimpleErrorBadArgument()
}

type goStringifier struct{}

func (goStringifier) FromSimpleType(value int32) string {
	return fmt.Sprintf("Go: %d", value)
}

func (goStringifier) FromComplexType(values *[]*float64) string {
	if values == nil {
		return "Go: nil"
	}

	var strNumbers []string
	for _, num := range *values {
		if num != nil {
			strNumbers = append(strNumbers, fmt.Sprintf("%f", *num))
		} else {
			strNumbers = append(strNumbers, "nil")
		}
	}

	return fmt.Sprintf("Go: %s", strings.Join(strNumbers, " "))
}

type testGetterInput[T any] struct {
	name          string
	value         T
	getError      bool
	expectedRes   T
	expectedError error
}

func testGetter[T any](t *testing.T, tt testGetterInput[T], getterFn func(callbacks fixture_callbacks.ForeignGetters, value T, flag bool) (T, error)) {
	foreignGetters := getters{}
	res, err := getterFn(foreignGetters, tt.value, tt.getError)
	assert.Equal(t, tt.expectedRes, res)
	assert.ErrorIs(t, err, tt.expectedError)
}

func TestRustGetters_GetNothing(t *testing.T) {
	foreignGetters := getters{}
	getters := fixture_callbacks.NewRustGetters()
	err := getters.GetNothing(foreignGetters, "bad-argument")
	assert.ErrorIs(t, err, fixture_callbacks.ErrSimpleErrorBadArgument)
	err = getters.GetNothing(foreignGetters, "unexpected-error")
	assert.ErrorIs(t, err, fixture_callbacks.ErrSimpleErrorUnexpectedError)
	err = getters.GetNothing(foreignGetters, "foo")
	assert.Equal(t, err, nil)
}

func TestRustGetters_GetBool(t *testing.T) {
	for _, tt := range []testGetterInput[bool]{
		{
			name:          "true",
			value:         true,
			getError:      false,
			expectedRes:   true,
			expectedError: nil,
		},
		{
			name:          "false",
			value:         false,
			getError:      false,
			expectedRes:   false,
			expectedError: nil,
		},
		{
			name:          "error 1",
			value:         false,
			getError:      true,
			expectedRes:   false,
			expectedError: fixture_callbacks.ErrSimpleErrorBadArgument,
		},
		{
			name:     "error 2",
			value:    true,
			getError: true,
			// Result is either a return value or an error in Rust, therefore default
			// value will be received in Go
			expectedRes:   false,
			expectedError: fixture_callbacks.ErrSimpleErrorBadArgument,
		},
	} {
		t.Run(tt.name, func(*testing.T) {
			testGetter(t, tt, fixture_callbacks.NewRustGetters().GetBool)
		})
	}
}

func TestRustGetters_GetString(t *testing.T) {
	for _, tt := range []testGetterInput[string]{
		{
			name:          "Case 1",
			value:         "cAse 1!",
			getError:      false,
			expectedRes:   "cAse 1!",
			expectedError: nil,
		},
		{
			name:          "Case 2",
			value:         "CaSE@#$%&*()_\n2",
			getError:      false,
			expectedRes:   "CaSE@#$%&*()_\n2",
			expectedError: nil,
		},
		{
			name:          "error",
			value:         "Error",
			getError:      true,
			expectedRes:   "",
			expectedError: fixture_callbacks.ErrSimpleErrorUnexpectedError,
		},
	} {

		t.Run(tt.name, func(*testing.T) {
			testGetter(t, tt, fixture_callbacks.NewRustGetters().GetString)
		})
	}
}

func TestRustGetters_GetOption(t *testing.T) {
	case1 := "cAse 1!"
	for _, tt := range []testGetterInput[*string]{
		{
			name:          "happy path",
			value:         &case1,
			getError:      false,
			expectedRes:   &case1,
			expectedError: nil,
		},
		{
			name:          "nil ok",
			value:         nil,
			getError:      false,
			expectedRes:   nil,
			expectedError: nil,
		},
	} {

		t.Run(tt.name, func(*testing.T) {
			testGetter(t, tt, fixture_callbacks.NewRustGetters().GetOption)
		})
	}
}

func TestRustGetters_GetOptionComplexError(t *testing.T) {
	res, err := fixture_callbacks.NewRustGetters().GetOption(
		getters{},
		nil,
		true,
	)
	assert.Nil(t, res)
	assert.Equal(
		t,
		fixture_callbacks.NewComplexErrorUnexpectedErrorWithReason("nil input"),
		err,
	)
}

func TestRustGetters_GetList(t *testing.T) {
	for _, tt := range []testGetterInput[[]int32]{
		{
			name:          "case 1",
			value:         []int32{1, 2, 3},
			getError:      false,
			expectedRes:   []int32{1, 2, 3},
			expectedError: nil,
		},
		{
			name:          "case 2",
			value:         []int32{3, 2, 1, 0, -1, -99},
			getError:      false,
			expectedRes:   []int32{3, 2, 1, 0, -1, -99},
			expectedError: nil,
		},
		{
			name:          "nil",
			value:         nil,
			getError:      false,
			expectedRes:   nil,
			expectedError: nil,
		},
		{
			name:          "error",
			value:         []int32{1, 2, 3},
			getError:      true,
			expectedRes:   nil,
			expectedError: fixture_callbacks.ErrSimpleErrorBadArgument,
		},
	} {

		t.Run(tt.name, func(*testing.T) {
			testGetter(t, tt, fixture_callbacks.NewRustGetters().GetList)
		})
	}
}

func TestRustGetters_GetStringOptionalCallback(t *testing.T) {
	caseOK := "case ok"
	var foreignGetters fixture_callbacks.ForeignGetters = getters{}
	var invalidForeignGetters fixture_callbacks.ForeignGetters = invalidGetters{}
	for _, tt := range []struct {
		name           string
		v              string
		flag           bool
		expectedRes    *string
		expectedError  error
		foreignGetters *fixture_callbacks.ForeignGetters
	}{
		{
			name:           "case OK",
			v:              caseOK,
			flag:           false,
			expectedRes:    &caseOK,
			expectedError:  nil,
			foreignGetters: &foreignGetters,
		},
		{
			name:           "error",
			v:              caseOK,
			flag:           true,
			expectedRes:    nil,
			expectedError:  fixture_callbacks.ErrSimpleErrorUnexpectedError,
			foreignGetters: &foreignGetters,
		},
		{
			name:           "nil getters OK",
			v:              caseOK,
			flag:           false,
			expectedRes:    nil,
			expectedError:  nil,
			foreignGetters: nil,
		},
		{
			name:           "nil getters flag true still OK",
			v:              caseOK,
			flag:           true,
			expectedRes:    nil,
			expectedError:  nil,
			foreignGetters: nil,
		},
		{
			name:           "invalid error",
			v:              caseOK,
			flag:           true,
			expectedRes:    nil,
			expectedError:  fixture_callbacks.ErrSimpleErrorUnexpectedError,
			foreignGetters: &invalidForeignGetters,
		},
	} {

		t.Run(tt.name, func(*testing.T) {
			rustGetters := fixture_callbacks.NewRustGetters()
			res, err := rustGetters.GetStringOptionalCallback(
				tt.foreignGetters,
				tt.v,
				tt.flag,
			)
			assert.Equal(t, tt.expectedRes, res)
			assert.ErrorIs(t, err, tt.expectedError)
		})
	}
}

func TestRustGetters_CallbackReferenceDoesNotInvalidateOtherReferences(t *testing.T) {
	stringifier := goStringifier{}
	rustStringifier1 := fixture_callbacks.NewRustStringifier(stringifier)

	{
		rustStringifier2 := fixture_callbacks.NewRustStringifier(stringifier)
		assert.Equal(t, "Go: 123", rustStringifier2.FromSimpleType(123))
		rustStringifier2.Destroy()
		// `stringifier` must remain valid after `rustStringifier2` drops the reference
	}

	assert.Equal(t, "Go: 321", rustStringifier1.FromSimpleType(321))
	rustStringifier1.Destroy()
}
