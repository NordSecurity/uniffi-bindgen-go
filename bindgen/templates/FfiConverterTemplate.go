{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

type bufLifter[GoType any] interface {
	lift(value C.RustBuffer) GoType
}

type bufLowerer[GoType any] interface {
	lower(value GoType) C.RustBuffer
}

type ffiConverter[GoType any, FfiType any] interface {
	lift(value FfiType) GoType
	lower(value GoType) FfiType
}

type bufReader[GoType any] interface {
	read(reader io.Reader) GoType
}

type bufWriter[GoType any] interface {
	write(writer io.Writer, value GoType)
}

type ffiRustBufConverter[GoType any, FfiType any] interface {
	ffiConverter[GoType, FfiType]
	bufReader[GoType]
}

func lowerIntoRustBuffer[GoType any](bufWriter bufWriter[GoType], value GoType) C.RustBuffer {
	// This might be not the most efficient way but it does not require knowing allocation size
	// beforehand
	var buffer bytes.Buffer
	bufWriter.write(&buffer, value)

	bytes, err := io.ReadAll(&buffer)
	if err != nil {
		panic(fmt.Errorf("reading written data: %w", err))
	}
	return goBytesToCRustBuffer(bytes)
}

func liftFromRustBuffer[GoType any](bufReader bufReader[GoType], rbuf rustBuffer) GoType {
	defer rbuf.free()
	reader := rbuf.asReader()
	item := bufReader.read(reader)
	if reader.Len() > 0 {
		// TODO: Remove this
		leftover, _ := io.ReadAll(reader)
		panic(fmt.Errorf("Junk remaining in buffer after lifting: %s", string(leftover)))
	}
	return item
}

