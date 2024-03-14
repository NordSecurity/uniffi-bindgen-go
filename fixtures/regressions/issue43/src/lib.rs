/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

#[uniffi::export]
async fn get_async_external_type() -> uniffi_go_issue45::Record {
    uniffi_go_issue45::Record {
        id: "foo".to_string(),
        tag: "bar".to_string(),
    }
}

include!(concat!(env!("OUT_DIR"), "/issue43.uniffi.rs"));
