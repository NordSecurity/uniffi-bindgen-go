const uniffiRustTaskCallbackSuccess byte = 0
const uniffiRustTaskCallbackCancelled byte = 1
const uniffiForeignExecutorCallbackSuccess byte = 0
const uniffiForeignExecutorCallbackCanceled byte = 1
const uniffiForeignExecutorCallbackError byte = 2

// Encapsulates an executor that can run Rust tasks
type UniFfiForeignExecutor struct {
	priority int
}


