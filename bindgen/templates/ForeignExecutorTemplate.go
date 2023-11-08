const uniffiRustTaskCallbackSuccess byte = 0
const uniffiRustTaskCallbackCancelled byte = 1
const uniffiForeignExecutorCallbackSuccess byte = 0
const uniffiForeignExecutorCallbackCanceled byte = 1
const uniffiForeignExecutorCallbackError byte = 2

{% if self.include_once_check("CallbackHelpers.go") %}{% include "CallbackHelpers.go" %}{% endif %}
{{- self.add_import("runtime") }}
{{- self.add_import("time") }}

// UniFfiForeignExecutor encapsulates an executor that can run Rust tasks.
type UniFfiForeignExecutor struct {}

type FfiConverterForeignExecutor struct {}
var FfiConverterForeignExecutorINSTANCE = FfiConverterForeignExecutor{}

func (c FfiConverterForeignExecutor) Lower(value UniFfiForeignExecutor) C.int {
	return 0;
}

func (c FfiConverterForeignExecutor) Write(writer io.Writer, value UniFfiForeignExecutor) {
	writeUint64(writer, uint64(c.Lower(value)))
}

func (c FfiConverterForeignExecutor) Lift(value C.int) UniFfiForeignExecutor {
	if value != 0 {
		panic(fmt.Errorf("Invalid executor pointer: %d", value))
	}
	return UniFfiForeignExecutor{}
}

func (c FfiConverterForeignExecutor) Read(reader io.Reader) UniFfiForeignExecutor {
	return c.Lift(C.int(readUint64(reader)))
}


//export uniffiForeignExecutorCallback{{config.package_name.as_ref().unwrap()}}
func uniffiForeignExecutorCallback{{config.package_name.as_ref().unwrap()}}(executor C.uint64_t, delay C.uint32_t, task C.RustTaskCallback, taskData *C.void) C.int8_t {
	if task != nil {
		_ = FfiConverterForeignExecutorINSTANCE.Lift(C.int(executor))
		go func() {
			if delay > 0 {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			} else {
				runtime.Gosched()
			}

			C.cgo_rust_task_callback_bridge_{{ config.package_name.as_ref().unwrap() }}(
				C.RustTaskCallback(unsafe.Pointer(task)),
				unsafe.Pointer(taskData),
				C.int8_t(uniffiCallbackResultSuccess),
			)
		}()
		return C.int8_t(uniffiCallbackResultSuccess)
	} else {
		// Drop the executor
		// nothing to do at the moment
		return C.int8_t(uniffiIdxCallbackFree)
	}
}

func uniffiInitForeignExecutor() {
	// Register the callback
	{%- match ci.ffi_foreign_executor_callback_set() %}
	{%- when Some with (fn) %}
	rustCall(func(status *C.RustCallStatus) bool {
		C.{{ fn.name() }}(C.ForeignExecutorCallback(C.uniffiForeignExecutorCallback{{config.package_name.as_ref().unwrap()}}), status)
		return false
	})
	{%- when None %}
	{#- No foreign executor, we dont set anything #}
        {% endmatch %}
}

type {{ type_|ffi_destroyer_name }} struct {}

func ({{ type_|ffi_destroyer_name }}) Destroy(_ {{ type_name }}) {}

