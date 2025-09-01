{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

{{- self.add_import("math") }}

{%- let inner_type_name = inner_type|type_name(ci) %}

type {{ ffi_converter_name }} struct{}

var {{ ffi_converter_instance }} = {{ ffi_converter_name }}{}

func (c {{ ffi_converter_name }}) Lift(rb RustBufferI) {{ type_name }} {
	return LiftFromRustBuffer[{{ type_name }}](c, rb)
}

func (c {{ ffi_converter_name }}) Read(reader io.Reader) {{ type_name }} {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make({{type_name}}, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, {{ inner_type|read_fn }}(reader))
	}
	return result
}

func (c {{ ffi_converter_name }}) Lower(value {{ type_name }}) C.RustBuffer {
	return LowerIntoRustBuffer[{{ type_name }}](c, value)
}

func (c {{ ffi_converter_name }}) LowerExternal(value {{ type_name }}) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[{{ type_name }}](c, value))
}

func (c {{ ffi_converter_name }}) Write(writer io.Writer, value {{ type_name }}) {
	if len(value) > math.MaxInt32 {
		panic("{{ type_name }} is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		{{ inner_type|write_fn }}(writer, item)
	}
}

type {{ ffi_destroyer_name }} struct {}

func ({{ ffi_destroyer_name }}) Destroy(sequence {{ type_name }}) {
	for _, value := range sequence {
		{{ inner_type|destroy_fn }}(value)	
	}
}
