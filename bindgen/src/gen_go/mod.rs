/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

use anyhow::{Context, Result};
use askama::Template;
use heck::{ToLowerCamelCase, ToSnakeCase, ToUpperCamelCase};
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
mod executor;
mod external;
mod miscellany;
mod object;
mod primitives;
mod record;

#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Config {
    cdylib_name: Option<String>,
    package_name: Option<String>,
    c_module_filename: Option<String>,
    #[serde(default)]
    custom_types: HashMap<String, CustomTypeConfig>,
    #[serde(default)]
    go_mod: Option<String>,
}

impl uniffi_bindgen::BindingsConfig for Config {
    fn update_from_ci(&mut self, ci: &ComponentInterface) {
        self.package_name
            .get_or_insert_with(|| ci.namespace().into());
        self.cdylib_name
            .get_or_insert_with(|| format!("uniffi_{}", ci.namespace()));
    }
    fn update_from_cdylib_name(&mut self, cdylib_name: &str) {
        self.cdylib_name
            .get_or_insert_with(|| cdylib_name.to_string());
    }

    fn update_from_dependency_configs(
        &mut self,
        _config_map: std::collections::HashMap<&str, &Self>,
    ) {
        // unused
    }
}

#[derive(Debug, Default, Clone, Serialize, Deserialize)]
pub struct CustomTypeConfig {
    imports: Option<Vec<String>>,
    type_name: Option<String>,
    into_custom: TemplateExpression,
    from_custom: TemplateExpression,
}

/// A struct to record a go import statement.
#[derive(Clone, Debug, Eq, Ord, PartialEq, PartialOrd)]
pub enum ImportRequirement {
    /// A simple module import.
    Module { mod_name: String },
}

impl ImportRequirement {
    fn render(&self) -> String {
        match &self {
            ImportRequirement::Module { mod_name } => format!("\"{mod_name}\""),
        }
    }
}

impl Config {
    /// The name of the go package containing the high-level foreign-language bindings.
    pub fn package_name(&self) -> String {
        self.package_name.clone().expect("missing package name")
    }

    /// The filename stem for the lower-level C module containing the FFI declarations.
    pub fn c_module_filename(&self) -> String {
        match self.c_module_filename.as_ref() {
            Some(name) => name.clone(),
            None => self.package_name(),
        }
    }

    /// The name of the `.h` file for the lower-level C module with FFI declarations.
    pub fn header_filename(&self) -> String {
        format!("{}.h", self.c_module_filename())
    }

    /// The name of the `.c` file for the lower-level C module with FFI declarations.
    pub fn c_filename(&self) -> String {
        format!("{}.c", self.c_module_filename())
    }
}

#[derive(Template)]
#[template(syntax = "go", escape = "none", path = "wrapper.go")]
pub struct GoWrapper<'a> {
    config: Config,
    ci: &'a ComponentInterface,
    type_helper_code: String,
    type_imports: BTreeSet<ImportRequirement>,
    has_async_fns: bool,
}

impl<'a> GoWrapper<'a> {
    pub fn new(config: Config, ci: &'a ComponentInterface) -> Self {
        let type_renderer = TypeRenderer::new(&config, ci);
        let type_helper_code = type_renderer.render().expect("type rendering");
        let type_imports = type_renderer.imports.into_inner();
        Self {
            config,
            ci,
            type_helper_code,
            type_imports,
            has_async_fns: ci.has_async_fns(),
        }
    }

    pub fn imports(&self) -> Vec<ImportRequirement> {
        self.type_imports.iter().cloned().collect()
    }

    pub fn initialization_fns(&self) -> Vec<String> {
        self.ci
            .iter_types()
            .map(|t| GoCodeOracle.find(t))
            .filter_map(|t| t.initialization_fn())
            .chain(
                self.has_async_fns
                    .then(|| "uniffiInitContinuationCallback".into()),
            )
            .collect()
    }
}

