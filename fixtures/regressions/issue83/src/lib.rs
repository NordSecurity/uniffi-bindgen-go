/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

use std::sync::Arc;

#[uniffi::trait_interface]
pub trait StringReceiver: Send + Sync {
    fn receive(&self, value: String);
}

/// Hand an empty `String` to a foreign-implemented receiver.
///
/// An empty `String` lowers into a `RustBuffer` whose `data` pointer
/// points to non-null, dangling value (Rust's `NonNull::dangling()`).
/// On the Go side that pointer lives on the cgo callback's stack frame
/// and must not be treated as Go pointer. Otherwise, if the stack is scanned
/// the runtime complains about invalid pointer.
pub fn invoke_string_receiver(receiver: Arc<dyn StringReceiver>) {
    receiver.receive(String::new());
}

include!(concat!(env!("OUT_DIR"), "/issue83.uniffi.rs"));
