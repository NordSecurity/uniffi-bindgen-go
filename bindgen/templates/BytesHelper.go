{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

type FfiConverterBytes struct{}

var FfiConverterBytesINSTANCE = FfiConverterBytes{}

func (c FfiConverterBytes) lower(value []byte) C.RustBuffer {
	return goBytesToCRustBuffer(value)
}

func (c FfiConverterBytes) write(writer io.Writer, value []byte) {
	if _, err := writer.Write(value); err != nil {
		panic(err)
	}
}

func (c FfiConverterBytes) lift(value C.RustBuffer) []byte {
	return cRustBufferToGoBytes(value)
}

func (c FfiConverterBytes) read(reader io.Reader) []byte {
	result, err := io.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	return result
}

type {{ type_|ffi_destroyer_name }} struct {}

func ({{ type_|ffi_destroyer_name }}) destroy(_ {{ type_name }}) {}

