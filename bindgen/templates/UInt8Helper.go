{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

type {{ ffi_converter_name }} struct{}

var {{ ffi_converter_name }}INSTANCE = {{ ffi_converter_name }}{}

func ({{ ffi_converter_name }}) lower(value uint8) C.uint8_t {
	return C.uint8_t(value)
}

func ({{ ffi_converter_name }}) write(writer io.Writer, value uint8) {
	writeUint8(writer, value)
}

func ({{ ffi_converter_name }}) lift(value C.uint8_t) uint8 {
	return uint8(value)
}

func ({{ ffi_converter_name }}) read(reader io.Reader) uint8 {
	return readUint8(reader)
}

type {{ type_|ffi_destroyer_name }} struct {}

func ({{ type_|ffi_destroyer_name }}) destroy(_ {{ type_name }}) {}
