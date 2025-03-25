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
	uniffiObj, ok := {{ ffi_converter_instance }}.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}
	
	{% if meth.is_async() %}
	{%- let result_struct = meth.foreign_future_ffi_result_struct().name()|ffi_struct_name %}
	result := make(chan C.{{ result_struct }}, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture {
		handle: C.uint64_t(guardHandle),
		free: C.UniffiForeignFutureFree(C.{{ config|free_gorutine_callback }}),
	}
	
	// Wait for compleation or cancel
	go func() {
		select {
			case <-cancel:
			case res := <-result:
				{{ ffi_callback|find_ffi_callback_helper -}}
					(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
        asyncResult := &C.{{ result_struct }}{};
    	{%- if meth.return_type().is_some() %}
    	uniffiOutReturn := &asyncResult.returnValue
    	{%- endif %}
    	{%- if meth.throws_type().is_some() %}
    	callStatus := &asyncResult.callStatus
    	{%- endif %}
    	defer func() {
    		result <- *asyncResult
    	}()
	{% endif %}

    {%- match (meth.return_type(), meth.throws_type()) -%}
    {%- when (Some(_), Some(_)) -%} res, err :=
    {%- when (None, Some(_)) -%} err :=
    {%- when (Some(_), None) -%} res :=
    {%- when (None, None) -%}
    {%- endmatch -%} {# space -#}
    uniffiObj.{{ meth.name()|fn_name }}(
        {%- for arg in meth.arguments() %}
        {{ arg|lift_fn }}({{ arg.name()|var_name }}),
        {%- endfor %}
    )
	
    {% if let Some(error_type) = meth.throws_type() -%}
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
    {%- endif %}

	{% if let Some(return_type) = meth.return_type() -%}
	*uniffiOutReturn = {{ return_type|lower_fn }}(res)
	{%- endif %}

	{%- if meth.is_async() %}
	}()
	{%- endif %}

	{#
    {%- else %}
    // CancellationTokenSource cts = new CancellationTokenSource();

    // Task.Run(async () => {
    go func() {
        ret := C.{{ meth.foreign_future_ffi_result_struct()|ffi_type_name_cgo_safe }}{};
        // ret.@callStatus = new UniffiRustCallStatus();

        {%- match meth.throws_type() %}
        {%- when Some with (error_type) %}
        try {
        {%- when None %}
        {%- endmatch %}

        {%- match meth.return_type() %}
        {%- when Some with (return_type) %}
        var result =
        {%- when None %}
        {%- endmatch %}

        await uniffiObj.{{ meth.name()|fn_name }}(
            {%- for arg in meth.arguments() %}
            {{ arg|lift_fn }}({{ arg.name()|var_name }}){%- if !loop.last %}, {% endif -%}
            {%- endfor %}).WaitAsync(cts.Token);


        {%- match meth.return_type() %}
        {%- when Some with (return_type) %}
        {%- let complete_fn_type = return_type|ffi_foreign_future_complete %}
        ret.@returnValue = {{ return_type|ffi_converter_name }}.INSTANCE.Lower(result);
        {%- when None %}
        {%- let complete_fn_type = "ForeignFutureCompleteVoid" %}
        {%- endmatch %}

        ret.@callStatus.code = UniffiCallbackResponseStatus.SUCCESS;
        {%- match meth.throws_type() %}
        {%- when Some with (error_type) %}
        } catch ({{ error_type|type_name(ci) }} e) {
            ret.@callStatus.code = UniffiCallbackResponseStatus.ERROR;
            ret.@callStatus.error_buf = {{ error_type|ffi_converter_name }}.INSTANCE.Lower(e);
        } catch (System.Exception e){
            ret.@callStatus.code = UniffiCallbackResponseStatus.UNEXPECTED_ERROR;
            try {
                ret.@callStatus.error_buf = FfiConverterString.INSTANCE.Lower(e.Message);
            }
            catch {
            }
        }
        {%- when None %}
        {%- endmatch %}

        {% match meth.return_type() %}
        {%- when Some with (return_type) %}
        {%- let complete_fn_type = return_type|ffi_foreign_future_complete %}
        var cb = Marshal.GetDelegateForFunctionPointer<_UniFFILib.{{ complete_fn_type }}>(@uniffiFutureCallback);
        {%- when None %}
        var cb = Marshal.GetDelegateForFunctionPointer<_UniFFILib.UniffiForeignFutureCompleteVoid>(@uniffiFutureCallback);
        {%- endmatch %}
        cb(@uniffiCallbackData, ret);
    }()

    var foreignHandle = _UniFFIAsync._foreign_futures_map.Insert(cts);
    @uniffiOutReturn.@handle = foreignHandle;
    @uniffiOutReturn.@free = Marshal.GetFunctionPointerForDelegate(_UniFFIAsync.UniffiForeignFutureFreeCallback.callback);
    {%- endif %}
    #}
}

{% endfor %}

{%- let free_callback = self::oracle().cgo_vtable_free_fn_name(name, module_path) %}
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
	{{ ffi_converter_instance }}.handleMap.remove(uint64(handle))
}

func (c {{ ffi_converter_name }}) register() {
	C.{{ ffi_init_callback.name() }}(&{{ vtable_name }})
}

