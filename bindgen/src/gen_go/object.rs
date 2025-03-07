/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

use uniffi_bindgen::{backend::Literal, ComponentInterface};
use uniffi_meta::ObjectImpl;

use super::{filters::oracle, CodeType};

#[derive(Debug)]
pub struct ObjectCodeType {
    id: String,
    imp: ObjectImpl,
}

impl ObjectCodeType {
    pub fn new(id: String, imp: ObjectImpl) -> Self {
        Self { id, imp }
    }
}

impl CodeType for ObjectCodeType {
    fn type_label(&self, _ci: &ComponentInterface) -> String {
        if self.imp.has_callback_interface() {
            // When object has callback interface, it is represented
            // as interface, that is already a fat pointer
            oracle().class_name(&self.id)
        } else {
            format!("*{}", super::GoCodeOracle.class_name(&self.id))
        }
    }

    fn canonical_name(&self) -> String {
        oracle().class_name(&self.id)
    }

    fn literal(&self, _literal: &Literal, _ci: &ComponentInterface) -> String {
        unreachable!();
    }

    fn initialization_fn(&self) -> Option<String> {
        self.imp
            .has_callback_interface()
            .then(|| format!("{}.register", self.ffi_converter_instance()))
    }
}
