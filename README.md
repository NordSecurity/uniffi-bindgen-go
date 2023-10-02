# uniffi-bindgen-go - UniFFI Go bindings generator

Generate [UniFFI](https://github.com/mozilla/uniffi-rs) bindings for Go. `uniffi-bindgen-go` lives
as a separate project from `uniffi-go`, as per
[uniffi-rs #1355](https://github.com/mozilla/uniffi-rs/issues/1355). Currently, `uniffi-bindgen-go`
uses `uniffi-rs` version `0.23.0`.

# How to install

Minimum Rust version required to install `uniffi-bindgen-go` is `1.64`.
Newer Rust versions should also work fine.

```
cargo install uniffi-bindgen-go --git https://github.com/NordSecurity/uniffi-bindgen-go
```

# How to generate bindings

```
uniffi-bindgen-go path/to/definitions.udl
```
Generates bindings file `path/to/uniffi/definitions/definitions.go`

# How to integrate bindings

To integrate the bindings into your projects, simply add the generated bindings file to your project.
Generated bindings require Go 1.19 or later to compile.


# Configuration options

It's possible to configure some settings by passing `--config` argument to the generator. All
configuration keys are defined in `bindings.go` section.
```bash
uniffi-bindgen-go path/to/definitions.udl --config path/to/uniffi.toml
```

- `package_name` - override the go package name.

- `cdylib_name` - override the dynamic library name linked by generated bindings, excluding `lib`
    prefix and `.dll` file extension.
    ```toml
    # For library `libgreeter.dll`
    [bindings.go]
    cdylib_name = "greeter"
    ```

- `custom_types` - properties for custom type defined in UDL with `[Custom] typedef string Url;`.
    ```toml
    # Represent URL as a native Go `url.Url`. The underlying type of URL is a string.
    imports = ["net/url"]
    type_name = "url.URL"
    into_custom = """u, err := url.Parse({})
        if err != nil {
             panic(err)
        }
        return *u
    """
    from_custom = "{}.String()"
    ```

    - `imports` (optional) - any imports required to satisfy this type.

    - `type_name` (optional) - the name to represent the type in generated bindings. Default is the
        type alias name from UDL, e.g. `Url`.

    - `into_custom` (required) - an expression to convert from the underlying type into custom type. `{}` will
        will be expanded into variable containing the underlying value. The expression is used in a
        return statement, i.e. `return <expression(value)>;`.

    - `from_custom` (required) - an expression to convert from the custom type into underlying type. `{}` will
        will be expanded into variable containing the custom value. The expression is used in a
        return statement, i.e. `return <expression(value);>`.

- `go_mod` (optional) - Specify the go module for the final package, used as imports source for external types.

- `c_module_filename`(optional) - override the name of the `C` module (`.h` and `.c`)

# Contributing

For contribution guidelines, read [CONTRIBUTING.md](CONTRIBUTING.md)

# Versioning

`uniffi-bindgen-go` is versioned separately from `uniffi-rs`. UniFFI follows the [SemVer rules from
the Cargo Book](https://doc.rust-lang.org/cargo/reference/resolver.html#semver-compatibility)
which states "Versions are considered compatible if their left-most non-zero
major/minor/patch component is the same". A breaking change is any modification to the Go bindings
that demands the consumer of the bindings to make corresponding changes to their code to ensure that
the bindings continue to function properly. `uniffi-bindgen-go` is young, and its unclear how stable
the generated bindings are going to be between versions. For this reason, major version is currently
0, and most changes are probably going to bump minor version.

To ensure consistent feature set across external binding generators, `uniffi-bindgen-go` targets
a specific `uniffi-rs` version. A consumer using Go bindings (in `uniffi-bindgen-go`) and Go
bindings (in `uniffi-bindgen-go`) expects the same features to be available across multiple bindings
generators. This means that the consumer should choose external binding generator versions such that
each generator targets the same `uniffi-rs` version.

To simplify this choice `uniffi-bindgen-cs` and `uniffi-bindgen-go` use tag naming convention
as follows: `vX.Y.Z+vA.B.C`, where `X.Y.Z` is the version of the generator itself, and `A.B.C` is
the version of uniffi-rs it is based on.
