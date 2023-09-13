/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

use uniffi_bindgen::backend::{CodeType, Literal};

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
    fn type_label(&self) -> String {
        super::GoCodeOracle.class_name(&self.name)
    }

    fn canonical_name(&self) -> String {
        format!("Type{}", self.name)
    }

    fn literal(&self, literal: &Literal) -> String {
        if let Literal::Enum(v, _) = literal {
            format!(
                "{}.{}",
                self.type_label(),
                super::GoCodeOracle.enum_variant_name(v)
            )
        } else {
            unreachable!();
        }
    }
}
