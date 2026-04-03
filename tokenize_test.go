package sqltoken

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
		{Type: Literal, Text: "'a''b'"},
		{Type: Semicolon, Text: ";"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "';\\''"},
	},
	{
		{Type: Word, Text: "c06_doubles"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: `""`},
		{Type: Whitespace, Text: " \t"},
		{Type: Literal, Text: `"a""b"`},
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
		{Type: Literal, Text: "'a''b'"},
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
		{Type: Literal, Text: `"a""b"`},
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

var commonMySQLS2Cases = []Tokens{
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
		{Type: Punctuation, Text: "=@:$"},
	},
	{
		{Type: Word, Text: "m15"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "q"},
		{Type: Literal, Text: "'!martha''s family!'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "m16"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "nq"},
		{Type: Literal, Text: "'!martha''s family!'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "m17"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "@"},
		{Type: Word, Text: "foo"},
		{Type: Punctuation, Text: "$"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "m18"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "f"},
		{Type: Comment, Text: "#o@o$ "},
	},
	{
		{Type: Word, Text: "m19"},
		{Type: Whitespace, Text: " "},
		{Type: Comment, Text: "#foo "},
	},
	{
		{Type: Word, Text: "m20"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "foo"},
		{Type: Punctuation, Text: "$"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "m21"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "foo"},
		{Type: Punctuation, Text: "@"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "m22"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "foo"},
		{Type: Comment, Text: "# "},
	},
	{
		{Type: Word, Text: "m23"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "_foo"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "m24"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: ":"},
		{Type: Word, Text: "名前"},
		{Type: Punctuation, Text: ")"},
	},
	{
		{Type: Word, Text: "m25"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "''"},
		{Type: Whitespace, Text: " \t"},
		{Type: Literal, Text: "'\\''"},
		{Type: Semicolon, Text: ";"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "';\\''"},
	},
	{
		{Type: Word, Text: "m26"},
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
		{Type: Word, Text: "m27"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: `""`},
		{Type: Whitespace, Text: " \t"},
		{Type: Literal, Text: `"\""`},
		{Type: Semicolon, Text: ";"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: `";\""`},
	},
	{
		{Type: Word, Text: "m28"},
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
		{Type: DelimiterStatement, Text: "DELIMITER    29\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "12"},
		{Type: Delimiter, Text: "29"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER /*foo30*/\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "7"},
		{Type: Delimiter, Text: "/*foo30*/"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER 'foo''bar31'\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "'bar''foo'"},
		{Type: Delimiter, Text: "foo'bar31"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER 'foo' 'bar32'\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "'bar''foo'"},
		{Type: Whitespace, Text: " "},
		{Type: Delimiter, Text: "foobar32"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER ---33\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "8"},
		{Type: Comment, Text: "-- comment ---\n"},
		{Type: Whitespace, Text: " "},
		{Type: Delimiter, Text: "---33"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER x y z 34\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "8"},
		{Type: Delimiter, Text: "x"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER ---35\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "9"},
		{Type: Comment, Text: "/* --- comment */"},
		{Type: Whitespace, Text: "\n "},
		{Type: Delimiter, Text: "---35"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER o'foo36$\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "'bar'"},
		{Type: Delimiter, Text: "o'foo36$"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "'baz'"},
		{Type: Delimiter, Text: "o'foo36$"},
		{Type: Whitespace, Text: "\n"},
		{Type: DelimiterStatement, Text: "Delimiter ;\n"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER foo37'\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "7"},
		{Type: Whitespace, Text: " "},
		{Type: Delimiter, Text: "foo37'"},
		{Type: Whitespace, Text: "\n"},
		{Type: DelimiterStatement, Text: "DELIMITER ;\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "2"},
		{Type: Delimiter, Text: ";"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		// 38
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
		// 39
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
	// delimiter at very end of input
	{
		// 40
		{Type: DelimiterStatement, Text: "DELIMITER $$\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "1"},
		{Type: Delimiter, Text: "$$"},
	},
	// delimiter-only content right after DELIMITER statement
	{
		// 41
		{Type: DelimiterStatement, Text: "DELIMITER $$\n"},
		{Type: Delimiter, Text: "$$"},
	},
	// DELIMITER after \r not recognized (only \n counts as line-start)
	{
		// 42
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
		// 43
		{Type: Word, Text: "DELIMITER"},
	},
	// DELIMITER followed only by spaces then EOF
	{
		// 44
		{Type: Word, Text: "DELIMITER"},
		{Type: Whitespace, Text: "   "},
	},
	// DELIMITER immediately followed by \n (no space) — not recognized
	{
		// 45
		{Type: Word, Text: "DELIMITER"},
		{Type: Whitespace, Text: "\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "1"},
		{Type: Delimiter, Text: ";"},
	},
	// single-character delimiter
	{
		// 46
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
		// 47
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
		// 48
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
		// 49
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
		// 50
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
		// 51
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
		// 52
		{Type: DelimiterStatement, Text: "DELIMITER $$\n"},
		{Type: Delimiter, Text: "$$"},
		{Type: Delimiter, Text: "$$"},
		{Type: Whitespace, Text: "\n"},
		{Type: DelimiterStatement, Text: "DELIMITER ;\n"},
	},
	// DELIMITER word mid-line treated as plain word
	{
		// 53
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "DELIMITER"},
		{Type: Whitespace, Text: " "},
		{Type: Punctuation, Text: "$$"},
		{Type: Whitespace, Text: "\n"},
	},
	// Delimiter and comment confusion
	{
		{Type: DelimiterStatement, Text: "DelImiTer *...\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "54"},
		{Type: Punctuation, Text: "/"},
		{Type: Delimiter, Text: "*..."},
		{Type: Whitespace, Text: "\n"},
	},
	// Delimiter and literal confusion
	{
		{Type: DelimiterStatement, Text: "dELiMItER '''55'\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "55"},
		{Type: Delimiter, Text: "'55"},
		{Type: Whitespace, Text: "\n"},
	},
	// Delimiter starting with ' preceded by e — E-quoted string handler must not consume it
	{
		{Type: DelimiterStatement, Text: "DELIMITER \"'X\"\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "56"},
		{Type: Punctuation, Text: "+"},
		{Type: Word, Text: "e"},
		{Type: Delimiter, Text: "'X"},
		{Type: Whitespace, Text: "\n"},
		{Type: DelimiterStatement, Text: "DELIMITER ;\n"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER \"e'57\"\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "57"},
		{Type: Delimiter, Text: "e'57"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER \"'58\"\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "58"},
		{Type: Delimiter, Text: "'58"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: DelimiterStatement, Text: "DELIMITER '\"59'\n"},
		{Type: Word, Text: "SELECT"},
		{Type: Whitespace, Text: " "},
		{Type: Number, Text: "59"},
		{Type: Delimiter, Text: "\"59"},
		{Type: Whitespace, Text: "\n"},
	},
	{
		{Type: Word, Text: "select"},
		{Type: Whitespace, Text: " "},
		{Type: QuestionMark, Text: "?"},
		{Type: Punctuation, Text: ","},
		{Type: Whitespace, Text: " "},
		{Type: QuestionMark, Text: "?"},
		{Type: Number, Text: "1"},
		{Type: Punctuation, Text: ","},
		{Type: Whitespace, Text: " "},
		{Type: QuestionMark, Text: "?"},
	},
}

var mySQLCases = []Tokens{
	{
		// MySQL does not support E'...' prefix — E is a plain word
		{Type: Word, Text: "m27a"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "e"},
		{Type: Literal, Text: "'foo\\'bar'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "m27b"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "E"},
		{Type: Literal, Text: "'foo\\'bar'"},
		{Type: Whitespace, Text: " "},
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
		{Type: Word, Text: "m16"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "n'martha''s family'"},
		{Type: Whitespace, Text: " "},
	},
}

var singleStoreCases = []Tokens{
	{
		// no support for n prefix for literals
		{Type: Word, Text: "m14"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "n"},
		{Type: Literal, Text: "'national charset'"},
	},
	{
		// _charset prefix IS supported by SingleStore
		{Type: Word, Text: "m14"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "_utf8'redundent'"},
	},
	{
		// no support for n prefix for literals
		{Type: Word, Text: "m16"},
		{Type: Whitespace, Text: " "},
		{Type: Word, Text: "n"},
		{Type: Literal, Text: "'martha''s family'"},
		{Type: Whitespace, Text: " "},
	},
	{
		// SingleStore supports E'...' prefix (unlike MySQL)
		{Type: Word, Text: "s01"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "E'foo\\'bar'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "s02"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "e'foo\\'bar'"},
		{Type: Whitespace, Text: " "},
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
		// PostgreSQL accepts n'...' as a national string literal
		{Type: Word, Text: "p16"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "n'mysql only'"},
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
		// PostgreSQL accepts n'...' as a national string literal
		{Type: Word, Text: "p18"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "n'martha''s family'"},
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
	{
		{Type: Word, Text: "p32"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "E'martha''s family'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p33"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "E'martha\\'s family'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p34"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "E'back \\ slash'"},
		{Type: Whitespace, Text: " "},
	},
	{
		{Type: Word, Text: "p35"},
		{Type: Whitespace, Text: " "},
		{Type: Literal, Text: "e'foo\\'bar'"},
		{Type: Whitespace, Text: " "},
	},
}

var sqliteCases = []Tokens{
	{
		{Type: Word, Text: "select"},
		{Type: Whitespace, Text: " "},
		{Type: QuestionMark, Text: "?"},
		{Type: Punctuation, Text: ","},
		{Type: Whitespace, Text: " "},
		{Type: QuestionMark, Text: "?1"},
		{Type: Punctuation, Text: ","},
		{Type: Whitespace, Text: " "},
		{Type: QuestionMark, Text: "?"},
	},
	{
		{Type: Word, Text: "select"},
		{Type: Whitespace, Text: " "},
		{Type: QuestionMark, Text: "?"},
		{Type: Punctuation, Text: ","},
		{Type: Whitespace, Text: " "},
		{Type: QuestionMark, Text: "?123"},
		{Type: Punctuation, Text: ","},
		{Type: Whitespace, Text: " "},
		{Type: ColonWord, Text: ":one"},
		{Type: Punctuation, Text: ","},
		{Type: Whitespace, Text: " "},
		{Type: AtWord, Text: "@two"},
		{Type: Punctuation, Text: ","},
		{Type: Whitespace, Text: " "},
		{Type: DollarNumber, Text: "$456"},
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

// SQLx, SQLite
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

func doTests(t *testing.T, prefix string, config Config, cases ...[]Tokens) {
	for _, tcl := range cases {
		for i, tc := range tcl {
			tc := tc
			desc := "null"
			if len(tc) > 0 {
				desc = tc[0].Text
			}
			t.Run(fmt.Sprintf("%s%03d_%s", prefix, i+1, desc), func(t *testing.T) {
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
	doTests(t, "common_", MySQLConfig(), commonCases, commonMySQLS2Cases)
	doTests(t, "mysql_", MySQLConfig(), commonCases, mySQLCases)
}

func TestSingleStoreTokenizing(t *testing.T) {
	doTests(t, "common_", SingleStoreConfig(), commonCases, commonMySQLS2Cases)
	doTests(t, "mysql_", SingleStoreConfig(), commonCases, singleStoreCases)
}

func TestPostgresSQLTokenizing(t *testing.T) {
	doTests(t, "psql_", PostgreSQLConfig(), commonCases, postgreSQLCases)
}

func TestSQLiteTokenizing(t *testing.T) {
	doTests(t, "sqlite_", SQLiteConfig(), commonCases, sqliteCases, oddball2Cases)
}

func TestOracleTokenizing(t *testing.T) {
	doTests(t, "oracle_", OracleConfig(), commonCases, oracleCases)
}

func TestSQLServerTokenizing(t *testing.T) {
	doTests(t, "sqlsvr_", SQLServerConfig(), commonCases, sqlServerCases)
}

func TestOddbal1Tokenizing(t *testing.T) {
	c := SQLServerConfig()
	c.NoticeAtWord = false
	doTests(t, "oddball_", c, commonCases, oddball1Cases)
}

func TestOddbal2Tokenizing(t *testing.T) {
	c := MySQLConfig()
	c.NoticeColonWord = true
	c.ColonWordIncludesUnicode = true
	doTests(t, "oddball2_", c, commonCases, oddball2Cases)
}

func TestSeparatePunctuationTokenizing(t *testing.T) {
	c := MySQLConfig()
	c.NoticeColonWord = true
	c.NoticeQuestionMark = true
	c.SeparatePunctuation = true
	doTests(t, "punct_", c, separatePunctuationCases)
}
