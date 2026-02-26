/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

use uniffi_bindgen::ComponentInterface;

use super::CodeType;

#[derive(Debug)]
pub struct ExternalCodeType {
    name: String,
    namespace: String,
    is_object: bool,
}

impl ExternalCodeType {
    pub fn new(name: String, namespace: String, is_object: bool) -> Self {
        ExternalCodeType { name, namespace, is_object }
    }
}

impl CodeType for ExternalCodeType {
    fn type_label(&self, _ci: &ComponentInterface) -> String {
        let label = format!("{}.{}", self.namespace, self.name);
        if self.is_object {
            format!("*{}", label)
        } else {
            label
        }
    }

    fn canonical_name(&self) -> String {
        self.name.clone()
    }

    fn ffi_converter_name(&self) -> String {
        format!("{}.FfiConverter{}", self.namespace, self.canonical_name())
    }

    fn ffi_destroyer_name(&self) -> String {
        format!("{}.FfiDestroyer{}", self.namespace, self.canonical_name())
    }

    fn requires_lower_external(&self) -> bool {
        !self.is_object
    }
}
