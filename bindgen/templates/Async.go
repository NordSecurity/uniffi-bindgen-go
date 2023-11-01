{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

const (
	uniffiRustFuturePollReady      C.int8_t = 0
	uniffiRustFuturePollMaybeReady C.int8_t = 1
)

func uniffiRustCallAsync(
	rustFutureFunc func(*C.RustCallStatus) *C.void,
	pollFunc func(*C.void, unsafe.Pointer, *C.RustCallStatus),
	completeFunc func(*C.void, *C.RustCallStatus),
	freeFunc func(*C.void, *C.RustCallStatus),
) {
	rustFuture, err := uniffiRustCallAsyncInner(nil, rustFutureFunc, pollFunc, freeFunc)
	if err != nil {
		panic(err)
	}
	defer func() {
		var status C.RustCallStatus
		freeFunc(rustFuture, &status)
		err := checkCallStatus(nil, status)
		if err != nil {
			panic(fmt.Sprintf("unexpected freeFunc error %v", err))
		}
	}()

	var status C.RustCallStatus
	completeFunc(rustFuture, &status)
	err = checkCallStatus(nil, status)
	if err != nil {
		panic(err)
	}
}

func uniffiRustCallAsyncWithResult[T any](
	rustFutureFunc func(*C.RustCallStatus) *C.void,
	pollFunc func(*C.void, unsafe.Pointer, *C.RustCallStatus),
	completeFunc func(*C.void, *C.RustCallStatus) T,
	freeFunc func(*C.void, *C.RustCallStatus),
) T {
	rustFuture, err := uniffiRustCallAsyncInner(nil, rustFutureFunc, pollFunc, freeFunc)
	if err != nil {
		panic(err)
	}

	defer func() {
		var status C.RustCallStatus
		freeFunc(rustFuture, &status)
		err := checkCallStatus(nil, status)
		if err != nil {
			panic(fmt.Sprintf("unexpected freeFunc error %v", err))
		}
	}()

	var status C.RustCallStatus
	returnValue := completeFunc(rustFuture, &status)
	err = checkCallStatus(nil, status)
	if err != nil {
		panic(err)
	}
	return returnValue
}

func uniffiRustCallAsyncWithError(
	converter BufLifter[error],
	rustFutureFunc func(*C.RustCallStatus) *C.void,
	pollFunc func(*C.void, unsafe.Pointer, *C.RustCallStatus),
	completeFunc func(*C.void, *C.RustCallStatus),
	freeFunc func(*C.void, *C.RustCallStatus),
) error {
	rustFuture, err := uniffiRustCallAsyncInner(converter, rustFutureFunc, pollFunc, freeFunc)
	if err != nil {
		return err
	}

	defer func() {
		var status C.RustCallStatus
		freeFunc(rustFuture, &status)
		err := checkCallStatus(converter, status)
		if err != nil {
			panic(fmt.Sprintf("unexpected freeFunc error %v", err))
		}
	}()

	var status C.RustCallStatus
	completeFunc(rustFuture, &status)
	err = checkCallStatus(converter, status)
	return err
}

func uniffiRustCallAsyncWithErrorAndResult[T any](
	converter BufLifter[error],
	rustFutureFunc func(*C.RustCallStatus) *C.void,
	pollFunc func(*C.void, unsafe.Pointer, *C.RustCallStatus),
	completeFunc func(*C.void, *C.RustCallStatus) T,
	freeFunc func(*C.void, *C.RustCallStatus),
) (T, error) {
	var returnValue T
	rustFuture, err := uniffiRustCallAsyncInner(converter, rustFutureFunc, pollFunc, freeFunc)
	if err != nil {
		return returnValue, err
	}

	defer func() {
		var status C.RustCallStatus
		freeFunc(rustFuture, &status)
		err := checkCallStatus(converter, status)
		if err != nil {
			panic(fmt.Sprintf("unexpected freeFunc error %v", err))
		}
	}()

	var status C.RustCallStatus
	returnValue = completeFunc(rustFuture, &status)
	err = checkCallStatus(converter, status)
	return returnValue, err
}

func uniffiRustCallAsyncInner(
	converter BufLifter[error],
	rustFutureFunc func(*C.RustCallStatus) *C.void,
	pollFunc func(*C.void, unsafe.Pointer, *C.RustCallStatus),
	freeFunc func(*C.void, *C.RustCallStatus),
) (*C.void, error) {
	pollResult := C.int8_t(-1)
	waiter := make(chan C.int8_t, 1)
	chanHandle := cgo.NewHandle(waiter)
	var status C.RustCallStatus

	rustFuture := rustFutureFunc(&status)
	err := checkCallStatus(converter, status)
	if err != nil {
		return nil, err
	}

	defer chanHandle.Delete()

	for pollResult != uniffiRustFuturePollReady {
		ptr := unsafe.Pointer(&chanHandle)
		pollFunc(rustFuture, ptr, &status)
		err = checkCallStatus(converter, status)
		if err != nil {
			return nil, err
		}
		res := <-waiter
		pollResult = res
	}

	return rustFuture, nil
}

// Callback handlers for an async calls.  These are invoked by Rust when the future is ready.  They
// lift the return value or error and resume the suspended function.

//export uniffiFutureContinuationCallback{{ config.package_name.as_ref().unwrap() }}
func uniffiFutureContinuationCallback{{ config.package_name.as_ref().unwrap() }}(ptr unsafe.Pointer, pollResult C.int8_t) {
	doneHandle := *(*cgo.Handle)(ptr)
	done := doneHandle.Value().((chan C.int8_t))
	done <- pollResult
}

func uniffiInitContinuationCallback() {
	rustCall(func(uniffiStatus *C.RustCallStatus) bool {
		C.{{ ci.ffi_rust_future_continuation_callback_set().name() }}(
			C.RustFutureContinuation(C.uniffiFutureContinuationCallback{{config.package_name.as_ref().unwrap()}}),
			uniffiStatus,
		)
		if uniffiStatus != nil {
			err := checkCallStatusUnknown(*uniffiStatus)
			if err != nil {
				panic(fmt.Errorf("Failed to initalize RustFutureContinuation %v", err))
			}
		}
		return false
	})
}
