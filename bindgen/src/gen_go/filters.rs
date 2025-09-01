use heck::ToShoutySnakeCase;
use uniffi_meta::LiteralMetadata;

use super::*;

pub fn oracle() -> &'static GoCodeOracle {
    &GoCodeOracle
}

pub fn ffi_converter_name(type_: &impl AsType) -> Result<String, askama::Error> {
    Ok(oracle().find(type_).ffi_converter_name())
}

pub fn ffi_converter_instance(type_: &impl AsType) -> Result<String, askama::Error> {
    Ok(oracle().find(type_).ffi_converter_instance())
}

pub fn ffi_destroyer_name(type_: &impl AsType) -> Result<String, askama::Error> {
    Ok(oracle().find(type_).ffi_destroyer_name())
}

pub fn read_fn(type_: &impl AsType) -> Result<String, askama::Error> {
    Ok(oracle().find(type_).read())
}

pub fn lift_fn(type_: &impl AsType) -> Result<String, askama::Error> {
    Ok(oracle().find(type_).lift())
}

pub fn write_fn(type_: &impl AsType) -> Result<String, askama::Error> {
    Ok(oracle().find(type_).write())
}

pub fn lower_fn(type_: &impl AsType) -> Result<String, askama::Error> {
    Ok(oracle().find(type_).lower())
}

pub fn lower_external_fn(type_: &impl AsType) -> Result<String, askama::Error> {
    Ok(oracle().find(type_).lower_external())
}

pub fn destroy_fn(type_: &impl AsType) -> Result<String, askama::Error> {
    Ok(oracle().find(type_).destroy())
}

pub fn var_name(nm: &str) -> Result<String, askama::Error> {
    Ok(oracle().var_name(nm))
}

/// If name is empty create one based on position of a variable
pub fn or_pos_var(nm: String, pos: &usize) -> Result<String, askama::Error> {
    if nm.is_empty() {
        Ok(format!("var{pos}"))
    } else {
        Ok(nm)
    }
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

/// If name is empty create one based on position of a field
pub fn or_pos_field(nm: String, pos: &usize) -> Result<String, askama::Error> {
    if nm.is_empty() {
        Ok(format!("Field{pos}"))
    } else {
        Ok(nm)
    }
}

pub fn type_name(type_: &impl AsType, ci: &ComponentInterface) -> Result<String, askama::Error> {
    Ok(oracle().find(type_).type_label(ci))
}

pub fn canonical_name(type_: &impl AsType) -> Result<String, askama::Error> {
    Ok(oracle().find(type_).canonical_name())
}

pub fn class_name(nm: &str) -> Result<String, askama::Error> {
    Ok(oracle().class_name(nm))
}

pub fn object_names(obj: &Object) -> Result<(String, String), askama::Error> {
    Ok(oracle().object_names(obj))
}

pub fn into_ffi_type(type_: &Type) -> Result<FfiType, askama::Error> {
    Ok(type_.into())
}

/// FFI type representation in C code
pub fn cgo_ffi_type(type_: &FfiType) -> Result<String, askama::Error> {
    let result = match type_ {
        FfiType::Reference(inner) => format!("{}*", cgo_ffi_type(inner)?),
        other => oracle().ffi_type_label(other),
    };

    Ok(result)
}

/// FFI function name to be used in as C to Go callback
pub fn cgo_callback_fn_name(
    f: &FfiCallbackFunction,
    module_path: &str,
) -> Result<String, askama::Error> {
    Ok(oracle().cgo_callback_fn_name(f, module_path))
}

/// FFI type name to be used to reference cgo types
pub fn ffi_type_name<T: Clone + Into<FfiType>>(type_: &T) -> Result<String, askama::Error> {
    let ffi_type: FfiType = type_.clone().into();
    let result = match ffi_type {
        FfiType::RustArcPtr(_) => "unsafe.Pointer".into(),
        FfiType::RustBuffer(_) => "RustBufferI".into(),
        FfiType::VoidPointer => "*C.void".into(),
        FfiType::Reference(inner) => format!("*{}", ffi_type_name(&*inner)?),
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
        FfiType::RustBuffer(_) => "C.RustBuffer".into(),
        FfiType::VoidPointer => "*C.void".into(),
        FfiType::Reference(inner) => format!("*{}", ffi_type_name_cgo_safe(&*inner)?),
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

// Get the idiomatic Go rendering of an individual enum variant's discriminant
pub fn variant_discr_literal(e: &Enum, index: &usize) -> Result<String, askama::Error> {
    let literal = e.variant_discr(*index).expect("invalid index");
    match literal {
        LiteralMetadata::UInt(v, _, _) => Ok(v.to_string()),
        LiteralMetadata::Int(v, _, _) => Ok(v.to_string()),
        _ => unreachable!("expected an int or uint!"),
    }
}

/// Get the idiomatic Go rendering of docstring
pub fn docstring(docstring: &str, tabs: &i32) -> Result<String, askama::Error> {
    let docstring = textwrap::indent(&textwrap::dedent(docstring), "// ");

    let tabs = usize::try_from(*tabs).unwrap_or_default();
    Ok(textwrap::indent(&docstring, &"\t".repeat(tabs)))
}

/// Get the idiomatic C rendering of an if guard name
pub fn if_guard_name(nm: &str) -> Result<String, askama::Error> {
    Ok(format!("UNIFFI_FFIDEF_{}", nm.to_shouty_snake_case()))
}

/// Get the idiomatic C rendering of an FFI callback function name
pub fn ffi_callback_name(nm: &str) -> Result<String, askama::Error> {
    Ok(oracle().ffi_callback_name(nm))
}

/// Get the idiomatic C rendering of an FFI callback function helper name
/// This is used to call C callbacks from go
pub fn ffi_callback_helper_name(nm: &str) -> Result<String, askama::Error> {
    Ok(format!("call_{}", oracle().ffi_callback_name(nm)))
}

/// Find C callback argument in cb and return name for helper funciton to call it
pub fn find_ffi_callback_helper(cb: &FfiCallbackFunction) -> Result<String, askama::Error> {
    Ok(cb
        .arguments()
        .iter()
        .filter_map(|arg| match arg.type_() {
            FfiType::Callback(name) => {
                Some(format!("C.{}", ffi_callback_helper_name(&name).unwrap()))
            }
            _ => None,
        })
        .next()
        .expect(
            "Must be called on async trait callback, as it is granteed to have future callback",
        ))
}

/// Get the idiomatic C rendering of an FFI struct name
pub fn ffi_struct_name(nm: &str) -> Result<String, askama::Error> {
    Ok(oracle().ffi_struct_name(nm))
}

pub fn has_display(obj: &Object) -> Result<bool, askama::Error> {
    Ok(obj
        .uniffi_traits()
        .into_iter()
        .any(|t| matches!(t, UniffiTrait::Display { .. })))
}

/// Exported go function consume rust poll continuation
pub fn future_continuation_name(config: &Config) -> Result<String, askama::Error> {
    Ok(format!(
        "{}_uniffiFutureContinuationCallback",
        config
            .package_name
            .as_ref()
            .expect("package name must be set")
    ))
}

/// Exported go function to "free"/"cancel" async callback
pub fn free_gorutine_callback(config: &Config) -> Result<String, askama::Error> {
    Ok(format!(
        "{}_uniffiFreeGorutine",
        config
            .package_name
            .as_ref()
            .expect("package name must be set")
    ))
}
