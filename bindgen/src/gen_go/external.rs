/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

use uniffi_bindgen::{backend::CodeType, interface::ExternalKind};

#[derive(Debug)]
pub struct ExternalCodeType {
    name: String,
    #[allow(dead_code)]
    module_path: String,
    kind: ExternalKind,
    namespace: String,
    #[allow(dead_code)]
    tagged: bool,
}

impl ExternalCodeType {
    pub fn new(
        name: String,
        module_path: String,
        kind: ExternalKind,
        namespace: String,
        tagged: bool,
    ) -> Self {
        ExternalCodeType {
            name,
            module_path,
            kind,
            namespace,
            tagged,
        }
    }
}

impl CodeType for ExternalCodeType {
    fn type_label(&self) -> String {
        match self.kind {
            ExternalKind::DataClass => format!("{}.{}", self.namespace, self.name),
            ExternalKind::Interface => format!("*{}.{}", self.namespace, self.name),
        }
    }

    fn canonical_name(&self) -> String {
        match self.kind {
            ExternalKind::DataClass => format!("Type{}", self.name),
            ExternalKind::Interface => self.name.clone(),
        }
    }

    fn ffi_converter_name(&self) -> String {
        format!("{}.FfiConverter{}", self.namespace, self.canonical_name())
    }
}
