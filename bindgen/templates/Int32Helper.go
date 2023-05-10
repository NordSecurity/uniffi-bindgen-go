{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

type {{ ffi_converter_name }} struct{}

var {{ ffi_converter_name }}INSTANCE = {{ ffi_converter_name }}{}

func ({{ ffi_converter_name }}) lower(value int32) C.int32_t {
	return C.int32_t(value)
}

func ({{ ffi_converter_name }}) write(writer io.Writer, value int32) {
	writeInt32(writer, value)
}

func ({{ ffi_converter_name }}) lift(value C.int32_t) int32 {
	return int32(value)
}

func ({{ ffi_converter_name }}) read(reader io.Reader) int32 {
	return readInt32(reader)
}

type {{ type_|ffi_destroyer_name }} struct {}

func ({{ type_|ffi_destroyer_name }}) destroy(_ {{ type_name }}) {}
