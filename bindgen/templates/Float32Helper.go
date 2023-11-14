{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

type {{ ffi_converter_name }} struct{}

var {{ ffi_converter_name }}INSTANCE = {{ ffi_converter_name }}{}

func ({{ ffi_converter_name }}) Lower(value float32) C.float {
	return C.float(value)
}

func ({{ ffi_converter_name }}) Write(writer io.Writer, value float32) {
	writeFloat32(writer, value)
}

func ({{ ffi_converter_name }}) Lift(value C.float) float32 {
	return float32(value)
}

func ({{ ffi_converter_name }}) Read(reader io.Reader) float32 {
	return readFloat32(reader)
}

type {{ type_|ffi_destroyer_name }} struct {}

func ({{ type_|ffi_destroyer_name }}) Destroy(_ {{ type_name }}) {}
