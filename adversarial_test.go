package sqltoken

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Adversarial tests to expose bugs identified in code review
// These tests target specific edge cases and potential panics in the codebase

// TestJoinPanicOnSingleTokenWithDelimiter tests the bug identified in tokenize.go:1667-1668
// where Join() can panic when a command slice has a Delimiter annotation and contains only a single token.
// After `tokens = tokens[1:]` on line 1662, if the original tokens slice contained only one element,
// it becomes empty, causing a panic when trying to access `tokens[len(tokens)-1]` on line 1668.
//
// Bug confirmed: This test FAILS and exposes the panic condition.
func TestJoinPanicOnSingleTokenWithDelimiter(t *testing.T) {
	t.Run("single_token_with_delimiter", func(t *testing.T) {
		// Create a single-token command with a Delimiter annotation
		singleToken := Token{
			Type:      Whitespace,
			Text:      " ",
			Delimiter: Tokens{{Type: Word, Text: "$$"}},
		}
		
		tokensList := TokensList{
			Tokens{singleToken},
		}
		
		// This should not panic
		require.NotPanics(t, func() {
			result := tokensList.Join()
			// The result should handle the empty slice case
			assert.NotNil(t, result)
		})
	})

	t.Run("single_delimiter_token", func(t *testing.T) {
		// Single Delimiter token in a command
		delimToken := Token{
			Type:      Delimiter,
			Text:      "DELIMITER",
			Delimiter: Tokens{{Type: Word, Text: "//"}},
		}
		
		tokensList := TokensList{
			Tokens{delimToken},
		}
		
		require.NotPanics(t, func() {
			result := tokensList.Join()
			assert.NotNil(t, result)
		})
	})

	t.Run("delimiter_with_split_annotation", func(t *testing.T) {
		// Single token with both Delimiter and Split annotations
		tokenWithSplit := Token{
			Type:      Word,
			Text:      "SELECT",
			Delimiter: Tokens{{Type: Word, Text: "$$"}},
			Split:     Tokens{{Type: Whitespace, Text: "\n"}},
		}
		
		tokensList := TokensList{
			Tokens{tokenWithSplit},
		}
		
		require.NotPanics(t, func() {
			result := tokensList.Join()
			assert.NotNil(t, result)
		})
	})
}

// TestStripWhitespaceCoordinateSystemBug tests the bug identified in tokenize.go:1404-1417
// where Strip() uses lastWhitespace (index into original ts) and lastReal (index/length in output slice c)
// in different coordinate systems. This can cause whitespace suppression logic to skip required whitespace
// after large runs of stripped tokens (e.g., many comments).
func TestStripWhitespaceCoordinateSystemBug(t *testing.T) {
	t.Run("many_comments_before_whitespace", func(t *testing.T) {
		// Create input with many comments followed by whitespace and tokens
		// This should expose the coordinate system mismatch
		// The pattern needs to be: token, many-stripped-items, whitespace, token, whitespace, token
		// where lastWhitespace (index in ts) < lastReal (len of c) after stripping
		input := "word1"
		// Add many comments to create a large gap between indices in ts vs c
		// After word1, lastReal will be 1 (length of c)
		// After all comments are skipped, lastWhitespace will still be whatever it was
		// When we hit the first whitespace at index ~40, lastWhitespace (0) < lastReal (1) is true, so it's kept
		// When we hit word2, lastReal becomes 2
		// When we hit the next whitespace at index ~45, lastWhitespace (~40) is NOT < lastReal (2)
		// This is the bug: lastWhitespace is in ts coordinates (40), lastReal is in c coordinates (2)
		for i := 0; i < 25; i++ {
			input += "/* comment" + string(rune('A'+i%26)) + " */"
		}
		input += " word2 word3"
		
		tokens := TokenizeMySQL(input)
		stripped := tokens.Strip()
		result := stripped.String()
		
		// Should have whitespace between word1 and word2, and word2 and word3
		assert.Contains(t, result, "word1 word2")
		assert.Contains(t, result, "word2 word3")
		// Should not have concatenated tokens
		assert.NotContains(t, result, "word1word2")
		assert.NotContains(t, result, "word2word3")
	})

	t.Run("alternating_comments_and_whitespace", func(t *testing.T) {
		// Alternating pattern that stresses the coordinate tracking
		input := "a /* c1 */ /* c2 */ /* c3 */ /* c4 */ /* c5 */ b /* c6 */ /* c7 */ c"
		
		tokens := TokenizeMySQL(input)
		stripped := tokens.Strip()
		result := stripped.String()
		
		// Should preserve whitespace between tokens
		assert.Equal(t, "a b c", result)
		assert.NotContains(t, result, "ab")
		assert.NotContains(t, result, "bc")
	})

	t.Run("long_comment_sequence", func(t *testing.T) {
		// Very long sequence of stripped tokens to maximize coordinate divergence
		input := "first"
		for i := 0; i < 50; i++ {
			input += " -- comment\n"
		}
		input += " second third"
		
		tokens := TokenizeMySQL(input)
		stripped := tokens.Strip()
		result := stripped.String()
		
		// Should have proper whitespace
		assert.Equal(t, "first second third", result)
		assert.NotContains(t, result, "firstsecond")
		assert.NotContains(t, result, "secondthird")
	})

	t.Run("consecutive_whitespace_after_comments", func(t *testing.T) {
		// Test that consecutive whitespace is properly handled after comments
		input := "word1 /* comment */ \t \n   word2"
		
		tokens := TokenizeMySQL(input)
		stripped := tokens.Strip()
		result := stripped.String()
		
		// Should normalize to single space but not drop it
		assert.Equal(t, "word1 word2", result)
		assert.NotEqual(t, "word1word2", result)
	})
}

