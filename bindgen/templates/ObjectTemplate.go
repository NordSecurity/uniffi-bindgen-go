{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

{{- self.add_import("runtime") }}

{%- let obj = ci.get_object_definition(name).expect("missing obj") %}
{%- let obj_name = obj.name()|class_name %}

{%- if self.include_once_check("ObjectRuntime.go") %}{% include "ObjectRuntime.go" %}{% endif %}

{%- call go::docstring(obj, 0) %}
type {{ obj_name }} struct {
	ffiObject FfiObject
}

{%- match obj.primary_constructor() %}
{%- when Some with (cons) %}
{%- call go::docstring(cons, 0) %}
func New{{ obj_name }}({% call go::arg_list_decl(cons) -%}) {% call go::return_type_decl(cons) %} {
	{% call go::ffi_call_binding(func, "") %}
}
{%- when None %}
{%- endmatch %}

{% for cons in obj.alternate_constructors() -%}
{%- call go::docstring(cons, 0) %}
func {{ obj_name }}{{ cons.name()|fn_name }}({% call go::arg_list_decl(cons) %}) {% call go::return_type_decl(cons) %} {
	{% call go::ffi_call_binding(func, "") %}
}
{% endfor %}

{% for func in obj.methods() -%}
{%- call go::docstring(func, 0) %}
func (_self {{ type_name }}){{ func.name()|fn_name }}({%- call go::arg_list_decl(func) -%}) {% call go::return_type_decl(func) %} {
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
func (_self {{ type_name }})String() string {
	_pointer := _self.ffiObject.incrementPointer("{{ type_name }}")
	defer _self.ffiObject.decrementPointer()
	{% call go::ffi_call_binding(fmt, "_pointer") %}
}
{% else %}
{% endmatch %}
{% endfor %}

func (object {{ type_name }})Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type {{ obj|ffi_converter_name }} struct {}

var {{ obj|ffi_converter_name }}INSTANCE = {{ obj|ffi_converter_name }}{}

func (c {{ obj|ffi_converter_name }}) Lift(pointer unsafe.Pointer) {{ type_name }} {
	result := &{{ obj_name }} {
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
	runtime.SetFinalizer(result, ({{ type_name }}).Destroy)
	return result
}

func (c {{ obj|ffi_converter_name }}) Read(reader io.Reader) {{ type_name }} {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c {{ obj|ffi_converter_name }}) Lower(value {{ type_name }}) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("{{ type_name }}")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c {{ obj|ffi_converter_name }}) Write(writer io.Writer, value {{ type_name }}) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type {{ obj|ffi_destroyer_name }} struct {}

func (_ {{ obj|ffi_destroyer_name }}) Destroy(value {{ type_name }}) {
	value.Destroy()
}
