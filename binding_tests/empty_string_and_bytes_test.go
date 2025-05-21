package binding_tests

import (
	"testing"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/uniffi_go_empty_string_and_bytes"
)

func TestEmptyString(t *testing.T) {
	if *uniffi_go_empty_string_and_bytes.EmptyStringTest() != "" {
		t.Error("EmptyStringTest() should return empty string")
	}
}

func TestEmptyBytes(t *testing.T) {
	if len(*uniffi_go_empty_string_and_bytes.EmptyBytesTest()) != 0 {
		t.Error("EmptyStringTest() should return empty string")
	}
}
