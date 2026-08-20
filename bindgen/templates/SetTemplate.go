{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

{{- self.add_import("math") }}

{%- let inner_type_name = inner_type|type_name(ci) %}

type {{ ffi_converter_name }} struct {}

var {{ ffi_converter_instance }} = {{ ffi_converter_name }}{}

func (c {{ ffi_converter_name }}) Lift(rb RustBufferI) {{ type_name }} {
	return LiftFromRustBuffer[{{ type_name }}](c, rb)
}

func (_ {{ ffi_converter_name }}) Read(reader io.Reader) {{ type_name }} {
	length := readInt32(reader)
	result := make({{ type_name }}, length)
	for i := int32(0); i < length; i++ {
		result[{{ inner_type|read_fn(ci) }}(reader)] = struct{}{}
	}
	return result
}

func (c {{ ffi_converter_name }}) Lower(value {{ type_name }}) C.RustBuffer {
	return LowerIntoRustBuffer[{{ type_name }}](c, value)
}

func (c {{ ffi_converter_name }}) LowerExternal(value {{ type_name }}) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[{{ type_name }}](c, value))
}

func (_ {{ ffi_converter_name }}) Write(writer io.Writer, setValue {{ type_name }}) {
	if len(setValue) > math.MaxInt32 {
		panic("{{ type_name }} is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(setValue)))
	for value := range setValue {
		{{ inner_type|write_fn(ci) }}(writer, value)
	}
}

type {{ ffi_destroyer_name }} struct {}

func (_ {{ ffi_destroyer_name }}) Destroy(setValue {{ type_name }}) {
	for value := range setValue {
		{{ inner_type|destroy_fn(ci) }}(value)
	}
}
