/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"testing"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/uniffi/destroy"
	"github.com/stretchr/testify/assert"
)

func TestDestroyObject(t *testing.T) {
	journal := destroy.CreateJournal()
	resource := destroy.NewResource()
	journal.Object = &resource
	assert.Equal(t, int32(1), destroy.GetLiveCount())
	journal.Destroy()
	assert.Equal(t, int32(0), destroy.GetLiveCount())
}

func TestDestroyRecord(t *testing.T) {
	journal := destroy.CreateJournal()
	smallJournal := destroy.SmallJournal{destroy.NewResource()}
	journal.Record = &smallJournal
	assert.Equal(t, int32(1), destroy.GetLiveCount())
	journal.Destroy()
	assert.Equal(t, int32(0), destroy.GetLiveCount())
}

func TestDestroyMap(t *testing.T) {
	journal := destroy.CreateJournal()
	resources := make(map[int32]*destroy.Resource)
	resources[0] = destroy.NewResource()
	resources[1] = destroy.NewResource()
	journal.Map = &resources
	assert.Equal(t, int32(2), destroy.GetLiveCount())
	journal.Destroy()
	assert.Equal(t, int32(0), destroy.GetLiveCount())
}

func TestDestroySequence(t *testing.T) {
	journal := destroy.CreateJournal()
	resources := []*destroy.Resource{
		destroy.NewResource(),
		destroy.NewResource(),
	}
	journal.List = &resources
	assert.Equal(t, int32(2), destroy.GetLiveCount())
	journal.Destroy()
	assert.Equal(t, int32(0), destroy.GetLiveCount())
}

func TestDestroyEnum(t *testing.T) {
	journal := destroy.CreateJournal()
	var enum destroy.EnumJournal = destroy.EnumJournalJournal{destroy.SmallJournal{destroy.NewResource()}}
	journal.Enum = &enum
	assert.Equal(t, int32(1), destroy.GetLiveCount())
	journal.Destroy()
	assert.Equal(t, int32(0), destroy.GetLiveCount())
}
