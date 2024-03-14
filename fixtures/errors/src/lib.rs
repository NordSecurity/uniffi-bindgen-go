/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

use std::collections::HashMap;

#[derive(Debug, thiserror::Error)]
pub enum BoobyTrapError {
    #[error("You slipped on deliberately poured ice")]
    IceSlip,
    #[error("A hot door knob has burnt your hand")]
    HotDoorKnob,
}

#[derive(Debug, thiserror::Error)]
pub enum ValidationError {
    #[error("Invalid user_id {user_id}")]
    InvalidUser { user_id: i32 },
    #[error("Invalid message {message}")]
    InvalidMessage { message: String },
    #[error("Invalid user {user_id} and message {message}")]
    InvalidUserAndMessage { user_id: i32, message: String },
    #[error("Unknown error")]
    UnknownError,
}

#[derive(Debug, thiserror::Error)]
pub enum ErrorNamedError {
    #[error("Error {error}")]
    Error { error: String },
}

#[derive(Debug, thiserror::Error)]
pub enum ComplexError {
    #[error("Struct")]
    Struct { position_a: Vec2, position_b: Vec2 },
    #[error("List")]
    List { list: Vec<Vec2> },
    #[error("List")]
    Map { map: HashMap<i32, Vec2> },
    #[error("Option")]
    Option {
        id_a: Option<i32>,
        id_b: Option<i32>,
    },
}

#[derive(Debug, thiserror::Error)]
pub enum NestedError {
    #[error(transparent)]
    Nested { source: ValidationError },
}

#[derive(Debug)]
pub struct Vec2 {
    x: f64,
    y: f64,
}

impl Vec2 {
    pub fn new(x: f64, y: f64) -> Vec2 {
        Vec2 { x, y }
    }
}

#[uniffi::export]
fn try_nested(trip: bool) -> Result<(), NestedError> {
    if trip {
        Err(NestedError::Nested {
            source: ValidationError::UnknownError,
        })
    } else {
        Ok(())
    }
}

fn try_void(trip: bool) -> Result<(), BoobyTrapError> {
    if trip {
        Err(BoobyTrapError::IceSlip)
    } else {
        Ok(())
    }
}

fn try_string(trip: bool) -> Result<String, BoobyTrapError> {
    if trip {
        Err(BoobyTrapError::IceSlip)
    } else {
        Ok("hello world".to_string())
    }
}

fn validate_message(user_id: i32, message: String) -> Result<(), ValidationError> {
    if user_id == 100 && message == "byebye" {
        Err(ValidationError::InvalidUserAndMessage { user_id, message })
    } else if user_id == 100 {
        Err(ValidationError::InvalidUser { user_id })
    } else if message == "byebye" {
        Err(ValidationError::InvalidMessage { message })
    } else {
        Ok(())
    }
}

fn get_complex_error(error: String) -> Result<(), ComplexError> {
    match error.as_ref() {
        "struct" => Err(ComplexError::Struct {
            position_a: Vec2::new(1.0, 1.0),
            position_b: Vec2::new(2.0, 2.0),
        }),
        "list" => Err(ComplexError::List {
            list: vec![Vec2::new(1.0, 1.0), Vec2::new(2.0, 2.0)],
        }),
        "map" => Err(ComplexError::Map {
            map: HashMap::from([(0, Vec2::new(1.0, 1.0)), (1, Vec2::new(2.0, 2.0))]),
        }),
        "option" => Err(ComplexError::Option {
            id_a: Some(123),
            id_b: None,
        }),
        _ => Ok(()),
    }
}

fn error_boolean() -> Result<bool, BoobyTrapError> {
    Err(BoobyTrapError::IceSlip)
}

fn error_duration() -> Result<std::time::Duration, BoobyTrapError> {
    Err(BoobyTrapError::IceSlip)
}

fn error_i8() -> Result<i8, BoobyTrapError> {
    Err(BoobyTrapError::IceSlip)
}

fn error_u8() -> Result<u8, BoobyTrapError> {
    Err(BoobyTrapError::IceSlip)
}

fn error_i16() -> Result<i16, BoobyTrapError> {
    Err(BoobyTrapError::IceSlip)
}

fn error_u16() -> Result<u16, BoobyTrapError> {
    Err(BoobyTrapError::IceSlip)
}

fn error_i32() -> Result<i32, BoobyTrapError> {
    Err(BoobyTrapError::IceSlip)
}

fn error_u32() -> Result<u32, BoobyTrapError> {
    Err(BoobyTrapError::IceSlip)
}

fn error_i64() -> Result<i64, BoobyTrapError> {
    Err(BoobyTrapError::IceSlip)
}

fn error_u64() -> Result<u64, BoobyTrapError> {
    Err(BoobyTrapError::IceSlip)
}

fn error_f32() -> Result<f32, BoobyTrapError> {
    Err(BoobyTrapError::IceSlip)
}

fn error_f64() -> Result<f64, BoobyTrapError> {
    Err(BoobyTrapError::IceSlip)
}

fn error_optional_timestamp() -> Result<Option<std::time::SystemTime>, BoobyTrapError> {
    Err(BoobyTrapError::IceSlip)
}

fn error_string() -> Result<String, BoobyTrapError> {
    Err(BoobyTrapError::IceSlip)
}

fn error_timestamp() -> Result<std::time::SystemTime, BoobyTrapError> {
    Err(BoobyTrapError::IceSlip)
}

include!(concat!(env!("OUT_DIR"), "/errors.uniffi.rs"));
