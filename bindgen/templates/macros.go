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
		({{ return_type|type_name(ci) }}, error)
		{%- when None -%}
		{{ return_type|type_name(ci) }}
		{%- endmatch %}
	{%- when None -%}
		{%- match func.throws_type() -%}
		{%- when Some with (throws_type) -%}
		error
		{%- when None -%}
		{%- endmatch %}
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
			return {{ return_type|lift_fn }}(_uniffiRV), nil
		}
		{%- when None -%}
		return {{ return_type|lift_fn }}({% call to_ffi_call(func, prefix) %})
		{%- endmatch -%}
	{%- when None -%}
		{%- match func.throws_type() -%}
		{%- when Some with (throws_type) -%}
		_, _uniffiErr := {% call to_ffi_call(func, prefix) %}
		return _uniffiErr.AsError()
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
		return {% call ffi_invoke(func, prefix) %}
	})
	{%- else -%}
	func(_uniffiStatus *C.RustCallStatus) bool {
		{% call ffi_invoke(func, prefix) %}
		return false
	})
	{%- endmatch %}
{%- endmacro -%}

{%- macro ffi_invoke(func, prefix) -%}
	{%- if let Some(FfiType::RustBuffer(_)) = func.ffi_func().return_type() -%}
	GoRustBuffer {
		inner: C.{{ func.ffi_func().name() }}({% call _arg_list_ffi_call(func, prefix) -%}),
	}
	{%- else -%}
	C.{{ func.ffi_func().name() }}({% call _arg_list_ffi_call(func, prefix) -%})
	{%- endif -%}
{%- endmacro -%}

{%- macro remap_ffi_val(type_, val) -%}
	{%- if let FfiType::RustBuffer(_) = type_|into_ffi_type -%}
	GoRustBuffer {
		inner: {{ val }},
	}
	{%- else -%}
	{{ val }}
	{%- endif -%}
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

{%- macro func_return_vars_pairs(func, prefix = "", suffix = "") -%}
    {%- match (func.return_type(), func.throws_type()) -%}
    {%- when (Some(_), Some(_)) -%} {{ prefix }} res, err {{ suffix }}
    {%- when (None, Some(_)) -%} {{ prefix }} _, err {{ suffix }}
    {%- when (Some(_), None) -%} {{ prefix }} res, _ {{ suffix }}
    {%- when (None, None) -%}
    {%- endmatch -%} {# space -#}
{%- endmacro -%}

{%- macro func_return_vars(func, prefix = "", suffix = "") -%}
    {% match (func.return_type(), func.throws_type()) -%}
    {%- when (Some(_), Some(_)) -%} {{ prefix }} res, err {{ suffix }}
    {%- when (None, Some(_)) -%} {{ prefix }} err {{ suffix }}
    {%- when (Some(_), None) -%} {{ prefix }} res {{ suffix }}
    {%- when (None, None) -%}
    {%- endmatch -%}
{%- endmacro -%}

{%- macro func_nil_err_check(func) -%}
    {% match (func.return_type(), func.throws_type()) -%}
    {%- when (Some(_), Some(_)) -%}
	if err == nil {
		return res, nil
	}
    {%- when (None, Some(_)) -%}
	if err == nil {
		return nil
	}
    {%- when (Some(_), None) -%}
    {%- when (None, None) -%}
    {%- endmatch -%}
{%- endmacro -%}

{%- macro async_ffi_call_binding(func, prefix) -%}
	{%- call func_return_vars_pairs(func, suffix = ":=") -%}
	
    {%- match (func.return_type(), func.throws_type()) %}
    {%- when (Some(return_type), Some(e)) -%}
	uniffiRustCallAsync[{{ e|canonical_name }}](
        {{ e|ffi_converter_instance }},
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) {{ return_type|ffi_type_name }} {
			res := C.{{ func.ffi_rust_future_complete(ci) }}(handle, status)
			return {% call remap_ffi_val(return_type, "res") %}
		},
		// liftFn
		func(ffi {{ return_type|ffi_type_name }}) {{ return_type|type_name(ci) }} {
			return {{ return_type|lift_fn }}(ffi)
		},
    {%- when (None, Some(e)) -%}
	uniffiRustCallAsync[{{ e|canonical_name }}](
        {{ e|ffi_converter_instance }},
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.{{ func.ffi_rust_future_complete(ci) }}(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
    {%- when (Some(return_type), None) -%}
	uniffiRustCallAsync[error](
        nil,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) {{ return_type|ffi_type_name }} {
			res := C.{{ func.ffi_rust_future_complete(ci) }}(handle, status)
			return {% call remap_ffi_val(return_type, "res") %}
		},
		// liftFn
		func(ffi {{ return_type|ffi_type_name }}) {{ return_type|type_name(ci) }} {
			return {{ return_type|lift_fn }}(ffi)
		},
    {%- when (None, None) -%}
	uniffiRustCallAsync[error](
        nil,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.{{ func.ffi_rust_future_complete(ci) }}(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
    {%- endmatch %}
		{% call ffi_invoke(func, prefix) %},
		// pollFn
		func (handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.{{ func.ffi_rust_future_poll(ci) }}(handle, continuation, data)
		},
		// freeFn
		func (handle C.uint64_t) {
			C.{{ func.ffi_rust_future_free(ci) }}(handle)
		},
	)

	{% call func_nil_err_check(func) %}

	{% call func_return_vars(func, prefix = "return") %}
{%- endmacro -%}

{%- macro lower_fn_call(arg) -%}
{%- match arg.as_type() -%}
{%- when Type::External with { kind, module_path, name, namespace, tagged } -%}
{%- match kind -%}
{%- when ExternalKind::DataClass -%}
CFromRustBuffer({{ arg|lower_external_fn }}({{ arg.name()|var_name }}))
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
