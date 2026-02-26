/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

use uniffi_bindgen::{
    interface::{FfiType, Type},
    ComponentInterface,
};

use super::{filters::oracle, CodeType};

#[derive(Debug)]
pub struct ExternalCodeType {
    name: String,
    namespace: String,
    kind: ExternalKind,
    is_error: bool,
    custom_builtin: Option<Type>,
}

#[derive(Debug, Clone, Copy)]
struct CustomFfiType {
    plain: &'static str,
    cgo: &'static str,
    needs_lower_external: bool,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ExternalKind {
    Value,
    Object,
    Interface,
}

impl ExternalCodeType {
    pub fn new(
        name: String,
        namespace: String,
        kind: ExternalKind,
        is_error: bool,
        custom_builtin: Option<Type>,
    ) -> Self {
        ExternalCodeType {
            name,
            namespace,
            kind,
            is_error,
            custom_builtin,
        }
    }

    fn rendered_type_label(&self) -> String {
        let label = format!("{}.{}", self.namespace, self.canonical_name());
        match self.kind {
            ExternalKind::Value | ExternalKind::Interface if self.is_error => format!("*{}", label),
            ExternalKind::Value | ExternalKind::Interface => label,
            ExternalKind::Object => format!("*{}", label),
        }
    }

    fn helper_stem(&self) -> String {
        if self.custom_builtin.is_some() {
            format!("Type{}", self.canonical_name())
        } else {
            self.canonical_name()
        }
    }

    fn custom_ffi_type(&self) -> Option<CustomFfiType> {
        let builtin = self.custom_builtin.as_ref()?;
        Some(match FfiType::from(builtin) {
            FfiType::Int8 => CustomFfiType {
                plain: "int8",
                cgo: "C.int8_t",
                needs_lower_external: false,
            },
            FfiType::UInt8 => CustomFfiType {
                plain: "uint8",
                cgo: "C.uint8_t",
                needs_lower_external: false,
            },
            FfiType::Int16 => CustomFfiType {
                plain: "int16",
                cgo: "C.int16_t",
                needs_lower_external: false,
            },
            FfiType::UInt16 => CustomFfiType {
                plain: "uint16",
                cgo: "C.uint16_t",
                needs_lower_external: false,
            },
            FfiType::Int32 => CustomFfiType {
                plain: "int32",
                cgo: "C.int32_t",
                needs_lower_external: false,
            },
            FfiType::UInt32 => CustomFfiType {
                plain: "uint32",
                cgo: "C.uint32_t",
                needs_lower_external: false,
            },
            FfiType::Int64 => CustomFfiType {
                plain: "int64",
                cgo: "C.int64_t",
                needs_lower_external: false,
            },
            FfiType::UInt64 => CustomFfiType {
                plain: "uint64",
                cgo: "C.uint64_t",
                needs_lower_external: false,
            },
            FfiType::Float32 => CustomFfiType {
                plain: "float32",
                cgo: "C.float",
                needs_lower_external: false,
            },
            FfiType::Float64 => CustomFfiType {
                plain: "float64",
                cgo: "C.double",
                needs_lower_external: false,
            },
            FfiType::RustBuffer(_) => CustomFfiType {
                plain: "RustBufferI",
                cgo: "RustBufferI",
                needs_lower_external: true,
            },
            _ => return None,
        })
    }
}

impl CodeType for ExternalCodeType {
    fn type_label(&self, _ci: &ComponentInterface) -> String {
        self.rendered_type_label()
    }

    fn canonical_name(&self) -> String {
        oracle().class_name(&self.name)
    }

    fn ffi_converter_name(&self) -> String {
        format!("{}.FfiConverter{}", self.namespace, self.helper_stem())
    }

    fn ffi_destroyer_name(&self) -> String {
        format!("{}.FfiDestroyer{}", self.namespace, self.helper_stem())
    }

    fn lower(&self) -> String {
        if let Some(ffi_type) = self.custom_ffi_type() {
            return format!(
                "func(value {}) {} {{ return {}({}.LowerToExternal{}(value)) }}",
                self.rendered_type_label(),
                ffi_type.cgo,
                ffi_type.cgo,
                self.namespace,
                self.helper_stem(),
            );
        }
        match self.kind {
            ExternalKind::Value => format!("{}.Lower", self.ffi_converter_instance()),
            ExternalKind::Object | ExternalKind::Interface => {
                let type_label = self.rendered_type_label();
                format!(
                    "func(value {type_label}) C.uint64_t {{ return C.uint64_t({}.LowerToExternal{}(value)) }}",
                    self.namespace,
                    self.helper_stem()
                )
            }
        }
    }

    fn lift(&self) -> String {
        if let Some(ffi_type) = self.custom_ffi_type() {
            return format!(
                "func(value {}) {} {{ return {}.LiftFromExternal{}({}(value)) }}",
                ffi_type.cgo,
                self.rendered_type_label(),
                self.namespace,
                self.helper_stem(),
                ffi_type.plain,
            );
        }
        match self.kind {
            ExternalKind::Value => format!("{}.Lift", self.ffi_converter_instance()),
            ExternalKind::Object | ExternalKind::Interface => {
                let type_label = self.rendered_type_label();
                format!(
                    "func(handle C.uint64_t) {type_label} {{ return {}.LiftFromExternal{}(uint64(handle)) }}",
                    self.namespace,
                    self.helper_stem()
                )
            }
        }
    }

    fn lower_external(&self) -> String {
        if self.custom_builtin.is_some() {
            return format!("{}.LowerToExternal{}", self.namespace, self.helper_stem());
        }
        format!("{}.LowerExternal", self.ffi_converter_instance())
    }

    fn requires_lower_external(&self) -> bool {
        self.custom_builtin
            .as_ref()
            .and_then(|_| {
                self.custom_ffi_type()
                    .map(|ffi_type| ffi_type.needs_lower_external)
            })
            .unwrap_or(self.kind == ExternalKind::Value)
    }
}
