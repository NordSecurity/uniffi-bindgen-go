/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

use anyhow::{Context, Result};
use askama::Template;
use camino::Utf8Path;
use fs_err::{self as fs};
use heck::{ToLowerCamelCase, ToSnakeCase, ToUpperCamelCase};
use serde::{Deserialize, Serialize};
use std::borrow::Borrow;
use std::cell::RefCell;
use std::collections::{BTreeSet, HashMap, HashSet};
use uniffi_bindgen::backend::{CodeType, TemplateExpression};
use uniffi_bindgen::{interface::*, BindingsConfig};

mod callback_interface;
mod compounds;
mod custom;
mod enum_;
mod error;
mod executor;
mod external;
mod miscellany;
mod object;
mod primitives;
mod record;

#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Config {
    cdylib_name: Option<String>,
    module_name: Option<String>,
    ffi_module_name: Option<String>,
    ffi_module_filename: Option<String>,
    package_name: Option<String>,
    #[serde(default)]
    custom_types: HashMap<String, CustomTypeConfig>,
    #[serde(default)]
    go_mod: Option<String>,
}

impl uniffi_bindgen::BindingsConfig for Config {
    const TOML_KEY: &'static str = "go";

    fn update_from_ci(&mut self, ci: &ComponentInterface) {
        self.module_name
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
    /// Import everything from a module.
    DotModule { mod_name: String },
}

impl ImportRequirement {
    fn render(&self) -> String {
        match &self {
            ImportRequirement::Module { mod_name } => format!("\"{mod_name}\""),
            ImportRequirement::DotModule { mod_name } => format!(". \"{mod_name}\""),
        }
    }
}

/// Load the config from the TOML value, falling back to an empty map if it doesn't exist.
fn load_toml_file(source: &Utf8Path) -> Result<toml::value::Table> {
    let contents = fs::read_to_string(source)
        .with_context(|| format!("Failed to read config file: {:?}", source))?;
    let mut cfg: toml::value::Table = toml::de::from_str(&contents)?;

    let res = cfg
        .remove("bindings")
        .and_then(|mut cfg| cfg.as_table_mut().map(|tbl| tbl.remove(Config::TOML_KEY)))
        .flatten()
        .and_then(|cfg| cfg.try_into().ok())
        .unwrap_or_else(|| toml::value::Table::default());
    Ok(res)
}

impl Config {
    pub fn load_initial(
        crate_root: &Utf8Path,
        config_file_override: Option<&Utf8Path>,
    ) -> Result<Self> {
        let default_config = crate_root.join("uniffi.toml").canonicalize_utf8().ok();
        let toml_config = match (default_config, config_file_override) {
            (Some(cfg), Some(over)) => {
                let mut cfg: toml::value::Table = load_toml_file(&cfg)?;
                let over: toml::value::Table = load_toml_file(over)?;

                fn merge(a: &mut toml::value::Table, b: toml::value::Table) {
                    for (key, value) in b.into_iter() {
                        match a.get_mut(&key) {
                            Some(existing_value) => match (existing_value, value) {
                                (toml::Value::Table(ref mut t0), toml::Value::Table(t1)) => {
                                    merge(t0, t1);
                                }
                                (v, value) => *v = value,
                            },
                            None => {
                                a.insert(key, value);
                            }
                        }
                    }
                }

                merge(&mut cfg, over);

                toml::Value::from(cfg)
            }
            (Some(cfg), None) => toml::Value::from(load_toml_file(&cfg)?),
            (None, Some(over)) => toml::Value::from(load_toml_file(&over)?),
            (None, None) => toml::Value::from(toml::value::Table::default()),
        };

        Ok(toml_config.try_into()?)
    }

    /// The name of the go package containing the high-level foreign-language bindings.
    pub fn package_name(&self) -> String {
        match self.module_name.as_ref() {
            Some(name) => name.clone(),
            None => "uniffi".into(),
        }
    }

    /// The name of the lower-level C module containing the FFI declarations.
    pub fn ffi_package_name(&self) -> String {
        match self.ffi_module_name.as_ref() {
            Some(name) => name.clone(),
            None => format!("{}FFI", self.package_name()),
        }
    }

    /// The filename stem for the lower-level C module containing the FFI declarations.
    pub fn ffi_package_filename(&self) -> String {
        match self.ffi_module_filename.as_ref() {
            Some(name) => name.clone(),
            None => self.ffi_package_name(),
        }
    }

    /// The name of the `.h` file for the lower-level C module with FFI declarations.
    pub fn header_filename(&self) -> String {
        format!("{}.h", self.ffi_package_filename())
    }

    /// The name of the `.c` file for the lower-level C module with FFI declarations.
    pub fn c_filename(&self) -> String {
        format!("{}.c", self.ffi_package_filename())
    }

