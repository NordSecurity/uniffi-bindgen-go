{%- if let Some(display_fmt) = trait_methods.display_fmt %}
func (_self {{ receiver_type }}) String() string {
	{{ self_binding }} := {{ ffi_converter_instance }}.Lower(_self)
	{% call go::ffi_call_binding(display_fmt, self_binding) %}
}

{%- endif %}
{%- if let Some(debug_fmt) = trait_methods.debug_fmt %}
func (_self {{ receiver_type }}) DebugString() string {
	{{ self_binding }} := {{ ffi_converter_instance }}.Lower(_self)
	{% call go::ffi_call_binding(debug_fmt, self_binding) %}
}

{%- endif %}
{%- if let Some(eq_eq) = trait_methods.eq_eq %}
func (_self {{ receiver_type }}) Eq(other {{ receiver_type }}) bool {
	{{ self_binding }} := {{ ffi_converter_instance }}.Lower(_self)
	{% call go::ffi_call_binding(eq_eq, self_binding) %}
}

{%- endif %}
{%- if let Some(eq_ne) = trait_methods.eq_ne %}
func (_self {{ receiver_type }}) Ne(other {{ receiver_type }}) bool {
	{{ self_binding }} := {{ ffi_converter_instance }}.Lower(_self)
	{% call go::ffi_call_binding(eq_ne, self_binding) %}
}

{%- endif %}
{%- if let Some(hash_hash) = trait_methods.hash_hash %}
func (_self {{ receiver_type }}) Hash() uint64 {
	{{ self_binding }} := {{ ffi_converter_instance }}.Lower(_self)
	{% call go::ffi_call_binding(hash_hash, self_binding) %}
}

{%- endif %}
{%- if let Some(ord_cmp) = trait_methods.ord_cmp %}
func (_self {{ receiver_type }}) Cmp(other {{ receiver_type }}) int8 {
	{{ self_binding }} := {{ ffi_converter_instance }}.Lower(_self)
	{% call go::ffi_call_binding(ord_cmp, self_binding) %}
}

{%- endif %}
