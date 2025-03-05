{%- let cbi = ci.get_callback_interface_definition(name).expect("missing cbi") %}
{%- let type_name = cbi|type_name %}
{%- let foreign_callback = format!("foreignCallback{}", canonical_type_name) %}

{%- call go::docstring(cbi, 0) %}
type {{ type_name }} interface {
	{% for meth in cbi.methods() -%}
	{%- call go::docstring(meth, 1) %}
	{{ meth.name()|fn_name }}({% call go::arg_list_decl(meth) %}) {% call go::return_type_decl_cb(meth) %}
	{% endfor %}
}

{%- let ffi_converter_type = cbi|ffi_converter_name %}
{%- let ffi_converter_var = format!("{ffi_converter_type}INSTANCE") %}

type {{ ffi_converter_type }} struct {
	handleMap *concurrentHandleMap[{{ type_name }}]
}

var {{ ffi_converter_var }} = {{ ffi_converter_type }} {
	handleMap: newConcurrentHandleMap[{{ type_name }}](),
}

// TODO(pna): where was this used?
func (c {{ ffi_converter_type }}) drop(handle uint64) RustBuffer {
	c.handleMap.remove(handle)
	return RustBuffer{}
}

func (c {{ ffi_converter_type }}) Lift(handle uint64) {{ type_name }} {
	val, ok := c.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}
	return val
}

func (c {{ ffi_converter_type }}) Read(reader io.Reader) {{ type_name }} {
	return c.Lift(readUint64(reader))
}

func (c {{ ffi_converter_type }}) Lower(value {{ type_name }}) C.uint64_t {
	return C.uint64_t(c.handleMap.insert(value))
}

func (c {{ ffi_converter_type }}) Write(writer io.Writer, value {{ type_name }}) {
	writeUint64(writer, uint64(c.Lower(value)))
}

type {{ cbi|ffi_destroyer_name }} struct {}

func ({{ cbi|ffi_destroyer_name }}) Destroy(value {{ type_name }}) {
}

{% let vtable = cbi.vtable() %}
{%- let vtable_methods = cbi.vtable_methods() %}
{%- let ffi_init_callback = cbi.ffi_init_callback() %}

{%- include "VTableImpl.go" %}
