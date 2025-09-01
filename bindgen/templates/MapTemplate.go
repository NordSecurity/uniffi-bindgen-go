{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

{%- let key_type_name = key_type|type_name(ci) %}
{%- let value_type_name = value_type|type_name(ci) %}

type {{ ffi_converter_name }} struct {}

var {{ ffi_converter_instance }} = {{ ffi_converter_name }}{}

func (c {{ ffi_converter_name }}) Lift(rb RustBufferI) {{ type_name }} {
	return LiftFromRustBuffer[{{ type_name }}](c, rb)
}

func (_ {{ ffi_converter_name }}) Read(reader io.Reader) {{ type_name }} {
	result := make({{ type_name }})
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := {{ key_type|read_fn }}(reader)
		value := {{ value_type|read_fn }}(reader)
		result[key] = value
	}
	return result
}

func (c {{ ffi_converter_name }}) Lower(value {{ type_name }}) C.RustBuffer {
	return LowerIntoRustBuffer[{{ type_name }}](c, value)
}

func (c {{ ffi_converter_name }}) LowerExternal(value {{ type_name }}) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[{{ type_name }}](c, value))
}

func (_ {{ ffi_converter_name }}) Write(writer io.Writer, mapValue {{ type_name }}) {
	if len(mapValue) > math.MaxInt32 {
		panic("{{ type_name }} is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		{{ key_type|write_fn }}(writer, key)
		{{ value_type|write_fn }}(writer, value)
	}
}

type {{ ffi_destroyer_name }} struct {}

func (_ {{ ffi_destroyer_name }}) Destroy(mapValue {{ type_name }}) {
	for key, value := range mapValue {
		{{ key_type|destroy_fn }}(key)
		{{ value_type|destroy_fn }}(value)	
	}
}
