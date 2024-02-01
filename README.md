# uniffi-bindgen-go - UniFFI Go bindings generator

Generate [UniFFI](https://github.com/mozilla/uniffi-rs) bindings for Go. `uniffi-bindgen-go` lives
as a separate project from `uniffi-go`, as per
[uniffi-rs #1355](https://github.com/mozilla/uniffi-rs/issues/1355). Currently, `uniffi-bindgen-go`
uses `uniffi-rs` version `0.25.0`.

# How to install

Minimum Rust version required to install `uniffi-bindgen-go` is `1.70`.
Newer Rust versions should also work fine.

```
cargo install uniffi-bindgen-go --git https://github.com/NordSecurity/uniffi-bindgen-go --tag v0.2.1+v0.25.0
```

# How to generate bindings

```
uniffi-bindgen-go path/to/definitions.udl
```

Generates bindings file `path/to/uniffi/definitions/definitions.go`

# Linking 

You will need to make sure the compiled Rust binaries are in your `LD_LIBRARY_PATH` (set this to `target/release/` if you have built with the `--release` flag, for instance).

You can also create a script to properly link these bindings - see `test_bindings.sh` in the root of this repository for an example. 

# How to integrate bindings

To integrate the bindings into your projects, simply add the generated bindings file to your project.
Generated bindings require Go 1.19 or later to compile.


# Configuration options

It's possible to [configure some settings](docs/CONFIGURATION.md) by passing `--config` argument to
the generator
```bash
uniffi-bindgen-go path/to/definitions.udl --config path/to/uniffi.toml
```

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

The table shows `uniffi-rs` version history for tags that were published before tag naming convention described above was introduced.

| uniffi-bindgen-go version                | uniffi-rs version                                |
|------------------------------------------|--------------------------------------------------|
| v0.2.0                                   | v0.25.0                                          |
| v0.1.0                                   | v0.23.0                                          |

# Documentation

More documentation is available in [docs](docs) directory.

# Contributing

For contribution guidelines, read [CONTRIBUTING.md](CONTRIBUTING.md).
