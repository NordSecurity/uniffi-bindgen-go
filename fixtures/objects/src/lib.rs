/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

use crossbeam::channel::{Receiver, Sender};
use once_cell::sync::Lazy;
use std::sync::{Arc, Mutex};

static RECEIVER_COUNT: Lazy<Mutex<i32>> = Lazy::new(|| Mutex::new(0));

#[derive(Debug, thiserror::Error)]
pub enum ObjectError {
    #[error("InvalidOperation")]
    InvalidOperation,
}

pub struct Object0 {}

impl Object0 {
    pub fn new() -> Object0 {
        return Object0 {};
    }

    pub fn new_custom() -> Object0 {
        return Object0 {};
    }
}

// An async function returning a struct that can throw.
#[uniffi::export]
pub async fn fallible_object0_async(do_fail: bool) -> Result<Arc<Object0>, ObjectError> {
    if do_fail {
        Err(ObjectError::InvalidOperation)
    } else {
        Ok(Arc::new(Object0::new()))
    }
}

pub struct Object1 {
    message: String,
}

impl Object1 {
    pub fn new(message: String) -> Object1 {
        return Object1 { message };
    }

    pub fn new_custom(message: String) -> Object1 {
        return Object1 { message };
    }

    pub fn get_message(&self) -> String {
        self.message.clone()
    }
}

pub struct FallibleObject0 {}

impl FallibleObject0 {
    pub fn new() -> Result<FallibleObject0, ObjectError> {
        Err(ObjectError::InvalidOperation)
    }

    pub fn new_custom() -> Result<FallibleObject0, ObjectError> {
        Err(ObjectError::InvalidOperation)
    }
}

pub struct FallibleObject1 {
    message: String,
}

impl FallibleObject1 {
    pub fn new(message: String) -> Result<FallibleObject1, ObjectError> {
        if message.is_empty() {
            Err(ObjectError::InvalidOperation)
        } else {
            Ok(FallibleObject1 { message })
        }
    }

    pub fn new_custom(message: String) -> Result<FallibleObject1, ObjectError> {
        if message.is_empty() {
            Err(ObjectError::InvalidOperation)
        } else {
            Ok(FallibleObject1 { message })
        }
    }

    pub fn get_message(&self) -> String {
        self.message.clone()
    }
}

pub struct SignalReceiver {
    sender: Sender<()>,
    receiver: Receiver<()>,
}

impl SignalReceiver {
    pub fn receive_signal(&self) {
        self.sender.send(()).unwrap();
        self.receiver.recv().unwrap();
        self.sender.send(()).unwrap();
    }

    pub fn heart_beat(&self) -> String {
        "whoosh".to_string()
    }
}

impl Drop for SignalReceiver {
    fn drop(&mut self) {
        *RECEIVER_COUNT.lock().unwrap() -= 1;
    }
}

pub struct SignalSender {
    sender: Sender<()>,
    receiver: Receiver<()>,
}

impl SignalSender {
    pub fn send_signal(&self) {
        self.sender.send(()).unwrap();
    }

    pub fn wait_for_receiver_to_appear(&self) {
        self.receiver.recv().unwrap();
    }

    pub fn wait_for_receiver_to_disappear(&self) {
        self.receiver.recv().unwrap();
    }
}

pub struct Channel {
    pub sender: Arc<SignalSender>,
    pub receiver: Arc<SignalReceiver>,
}

pub fn create_channel() -> Channel {
    let (tx1, rx1) = crossbeam::channel::unbounded::<()>();
    let (tx2, rx2) = crossbeam::channel::unbounded::<()>();

    let receiver = SignalReceiver {
        sender: tx2,
        receiver: rx1,
    };

    let sender = SignalSender {
        sender: tx1,
        receiver: rx2,
    };

    *RECEIVER_COUNT.lock().unwrap() += 1;

    Channel {
        sender: Arc::new(sender),
        receiver: Arc::new(receiver),
    }
}

pub fn create_fallible_object1(message: String) -> Result<Arc<FallibleObject1>, ObjectError> {
    FallibleObject1::new(message).map(|o| Arc::new(o))
}

pub fn return_object1(object: Arc<Object1>) -> Arc<Object1> {
    object
}

fn get_live_receiver_count() -> i32 {
    *RECEIVER_COUNT.lock().unwrap()
}

include!(concat!(env!("OUT_DIR"), "/objects.uniffi.rs"));
