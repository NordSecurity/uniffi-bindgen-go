{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

// Callbacks for async functions

// Callback handlers for an async calls. These are invoked by Rust when the future is ready.
// They lift the return value or error and resume the suspended function.
{%- for result_type in ci.iter_async_result_types() %}

//export {{ result_type|future_callback }}
func {{ result_type|future_callback }}(
	rawChan unsafe.Pointer,
	returnValue {{ result_type.future_callback_param().borrow()|ffi_type_name_cgo_safe }},
	status C.RustCallStatus,
) {
	done := *(*chan {{ result_type|future_chan_type }})(rawChan)

	{%- match result_type.throws_type %}
	{%- when Some with (e) %}
	err := checkCallStatus({{ e|ffi_converter_name }}{}, status)
	{%- else %}
	err := checkCallStatusUnknown(status)
	{%- endmatch %}
	if err != nil {
		{%- match result_type.return_type -%}
		{%- when Some with (return_type) -%}
		// {{ format!("{:?}", result_type.return_type) }}
		done <- {{ result_type|future_chan_type }} {
			val: {{ result_type.return_type|default_type }},
			err: err,
		}
		{%- else %}
		done <- {{ result_type|future_chan_type }} {
			err,
		}
		{%- endmatch %}
		return
	}

	{%- match result_type.return_type %}
        {%- when Some(return_type) %}
        done <- {{ result_type|future_chan_type }} {
		val: {{ return_type|lift_fn }}(returnValue),
		err: nil,
	}
        {%- when None %}
        done <- {{ result_type|future_chan_type }} {}
        {%- endmatch %}
}
{%- endfor %}
