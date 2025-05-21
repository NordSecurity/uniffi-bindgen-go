#[uniffi::export]
fn empty_string_test() -> Option<String> {
    Some(String::new())
}

#[uniffi::export]
fn empty_bytes_test() -> Option<Vec<u8>> {
    Some(vec![])
}

uniffi::setup_scaffolding!();
