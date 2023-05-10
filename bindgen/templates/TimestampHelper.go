{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

{{- self.add_import("time") }}

type FfiConverterTimestamp struct{}

var FfiConverterTimestampINSTANCE = FfiConverterTimestamp{}

func (c FfiConverterTimestamp) lift(cRustBuf C.RustBuffer) time.Time {
	rustBuffer := fromCRustBuffer(cRustBuf)
	return liftFromRustBuffer[time.Time](c, rustBuffer)
}

func (c FfiConverterTimestamp) read(reader io.Reader) time.Time {
	sec := readInt64(reader)
	nsec := readUint32(reader)

	var sign int64 = 1
	if sec < 0 {
		sign = -1
	}

	return time.Unix(sec, int64(nsec) * sign)
}

func (c FfiConverterTimestamp) lower(value time.Time) C.RustBuffer {
	return lowerIntoRustBuffer[time.Time](c, value)
}

func (c FfiConverterTimestamp) write(writer io.Writer, value time.Time) {
	sec := value.Unix()
	nsec := uint32(value.Nanosecond())
	if value.Unix() < 0 {
		nsec = 1_000_000_000 - nsec
		sec += 1
	}

	writeInt64(writer, sec)
	writeUint32(writer, nsec)
}

type {{ type_|ffi_destroyer_name }} struct {}

func ({{ type_|ffi_destroyer_name }}) destroy(_ {{ type_name }}) {}
