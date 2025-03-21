/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

use uniffi_bindgen::{backend::Literal, ComponentInterface};

use super::CodeType;

#[derive(Debug)]
pub struct RecordCodeType {
    name: String,
}

impl RecordCodeType {
    pub fn new(name: String) -> Self {
        Self { name }
    }
}

impl CodeType for RecordCodeType {
    fn type_label(&self, _ci: &ComponentInterface) -> String {
        super::GoCodeOracle.class_name(&self.name)
    }

    fn canonical_name(&self) -> String {
        super::GoCodeOracle.class_name(&self.name)
    }

    fn literal(&self, _literal: &Literal, _ci: &ComponentInterface) -> String {
        unreachable!();
    }
}
