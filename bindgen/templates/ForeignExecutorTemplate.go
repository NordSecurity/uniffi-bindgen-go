const uniffiRustTaskCallbackSuccess byte = 0
const uniffiRustTaskCallbackCancelled byte = 1
const uniffiForeignExecutorCallbackSuccess byte = 0
const uniffiForeignExecutorCallbackCanceled byte = 1
const uniffiForeignExecutorCallbackError byte = 2

{% if self.include_once_check("CallbackInterfaceRuntime.go") %}{% include "CallbackInterfaceRuntime.go" %}{% endif %}
{{- self.add_import("sync") }}

// Encapsulates an executor that can run Rust tasks
type UniFfiForeignExecutor struct {
	inner InnerExecutor
}

type InnerExecutor struct {}

func (c UniFfiForeignExecutor) Lower(value UniFfiForeignExecutor) C.int {
	return FfiConverterForeignExecutorINSTANCE.Lower(value)
}

func (c UniFfiForeignExecutor) Write(writer io.Writer, value UniFfiForeignExecutor) {
	FfiConverterForeignExecutorINSTANCE.Write(writer, value)
}

func (c UniFfiForeignExecutor) Lift(value C.int) UniFfiForeignExecutor {
	return FfiConverterForeignExecutorINSTANCE.Lift(value)
}

func (c UniFfiForeignExecutor) Read(reader io.Reader) UniFfiForeignExecutor {
	return FfiConverterForeignExecutorINSTANCE.Read(reader)
}



type FfiConverterForeignExecutor struct {
	handleMap *concurrentHandleMap[InnerExecutor]
}

func (c *FfiConverterForeignExecutor) drop(handle uint64) RustBuffer {
	c.handleMap.remove(handle)
	return RustBuffer{}
}

func (c *FfiConverterForeignExecutor) Lift(handle C.int) UniFfiForeignExecutor {
	val, ok := c.handleMap.tryGet(uint64(handle))
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}
	inner := *val
	return UniFfiForeignExecutor{
		inner,
	}
}

func (c *FfiConverterForeignExecutor) Read(reader io.Reader) UniFfiForeignExecutor {
	return c.Lift(C.int(readUint64(reader)))
}

func (c *FfiConverterForeignExecutor) Lower(value UniFfiForeignExecutor) C.int {
	return C.int(c.handleMap.insert(&value.inner))
}

func (c *FfiConverterForeignExecutor) Write(writer io.Writer, value UniFfiForeignExecutor) {
	writeUint64(writer, uint64(c.Lower(value)))
}

var FfiConverterForeignExecutorINSTANCE = FfiConverterForeignExecutor{
	handleMap: newConcurrentHandleMap[InnerExecutor](),
}

//export uniffiForeignExecutorCallback
func uniffiForeignExecutorCallback(executor C.uint64_t, delay C.uint32_t, task C.RustTaskCallback, taskData *C.void) C.int8_t {
	fmt.Printf("Executor callback called\n")
	ex := FfiConverterForeignExecutorINSTANCE.Lift(C.int(executor))
	fmt.Printf("Executor callback with %v\n", ex)
	// TODO: converte the handle into the correct instance
	// trigger call
        return 0
}

func uniffiInitForeignExecutor() {
	// Register the callback
	{%- match ci.ffi_foreign_executor_callback_set() %}
	{%- when Some with (fn) %}
	rustCall(func(uniffiStatus *C.RustCallStatus) bool {
		C.{{ fn.name() }}(C.ForeignExecutorCallback(C.uniffiForeignExecutorCallback), uniffiStatus)
		// TODO: handle error
		return false
	})
	fmt.Printf("Registered callback handler\n")
	{%- when None %}
	{#- No foreign executor, we dont set anything #}
        {% endmatch %}
}

type {{ type_|ffi_destroyer_name }} struct {}

func ({{ type_|ffi_destroyer_name }}) Destroy(_ {{ type_name }}) {}