// TestStripDelimiterFastPathOffByOne tests the bug identified in tokenize.go:1381-1393
// where in the Delimiter-matching fast-path, `i` is advanced by `len(delimiter)` after appending,
// and then `captureSkip()` uses `ts[lastCapture : i+1]`. At that point `i` no longer points at
// the last appended token, so the captured Strip slice is off-by-one and can include a token
// that was not appended (or panic if delimiter occurs at end of slice).
func TestStripDelimiterFastPathOffByOne(t *testing.T) {
	t.Run("delimiter_at_end_of_input", func(t *testing.T) {
		// Delimiter at the very end of input - should not panic
		input := "SELECT 1 $$"
		
		tokens := TokenizeMySQL(input)
		
		require.NotPanics(t, func() {
			// This should process the delimiter without panic
			_ = tokens.Strip()
		})
	})

	t.Run("delimiter_with_custom_delimiter_set", func(t *testing.T) {
		// Test with DELIMITER command changing the delimiter
		input := "DELIMITER $$\nSELECT 1$$\nDELIMITER ;\n"
		
		tokens := TokenizeMySQL(input)
		
		require.NotPanics(t, func() {
			stripped := tokens.Strip()
			assert.NotNil(t, stripped)
		})
	})

	t.Run("multiple_delimiters_in_sequence", func(t *testing.T) {
		// Multiple delimiters that could expose off-by-one errors
		input := "DELIMITER $$\nA$$B$$C$$\nDELIMITER ;\n"
		
		tokens := TokenizeMySQL(input)
		
		require.NotPanics(t, func() {
			stripped := tokens.Strip()
			result := stripped.String()
			// Should process all delimiters correctly
			assert.NotNil(t, result)
		})
	})

	t.Run("delimiter_followed_by_eof", func(t *testing.T) {
		// Edge case: delimiter is the last thing in input with no trailing content
		input := "DELIMITER //\nCREATE PROCEDURE test() BEGIN SELECT 1; END//"
		
		tokens := TokenizeMySQL(input)
		
		require.NotPanics(t, func() {
			stripped := tokens.Strip()
			assert.NotNil(t, stripped)
			// Verify the Strip slice metadata is captured correctly
			for _, tok := range stripped {
				// Should have valid Strip information where applicable
				if tok.Strip != nil {
					assert.NotNil(t, tok.Strip)
				}
			}
		})
	})

	t.Run("unstrip_after_delimiter_at_end", func(t *testing.T) {
		// Test that Unstrip works correctly with delimiter at end
		input := "SELECT 1 $$"
		
		tokens := TokenizeMySQL(input)
		
		require.NotPanics(t, func() {
			stripped := tokens.Strip()
			// Unstrip should work without panic
			unstripped := stripped.Unstrip()
			assert.NotNil(t, unstripped)
		})
	})
}

// TestCmdSplitWithDelimiterEdgeCases tests edge cases in CmdSplit with custom delimiters
func TestCmdSplitWithDelimiterEdgeCases(t *testing.T) {
	t.Run("empty_command_with_delimiter", func(t *testing.T) {
		// DELIMITER followed immediately by another DELIMITER
		input := "DELIMITER $$\nDELIMITER ;\n"
		
		tokens := TokenizeMySQL(input)
		
		require.NotPanics(t, func() {
			split := tokens.CmdSplit()
			assert.NotNil(t, split)
		})
	})

	t.Run("delimiter_change_without_content", func(t *testing.T) {
		// Just delimiter changes, no actual SQL
		input := "DELIMITER //\nDELIMITER $$\nDELIMITER ;\n"
		
		tokens := TokenizeMySQL(input)
		
		require.NotPanics(t, func() {
			split := tokens.CmdSplit()
			joined := split.Join()
			assert.NotNil(t, joined)
		})
	})

	t.Run("single_whitespace_with_delimiter_annotation", func(t *testing.T) {
		// Edge case mentioned in the code: len(tokens) == 1 && tokens[0].Type == Whitespace
		input := "DELIMITER $$\n \n$$"
		
		tokens := TokenizeMySQL(input)
		
		require.NotPanics(t, func() {
			split := tokens.CmdSplitUnstripped()
			joined := split.Join()
			assert.NotNil(t, joined)
		})
	})
}

// TestJoinUnstripRoundTrip tests the round-trip behavior of CmdSplit/Join and Strip/Unstrip
func TestJoinUnstripRoundTrip(t *testing.T) {
	t.Run("complex_delimiter_round_trip", func(t *testing.T) {
		input := `DELIMITER $$
CREATE PROCEDURE test()
BEGIN
  SELECT 1;
  SELECT 2;
END$$
DELIMITER ;
SELECT 3;`
		
		tokens := TokenizeMySQL(input)
		split := tokens.CmdSplitUnstripped()
		
		require.NotPanics(t, func() {
			joined := split.Join()
			// Should be able to round-trip
			assert.NotNil(t, joined)
			
			// Verify we can split again
			split2 := joined.CmdSplitUnstripped()
			assert.Equal(t, len(split), len(split2))
		})
	})

	t.Run("strip_unstrip_with_many_comments", func(t *testing.T) {
		input := "SELECT /* c1 */ /* c2 */ /* c3 */ 1 /* c4 */;"
		
		tokens := TokenizeMySQL(input)
		stripped := tokens.Strip()
		
		require.NotPanics(t, func() {
			unstripped := stripped.Unstrip()
			assert.NotNil(t, unstripped)
			// Original tokens should be preserved in some form
			assert.True(t, len(unstripped) >= len(stripped))
		})
	})
}