pub fn generate_go_bindings(
    config: &Config,
    ci: &ComponentInterface,
) -> Result<(String, String, String)> {
    let header = BridgingHeader::new(config, ci)
        .render()
        .context("failed to render Go bridging header")?;
    let c_content = BridgingCFile::new(config, ci)
        .render()
        .context("failed to render Go bridging file")?;
    let wrapper = GoWrapper::new(config.clone(), ci)
        .render()
        .context("failed to render go bindings")?;
    Ok((header, c_content, wrapper))
}

/// Template for generating the `.h` file that defines the low-level C FFI.
///
/// This file defines only the low-level structs and functions that are exposed
/// by the compiled Rust code. It gets wrapped into a higher-level API by the
/// code from [`GoWrapper`].
#[derive(Template)]
#[template(syntax = "c", escape = "none", path = "BridgingHeaderTemplate.h")]
pub struct BridgingHeader<'config, 'ci> {
    config: &'config Config,
    ci: &'ci ComponentInterface,
}

impl<'config, 'ci> BridgingHeader<'config, 'ci> {
    pub fn new(config: &'config Config, ci: &'ci ComponentInterface) -> Self {
        Self { config, ci }
    }

    // This represents true callback functions used in CGo layer. Thi is needed due to
    // https://github.com/golang/go/issues/19837
    pub fn cgo_callback_fns(&self) -> Vec<String> {
        self.ci
            .callback_interface_definitions()
            .iter()
            .map(|d| format!("{}_cgo_{}", module_path(d), d.name()))
            .collect()
    }
}

/// Template for generating the `.c` file that defines the low-level C FFI.
///
/// This file defines only the low-level structs and functions that are exposed
/// by the compiled Rust code. It gets wrapped into a higher-level API by the
/// code from [`GoWrapper`].
#[derive(Template)]
#[template(syntax = "c", escape = "none", path = "BridgingCTemplate.c")]
pub struct BridgingCFile<'config, 'ci> {
    config: &'config Config,
    _ci: &'ci ComponentInterface,
}

impl<'config, 'ci> BridgingCFile<'config, 'ci> {
    pub fn new(config: &'config Config, ci: &'ci ComponentInterface) -> Self {
        Self { config, _ci: ci }
    }
}

fn module_path(cbi: &CallbackInterface) -> String {
    if let Type::CallbackInterface { module_path, .. } = cbi.as_type() {
        module_path
    } else {
        unreachable!()
    }
}

#[derive(Clone)]
pub struct GoCodeOracle;

impl GoCodeOracle {
    // Map `Type` instances to a `Box<dyn CodeType>` for that type.
    //
    // There is a companion match in `templates/Types.go` which performs a similar function for the
    // template code.
    //
    //   - When adding additional types here, make sure to also add a match arm to the `Types.go` template.
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
            Type::Bytes => Box::new(primitives::BytesCodeType),
            Type::Duration => Box::new(miscellany::DurationCodeType),
            Type::Map {
                key_type,
                value_type,
            } => Box::new(compounds::MapCodeType::new(*key_type, *value_type)),
            Type::Object { name, .. } => Box::new(object::ObjectCodeType::new(name)),
            Type::Optional { inner_type } => {
                Box::new(compounds::OptionalCodeType::new(*inner_type))
            }
            Type::Record { name, .. } => Box::new(record::RecordCodeType::new(name)),
            Type::Sequence { inner_type } => {
                Box::new(compounds::SequenceCodeType::new(*inner_type))
            }
            Type::Timestamp => Box::new(miscellany::TimestampCodeType),
            Type::Custom { name, .. } => Box::new(custom::CustomCodeType::new(name)),

