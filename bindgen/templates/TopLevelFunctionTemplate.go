{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

func {{ func.name()|fn_name}}({%- call go::arg_list_decl(func) -%}) {% call go::return_type_decl(func) %} {
	{% call go::ffi_call_binding(func, "") %}
}
