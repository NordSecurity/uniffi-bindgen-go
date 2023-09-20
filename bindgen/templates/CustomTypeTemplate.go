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
type {{ name }} = {{ builtin|type_name }}
type {{ ffi_converter_name }} = {{ builtin|ffi_converter_name }}
type {{ ffi_destroyer_name }} = {{ builtin|ffi_destroyer_name }}
var {{ ffi_converter_name }}INSTANCE = {{ builtin|ffi_converter_name }}{}

{%- when Some with (config) %}

{%- let ffi_type_tmp=builtin|into_ffi_type %}
{%- let ffi_type_name=ffi_type_tmp.borrow()|ffi_type_name %}

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

var {{ ffi_converter_name }}INSTANCE = {{ ffi_converter_name }}{}

func ({{ ffi_converter_name }}) Lower(value {{ name }}) {{ ffi_type_name }} {
    builtinValue := {{ config.from_custom.render("value") }}
    return {{ builtin|lower_fn }}(builtinValue)
}

func ({{ ffi_converter_name }}) Write(writer io.Writer, value {{ name }}) {
    builtinValue := {{ config.from_custom.render("value") }}
    {{ builtin|write_fn }}(writer, builtinValue)
}

func ({{ ffi_converter_name }}) Lift(value {{ ffi_type_name }}) {{ name }} {
    builtinValue := {{ builtin|lift_fn }}(value)
    {{ config.into_custom.render("builtinValue") }}
}

func ({{ ffi_converter_name }}) Read(reader io.Reader) {{ name }} {
    builtinValue := {{ builtin|read_fn }}(reader)
    {{ config.into_custom.render("builtinValue") }}
}

type {{ type_|ffi_destroyer_name }} struct {}

func ({{ type_|ffi_destroyer_name }}) Destroy(value {{ name }}) {
	builtinValue := {{ config.from_custom.render("value") }}
	{{ builtin|destroy_fn }}(builtinValue)
}

{%- endmatch %}
