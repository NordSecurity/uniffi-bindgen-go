/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

use paste::paste;
use uniffi_bindgen::{backend::Literal, interface::Type, ComponentInterface};

use super::CodeType;

fn render_literal(literal: &Literal, inner: &Type, ci: &ComponentInterface) -> String {
    match literal {
        Literal::None => "nil".into(),

        // For optionals
        _ => super::GoCodeOracle.find(inner, ci).literal(literal, ci),
    }
}

macro_rules! impl_code_type_for_compound {
     ($T:ty, $type_label_pattern:literal, $canonical_name_pattern: literal) => {
         paste! {
             #[derive(Debug)]
             pub struct $T<'a> {
                 inner: Type,
                 ci: &'a ComponentInterface,
             }

             impl<'a> $T<'a> {
                 pub fn new(inner: Type, ci: &'a ComponentInterface) -> Self {
                     Self { inner, ci }
                 }
                 fn inner(&self) -> &Type {
                     &self.inner
                 }
             }

             impl<'a> CodeType for $T<'a>  {
                 fn type_label(&self, ci: &ComponentInterface) -> String {
                     format!($type_label_pattern, $crate::gen_go::GoCodeOracle.find(self.inner(), ci).type_label(ci))
                 }

                 fn canonical_name(&self) -> String {
                     format!($canonical_name_pattern, $crate::gen_go::GoCodeOracle.find(self.inner(), self.ci).canonical_name())
                 }

                 fn literal(&self, literal: &Literal, ci: &ComponentInterface) -> String {
                     render_literal(literal, self.inner(), ci)
                 }
             }
         }
     }
}

impl_code_type_for_compound!(OptionalCodeType, "*{}", "Optional{}");
impl_code_type_for_compound!(SequenceCodeType, "[]{}", "Sequence{}");

#[derive(Debug)]
pub struct MapCodeType<'a> {
    key: Type,
    value: Type,
    ci: &'a ComponentInterface,
}

impl<'a> MapCodeType<'a> {
    pub fn new(key: Type, value: Type, ci: &'a ComponentInterface) -> Self {
        Self { key, value, ci }
    }

    fn key(&self) -> &Type {
        &self.key
    }

    fn value(&self) -> &Type {
        &self.value
    }
}

impl<'a> CodeType for MapCodeType<'a> {
    fn type_label(&self, ci: &ComponentInterface) -> String {
        format!(
            "map[{}]{}",
            super::GoCodeOracle.find(self.key(), ci).type_label(ci),
            super::GoCodeOracle.find(self.value(), ci).type_label(ci),
        )
    }

    fn canonical_name(&self) -> String {
        format!(
            "Map{}{}",
            super::GoCodeOracle
                .find(self.key(), self.ci)
                .canonical_name(),
            super::GoCodeOracle
                .find(self.value(), self.ci)
                .canonical_name(),
        )
    }

    fn literal(&self, literal: &Literal, ci: &ComponentInterface) -> String {
        render_literal(literal, &self.value, ci)
    }
}
