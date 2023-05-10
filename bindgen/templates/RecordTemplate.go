{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

{%- let rec = ci.get_record_definition(name).unwrap() %}

type {{ type_name }} struct {
	{%- for field in rec.fields() %}
	{{ field.name()|field_name }} {{ field|type_name -}}
	{%- endfor %}
}

func (r *{{ type_name }}) Destroy() {
	{%- for field in rec.fields() %}
		{{ field|destroy_fn }}(r.{{ field.name()|field_name }});
	{%- endfor %}
}

type {{ rec|ffi_converter_name }} struct {}

var {{ rec|ffi_converter_name }}INSTANCE = {{ rec|ffi_converter_name }}{}

func (c {{ rec|ffi_converter_name }}) lift(cRustBuf C.RustBuffer) {{ type_name }} {
	rustBuffer := fromCRustBuffer(cRustBuf)
	return liftFromRustBuffer[{{ type_name }}](c, rustBuffer)
}

func (c {{ rec|ffi_converter_name }}) read(reader io.Reader) {{ type_name }} {
	return {{ type_name }} {
		{%- for field in rec.fields() %}
			{{ field|read_fn }}(reader),
		{%- endfor %}
	}
}

func (c {{ rec|ffi_converter_name }}) lower(value {{ type_name }}) C.RustBuffer {
	return lowerIntoRustBuffer[{{ type_name }}](c, value)
}

func (c {{ rec|ffi_converter_name }}) write(writer io.Writer, value {{ type_name }}) {
	{%- for field in rec.fields() %}
		{{ field|write_fn }}(writer, value.{{ field.name()|field_name }});
	{%- endfor %}
}

type {{ ffi_destroyer_name }} struct {}

func (_ {{ ffi_destroyer_name }}) destroy(value {{ type_name }}) {
	value.Destroy()
}
