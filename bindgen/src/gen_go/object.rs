/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

use uniffi_bindgen::backend::{CodeType, Literal};

#[derive(Debug)]
pub struct ObjectCodeType {
    name: String,
}

impl ObjectCodeType {
    pub fn new(name: String) -> Self {
        Self { name }
    }
}

impl CodeType for ObjectCodeType {
    fn type_label(&self) -> String {
        format!("*{}", super::GoCodeOracle.class_name(&self.name))
    }

    fn canonical_name(&self) -> String {
        self.name.clone()
    }

    fn literal(&self, _literal: &Literal) -> String {
        unreachable!();
    }
}
