{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

type rustBuffer struct {
	self     C.RustBuffer
}

func fromCRustBuffer(crbuf C.RustBuffer) rustBuffer {
	capacity := int(crbuf.capacity)
	length := int(crbuf.len)
	data := unsafe.Pointer(crbuf.data)
	
	if data == nil && (capacity > 0 || length > 0) {
		panic(fmt.Sprintf("null in valid C.RustBuffer, capacity non null on null data: %d, %d, %s", capacity, length, data))
	}
	return rustBuffer{
		self:     crbuf,
	}
}

// asByteBuffer reads the full rust buffer and then converts read bytes to a new reader which makes
// it quite inefficient
// TODO: Return an implementation which reads only when needed
func (rb rustBuffer) asReader() *bytes.Reader {
	b := C.GoBytes(unsafe.Pointer(rb.self.data), C.int(rb.self.len))
	return bytes.NewReader(b)
}

func (rb rustBuffer) asCRustBuffer() C.RustBuffer {
	return rb.self
}

func stringToCRustBuffer(str string) C.RustBuffer {
	return goBytesToCRustBuffer([]byte(str))
}

func (rb rustBuffer) free() {
	rustCall(func( status *C.RustCallStatus) bool {
		C.{{ ci.ffi_rustbuffer_free().name() }}(rb.self, status)
		return false
	})
}

func goBytesToCRustBuffer(b []byte) C.RustBuffer {
	// We can pass the pointer along here, as it is pinned
	// for the duration of this call
	foreign := C.ForeignBytes {
		len: C.int(len(b)),
		data: (*C.uchar)(unsafe.Pointer(&b)),
	}
	
	return rustCall(func( status *C.RustCallStatus) C.RustBuffer {
		return C.{{ ci.ffi_rustbuffer_from_bytes().name() }}(foreign, status)
	})
}

func cRustBufferToGoBytes(b C.RustBuffer) []byte {
	return C.GoBytes(unsafe.Pointer(b.data), C.int(b.len))
}
