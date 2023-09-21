{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

// Callbacks for async functions

// Callback handlers for an async calls. These are invoked by Rust when the future is ready.
// They lift the return value or error and resume the suspended function.
{%- for result_type in ci.iter_async_result_types() %}

// TODO:
// {{ format!("{:?}", result_type) }}

{%- endfor %}
