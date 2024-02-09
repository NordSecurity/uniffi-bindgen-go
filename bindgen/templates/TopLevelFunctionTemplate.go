{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

{%- call go::docstring(func, 0) %}
func {{ func.name()|fn_name}}({%- call go::arg_list_decl(func) -%}) {% call go::return_type_decl(func) %} {
{%- if func.is_async() %}
	{% call go::async_ffi_call_binding(func, "") %}
{%- else %}
	{% call go::ffi_call_binding(func, "") %}
{%- endif %}
}
