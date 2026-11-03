package core

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/v2/graphemes"
	"github.com/go-playground/validator/v10"
)

const (
	// UsernameGlyphsMax caps the glyphs of a username.
	UsernameGlyphsMax = 32
	// UsernameBytesMax caps the bytes of a username, so invisible characters cannot pad it without
	// bound. The schema holds the same cap.
	UsernameBytesMax = 2048
	// usernameGlyphBytesMax caps the bytes of one glyph. The longest emoji takes 35, and UAX #15 caps a
	// run of combining marks at 30, so a stack of marks that spills over the lines around it does not
	// fit.
	usernameGlyphBytesMax = 64
)

// ValidateUsername reports whether a string, once trimmed of U+0020 at both ends, is empty, which
// clears the username, or a username. A username is valid UTF-8 of [UsernameBytesMax] bytes at most.
// It holds no control character, no space but U+0020, and no bidirectional embedding, override or
// isolate control, which reorder the text after the name.
//
// A username counts 1 to [UsernameGlyphsMax] glyphs. A glyph is a grapheme cluster (UAX #29) holding a
// code point a reader sees, one that is not Default_Ignorable_Code_Point. The package registers it as
// the "username" struct tag.
func ValidateUsername(fl validator.FieldLevel) bool {
	username := strings.Trim(fl.Field().String(), " ")
	if username == "" {
		return true
	}

	if len(username) > UsernameBytesMax || !utf8.ValidString(username) ||
		strings.ContainsFunc(username, refusedInUsername) {
		return false
	}

	var glyphs int

	clusters := graphemes.FromString(username)
	for clusters.Next() {
		cluster := clusters.Value()
		if len(cluster) > usernameGlyphBytesMax {
			return false
		}

		if strings.ContainsFunc(cluster, isVisible) {
			glyphs++
		}
	}

	return glyphs >= 1 && glyphs <= UsernameGlyphsMax
}

// refusedInUsername reports whether a username may not hold r: a control character, a space other
// than U+0020, or a bidirectional embedding, override or isolate control.
func refusedInUsername(r rune) bool {
	return unicode.Is(unicode.Cc, r) ||
		(r != ' ' && unicode.Is(unicode.White_Space, r)) ||
		(r >= '\u202A' && r <= '\u202E') ||
		(r >= '\u2066' && r <= '\u2069')
}

// isVisible reports whether r is no Default_Ignorable_Code_Point, the code points a renderer draws as
// nothing. It derives that property as DerivedCoreProperties.txt does (UAX #44):
//
//	Other_Default_Ignorable_Code_Point + Cf + Variation_Selector
//	- White_Space - U+FFF9..U+FFFB - U+13430..U+1343F - Prepended_Concatenation_Mark
func isVisible(r rune) bool {
	switch {
	case unicode.Is(unicode.White_Space, r),
		r >= '\uFFF9' && r <= '\uFFFB',
		r >= 0x13430 && r <= 0x1343F,
		unicode.Is(unicode.Prepended_Concatenation_Mark, r):
		return true
	default:
		return !unicode.In(r, unicode.Other_Default_Ignorable_Code_Point, unicode.Cf, unicode.Variation_Selector)
	}
}
