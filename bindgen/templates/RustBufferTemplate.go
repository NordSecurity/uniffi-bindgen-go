{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

type rustBuffer struct {
	capacity int
	length   int
	data     unsafe.Pointer
	self     C.RustBuffer
}

func fromCRustBuffer(crbuf C.RustBuffer) rustBuffer {
	return rustBuffer{
		capacity: int(crbuf.capacity),
		length:   int(crbuf.len),
		data:     unsafe.Pointer(crbuf.data),
		self:     crbuf,
	}
}

// asByteBuffer reads the full rust buffer and then converts read bytes to a new reader which makes
// it quite inefficient
// TODO: Return an implementation which reads only when needed
func (rb rustBuffer) asReader() *bytes.Reader {
	b := C.GoBytes(rb.data, C.int(rb.length))
	return bytes.NewReader(b)
}

func (rb rustBuffer) asCRustBuffer() C.RustBuffer {
	return C.RustBuffer{
		capacity: C.int(rb.capacity),
		len: C.int(rb.length),
		data: (*C.uchar)(unsafe.Pointer(rb.data)),
	}
}

func stringToCRustBuffer(str string) C.RustBuffer {
	b := []byte(str)
	cs := C.CString(str)
	return C.RustBuffer{
		capacity: C.int(len(b)),
		len: C.int(len(b)),
		data: (*C.uchar)(unsafe.Pointer(cs)),
	}
}

func (rb rustBuffer) free() {
	rustCall(func( status *C.RustCallStatus) bool {
		C.{{ ci.ffi_rustbuffer_free().name() }}(rb.self, status)
		return false
	})
}

func goBytesToCRustBuffer(b []byte) C.RustBuffer {
	cs := C.CBytes(b)
	return C.RustBuffer{
		capacity: C.int(len(b)),
		len: C.int(len(b)),
		data: (*C.uchar)(unsafe.Pointer(cs)),
	}
}

func cRustBufferToGoBytes(b C.RustBuffer) []byte {
	return C.GoBytes(unsafe.Pointer(b.data), C.int(b.len))
}
