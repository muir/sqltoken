package sqltoken

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var commonCases = []Tokens{
	{},
	{
		{Type: Word, Text: "c01"},
	},
	{
		{Type: Word, Text: "c02"},
		{Type: Delimiter, Text: ";"},
		{Type: Word, Text: "morestuff"},
	},
	{
		{Type: Word, Text: "c03"},
		{Type: Comment, Text: "--cmt;\n"},
		{Type: Word, Text: "stuff2"},
	},
	{
		{Type: Word, Text: "c04"},
		{Type: Punctuation, Text: "-"},
		{Type: Word, Text: "an"},
		{Type: Punctuation, Text: "-"},
		{Type: Word, Text: "dom"},
	},
	{
		{Type: Word, Text: "c05_singles"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "''"},
		{Type: Whitespace, Text: " \t"},
		{Type: Literal, Text: "'\\''"},
		{Type: Semicolon, Text: ";"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "';\\''"},
	},
	{
		{Type: Word, Text: "c06_doubles"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: `""`},
		{Type: Whitespace, Text: " \t"},
		{Type: Literal, Text: `"\""`},
		{Type: Semicolon, Text: ";"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: `";\""`},
	},
	{
		{Type: Word, Text: "c07_singles"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "-"},
		{Type: Literal, Text: "''"},
		{Type: Whitespace, Text: " \t"},
		{Type: Punctuation, Text: "-"},
		{Type: Literal, Text: "'\\''"},
		{Type: Semicolon, Text: ";"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "-"},
		{Type: Literal, Text: "';\\''"},
	},
	{
		{Type: Word, Text: "c08_doubles"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "-"},
		{Type: Literal, Text: `""`},
		{Type: Whitespace, Text: " \t"},
		{Type: Punctuation, Text: "-"},
		{Type: Literal, Text: `"\""`},
		{Type: Semicolon, Text: ";"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "-"},
		{Type: Literal, Text: `";\""`},
	},
	{
		{Type: Word, Text: "c09"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "r"},
		{Type: Punctuation, Text: "-"},
		{Type: Word, Text: "an"},
		{Type: Punctuation, Text: "-"},
		{Type: Word, Text: "dom"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: `";;"`},
		{Type: Semicolon, Text: ";"},
		{Type: Literal, Text: "';'"},
		{Type: Punctuation, Text: "-"},
		{Type: Literal, Text: `";"`},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "-"},
		{Type: Literal, Text: "';'"},
		{Type: Punctuation, Text: "-"},
	},
	{
		{Type: Word, Text: "c10"},
		{Type: Punctuation, Text: "-//"},
	},
	{
		{Type: Word, Text: "c11"},
		{Type: Punctuation, Text: "-//-/-"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "c12"},
		{Type: Punctuation, Text: "/"},
		{Type: Literal, Text: `";"`},
		{Type: Whitespace, Text: "\r\n"},
		{Type: Literal, Text: `";"`},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "-/"},
		{Type: Literal, Text: `";"`},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "c13"},
		{Type: Punctuation, Text: "/"},
		{Type: Literal, Text: "';'"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "';'"},
		{Type: Whitespace, Text: " "},
		{Type: Comment, Text: "/*;*/"},
		{Type: Punctuation, Text: "-/"},
		{Type: Literal, Text: "';'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "c14"},
		{Type: Punctuation, Text: "-"},
		{Type: Comment, Text: "/*;*/"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "-/"},
		{Type: Comment, Text: "/*\n\t;*/"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "c15"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: ".5"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "c16"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: ".5"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "0.5"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "30.5"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "40"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "40.13"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "40.15e8"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "40e8"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: ".4e8"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: ".4e20"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "c17"},
		{Type: Whitespace, Text: " "},
		{Type: Comment, Text: "/* foo \n */"},
	},
	{
		{Type: Word, Text: "c18"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "'unterminated "},
	},
	{
		{Type: Word, Text: "c19"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: `"unterminated `},
	},
	{
		{Type: Word, Text: "c20"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: `'unterminated \`},
	},
	{
		{Type: Word, Text: "c21"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: `"unterminated \`},
	},
	{
		{Type: Word, Text: "c22"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: ".@"},
	},
	{
		{Type: Word, Text: "c23"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: ".@"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "c24"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "7"},
		{Type: Word, Text: "ee"},
	},
	{
		{Type: Word, Text: "c25"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "7"},
		{Type: Word, Text: "eg"},
	},
	{
		{Type: Word, Text: "c26"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "7"},
		{Type: Word, Text: "ee"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "c27"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "7"},
		{Type: Word, Text: "eg"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "c28"},
		{Type: Whitespace, Text: " "},
		{Type: Comment, Text: "/* foo "},
	},
	{
		{Type: Word, Text: "c29"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "7e8"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "c30"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "7e8"},
	},
	{
		{Type: Word, Text: "c31"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "7.0"},
		{Type: Word, Text: "e"},
	},
	{
		{Type: Word, Text: "c32"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "7.0"},
		{Type: Word, Text: "e"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "c33"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "eèҾ"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "ҾeèҾ"},
	},
	{
		{Type: Word, Text: "c34"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "⁖"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "+⁖"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "+⁖*"},
	},
	{
		{Type: Word, Text: "c35"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "๒"},
	},
	{
		{Type: Word, Text: "c36"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "๒"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "c37"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "๒⎖๒"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "c38"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "⎖๒"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "c39"},
		{Type: Whitespace, Text: " "},
		{Type: Comment, Text: "-- comment w/o end"},
	},
	{
		{Type: Word, Text: "c40"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: ".๒"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "c40"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "abnormal"},
		{Type: Whitespace, Text: " "}, // this is a unicode space character
		{Type: Word, Text: "space"},
	},
	{
		{Type: Word, Text: "c41"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "abnormal"},
		{Type: Whitespace, Text: "  "}, // this is a unicode space character
		{Type: Word, Text: "space"},
	},
	{
		{Type: Word, Text: "c42"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "abnormal"},
		{Type: Whitespace, Text: "  "}, // this is a unicode space character
		{Type: Word, Text: "space"},
	},
	{
		{Type: Word, Text: "c43"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "."},
	},
	{
		{Type: Word, Text: "c44"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "๒๒"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "c45"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "3e๒๒๒"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "c46"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "3.7"},
	},
	{
		{Type: Word, Text: "c47"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "3.7e19"},
	},
	{
		{Type: Word, Text: "c48"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "3.7e2"},
	},
	{
		{Type: Word, Text: "c49"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "😀"}, // I'm not sure I agree with the classification
	},
	{
		{Type: Word, Text: "c50"},
		{Type: Whitespace, Text: " \x00"},
	},
	{
		{Type: Word, Text: "c51"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "x"},
		{Type: Whitespace, Text: "\x00"},
	},
	{
		{Type: Comment, Text: "-- c52\n"},
	},
	{
		{Type: Word, Text: "c53"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "z"},
		{Type: Literal, Text: "'not a prefixed literal'"},
	},
	{
		{Type: Word, Text: "c54"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "z"},
	},
}

