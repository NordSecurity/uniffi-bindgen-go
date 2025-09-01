{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

{%- call go::docstring(e, 0) %}
type {{ canonical_type_name }} struct {
	err error
}

// Convience method to turn *{{canonical_type_name}} into error
// Avoiding treating nil pointer as non nil error interface
func (err *{{ canonical_type_name }}) AsError() error {
	if err == nil {
		return nil
	} else {
		return err
	}
}

func (err {{ canonical_type_name }}) Error() string {
	return fmt.Sprintf("{{ canonical_type_name }}: %s", err.err.Error())
}

func (err {{ canonical_type_name }}) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
{%- for variant in e.variants() %}
{%- let variant_class_name = (canonical_type_name.clone() + variant.name())|class_name %}
var Err{{ variant_class_name }} = fmt.Errorf("{{ variant_class_name }}")
{%- endfor %}

// Variant structs
{%- for variant in e.variants() %}
{%- let variant_class_name = (canonical_type_name.clone() + variant.name())|class_name %}
{%- call go::docstring(variant, 0) %}
type {{ variant_class_name }} struct {
	{%- if e.is_flat() %}
	message string
	{%- else %}
	{%- for field in variant.fields() %}
	{{ field.name()|error_field_name|or_pos_field(loop.index0) }} {{ self.field_type_name(field, ci) }}
	{%- endfor %}
	{%- endif %}
}

{%- call go::docstring(variant, 0) %}
func New{{ variant_class_name }}(
	{%- if !e.is_flat() %}
	{%- for field in variant.fields() %}
	{{ field.name()|var_name|or_pos_var(loop.index0) }} {{ self.field_type_name(field, ci) }},
	{%- endfor %}
	{%- endif %}
) {{ type_name }} {
	return &{{ canonical_type_name }} { err: &{{ variant_class_name }} {
		{%- if !e.is_flat() %}
		{%- for field in variant.fields() %}
			{{ field.name()|error_field_name|or_pos_field(loop.index0) -}}
				: {{ field.name()|var_name|or_pos_var(loop.index0) }},
		{%- endfor -%}
		{%- endif -%}
		} }
}

func (e {{ variant_class_name }}) destroy() {
	{%- for field in variant.fields() %}
		{{ field|destroy_fn }}(e.{{ field.name()|error_field_name|or_pos_field(loop.index0) }})
	{%- endfor %}
}

{% if e.is_flat() %}
func (err {{ variant_class_name }}) Error() string {
	return fmt.Sprintf("{{ variant.name()|class_name }}: %s", err.message)
}
{%- else %}
func (err {{ variant_class_name }}) Error() string {
	return fmt.Sprint("{{ variant.name()|class_name }}",
		{% if !variant.fields().is_empty() %}": ",{% endif %}
		{%- for field in variant.fields() %}
		{% if !loop.first %}", ",{% endif %}
		"{{ field.name()|error_field_name|or_pos_field(loop.index0) }}=",
		err.{{ field.name()|error_field_name|or_pos_field(loop.index0) }},
		{%- endfor %}
	)
}
{%- endif %}

func (self {{ variant_class_name }}) Is(target error) bool {
	return target == Err{{ variant_class_name }}
}

{%- endfor %}

type {{ ffi_converter_name }} struct{}

var {{ ffi_converter_instance }} = {{ ffi_converter_name }}{}

func (c {{ ffi_converter_name }}) Lift(eb RustBufferI) {{ type_name }} {
	return LiftFromRustBuffer[{{ type_name }}](c, eb)
}

func (c {{ ffi_converter_name }}) Lower(value {{ type_name }}) C.RustBuffer {
	return LowerIntoRustBuffer[{{ type_name }}](c, value)
}

func (c {{ ffi_converter_name }}) LowerExternal(value {{ type_name }}) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[{{ type_name }}](c, value))
}

func (c {{ ffi_converter_name }}) Read(reader io.Reader) {{ type_name }} {
	errorID := readUint32(reader)

	{%- if e.is_flat() %}

	message := {{ Type::String.borrow()|read_fn }}(reader)
	switch errorID {
	{%- for variant in e.variants() %}
	case {{ loop.index }}:
		return &{{ canonical_type_name }}{ &{{- canonical_type_name }}{{ variant.name()|class_name }}{message}}
	{%- endfor %}
	default:
		panic(fmt.Sprintf("Unknown error code %d in {{ e|ffi_converter_name }}.Read()", errorID))
	}

	{% else %}

	switch errorID {
	{%- for variant in e.variants() %}
	case {{ loop.index }}:
		return &{{ canonical_type_name }}{ &{{- canonical_type_name }}{{ variant.name()|class_name }}{
			{%- for field in variant.fields() %}
			{{ field.name()|error_field_name|or_pos_field(loop.index0) }}: {{ field|read_fn }}(reader),
			{%- endfor %}
		}}
	{%- endfor %}
	default:
		panic(fmt.Sprintf("Unknown error code %d in {{ e|ffi_converter_name}}.Read()", errorID))
	}

	{%- endif %}
}

func (c {{ ffi_converter_name }}) Write(writer io.Writer, value {{ type_name }}) {
	switch variantValue := value.err.(type) {
		{%- for variant in e.variants() %}
		case *{{ canonical_type_name }}{{ variant.name()|class_name }}:
			writeInt32(writer, {{ loop.index }})
			{%- for field in variant.fields() %}
			{{ field|write_fn }}(writer, variantValue.{{ field.name()|error_field_name|or_pos_field(loop.index0) }})
			{%- endfor %}
		{%- endfor %}
		default:
			_ = variantValue
			panic(fmt.Sprintf("invalid error value `%v` in {{ e|ffi_converter_name }}.Write", value))
	}
}

type {{ ffi_destroyer_name }} struct {}

func (_ {{ ffi_destroyer_name }}) Destroy(value {{ type_name }}) {
	switch variantValue := value.err.(type) {
		{%- for variant in e.variants() %}
		case {{ canonical_type_name }}{{ variant.name()|class_name }}:
			variantValue.destroy()
		{%- endfor %}
		default:
			_ = variantValue
			panic(fmt.Sprintf("invalid error value `%v` in {{ ffi_destroyer_name }}.Destroy", value))
	}
}

