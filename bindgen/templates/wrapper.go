{#/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */#}

package {{ ci.namespace() }}

// #include <{{ config.header_filename() }}>
import "C"

import (
	"bytes"
	"fmt"
	"io"
	"unsafe"
	"encoding/binary"
	{%- if ci.has_async_fns() %}
	"runtime/cgo"
	{%- endif %}
	{%- for imported_package in self.imports() %}
	{{ imported_package.render() }}
	{%- endfor %}
)

{% include "RustBufferTemplate.go" %}
{% include "FfiConverterTemplate.go" %}
{% include "Helpers.go" %}
{% include "BinaryWrite.go" %}
{% include "BinaryRead.go" %}

{% include "NamespaceLibraryTemplate.go" %}

{{ type_helper_code }}

{%- if ci.has_async_fns() %}
{% include "Async.go" %}
{%- endif %}

{%- for func in ci.function_definitions() %}
{% include "TopLevelFunctionTemplate.go" %}
{%- endfor %}

{% import "macros.go" as go %}