    /// The name of the compiled Rust library containing the FFI implementation.
    pub fn cdylib_name(&self) -> String {
        if let Some(cdylib_name) = &self.cdylib_name {
            cdylib_name.clone()
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
    type_imports: BTreeSet<ImportRequirement>,
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
    _config: &'config Config,
    ci: &'ci ComponentInterface,
}

impl<'config, 'ci> BridgingHeader<'config, 'ci> {
    pub fn new(config: &'config Config, ci: &'ci ComponentInterface) -> Self {
        Self {
            _config: config,
            ci,
        }
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
            } => Box::new(external::ExternalCodeType::new(name, module_path, kind)),
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

    /// Get the idiomatic Go rendering of an exception name.
    fn error_name(&self, nm: &str) -> String {
        self.class_name(nm)
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
            FfiType::FutureCallback { return_type } => {
                let return_type = filters::cgo_ffi_callback_type(return_type).unwrap();
                format!("UniFfiFutureCallback{}", return_type)
            }
            FfiType::FutureCallbackData => "void*".into(),
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
            Type::External { module_path, .. } => Ok(format!("{}.{}", module_path, class_name)),
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

    pub fn lower_fn_call(arg: &Argument) -> Result<String, askama::Error> {
        let res = match arg.as_type() {
            Type::External {
                kind: ExternalKind::DataClass,
                ..
            } => {
                format!(
                    "RustBufferFromForeign({}({}))",
                    lower_fn(arg)?,
                    var_name(arg.name())?
                )
            }
            _ => format!("{}({})", lower_fn(arg)?, var_name(arg.name())?),
        };

        Ok(res)
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

    pub fn type_name(type_: &impl AsType) -> Result<String, askama::Error> {
        Ok(oracle().find(type_).type_label())
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

    pub fn cgo_ffi_callback_type(type_: &FfiType) -> Result<String, askama::Error> {
        let res = match type_ {
            FfiType::FutureCallbackData => "FutureCallbackData".into(),
            FfiType::RustArcPtr(_) => "RustArcPtr".into(),
            _ => oracle().ffi_type_label(type_),
        };
        Ok(res)
    }

    /// FFI type name to be used to reference cgo types
    pub fn ffi_type_name<T: Clone + Into<FfiType>>(type_: &T) -> Result<String, askama::Error> {
        let ffi_type: FfiType = type_.clone().into();
        let result = match ffi_type {
            FfiType::RustArcPtr(_) => "unsafe.Pointer".into(),
            FfiType::RustBuffer(name) => match name {
                Some(_name) => {
                    // External buffer
                    format!("RustBufferI")
                }
                None => {
                    // Our "own"
                    "RustBufferI".into()
                }
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

    pub fn default_type(type_: &Option<Type>) -> Result<String, askama::Error> {
        let res = match type_ {
            Some(ty) => match ty {
                Type::UInt8
                | Type::Int8
                | Type::UInt16
                | Type::Int16
                | Type::UInt32
                | Type::Int32
                | Type::UInt64
                | Type::Int64
                | Type::Float32
                | Type::Float64 => "0".into(),
                Type::Boolean => "false".into(),
                Type::String => "\"\"".into(),
                Type::Bytes => "[]byte{{}}".into(),
                Type::Duration => "time.Duration{}".into(),
                Type::Map {
                    key_type,
                    value_type,
                } => format!(
                    "map[{}]{}{{}}",
                    type_name(&key_type)?,
                    type_name(&value_type)?
                ),
                Type::Object { .. } => "nil".into(),
                Type::Optional { .. } => "nil".into(),
                Type::Record { name, .. } => format!("{}{{}}", name),
                Type::Sequence { inner_type } => {
                    format!("[]{}{{}}", type_name(inner_type)?)
                }
                Type::Timestamp => "time.Time{}".into(),
                Type::Custom { name, .. } => format!("{}{{}}", name),

                Type::Enum { .. } => "0".into(), // enums are respresented as uint
                Type::CallbackInterface { .. } => "nil".into(),
                Type::ForeignExecutor => "nil".into(),
                Type::External { name, kind, .. } => match kind {
                    ExternalKind::Interface => "nil".into(),
                    ExternalKind::DataClass => format!("{}{{}}", name),
                },
            },
            None => "nil".into(),
        };
        Ok(res)
    }

    /// Name of the callback function to handle an async result
    pub fn future_callback(result: &ResultType) -> Result<String, askama::Error> {
        Ok(format!(
            "uniffiFutureCallbackHandler{}{}",
            match &result.return_type {
                Some(t) => oracle().find(t).canonical_name(),
                None => "Void".into(),
            },
            match &result.throws_type {
                Some(t) => oracle().find(t).canonical_name(),
                None => "".into(),
            }
        ))
    }

    pub fn future_chan_type(result: &ResultType) -> Result<String, askama::Error> {
        Ok(format!(
            "struct {{ {} err error }}",
            match &result.return_type {
                Some(return_type) => format!("val {};", type_name(return_type)?),
                None => "".into(),
            }
        ))
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
            format!("{go_mod}/{mod_name}/{mod_name}")
        } else {
            format!("{mod_name}/{mod_name}")
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
