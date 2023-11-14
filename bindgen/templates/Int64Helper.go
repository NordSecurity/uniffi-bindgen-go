{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

type {{ ffi_converter_name }} struct{}

var {{ ffi_converter_name }}INSTANCE = {{ ffi_converter_name }}{}

func ({{ ffi_converter_name }}) Lower(value int64) C.int64_t {
	return C.int64_t(value)
}

func ({{ ffi_converter_name }}) Write(writer io.Writer, value int64) {
	writeInt64(writer, value)
}

func ({{ ffi_converter_name }}) Lift(value C.int64_t) int64 {
	return int64(value)
}

func ({{ ffi_converter_name }}) Read(reader io.Reader) int64 {
	return readInt64(reader)
}

type {{ type_|ffi_destroyer_name }} struct {}

func ({{ type_|ffi_destroyer_name }}) Destroy(_ {{ type_name }}) {}
