package core_test

import (
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"

	"github.com/a-novel/service-profile/internal/core"
)

func TestValidateUsername(t *testing.T) {
	t.Parallel()

	usernameValidator := validator.New()
	require.NoError(t, usernameValidator.RegisterValidation("username", core.ValidateUsername))

	// Each case gives the glyphs its username counts, or 0 for a refused username. Escapes spell out the
	// characters a reader cannot see.
	testCases := []struct {
		name string

		username string

		expectGlyphs int
	}{
		{name: "Success", username: "Jean Paul", expectGlyphs: 9},
		{name: "Success/Trimmed", username: " " + strings.Repeat("x", 32) + " ", expectGlyphs: 32},
		{name: "Success/ZeroWidthSpace", username: "a\u200Bb", expectGlyphs: 2},
		{name: "Success/FamilyEmoji", username: "\U0001F468\u200D\U0001F469\u200D\U0001F467", expectGlyphs: 1},
		{
			name:         "Success/FlagTagSequence",
			username:     "\U0001F3F4\U000E0067\U000E0062\U000E0065\U000E006E\U000E0067\U000E007F",
			expectGlyphs: 1,
		},
		{
			// Persian for "I want", whose zero-width non-joiner keeps two letters apart.
			name:         "Success/Persian",
			username:     "\u0645\u06CC\u200C\u062E\u0648\u0627\u0647\u0645",
			expectGlyphs: 7,
		},
		{name: "Success/CombiningMarks", username: "e" + strings.Repeat("\u0301", 31), expectGlyphs: 1},
		{name: "Success/PrependedConcatenationMark", username: "\u0600", expectGlyphs: 1},
		{name: "Success/InterlinearAnnotation", username: "\uFFF9", expectGlyphs: 1},
		{name: "Success/InvisibleAround", username: "\u00ADa\uFE0F", expectGlyphs: 1},

		{name: "Error/HangulFiller", username: "\u3164"},
		{name: "Error/SoftHyphen", username: "\u00AD"},
		{name: "Error/CarriageReturn", username: "a\rb"},
		{name: "Error/LineFeed", username: "a\nb"},
		{name: "Error/Tab", username: "a\tb"},
		{name: "Error/NoBreakSpace", username: "a\u00A0b"},
		{name: "Error/IdeographicSpace", username: "a\u3000b"},
		{name: "Error/RightToLeftOverride", username: "a\u202Eb"},
		{name: "Error/FirstStrongIsolate", username: "a\u2068b"},
		{name: "Error/TooManyGlyphs", username: strings.Repeat("x", 33)},
		{name: "Error/GlyphTooLong", username: "e" + strings.Repeat("\u0301", 32)},
		{name: "Error/FortyCombiningMarks", username: "e" + strings.Repeat("\u0301", 40)},
		{name: "Error/TooManyBytes", username: "a" + strings.Repeat("\u200B", 683)},
		{name: "Error/InvalidUTF8", username: "a\xffb"},
	}

	// A name that trims to nothing is no username but is valid: it clears the profile's username.
	t.Run("Success/Clear", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, usernameValidator.Var("", "username"))
		require.NoError(t, usernameValidator.Var("   ", "username"))
	})

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if testCase.expectGlyphs == 0 {
				require.Error(t, usernameValidator.Var(testCase.username, "username"))

				return
			}

			// The username counts exactly its glyphs when padding it to the cap fits, and one more glyph
			// does not. The padding goes first, since a prepended mark joins the character after it.
			padded := strings.Repeat("x", core.UsernameGlyphsMax-testCase.expectGlyphs) + testCase.username
			require.NoError(t, usernameValidator.Var(padded, "username"))
			require.Error(t, usernameValidator.Var("x"+padded, "username"))
		})
	}
}
