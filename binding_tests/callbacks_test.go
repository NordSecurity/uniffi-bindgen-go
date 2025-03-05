// TODO(pna): need to use new callback logic

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

type GoSim struct{}

func (sim *GoSim) Name() string {
	return "go"
}

func TestCallbackWorks(t *testing.T) {
	cases := []struct {
		name  string
		phone callbacks.TelephoneIterface
	}{
		{"simple", callbacks.NewTelephone()},
		{"fancy", callbacks.NewFancyTelephone()},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			telephone := callbacks.NewTelephone()
			callback := &OnCallAnswerImpl{}
			msg, err := c.phone.Call(callbacks.GetSimCards()[0], callback)
			if assert.NoError(t, err) {
				assert.Equal(t, 1, callback.answerCount)
				assert.Equal(t, "hello, 1", msg)
			}

			msg, err = telephone.Call(callbacks.GetSimCards()[0], callback)
			if assert.NoError(t, err) {
				assert.Equal(t, 2, callback.answerCount)
				assert.Equal(t, "hello, 2", msg)
			}

			callback = &OnCallAnswerImpl{}
			msg, err = telephone.Call(callbacks.GetSimCards()[0], callback)
			if assert.NoError(t, err) {
				assert.Equal(t, 1, callback.answerCount)
				assert.Equal(t, "hello, 1", msg)
			}

			callbackBusy := OnCallAnswerBusyImpl{}
			msg, err = telephone.Call(callbacks.GetSimCards()[0], callbackBusy)
			if assert.Error(t, err) {
				assert.ErrorIs(t, err, callbacks.ErrTelephoneErrorBusy)
			}

			sim := &GoSim{}
			msg, err = telephone.Call(sim, callback)
			if assert.NoError(t, err) {
				assert.Equal(t, "go est bon marché", msg)
			}
		})
	}
}

func TestCallbackRegistrationIsNotAffectedByGC(t *testing.T) {
	telephone := callbacks.NewTelephone()
	callback := &OnCallAnswerImpl{}
	runtime.GC()
	telephone.Call(callbacks.GetSimCards()[0], callback)
}

func TestCallbackReferenceIsDropped(t *testing.T) {
	telephone := callbacks.NewTelephone()
	dropped := false
	done := make(chan struct{})
	func() {
		callback := &OnCallAnswerImpl{}
		runtime.SetFinalizer(callback, func(cb *OnCallAnswerImpl) {
			dropped = true
			done <- struct{}{}
		})
		telephone.Call(callbacks.GetSimCards()[0], callback)
	}()
	runtime.GC()
	// runtime.GC() is not a fully blocking call
	select {
	case <-time.After(time.Millisecond * 100):
		panic("timed out")
	case <-done:
		assert.Equal(t, true, dropped)
	}
}
