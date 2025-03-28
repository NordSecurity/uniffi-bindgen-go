### v0.3.0+v0.28.3

- **BREAKING** Upgrade to v0.28.3
  - Support for new WithForeign Traits
  - Support for object's as Error
  - Support for Async traits
  - All uniffi trait's get generated methods (String, DebugString, Hash, Eq, Ne)
  - #[repr(x)] Now represented more closely to underlying data
  - Support for Tuple Enums and Erros added
- **BREAKING** Error's are now returned as concreate types,
not an error interface. For convience all error types have AsError
to simplify convertion, avoiding footgun of treating nil pointer as non nil
error interface
- **BREAKING** Dropped support for ExternalTypes


----

### v0.2.2+v0.25.0

- Fix multiple async packages not working together [#45](https://github.com/NordSecurity/uniffi-bindgen-go/issues/45).
- Add package name to `RustBufferI` if it needs it [#43](https://github.com/NordSecurity/uniffi-bindgen-go/issues/43).
- Fix several issues with the error interface [#36](https://github.com/NordSecurity/uniffi-bindgen-go/issues/36).
- Cleanup callback code.

### v0.2.1+v0.25.0

- Fix generated code when `error` field name is used in an associated error variant
- Implement docstrings

### v0.2.0+v0.25.0

- **BREAKING**: Update to uniffi 0.25.0.
- **IMPORTANT**: Fix race condition in callback handling code [#28](https://github.com/NordSecurity/uniffi-bindgen-go/issues/28).
- Implement `--library-mode` command line option.
- Implement async functions and methods.
- implement foreign executor.
- Implement `bytes` type.
- Implement external types.
- Fix incorrect code emitted for all caps acronyms in objects and callbacks, e.g. `HTTPClient`.

----

### v0.1.5+v0.23.0

- **IMPORTANT**: Fix memory leak for all strings being read from FFI.

### v0.1.4+v0.23.0

- Fix typo in generated Go bindings for associated enum case with no fields.

### v0.1.3+v0.23.0

- Closing generated binding file before formatting.
- Removed unnecessery import from EnumTemplate.go.

### v0.1.2+v0.23.0

- Fix 0.1 release to be compatible with mozilla/uniffi-rs 0.23.0 after docstring changes.

### v0.1.1+v0.23.0

- Changed callback return type to `C.uint64_t`.

### v0.1.0+v0.23.0

- Updated version tag pattern.

----
