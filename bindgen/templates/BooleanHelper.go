{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

type {{ ffi_converter_name }} struct{}

var {{ ffi_converter_name }}INSTANCE = {{ ffi_converter_name }}{}

func ({{ ffi_converter_name }}) Lower(value bool) C.int8_t {
	if value {
		return C.int8_t(1)
	}
	return C.int8_t(0)
}

func ({{ ffi_converter_name }}) Write(writer io.Writer, value bool) {
	if value {
		writeInt8(writer, 1)
	} else {
		writeInt8(writer, 0)
	}
}

func ({{ ffi_converter_name }}) Lift(value C.int8_t) bool {
	return value != 0
}

func ({{ ffi_converter_name }}) Read(reader io.Reader) bool {
	return readInt8(reader) != 0
}

type {{ type_|ffi_destroyer_name }} struct {}

func ({{ type_|ffi_destroyer_name }}) Destroy(_ {{ type_name }}) {}
