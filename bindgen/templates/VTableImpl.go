{% if self.include_once_check("VTableRuntime.go") %}{% include "VTableRuntime.go" %}{% endif %}
{{- self.add_import("sync") }}

{%- for (ffi_callback, meth) in vtable_methods.iter() %}

{% let callback_name = ffi_callback|cgo_callback_fn_name(module_path) -%}

//export {{ callback_name }}
func {{ callback_name }}(
	    {%- for arg in ffi_callback.arguments() -%}
	    {{- arg.name().borrow()|var_name }} {{ arg.type_().borrow()|ffi_type_name_cgo_safe }},
	    {%- endfor -%}
	    {%- if ffi_callback.has_rust_call_status_arg() -%}
	    callStatus *C.RustCallStatus,
	    {%- endif -%}	
	) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := {{ ffi_converter_var }}.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}
	
	{% match meth.return_type() -%}
	{%- when Some with (return_type) -%}
        {%- match meth.throws_type() -%}
        {%- when Some(error_type) -%}
	result, err :=
        {%- when None -%}
	result :=
        {%- endmatch -%}
	{%- when None -%}
	{%- match meth.throws_type() -%}
        {%- when Some(error_type) -%}
	err :=
        {%- when None -%}
        {%- endmatch -%}
	{%- endmatch -%}
    uniffiObj.{{ meth.name()|fn_name }}(
        {%- for arg in meth.arguments() %}
        {{ arg|lift_fn }}({{ arg.name()|var_name }}),
        {%- endfor %}
    )
	
    {% match meth.throws_type() -%}
    {%- when Some(error_type) -%}
	if err != nil {
		// The only way to bypass an unexpected error is to bypass pointer to an empty
		// instance of the error
		if err.err == nil {
			*callStatus = C.RustCallStatus {
				code: C.int8_t(uniffiCallbackUnexpectedResultError),
			}
			return
		}
		
		*callStatus = C.RustCallStatus {
			code: C.int8_t(uniffiCallbackResultError),
			errorBuf: {{ error_type|lower_fn }}(err),
		}
		return
	}
    {%- when None -%}
    {%- endmatch %}

	{% match meth.return_type() -%}
	{%- when Some(return_type) -%}
	*uniffiOutReturn = {{ return_type|lower_fn }}(result)
	{%- when None -%}
	{%- endmatch %}
	return
}

{% endfor %}

{# TODO(pna): make this part of oracle / filter api #}
{%- let free_callback = format!("{module_path}_cgo_dispatchCallbackInterface{name}Free") %}
{%- let free_type = "CallbackInterfaceFree"|ffi_callback_name %}

{%- let vtable_name = vtable|cgo_ffi_type -%}
{%- let vtable_name = format!("{vtable_name}INSTANCE") -%}

var {{ vtable_name }} = {{ vtable|ffi_type_name_cgo_safe }} {
	{%- for (ffi_callback, meth) in vtable_methods.iter() %}
	{% let callback_name = ffi_callback|cgo_callback_fn_name(module_path) -%}
	{% let callback_type = ffi_callback.name()|ffi_callback_name -%}
	
	{{ meth.name()|var_name }}: (C.{{ callback_type }})(C.{{ callback_name }}),
	
	{%- endfor %}

	uniffiFree: (C.{{ free_type }})(C.{{ free_callback }}),
}

//export {{ free_callback }}
func {{ free_callback }}(handle C.uint64_t) {
	{{ ffi_converter_var }}.handleMap.remove(uint64(handle))
}

func (c {{ ffi_converter_type }}) register() {
	C.{{ ffi_init_callback.name() }}(&{{ vtable_name }})
}

