package sqltoken

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStrip(t *testing.T) {
	cases := []struct {
		before string
		after  string
	}{
		{
			before: "",
			after:  "",
		},
		{
			before: "-- stuff1\n",
			after:  "",
		},
		{
			before: "/* foo2 */ bar \n baz  ; ",
			after:  "bar baz",
		},
		{
			before: " /* foo3 */ bar \n baz  ; ",
			after:  "bar baz",
		},
		{
			before: "\t\talpha4  \n\n beta\t ;  ",
			after:  "alpha4 beta",
		},
		{
			before: "word/*c5*/\t/*c6*/word2 ; ",
			after:  "word word2",
		},
		{
			before: "  -- c7\n  word ; -- c2\n",
			after:  "word",
		},
		{
			before: "x8 ; ; ;  ",
			after:  "x8",
		},
		{
			before: "a9;b;c;",
			after:  "a9;b;c",
		},
		{
			before: ";\n\t ; /* c10 */ ; -- tail",
			after:  "",
		},
		{
			before: "  ';--not-comment11'  ;  ",
			after:  "';--not-comment11'",
		},
		{
			before: "/*a12*/\n/*b*/\n",
			after:  "",
		},
		{
			before: "a/*x13*/;/*y*/b;",
			after:  "a;b",
		},
		{
			before: "a14\n\t\tb\n c ;",
			after:  "a14 b c",
		},
		{
			before: "a /* c15 */ /* c2 */ /* c3 */ /* c4 */ /* c5 */ b /* c6 */ /* c7 */ c",
			after:  "a b c",
		},
		{
			before: "word16 /* comment */ \t \n   word17",
			after:  "word16 word17",
		},
		{
			before: "SELECT 18 $$",
			after:  "SELECT 18 $$",
		},
		{
			before: "DELIMITER $$\nSELECT 19$$\nDELIMITER ;\n",
			after:  "DELIMITER $$\nSELECT 19$$\nDELIMITER ;\n",
		},
		{
			before: "DELIMITER $$\nDELIMITER ;\n",
			after:  "DELIMITER $$\nDELIMITER ;\n",
		},
	}

	for _, tc := range cases {
		t.Logf("input: %q", tc.before)
		ts := TokenizeMySQL(tc.before)
		tsCopy := ts.Copy()
		stripped := ts.Strip()
		if !assert.Equal(t, tsCopy, ts, "ts unchanged") {
			dumpTokens(t, "original", tsCopy)
			dumpTokens(t, "after-strip", ts)
			t.FailNow()
		}
		if !assert.Equal(t, tc.after, stripped.String(), "strip") {
			dumpTokens(t, "original", ts)
			dumpTokens(t, "stripped", stripped)
			t.FailNow()
		}
		if tc.after != "" {
			strippedCopy := stripped.Copy()
			unstripped := stripped.Unstrip()
			if !assert.Equal(t, strippedCopy, stripped, "stripped is unchanged") {
				dumpTokens(t, "stripped", stripped)
				dumpTokens(t, "strippedCopy", strippedCopy)
				t.FailNow()
			}
			if !assert.Equal(t, tc.before, unstripped.String(), "unstrip") {
				dumpTokens(t, "original", ts)
				dumpTokens(t, "stripped", stripped)
				dumpTokens(t, "unstripped", unstripped)
				t.FailNow()
			}
		}
	}
}

var nlDelimiterRE = regexp.MustCompile("\n+DELIMITER ")

