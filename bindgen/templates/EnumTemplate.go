{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

{% let e = ci.get_enum_definition(name).unwrap() -%}
{%- if e.is_flat() -%}

type {{ type_name }} uint

const (
	{%- for variant in e.variants() %}
	{{ type_name }}{{ variant.name()|enum_variant_name }} {{ type_name }} = {{ loop.index }}
	{%- endfor %}
)
{%- else %}

type {{ type_name }} interface {
	Destroy()
}

{%- for variant in e.variants() %}
type {{ type_name }}{{ variant.name()|class_name }} struct {
	{%- for field in variant.fields() %}
	{{ field.name()|field_name }} {{ field|type_name}}
	{%- endfor %}
}

func (e {{ type_name }}{{ variant.name()|class_name }}) Destroy() {
	{%- for field in variant.fields() %}
		{{ field|destroy_fn }}(e.{{ field.name()|field_name }});
	{%- endfor %}
}
{%- endfor %}

{%- endif %}

type {{ e|ffi_converter_name}} struct {}

var {{ e|ffi_converter_name }}INSTANCE = {{ e|ffi_converter_name }}{}

func (c {{ e|ffi_converter_name }}) lift(cRustBuf C.RustBuffer) {{ type_name }} {
	return liftFromRustBuffer[{{ type_name }}](c, fromCRustBuffer(cRustBuf))
}

func (c {{ e|ffi_converter_name }}) lower(value {{ type_name }}) C.RustBuffer {
	return lowerIntoRustBuffer[{{ type_name }}](c, value)
}

{%- if e.is_flat() %}
func ({{ e|ffi_converter_name }}) read(reader io.Reader) {{ type_name }} {
	id := readInt32(reader)
	return {{ type_name }}(id)
}

func ({{ e|ffi_converter_name }}) write(writer io.Writer, value {{ type_name }}) {
	writeInt32(writer, int32(value))
}
{%- else %}
func ({{ e|ffi_converter_name }}) read(reader io.Reader) {{ type_name }} {
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
			panic(fmt.Sprintf("invalid enum value %v in {{ e|ffi_converter_name }}.read()", id));
	}
}

func ({{ e|ffi_converter_name }}) write(writer io.Writer, value {{ type_name }}) {
	switch variant_value := value.(type) {
		{%- for variant in e.variants() %}
		case {{ type_name }}{{ variant.name()|class_name }}:
			writeInt32(writer, {{ loop.index }})
			{%- for field in variant.fields() %}
			{{ field|write_fn }}(writer, variant_value.{{ field.name()|field_name }})
			{%- endfor %}
		{%- endfor %}
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in {{ e|ffi_converter_name }}.write", value))
	}
}
{%- endif %}

type {{ ffi_destroyer_name }} struct {}

func (_ {{ ffi_destroyer_name }}) destroy(value {{ type_name }}) {
	{%- if e.is_flat() %}
	{%- else %}
	value.Destroy()
	{%- endif %}
}
