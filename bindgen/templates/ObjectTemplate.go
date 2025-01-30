{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

{{- self.add_import("runtime") }}

{%- let obj = ci.get_object_definition(name).expect("missing obj") %}
{%- let (interface_name, impl_name) = obj|object_names %}
{%- let impl_type_name = format!("*{impl_name}") %}

{%- if self.include_once_check("ObjectRuntime.go") %}{% include "ObjectRuntime.go" %}{% endif %}

{%- call go::docstring(obj, 0) %}
type {{ interface_name }} interface {
	{%- for func in obj.methods() -%}
	{%- call go::docstring(func, 1) %}
	{{ func.name()|fn_name }}({%- call go::arg_list_decl(func) -%}) {% call go::return_type_decl(func) %}
	{%- endfor %}
}

{%- call go::docstring(obj, 0) %}
type {{ impl_name }} struct {
	ffiObject FfiObject
}

{%- match obj.primary_constructor() %}
{%- when Some with (cons) %}
{%- call go::docstring(cons, 0) %}
func New{{ impl_name }}({% call go::arg_list_decl(cons) -%}) {% call go::return_type_decl(cons) %} {
	{% call go::ffi_call_binding(cons, "") %}
}
{%- when None %}
{%- endmatch %}

{% for cons in obj.alternate_constructors() -%}
{%- call go::docstring(cons, 0) %}
func {{ impl_name }}{{ cons.name()|fn_name }}({% call go::arg_list_decl(cons) %}) {% call go::return_type_decl(cons) %} {
	{% call go::ffi_call_binding(cons, "") %}
}
{% endfor %}

{% for func in obj.methods() -%}
{%- call go::docstring(func, 0) %}
func (_self {{ impl_type_name }}) {{ func.name()|fn_name }}({%- call go::arg_list_decl(func) -%}) {% call go::return_type_decl(func) %} {
	_pointer := _self.ffiObject.incrementPointer("{{ type_name }}")
	defer _self.ffiObject.decrementPointer()
{%- if func.is_async() %}
	{% call go::async_ffi_call_binding(func, "_pointer") %}
}
{%- else %}
	{% call go::ffi_call_binding(func, "_pointer") %}
}
{%endif %}
{% endfor %}

{%- for tm in obj.uniffi_traits() -%}
{%- match tm %}
{%- when UniffiTrait::Display { fmt } %}
func (_self {{ impl_type_name }}) String() string {
	_pointer := _self.ffiObject.incrementPointer("{{ type_name }}")
	defer _self.ffiObject.decrementPointer()
	{% call go::ffi_call_binding(fmt, "_pointer") %}
}
{% else %}
{% endmatch %}
{% endfor %}

func (object {{ impl_type_name }}) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

{%- let ffi_converter_type = obj|ffi_converter_name %}
{%- let ffi_converter_var = format!("{ffi_converter_type}INSTANCE") %}

{%- if obj.has_callback_interface() %}
{%- let vtable = obj.vtable().expect("trait interface should have a vtable") %}
{%- let vtable_methods = obj.vtable_methods() %}
{%- let ffi_init_callback = obj.ffi_init_callback() %}

{% if self.include_once_check("CallbackInterfaceRuntime.go") %}{% include "CallbackInterfaceRuntime.go" %}{% endif %}
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
				code: uniffiCallbackUnexpectedResultError,
			}
			return
		}
		
		*callStatus = C.RustCallStatus {
			code: uniffiCallbackResultError,
			errorBuf: {{ error_type|lower_fn }}](err),
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
	fmt.Println("Virtual Free")
	{{ ffi_converter_var }}.handleMap.remove(uint64(handle))
}

func (c {{ ffi_converter_type }}) register() {
	C.{{ ffi_init_callback.name() }}(&{{ vtable_name }})
}

{%- endif %}

type {{ ffi_converter_type }} struct {
	{%- if obj.has_callback_interface() %}
	handleMap *concurrentHandleMap[{{ type_name }}]
	{% endif -%}
}

var {{ ffi_converter_var }} = {{ ffi_converter_type }}{
	{%- if obj.has_callback_interface() %}
	handleMap: newConcurrentHandleMap[{{ type_name }}](),
	{% endif -%}
}

func (c {{ ffi_converter_type }}) Lift(pointer unsafe.Pointer) {{ type_name }} {
	result := &{{ impl_name }} {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.{{ obj.ffi_object_clone().name() }}(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.{{ obj.ffi_object_free().name() }}(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, ({{ impl_type_name }}).Destroy)
	return result
}

func (c {{ ffi_converter_type }}) Read(reader io.Reader) {{ type_name }} {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c {{ ffi_converter_type }}) Lower(value {{ type_name }}) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	{%- if obj.has_callback_interface() %}
	pointer := unsafe.Pointer(uintptr(c.handleMap.insert(value)))
	{%- else %}
	pointer := value.ffiObject.incrementPointer("{{ type_name }}")
	defer value.ffiObject.decrementPointer()
	{%- endif %}
	return pointer
	
}

func (c {{ ffi_converter_type }}) Write(writer io.Writer, value {{ type_name }}) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type {{ obj|ffi_destroyer_name }} struct {}

func (_ {{ obj|ffi_destroyer_name }}) Destroy(value {{ type_name }}) {
	// TODO(pna): Not yet confident this is actualy ok
	// techinaly we will only get destroy for type that we actualy own here
	{%- if obj.has_callback_interface() %}
	if val, ok := value.({{ impl_type_name }}); ok {
		val.Destroy()
	} else {
		fmt.Println("Hmm what is this type exacly?")
	}
	{%- else %}
		value.Destroy()
	{%- endif %}
}