func TestCmdSplit(t *testing.T) {
	cases := []struct {
		name            string
		input           string
		stripped        []string
		notStripped     []string
		joinNotStripped string
		joinStripped    string
	}{
		{
			name:            "delimiter",
			input:           "DELIMITER ;\n",
			notStripped:     []string{"DELIMITER ;\n"},
			stripped:        []string{"DELIMITER ;\n"},
			joinNotStripped: "",
			joinStripped:    "",
		},
		{
			name:            "empty",
			input:           "",
			notStripped:     []string{},
			stripped:        []string{},
			joinNotStripped: "",
			joinStripped:    "",
		},
		{
			name:        "comment_only",
			input:       "-- stuff100\n",
			notStripped: []string{"-- stuff100\n"},
			stripped:    []string{},
		},
		{
			name:         "two-commands",
			input:        "select 101;\nselect 2;\n",
			notStripped:  []string{"select 101", "\nselect 2", "\n"},
			stripped:     []string{"select 101", "select 2"},
			joinStripped: "select 101;select 2;",
		},
		{
			name:            "extra_semicolons_preserved_or_not",
			input:           "SELECT 102;;SELECT 2;;;",
			notStripped:     []string{"SELECT 102", "SELECT 2"},
			stripped:        []string{"SELECT 102", "SELECT 2"},
			joinNotStripped: "SELECT 102;SELECT 2;",
			joinStripped:    "SELECT 102;SELECT 2;",
		},
		{
			name:            "semicolons_inside_literal_and_comment",
			input:           "SELECT ';';/*x103;*/SELECT 2;",
			notStripped:     []string{"SELECT ';'", "/*x103;*/SELECT 2"},
			stripped:        []string{"SELECT ';'", "SELECT 2"},
			joinNotStripped: "SELECT ';';/*x103;*/SELECT 2;",
			joinStripped:    "SELECT ';';SELECT 2;",
		},
		{
			name:            "whitespace_only_commands_preserved_unstripped",
			input:           " ;SELECT 104;  ;SELECT 2; ",
			notStripped:     []string{" ", "SELECT 104", "  ", "SELECT 2", " "},
			stripped:        []string{"SELECT 104", "SELECT 2"},
			joinNotStripped: " ;SELECT 104;  ;SELECT 2; ",
			joinStripped:    "SELECT 104;SELECT 2;",
		},
		{
			name:            "delimiter_text_inside_literal_not_split",
			input:           "DELIMITER $$\nSELECT '$$';$$\nDELIMITER ;\nSELECT 105;\n",
			notStripped:     []string{"DELIMITER $$\nSELECT '$$';$$\nDELIMITER ;\n", "SELECT 105", "\n"},
			stripped:        []string{"DELIMITER $$\nSELECT '$$';$$\nDELIMITER ;\n", "SELECT 105"},
			joinNotStripped: "DELIMITER $$\nSELECT '$$';$$\nDELIMITER ;\nSELECT 105;\n",
			joinStripped:    "DELIMITER $$\nSELECT '$$';$$\nDELIMITER ;\nSELECT 105;",
		},
		{
			name:            "leading_and_trailing_semicolon_runs",
			input:           ";;;SELECT 106;SELECT 2;;;",
			notStripped:     []string{"SELECT 106", "SELECT 2"},
			stripped:        []string{"SELECT 106", "SELECT 2"},
			joinNotStripped: "SELECT 106;SELECT 2;",
			joinStripped:    "SELECT 106;SELECT 2;",
		},
		{
			name:         "command_with_comment",
			input:        " /* foo107 */ bat \n baz  ; ",
			notStripped:  []string{" /* foo107 */ bat \n baz  ", " "},
			stripped:     []string{"bat baz"},
			joinStripped: "bat baz;",
		},
		{
			name:         "commands_with_comment",
			input:        " /* foo108 */ bar \n ;baz  ; ",
			notStripped:  []string{" /* foo108 */ bar \n ", "baz  ", " "},
			stripped:     []string{"bar", "baz"},
			joinStripped: "bar;baz;",
		},
		{
			name:            "two_delimited_commands",
			input:           "DELIMITER $$\nSELECT 109; $$\nSELECT 2$$\n",
			notStripped:     []string{"DELIMITER $$\nSELECT 109; $$\nDELIMITER ;\n", "DELIMITER $$\n\nSELECT 2$$\nDELIMITER ;\n", "\n"},
			stripped:        []string{"DELIMITER $$\nSELECT 109; $$\nDELIMITER ;\n", "DELIMITER $$\nSELECT 2$$\nDELIMITER ;\n"},
			joinNotStripped: "DELIMITER $$\nSELECT 109; $$\nSELECT 2$$\n\nDELIMITER ;\n\n",
			joinStripped:    "DELIMITER $$\nSELECT 109; $$SELECT 2$$\nDELIMITER ;\n",
		},
		{
			name:            "delimited_then_not_delimited1",
			input:           "DELIMITER //\nCREATE PROCEDURE p()\nBEGIN;\nSELECT 110;\nEND//\nDELIMITER ;\nSELECT 2;\n",
			stripped:        []string{"DELIMITER //\nCREATE PROCEDURE p() BEGIN; SELECT 110; END//\nDELIMITER ;\n", "SELECT 2"},
			notStripped:     []string{"DELIMITER //\nCREATE PROCEDURE p()\nBEGIN;\nSELECT 110;\nEND//\nDELIMITER ;\n", "SELECT 2", "\n"},
			joinNotStripped: "DELIMITER //\nCREATE PROCEDURE p()\nBEGIN;\nSELECT 110;\nEND//\nDELIMITER ;\nSELECT 2;\n",
			joinStripped:    "DELIMITER //\nCREATE PROCEDURE p() BEGIN; SELECT 110; END//\nDELIMITER ;\nSELECT 2;",
		},
		{
			name:         "delimited_then_not_delimited2",
			input:        "DELIMITER //\nCREATE PROCEDURE p()\nBEGIN;\nSELECT 111;\nEND//\nDELIMITER ;\nSELECT 2;\n",
			stripped:     []string{"DELIMITER //\nCREATE PROCEDURE p() BEGIN; SELECT 111; END//\nDELIMITER ;\n", "SELECT 2"},
			notStripped:  []string{"DELIMITER //\nCREATE PROCEDURE p()\nBEGIN;\nSELECT 111;\nEND//\nDELIMITER ;\n", "SELECT 2", "\n"},
			joinStripped: "DELIMITER //\nCREATE PROCEDURE p() BEGIN; SELECT 111; END//\nDELIMITER ;\nSELECT 2;",
		},
		{
			name:         "delimited_then_not_delimited3",
			input:        "DELIMITER //\nCREATE PROCEDURE p()\nBEGIN;\nSELECT 1;\nEND// /* comment */\nDELIMITER ;\nSELECT 2;\n",
			stripped:     []string{"DELIMITER //\nCREATE PROCEDURE p() BEGIN; SELECT 1; END//\nDELIMITER ;\n", "SELECT 2"},
			notStripped:  []string{"DELIMITER //\nCREATE PROCEDURE p()\nBEGIN;\nSELECT 1;\nEND// /* comment */\nDELIMITER ;\n", "SELECT 2", "\n"},
			joinStripped: "DELIMITER //\nCREATE PROCEDURE p() BEGIN; SELECT 1; END//\nDELIMITER ;\nSELECT 2;",
		},
		{
			name:         "delimited_trailing_comment_before_delimiter_cmd",
			input:        "DELIMITER //\nSELECT 1// -- c1\nDELIMITER ;\nSELECT 2;\n",
			stripped:     []string{"DELIMITER //\nSELECT 1//\nDELIMITER ;\n", "SELECT 2"},
			notStripped:  []string{"DELIMITER //\nSELECT 1// -- c1\nDELIMITER ;\n", "SELECT 2", "\n"},
			joinStripped: "DELIMITER //\nSELECT 1//\nDELIMITER ;\nSELECT 2;",
		},
		{
			name:         "delimited_comment_tail_is_wrapped",
			input:        "DELIMITER $$\nSELECT 1$$/* c */\nDELIMITER ;\nSELECT 2;\n",
			stripped:     []string{"DELIMITER $$\nSELECT 1$$\nDELIMITER ;\n", "SELECT 2"},
			notStripped:  []string{"DELIMITER $$\nSELECT 1$$/* c */\nDELIMITER ;\n", "SELECT 2", "\n"},
			joinStripped: "DELIMITER $$\nSELECT 1$$\nDELIMITER ;\nSELECT 2;",
		},
		{
			name:         "delimited_multitoken_comment_tail_is_wrapped",
			input:        "DELIMITER END-IF\nSELECT 1END-IF/* c */\nDELIMITER ;\nSELECT 2;\n",
			stripped:     []string{"DELIMITER END-IF\nSELECT 1END-IF\nDELIMITER ;\n", "SELECT 2"},
			notStripped:  []string{"DELIMITER END-IF\nSELECT 1END-IF/* c */\nDELIMITER ;\n", "SELECT 2", "\n"},
			joinStripped: "DELIMITER END-IF\nSELECT 1END-IF\nDELIMITER ;\nSELECT 2;",
		},
		{
			name:         "delimited_whitespace_tail_is_unwrapped",
			input:        "DELIMITER //\nSELECT 1// \nDELIMITER ;\nSELECT 2;\n",
			stripped:     []string{"DELIMITER //\nSELECT 1//\nDELIMITER ;\n", "SELECT 2"},
			notStripped:  []string{"DELIMITER //\nSELECT 1// \nDELIMITER ;\n", "SELECT 2", "\n"},
			joinStripped: "DELIMITER //\nSELECT 1//\nDELIMITER ;\nSELECT 2;",
		},
		{
			name:         "delimited_comment_tail_no_leading_space_wrapped",
			input:        "DELIMITER $$\nSELECT 1$$/*c*/\nDELIMITER ;\nSELECT 2;\n",
			stripped:     []string{"DELIMITER $$\nSELECT 1$$\nDELIMITER ;\n", "SELECT 2"},
			notStripped:  []string{"DELIMITER $$\nSELECT 1$$/*c*/\nDELIMITER ;\n", "SELECT 2", "\n"},
			joinStripped: "DELIMITER $$\nSELECT 1$$\nDELIMITER ;\nSELECT 2;",
		},
		{
			name:            "delimiter_command_without_newline_activated",
			input:           "DELIMITER $$ SELECT 1;",
			stripped:        []string{"DELIMITER $$ SELECT 1;\nDELIMITER ;\n"},
			notStripped:     []string{"DELIMITER $$ SELECT 1;\nDELIMITER ;\n"},
			joinStripped:    "DELIMITER $$ SELECT 1;\nDELIMITER ;\n",
			joinNotStripped: "DELIMITER $$ SELECT 1;\nDELIMITER ;\n",
		},
		{
			name:  "join_notices_delimiter_changes_ignoring_whitespace_only_segments",
			input: "DELIMITER $$\nSELECT 1$$\nDELIMITER //\nSELECT 2//\nDELIMITER ;\nSELECT 3;\n",
			stripped: []string{
				"DELIMITER $$\nSELECT 1$$\nDELIMITER ;\n",
				"DELIMITER //\nSELECT 2//\nDELIMITER ;\n",
				"SELECT 3",
			},
			notStripped: []string{
				"DELIMITER $$\nSELECT 1$$\nDELIMITER ;\n",
				"\nDELIMITER //\nSELECT 2//\nDELIMITER ;\n",
				"SELECT 3",
				"\n",
			},
			joinStripped: "DELIMITER $$\nSELECT 1$$\nDELIMITER //\nSELECT 2//\nDELIMITER ;\nSELECT 3;",
		},
		{
			name:  "delimiter_set_at_end",
			input: "DELIMITER $$\nA$$B$$C$$\nDELIMITER ;\n",
			stripped: []string{
				"DELIMITER $$\nA$$\nDELIMITER ;\n",
				"DELIMITER $$\nB$$\nDELIMITER ;\n",
				"DELIMITER $$\nC$$\nDELIMITER ;\n",
			},
			notStripped: []string{
				"DELIMITER $$\nA$$\nDELIMITER ;\n",
				"DELIMITER $$\nB$$\nDELIMITER ;\n",
				"DELIMITER $$\nC$$\nDELIMITER ;\n",
			},
			joinStripped: "DELIMITER $$\nA$$B$$C$$\nDELIMITER ;\n",
		},
		{
			name:        "delimiters_without_content",
			input:       "DELIMITER $$\nDELIMITER ;\n",
			stripped:    []string{"DELIMITER $$\nDELIMITER ;\n"},
			notStripped: []string{"DELIMITER $$\nDELIMITER ;\n"},
		},
		{
			name:        "just_delimiter_changes",
			input:       "DELIMITER //\nDELIMITER $$\nDELIMITER ;\n",
			stripped:    []string{"DELIMITER //\nDELIMITER $$\nDELIMITER ;\n"},
			notStripped: []string{"DELIMITER //\nDELIMITER $$\nDELIMITER ;\n"},
		},
		{
			name:            "just_delimited_whitespace",
			input:           "DELIMITER $$\n \n$$",
			stripped:        []string{"DELIMITER $$\n$$\nDELIMITER ;\n"},
			notStripped:     []string{"DELIMITER $$\n \n$$\nDELIMITER ;\n"},
			joinNotStripped: "DELIMITER $$\n \n$$\nDELIMITER ;\n",
			joinStripped:    "DELIMITER $$\n$$\nDELIMITER ;\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Log(tc.name)
			ts := TokenizeMySQL(tc.input)
			tsCopy := ts.Copy()
			t.Logf("input: %q", tc.input)
			dumpTokens(t, "raw", ts)
			notStripped := ts.CmdSplitUnstripped()
			if !assert.Equal(t, tsCopy, ts, "ts changed 1") {
				dumpTokens(t, "raw#2", ts)
				t.FailNow()
			}
			notStrippedCopy := notStripped.Copy()
			dumpTokens(t, "notStripped", notStripped...)
			stripped := ts.CmdSplit()
			if !assert.Equal(t, tsCopy, ts, "ts changed 2") {
				dumpTokens(t, "raw#3", ts)
				t.FailNow()
			}
			dumpTokens(t, "stripped", stripped...)
			strippedCopy := stripped.Copy()
			if !assert.Equal(t, notStrippedCopy, notStripped, "notStripped changed") {
				dumpTokens(t, "notStripped#2", notStripped...)
				t.FailNow()
			}
			require.Equal(t, tc.notStripped, notStripped.Strings(), "notStripped commands")
			require.Equal(t, tc.stripped, stripped.Strings(), "stripped commands")
			if tc.joinNotStripped == "" && len(tc.notStripped) != 0 {
				tc.joinNotStripped = tc.input
				t.Logf("joinNotStripped not set, using %q", tc.input)
			}
			if tc.joinStripped == "" && len(tc.stripped) != 0 {
				tc.joinStripped = ts.Strip().String()
				t.Logf("joinStripped not set, using %q", tc.joinStripped)
			}
			notStrippedJoin := notStripped.Join()
			if !assert.Equal(t, notStrippedCopy, notStripped, "notStripped changed") {
				dumpTokens(t, "notStripped after", notStripped...)
				t.FailNow()
			}
			noSplit(t, notStrippedJoin, "not stripped join")
			if !assert.Equal(t, simplifyNLs(tc.joinNotStripped), simplifyNLs(notStrippedJoin.String()), "notStripped join") {
				dumpTokens(t, "notStripped.Join", notStrippedJoin)
				t.FailNow()
			}
			strippedJoin := stripped.Join()
			if !assert.Equal(t, strippedCopy, stripped, "stripped changed") {
				dumpTokens(t, "stripped after", stripped...)
				t.FailNow()
			}
			noSplit(t, strippedJoin, "stripped join")
			if !assert.Equal(t, tc.joinStripped, strippedJoin.String(), "stripped join") {
				dumpTokens(t, "stripped.Join", strippedJoin)
			}
		})
	}
}

