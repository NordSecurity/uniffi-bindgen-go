/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

use uniffi_bindgen::ComponentInterface;

use super::CodeType;

#[derive(Debug)]
pub struct ExternalCodeType {
    name: String,
    namespace: String,
}

impl ExternalCodeType {
    pub fn new(name: String, namespace: String) -> Self {
        ExternalCodeType { name, namespace }
    }
}

impl CodeType for ExternalCodeType {
    fn type_label(&self, _ci: &ComponentInterface) -> String {
        format!("{}.{}", self.namespace, self.name)
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
}
