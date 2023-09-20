{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

type {{ ffi_converter_name }} struct{}

var {{ ffi_converter_name }}INSTANCE = {{ ffi_converter_name }}{}

func ({{ ffi_converter_name }}) Lower(value uint16) C.uint16_t {
	return C.uint16_t(value)
}

func ({{ ffi_converter_name }}) Write(writer io.Writer, value uint16) {
	writeUint16(writer, value)
}

func ({{ ffi_converter_name }}) Lift(value C.uint16_t) uint16 {
	return uint16(value)
}

func ({{ ffi_converter_name }}) Read(reader io.Reader) uint16 {
	return readUint16(reader)
}

type {{ type_|ffi_destroyer_name }} struct {}

func ({{ type_|ffi_destroyer_name }}) Destroy(_ {{ type_name }}) {}
