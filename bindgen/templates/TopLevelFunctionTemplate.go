{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

{%- if func.is_async() %}
func {{ func.name()|fn_name}}({%- call go::arg_list_decl(func) -%}) {% call go::return_type_decl(func) %} {
	// We create a channel, that this function blocks on, until the callback sends a result on it
	
	done := make(chan {{ func.result_type().borrow()|future_chan_type }})
	chanHandle := cgo.NewHandle(done)
	defer chanHandle.Delete()

	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.{{ func.ffi_func().name() }}(
			{%- for arg in func.arguments() %}
			{{- arg|lower_fn_call }},
			{%- endfor %}
			FfiConverterForeignExecutorINSTANCE.Lower(UniFfiForeignExecutor {}),
			C.UniFfiFutureCallback{{ func.result_type().future_callback_param().borrow()|cgo_ffi_callback_type }}(C.{{ func.result_type().borrow()|future_callback }}),
			unsafe.Pointer(chanHandle),
			_uniffiStatus,
		)
		return false
	})

	// wait for things to be done
        res := <- done
	
	{%- match func.return_type() -%}
	{%- when Some with (return_type) -%}
		{%- match func.throws_type() -%}
		{%- when Some with (throws_type) %}
	 return res.val, res.err
		{%- when None %}
	 return res.val
		{%- endmatch %}
	{%- when None -%}
		{%- match func.throws_type() -%}
		{%- when Some with (throws_type) %}
	 return res.err
		{%- when None %}
         _ = res
		{%- endmatch %}
	{%- endmatch %}
}
{%- else %}
func {{ func.name()|fn_name}}({%- call go::arg_list_decl(func) -%}) {% call go::return_type_decl(func) %} {
	{% call go::ffi_call_binding(func, "") %}
}
{% endif %}
