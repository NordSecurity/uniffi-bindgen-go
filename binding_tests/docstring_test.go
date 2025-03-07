package binding_tests

import (
	"testing"

	docstrings "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/uniffi_docstring"

	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"strings"
)

// See https://github.com/mozilla/uniffi-rs/blob/v0.28.3/fixtures/docstring/tests/test_generated_bindings.rs#L13
var DOCSTRINGS []string = []string{
	"<docstring-alternate-constructor>",
	"<docstring-associated-enum-variant-2>",
	"<docstring-associated-enum-variant>",
	"<docstring-associated-enum>",
	"<docstring-associated-error-variant-2>",
	"<docstring-associated-error-variant>",
	"<docstring-associated-error>",
	"<docstring-callback-method>",
	"<docstring-callback>",
	"<docstring-enum-variant-2>",
	"<docstring-enum-variant>",
	"<docstring-enum>",
	"<docstring-error-variant-2>",
	"<docstring-error-variant>",
	"<docstring-error>",
	"<docstring-function>",
	"<docstring-method>",
	"<docstring-namespace>",
	"<docstring-object>",
	"<docstring-primary-constructor>",
	"<docstring-record-field>",
	"<docstring-record>",
}

type ExampleCallback struct{}

func (_ ExampleCallback) Test() {
}

// Make sure symbols are not accidentally commented out by docstrings
func TestSymbolsWithDocstringsExist(t *testing.T) {
	docstrings.Test()

	_ = docstrings.EnumTestOne
	_ = docstrings.EnumTestTwo

	_ = docstrings.AssociatedEnumTestTest{0}
	_ = docstrings.AssociatedEnumTestTest2{0}

	_ = docstrings.ErrErrorTestOne
	_ = docstrings.ErrErrorTestTwo

	_ = docstrings.NewAssociatedErrorTestTest(0)
	_ = docstrings.NewAssociatedErrorTestTest2(0)

	obj := docstrings.NewObjectTest()
	obj = docstrings.ObjectTestNewAlternate()
	obj.Test()

	record := docstrings.RecordTest{0}
	_ = record.Test

	var callback docstrings.CallbackTest
	callback = ExampleCallback{}
	callback.Test()
}

func TestDocstringsAppearInBindings(t *testing.T) {
	bindingsContent := readDocstringBindingsFile(t)
	expectedDocstrings := getExpectedDocstrings(t)

	var missingDocstrings []string

	for _, docstring := range expectedDocstrings {
		if !strings.Contains(bindingsContent, docstring) {
			missingDocstrings = append(missingDocstrings, docstring)
		}
	}

	assert.Empty(t, missingDocstrings)
}

func readDocstringBindingsFile(t *testing.T) string {
	cwd, err := os.Getwd()
	assert.NoError(t, err)

	file, err := os.Open(fmt.Sprintf("%s/generated/uniffi_docstring/uniffi_docstring.go", cwd))
	assert.NoError(t, err)
	defer file.Close()

	bytes, err := io.ReadAll(file)
	assert.NoError(t, err)

	return string(bytes)
}

func getExpectedDocstrings(t *testing.T) []string {
	return DOCSTRINGS
}
