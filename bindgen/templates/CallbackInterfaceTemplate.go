{%- let cbi = ci.get_callback_interface_definition(name).unwrap() %}
{%- let type_name = cbi|type_name %}
{%- let foreign_callback = format!("foreignCallback{}", canonical_type_name) %}

{% if self.include_once_check("CallbackInterfaceRuntime.go") %}{% include "CallbackInterfaceRuntime.go" %}{% endif %}
{{- self.add_import("sync") }}

// Declaration and FfiConverters for {{ type_name }} Callback Interface
type {{ type_name }} interface {
	{% for meth in cbi.methods() -%}
	{{ meth.name()|fn_name }}({% call go::arg_list_decl(meth) %}) {% call go::return_type_decl_cb(meth) %}
	{% endfor %}
}

// {{ foreign_callback }} cannot be callable be a compiled function at a same time
type {{ foreign_callback }} struct {}

{% let cgo_callback_fn = self.cgo_callback_fn(type_name) -%}
//export {{ cgo_callback_fn }}
func {{ cgo_callback_fn }}(handle C.uint64_t, method C.int32_t, argsPtr *C.uint8_t, argsLen C.int32_t, outBuf *C.RustBuffer) C.int32_t {
	cb := {{ type_|lift_fn }}(uint64(handle));
	switch method {
	case 0:
		// 0 means Rust is done with the callback, and the callback
		// can be dropped by the foreign language.
		*outBuf = {{ ffi_converter_name }}INSTANCE.drop(uint64(handle)).asCRustBuffer()
		// See docs of ForeignCallback in `uniffi/src/ffi/foreigncallbacks.rs`
		return C.int32_t(idxCallbackFree)

	{% for meth in cbi.methods() -%}
	{% let method_name = meth.name()|fn_name -%}
	{% let method_name = format!("Invoke{}", method_name) -%}
	case {{ loop.index }}:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = {{ foreign_callback}}{}.{{ method_name }}(cb, args, outBuf);
		return C.int32_t(result)
	{% endfor %}
	default:
		// This should never happen, because an out of bounds method index won't
		// ever be used. Once we can catch errors, we should return an InternalException.
		// https://github.com/mozilla/uniffi-rs/issues/351
		return C.int32_t(uniffiCallbackUnexpectedResultError)
	}
}

{% for meth in cbi.methods() -%}
{% let method_name = meth.name()|fn_name -%}
{% let method_name = format!("Invoke{}", method_name) -%}
func ({{ foreign_callback }}) {{ method_name }} (callback {{ type_name }}, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	{% if meth.arguments().len() != 0 -%}
	reader := bytes.NewReader(args)
	{% endif -%}
	{%- match meth.return_type() -%}
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
	callback.{{ meth.name()|fn_name }}(
		{%- for arg in meth.arguments() -%}
		{{ arg|read_fn }}(reader)
		{%- if !loop.last %}, {% endif -%}
		{%- endfor -%}
		);

        {% match meth.throws_type() -%}
        {%- when Some(error_type) -%}
	if err != nil {
		// The only way to bypass an unexpected error is to bypass pointer to an empty
		// instance of the error
		if err.err == nil {
			return uniffiCallbackUnexpectedResultError
		}
		*outBuf = lowerIntoRustBuffer[*{{ error_type|type_name }}]({{ error_type|ffi_converter_name }}INSTANCE, err)
		return uniffiCallbackResultError
	}
        {%- when None -%}
        {%- endmatch %}
	{% match meth.return_type() -%}
	{%- when Some with (return_type) -%}
	*outBuf = lowerIntoRustBuffer[{{ return_type|type_name }}]({{ return_type|ffi_converter_name }}INSTANCE, result)
	return uniffiCallbackResultSuccess 
	{%- else -%}
	return uniffiCallbackResultSuccess
	{%- endmatch %}
}
{% endfor %}

type {{ ffi_converter_name }} struct {
	FfiConverterCallbackInterface[{{ type_name }}]
}

var {{ ffi_converter_name }}INSTANCE = &{{ ffi_converter_name }} {
	FfiConverterCallbackInterface: FfiConverterCallbackInterface[{{ type_name }}]{
		handleMap: newConcurrentHandleMap[{{ type_name }}](),
	},
}

// This is a static function because only 1 instance is supported for registering
func (c *{{ ffi_converter_name }}) register() {
	rustCall(func(status *C.RustCallStatus) int32 {
		C.{{ cbi.ffi_init_callback().name() }}(C.ForeignCallback(C.{{ cgo_callback_fn }}), status)
		return 0
	})
}

type {{ cbi|ffi_destroyer_name }} struct {}

func ({{ cbi|ffi_destroyer_name }}) destroy(value {{ type_name }}) {
}

