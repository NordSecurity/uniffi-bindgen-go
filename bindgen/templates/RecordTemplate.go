{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

{%- let rec = ci.get_record_definition(name).expect("missing record") %}

{%- call go::docstring(rec, 0) %}
type {{ type_name }} struct {
	{%- for field in rec.fields() %}
	{%- call go::docstring(field, 1) %}
	{{ field.name()|field_name }} {{ field|type_name(ci) -}}
	{%- endfor %}
}

func (r *{{ type_name }}) Destroy() {
	{%- for field in rec.fields() %}
		{{ field|destroy_fn }}(r.{{ field.name()|field_name }});
	{%- endfor %}
}

type {{ ffi_converter_name }} struct {}

var {{ ffi_converter_instance }} = {{ ffi_converter_name }}{}

func (c {{ rec|ffi_converter_name }}) Lift(rb RustBufferI) {{ type_name }} {
	return LiftFromRustBuffer[{{ type_name }}](c, rb)
}

func (c {{ rec|ffi_converter_name }}) Read(reader io.Reader) {{ type_name }} {
	return {{ type_name }} {
		{%- for field in rec.fields() %}
			{{ field|read_fn }}(reader),
		{%- endfor %}
	}
}

func (c {{ rec|ffi_converter_name }}) Lower(value {{ type_name }}) C.RustBuffer {
	return LowerIntoRustBuffer[{{ type_name }}](c, value)
}

func (c {{ rec|ffi_converter_name }}) LowerExternal(value {{ type_name }}) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[{{ type_name }}](c, value))
}

func (c {{ rec|ffi_converter_name }}) Write(writer io.Writer, value {{ type_name }}) {
	{%- for field in rec.fields() %}
		{{ field|write_fn }}(writer, value.{{ field.name()|field_name }});
	{%- endfor %}
}

type {{ ffi_destroyer_name }} struct {}

func (_ {{ ffi_destroyer_name }}) Destroy(value {{ type_name }}) {
	value.Destroy()
}
