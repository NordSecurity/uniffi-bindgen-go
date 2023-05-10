{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

{{- self.add_import("runtime") }}

{%- let obj = ci.get_object_definition(name).unwrap() %}
{%- let canonical_name = type_|canonical_name %}
{%- if self.include_once_check("ObjectRuntime.go") %}{% include "ObjectRuntime.go" %}{% endif %}

type {{ canonical_name }} struct {
	ffiObject FfiObject
}

{%- match obj.primary_constructor() %}
{%- when Some with (cons) %}
func New{{ canonical_name }}({% call go::arg_list_decl(cons) -%}) {% call go::return_type_decl(cons) %} {
	{% call go::ffi_call_binding(func, "") %}
}
{%- when None %}
{%- endmatch %}

{% for cons in obj.alternate_constructors() -%}
func {{ canonical_name }}{{ cons.name()|fn_name }}({% call go::arg_list_decl(cons) %}) {% call go::return_type_decl(cons) %} {
	{% call go::ffi_call_binding(func, "") %}
}
{% endfor %}

{% for func in obj.methods() -%}
func (_self {{ type_name }}){{ func.name()|fn_name }}({%- call go::arg_list_decl(func) -%}) {% call go::return_type_decl(func) %} {
	_pointer := _self.ffiObject.incrementPointer("{{ type_name }}")
	defer _self.ffiObject.decrementPointer()
	{% call go::ffi_call_binding(func, "_pointer") %}
}
{% endfor %}

func (object {{ type_name }})Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type {{ obj|ffi_converter_name }} struct {}

var {{ obj|ffi_converter_name }}INSTANCE = {{ obj|ffi_converter_name }}{}

func (c {{ obj|ffi_converter_name }}) lift(pointer unsafe.Pointer) {{ type_name }} {
	result := &{{ canonical_name }} {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.{{ obj.ffi_object_free().name() }}(pointer, status)
		}),
	}
	runtime.SetFinalizer(result, ({{ type_name }}).Destroy)
	return result
}

func (c {{ obj|ffi_converter_name }}) read(reader io.Reader) {{ type_name }} {
	return c.lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c {{ obj|ffi_converter_name }}) lower(value {{ type_name }}) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("{{ type_name }}")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c {{ obj|ffi_converter_name }}) write(writer io.Writer, value {{ type_name }}) {
	writeUint64(writer, uint64(uintptr(c.lower(value))))
}

type {{ obj|ffi_destroyer_name }} struct {}

func (_ {{ obj|ffi_destroyer_name }}) destroy(value {{ type_name }}) {
	value.Destroy()
}
