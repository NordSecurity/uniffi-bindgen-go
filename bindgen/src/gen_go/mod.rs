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
use uniffi_bindgen::backend::TemplateExpression;
use uniffi_bindgen::interface::*;

use self::filters::oracle;

pub mod filters;

mod callback_interface;
mod compounds;
mod custom;
mod enum_;
mod external;
mod miscellany;
mod object;
mod primitives;
mod record;

pub trait CodeType: std::fmt::Debug {
    /// The language specific label used to reference this type. This will be used in
    /// method signatures and property declarations.
    fn type_label(&self, ci: &ComponentInterface) -> String;

    /// A representation of this type label that can be used as part of another
    /// identifier. e.g. `read_foo()`, or `FooInternals`.
    ///
    /// This is especially useful when creating specialized objects or methods to deal
    /// with this type only.
    fn canonical_name(&self) -> String;

    fn literal(&self, _literal: &Literal, ci: &ComponentInterface) -> String {
        unimplemented!("Unimplemented for {}", self.type_label(ci))
    }

    /// Name of the FfiConverter
    ///
    /// This is the object that contains the lower, write, lift, and read methods for this type.
    /// Depending on the binding this will either be a singleton or a class with static methods.
    ///
    /// This is the newer way of handling these methods and replaces the lower, write, lift, and
    /// read CodeType methods.  Currently only used by Kotlin, but the plan is to move other
    /// backends to using this.
    fn ffi_converter_name(&self) -> String {
        format!("FfiConverter{}", self.canonical_name())
    }

    fn ffi_converter_instance(&self) -> String {
        format!("{}INSTANCE", self.ffi_converter_name())
    }

    fn ffi_destroyer_name(&self) -> String {
        format!("FfiDestroyer{}", self.canonical_name())
    }

    /// An expression for lowering a value into something we can pass over the FFI.
    fn lower(&self) -> String {
        format!("{}.Lower", self.ffi_converter_instance())
    }

    /// An expression for lowering a value into something we can pass over the FFI,
    /// when the type is external
    fn lower_external(&self) -> String {
        format!("{}.LowerExternal", self.ffi_converter_instance())
    }

    /// An expression for writing a value into a byte buffer.
    fn write(&self) -> String {
        format!("{}.Write", self.ffi_converter_instance())
    }

    /// An expression for lifting a value from something we received over the FFI.
    fn lift(&self) -> String {
        format!("{}.Lift", self.ffi_converter_instance())
    }

    /// An expression for reading a value from a byte buffer.
    fn read(&self) -> String {
        format!("{}.Read", self.ffi_converter_instance())
    }

    // An expression to destroy this type
    fn destroy(&self) -> String {
        format!("{}{{}}.Destroy", self.ffi_destroyer_name())
    }

    /// A list of imports that are needed if this type is in use.
    /// Classes are imported exactly once.
    fn imports(&self) -> Option<Vec<String>> {
        None
    }

    /// Function to run at startup
    fn initialization_fn(&self) -> Option<String> {
        None
    }
}

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

impl Config {
    pub fn update_from_ci(&mut self, ci: &ComponentInterface) {
        self.package_name
            .get_or_insert_with(|| ci.namespace().into());
        self.cdylib_name
            .get_or_insert_with(|| format!("uniffi_{}", ci.namespace()));
    }

