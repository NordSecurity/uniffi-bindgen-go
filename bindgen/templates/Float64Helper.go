{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

type {{ ffi_converter_name }} struct{}

var {{ ffi_converter_name }}INSTANCE = {{ ffi_converter_name }}{}

func ({{ ffi_converter_name }}) lower(value float64) C.double {
	return C.double(value)
}

func ({{ ffi_converter_name }}) write(writer io.Writer, value float64) {
	writeFloat64(writer, value)
}

func ({{ ffi_converter_name }}) lift(value C.double) float64 {
	return float64(value)
}

func ({{ ffi_converter_name }}) read(reader io.Reader) float64 {
	return readFloat64(reader)
}

type {{ type_|ffi_destroyer_name }} struct {}

func ({{ type_|ffi_destroyer_name }}) destroy(_ {{ type_name }}) {}
