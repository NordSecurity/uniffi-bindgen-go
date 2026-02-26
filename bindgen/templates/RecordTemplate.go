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
{%- if let Some(display_fmt) = trait_methods.display_fmt %}
func (_self {{ type_name }}) String() string {
	_selfBuf := {{ ffi_converter_instance }}.Lower(_self)
	{% call go::ffi_call_binding(display_fmt, "_selfBuf") %}
}

{%- endif %}
{%- if let Some(debug_fmt) = trait_methods.debug_fmt %}
func (_self {{ type_name }}) DebugString() string {
	_selfBuf := {{ ffi_converter_instance }}.Lower(_self)
	{% call go::ffi_call_binding(debug_fmt, "_selfBuf") %}
}

{%- endif %}
{%- if let Some(eq_eq) = trait_methods.eq_eq %}
func (_self {{ type_name }}) Eq(other {{ type_name }}) bool {
	_selfBuf := {{ ffi_converter_instance }}.Lower(_self)
	{% call go::ffi_call_binding(eq_eq, "_selfBuf") %}
}

{%- endif %}
{%- if let Some(eq_ne) = trait_methods.eq_ne %}
func (_self {{ type_name }}) Ne(other {{ type_name }}) bool {
	_selfBuf := {{ ffi_converter_instance }}.Lower(_self)
	{% call go::ffi_call_binding(eq_ne, "_selfBuf") %}
}

{%- endif %}
{%- if let Some(hash_hash) = trait_methods.hash_hash %}
func (_self {{ type_name }}) Hash() uint64 {
	_selfBuf := {{ ffi_converter_instance }}.Lower(_self)
	{% call go::ffi_call_binding(hash_hash, "_selfBuf") %}
}

{%- endif %}
{%- if let Some(ord_cmp) = trait_methods.ord_cmp %}
func (_self {{ type_name }}) Cmp(other {{ type_name }}) int8 {
	_selfBuf := {{ ffi_converter_instance }}.Lower(_self)
	{% call go::ffi_call_binding(ord_cmp, "_selfBuf") %}
}

{%- endif %}

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
