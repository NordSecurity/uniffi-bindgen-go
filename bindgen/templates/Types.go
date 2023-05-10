{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

{%- import "macros.go" as go %}

{%- for type_ in ci.iter_types() %}
{%- let type_name = type_|type_name %}
{%- let ffi_converter_name = type_|ffi_converter_name %}
{%- let ffi_destroyer_name = type_|ffi_destroyer_name %}
{%- let canonical_type_name = type_|canonical_name %}
{#
 # Map `Type` instances to an include statement for that type.
 #
 # There is a companion match in `KotlinCodeOracle::create_code_type()` which performs a similar function for the
 # Rust code.
 #
 #   - When adding additional types here, make sure to also add a match arm to that function.
 #   - To keep things managable, let's try to limit ourselves to these 2 mega-matches
 #}
{%- match type_ %}

{%- when Type::Boolean %}
{%- include "BooleanHelper.go" %}

{%- when Type::Int8 %}
{%- include "Int8Helper.go" %}

{%- when Type::Int16 %}
{%- include "Int16Helper.go" %}

{%- when Type::Int32 %}
{%- include "Int32Helper.go" %}

{%- when Type::Int64 %}
{%- include "Int64Helper.go" %}

{%- when Type::UInt8 %}
{%- include "UInt8Helper.go" %}

{%- when Type::UInt16 %}
{%- include "UInt16Helper.go" %}

{%- when Type::UInt32 %}
{%- include "UInt32Helper.go" %}

{%- when Type::UInt64 %}
{%- include "UInt64Helper.go" %}

{%- when Type::Float32 %}
{%- include "Float32Helper.go" %}

{%- when Type::Float64 %}
{%- include "Float64Helper.go" %}

{%- when Type::String %}
{%- include "StringHelper.go" %}

{%- when Type::Error(name) %}
{% include "ErrorTemplate.go" %}

{%- when Type::Timestamp %}
{% include "TimestampHelper.go" %}

{%- when Type::Duration %}
{% include "DurationHelper.go" %}

{%- when Type::Enum(name) %}
{% include "EnumTemplate.go" %}

{%- when Type::Optional(inner_type) %}
{% include "OptionalTemplate.go" %}

{%- when Type::Object(name) %}
{% include "ObjectTemplate.go" %}

{%- when Type::Record(name) %}
{% include "RecordTemplate.go" %}

{%- when Type::Sequence(inner_type) %}
{% include "SequenceTemplate.go" %}

{%- when Type::Map(key_type, value_type) %}
{% include "MapTemplate.go" %}

{%- when Type::CallbackInterface(name) %}
{% include "CallbackInterfaceTemplate.go" %}

{%- when Type::Custom { name, builtin } %}
{% include "CustomTypeTemplate.go" %}

{%- else %}
{%- endmatch %}
{%- endfor %}