            Type::Enum { name, .. } => Box::new(enum_::EnumCodeType::new(name)),
            Type::CallbackInterface { name, .. } => {
                Box::new(callback_interface::CallbackInterfaceCodeType::new(name))
            }
            Type::ForeignExecutor => Box::new(executor::ForeignExecutorCodeType),
            Type::External {
                name,
                module_path,
                kind,
                namespace,
                tagged,
            } => Box::new(external::ExternalCodeType::new(
                name,
                module_path,
                kind,
                namespace,
                tagged,
            )),
        }
    }

    fn find(&self, type_: &impl AsType) -> Box<dyn CodeType> {
        self.create_code_type(type_.as_type())
    }

    /// Get the idiomatic Go rendering of a class name (for enums, records, errors, etc).
    fn class_name(&self, nm: &str) -> String {
        nm.to_string().to_upper_camel_case()
    }

    /// Get the idiomatic Go rendering of a function name.
    fn fn_name(&self, nm: &str) -> String {
        nm.to_string().to_upper_camel_case()
    }

    /// Get the idiomatic Go rendering of a variable name.
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

    /// Get the idiomatic Go rendering of an individual enum variant.
    fn enum_variant_name(&self, nm: &str) -> String {
        nm.to_string().to_upper_camel_case()
    }

    /// Get the import path for a external type
    fn import_name(&self, nm: &str) -> String {
        nm.to_snake_case()
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
            FfiType::ForeignExecutorHandle => "int".into(),
            FfiType::ForeignExecutorCallback => "ForeignExecutorCallback".into(),
            FfiType::RustFutureHandle | FfiType::RustFutureContinuationData => "void*".into(),
            FfiType::RustFutureContinuationCallback => "RustFutureContinuation".into(),
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
        let class_name = oracle().class_name(&format!(
            "FfiDestroyer{}",
            oracle().find(type_).canonical_name()
        ));
        match type_.as_type() {
            Type::External { namespace, .. } => Ok(format!("{}.{}", namespace, class_name)),
            _ => Ok(class_name),
        }
    }

    pub fn read_fn(type_: &impl AsType) -> Result<String, askama::Error> {
        Ok(format!(
            "{}INSTANCE.Read",
            oracle().find(type_).ffi_converter_name()
        ))
    }

    pub fn lift_fn(type_: &impl AsType) -> Result<String, askama::Error> {
        Ok(format!(
            "{}INSTANCE.Lift",
            oracle().find(type_).ffi_converter_name()
        ))
    }

    pub fn write_fn(type_: &impl AsType) -> Result<String, askama::Error> {
        Ok(format!(
            "{}INSTANCE.Write",
            oracle().find(type_).ffi_converter_name()
        ))
    }

    pub fn lower_fn(type_: &impl AsType) -> Result<String, askama::Error> {
        Ok(format!(
            "{}INSTANCE.Lower",
            oracle().find(type_).ffi_converter_name()
        ))
    }

    pub fn destroy_fn(type_: &impl AsType) -> Result<String, askama::Error> {
        Ok(format!("{}{{}}.Destroy", ffi_destroyer_name(type_)?))
    }

    pub fn var_name(nm: &str) -> Result<String, askama::Error> {
        Ok(oracle().var_name(nm))
    }

    pub fn import_name(nm: &str) -> Result<String, askama::Error> {
        Ok(oracle().import_name(nm))
    }

    /// Get the idiomatic Go rendering of a struct field name.
    pub fn field_name(nm: &str) -> Result<String, askama::Error> {
        Ok(nm.to_string().to_upper_camel_case())
    }

    pub fn error_field_name(nm: &str) -> Result<String, askama::Error> {
        // Fields called 'Error' can clash with structs which implement the error
        // interface, causing a compilation error. Suffix with _ similar to reserved
        // keywords in var names.
        if nm == "error" {
            return Ok(String::from("Error_"));
        }
        Ok(nm.to_string().to_upper_camel_case())
    }

    // Return the runtime type cast of this field if it is an Enum type. In most cases
    // we want to pass around the `error` interface and let the caller type cast, but in
    // some cases (e.g when writing nested errors) we need to work with concrete error types
    // which involve type casting from `error` to `ConcreteError`.
    pub fn error_type_cast(type_: &impl AsType) -> Result<String, askama::Error> {
        let result = match type_.as_type() {
            Type::Enum { .. } => format!(".(*{})", oracle().find(type_).type_label()),
            _ => String::from(""),
        };
        Ok(result)
    }

    pub fn type_name(type_: &impl AsType) -> Result<String, askama::Error> {
        Ok(oracle().find(type_).type_label())
    }

    pub fn variant_type_name(type_: &impl AsType) -> Result<String, askama::Error> {
        let result = match type_.as_type() {
            Type::Enum { .. } => format!("*{}", oracle().find(type_).type_label()),
            _ => oracle().find(type_).type_label(),
        };
        Ok(result)
    }

    pub fn canonical_name(type_: &impl AsType) -> Result<String, askama::Error> {
        Ok(oracle().find(type_).canonical_name())
    }

    pub fn class_name(nm: &str) -> Result<String, askama::Error> {
        Ok(oracle().class_name(nm))
    }

    pub fn into_ffi_type(type_: &Type) -> Result<FfiType, askama::Error> {
        Ok(type_.into())
    }

    pub fn cgo_ffi_type(type_: &FfiType) -> Result<String, askama::Error> {
        Ok(oracle().ffi_type_label(&type_))
    }

    /// FFI type name to be used to reference cgo types
    pub fn ffi_type_name(type_: &Type) -> Result<String, askama::Error> {
        let ffi_type: FfiType = type_.clone().into();
        let result = match ffi_type {
            FfiType::RustArcPtr(_) => "unsafe.Pointer".into(),
            FfiType::RustBuffer(_) => match type_ {
                Type::External { namespace, .. } => format!("{}.RustBufferI", namespace),
                _ => "RustBufferI".into(),
            },
            _ => format!("C.{}", oracle().ffi_type_label(&ffi_type)),
        };
        Ok(result)
    }

    /// FFI type name to be used to reference cgo types. Such that they exactly match to the cgo bindings and can be used with `//export`.
    pub fn ffi_type_name_cgo_safe<T: Clone + Into<FfiType>>(
        type_: &T,
    ) -> Result<String, askama::Error> {
        let ffi_type: FfiType = type_.clone().into();
        let result = match ffi_type {
            FfiType::RustArcPtr(_) => "unsafe.Pointer".into(),
            FfiType::RustBuffer(_) => "RustBuffer".into(),
            _ => format!("C.{}", oracle().ffi_type_label(&ffi_type)),
        };
        Ok(result)
    }

    /// Get the idiomatic Go rendering of a function name.
    pub fn fn_name(nm: &str) -> Result<String, askama::Error> {
        Ok(oracle().fn_name(nm))
    }

    /// Get the idiomatic Go rendering of a function name.
    pub fn enum_variant_name(nm: &str) -> Result<String, askama::Error> {
        Ok(oracle().enum_variant_name(nm))
    }

    /// Get the idiomatic Go rendering of docstring
    pub fn docstring(docstring: &str, tabs: &i32) -> Result<String, askama::Error> {
        let docstring = textwrap::indent(&textwrap::dedent(docstring), "// ");

        let tabs = usize::try_from(*tabs).unwrap_or_default();
        Ok(textwrap::indent(&docstring, &"\t".repeat(tabs)))
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
    imports: RefCell<BTreeSet<ImportRequirement>>,
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
        self.imports.borrow_mut().insert(ImportRequirement::Module {
            mod_name: name.to_owned(),
        });
        ""
    }

    fn add_local_import(&self, mod_name: &str) -> &str {
        let mod_name = if let Some(ref go_mod) = self.config.go_mod {
            let go_mod = go_mod.trim_end_matches("/");
            format!("{go_mod}/{mod_name}")
        } else {
            format!("{mod_name}")
        };

        self.imports
            .borrow_mut()
            .insert(ImportRequirement::Module { mod_name });
        ""
    }

    pub fn cgo_callback_fn(&self, name: &str, module_path: &str) -> String {
        format!("{module_path}_cgo_{name}")
    }
}
