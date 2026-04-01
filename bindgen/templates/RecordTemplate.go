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
		{{ field|destroy_fn(ci) }}(r.{{ field.name()|field_name }});
	{%- endfor %}
}

{%- let trait_methods = rec.uniffi_trait_methods() %}
{%- let receiver_type = type_name %}
{%- let self_binding = "_selfBuf" %}
{%- include "TraitMethods.go" %}

{%- for meth in rec.methods() %}
{%- call go::docstring(meth, 0) %}
func (_self {{ type_name }}) {{ meth.name()|fn_name }}({%- call go::arg_list_decl(meth) -%}) {% call go::return_type_decl(meth) %} {
	_selfBuf := {{ ffi_converter_instance }}.Lower(_self)
	{% if meth.is_async() %}
	{% call go::async_ffi_call_binding(meth, "_selfBuf") %}
	{% else %}
	{% call go::ffi_call_binding(meth, "_selfBuf") %}
	{% endif %}
}

{%- endfor %}

type {{ ffi_converter_name }} struct {}

var {{ ffi_converter_instance }} = {{ ffi_converter_name }}{}

func (c {{ rec|ffi_converter_name(ci) }}) Lift(rb RustBufferI) {{ type_name }} {
	return LiftFromRustBuffer[{{ type_name }}](c, rb)
}

func (c {{ rec|ffi_converter_name(ci) }}) Read(reader io.Reader) {{ type_name }} {
	return {{ type_name }} {
		{%- for field in rec.fields() %}
			{{ field|read_fn(ci) }}(reader),
		{%- endfor %}
	}
}

func (c {{ rec|ffi_converter_name(ci) }}) Lower(value {{ type_name }}) C.RustBuffer {
	return LowerIntoRustBuffer[{{ type_name }}](c, value)
}

func (c {{ rec|ffi_converter_name(ci) }}) LowerExternal(value {{ type_name }}) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[{{ type_name }}](c, value))
}

func (c {{ rec|ffi_converter_name(ci) }}) Write(writer io.Writer, value {{ type_name }}) {
	{%- for field in rec.fields() %}
		{{ field|write_fn(ci) }}(writer, value.{{ field.name()|field_name }});
	{%- endfor %}
}

type {{ ffi_destroyer_name }} struct {}

func (_ {{ ffi_destroyer_name }}) Destroy(value {{ type_name }}) {
	value.Destroy()
}
