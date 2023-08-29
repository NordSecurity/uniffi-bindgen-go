{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

{%- let e = ci.get_error_definition(name).unwrap() -%}
type {{ type_name|class_name }} struct {
	err error
}

func (err {{ type_name|class_name }}) Error() string {
	return fmt.Sprintf("{{ type_name|class_name }}: %s", err.err.Error())
}

func (err {{ type_name|class_name }}) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
{%- for variant in e.variants() %}
{%- let variant_class_name = (type_name.clone() + variant.name())|class_name %}
var Err{{ variant_class_name }} = fmt.Errorf("{{ variant_class_name }}")
{%- endfor %}

// Variant structs
{%- for variant in e.variants() %}
{%- let variant_class_name = (type_name.clone() + variant.name())|class_name %}
type {{ variant_class_name }} struct {
	{%- if e.is_flat() %}
	message string
	{%- else %}
	{%- for field in variant.fields() %}
	{{ field.name()|field_name }} {{ field|type_name}}
	{%- endfor %}
	{%- endif %}
}

func New{{ variant_class_name }}(
	{%- if !e.is_flat() %}
	{%- for field in variant.fields() %}
	{{ field.name()|var_name }} {{ field|type_name}},
	{%- endfor %}
	{%- endif %}
) *{{ type_name.clone() }} {
	return &{{ type_name.clone() }}{
		err: &{{ variant_class_name }}{
		{%- if !e.is_flat() %}
		{%- for field in variant.fields() %}
			{{ field.name()|field_name }}: {{ field.name()|var_name }},
		{%- endfor %}
		{%- endif %}
		},
	}
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
		"{{ field.name()|field_name }}=",
		err.{{ field.name()|field_name }},
		{%- endfor %}
	)
}
{%- endif %}

func (self {{ variant_class_name }}) Is(target error) bool {
	return target == Err{{ variant_class_name }}
}

{%- endfor %}

type {{ e|ffi_converter_name }} struct{}

var {{ e|ffi_converter_name }}INSTANCE = {{ e|ffi_converter_name }}{}

func (c {{ e|ffi_converter_name }}) lift(cErrBuf C.RustBuffer) error {
	errBuf := fromCRustBuffer(cErrBuf)
	return liftFromRustBuffer[error](c, errBuf)
}

func (c {{ e|ffi_converter_name }}) lower(value *{{ type_name|class_name }}) C.RustBuffer {
	return lowerIntoRustBuffer[*{{ type_name|class_name }}](c, value)
}

func (c {{ e|ffi_converter_name }}) read(reader io.Reader) error {
	errorID := readUint32(reader)

	{%- if e.is_flat() %}

	message := {{ TypeIdentifier::String.borrow()|read_fn }}(reader)
	switch errorID {
	{%- for variant in e.variants() %}
	case {{ loop.index }}:
		return &{{ type_name|class_name }}{&{{ type_name|class_name }}{{ variant.name()|class_name }}{message}}
	{%- endfor %}
	default:
		panic(fmt.Sprintf("Unknown error code %d in {{ e|ffi_converter_name}}.read()", errorID))
	}

	{% else %}

	switch errorID {
	{%- for variant in e.variants() %}
	case {{ loop.index }}:
		return &{{ type_name|class_name }}{&{{ type_name|class_name }}{{ variant.name()|class_name }}{
			{%- for field in variant.fields() %}
			{{ field.name()|field_name }}: {{ field|read_fn }}(reader),
			{%- endfor %}
		}}
	{%- endfor %}
	default:
		panic(fmt.Sprintf("Unknown error code %d in {{ e|ffi_converter_name}}.read()", errorID))
	}

	{%- endif %}
}

func (c {{ e|ffi_converter_name }}) write(writer io.Writer, value *{{ type_name|class_name }}) {
	switch variantValue := value.err.(type) {
		{%- for variant in e.variants() %}
		case *{{ type_name }}{{ variant.name()|class_name }}:
			writeInt32(writer, {{ loop.index }})
			{%- for field in variant.fields() %}
			{{ field|write_fn }}(writer, variantValue.{{ field.name()|field_name }})
			{%- endfor %}
		{%- endfor %}
		default:
			_ = variantValue
			panic(fmt.Sprintf("invalid error value `%v` in {{ e|ffi_converter_name }}.write", value))
	}
}
