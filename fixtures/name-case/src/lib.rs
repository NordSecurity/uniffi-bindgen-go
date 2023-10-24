/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

include!(concat!(env!("OUT_DIR"), "/name-case.uniffi.rs"));

pub enum ENUMTest {
    VARIANTOne,
}

pub enum AssociatedENUMTest {
    VARIANTTest { code: i16 },
}

#[derive(Debug, thiserror::Error)]
pub enum ERRORTest {
    #[error("Test")]
    VARIANTOne,
}

#[derive(Debug, thiserror::Error)]
pub enum AssociatedERRORTest {
    #[error("Test")]
    VARIANTTest { code: i16 },
}

pub struct OBJECTTest {}

impl OBJECTTest {
    pub fn new() -> Self {
        OBJECTTest {}
    }

    pub fn new_alternate() -> Self {
        OBJECTTest {}
    }

    pub fn test(&self) {}
}

pub struct RECORDTest {
    test: i32,
}

pub fn test() {
    let _ = ERRORTest::VARIANTOne;
}

pub trait CALLBACKTest {
    fn test(&self);
}
