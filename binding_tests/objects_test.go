/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"runtime"
	"testing"
	"time"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/objects"

	"github.com/stretchr/testify/assert"
)

func TestObject0(t *testing.T) {
	object := objects.NewObject0()
	assert.NotNil(t, object)

	object = objects.Object0NewCustom()
	assert.NotNil(t, object)
}

func TestObject1(t *testing.T) {
	object := objects.NewObject1("hello, world")
	assert.Equal(t, "hello, world", object.GetMessage())

	object = objects.Object1NewCustom("bye bye")
	assert.Equal(t, "bye bye", object.GetMessage())
}

func TestFallibleObject0(t *testing.T) {
	object, err := objects.NewFallibleObject0()
	assert.ErrorIs(t, err, objects.ErrObjectErrorInvalidOperation)
	assert.Nil(t, object)

	object, err = objects.FallibleObject0NewCustom()
	assert.ErrorIs(t, err, objects.ErrObjectErrorInvalidOperation)
	assert.Nil(t, object)
}

func TestFallibleObject1(t *testing.T) {
	object, err := objects.NewFallibleObject1("")
	assert.ErrorIs(t, err, objects.ErrObjectErrorInvalidOperation)
	assert.Nil(t, object)

	object, err = objects.FallibleObject1NewCustom("")
	assert.ErrorIs(t, err, objects.ErrObjectErrorInvalidOperation)
	assert.Nil(t, object)

	object, err = objects.NewFallibleObject1("hello, world")
	if assert.Nil(t, err) {
		assert.Equal(t, "hello, world", object.GetMessage())
	}

	object, err = objects.FallibleObject1NewCustom("hello, world")
	if assert.Nil(t, err) {
		assert.Equal(t, "hello, world", object.GetMessage())
	}
}

func TestReturnObject(t *testing.T) {
	// This test verifies a very weird situation. `object1` and `object2` point to the same
	// underlying Rust object. However, the Go object represents not the instance of the object,
	// but an instance of `std::sync::Arc`. Therefore, both `object1` and `object2` must be
	// `.Destroy()`ed to actually drop the underlying Rust object.

	arc1 := objects.NewObject1("hello, world")
	defer arc1.Destroy()

	arc2 := objects.ReturnObject1(arc1)
	arc2.Destroy()

	assert.Equal(t, "hello, world", arc1.GetMessage())
	assert.PanicsWithError(t, "*Object1 object has already been destroyed", func() {
		_ = arc2.GetMessage()
	})
}

func TestDoubleDestroy(t *testing.T) {
	object := objects.NewObject1("hello world")
	object.Destroy()
	object.Destroy()
	assert.PanicsWithError(t, "*Object1 object has already been destroyed", func() {
		_ = object.GetMessage()
	})
}

func TestDestroyAcrossReferences(t *testing.T) {
	object1 := objects.NewObject1("hello world")
	object2 := object1
	object2.Destroy()
	assert.PanicsWithError(t, "*Object1 object has already been destroyed", func() {
		_ = object1.GetMessage()
	})
}

func TestDestroyWithInFlightCalls(t *testing.T) {
	channel := objects.CreateChannel()

	go func() {
		channel.Receiver.ReceiveSignal()
	}()
	channel.Sender.WaitForReceiverToAppear()

	// At this point `ReceiveSignal` is in-progress, and `Destroy` won't have an effect
	// until `SendSignal` is dispatched.
	channel.Receiver.Destroy()
	channel.Receiver.Destroy()
	assert.Equal(t, "whoosh", channel.Receiver.HeartBeat())

	channel.Sender.SendSignal()
	channel.Sender.WaitForReceiverToDisappear()
	// When the receiver disappears from Rust, the other thread still needs a moment to
	// return from `ReceiveSignal`, and to call `freeRustArcPtr` upon ending the call
	time.Sleep(1 * time.Millisecond)

	// At this point all in-progress calls are finished, and object should be destroyed
	assert.PanicsWithError(t, "*SignalReceiver object has already been destroyed", func() {
		_ = channel.Receiver.HeartBeat()
	})
}

func TestDestroyWithGC(t *testing.T) {
	channel := objects.CreateChannel()
	assert.Equal(t, int32(1), objects.GetLiveReceiverCount())

	// Receiver is alive until this line
	runtime.KeepAlive(channel.Receiver)

	// GC the Receiver
	runtime.GC()
	// a brief moment for finalizer to run
	time.Sleep(1 * time.Millisecond)

	assert.Equal(t, int32(0), objects.GetLiveReceiverCount())
}

func TestGcDoesNotDestroyObjectsWithInFlightCalls(t *testing.T) {
	channel := objects.CreateChannel()
	assert.Equal(t, int32(1), objects.GetLiveReceiverCount())

	go func() {
		channel.Receiver.ReceiveSignal()
	}()
	// at this point Receiver is no longer referenced anywhere, except for in-flight binding call

	channel.Sender.WaitForReceiverToAppear()

	// attempt to GC the Receiver
	runtime.GC()
	// a brief moment for the finalizer to run
	time.Sleep(1 * time.Millisecond)
	// the signal hasn't been sent yet, so its expected that `ReceiveSignal` is still in-flight
	// and Receiver should not have been GC'ed
	assert.Equal(t, int32(1), objects.GetLiveReceiverCount())

	channel.Sender.SendSignal()
	channel.Sender.WaitForReceiverToDisappear()
	// When the Receiver disappears from Rust, the other thread still needs a moment to
	// return from `ReceiveSignal`, and to call `freeRustArcPtr` upon ending the call
	time.Sleep(1 * time.Millisecond)

	// GC the Receiver
	runtime.GC()
	// a brief moment for the finalizer to run
	time.Sleep(1 * time.Millisecond)
	// at this point the Receiver should have been GC'ed
	assert.Equal(t, int32(0), objects.GetLiveReceiverCount())
}
