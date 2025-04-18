{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

type {{ ffi_converter_name }} struct{}

var {{ ffi_converter_instance }} = {{ ffi_converter_name }}{}

func ({{ ffi_converter_name }}) Lower(value uint8) C.uint8_t {
	return C.uint8_t(value)
}

func ({{ ffi_converter_name }}) Write(writer io.Writer, value uint8) {
	writeUint8(writer, value)
}

func ({{ ffi_converter_name }}) Lift(value C.uint8_t) uint8 {
	return uint8(value)
}

func ({{ ffi_converter_name }}) Read(reader io.Reader) uint8 {
	return readUint8(reader)
}

type {{ ffi_destroyer_name }} struct {}

func ({{ ffi_destroyer_name }}) Destroy(_ {{ type_name }}) {}
