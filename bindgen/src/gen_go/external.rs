/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

use uniffi_bindgen::{backend::CodeType, interface::ExternalKind};

#[derive(Debug)]
pub struct ExternalCodeType {
    name: String,
    module_path: String,
    kind: ExternalKind,
}

impl ExternalCodeType {
    pub fn new(name: String, module_path: String, kind: ExternalKind) -> Self {
        ExternalCodeType {
            name,
            module_path,
            kind,
        }
    }
}

impl CodeType for ExternalCodeType {
    fn type_label(&self) -> String {
        match self.kind {
            ExternalKind::DataClass => format!("{}.{}", self.module_path, self.name),
            ExternalKind::Interface => format!("*{}.{}", self.module_path, self.name),
        }
    }

    fn canonical_name(&self) -> String {
        match self.kind {
            ExternalKind::DataClass => format!("Type{}", self.name),
            ExternalKind::Interface => self.name.clone(),
        }
    }

    fn ffi_converter_name(&self) -> String {
        format!("{}.FfiConverter{}", self.module_path, self.canonical_name())
    }
}
