package server

import (
	"slices"
	"testing"

	"github.com/goplus/goxlsw/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerSpxGetDefinitions(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		s := New(newMapFSWithoutModTime(map[string][]byte{
			"main.spx": []byte(`
var (
	MySprite Sprite
)
MySprite.turn Left
run "assets", {Title: "My Game"}
`),
			"MySprite.spx": []byte(`
onStart => {
	MySprite.turn Right
}
`),
			"assets/index.json":                  []byte(`{}`),
			"assets/sprites/MySprite/index.json": []byte(`{}`),
		}), nil)

		mainSpxFileScopeParams := []SpxGetDefinitionsParams{
			{
				TextDocumentPositionParams: TextDocumentPositionParams{
					TextDocument: TextDocumentIdentifier{URI: "file:///main.spx"},
					Position:     Position{Line: 0, Character: 0},
				},
			},
		}
		mainSpxFileScopeDefs, err := s.spxGetDefinitions(mainSpxFileScopeParams)
		require.NoError(t, err)
		require.NotNil(t, mainSpxFileScopeDefs)
		assert.True(t, spxDefinitionIdentifierSliceContains(mainSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr("builtin"),
			Name:    util.ToPtr("println"),
		}))
		assert.True(t, spxDefinitionIdentifierSliceContains(mainSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr("main"),
			Name:    util.ToPtr("MySprite"),
		}))
		assert.True(t, spxDefinitionIdentifierSliceContains(mainSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package:    util.ToPtr(GetSpxPkg().Path()),
			Name:       util.ToPtr("Game.play"),
			OverloadID: util.ToPtr("1"),
		}))
		assert.True(t, spxDefinitionIdentifierSliceContains(mainSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr(GetSpxPkg().Path()),
			Name:    util.ToPtr("Game.onStart"),
		}))
		assert.True(t, spxDefinitionIdentifierSliceContains(mainSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr(GetSpxPkg().Path()),
			Name:    util.ToPtr("Game.onClick"),
		}))
		assert.False(t, spxDefinitionIdentifierSliceContains(mainSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package:    util.ToPtr(GetSpxPkg().Path()),
			Name:       util.ToPtr("Sprite.turn"),
			OverloadID: util.ToPtr("1"),
		}))
		assert.False(t, spxDefinitionIdentifierSliceContains(mainSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr(GetSpxPkg().Path()),
			Name:    util.ToPtr("Sprite.onStart"),
		}))

		mySpriteSpxFileScopeParams := []SpxGetDefinitionsParams{
			{
				TextDocumentPositionParams: TextDocumentPositionParams{
					TextDocument: TextDocumentIdentifier{URI: "file:///MySprite.spx"},
					Position:     Position{Line: 0, Character: 0},
				},
			},
		}
		mySpriteSpxFileScopeDefs, err := s.spxGetDefinitions(mySpriteSpxFileScopeParams)
		require.NoError(t, err)
		require.NotNil(t, mySpriteSpxFileScopeDefs)
		assert.True(t, spxDefinitionIdentifierSliceContains(mySpriteSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr("builtin"),
			Name:    util.ToPtr("println"),
		}))
		assert.True(t, spxDefinitionIdentifierSliceContains(mySpriteSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package:    util.ToPtr(GetSpxPkg().Path()),
			Name:       util.ToPtr("Game.play"),
			OverloadID: util.ToPtr("1"),
		}))
		assert.False(t, spxDefinitionIdentifierSliceContains(mySpriteSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr(GetSpxPkg().Path()),
			Name:    util.ToPtr("Game.onStart"),
		}))
		assert.True(t, spxDefinitionIdentifierSliceContains(mySpriteSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package:    util.ToPtr(GetSpxPkg().Path()),
			Name:       util.ToPtr("Sprite.turn"),
			OverloadID: util.ToPtr("1"),
		}))
		assert.True(t, spxDefinitionIdentifierSliceContains(mySpriteSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr(GetSpxPkg().Path()),
			Name:    util.ToPtr("Sprite.onStart"),
		}))
		assert.True(t, spxDefinitionIdentifierSliceContains(mySpriteSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr(GetSpxPkg().Path()),
			Name:    util.ToPtr("Sprite.onClick"),
		}))

		mySpriteSpxOnStartScopeParams := []SpxGetDefinitionsParams{
			{
				TextDocumentPositionParams: TextDocumentPositionParams{
					TextDocument: TextDocumentIdentifier{URI: "file:///MySprite.spx"},
					Position:     Position{Line: 2, Character: 0},
				},
			},
		}
		mySpriteSpxOnStartScopeDefs, err := s.spxGetDefinitions(mySpriteSpxOnStartScopeParams)
		require.NoError(t, err)
		require.NotNil(t, mySpriteSpxOnStartScopeDefs)
		assert.True(t, spxDefinitionIdentifierSliceContains(mySpriteSpxOnStartScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr("builtin"),
			Name:    util.ToPtr("println"),
		}))
		assert.True(t, spxDefinitionIdentifierSliceContains(mySpriteSpxOnStartScopeDefs, SpxDefinitionIdentifier{
			Package:    util.ToPtr(GetSpxPkg().Path()),
			Name:       util.ToPtr("Game.play"),
			OverloadID: util.ToPtr("1"),
		}))
		assert.False(t, spxDefinitionIdentifierSliceContains(mySpriteSpxOnStartScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr(GetSpxPkg().Path()),
			Name:    util.ToPtr("Game.onStart"),
		}))
		assert.False(t, spxDefinitionIdentifierSliceContains(mySpriteSpxOnStartScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr(GetSpxPkg().Path()),
			Name:    util.ToPtr("Game.onClick"),
		}))
		assert.True(t, spxDefinitionIdentifierSliceContains(mySpriteSpxOnStartScopeDefs, SpxDefinitionIdentifier{
			Package:    util.ToPtr(GetSpxPkg().Path()),
			Name:       util.ToPtr("Sprite.turn"),
			OverloadID: util.ToPtr("1"),
		}))
		assert.False(t, spxDefinitionIdentifierSliceContains(mySpriteSpxOnStartScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr(GetSpxPkg().Path()),
			Name:    util.ToPtr("Sprite.onStart"),
		}))
		assert.False(t, spxDefinitionIdentifierSliceContains(mySpriteSpxOnStartScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr(GetSpxPkg().Path()),
			Name:    util.ToPtr("Sprite.onClick"),
		}))
		assert.True(t, spxDefinitionIdentifierSliceContains(mySpriteSpxOnStartScopeDefs, SpxDefinitionIdentifier{
			Package:    util.ToPtr(GetSpxPkg().Path()),
			Name:       util.ToPtr("Sprite.clone"),
			OverloadID: util.ToPtr("1"),
		}))
	})

	t.Run("ParseError", func(t *testing.T) {
		s := New(newMapFSWithoutModTime(map[string][]byte{
			"main.spx": []byte(`
// Invalid syntax
var (
	MySprite Sprite
`),
		}), nil)

		mainSpxFileScopeParams := []SpxGetDefinitionsParams{
			{
				TextDocumentPositionParams: TextDocumentPositionParams{
					TextDocument: TextDocumentIdentifier{URI: "file:///main.spx"},
					Position:     Position{Line: 0, Character: 0},
				},
			},
		}
		mainSpxFileScopeDefs, err := s.spxGetDefinitions(mainSpxFileScopeParams)
		require.NoError(t, err)
		require.NotNil(t, mainSpxFileScopeDefs)
		assert.True(t, spxDefinitionIdentifierSliceContains(mainSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr("builtin"),
			Name:    util.ToPtr("println"),
		}))
		assert.True(t, spxDefinitionIdentifierSliceContains(mainSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr("main"),
			Name:    util.ToPtr("Game.MySprite"),
		}))
		assert.True(t, spxDefinitionIdentifierSliceContains(mainSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr(GetSpxPkg().Path()),
			Name:    util.ToPtr("Game.onStart"),
		}))
		assert.False(t, spxDefinitionIdentifierSliceContains(mainSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr(GetSpxPkg().Path()),
			Name:    util.ToPtr("Sprite.onStart"),
		}))
	})

	t.Run("TrailingEmptyLinesOfSpriteCode", func(t *testing.T) {
		s := New(newMapFSWithoutModTime(map[string][]byte{
			"main.spx": []byte(`
var (
	MySprite Sprite
)
MySprite.turn Left
run "assets", {Title: "My Game"}
`),
			"MySprite.spx": []byte(`
onStart => {
	MySprite.turn Right
}


`),
			"assets/index.json":                  []byte(`{}`),
			"assets/sprites/MySprite/index.json": []byte(`{}`),
		}), nil)

		mySpriteSpxFileScopeParams := []SpxGetDefinitionsParams{
			{
				TextDocumentPositionParams: TextDocumentPositionParams{
					TextDocument: TextDocumentIdentifier{URI: "file:///MySprite.spx"},
					Position:     Position{Line: 5, Character: 0},
				},
			},
		}
		mySpriteSpxFileScopeDefs, err := s.spxGetDefinitions(mySpriteSpxFileScopeParams)
		require.NoError(t, err)
		require.NotNil(t, mySpriteSpxFileScopeDefs)
		assert.True(t, spxDefinitionIdentifierSliceContains(mySpriteSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package:    util.ToPtr(GetSpxPkg().Path()),
			Name:       util.ToPtr("Game.play"),
			OverloadID: util.ToPtr("1"),
		}))
		assert.False(t, spxDefinitionIdentifierSliceContains(mySpriteSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr(GetSpxPkg().Path()),
			Name:    util.ToPtr("Game.onStart"),
		}))
		assert.True(t, spxDefinitionIdentifierSliceContains(mySpriteSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr(GetSpxPkg().Path()),
			Name:    util.ToPtr("Sprite.onStart"),
		}))
		assert.True(t, spxDefinitionIdentifierSliceContains(mySpriteSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr(GetSpxPkg().Path()),
			Name:    util.ToPtr("Sprite.onClick"),
		}))
	})

	t.Run("EOF", func(t *testing.T) {
		s := New(newMapFSWithoutModTime(map[string][]byte{
			"main.spx": []byte(`
onStart => {}
`),
		}), nil)

		mainSpxOnStartScopeParams := []SpxGetDefinitionsParams{
			{
				TextDocumentPositionParams: TextDocumentPositionParams{
					TextDocument: TextDocumentIdentifier{URI: "file:///main.spx"},
					Position:     Position{Line: 1, Character: 13},
				},
			},
		}
		mainSpxOnStartScopeDefs, err := s.spxGetDefinitions(mainSpxOnStartScopeParams)
		require.NoError(t, err)
		require.NotNil(t, mainSpxOnStartScopeDefs)
		assert.False(t, spxDefinitionIdentifierSliceContains(mainSpxOnStartScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr(GetSpxPkg().Path()),
			Name:    util.ToPtr("Game.onStart"),
		}))

		mainSpxFileScopeParams := []SpxGetDefinitionsParams{
			{
				TextDocumentPositionParams: TextDocumentPositionParams{
					TextDocument: TextDocumentIdentifier{URI: "file:///main.spx"},
					Position:     Position{Line: 2, Character: 0},
				},
			},
		}
		mainSpxFileScopeDefs, err := s.spxGetDefinitions(mainSpxFileScopeParams)
		require.NoError(t, err)
		require.NotNil(t, mainSpxFileScopeDefs)
		assert.True(t, spxDefinitionIdentifierSliceContains(mainSpxFileScopeDefs, SpxDefinitionIdentifier{
			Package: util.ToPtr(GetSpxPkg().Path()),
			Name:    util.ToPtr("Game.onStart"),
		}))
	})

	// See https://github.com/goplus/builder/issues/1398.
	t.Run("Issue#1398", func(t *testing.T) {
		s := New(newMapFSWithoutModTime(map[string][]byte{
			"main.spx": []byte(`
run "assets", {Title: "My Game"}
`),
			"MySprite.spx": []byte(`
onTouchStart "" => {
	say "touched by someone"
}
`),
			"assets/index.json":                  []byte(`{}`),
			"assets/sprites/MySprite/index.json": []byte(`{}`),
		}), nil)

		params := []SpxGetDefinitionsParams{
			{
				TextDocumentPositionParams: TextDocumentPositionParams{
					TextDocument: TextDocumentIdentifier{URI: "file:///MySprite.spx"},
					Position:     Position{Line: 1, Character: 15},
				},
			},
		}
		defs, err := s.spxGetDefinitions(params)
		require.NoError(t, err)
		require.NotNil(t, defs)
		assert.False(t, spxDefinitionIdentifierSliceContains(defs, SpxDefinitionIdentifier{
			Package: util.ToPtr(GetSpxPkg().Path()),
			Name:    util.ToPtr("Sprite.onTouchStart"),
		}))
		assert.True(t, spxDefinitionIdentifierSliceContains(defs, SpxDefinitionIdentifier{
			Package:    util.ToPtr(GetSpxPkg().Path()),
			Name:       util.ToPtr("Sprite.say"),
			OverloadID: util.ToPtr("0"),
		}))
	})
}

// spxDefinitionIdentifierSliceContains reports whether a slice of [SpxDefinitionIdentifier]
// contains a specific [SpxDefinitionIdentifier].
func spxDefinitionIdentifierSliceContains(defs []SpxDefinitionIdentifier, def SpxDefinitionIdentifier) bool {
	return slices.ContainsFunc(defs, func(d SpxDefinitionIdentifier) bool {
		return util.FromPtr(d.Package) == util.FromPtr(def.Package) &&
			util.FromPtr(d.Name) == util.FromPtr(def.Name) &&
			util.FromPtr(d.OverloadID) == util.FromPtr(def.OverloadID)
	})
}