var mySQLCases = []Tokens{
	{
		{Type: Word, Text: "m01"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "-"},
		{Type: Comment, Text: "# /# #;\n"},
		{Type: Whitespace, Text: "\t"},
		{Type: Word, Text: "foo"},
	},
	{
		{Type: Word, Text: "m02"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "'#;'"},
		{Type: Punctuation, Text: ","},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: `"#;"`},
		{Type: Punctuation, Text: ","},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "-"},
		{Type: Comment, Text: "# /# #;\n"},
		{Type: Whitespace, Text: "\t"},
		{Type: Word, Text: "foo"},
	},
	{
		{Type: Word, Text: "m03"},
		{Type: Whitespace, Text: " "},
		{Type: QuestionMark, Text: "?"},
		{Type: QuestionMark, Text: "?"},
	},
	{
		{Type: Word, Text: "m04"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "$"},
		{Type: Number, Text: "5"},
	},
	{
		{Type: Word, Text: "m05"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "U"},
		{Type: Punctuation, Text: "&"},
		{Type: Literal, Text: `'d\0061t\+000061'`},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "m06"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "0x1f"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "x'1f'"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "X'1f'"},
	},
	{
		{Type: Word, Text: "m07"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "0b01"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "b'010'"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "B'110'"},
	},
	{
		{Type: Word, Text: "m08"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "0b01"},
	},
	{
		{Type: Word, Text: "m09"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "0x01"},
	},
	{
		{Type: Word, Text: "m10"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "x'1f"},
		{Type: Punctuation, Text: "&"},
	},
	{
		{Type: Word, Text: "m10"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "b'1"},
		{Type: Number, Text: "7"},
	},
	{
		{Type: Word, Text: "m11"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "$$"},
		{Type: Word, Text: "footext"},
		{Type: Punctuation, Text: "$$"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "m12"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "b'10"},
	},
	{
		{Type: Word, Text: "m13"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "x'1f"},
	},
	{
		{Type: Word, Text: "m14"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "n'national charset'"},
	},
	{
		{Type: Word, Text: "m14"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "_utf8'redundent'"},
	},
	{
		{Type: Word, Text: "m15"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "=@:$"},
	},
	{
		{Type: Word, Text: "m16"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "n'martha''s family'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "m17"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "q"},
		{Type: Literal, Text: "'!martha''s family!'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "m18"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "nq"},
		{Type: Literal, Text: "'!martha''s family!'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "m19"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "@"},
		{Type: Word, Text: "foo"},
		{Type: Punctuation, Text: "$"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "m20"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "f"},
		{Type: Comment, Text: "#o@o$ "},
	},
	{
		{Type: Word, Text: "m21"},
		{Type: Whitespace, Text: " "},
		{Type: Comment, Text: "#foo "},
	},
	{
		{Type: Word, Text: "m22"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "foo"},
		{Type: Punctuation, Text: "$"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "m23"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "foo"},
		{Type: Punctuation, Text: "@"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "m24"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "foo"},
		{Type: Comment, Text: "# "},
	},
	{
		{Type: Word, Text: "m25"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "_foo"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "m26"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: ":"},
		{Type: Word, Text: "名前"},
		{Type: Punctuation, Text: ")"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER    78\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "12"},
		{Type: Delimiter, Text: "78"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER /*foo*/\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "7"},
		{Type: Delimiter, Text: "/*foo*/"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER 'foo''bar'\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "'bar''foo'"},
		{Type: Delimiter, Text: "foo'bar"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER 'foo' 'bar'\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "'bar''foo'"},
		{Type: Whitespace, Text: " "},
		{Type: Delimiter, Text: "foobar"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER ---\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "8"},
		{Type: Comment, Text: "-- comment ---\n"},
		{Type: Whitespace, Text: " "},
		{Type: Delimiter, Text: "---"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER x y z\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "8"},
		{Type: Delimiter, Text: "x"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER ---\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "9"},
		{Type: Comment, Text: "/* --- comment */"},
		{Type: Whitespace, Text: "\n "},
		{Type: Delimiter, Text: "---"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER o'foo$\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "'bar'"},
		{Type: Delimiter, Text: "o'foo$"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "'baz'"},
		{Type: Delimiter, Text: "o'foo$"},
		{Type: Whitespace, Text: "\n"},
		{Type: DelimiterStatement, Text: "Delimiter ;\n"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER foo'\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "7"},
		{Type: Whitespace, Text: " "},
		{Type: Delimiter, Text: "foo'"},
		{Type: Whitespace, Text: "\n"},
		{Type: DelimiterStatement, Text: "DELIMITER ;\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "2"},
		{Type: Delimiter, Text: ";"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER $$\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "'$$'"},
		{Type: Punctuation, Text: ";"},
		{Type: Delimiter, Text: "$$"},
		{Type: Whitespace, Text: "\n"},
		{Type: DelimiterStatement, Text: "DELIMITER ;\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "2"},
		{Type: Delimiter, Text: ";"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: DelimiterStatement, Text: "delimiter $$\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "19"},
		{Type: Delimiter, Text: "$$"},
		{Type: Word, Text: "delimiter"}, // not recognized because not at start of line
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: ";"},
		{Type: Whitespace, Text: "\n"},
	},
	// delimiter at very end of input — off-by-one: currently not recognized as Delimiter
	{
		{Type: DelimiterStatement, Text: "DELIMITER $$\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "1"},
		{Type: Punctuation, Text: "$$"},
	},
	// delimiter-only content right after DELIMITER statement — off-by-one
	{
		{Type: DelimiterStatement, Text: "DELIMITER $$\n"},
		{Type: Punctuation, Text: "$$"},
	},
	// DELIMITER after \r not recognized (only \n counts as line-start)
	{
		{Type: Whitespace, Text: "\r"},
		{Type: Word, Text: "DELIMITER"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "$$"},
		{Type: Whitespace, Text: "\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "1"},
		{Type: Punctuation, Text: "$$"},
	},
	// DELIMITER with no value — just EOF
	{
		{Type: Word, Text: "DELIMITER"},
	},
	// DELIMITER followed only by spaces then EOF
	{
		{Type: Word, Text: "DELIMITER"},
		{Type: Whitespace, Text: "   "},
	},
	// DELIMITER immediately followed by \n (no space) — not recognized
	{
		{Type: Word, Text: "DELIMITER"},
		{Type: Whitespace, Text: "\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "1"},
		{Type: Delimiter, Text: ";"},
	},
	// single-character delimiter
	{
		{Type: DelimiterStatement, Text: "DELIMITER /\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "1"},
		{Type: Delimiter, Text: "/"},
		{Type: Whitespace, Text: "\n"},
		{Type: DelimiterStatement, Text: "DELIMITER ;\n"},
	},
	// mixed-case delimiter command
	{
		{Type: DelimiterStatement, Text: "Delimiter $$\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "1"},
		{Type: Delimiter, Text: "$$"},
		{Type: Whitespace, Text: "\n"},
		{Type: DelimiterStatement, Text: "DeLiMiTeR ;\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "2"},
		{Type: Delimiter, Text: ";"},
		{Type: Whitespace, Text: "\n"},
	},
	// multiple semicolons inside custom delimiter mode become Punctuation
	{
		{Type: DelimiterStatement, Text: "DELIMITER $$\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "1"},
		{Type: Punctuation, Text: ";"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "2"},
		{Type: Punctuation, Text: ";"},
		{Type: Delimiter, Text: "$$"},
		{Type: Whitespace, Text: "\n"},
		{Type: DelimiterStatement, Text: "DELIMITER ;\n"},
	},
	// single-quoted delimiter value
	{
		{Type: DelimiterStatement, Text: "DELIMITER '$$'\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "1"},
		{Type: Delimiter, Text: "$$"},
		{Type: Whitespace, Text: "\n"},
		{Type: DelimiterStatement, Text: "DELIMITER ;\n"},
	},
	// double-quoted delimiter value
	{
		{Type: DelimiterStatement, Text: "DELIMITER \"$$\"\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "1"},
		{Type: Delimiter, Text: "$$"},
		{Type: Whitespace, Text: "\n"},
		{Type: DelimiterStatement, Text: "DELIMITER ;\n"},
	},
	// trailing junk after delimiter value is part of statement
	{
		{Type: DelimiterStatement, Text: "DELIMITER $$ extra\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "1"},
		{Type: Delimiter, Text: "$$"},
		{Type: Whitespace, Text: "\n"},
		{Type: DelimiterStatement, Text: "DELIMITER ;\n"},
	},
	// empty commands between consecutive delimiters
	{
		{Type: DelimiterStatement, Text: "DELIMITER $$\n"},
		{Type: Delimiter, Text: "$$"},
		{Type: Delimiter, Text: "$$"},
		{Type: Whitespace, Text: "\n"},
		{Type: DelimiterStatement, Text: "DELIMITER ;\n"},
	},
	// DELIMITER word mid-line treated as plain word
	{
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "DELIMITER"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "$$"},
		{Type: Whitespace, Text: "\n"},
	},
}

var postgreSQLCases = []Tokens{
	{
		{Type: Word, Text: "p01"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "#"},
		{Type: Word, Text: "foo"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: Word, Text: "p02"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "?"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: Word, Text: "p03"},
		{Type: Whitespace, Text: " "},
		{Type: DollarNumber, Text: "$17"},
		{Type: DollarNumber, Text: "$8"},
	},
	{
		{Type: Word, Text: "p04"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: `U&'d\0061t\+000061'`},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p05"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "0"},
		{Type: Word, Text: "x1f"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "x"},
		{Type: Literal, Text: "'1f'"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "X"},
		{Type: Literal, Text: "'1f'"},
	},
	{
		{Type: Word, Text: "p06"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "0"},
		{Type: Word, Text: "b01"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "b"},
		{Type: Literal, Text: "'010'"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "B"},
		{Type: Literal, Text: "'110'"},
	},
	{
		{Type: Word, Text: "p07"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "$$footext$$"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p08"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "$$foo!text$$"},
	},
	{
		{Type: Word, Text: "p09"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "$q$foo$$text$q$"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p10"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "$q$foo$$text$q$"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p11"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "$$"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p12"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "$$"},
	},
	{
		{Type: Word, Text: "p13"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "$"},
		{Type: Word, Text: "q"},
		{Type: Punctuation, Text: "$"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p14"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "$ҾeèҾ$ $ DLa 32498 $ҾeèҾ$"},
		{Type: Punctuation, Text: "$"},
	},
	{
		{Type: Word, Text: "p15"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "$ҾeèҾ$ $ DLa 32498 $ҾeèҾ$"},
	},
	{
		{Type: Word, Text: "p16"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "$"},
		{Type: Word, Text: "foo"},
		{Type: Punctuation, Text: "-$"},
		{Type: Word, Text: "bar"},
		{Type: Punctuation, Text: "$"},
		{Type: Word, Text: "foo"},
		{Type: Punctuation, Text: "-$"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p16"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "n"},
		{Type: Literal, Text: "'mysql only'"},
	},
	{
		{Type: Word, Text: "p16"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "_utf8"},
		{Type: Literal, Text: "'mysql only'"},
	},
	{
		{Type: Word, Text: "p17"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "=@:?"},
	},
	{
		{Type: Word, Text: "p18"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "n"},
		{Type: Literal, Text: "'martha''s family'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p19"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "q"},
		{Type: Literal, Text: "'!martha''s family!'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p20"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "nq"},
		{Type: Literal, Text: "'!martha''s family!'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p21"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "3.7E9"},
		{Type: Word, Text: "f"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p22"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "3.7"},
		{Type: Word, Text: "D"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p23"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "3"},
		{Type: Word, Text: "d"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p24"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "3.7E11"},
		{Type: Word, Text: "F"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p25"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "@"},
		{Type: Word, Text: "foo"},
		{Type: Punctuation, Text: "$"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p26"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "f"},
		{Type: Punctuation, Text: "#"},
		{Type: Word, Text: "o"},
		{Type: Punctuation, Text: "@"},
		{Type: Word, Text: "o"},
		{Type: Punctuation, Text: "$"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p27"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "#"},
		{Type: Word, Text: "foo"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p28"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "foo"},
		{Type: Punctuation, Text: "$"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p29"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "foo"},
		{Type: Punctuation, Text: "@"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p30"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "foo"},
		{Type: Punctuation, Text: "#"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p31"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "_foo"},
		{Type: Whitespace, Text: " "},
	},
}

var oracleCases = []Tokens{
	{
		{Type: Word, Text: "o1"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "'martha''s family'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o2"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "n'martha''s family'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o3"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "N'martha''s family'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o4"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "q'!martha's family!'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o5"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "q'<martha's >< family>'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o6"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "Nq'(martha's )( family)'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o7"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "Q'{martha's  family}'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o8"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "nq'[martha's  family]'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o9"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "nq"},
		{Type: Literal, Text: "' martha''s  family '"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o10"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "Q"},
		{Type: Literal, Text: "' martha''s  family '"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o11"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "3.7E9f"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o12"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "3.7D"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o13"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "3d"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o14"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "3.7E11F"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o15"},
		{Type: Whitespace, Text: " "},
		{Type: ColonWord, Text: ":foo"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o16"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: ":"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o17"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: ":"},
		{Type: Number, Text: "3"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o17"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "@"},
		{Type: Word, Text: "foo"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o16"},
		{Type: Whitespace, Text: " "},
		{Type: ColonWord, Text: ":f"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o17"},
		{Type: Whitespace, Text: " "},
		{Type: ColonWord, Text: ":f"},
	},
	{
		{Type: Word, Text: "o18"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "::"},
		{Type: Word, Text: "foo"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o19"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: ":::"},
		{Type: Word, Text: "foo"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "o20"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "::::"},
		{Type: Word, Text: "foo"},
		{Type: Whitespace, Text: " "},
	},
}

var sqlServerCases = []Tokens{
	{
		{Type: Word, Text: "s01"},
		{Type: Whitespace, Text: " "},
		{Type: Identifier, Text: "@foo$"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "s02"},
		{Type: Whitespace, Text: " "},
		{Type: Identifier, Text: "f#o@o$"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "s03"},
		{Type: Whitespace, Text: " "},
		{Type: Identifier, Text: "#foo"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "s04"},
		{Type: Whitespace, Text: " "},
		{Type: Identifier, Text: "foo$"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "s05"},
		{Type: Whitespace, Text: " "},
		{Type: Identifier, Text: "foo@"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "s06"},
		{Type: Whitespace, Text: " "},
		{Type: Identifier, Text: "foo#"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "s07"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "_foo"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "s08"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "'martha''s family'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "s09"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "n'martha''s family'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "s10"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "N'martha''s family'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "s11"},
		{Type: Whitespace, Text: " "},
		{Type: AtWord, Text: "@foo"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "s12"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "@"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "s13"},
		{Type: Whitespace, Text: " "},
		{Type: Identifier, Text: "@8"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "s14"},
		{Type: Whitespace, Text: " "},
		{Type: Identifier, Text: "@88"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "s15"},
		{Type: Whitespace, Text: " "},
		{Type: Identifier, Text: "@foo$b"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "s16"},
		{Type: Whitespace, Text: " "},
		{Type: Identifier, Text: "@foo$b"},
	},
	{
		{Type: Word, Text: "s17"},
		{Type: Whitespace, Text: " "},
		{Type: Identifier, Text: "@88"},
	},
	{
		{Type: Word, Text: "s18"},
		{Type: Whitespace, Text: " "},
		{Type: Identifier, Text: "@8"},
	},
	{
		{Type: Word, Text: "s19"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "@"},
	},
}

// SQLServer w/o AtWord
var oddball1Cases = []Tokens{
	{
		{Type: Word, Text: "od1"},
		{Type: Whitespace, Text: " "},
		{Type: Identifier, Text: "@foo$"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "od2"},
		{Type: Whitespace, Text: " "},
		{Type: Identifier, Text: "@foo"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "od3"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "@"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "od4"},
		{Type: Whitespace, Text: " "},
		{Type: Identifier, Text: "@8"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "od5"},
		{Type: Whitespace, Text: " "},
		{Type: Identifier, Text: "@88"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "od6"},
		{Type: Whitespace, Text: " "},
		{Type: Identifier, Text: "@88"},
	},
	{
		{Type: Word, Text: "od7"},
		{Type: Whitespace, Text: " "},
		{Type: Identifier, Text: "@8"},
	},
	{
		{Type: Word, Text: "od8"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "@"},
	},
}

// SQLx
var oddball2Cases = []Tokens{
	{
		{Type: Word, Text: "sqlx1"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "INSERT"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "INTO"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "foo"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "("},
		{Type: Word, Text: "a"},
		{Type: Punctuation, Text: ","},
		{Type: Word, Text: "b"},
		{Type: Punctuation, Text: ","},
		{Type: Word, Text: "c"},
		{Type: Punctuation, Text: ","},
		{Type: Word, Text: "d"},
		{Type: Punctuation, Text: ")"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "VALUES"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "("},
		{Type: ColonWord, Text: ":あ"},
		{Type: Punctuation, Text: ","},
		{Type: Whitespace, Text: " "},
		{Type: ColonWord, Text: ":b"},
		{Type: Punctuation, Text: ","},
		{Type: Whitespace, Text: " "},
		{Type: ColonWord, Text: ":キコ"},
		{Type: Punctuation, Text: ","},
		{Type: Whitespace, Text: " "},
		{Type: ColonWord, Text: ":名前"},
		{Type: Punctuation, Text: ")"},
	},
}

// SQLx
var separatePunctuationCases = []Tokens{
	{
		{Type: Word, Text: "INSERT"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "INTO"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "foo"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "("},
		{Type: Word, Text: "a"},
		{Type: Punctuation, Text: ","},
		{Type: Word, Text: "b"},
		{Type: Punctuation, Text: ","},
		{Type: Word, Text: "c"},
		{Type: Punctuation, Text: ","},
		{Type: Word, Text: "d"},
		{Type: Punctuation, Text: ")"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "VALUES"},
		{Type: Punctuation, Text: "("},
		{Type: Word, Text: "NOW"},
		{Type: Punctuation, Text: "("},
		{Type: Punctuation, Text: ")"},
		{Type: Punctuation, Text: ","},
		{Type: Whitespace, Text: " "},
		{Type: ColonWord, Text: ":b"},
		{Type: Punctuation, Text: ")"},
	},
}

func doTests(t *testing.T, config Config, cases ...[]Tokens) {
	for _, tcl := range cases {
		for i, tc := range tcl {
			tc := tc
			desc := "null"
			if len(tc) > 0 {
				desc = tc[0].Text
			}
			t.Run(fmt.Sprintf("%03d_%s", i+1, desc), func(t *testing.T) {
				text := tc.String()
				t.Log("---------------------------------------")
				t.Log(text)
				t.Log("-----------------")
				got := Tokenize(text, config)
				if !assert.Equal(t, text, got.String(), tc.String()) || !assert.Equal(t, tc, got, tc.String()) {
					dumpTokens(t, "want", tc)
					dumpTokens(t, "got", got)
				}
			})
		}
	}
}

func TestMySQLTokenizing(t *testing.T) {
	doTests(t, MySQLConfig(), commonCases, mySQLCases)
}

func TestPostgresSQLTokenizing(t *testing.T) {
	doTests(t, PostgreSQLConfig(), commonCases, postgreSQLCases)
}

func TestOracleTokenizing(t *testing.T) {
	doTests(t, OracleConfig(), commonCases, oracleCases)
}

func TestSQLServerTokenizing(t *testing.T) {
	doTests(t, SQLServerConfig(), commonCases, sqlServerCases)
}

func TestOddbal1Tokenizing(t *testing.T) {
	c := SQLServerConfig()
	c.NoticeAtWord = false
	doTests(t, c, commonCases, oddball1Cases)
}

func TestOddbal2Tokenizing(t *testing.T) {
	c := MySQLConfig()
	c.NoticeColonWord = true
	c.ColonWordIncludesUnicode = true
	doTests(t, c, commonCases, oddball2Cases)
}

func TestSeparatePunctuationTokenizing(t *testing.T) {
	c := MySQLConfig()
	c.NoticeColonWord = true
	c.NoticeQuestionMark = true
	c.SeparatePunctuation = true
	doTests(t, c, separatePunctuationCases)
}

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
			before: "-- stuff\n",
			after:  "",
		},
		{
			before: "/* foo */ bar \n baz  ; ",
			after:  "bar baz",
		},
		{
			before: " /* foo */ bar \n baz  ; ",
			after:  "bar baz",
		},
		{
			before: "\t\talpha  \n\n beta\t ;  ",
			after:  "alpha beta",
		},
		{
			before: "word/*c1*/\t/*c2*/word2 ; ",
			after:  "word word2",
		},
		{
			before: "  -- c1\n  word ; -- c2\n",
			after:  "word",
		},
		{
			before: "x ; ; ;  ",
			after:  "x",
		},
		{
			before: "a;b;c;",
			after:  "a;b;c",
		},
		{
			before: ";\n\t ; /* c */ ; -- tail",
			after:  "",
		},
		{
			before: "  ';--not-comment'  ;  ",
			after:  "';--not-comment'",
		},
		{
			before: "/*a*/\n/*b*/\n",
			after:  "",
		},
		{
			before: "a/*x*/;/*y*/b;",
			after:  "a;b",
		},
		{
			before: "a\n\t\tb\n c ;",
			after:  "a b c",
		},
		{
			before: "a /* c1 */ /* c2 */ /* c3 */ /* c4 */ /* c5 */ b /* c6 */ /* c7 */ c",
			after:  "a b c",
		},
		{
			before: "word1 /* comment */ \t \n   word2",
			after:  "word1 word2",
		},
		{
			before: "SELECT 1 $$",
			after:  "SELECT 1 $$",
		},
		{
			before: "SELECT 1 $$",
			after:  "SELECT 1 $$",
		},
		{
			before: "DELIMITER $$\nSELECT 1$$\nDELIMITER ;\n",
			after:  "DELIMITER $$\nSELECT 1$$\nDELIMITER ;\n",
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
			input:       "-- stuff\n",
			notStripped: []string{"-- stuff\n"},
			stripped:    []string{},
		},
		{
			name:         "two-commands",
			input:        "select 1;\nselect 2;\n",
			notStripped:  []string{"select 1", "\nselect 2", "\n"},
			stripped:     []string{"select 1", "select 2"},
			joinStripped: "select 1;select 2;",
		},
		{
			name:            "extra_semicolons_preserved_or_not",
			input:           "SELECT 1;;SELECT 2;;;",
			notStripped:     []string{"SELECT 1", "SELECT 2"},
			stripped:        []string{"SELECT 1", "SELECT 2"},
			joinNotStripped: "SELECT 1;SELECT 2;",
			joinStripped:    "SELECT 1;SELECT 2;",
		},
		{
			name:            "semicolons_inside_literal_and_comment",
			input:           "SELECT ';';/*x;*/SELECT 2;",
			notStripped:     []string{"SELECT ';'", "/*x;*/SELECT 2"},
			stripped:        []string{"SELECT ';'", "SELECT 2"},
			joinNotStripped: "SELECT ';';/*x;*/SELECT 2;",
			joinStripped:    "SELECT ';';SELECT 2;",
		},
		{
			name:            "whitespace_only_commands_preserved_unstripped",
			input:           " ;SELECT 1;  ;SELECT 2; ",
			notStripped:     []string{" ", "SELECT 1", "  ", "SELECT 2", " "},
			stripped:        []string{"SELECT 1", "SELECT 2"},
			joinNotStripped: " ;SELECT 1;  ;SELECT 2; ",
			joinStripped:    "SELECT 1;SELECT 2;",
		},
		{
			name:            "delimiter_text_inside_literal_not_split",
			input:           "DELIMITER $$\nSELECT '$$';$$\nDELIMITER ;\nSELECT 2;\n",
			notStripped:     []string{"DELIMITER $$\nSELECT '$$';$$\nDELIMITER ;\n", "SELECT 2", "\n"},
			stripped:        []string{"DELIMITER $$\nSELECT '$$';$$\nDELIMITER ;\n", "SELECT 2"},
			joinNotStripped: "DELIMITER $$\nSELECT '$$';$$\nDELIMITER ;\nSELECT 2;\n",
			joinStripped:    "DELIMITER $$\nSELECT '$$';$$\nDELIMITER ;\nSELECT 2;",
		},
		{
			name:            "leading_and_trailing_semicolon_runs",
			input:           ";;;SELECT 1;SELECT 2;;;",
			notStripped:     []string{"SELECT 1", "SELECT 2"},
			stripped:        []string{"SELECT 1", "SELECT 2"},
			joinNotStripped: "SELECT 1;SELECT 2;",
			joinStripped:    "SELECT 1;SELECT 2;",
		},
		{
			name:         "command_with_comment",
			input:        " /* foo1 */ bat \n baz  ; ",
			notStripped:  []string{" /* foo1 */ bat \n baz  ", " "},
			stripped:     []string{"bat baz"},
			joinStripped: "bat baz;",
		},
		{
			name:         "commands_with_comment",
			input:        " /* foo2 */ bar \n ;baz  ; ",
			notStripped:  []string{" /* foo2 */ bar \n ", "baz  ", " "},
			stripped:     []string{"bar", "baz"},
			joinStripped: "bar;baz;",
		},
		{
			name:            "two_delimited_commands",
			input:           "DELIMITER $$\nSELECT 1; $$\nSELECT 2$$\n",
			notStripped:     []string{"DELIMITER $$\nSELECT 1; $$\nDELIMITER ;\n", "DELIMITER $$\n\nSELECT 2$$\nDELIMITER ;\n", "\n"},
			stripped:        []string{"DELIMITER $$\nSELECT 1; $$\nDELIMITER ;\n", "DELIMITER $$\nSELECT 2$$\nDELIMITER ;\n"},
			joinNotStripped: "DELIMITER $$\nSELECT 1; $$\nSELECT 2$$\n\nDELIMITER ;\n\n",
			joinStripped:    "DELIMITER $$\nSELECT 1; $$SELECT 2$$\nDELIMITER ;\n",
		},
		{
			name:            "delimited_then_not_delimited1",
			input:           "DELIMITER //\nCREATE PROCEDURE p()\nBEGIN;\nSELECT 1;\nEND//\nDELIMITER ;\nSELECT 2;\n",
			stripped:        []string{"DELIMITER //\nCREATE PROCEDURE p() BEGIN; SELECT 1; END//\nDELIMITER ;\n", "SELECT 2"},
			notStripped:     []string{"DELIMITER //\nCREATE PROCEDURE p()\nBEGIN;\nSELECT 1;\nEND//\nDELIMITER ;\n", "SELECT 2", "\n"},
			joinNotStripped: "DELIMITER //\nCREATE PROCEDURE p()\nBEGIN;\nSELECT 1;\nEND//\nDELIMITER ;\nSELECT 2;\n",
			joinStripped:    "DELIMITER //\nCREATE PROCEDURE p() BEGIN; SELECT 1; END//\nDELIMITER ;\nSELECT 2;",
		},
		{
			name:         "delimited_then_not_delimited2",
			input:        "DELIMITER //\nCREATE PROCEDURE p()\nBEGIN;\nSELECT 1;\nEND//\nDELIMITER ;\nSELECT 2;\n",
			stripped:     []string{"DELIMITER //\nCREATE PROCEDURE p() BEGIN; SELECT 1; END//\nDELIMITER ;\n", "SELECT 2"},
			notStripped:  []string{"DELIMITER //\nCREATE PROCEDURE p()\nBEGIN;\nSELECT 1;\nEND//\nDELIMITER ;\n", "SELECT 2", "\n"},
			joinStripped: "DELIMITER //\nCREATE PROCEDURE p() BEGIN; SELECT 1; END//\nDELIMITER ;\nSELECT 2;",
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
			name:         "delimiter_command_without_newline_not_activated",
			input:        "DELIMITER $$ SELECT 1;",
			stripped:     []string{"DELIMITER $$ SELECT 1"},
			notStripped:  []string{"DELIMITER $$ SELECT 1"},
			joinStripped: "DELIMITER $$ SELECT 1;",
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
		{
			name:        "only_semicolons",
			input:       ";;;",
			notStripped: []string{},
			stripped:    []string{},
		},
		{
			name:            "single_cmd_no_trailing_semicolon",
			input:           "SELECT 1",
			notStripped:     []string{"SELECT 1"},
			stripped:        []string{"SELECT 1"},
			joinNotStripped: "SELECT 1",
			joinStripped:    "SELECT 1",
		},
		{
			name:            "two_cmds_last_without_semicolon",
			input:           "SELECT 1;SELECT 2",
			notStripped:     []string{"SELECT 1", "SELECT 2"},
			stripped:        []string{"SELECT 1", "SELECT 2"},
			joinNotStripped: "SELECT 1;SELECT 2",
			joinStripped:    "SELECT 1;SELECT 2",
		},
		{
			name:         "whitespace_only_input",
			input:        "   \n  \t  ",
			notStripped:  []string{"   \n  \t  "},
			stripped:     []string{},
			joinStripped: "",
		},
		{
			name:            "comment_before_semicolon",
			input:           "/* c */;SELECT 1;",
			notStripped:     []string{"/* c */", "SELECT 1"},
			stripped:        []string{"SELECT 1"},
			joinNotStripped: "/* c */;SELECT 1;",
			joinStripped:    "SELECT 1;",
		},
		{
			name:            "delimiter_no_reset_to_semicolon",
			input:           "DELIMITER $$\nSELECT 1$$\n",
			notStripped:     []string{"DELIMITER $$\nSELECT 1$$\nDELIMITER ;\n", "\n"},
			stripped:        []string{"DELIMITER $$\nSELECT 1$$\nDELIMITER ;\n"},
			joinNotStripped: "DELIMITER $$\nSELECT 1$$\nDELIMITER ;\n\n",
			joinStripped:    "DELIMITER $$\nSELECT 1$$\nDELIMITER ;\n",
		},
		{
			name:            "delimiter_immediately_after_stmt",
			input:           "DELIMITER $$\n$$\nDELIMITER ;\n",
			notStripped:     []string{"DELIMITER $$\n$$\nDELIMITER ;\n"},
			stripped:        []string{"DELIMITER $$\n$$\nDELIMITER ;\n"},
			joinNotStripped: "DELIMITER $$\n$$\nDELIMITER ;\n",
			joinStripped:    "DELIMITER $$\n$$\nDELIMITER ;\n",
		},
		{
			name:            "delimiter_literal_contains_delim_text",
			input:           "DELIMITER $$\nSELECT '$$', 1$$\nDELIMITER ;\n",
			notStripped:     []string{"DELIMITER $$\nSELECT '$$', 1$$\nDELIMITER ;\n"},
			stripped:        []string{"DELIMITER $$\nSELECT '$$', 1$$\nDELIMITER ;\n"},
			joinNotStripped: "DELIMITER $$\nSELECT '$$', 1$$\nDELIMITER ;\n",
			joinStripped:    "DELIMITER $$\nSELECT '$$', 1$$\nDELIMITER ;\n",
		},
		{
			name:            "delimiter_comment_contains_delim_text",
			input:           "DELIMITER $$\nSELECT /* $$ */ 1$$\nDELIMITER ;\n",
			notStripped:     []string{"DELIMITER $$\nSELECT /* $$ */ 1$$\nDELIMITER ;\n"},
			stripped:        []string{"DELIMITER $$\nSELECT 1$$\nDELIMITER ;\n"},
			joinNotStripped: "DELIMITER $$\nSELECT /* $$ */ 1$$\nDELIMITER ;\n",
			joinStripped:    "DELIMITER $$\nSELECT 1$$\nDELIMITER ;\n",
		},
		{
			name:  "three_delimiter_switches",
			input: "DELIMITER $$\nA$$\nDELIMITER //\nB//\nDELIMITER ||\nC||\nDELIMITER ;\nD;\n",
			stripped: []string{
				"DELIMITER $$\nA$$\nDELIMITER ;\n",
				"DELIMITER //\nB//\nDELIMITER ;\n",
				"DELIMITER ||\nC||\nDELIMITER ;\n",
				"D",
			},
			notStripped: []string{
				"DELIMITER $$\nA$$\nDELIMITER ;\n",
				"\nDELIMITER //\nB//\nDELIMITER ;\n",
				"\nDELIMITER ||\nC||\nDELIMITER ;\n",
				"D",
				"\n",
			},
			joinStripped: "DELIMITER $$\nA$$\nDELIMITER //\nB//\nDELIMITER ||\nC||\nDELIMITER ;\nD;",
		},
		{
			name:            "delimiter_semicolons_preserved_in_content",
			input:           "DELIMITER $$\nSELECT 1; SELECT 2;$$\nDELIMITER ;\n",
			notStripped:     []string{"DELIMITER $$\nSELECT 1; SELECT 2;$$\nDELIMITER ;\n"},
			stripped:        []string{"DELIMITER $$\nSELECT 1; SELECT 2;$$\nDELIMITER ;\n"},
			joinNotStripped: "DELIMITER $$\nSELECT 1; SELECT 2;$$\nDELIMITER ;\n",
			joinStripped:    "DELIMITER $$\nSELECT 1; SELECT 2;$$\nDELIMITER ;\n",
		},
		{
			name:            "delimiter_word_mid_line_not_recognized",
			input:           "DELIMITER $$ SELECT 1;",
			notStripped:     []string{"DELIMITER $$ SELECT 1"},
			stripped:        []string{"DELIMITER $$ SELECT 1"},
			joinNotStripped: "DELIMITER $$ SELECT 1;",
			joinStripped:    "DELIMITER $$ SELECT 1;",
		},
		{
			name:            "delimiter_change_without_any_content",
			input:           "DELIMITER $$\nDELIMITER //\nDELIMITER ;\n",
			notStripped:     []string{"DELIMITER $$\nDELIMITER //\nDELIMITER ;\n"},
			stripped:        []string{"DELIMITER $$\nDELIMITER //\nDELIMITER ;\n"},
			joinNotStripped: "DELIMITER $$\nDELIMITER //\nDELIMITER ;\n",
			joinStripped:    "DELIMITER $$\nDELIMITER //\nDELIMITER ;\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
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
			}
			if tc.joinStripped == "" && len(tc.stripped) != 0 {
				tc.joinStripped = ts.Strip().String()
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
	for i, s := range tokens {
		t.Logf(" %s-%d: %q", prefix, i, s.String())
		for j, token := range s {
			t.Logf("  %s-%d-%d: %s %q%s%s", prefix, i, j, token.Type, token.Text,
				tokenPointerString("split", token.Split),
				tokensString("strip", token.Strip...))
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
		{
			raw:      []string{"SELECT \";\"", "';'"},
			stripped: []string{"SELECT \";\"", "';'"},
		},
		{
			raw:      []string{"SELECT 1"},
			stripped: []string{"SELECT 1"},
		},
		{
			raw:      []string{"-- c1\n", "-- c2\n"},
			stripped: []string{},
		},
		{
			raw:      []string{"SELECT 1 /* inline */", "SELECT 2"},
			stripped: []string{"SELECT 1", "SELECT 2"},
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
