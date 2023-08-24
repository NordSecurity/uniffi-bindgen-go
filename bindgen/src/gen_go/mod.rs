/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

use anyhow::{Context, Result};
use askama::Template;
use heck::{ToLowerCamelCase, ToUpperCamelCase};
use serde::{Deserialize, Serialize};
use std::borrow::Borrow;
use std::cell::RefCell;
use std::collections::{BTreeSet, HashMap, HashSet};
use uniffi_bindgen::backend::{CodeType, TemplateExpression};
use uniffi_bindgen::interface::*;

mod callback_interface;
mod compounds;
mod custom;
mod enum_;
mod error;
mod miscellany;
mod object;
mod primitives;
mod record;

#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Config {
    package_name: Option<String>,
    #[serde(default)]
    custom_types: HashMap<String, CustomTypeConfig>,
}

#[derive(Debug, Default, Clone, Serialize, Deserialize)]
pub struct CustomTypeConfig {
    imports: Option<Vec<String>>,
    type_name: Option<String>,
    into_custom: TemplateExpression,
    from_custom: TemplateExpression,
}

impl From<&ComponentInterface> for Config {
    fn from(ci: &ComponentInterface) -> Self {
        Config {
            package_name: Some(format!("{}", ci.namespace())),
            custom_types: HashMap::new(),
        }
    }
}

impl Config {
    pub fn package_name(&self) -> String {
        if let Some(package_name) = &self.package_name {
            package_name.clone()
        } else {
            "uniffi".into()
        }
    }
}

#[derive(Template)]
#[template(syntax = "go", escape = "none", path = "wrapper.go")]
pub struct GoWrapper<'a> {
    config: Config,
    ci: &'a ComponentInterface,
    type_helper_code: String,
    type_imports: BTreeSet<String>,
}

impl<'a> GoWrapper<'a> {
    pub fn new(config: Config, ci: &'a ComponentInterface) -> Self {
        let type_renderer = TypeRenderer::new(&config, ci);
        let type_helper_code = type_renderer.render().unwrap();
        let type_imports = type_renderer.imports.into_inner();
        Self {
            config,
            ci,
            type_helper_code,
            type_imports,
        }
    }

    pub fn imports(&self) -> Vec<String> {
        self.type_imports.iter().cloned().collect()
    }

    pub fn initialization_fns(&self) -> Vec<String> {
        self.ci
            .iter_types()
            .map(|t| GoCodeOracle.find(t))
            .filter_map(|t| t.initialization_fn())
            .collect()
    }

    // This represents true callback functions used in CGo layer. It is this is needed due to
    // https://github.com/golang/go/issues/19837
    pub fn cgo_callback_fns(&self) -> Vec<String> {
        self.ci
            .callback_interface_definitions()
            .iter()
            .map(|d| format!("{}_cgo_{}", self.ci.ffi_namespace(), d.name()))
            .collect()
    }
}

pub fn generate_go_bindings(config: &Config, ci: &ComponentInterface) -> Result<String> {
    GoWrapper::new(config.clone(), ci)
        .render()
        .context("failed to render go bindings")
}

#[derive(Clone)]
pub struct GoCodeOracle;

impl GoCodeOracle {
    // Map `Type` instances to a `Box<dyn CodeType>` for that type.
    //
    // There is a companion match in `templates/Types.swift` which performs a similar function for the
    // template code.
    //
    //   - When adding additional types here, make sure to also add a match arm to the `Types.swift` template.
    //   - To keep things managable, let's try to limit ourselves to these 2 mega-matches
    fn create_code_type(&self, type_: Type) -> Box<dyn CodeType> {
        match type_ {
            Type::UInt8 => Box::new(primitives::UInt8CodeType),
            Type::Int8 => Box::new(primitives::Int8CodeType),
            Type::UInt16 => Box::new(primitives::UInt16CodeType),
            Type::Int16 => Box::new(primitives::Int16CodeType),
            Type::UInt32 => Box::new(primitives::UInt32CodeType),
            Type::Int32 => Box::new(primitives::Int32CodeType),
            Type::UInt64 => Box::new(primitives::UInt64CodeType),
            Type::Int64 => Box::new(primitives::Int64CodeType),
            Type::Float32 => Box::new(primitives::Float32CodeType),
            Type::Float64 => Box::new(primitives::Float64CodeType),
            Type::Boolean => Box::new(primitives::BooleanCodeType),
            Type::String => Box::new(primitives::StringCodeType),

            Type::Duration => Box::new(miscellany::DurationCodeType),
            Type::Map {
                key_type,
                value_type,
            } => Box::new(compounds::MapCodeType::new(*key_type, *value_type)),
            Type::Object {
                module_path,
                name,
                imp: _,
            } => Box::new(object::ObjectCodeType::new(name, module_path)),
            Type::Optional { inner_type } => {
                Box::new(compounds::OptionalCodeType::new(*inner_type))
            }
            Type::Record { module_path, name } => {
                Box::new(record::RecordCodeType::new(name, module_path))
            }
            Type::Sequence { inner_type } => {
                Box::new(compounds::SequenceCodeType::new(*inner_type))
            }
            Type::Timestamp => Box::new(miscellany::TimestampCodeType),
            Type::Custom { name, .. } => Box::new(custom::CustomCodeType::new(name)),

            Type::Enum { module_path, name } => {
                Box::new(enum_::EnumCodeType::new(name, module_path))
            }
            Type::CallbackInterface { module_path, name } => Box::new(
                callback_interface::CallbackInterfaceCodeType::new(name, module_path),
            ),

            _ => panic!("{:?} unsupported", type_),
        }
    }