    pub fn update_from_cdylib_name(&mut self, cdylib_name: &str) {
        self.cdylib_name
            .get_or_insert_with(|| cdylib_name.to_string());
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

pub fn generate_go_bindings(config: &Config, ci: &ComponentInterface) -> Result<(String, String)> {
    let header = BridgingHeader::new(config, ci)
        .render()
        .context("failed to render Go bridging header")?;
    let wrapper = GoWrapper::new(config.clone(), ci)
        .render()
        .context("failed to render go bindings")?;
    Ok((header, wrapper))
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

    // This represents true callback functions used in CGo layer. This is needed due to
    // https://github.com/golang/go/issues/19837
    /// Returns (name, return_type, args, has_call_status)
    /// Idealy we would use FfiCalback
    pub fn cgo_callback_fns(&self) -> Vec<(String, Option<FfiType>, Vec<FfiArgument>, bool)> {
        let free_callback =
            |name: &str, module: &str| -> (String, Option<FfiType>, Vec<FfiArgument>, bool) {
                (
                    oracle().cgo_vtable_free_fn_name(name, module),
                    None,
                    vec![FfiArgument::new("handle", FfiType::Handle)],
                    false,
                )
            };

        let obj_callbacks = self
            .ci
            .object_definitions()
            .iter()
            .filter(|obj| obj.has_callback_interface())
            .map(|def| (module_path(def), def.name(), def.vtable_methods()));

        let cbi_callbacks = self
            .ci
            .callback_interface_definitions()
            .iter()
            .map(|def| (module_path(def), def.name(), def.vtable_methods()));

        obj_callbacks
            .chain(cbi_callbacks)
            .flat_map(|(module, name, vtable)| {
                let free = free_callback(name, &module);
                vtable
                    .into_iter()
                    .map(move |(ffi_cb, _)| {
                        (
                            oracle().cgo_callback_fn_name(&ffi_cb, &module),
                            ffi_cb.return_type().cloned(),
                            ffi_cb.arguments().into_iter().cloned().collect(),
                            ffi_cb.has_rust_call_status_arg(),
                        )
                    })
                    .chain([free])
            })
            .collect()
    }
}

fn module_path(type_: &impl AsType) -> String {
    match type_.as_type() {
        Type::CallbackInterface { module_path, .. } => module_path,
        Type::Object { module_path, .. } => module_path,
        _ => unreachable!(),
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
            Type::Object {
                name,
                module_path: _,
                imp,
            } => Box::new(object::ObjectCodeType::new(name, imp)),
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
        nm.to_upper_camel_case()
    }

    fn interface_name(&self, nm: &str) -> String {
        nm.to_string() + "Interface"
    }

    fn impl_name(&self, nm: &str) -> String {
        nm.to_string() + "Impl"
    }

    fn object_names(&self, obj: &Object) -> (String, String) {
        let class_name = self.class_name(obj.name());
        if obj.has_callback_interface() {
            let imp = self.impl_name(&class_name);
            (class_name, imp)
        } else {
            (self.interface_name(&class_name), class_name)
        }
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

    /// Get cgo symbol name for a callback function
    fn cgo_callback_fn_name(&self, f: &FfiCallbackFunction, module_path: &str) -> String {
        format!("{module_path}_cgo_dispatch{}", f.name())
    }

    /// Get cgo symbol name for a vtable free function
    fn cgo_vtable_free_fn_name(&self, nm: &str, module_path: &str) -> String {
        format!("{module_path}_cgo_dispatchCallbackInterface{nm}Free")
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
            FfiType::Handle => "uint64_t".into(),
            FfiType::VoidPointer => "void*".into(),
            FfiType::RustBuffer(_) => "RustBuffer".into(),
            FfiType::ForeignBytes => "ForeignBytes".into(),
            FfiType::RustCallStatus => "RustCallStatus".into(),

            FfiType::Callback(nm) => self.ffi_callback_name(nm),
            FfiType::Struct(nm) => self.ffi_struct_name(nm),
            FfiType::Reference(_ffi_type) => {
                panic!("Cannot be constructed at this level, ffi_type_name_cgo_safe should be used")
            }
        }
    }

    /// Get the idiomatic C rendering of an FFI callback function name
    fn ffi_callback_name(&self, nm: &str) -> String {
        format!("Uniffi{}", nm.to_upper_camel_case())
    }

    /// Get the idiomatic C rendering of an FFI struct name
    fn ffi_struct_name(&self, nm: &str) -> String {
        format!("Uniffi{}", nm.to_upper_camel_case())
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

    pub fn field_type_name(&self, field: &Field, ci: &ComponentInterface) -> String {
        let name = oracle().find(&field.as_type()).type_label(ci);
        match self.ci.is_name_used_as_error(&name) {
            true => format!("*{name}"),
            false => name.to_string(),
        }
    }
}
