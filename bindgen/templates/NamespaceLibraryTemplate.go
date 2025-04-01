func init() {
        {% let initialization_fns = self.initialization_fns() %}
        {% for func in initialization_fns -%}
        {{ func }}();
        {% endfor -%}

	uniffiCheckChecksums()
}


func uniffiCheckChecksums() {
	// Get the bindings contract version from our ComponentInterface
	bindingsContractVersion := {{ ci.uniffi_contract_version() }}
	// Get the scaffolding contract version by calling the into the dylib
	scaffoldingContractVersion := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint32_t {
		return C.{{ ci.ffi_uniffi_contract_version().name() }}()
	})
	if bindingsContractVersion != int(scaffoldingContractVersion) {
		// If this happens try cleaning and rebuilding your project
		panic("{{ ci.namespace() }}: UniFFI contract version mismatch")
	}
	
	{%- for (name, expected_checksum) in ci.iter_checksums() %}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.{{ name }}()
	})
	if checksum != {{ expected_checksum }} {
		// If this happens try cleaning and rebuilding your project
		panic("{{ ci.namespace() }}: {{ name }}: UniFFI API checksum mismatch")
	}
	}
	{%- endfor %}
}