    fn find(&self, type_: &impl AsType) -> Box<dyn CodeType> {
        self.create_code_type(type_.as_type())
    }

    /// Get the idiomatic Swift rendering of a class name (for enums, records, errors, etc).
    fn class_name(&self, nm: &str) -> String {
        nm.to_string().to_upper_camel_case()
    }

    /// Get the idiomatic Swift rendering of a function name.
    fn fn_name(&self, nm: &str) -> String {
        nm.to_string().to_upper_camel_case()
    }

    /// Get the idiomatic Swift rendering of a variable name.
    fn var_name(&self, nm: &str) -> String {
        // source: https://go.dev/ref/spec#Keywords
        if [
            "break",
            "case",
            "chan",
            "const",
            "continue",
            "default",
            "defer",
            "else",
            "fallthrough",
            "for",
            "func",
            "go",
            "goto",
            "if",
            "import",
            "interface",
            "map",
            "package",
            "range",
            "return",
            "select",
            "struct",
            "switch",
            "type",
            "var",
        ]
        .contains(&nm)
        {
            // This is done in order to avoid variables named as keywords and causing
            // compilation issues
            let mut s = String::from("var_");
            s.push_str(nm);
            s
        } else {
            nm.to_string()
        }
        .to_lower_camel_case()
    }

    /// Get the idiomatic Swift rendering of an individual enum variant.
    fn enum_variant_name(&self, nm: &str) -> String {
        nm.to_string().to_upper_camel_case()
    }

    /// Get the idiomatic Swift rendering of an exception name.
    fn error_name(&self, nm: &str) -> String {
        self.class_name(nm)
    }

    fn ffi_type_label(&self, ffi_type: &FfiType) -> String {
        match ffi_type {
            FfiType::Int8 => "int8_t".into(),
            FfiType::UInt8 => "uint8_t".into(),
            FfiType::Int16 => "int16_t".into(),
            FfiType::UInt16 => "uint16_t".into(),
            FfiType::Int32 => "int32_t".into(),
            FfiType::UInt32 => "uint32_t".into(),
            FfiType::Int64 => "int64_t".into(),
            FfiType::UInt64 => "uint64_t".into(),
            FfiType::Float32 => "float".into(),
            FfiType::Float64 => "double".into(),
            FfiType::RustArcPtr(_) => "void*".into(),
            FfiType::RustBuffer(_) => "RustBuffer".into(),
            FfiType::ForeignBytes => "ForeignBytes".into(),
            FfiType::ForeignCallback => "ForeignCallback".to_string(),
            FfiType::ForeignBytes => todo!(),
            FfiType::ForeignCallback => todo!(),
            FfiType::ForeignExecutorHandle => todo!(),
            FfiType::ForeignExecutorCallback => todo!(),
            FfiType::FutureCallback { .. } => todo!(),
            FfiType::FutureCallbackData => todo!(),
        }
    }
}

pub mod filters {
    use super::*;

