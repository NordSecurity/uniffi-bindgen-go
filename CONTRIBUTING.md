# How to submit changes

Pull requests are welcome!

# How to report issues/bugs?

Create an issue on Github, we will try to get back to you ASAP.

# Checkout the code

```
git clone https://github.com/NordSecurity/uniffi-bindgen-go.git
cd uniffi-bindgen-go
git submodule update --init --recursive
```

# Run tests

To run tests, `go` installation is required. Unlike `uniffi-rs`, there is no integration with
`cargo test`. Tests are written using [testing](https://pkg.go.dev/testing) package.

- Build `uniffi-bindgen-go` executable, and `libuniffi_fixtures.so` shared library.
    ```
    ./build.sh
    ```

- Generate test bindings using `uniffi-bindgen-go`, and run `go test` command.
    ```
    ./test_bindings.sh
    ```

# Run tests in Docker

Running tests in Docker containers is easier, because manual `rust`/`go` installations are not required.

```
./docker_build.sh
./docker_test_bindings.sh
```

# Directory structure

| Directory                                | Description                                      |
|------------------------------------------|--------------------------------------------------|
| 3rd-party/uniffi-rs/                     | fork of uniffi-rs, used for tests                |
| binding_tests/generated                  | generated test bindings                          |
| binding_tests/                           | Go tests for bindings                            |
| fixtures/                                | additional test fixtures specific to Go bindings |
| src/gen_go/                              | generator CLI code                               |
| bindgen/templates/                       | generator Go templates                           |


# Thank you!
