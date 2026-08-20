# Configuration options

Settings come from each crate's own `uniffi.toml`, under a `bindings.go` section. All
configuration keys below are defined there.

A crate's `uniffi.toml` is picked up automatically. To set values across crates, pass a
global config file with `--config`, which nests the same keys under `[defaults]` (applied
under every crate's own config) or `[crates.<name>]` (applied over it).
```bash
uniffi-bindgen-go path/to/definitions.udl --config path/to/global.toml
```
```toml
# global.toml
[defaults.bindings.go]
go_mod = "github.com/example/generated"

[crates.my_crate.bindings.go]
package_name = "mycrate"
```

- `package_name` - override the go package name.

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