    fn oracle() -> &'static GoCodeOracle {
        &GoCodeOracle
    }

    pub fn ffi_converter_name(type_: &impl AsType) -> Result<String, askama::Error> {
        Ok(oracle().find(type_).ffi_converter_name())
    }

    pub fn ffi_destroyer_name(type_: &impl AsType) -> Result<String, askama::Error> {
        Ok(oracle().class_name(&format!(
            "ffiDestroyer{}",
            oracle().find(type_).canonical_name()
        )))
    }

    pub fn read_fn(type_: &impl AsType) -> Result<String, askama::Error> {
        Ok(format!(
            "{}INSTANCE.read",
            oracle().find(type_).ffi_converter_name()
        ))
    }

    pub fn lift_fn(type_: &impl AsType) -> Result<String, askama::Error> {
        Ok(format!(
            "{}INSTANCE.lift",
            oracle().find(type_).ffi_converter_name()
        ))
    }

    pub fn write_fn(type_: &impl AsType) -> Result<String, askama::Error> {
        Ok(format!(
            "{}INSTANCE.write",
            oracle().find(type_).ffi_converter_name()
        ))
    }

    pub fn lower_fn(type_: &impl AsType) -> Result<String, askama::Error> {
        Ok(format!(
            "{}INSTANCE.lower",
            oracle().find(type_).ffi_converter_name()
        ))
    }

    pub fn destroy_fn(type_: &impl AsType) -> Result<String, askama::Error> {
        Ok(format!("{}{{}}.destroy", ffi_destroyer_name(type_)?))
    }

    pub fn var_name(nm: &str) -> Result<String, askama::Error> {
        Ok(oracle().var_name(nm))
    }

    /// Get the idiomatic Go rendering of a struct field name.
    pub fn field_name(nm: &str) -> Result<String, askama::Error> {
        Ok(nm.to_string().to_upper_camel_case())
    }

    pub fn type_name(type_: &impl AsType) -> Result<String, askama::Error> {
        Ok(oracle().find(type_).type_label())
    }

    pub fn canonical_name(type_: &impl AsType) -> Result<String, askama::Error> {
        Ok(oracle().find(type_).canonical_name())
    }

    pub fn class_name(nm: &str) -> Result<String, askama::Error> {
        Ok(oracle().class_name(nm))
    }

    /// FFI type name to be used to define cgo types inside bridging header
    pub fn cgo_ffi_type_name(type_: &FfiType) -> Result<String, askama::Error> {
        Ok(oracle().ffi_type_label(type_))
    }

    pub fn into_ffi_type(type_: &Type) -> Result<FfiType, askama::Error> {
        Ok(type_.into())
    }

    /// FFI type name to be used to reference cgo types
    pub fn ffi_type_name(type_: &FfiType) -> Result<String, askama::Error> {
        let result = match type_ {
            FfiType::RustArcPtr(_) => "unsafe.Pointer".into(),
            _ => format!("C.{}", oracle().ffi_type_label(type_)),
        };
        Ok(result)
    }

    /// Get the idiomatic Kotlin rendering of a function name.
    pub fn fn_name(nm: &str) -> Result<String, askama::Error> {
        Ok(oracle().fn_name(nm))
    }

    /// Get the idiomatic Kotlin rendering of a function name.
    pub fn enum_variant_name(nm: &str) -> Result<String, askama::Error> {
        Ok(oracle().enum_variant_name(nm))
    }
}

/// Renders Go helper code for all types
///
/// This template is a bit different than others in that it stores internal state from the render
/// process.  Make sure to only call `render()` once.
#[derive(Template)]
#[template(syntax = "go", escape = "none", path = "Types.go")]
pub struct TypeRenderer<'a> {
    config: &'a Config,

    ci: &'a ComponentInterface,

    // Track included modules for the `include_once()` macro
    include_once_names: RefCell<HashSet<String>>,

    // Track imports added with the `add_import()` macro
    imports: RefCell<BTreeSet<String>>,
}

impl<'a> TypeRenderer<'a> {
    fn new(config: &'a Config, ci: &'a ComponentInterface) -> Self {
        Self {
            config,
            ci,
            include_once_names: RefCell::new(HashSet::new()),
            imports: RefCell::new(BTreeSet::new()),
        }
    }

    // Helper for the including a template, but only once.
    //
    // The first time this is called with a name it will return true, indicating that we should
    // include the template. Subsequent calls will return false.
    fn include_once_check(&self, name: &str) -> bool {
        self.include_once_names
            .borrow_mut()
            .insert(name.to_string())
    }

    // Helper to add an import statement
    //
    // Call this inside your template to cause an import statement to be added at the top of the
    // file. Imports will be sorted and de-deuped.
    //
    // Returns an empty string so that it can be used inside an askama `{{ }}` block.
    fn add_import(&self, name: &str) -> &str {
        self.imports.borrow_mut().insert(name.to_owned());
        ""
    }

    pub fn cgo_callback_fn(&self, name: &str) -> String {
        format!("{}_cgo_{}", self.ci.ffi_namespace(), name)
    }
}
