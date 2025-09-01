{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

{% let e = ci.get_enum_definition(name).expect("missing enum") -%}
{%- if e.is_flat() -%}

{%- call go::docstring(e, 0) %}
{%- if let Some(variant_discr_type) = e.variant_discr_type() %}
type {{ type_name }} {{ variant_discr_type|type_name(ci) }}

const (
	{%- for variant in e.variants() %}
	{%- call go::docstring(variant, 1) %}
	{{ type_name }}{{ variant.name()|enum_variant_name }} {{ type_name }} = {{ e|variant_discr_literal(loop.index0) }}
	{%- endfor %}
)
{%- else %}
type {{ type_name }} uint

const (
	{%- for variant in e.variants() %}
	{%- call go::docstring(variant, 1) %}
	{{ type_name }}{{ variant.name()|enum_variant_name }} {{ type_name }} = {{ loop.index }}
	{%- endfor %}
)
{%- endif %}

{%- else %}

{%- call go::docstring(e, 0) %}
type {{ type_name }} interface {
	Destroy()
}

{%- for variant in e.variants() %}
{%- call go::docstring(variant, 0) %}
type {{ type_name }}{{ variant.name()|class_name }} struct {
	{%- for field in variant.fields() %}
	{{ field.name()|field_name|or_pos_field(loop.index0) }} {{ field|type_name(ci) }}
	{%- endfor %}
}

func (e {{ type_name }}{{ variant.name()|class_name }}) Destroy() {
	{%- for field in variant.fields() %}
		{{ field|destroy_fn }}(e.{{ field.name()|field_name|or_pos_field(loop.index0) }});
	{%- endfor %}
}
{%- endfor %}

{%- endif %}

type {{ ffi_converter_name }} struct {}

var {{ ffi_converter_instance }} = {{ ffi_converter_name }}{}

func (c {{ ffi_converter_name }}) Lift(rb RustBufferI) {{ type_name }} {
	return LiftFromRustBuffer[{{ type_name }}](c, rb)
}

func (c {{ ffi_converter_name }}) Lower(value {{ type_name }}) C.RustBuffer {
	return LowerIntoRustBuffer[{{ type_name }}](c, value)
}

func (c {{ ffi_converter_name }}) LowerExternal(value {{ type_name }}) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[{{ type_name }}](c, value))
}

{%- if e.is_flat() %}
func ({{ ffi_converter_name }}) Read(reader io.Reader) {{ type_name }} {
	id := readInt32(reader)
	return {{ type_name }}(id)
}

func ({{ ffi_converter_name }}) Write(writer io.Writer, value {{ type_name }}) {
	writeInt32(writer, int32(value))
}
{%- else %}
func ({{ ffi_converter_name }}) Read(reader io.Reader) {{ type_name }} {
	id := readInt32(reader)
	switch (id) {
		{%- for variant in e.variants() %}
		case {{ loop.index }}:
			return {{ type_name }}{{ variant.name()|class_name }}{
				{%- for field in variant.fields() %}
				{{ field|read_fn }}(reader),
				{%- endfor %}
			};
		{%- endfor %}
		default:
			panic(fmt.Sprintf("invalid enum value %v in {{ ffi_converter_name }}.Read()", id));
	}
}

func ({{ ffi_converter_name }}) Write(writer io.Writer, value {{ type_name }}) {
	switch variant_value := value.(type) {
		{%- for variant in e.variants() %}
		case {{ type_name }}{{ variant.name()|class_name }}:
			writeInt32(writer, {{ loop.index }})
			{%- for field in variant.fields() %}
			{{ field|write_fn }}(writer, variant_value.{{ field.name()|field_name|or_pos_field(loop.index0) }})
			{%- endfor %}
		{%- endfor %}
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in {{ ffi_converter_name }}.Write", value))
	}
}
{%- endif %}

type {{ ffi_destroyer_name }} struct {}

func (_ {{ ffi_destroyer_name }}) Destroy(value {{ type_name }}) {
	{%- if e.is_flat() %}
	{%- else %}
	value.Destroy()
	{%- endif %}
}
