{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

type {{ ffi_converter_name }} struct{}

var {{ ffi_converter_instance }} = {{ ffi_converter_name }}{}

func ({{ ffi_converter_name }}) Lower(value float64) C.double {
	return C.double(value)
}

func ({{ ffi_converter_name }}) Write(writer io.Writer, value float64) {
	writeFloat64(writer, value)
}

func ({{ ffi_converter_name }}) Lift(value C.double) float64 {
	return float64(value)
}

func ({{ ffi_converter_name }}) Read(reader io.Reader) float64 {
	return readFloat64(reader)
}

type {{ ffi_destroyer_name }} struct {}

func ({{ ffi_destroyer_name }}) Destroy(_ {{ type_name }}) {}
