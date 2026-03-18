{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

{%- match config.custom_types.get(name.as_str())  %}
{%- when None %}
{#- Define the type using typealiases to the builtin #}
/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type {{ name }} = {{ builtin|type_name(ci) }}
type {{ ffi_converter_name }} = {{ builtin|ffi_converter_name(ci) }}
type {{ ffi_destroyer_name }} = {{ builtin|ffi_destroyer_name(ci) }}
var {{ ffi_converter_instance }} = {{ builtin|ffi_converter_name(ci) }}{}

{%- if !ci.is_external(builtin.as_ref()) %}
{%- match builtin.as_ref() %}
{%- when Type::Object { .. } | Type::CallbackInterface { .. } %}
{%- else %}
{%- if let FfiType::RustBuffer(_) = builtin.as_ref()|into_ffi_type %}
func LiftFromExternal{{ canonical_type_name }}(value ExternalCRustBuffer) {{ name }} {
	return {{ ffi_converter_instance }}.Lift(RustBufferFromExternal(value))
}

func LowerToExternal{{ canonical_type_name }}(value {{ name }}) ExternalCRustBuffer {
	return RustBufferFromC({{ ffi_converter_instance }}.Lower(value))
}
{%- else %}
func LiftFromExternal{{ canonical_type_name }}(value {{ builtin|type_name(ci) }}) {{ name }} {
	return {{ ffi_converter_instance }}.Lift({{ builtin.as_ref()|ffi_type_name }}(value))
}

func LowerToExternal{{ canonical_type_name }}(value {{ name }}) {{ builtin|type_name(ci) }} {
	return {{ builtin|type_name(ci) }}({{ ffi_converter_instance }}.Lower(value))
}
{%- endif %}
{%- endmatch %}
{%- endif %}

{%- when Some with (config) %}

{%- let ffi_type_name=builtin.as_ref()|ffi_type_name %}

{# When the config specifies a different type name, create a typealias for it #}
{%- match config.type_name %}
{%- when Some(concrete_type_name) %}
/**
 * Typealias from the type name used in the UDL file to the custom type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type {{ name }} = {{ concrete_type_name }}
{%- else %}
{%- endmatch %}

{%- match config.imports %}
{%- when Some(imports) %}
{%- for import_name in imports %}
{{ self.add_import(import_name) }}
{%- endfor %}
{%- else %}
{%- endmatch %}

type {{ ffi_converter_name }} struct{}

var {{ ffi_converter_instance }} = {{ ffi_converter_name }}{}

func ({{ ffi_converter_name }}) Lower(value {{ name }}) {{ ffi_type_name }} {
	builtinValue := {{ config.lower("value") }}
	ffiValue := {{ builtin|lower_fn(ci) }}(builtinValue)
	return {% call go::remap_ffi_val(builtin, "ffiValue") %}
}

func ({{ ffi_converter_name }}) Write(writer io.Writer, value {{ name }}) {
	builtinValue := {{ config.lower("value") }}
	{{ builtin|write_fn(ci) }}(writer, builtinValue)
}

func ({{ ffi_converter_name }}) Lift(value {{ ffi_type_name }}) {{ name }} {
	builtinValue := {{ builtin|lift_fn(ci) }}(value)
	{{ config.lift("builtinValue") }}
}

{%- match builtin.as_ref() %}
{%- when Type::Object { .. } | Type::CallbackInterface { .. } %}
{%- else %}
{%- if let FfiType::RustBuffer(_) = builtin.as_ref()|into_ffi_type %}
func LiftFromExternal{{ canonical_type_name }}(value ExternalCRustBuffer) {{ name }} {
	return {{ ffi_converter_instance }}.Lift(RustBufferFromExternal(value))
}

func LowerToExternal{{ canonical_type_name }}(value {{ name }}) ExternalCRustBuffer {
	return {{ ffi_converter_instance }}.Lower(value)
}
{%- else %}
func LiftFromExternal{{ canonical_type_name }}(value {{ builtin|type_name(ci) }}) {{ name }} {
	return {{ ffi_converter_instance }}.Lift({{ ffi_type_name }}(value))
}

func LowerToExternal{{ canonical_type_name }}(value {{ name }}) {{ builtin|type_name(ci) }} {
	return {{ builtin|type_name(ci) }}({{ ffi_converter_instance }}.Lower(value))
}
{%- endif %}
{%- endmatch %}

func ({{ ffi_converter_name }}) Read(reader io.Reader) {{ name }} {
	builtinValue := {{ builtin|read_fn(ci) }}(reader)
	{{ config.lift("builtinValue") }}
}

type {{ ffi_destroyer_name }} struct {}

func ({{ ffi_destroyer_name }}) Destroy(value {{ name }}) {
	builtinValue := {{ config.lower("value") }}
	{{ builtin|destroy_fn(ci) }}(builtinValue)
}

{%- endmatch %}
