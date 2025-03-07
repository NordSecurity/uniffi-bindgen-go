{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

const (
	uniffiRustFuturePollReady      int8 = 0
	uniffiRustFuturePollMaybeReady int8 = 1
)

func uniffiRustCallAsync(
	rustFutureFunc func(*C.RustCallStatus) *C.void,
	pollFunc func(*C.void, unsafe.Pointer, *C.RustCallStatus),
	completeFunc func(*C.void, *C.RustCallStatus),
	_liftFunc func(bool),
	freeFunc func(*C.void, *C.RustCallStatus),
) {
	rustFuture, err := uniffiRustCallAsyncInner(nil, rustFutureFunc, pollFunc, freeFunc)
	if err != nil {
		panic(err)
	}
	defer rustCall(func(status *C.RustCallStatus) int {
		freeFunc(rustFuture, status)
		return 0
	})

	rustCall(func(status *C.RustCallStatus) int {
		completeFunc(rustFuture, status)
		return 0
	})
}

func uniffiRustCallAsyncWithResult[T any, U any](
	rustFutureFunc func(*C.RustCallStatus) *C.void,
	pollFunc func(*C.void, unsafe.Pointer, *C.RustCallStatus),
	completeFunc func(*C.void, *C.RustCallStatus) T,
	liftFunc func(T) U,
	freeFunc func(*C.void, *C.RustCallStatus),
) U {
	rustFuture, err := uniffiRustCallAsyncInner(nil, rustFutureFunc, pollFunc, freeFunc)
	if err != nil {
		panic(err)
	}

	defer rustCall(func(status *C.RustCallStatus) int {
		freeFunc(rustFuture, status)
		return 0
	})

	res := rustCall(func(status *C.RustCallStatus) T {
		return completeFunc(rustFuture, status)
	})
	return liftFunc(res)
}

func uniffiRustCallAsyncWithError[E error](
	converter BufReader[*E],
	rustFutureFunc func(*C.RustCallStatus) *C.void,
	pollFunc func(*C.void, unsafe.Pointer, *C.RustCallStatus),
	completeFunc func(*C.void, *C.RustCallStatus),
	_liftFunc func(bool),
	freeFunc func(*C.void, *C.RustCallStatus),
) *E {
	rustFuture, err := uniffiRustCallAsyncInner(converter, rustFutureFunc, pollFunc, freeFunc)
	if err != nil {
		return err
	}

	defer rustCall(func(status *C.RustCallStatus) int {
		freeFunc(rustFuture, status)
		return 0
	})

	_, err = rustCallWithError(converter, func(status *C.RustCallStatus) int {
		completeFunc(rustFuture, status)	
		return 0
	})
	return err
}

func uniffiRustCallAsyncWithErrorAndResult[E error, T any, U any](
	converter BufReader[*E],
	rustFutureFunc func(*C.RustCallStatus) *C.void,
	pollFunc func(*C.void, unsafe.Pointer, *C.RustCallStatus),
	completeFunc func(*C.void, *C.RustCallStatus) T,
	liftFunc func(T) U,
	freeFunc func(*C.void, *C.RustCallStatus),
) (U, *E) {
	var returnValue U
	rustFuture, err := uniffiRustCallAsyncInner(converter, rustFutureFunc, pollFunc, freeFunc)
	if err != nil {
		return returnValue, err
	}

	defer rustCall(func(status *C.RustCallStatus) int {
		freeFunc(rustFuture, status)
		return 0
	})

	res, err := rustCallWithError(converter, func(status *C.RustCallStatus) T {
		return completeFunc(rustFuture, status)	
	})
	if err != nil {
		return returnValue, err
	}
	return liftFunc(res), nil
}

func uniffiRustCallAsyncInner[E error](
	converter BufReader[*E],
	rustFutureFunc func(*C.RustCallStatus) *C.void,
	pollFunc func(*C.void, unsafe.Pointer, *C.RustCallStatus),
	freeFunc func(*C.void, *C.RustCallStatus),
) (*C.void, *E) {
	pollResult := int8(-1)
	waiter := make(chan int8, 1)
	chanHandle := cgo.NewHandle(waiter)

	rustFuture, err := rustCallWithError(converter, func(status *C.RustCallStatus) *C.void {
		return rustFutureFunc(status)
	})
	if err != nil {
		return nil, err
	}

	defer chanHandle.Delete()

	for pollResult != uniffiRustFuturePollReady {
		ptr := unsafe.Pointer(&chanHandle)
		_, err = rustCallWithError(converter, func(status *C.RustCallStatus) int {
			pollFunc(rustFuture, ptr, status)
			return 0
		})
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
	done := doneHandle.Value().((chan int8))
	done <- int8(pollResult)
}

func uniffiInitContinuationCallback() {
	rustCall(func(uniffiStatus *C.RustCallStatus) bool {
		// TODO(pna): fix this with async
		{#
		C.{{ ci.ffi_rust_future_continuation_callback_set().name() }}(
			C.RustFutureContinuation(C.uniffiFutureContinuationCallback{{config.package_name.as_ref().unwrap()}}),
			uniffiStatus,
		)
		#}
		return false
	})
}
