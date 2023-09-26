/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/callbacks"
	"github.com/stretchr/testify/assert"
)

type OnCallAnswerImpl struct {
	answerCount int
	// This will cause instances of `OnCallAnswerImpl` to be allocated on heap and will ensure
	// that finalizer functions are executed in tests
	ignored string
}

func (c *OnCallAnswerImpl) Answer() (string, *callbacks.TelephoneError) {
	c.answerCount += 1
	return fmt.Sprintf("hello, %d", c.answerCount), nil
}

type OnCallAnswerBusyImpl struct{}

func (OnCallAnswerBusyImpl) Answer() (string, *callbacks.TelephoneError) {
	return "", callbacks.NewTelephoneErrorBusy()
}

func TestCallbackWorks(t *testing.T) {
	telephone := callbacks.NewTelephone()
	callback := &OnCallAnswerImpl{}
	telephone.Call(callback)
	assert.Equal(t, 1, callback.answerCount)

	telephone.Call(callback)
	assert.Equal(t, 2, callback.answerCount)

	callback = &OnCallAnswerImpl{}
	telephone.Call(callback)
	assert.Equal(t, 1, callback.answerCount)

	callbackBusy := OnCallAnswerBusyImpl{}
	telephone.Call(callbackBusy)
}

func TestCallbackRegistrationIsNotAffectedByGC(t *testing.T) {
	telephone := callbacks.NewTelephone()
	callback := &OnCallAnswerImpl{}
	runtime.GC()
	telephone.Call(callback)
}

func TestCallbackReferenceIsDropped(t *testing.T) {
	telephone := callbacks.NewTelephone()
	dropped := false
	done :=  make(chan struct{})
	func() {
		callback := &OnCallAnswerImpl{}
		runtime.SetFinalizer(callback, func(cb *OnCallAnswerImpl) {
			dropped = true
			done <- struct{}{}
		})
		telephone.Call(callback)
	}()
	runtime.GC()
	// runtime.GC() is not a fully blocking call
	select {
	case <- time.After(time.Millisecond * 100):
		panic("timed out")
	case <- done:
		assert.Equal(t, true, dropped)
	}
}
