/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"testing"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/todolist"

	"github.com/stretchr/testify/assert"
)

func TestTodolistWorks(t *testing.T) {
	todo := todolist.NewTodoList()

	_, err := todo.GetLast()
	assert.ErrorIs(t, err, todolist.ErrTodoErrorEmptyTodoList)

	_, err = todolist.CreateEntryWith("")
	assert.ErrorIs(t, err, todolist.ErrTodoErrorEmptyString)

	todo.AddItem("Write strings support")
	last, err := todo.GetLast()
	if assert.NoError(t, err) {
		assert.Equal(t, "Write strings support", last)
	}

	todo.AddItem("Write tests for strings support")
	last, err = todo.GetLast()
	if assert.NoError(t, err) {
		assert.Equal(t, "Write tests for strings support", last)
	}

	entry, err := todolist.CreateEntryWith("Write bindings for strings as record members")
	todo.AddEntry(entry)

	last, err = todo.GetLast()
	if assert.NoError(t, err) {
		assert.Equal(t, "Write bindings for strings as record members", last)
	}

	last_entry, err := todo.GetLastEntry()
	if assert.NoError(t, err) {
		assert.Equal(t, "Write bindings for strings as record members", last_entry.Text)
	}

	todo.AddItem("Test Ünicode hàndling without an entry 🤣")
	last, err = todo.GetLast()
	if assert.NoError(t, err) {
		assert.Equal(t, "Test Ünicode hàndling without an entry 🤣", last)
	}

	entry2 := todolist.TodoEntry{"Test Ünicode hàndling in an entry 🤣"}
	todo.AddEntry(entry2)
	last, err = todo.GetLast()
	if assert.NoError(t, err) {
		assert.Equal(t, "Test Ünicode hàndling in an entry 🤣", last)
	}

	assert.Equal(t, 5, len(todo.GetEntries()))

	todo.AddEntries([]todolist.TodoEntry{todolist.TodoEntry{"foo"}, todolist.TodoEntry{"bar"}})
	assert.Equal(t, 7, len(todo.GetEntries()))
	last, err = todo.GetLast()
	if assert.NoError(t, err) {
		assert.Equal(t, "bar", last)
	}

	todo.AddItems([]string{"bobo", "fofo"})
	assert.Equal(t, 9, len(todo.GetEntries()))
	last, err = todo.GetLast()
	if assert.NoError(t, err) {
		assert.Equal(t, "fofo", last)
	}
}

func TestTodolistDefault(t *testing.T) {
	assert.Nil(t, todolist.GetDefaultList())

	todo := todolist.NewTodoList()
	defer todo.Destroy()
	todo.AddItems([]string{"foo", "bar"})

	todo2 := todolist.NewTodoList()
	defer todo2.Destroy()

	todolist.SetDefaultList(todo)
	{
		defaultList := *todolist.GetDefaultList()
		defer defaultList.Destroy()
		assert.Equal(t, todo.GetEntries(), defaultList.GetEntries())
		assert.NotEqual(t, todo2.GetEntries(), defaultList.GetEntries())
	}

	todo2.MakeDefault()
	{
		defaultList := *todolist.GetDefaultList()
		defer defaultList.Destroy()
		assert.NotEqual(t, todo.GetEntries(), defaultList.GetEntries())
		assert.Equal(t, todo2.GetEntries(), defaultList.GetEntries())
	}

	todo.AddItem("Test liveness after being demoted from default")
	last, err := todo.GetLast()
	if assert.NoError(t, err) {
		assert.Equal(t, "Test liveness after being demoted from default", last)
	}

	todo2.AddItem("Test shared state through local vs default reference")
	{
		defaultList := *todolist.GetDefaultList()
		defer defaultList.Destroy()
		last, err = defaultList.GetLast()
		if assert.NoError(t, err) {
			assert.Equal(t, "Test shared state through local vs default reference", last)
		}
	}
}