func noSplit(t *testing.T, ts Tokens, what string) {
	for _, token := range ts {
		if !assert.Emptyf(t, token.Split, "%s: split tag inappropriate", what) {
			dumpTokens(t, what, ts)
			t.FailNow()
		}
	}
}

func simplifyNLs(s string) string {
	return nlDelimiterRE.ReplaceAllLiteralString(s, "\nDELIMITER ")
}

func dumpTokens(t *testing.T, prefix string, tokens ...Tokens) {
	tokensString := func(what string, tkns ...Token) string {
		if len(tkns) == 0 {
			return ""
		}
		s := make([]string, len(tkns))
		for i, tok := range tkns {
			s[i] = fmt.Sprintf("%q", tok.Text)
		}
		return fmt.Sprintf(" (%s: %d: %s)", what, len(tkns), strings.Join(s, ", "))
	}
	tokenPointerString := func(what string, tkn *Token) string {
		if tkn == nil {
			return ""
		}
		return fmt.Sprintf(" (%s: %s: %q)", what, tkn.Type, tkn.Text)
	}
	debugString := func(debug string) string {
		if debug != "" {
			return fmt.Sprintf(" (debug:%q)", debug)
		}
		return ""
	}
	for i, s := range tokens {
		t.Logf(" %s-%d: %q", prefix, i, s.String())
		for j, token := range s {
			t.Logf("  %s-%d-%d: %s %q%s%s%s", prefix, i, j, token.Type, token.Text,
				tokenPointerString("split", token.Split),
				tokensString("strip", token.Strip...),
				debugString(token.Debug()))
		}
	}
}

