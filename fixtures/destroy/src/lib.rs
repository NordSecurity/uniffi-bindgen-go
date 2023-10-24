/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

use once_cell::sync::Lazy;
use std::collections::HashMap;
use std::sync::{Arc, RwLock};

static LIVE_COUNT: Lazy<RwLock<i32>> = Lazy::new(|| RwLock::new(0));

#[derive(Debug, Clone)]
pub struct Resource {}

impl Resource {
    pub fn new() -> Self {
        *LIVE_COUNT.write().unwrap() += 1;
        Resource {}
    }
}

impl Drop for Resource {
    fn drop(&mut self) {
        *LIVE_COUNT.write().unwrap() -= 1;
    }
}

#[derive(Debug, Clone)]
pub struct ResourceJournal {
    map: Option<HashMap<i32, Arc<Resource>>>,
    list: Option<Vec<Arc<Resource>>>,
    object: Option<Arc<Resource>>,
    record: Option<SmallJournal>,
    r#enum: Option<EnumJournal>,
    duration: Option<std::time::Duration>,
    timestamp: Option<std::time::SystemTime>,
    bool: Option<bool>,
    i8: Option<i8>,
    i16: Option<i16>,
    i32: Option<i32>,
    i64: Option<i64>,
    u8: Option<u8>,
    u16: Option<u16>,
    u32: Option<u32>,
    u64: Option<u64>,
    float32: Option<f32>,
    float64: Option<f64>,
    str: String,
}

#[derive(Debug, Clone)]
pub struct SmallJournal {
    resource: Arc<Resource>,
}

#[derive(Debug, Clone)]
pub enum EnumJournal {
    Journal { journal: SmallJournal },
}

fn create_journal() -> ResourceJournal {
    ResourceJournal {
        map: None,
        list: None,
        object: None,
        record: None,
        r#enum: None,
        duration: None,
        timestamp: None,
        bool: None,
        i8: None,
        i16: None,
        i32: None,
        i64: None,
        u8: None,
        u16: None,
        u32: None,
        u64: None,
        float32: None,
        float64: None,
        str: "hello".to_string(),
    }
}

fn get_live_count() -> i32 {
    *LIVE_COUNT.read().unwrap()
}

include!(concat!(env!("OUT_DIR"), "/destroy.uniffi.rs"));
