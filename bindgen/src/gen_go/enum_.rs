/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

use uniffi_bindgen::{backend::Literal, ComponentInterface};

use super::{filters::oracle, CodeType};

#[derive(Debug)]
pub struct EnumCodeType {
    name: String,
}

impl EnumCodeType {
    pub fn new(name: String) -> Self {
        Self { name }
    }
}

impl CodeType for EnumCodeType {
    fn type_label(&self, ci: &ComponentInterface) -> String {
        let name = self.canonical_name();
        if ci.is_name_used_as_error(&self.name) {
            format!("*{name}")
        } else {
            name
        }
    }

    fn canonical_name(&self) -> String {
        oracle().class_name(&self.name)
    }

    fn literal(&self, literal: &Literal, ci: &ComponentInterface) -> String {
        if let Literal::Enum(v, _) = literal {
            format!(
                "{}.{}",
                self.type_label(ci),
                super::GoCodeOracle.enum_variant_name(v)
            )
        } else {
            unreachable!();
        }
    }
}
