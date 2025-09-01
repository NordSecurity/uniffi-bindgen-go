{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

type {{ ffi_converter_name }} struct{}

var {{ ffi_converter_instance }} = {{ ffi_converter_name }}{}

func (c {{ ffi_converter_name }}) Lift(rb RustBufferI) {{ type_name }} {
	return LiftFromRustBuffer[{{ type_name }}](c, rb)
}

func (_ {{ ffi_converter_name }}) Read(reader io.Reader) {{ type_name }} {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := {{ inner_type|read_fn }}(reader)
	return &temp
}

func (c {{ ffi_converter_name }}) Lower(value {{ type_name }}) C.RustBuffer {
	return LowerIntoRustBuffer[{{ type_name }}](c, value)
}

func (c {{ ffi_converter_name }}) LowerExternal(value {{ type_name }}) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[{{ type_name }}](c, value))
}

func (_ {{ ffi_converter_name }}) Write(writer io.Writer, value {{ type_name }}) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		{{ inner_type|write_fn }}(writer, *value)
	}
}

type {{ ffi_destroyer_name }} struct {}

func (_ {{ ffi_destroyer_name }}) Destroy(value {{ type_name }}) {
	if value != nil {
		{{ inner_type|destroy_fn }}(*value)
	}
}
