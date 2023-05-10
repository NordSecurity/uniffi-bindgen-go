{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

type {{ ffi_converter_name }} struct{}

var {{ ffi_converter_name }}INSTANCE = {{ ffi_converter_name }}{}

func ({{ ffi_converter_name }}) lower(value uint64) C.uint64_t {
	return C.uint64_t(value)
}

func ({{ ffi_converter_name }}) write(writer io.Writer, value uint64) {
	writeUint64(writer, value)
}

func ({{ ffi_converter_name }}) lift(value C.uint64_t) uint64 {
	return uint64(value)
}

func ({{ ffi_converter_name }}) read(reader io.Reader) uint64 {
	return readUint64(reader)
}

type {{ type_|ffi_destroyer_name }} struct {}

func ({{ type_|ffi_destroyer_name }}) destroy(_ {{ type_name }}) {}
