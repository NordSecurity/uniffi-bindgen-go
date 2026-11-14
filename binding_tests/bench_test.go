/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"strings"
	"testing"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/coverall"
	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/futures"
	usf "github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/uniffi_simple_fns"
)

func BenchmarkLowerBytes(b *testing.B) {
	sizes := []struct {
		name string
		n    int
	}{
		{"16B", 16},
		{"1KiB", 1 << 10},
		{"64KiB", 64 << 10},
	}
	coveralls := coverall.NewCoveralls("bench")
	defer coveralls.Destroy()

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			input := make([]byte, size.n)
			for i := range input {
				input[i] = byte(i)
			}
			b.ReportAllocs()
			b.SetBytes(int64(size.n))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if got := coveralls.Reverse(input); len(got) != size.n {
					b.Fatalf("Reverse returned %d bytes, want %d", len(got), size.n)
				}
			}
		})
	}
}

func BenchmarkLowerString(b *testing.B) {
	sizes := []struct {
		name string
		n    int
	}{
		{"16B", 16},
		{"1KiB", 1 << 10},
	}
	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			input := strings.Repeat("a", size.n)
			b.ReportAllocs()
			b.SetBytes(int64(size.n))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if got := usf.StringIdentity(input); len(got) != size.n {
					b.Fatalf("StringIdentity returned %d bytes, want %d", len(got), size.n)
				}
			}
		})
	}
}

func BenchmarkAsyncAlwaysReady(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if !futures.AlwaysReady() {
			b.Fatal("AlwaysReady returned false")
		}
	}
}

func BenchmarkAsyncAlwaysReadyParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if !futures.AlwaysReady() {
				b.Error("AlwaysReady returned false")
				return
			}
		}
	})
}
