/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package binding_tests

import (
	"testing"

	"github.com/NordSecurity/uniffi-bindgen-go/binding_tests/generated/sprites"

	"github.com/stretchr/testify/assert"
)

func TestSpritesWork(t *testing.T) {
	{
		sprite_empty := sprites.NewSprite(nil)
		defer sprite_empty.Destroy()

		assert.Equal(t, sprites.Point{0, 0}, sprite_empty.GetPosition())
	}

	{
		sprite := sprites.NewSprite(&sprites.Point{0, 1})
		defer sprite.Destroy()

		assert.Equal(t, sprites.Point{0, 1}, sprite.GetPosition())

		sprite.MoveTo(sprites.Point{1, 2})
		assert.Equal(t, sprites.Point{1, 2}, sprite.GetPosition())

		sprite.MoveBy(sprites.Vector{-4, 2})
		assert.Equal(t, sprites.Point{-3, 4}, sprite.GetPosition())
	}

	{
		sprite := sprites.SpriteNewRelativeTo(sprites.Point{0, 1}, sprites.Vector{1, 1.5})
		defer sprite.Destroy()

		assert.Equal(t, sprites.Point{1, 2.5}, sprite.GetPosition())
	}
}
