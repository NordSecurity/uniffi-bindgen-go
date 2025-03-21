package binding_tests

import (
	"regexp"
	"testing"

	docstrings "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/uniffi_docstring"

	"fmt"
	"io"
	"os"

	"github.com/stretchr/testify/assert"
)

var DOCSTRINGS []string = []string{
	"<docstring-alternate-constructor>",
	"<docstring-associated-enum>",
	"<docstring-associated-enum-variant>",
	"<docstring-associated-enum-variant-2>",
	"<docstring-associated-error>",
	"<docstring-associated-error-variant>",
	"<docstring-associated-error-variant-2>",
	"<docstring-callback>",
	"<docstring-callback-method>",
	"<docstring-enum>",
	"<docstring-enum-variant>",
	"<docstring-enum-variant-2>",
	"<docstring-error>",
	"<docstring-error-variant>",
	"<docstring-error-variant-2>",
	"<docstring-function>",
	"<docstring-method>",
	"<docstring-multiline-function>",
	"<docstring-namespace>",
	"<docstring-object>",
	"<docstring-primary-constructor>",
	"<docstring-record>",
	"<docstring-record-field>",
	"<second-line>",
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
	actualDocstrings := getDocstringsFromBindingsFile(t)
	expectedDocstrings := getExpectedDocstrings(t)

	for docstring, _ := range expectedDocstrings {
		assert.Containsf(t, actualDocstrings, docstring, "Missing %s", docstring)
	}

	for docstring, _ := range actualDocstrings {
		assert.Containsf(t, expectedDocstrings, docstring, "Unexpected %s", docstring)
	}
}

func getDocstringsFromBindingsFile(t *testing.T) map[string]struct{} {
	cwd, err := os.Getwd()
	assert.NoError(t, err)

	file, err := os.Open(fmt.Sprintf("%s/generated/uniffi_docstring/uniffi_docstring.go", cwd))
	assert.NoError(t, err)
	defer file.Close()

	bytes, err := io.ReadAll(file)
	assert.NoError(t, err)

	re, err := regexp.Compile(`// (<[^>]*>)`)
	assert.NoError(t, err)

	found := re.FindAllStringSubmatch(string(bytes), -1)
	assert.NotNil(t, found, "Failed to find any of the docstrings")

	set := make(map[string]struct{})
	for _, doc := range found {
		set[doc[1]] = struct{}{}
	}

	return set
}

func getExpectedDocstrings(t *testing.T) map[string]struct{} {
	set := make(map[string]struct{})
	for _, doc := range DOCSTRINGS {
		set[doc] = struct{}{}
	}
	return set
}
