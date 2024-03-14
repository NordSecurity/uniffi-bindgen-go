/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

#[derive(uniffi::Record)]
struct Record {
    id: String,
    tag: String,
}

// Ensure multiple futures packages work fine together, the other one being
// the "futures" fixture from uniffi-rs.
// https://github.com/NordSecurity/uniffi-bindgen-go/issues/45
#[uniffi::export]
async fn get_async_record() -> Record {
    Record {
        id: "foo".to_string(),
        tag: "bar".to_string(),
    }
}

include!(concat!(env!("OUT_DIR"), "/issue45.uniffi.rs"));
