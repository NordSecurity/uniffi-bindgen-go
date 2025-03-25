package binding_tests

import (
	"testing"
	"time"

	. "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/uniffi_example_futures"

	"github.com/stretchr/testify/assert"
)

func TestSayAfter(t *testing.T) {
	start := time.Now()
	msg := SayAfter(20, "Alice")
	elapsed := time.Since(start)
	assert.Equal(t, "Hello, Alice!", msg)
	assert.GreaterOrEqual(t, elapsed, 20*time.Millisecond)
}
