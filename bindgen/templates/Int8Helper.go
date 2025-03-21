{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

type {{ ffi_converter_name }} struct{}

var {{ ffi_converter_instance }} = {{ ffi_converter_name }}{}

func ({{ ffi_converter_name }}) Lower(value int8) C.int8_t {
	return C.int8_t(value)
}

func ({{ ffi_converter_name }}) Write(writer io.Writer, value int8) {
	writeInt8(writer, value)
}

func ({{ ffi_converter_name }}) Lift(value C.int8_t) int8 {
	return int8(value)
}

func ({{ ffi_converter_name }}) Read(reader io.Reader) int8 {
	return readInt8(reader)
}

type {{ ffi_destroyer_name }} struct {}

func ({{ ffi_destroyer_name }}) Destroy(_ {{ type_name }}) {}
