{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

{%- let type_name = type_|type_name %}

type {{ ffi_converter_name }} struct{}

var {{ ffi_converter_name }}INSTANCE = {{ ffi_converter_name }}{}

func (c {{ ffi_converter_name }}) lift(cRustBuf C.RustBuffer) {{ type_name }} {
	return liftFromRustBuffer[{{ type_name }}](c, fromCRustBuffer(cRustBuf))
}

func (_ {{ ffi_converter_name }}) read(reader io.Reader) {{ type_name }} {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := {{ inner_type|read_fn }}(reader)
	return &temp
}

func (c {{ ffi_converter_name }}) lower(value {{ type_name }}) C.RustBuffer {
	return lowerIntoRustBuffer[{{ type_name }}](c, value)
}

func (_ {{ ffi_converter_name }}) write(writer io.Writer, value {{ type_name }}) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		{{ inner_type|write_fn }}(writer, *value)
	}
}

type {{ ffi_destroyer_name }} struct {}

func (_ {{ ffi_destroyer_name }}) destroy(value {{ type_name }}) {
	if value != nil {
		{{ inner_type|destroy_fn }}(*value)
	}
}