// New tests exercising CmdSplitUnstripped and raw Strings output
func TestCmdSplitUnstrippedStrings(t *testing.T) {
	cases := []struct {
		raw      []string // expected Strings() from CmdSplitUnstripped (unmodified slices)
		stripped []string // expected Strings() from CmdSplit (after Strip)
	}{
		{
			raw:      []string{},
			stripped: []string{},
		},
		{
			raw:      []string{"-- cmt\n"},
			stripped: []string{},
		},
		{
			raw:      []string{"SELECT 1", " SELECT 2"},
			stripped: []string{"SELECT 1", "SELECT 2"},
		},
		{
			raw:      []string{" /* foo */ bar \n baz  ", " "}, // trailing whitespace segment retained
			stripped: []string{"bar baz"},
		},
		{
			raw:      []string{" /* foo */ bar \n ", "baz  ", " "},
			stripped: []string{"bar", "baz"},
		},
		{
			raw:      []string{"-- cmt\nSELECT 3 ", "  SELECT 4  "},
			stripped: []string{"SELECT 3", "SELECT 4"},
		},
		{
			raw:      []string{"SELECT ';' ", " /*x;*/ SELECT 2"},
			stripped: []string{"SELECT ';'", "SELECT 2"},
		},
		{
			raw:      []string{"/* ; */", "SELECT 1"},
			stripped: []string{"SELECT 1"},
		},
		{
			raw:      []string{" ", "SELECT 1", "  ", "SELECT 2", " "},
			stripped: []string{"SELECT 1", "SELECT 2"},
		},
		{
			raw:      []string{"SELECT 'a;b;c'", "/* ; ; */", "SELECT 2"},
			stripped: []string{"SELECT 'a;b;c'", "SELECT 2"},
		},
	}
	for _, tc := range cases {
		input := strings.Join(tc.raw, ";")
		ts := TokenizeMySQL(input)
		// Baseline: original String() matches original input text
		require.Equal(t, input, ts.String(), input)
		rawList := ts.CmdSplitUnstripped().Strings()
		require.Equalf(t, tc.raw, rawList, "raw mismatch for %q", input)
		stripList := ts.CmdSplit().Strings()
		require.Equalf(t, tc.stripped, stripList, "stripped mismatch for %q", input)
	}
}
