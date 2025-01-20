#include <{{ config.header_filename() }}>

// This file exists beacause of
// https://github.com/golang/go/issues/11263
{# TODO(pna): ensure regression was not reintroduced #}
{# void cgo_rust_task_callback_bridge_{{ config.package_name.as_ref().unwrap() }}(RustTaskCallback cb, const void * taskData, int8_t status) {
  cb(taskData, status);
} #}
