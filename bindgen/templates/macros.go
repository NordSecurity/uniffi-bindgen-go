{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

{% macro arg_list_decl(func) %}
	{%- for arg in func.arguments() -%}
	    {%- let type_ = arg.as_type() -%}
	    {{ arg.name()|var_name }} {{ arg|type_name(ci) }}
		{%- if !loop.last %}, {% endif -%}
	{%- endfor -%}
{%- endmacro %}

{% macro return_type_decl(func) %}
	{%- match func.return_type() -%}
	{%- when Some with (return_type) -%}
		{%- match func.throws_type() -%}
		{%- when Some with (throws_type) -%}
		({{ return_type|type_name(ci) }}, {{ throws_type|type_name(ci) }})
		{%- when None -%}
		{{ return_type|type_name(ci) }}
		{%- endmatch %}
	{%- when None -%}
		{%- match func.throws_type() -%}
		{%- when Some with (throws_type) -%}
		{{ throws_type|type_name(ci) }}
		{%- when None -%}
		{%- endmatch %}
	{%- endmatch %}
{%- endmacro %}

{% macro return_type_decl_cb(func) %}
	{%- match func.return_type() -%}
	{%- when Some with (return_type) -%}
		{%- match func.throws_type() -%}
		{%- when Some with (throws_type) -%}
		({{ return_type|type_name(ci) }}, {{ throws_type|type_name(ci) }})
		{%- when None -%}
		{{ return_type|type_name(ci) }}
		{%- endmatch %}
	{%- when None -%}
		{%- match func.throws_type() -%}
		{%- when Some with (throws_type) -%}
		{{ throws_type|type_name(ci) }}
		{%- when None -%}
		{%- endmatch %}
	{%- endmatch %}
{%- endmacro %}


{% macro return_type_decl_async(func) %}
	{%- match func.return_type() -%}
	{%- when Some with (return_type) -%}
		{{ return_type|ffi_type_name }}
        {%- when None -%}
	{%- endmatch %}
{%- endmacro %}

{% macro ffi_call_binding(func, prefix) %}
	{%- match func.return_type() -%}
	{%- when Some with (return_type) -%}
		{%- match func.throws_type() -%}
		{%- when Some with (throws_type) -%}
		_uniffiRV, _uniffiErr := {% call to_ffi_call(func, prefix) %}
		if _uniffiErr != nil {
			var _uniffiDefaultValue {{ return_type|type_name(ci) }}
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return {{ return_type|lift_fn }}(_uniffiRV), _uniffiErr
		}
		{%- when None -%}
		return {{ return_type|lift_fn }}({% call to_ffi_call(func, prefix) %})
		{%- endmatch -%}
	{%- when None -%}
		{%- match func.throws_type() -%}
		{%- when Some with (throws_type) -%}
		_, _uniffiErr := {% call to_ffi_call(func, prefix) %}
		return _uniffiErr
		{%- when None -%}
		{% call to_ffi_call(func, prefix) %}
		{%- endmatch -%}
	{%- endmatch -%}
{% endmacro %}

{%- macro to_ffi_call(func, prefix) -%}
	{%- match func.throws_type() %}
	{%- when Some with (e) -%}
	rustCallWithError[{{ e|canonical_name }}]({{ e|ffi_converter_name }}{},
	{%- else -%}
	rustCall(
	{%- endmatch %}
	{%- match func.return_type() -%}
	{%- when Some with (return_type) -%}
	func(_uniffiStatus *C.RustCallStatus) {{ return_type|ffi_type_name }} {
		return C.{{ func.ffi_func().name() }}({% call _arg_list_ffi_call(func, prefix) -%})
	})
	{%- else -%}
	func(_uniffiStatus *C.RustCallStatus) bool {
		C.{{ func.ffi_func().name() }}({% call _arg_list_ffi_call(func, prefix) -%})
		return false
	})
	{%- endmatch %}
{%- endmacro -%}

{%- macro _arg_list_ffi_call(func, prefix) %}
	{%- if !prefix.is_empty() %}
		{{ prefix }},
	{%- endif %}
	{%- for arg in func.arguments() %}
		{%- call lower_fn_call(arg) -%}
		{%- if !loop.last %}, {% endif %}
	{%- endfor %}
	{%- if func.ffi_func().has_rust_call_status_arg() -%}
	{%- if func.arguments().len() > 0 %},{% endif -%}
	_uniffiStatus
	{%- endif %}
{%- endmacro -%}

// Arglist as used in the _UniFFILib function declations.
// Note unfiltered name but ffi_type_name filters.
{%- macro arg_list_ffi_decl(args, has_call_status) %}
	{%- for arg in args %}
		{{- arg.type_().borrow()|cgo_ffi_type }} {{ arg.name() -}}
		{%- if !loop.last %}, {% endif -%}
	{% endfor -%}
	{%- if has_call_status %}, RustCallStatus* callStatus {% endif -%}
{%- endmacro -%}

{%- macro async_ffi_call_binding(func, prefix) -%}
    {% match (func.return_type(), func.throws_type()) -%}
    {%- when (Some(_), Some(_)) -%} res, err :=
    {%- when (None, Some(_)) -%} _, err :=
    {%- when (Some(_), None) -%} res, _ :=
    {%- when (None, None) -%}
    {%- endmatch -%} {# space -#}
	
    {%- match (func.return_type(), func.throws_type()) %}
    {%- when (Some(return_type), Some(e)) -%}
	uniffiRustCallAsync[{{ e|canonical_name }}](
        {{ e|ffi_converter_instance }},
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) {{ return_type|type_name(ci) }} {
			return {{ return_type|lift_fn }}(
				C.{{ func.ffi_rust_future_complete(ci) }}(handle, status),
			)
		},
    {%- when (None, Some(e)) -%}
	uniffiRustCallAsync[{{ e|canonical_name }}](
        {{ e|ffi_converter_instance }},
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.{{ func.ffi_rust_future_complete(ci) }}(handle, status)
			return struct{}{}
		},
    {%- when (Some(return_type), None) -%}
	uniffiRustCallAsync[struct{}](
        nil,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) {{ return_type|type_name(ci) }} {
			return {{ return_type|lift_fn }}(
				C.{{ func.ffi_rust_future_complete(ci) }}(handle, status),
			)
		},
    {%- when (None, None) -%}
	uniffiRustCallAsync[struct{}](
        nil,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.{{ func.ffi_rust_future_complete(ci) }}(handle, status)
			return struct{}{}
		},
    {%- endmatch %}
		C.{{ func.ffi_func().name() }}({% call _arg_list_ffi_call(func, prefix) %}),
		// pollFn
		func (handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.{{ func.ffi_rust_future_poll(ci) }}(handle, continuation, data)
		},
		// freeFn
		func (handle C.uint64_t) {
			C.{{ func.ffi_rust_future_free(ci) }}(handle)
		},
	)

    {% match (func.return_type(), func.throws_type()) -%}
    {%- when (Some(_), Some(_)) -%} return res, err
    {%- when (None, Some(_)) -%} return err
    {%- when (Some(_), None) -%} return res
    {%- when (None, None) -%}
    {%- endmatch -%}
{%- endmacro -%}

{%- macro lower_fn_call(arg) -%}
{%- match arg.as_type() -%}
{%- when Type::External with { kind, module_path, name, namespace, tagged } -%}
{%- match kind -%}
{%- when ExternalKind::DataClass -%}
RustBufferFromExternal({{ arg|lower_fn }}({{ arg.name()|var_name }}))
{%- else -%}
{{ arg|lower_fn }}({{ arg.name()|var_name }})
{%- endmatch -%}
{%- else -%}
{{ arg|lower_fn }}({{ arg.name()|var_name }})
{%- endmatch -%}
{%- endmacro -%}

{%- macro docstring(defn, indent_tabs) %}
{%- match defn.docstring() %}
{%- when Some(docstring) %}
{{ docstring|docstring(indent_tabs) }}
{%- else %}
{%- endmatch %}
{%- endmacro %}
