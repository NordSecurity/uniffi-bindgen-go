{%- let cbi = ci.get_callback_interface_definition(name).expect("missing cbi") %}
{%- let type_name = cbi|type_name(ci) %}
{%- let foreign_callback = format!("foreignCallback{}", canonical_type_name) %}

{%- call go::docstring(cbi, 0) %}
type {{ type_name }} interface {
	{% for meth in cbi.methods() -%}
	{%- call go::docstring(meth, 1) %}
	{{ meth.name()|fn_name }}({% call go::arg_list_decl(meth) %}) {% call go::return_type_decl_cb(meth) %}
	{% endfor %}
}


type {{ ffi_converter_name }} struct {
	handleMap *concurrentHandleMap[{{ type_name }}]
}

var {{ ffi_converter_instance }} = {{ ffi_converter_name }} {
	handleMap: newConcurrentHandleMap[{{ type_name }}](),
}

func (c {{ ffi_converter_name }}) Lift(handle uint64) {{ type_name }} {
	val, ok := c.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}
	return val
}

func (c {{ ffi_converter_name }}) Read(reader io.Reader) {{ type_name }} {
	return c.Lift(readUint64(reader))
}

func (c {{ ffi_converter_name }}) Lower(value {{ type_name }}) C.uint64_t {
	return C.uint64_t(c.handleMap.insert(value))
}

func (c {{ ffi_converter_name }}) Write(writer io.Writer, value {{ type_name }}) {
	writeUint64(writer, uint64(c.Lower(value)))
}

type {{ ffi_destroyer_name }} struct {}

func ({{ ffi_destroyer_name }}) Destroy(value {{ type_name }}) {}

{% let vtable = cbi.vtable() %}
{%- let vtable_methods = cbi.vtable_methods() %}
{%- let ffi_init_callback = cbi.ffi_init_callback() %}

{%- include "VTableImpl.go" %}
