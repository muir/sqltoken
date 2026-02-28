package sqltoken

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/muir/list"
)

//go:generate enumer -type=TokenType -json

type TokenType int

const (
	Comment            TokenType = iota // 0
	Whitespace                          // 1
	QuestionMark                        // 2, used in MySQL substitution
	AtSign                              // 3, used in sqlserver substitution
	DollarNumber                        // 4, used in PostgreSQL substitution
	ColonWord                           // 5, used in sqlx substitution
	Literal                             // 6, strings
	Identifier                          // 7, used in SQL Server for many things
	AtWord                              // 8, used in SQL Server, subset of Identifier
	Number                              // 9
	Delimiter                           // 10, semicolon except for MySQL when DELIMITER is used
	Punctuation                         // 11
	Word                                // 12
	Other                               // 13, control characters and other non-printables
	DelimiterStatement                  // 14, DELIMITER command - MySQL only
	Empty                               // 15, marker used in Split for a token that should be elimited in join
)

const Semicolon = Delimiter // Deprecated: for backwards compatibility only

type Token struct {
	Type TokenType
	Text string
	// Split is set on the last token in a Tokens after CmdSplit/CmdSplitUnstripped
	// to capture the ; discarded by splitting.
	// Do not set manually.
	Split *Token
	// Strip captures what was before when you Strip() a tokens. It can be set on
	// any Token and it includes the Token itself
	// Do not set manually.
	Strip Tokens
}

func (t Token) Copy() Token {
	t.Strip = copySlice(t.Strip)
	return t
}

// Config specifies the behavior of Tokenize as relates to behavior
// that differs between SQL implementations
type Config struct {
	// Tokenize ? as type Question (used by MySQL)
	NoticeQuestionMark bool

	// Tokenize $7 as type DollarNumber (PostgreSQL)
	NoticeDollarNumber bool

	// Tokenize :word as type ColonWord (sqlx, Oracle)
	NoticeColonWord bool

	// Tokenize :word with unicode as ColonWord (sqlx)
	ColonWordIncludesUnicode bool

	// Tokenize # as type comment (MySQL)
	NoticeHashComment bool

	// $q$ stuff $q$ and $$stuff$$ quoting (PostgreSQL)
	NoticeDollarQuotes bool

	// NoticeHexValues 0xa0 x'af' X'AF' (MySQL)
	NoticeHexNumbers bool

	// NoticeBinaryValues 0x01 b'01' B'01' (MySQL)
	NoticeBinaryNumbers bool

	// NoticeUAmpPrefix U& utf prefix U&"\0441\043B\043E\043D" (PostgreSQL)
	NoticeUAmpPrefix bool

	// NoticeCharsetLiteral _latin1'string' n'string' (MySQL)
	NoticeCharsetLiteral bool

	// NoticeNotionalStrings [nN]'...''...' (Oracle, SQL Server)
	NoticeNotionalStrings bool

	// NoticeDelimitedStrings [nN]?[qQ]'DELIM .... DELIM' (Oracle)
	NoticeDeliminatedStrings bool

	// NoticeTypedNumbers nn.nnEnn[fFdD] (Oracle)
	NoticeTypedNumbers bool

	// NoticeMoneyConstants $10 $10.32 (SQL Server)
	NoticeMoneyConstants bool

	// NoticeAtWord @foo (SQL Server)
	NoticeAtWord bool

	// NoticeAtIdentifiers _baz @fo$o @@b#ar #foo ##b@ar(SQL Server)
	NoticeIdentifiers bool

	// SeparatePunctuation prevents merging successive punctuation into a single token
	SeparatePunctuation bool

	// NoticeDelimiter DELIMITER END;
	NoticeDelimiter bool
}

// WithNoticeQuestionMark enables parsing question marks using the QuestionMark token
func (c Config) WithNoticeQuestionMark() Config {
	c.NoticeQuestionMark = true
	return c
}

// WithNoticeDollarNumber enables parsing dollar parameters ($1) for PostgreSQL using the DollarNumber token
func (c Config) WithNoticeDollarNumber() Config {
	c.NoticeDollarNumber = true
	return c
}

// WithNoticeColonWord enables parsing for named parameters using the ColonWord token
func (c Config) WithNoticeColonWord() Config {
	c.NoticeColonWord = true
	return c
}

// WithColonWordIncludesUnicode enables unicode name parsing at a small performance cost
func (c Config) WithColonWordIncludesUnicode() Config {
	c.ColonWordIncludesUnicode = true
	return c
}

// WithNoticeHashComment enables parsing for '#' comments (MySQL) using the Punctuation token
func (c Config) WithNoticeHashComment() Config {
	c.NoticeHashComment = true
	return c
}

// WithNoticeDollarQuotes enables dollar quote $$parsing$$ for PostgreSQL using the DollarNumber token
func (c Config) WithNoticeDollarQuotes() Config {
	c.NoticeDollarQuotes = true
	return c
}

// WithNoticeHexNumbers enables quoted hex number parsing (x'af') using the HexNumber token (MySQL)
func (c Config) WithNoticeHexNumbers() Config {
	c.NoticeHexNumbers = true
	return c
}

// WithNoticeBinaryNumbers enables quoted binary number parsing (b'01') using the BinaryNumber token (MySQL)
func (c Config) WithNoticeBinaryNumbers() Config {
	c.NoticeBinaryNumbers = true
	return c
}

// WithNoticeUAmpPrefix enables U& prefix parsing (U&"\0441\043B\043E\043D") using the Literal token (PostgreSQL)
func (c Config) WithNoticeUAmpPrefix() Config {
	c.NoticeUAmpPrefix = true
	return c
}

// WithNoticeCharsetLiteral enables charset literal parsing (_latin1'string') using the Literal token (MySQL)
func (c Config) WithNoticeCharsetLiteral() Config {
	c.NoticeCharsetLiteral = true
	return c
}

// WithNoticeNotionalStrings enables notional string parsing (n'string') using the Literal token (Oracle, SQL Server)
func (c Config) WithNoticeNotionalStrings() Config {
	c.NoticeNotionalStrings = true
	return c
}

// WithNoticeDeliminatedStrings enables deliminated string parsing (q'DELIM .... DELIM') using the Literal token (Oracle)
func (c Config) WithNoticeDeliminatedStrings() Config {
	c.NoticeDeliminatedStrings = true
	return c
}

// WithNoticeTypedNumbers enables typed number parsing (nn.nnEnn[fFdD]) using the Number token (Oracle)
func (c Config) WithNoticeTypedNumbers() Config {
	c.NoticeTypedNumbers = true
	return c
}

// WithNoticeMoneyConstants enables money constant parsing ($10 $10.32) using the DollarNumber token (SQL Server)
func (c Config) WithNoticeMoneyConstants() Config {
	c.NoticeMoneyConstants = true
	return c
}

// WithNoticeAtWord enables parsing for '@foo' (SQL Server) using the AtWord token
func (c Config) WithNoticeAtWord() Config {
	c.NoticeAtWord = true
	return c
}

// WithNoticeIdentifiers enables parsing for identifiers (SQL Server) using the Identifier token
func (c Config) WithNoticeIdentifiers() Config {
	c.NoticeIdentifiers = true
	return c
}

// WithSeparatePunctuation enables separating successive punctuation into separate tokens
func (c Config) WithSeparatePunctuation() Config {
	c.SeparatePunctuation = true
	return c
}

func (c Config) WithNoticeDelimiter() Config {
	c.NoticeDelimiter = true
	return c
}

func (c Config) combineOkay(t TokenType) bool {
	// nolint:exhaustive
	switch t {
	case Number, QuestionMark, DollarNumber, ColonWord, Delimiter, DelimiterStatement:
		return false
	case Punctuation:
		return !c.SeparatePunctuation
	}
	return true
}

type Tokens []Token

func (ts Tokens) Copy() Tokens {
	if ts == nil {
		return nil
	}
	n := make(Tokens, len(ts))
	for i, t := range ts {
		n[i] = t.Copy()
	}
	return n
}

type TokensList []Tokens

func (tl TokensList) Copy() TokensList {
	if tl == nil {
		return nil
	}
	n := make(TokensList, len(tl))
	for i, ts := range tl {
		n[i] = ts.Copy()
	}
	return n
}

// OracleConfig returns a parsing configuration that is appropriate
// for parsing Oracle's SQL
func OracleConfig() Config {
	// https://docs.oracle.com/en/database/oracle/oracle-database/19/sqlrf/Literals.html
	return Config{}.
		WithNoticeNotionalStrings().
		WithNoticeDeliminatedStrings().
		WithNoticeTypedNumbers().
		WithNoticeColonWord()
}

// SQLServerConfig returns a parsing configuration that is appropriate
// for parsing SQLServer's SQL
func SQLServerConfig() Config {
	return Config{}.
		WithNoticeNotionalStrings().
		WithNoticeHexNumbers().
		WithNoticeMoneyConstants().
		WithNoticeAtWord().
		WithNoticeIdentifiers()
}

// MySQL returns a parsing configuration that is appropriate
// for parsing MySQL, MariaDB, and SingleStore SQL.
func MySQLConfig() Config {
	return Config{}.
		WithNoticeQuestionMark().
		WithNoticeHashComment().
		WithNoticeHexNumbers().
		WithNoticeBinaryNumbers().
		WithNoticeCharsetLiteral().
		WithNoticeDelimiter()
}

// PostgreSQL returns a parsing configuration that is appropriate
// for parsing PostgreSQL and CockroachDB SQL.
func PostgreSQLConfig() Config {
	return Config{}.
		WithNoticeDollarNumber().
		WithNoticeDollarQuotes().
		WithNoticeUAmpPrefix()
}

// TokenizeMySQL breaks up MySQL / MariaDB / SingleStore SQL strings into
// Token objects.
func TokenizeMySQL(s string) Tokens {
	return Tokenize(s, MySQLConfig())
}

// TokenizePostgreSQL breaks up PostgreSQL / CockroachDB SQL strings into
// Token objects.
func TokenizePostgreSQL(s string) Tokens {
	return Tokenize(s, PostgreSQLConfig())
}

const debug = false

// Tokenize breaks up SQL strings into Token objects.  No attempt is made
// to break successive punctuation.
func Tokenize(s string, config Config) Tokens {
	if len(s) == 0 {
		return []Token{}
	}
	stop := len(s)
	tokens := make([]Token, 0, len(s)/5)
	tokenStart := 0
	var i int
	var firstDollarEnd int
	var runeDelim rune
	var charDelim byte
	var endDelimiterWord int
	var delimiterBuffer string
	var delimiter string

	// Why is this written with Goto you might ask?  It's written
	// with goto because RE2 can't handle complex regex and PCRE
	// has external dependencies and thus isn't friendly for libraries.
	// So, it could have had a switch with a state variable, but that's
	// just a way to do goto that's lower performance.  Might as
	// well do goto the natural way.

	token := func(t TokenType) {
		if debug {
			fmt.Printf("> %s: {%s}\n", t, s[tokenStart:i])
		}
		if i-tokenStart == 0 {
			return
		}
		if len(tokens) > 0 && tokens[len(tokens)-1].Type == t && config.combineOkay(t) {
			tokens[len(tokens)-1].Text = s[tokenStart-len(tokens[len(tokens)-1].Text) : i]
		} else {
			tokens = append(tokens, Token{
				Type: t,
				Text: s[tokenStart:i],
			})
		}
		tokenStart = i
	}

BaseState:
	for i < stop {
		c := s[i]
		i++
		switch c {
		case '/':
			if i < len(s) && s[i] == '*' {
				goto CStyleComment
			}
			token(Punctuation)
		case '\'':
			goto SingleQuoteString
		case '"':
			goto DoubleQuoteString
		case '-':
			if i < stop && s[i] == '-' {
				goto SkipToEOL
			}
			token(Punctuation)
		case '#':
			if config.NoticeHashComment {
				goto SkipToEOL
			}
			if config.NoticeIdentifiers {
				goto Identifier
			}
			token(Punctuation)
		case '@':
			if config.NoticeAtWord {
				goto AtWordStart
			} else if config.NoticeIdentifiers {
				goto Identifier
			} else {
				token(Punctuation)
			}
		case ';':
			if config.NoticeDelimiter && delimiter != "" {
				token(Punctuation)
			} else {
				token(Delimiter)
			}
		case '?':
			if config.NoticeQuestionMark {
				token(QuestionMark)
			} else {
				token(Punctuation)
			}
		case ' ', '\n', '\r', '\t', '\b', '\v', '\f':
			goto Whitespace
		case '.':
			goto PossibleNumber
		case ':':
			if config.NoticeColonWord {
				goto ColonWordStart
			}
			token(Punctuation)
		case '~', '`', '!', '%', '^', '&', '*', '(', ')', '+', '=', '{', '}', '[', ']',
			'|', '\\', '<', '>', ',':
			token(Punctuation)
		case '$':
			// $1
			// $seq$ stuff $seq$
			// $$stuff$$
			if config.NoticeDollarQuotes || config.NoticeDollarNumber {
				goto Dollar
			}
			token(Punctuation)
		case 'U':
			// U&'d\0061t\+000061'
			if config.NoticeUAmpPrefix && i+1 < stop && s[i] == '&' && s[i+1] == '\'' {
				i += 2
				goto SingleQuoteString
			}
			goto Word
		case 'x', 'X':
			// X'1f' x'1f'
			if config.NoticeHexNumbers && i < stop && s[i] == '\'' {
				i++
				goto QuotedHexNumber
			}
			goto Word
		case 'b', 'B':
			if config.NoticeBinaryNumbers && i < stop && s[i] == '\'' {
				i++
				goto QuotedBinaryNumber
			}
			goto Word
		case 'n', 'N':
			if config.NoticeNotionalStrings && i < stop-1 {
				switch s[i] {
				case 'q', 'Q':
					if config.NoticeDeliminatedStrings && i < stop-2 && s[i+1] == '\'' {
						i += 2
						goto DeliminatedString
					}
				case '\'':
					i++
					goto SingleQuoteString
				}
			}
			goto Word
		case 'q', 'Q':
			if config.NoticeDeliminatedStrings && i < stop && s[i] == '\'' {
				i++
				goto DeliminatedString
			}
			goto Word
		case 'a' /*b*/, 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			/*n*/ 'o', 'p' /*q*/, 'r', 's', 't', 'u', 'v', 'w' /*x*/, 'y', 'z',
			'A' /*B*/, 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			/*N*/ 'O', 'P' /*Q*/, 'R', 'S', 'T' /*U*/, 'V', 'W' /*X*/, 'Y', 'Z',
			'_':
			// This covers the entire alphabet except specific letters that have
			// been handled above.  This case is actually just a performance
			// hack: if there were a letter missing it would be caught below
			// by unicode.IsLetter()
			goto Word
		case '0':
			if config.NoticeHexNumbers && i < stop && s[i] == 'x' {
				i++
				goto HexNumber
			}
			if config.NoticeBinaryNumbers && i < stop && s[i] == 'b' {
				i++
				goto BinaryNumber
			}
			goto Number
		case /*0*/ '1', '2', '3', '4', '5', '6', '7', '8', '9':
			goto Number
		default:
			r, w := utf8.DecodeRuneInString(s[i-1:])
			switch {
			case r == '⎖':
				// "⎖" is the unicode decimal separator -- an alternative to "."
				i += w - 1
				goto NumberNoDot
			case unicode.IsDigit(r):
				i += w - 1
				goto Number
			case unicode.IsPunct(r) || unicode.IsSymbol(r) || unicode.IsMark(r):
				i += w - 1
				token(Punctuation)
			case unicode.IsLetter(r):
				i += w - 1
				goto Word
			case unicode.IsControl(r) || unicode.IsSpace(r):
				i += w - 1
				goto Whitespace
			default:
				i += w - 1
				token(Other)
			}
		}
	}
	goto Done

CStyleComment:
	for i < stop {
		c := s[i]
		i++
		switch c {
		case '*':
			if i < stop && s[i] == '/' {
				i++
				token(Comment)
				goto BaseState
			}
		}
	}
	token(Comment)
	goto Done

SingleQuoteString:
	for i < stop {
		c := s[i]
		i++
		switch c {
		case '\'':
			token(Literal)
			goto BaseState
		case '\\':
			if i < stop {
				i++
			} else {
				token(Literal)
				goto Done
			}
		}
	}
	token(Literal)
	goto Done

DoubleQuoteString:
	for i < stop {
		c := s[i]
		i++
		switch c {
		case '"':
			token(Literal)
			goto BaseState
		case '\\':
			if i < stop {
				i++
			} else {
				token(Literal)
				goto Done
			}
		}
	}
	token(Literal)
	goto Done

SkipToEOL:
	for i < stop {
		c := s[i]
		i++
		switch c {
		case '\n':
			token(Comment)
			goto BaseState
		}
	}
	token(Comment)
	goto Done

Word:
	for i < stop {
		c := s[i]
		switch c {
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
			'_',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// This covers the entire alphabet and numbers.
			// This case is actually just a performance
			// hack: if there were a letter missing it would be caught below
			// by unicode.IsLetter()
			i++
			continue
		case '#', '@', '$':
			if config.NoticeIdentifiers {
				goto Identifier
			}
			token(Word)
			goto BaseState
		case ' ', '\t':
			if config.NoticeDelimiter && strings.EqualFold(s[tokenStart:i], "delimiter") && (tokenStart == 0 || s[tokenStart-1] == '\n') {
				goto DelimiterStatementStart
			}
			token(Word)
			goto BaseState
		case '\n', '\r', '\b', '\v', '\f',
			'!', '"' /*#*/ /*$*/, '%', '&' /*'*/, '(', ')', '*', '+', '-', '.', '/',
			':', ';', '<', '=', '>', '?', /*@*/
			'[', '\\', ']', '^' /*_*/, '`',
			'{', '|', '}', '~':
			// minor optimization to avoid DecodeRuneInString
			token(Word)
			goto BaseState
		case '\'':
			if config.NoticeCharsetLiteral {
				switch s[tokenStart] {
				case 'n', 'N':
					if i-tokenStart == 1 {
						i++
						goto SingleQuoteString
					}
				case '_':
					i++
					goto SingleQuoteString
				}
			}
		}
		r, w := utf8.DecodeRuneInString(s[i:])
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			i += w
			continue
		}
		token(Word)
		goto BaseState
	}
	token(Word)
	goto Done

DelimiterStatementStart:
	// we may have to backtrack
	endDelimiterWord = i
	delimiterBuffer = ""

	for i < stop { // len(c) ? XXX
		c := s[i]
		switch c {
		case ' ', '\t', '\r', '\b', '\v', '\f':
			i++
		case '\n':
			// invaliad delimiter command: backtrack
			goto NotDelimiter
		case '\'':
			i++
			goto DelimiterSingleQuote
		case '"':
			i++
			goto DelimiterDoubleQuote
		default:
			goto DelimiterUnquoted
		}
	}
	// fallthrough

NotDelimiter:
	i = endDelimiterWord
	token(Word)
	goto BaseState

DelimiterContinues:
	for i < stop { // len(c) ? XXX
		c := s[i]
		switch c {
		case ' ', '\t', '\r', '\b', '\v', '\f':
			i++
		case '\n':
			i++
			goto DelimiterFound
		case '\'':
			i++
			goto DelimiterSingleQuote
		case '"':
			i++
			goto DelimiterDoubleQuote
		default:
			goto DelimiterIgnoring
		}
	}

DelimiterSingleQuote:
	{
		delimiterStart := i
		for i < stop { // len(c)? XXX
			c := s[i]
			switch c {
			case '\\', '\n':
				// invalid
				goto NotDelimiter
			case '\'':
				if i+1 < stop {
					switch s[i+1] {
					case '\'':
						// '' inside ' quoted string
						i++
						delimiterBuffer += s[delimiterStart:i]
						i++
						delimiterStart = i
						continue
					case ' ', '\t', '\b':
						delimiterBuffer += s[delimiterStart:i]
						i++
						goto DelimiterContinues
					case '\n':
						delimiterBuffer += s[delimiterStart:i]
						i += 2
						goto DelimiterFound
					}
				}
				delimiterBuffer += s[delimiterStart:i]
				i++
				goto DelimiterContinues
			default:
				i++
			}
		}
		goto NotDelimiter
	}

DelimiterDoubleQuote:
	{
		delimiterStart := i
		for i < stop { // len(c)? XXX
			c := s[i]
			switch c {
			case '\\', '\n':
				// invalid
				goto NotDelimiter
			case '"':
				if i+1 < stop {
					switch s[i+1] {
					case '"':
						// "" inside " quoted string
						i++
						delimiterBuffer += s[delimiterStart:i]
						i++
						delimiterStart = i
						continue
					case ' ', '\t', '\b':
						delimiterBuffer += s[delimiterStart:i]
						i++
						goto DelimiterContinues
					case '\n':
						delimiterBuffer += s[delimiterStart:i]
						i += 2
						goto DelimiterFound
					}
				}
				delimiterBuffer += s[delimiterStart:i]
				i++
				goto DelimiterContinues
			default:
				i++
			}
		}
		goto NotDelimiter
	}

DelimiterUnquoted:
	{
		delimiterStart := i
		for i < stop { // len(c)? XXX
			c := s[i]
			switch c {
			case '\\':
				// invalid
				goto NotDelimiter
			case '\n':
				delimiterBuffer = s[delimiterStart:i]
				i++
				goto DelimiterFound
			case ' ', '\t', '\b', '\r':
				delimiterBuffer = s[delimiterStart:i]
				i++
				goto DelimiterIgnoring
			default:
				i++
			}
		}
		goto NotDelimiter
	}

DelimiterIgnoring:
	{
		for i < stop { // len(c)? XXX
			c := s[i]
			switch c {
			case '\n':
				i++
				goto DelimiterFound
			default:
				i++
			}
		}
		goto NotDelimiter
	}

DelimiterFound:
	delimiter = delimiterBuffer
	token(DelimiterStatement)

	// fallthrough

DelimiterSearchStart:
	{
		restoreI := i
		if delimiter == "" {
			stop = len(s)
			goto BaseState
		}
		end := len(s) - len(delimiter)
		for i < end {
			if strings.HasPrefix(s[i:], delimiter) {
				stop = i
				i = restoreI
				goto BaseState
			}
			c := s[i]
			i++
			switch c {
			case 'E', 'e':
				if i < len(s) && s[i] == '\'' {
					i++
				EQuotedString:
					for i < end {
						c = s[i]
						i++
						switch c {
						case '\\':
							if i+1 < len(s) {
								// skip one
								i++
							} else {
								break EQuotedString
							}
						case '\'':
							break EQuotedString
						}
					}
				}
			case '\'':
			SingleQuotedString:
				for i < end {
					c = s[i]
					i++
					switch c {
					case '\'':
						if s[i] == '\'' {
							i++
						} else {
							break SingleQuotedString
						}
					}
				}
			case '"':
			DoubleQuotedString:
				for i < end {
					c = s[i]
					i++
					switch c {
					case '"':
						if s[i] == '"' {
							i++
						} else {
							break DoubleQuotedString
						}
					}
				}
			case '/':
				if i < len(s) && s[i] == '*' {
					i++
					for ; i+1 < len(s); i++ {
						if s[i] == '*' && s[i+1] == '/' {
							i++
							break
						}
					}
				}
			case '-':
				if i < len(s) && s[i] == '-' {
					i++
					for ; i < len(s); i++ {
						if s[i] == '\n' {
							break
						}
					}
				}
			}
		}

		stop = len(s)
		i = restoreI
		goto BaseState
	}

ColonWordStart:
	if i < stop {
		c := s[i]
		switch c {
		case ':':
			// ::word is :: word, not : :word
			i++
			for i < stop && s[i] == ':' {
				i++
			}
			token(Punctuation)
			goto BaseState
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
			i++
			goto ColonWord
		case '\n', '\r', '\t', '\b', '\v', '\f', ' ',
			'!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', '-', '.', '/',
			/*:*/ ';', '<', '=', '>', '?', '@',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'[', '\\', ']', '^', '_', '`',
			'{', '|', '}', '~':
			// minor optimization to avoid DecodeRuneInString
			token(Punctuation)
			goto BaseState
		default:
			if config.ColonWordIncludesUnicode {
				r, w := utf8.DecodeRuneInString(s[i:])
				if unicode.IsLetter(r) {
					i += w
					goto ColonWord
				}
			}
			token(Punctuation)
			goto BaseState
		}
	}
	token(Punctuation)
	goto Done

ColonWord:
	for i < stop {
		c := s[i]
		switch c {
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'_':
			i++
			continue
		case '.':
			if config.ColonWordIncludesUnicode {
				i++
				continue
			}
		case '\n', '\r', '\t', '\b', '\v', '\f', ' ',
			'!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', '-' /*.*/, '/',
			':', ';', '<', '=', '>', '?', '@',
			'[', '\\', ']', '^' /*_*/, '`',
			'{', '|', '}', '~':
			// minor optimization to avoid DecodeRuneInString
			token(ColonWord)
			goto BaseState
		default:
			if config.ColonWordIncludesUnicode {
				r, w := utf8.DecodeRuneInString(s[i:])
				if unicode.IsLetter(r) || unicode.IsDigit(r) {
					i += w
					continue
				}
			}
			token(ColonWord)
			goto BaseState
		}
	}
	token(ColonWord)
	goto Done

Identifier:
	for i < stop {
		c := s[i]
		switch c {
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'#', '@', '$', '_':
			i++
			continue
		default:
			if i-tokenStart == 1 {
				// # @ $ or _
				token(Punctuation)
			} else {
				token(Identifier)
			}
			goto BaseState
		}
	}
	if i-tokenStart == 1 {
		// # @ $ or _
		token(Punctuation)
	} else {
		token(Identifier)
	}
	goto Done

AtWordStart:
	if i < stop {
		c := s[i]
		switch c {
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
			i++
			goto AtWord
		default:
			if config.NoticeIdentifiers {
				goto Identifier
			}
			// @
			token(Punctuation)
			goto BaseState
		}
	}
	// @
	token(Punctuation)
	goto Done

AtWord:
	for i < stop {
		c := s[i]
		switch c {
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			i++
			continue
		case '#', '@', '$', '_':
			if config.NoticeIdentifiers {
				goto Identifier
			}
			token(AtWord)
			goto BaseState
		default:
			token(AtWord)
			goto BaseState
		}
	}
	token(AtWord)
	goto Done

PossibleNumber:
	if i < stop {
		c := s[i]
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			i++
			goto NumberNoDot
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
			'\n', '\r', '\t', '\b', '\v', '\f', ' ',
			'!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', '-', '.', '/',
			':', ';', '<', '=', '>', '?', '@',
			'[', '\\', ']', '^', '_', '`',
			'{', '|', '}', '~':
			// minor optimization to avoid DecodeRuneInString
			token(Punctuation)
			goto BaseState
		default:
			r, w := utf8.DecodeRuneInString(s[i:])
			i += w
			if unicode.IsDigit(r) {
				goto NumberNoDot
			}
			// .
			token(Punctuation)
			goto BaseState
		}
	}
	// .
	token(Punctuation)
	goto Done

Number:
	for i < stop {
		c := s[i]
		i++
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// okay
		case '.':
			goto NumberNoDot
		case 'e', 'E':
			if i < stop {
				switch s[i] {
				case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
					i++
					goto Exponent
				case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
					'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
					'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
					'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
					'\n', '\r', '\t', '\b', '\v', '\f', ' ',
					'!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', '-', '.', '/',
					':', ';', '<', '=', '>', '?', '@',
					'[', '\\', ']', '^', '_', '`',
					'{', '|', '}', '~':
					// minor optimization to avoid DecodeRuneInString
				default:
					r, w := utf8.DecodeRuneInString(s[i:])
					if unicode.IsDigit(r) {
						i += w
						goto Exponent
					}
				}
			}
			i--
			token(Number)
			goto Word
		case 'd', 'D', 'f', 'F':
			if !config.NoticeTypedNumbers {
				i--
			}
			token(Number)
			goto BaseState
		case 'a', 'b', 'c' /*d*/ /*e*/ /*f*/, 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'A', 'B', 'C' /*D*/ /*E*/ /*F*/, 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
			'\n', '\r', '\t', '\b', '\v', '\f', ' ',
			'!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', '-' /*.*/, '/',
			':', ';', '<', '=', '>', '?', '@',
			'[', '\\', ']', '^', '_', '`',
			'{', '|', '}', '~':
			// minor optimization to avoid DecodeRuneInString
			i--
			token(Number)
			goto BaseState
		default:
			r, w := utf8.DecodeRuneInString(s[i-1:])
			if r == '⎖' {
				i += w - 1
				goto NumberNoDot
			}
			if !unicode.IsDigit(r) {
				i--
				token(Number)
				goto BaseState
			}
			i += w - 1
		}
	}
	token(Number)
	goto Done

NumberNoDot:
	for i < stop {
		c := s[i]
		i++
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// okay
		case 'e', 'E':
			if i < stop {
				switch s[i] {
				case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
					i++
					goto Exponent
				}
			}
			i--
			token(Number)
			goto Word
		case 'd', 'D', 'f', 'F':
			if !config.NoticeTypedNumbers {
				i--
			}
			token(Number)
			goto BaseState
		case 'a', 'b', 'c' /*d*/ /*e*/ /*f*/, 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'A', 'B', 'C' /*D*/ /*E*/ /*F*/, 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
			'\n', '\r', '\t', '\b', '\v', '\f', ' ',
			'!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', '-', '.', '/',
			':', ';', '<', '=', '>', '?', '@',
			'[', '\\', ']', '^', '_', '`',
			'{', '|', '}', '~':
			// minor optimization to avoid DecodeRuneInString
			i--
			token(Number)
			goto BaseState
		default:
			r, w := utf8.DecodeRuneInString(s[i-1:])
			if !unicode.IsDigit(r) {
				i--
				token(Number)
				goto BaseState
			}
			i += w - 1
		}
	}
	token(Number)
	goto Done

Exponent:
	if i < stop {
		c := s[i]
		i++
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			goto ExponentConfirmed
		case 'd', 'D', 'f', 'F':
			if !config.NoticeTypedNumbers {
				i--
			}
			token(Number)
			goto BaseState
		case 'a', 'b', 'c' /*d*/, 'e' /*f*/, 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'A', 'B', 'C' /*D*/, 'E' /*F*/, 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
			'\n', '\r', '\t', '\b', '\v', '\f', ' ',
			'!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', '-', '.', '/',
			':', ';', '<', '=', '>', '?', '@',
			'[', '\\', ']', '^', '_', '`',
			'{', '|', '}', '~':
			// minor optimization to avoid DecodeRuneInString
			i--
			token(Number)
			goto BaseState
		default:
			r, w := utf8.DecodeRuneInString(s[i-1:])
			if !unicode.IsDigit(r) {
				i--
				token(Number)
				goto BaseState
			}
			i += w - 1
			goto ExponentConfirmed
		}
	}
	token(Number)
	goto BaseState

ExponentConfirmed:
	for i < stop {
		c := s[i]
		i++
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// okay
		case 'd', 'D', 'f', 'F':
			if !config.NoticeTypedNumbers {
				i--
			}
			token(Number)
			goto BaseState
		case 'a', 'b', 'c' /*d*/, 'e' /*f*/, 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'A', 'B', 'C' /*D*/, 'E' /*F*/, 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
			'\n', '\r', '\t', '\b', '\v', '\f', ' ',
			'!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', '-', '.', '/',
			':', ';', '<', '=', '>', '?', '@',
			'[', '\\', ']', '^', '_', '`',
			'{', '|', '}', '~':
			// minor optimization to avoid DecodeRuneInString
			i--
			token(Number)
			goto BaseState
		default:
			r, w := utf8.DecodeRuneInString(s[i-1:])
			if !unicode.IsDigit(r) {
				i--
				token(Number)
				goto BaseState
			}
			i += w - 1
		}
	}
	token(Number)
	goto Done

HexNumber:
	for i < stop {
		c := s[i]
		i++
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'a', 'b', 'c', 'd', 'e', 'f',
			'A', 'B', 'C', 'D', 'E', 'F':
			// okay
		default:
			i--
			token(Number)
			goto BaseState
		}
	}
	token(Number)
	goto Done

BinaryNumber:
	for i < stop {
		c := s[i]
		i++
		switch c {
		case '0', '1':
			// okay
		default:
			i--
			token(Number)
			goto BaseState
		}
	}
	token(Number)
	goto Done

Whitespace:
	for i < stop {
		c := s[i]
		i++
		switch c {
		case ' ', '\n', '\r', '\t', '\b', '\v', '\f':
			// whitespace!
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
			'!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', '-', '.', '/',
			':', ';', '<', '=', '>', '?', '@',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'[', '\\', ']', '^', '_', '`',
			'{', '|', '}', '~':
			// minor optimization to avoid DecodeRuneInString
			i--
			token(Whitespace)
			goto BaseState
		default:
			r, w := utf8.DecodeRuneInString(s[i-1:])
			if !unicode.IsSpace(r) && !unicode.IsControl(r) {
				i--
				token(Whitespace)
				goto BaseState
			}
			i += w - 1
		}
	}
	token(Whitespace)
	goto Done

QuotedHexNumber:
	for i < stop {
		c := s[i]
		i++
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'a', 'b', 'c', 'd', 'e', 'f',
			'A', 'B', 'C', 'D', 'E', 'F':
			// okay
		case '\'':
			token(Number)
			goto BaseState
		default:
			i--
			token(Number)
			goto BaseState
		}
	}
	token(Number)
	goto Done

QuotedBinaryNumber:
	for i < stop {
		c := s[i]
		i++
		switch c {
		case '0', '1':
			// okay
		case '\'':
			token(Number)
			goto BaseState
		default:
			i--
			token(Number)
			goto BaseState
		}
	}
	token(Number)
	goto Done

DeliminatedString:
	// We arrive here with s[i] being on the opening delimiter
	// 'Foo''Bar'
	// n'Foo'
	// q'XlsXldsaX'
	// Q'(ls)(Xldsa)'
	// Nq'(ls)(Xldsa)'
	if i < stop {
		c := s[i]
		i++
		switch c {
		case '(':
			charDelim = ')'
			goto DeliminatedStringCharacter
		case '<':
			charDelim = '>'
			goto DeliminatedStringCharacter
		case '[':
			charDelim = ']'
			goto DeliminatedStringCharacter
		case '{':
			charDelim = '}'
			goto DeliminatedStringCharacter
		// [{<(
		case ')', '>', '}', ']',
			'\n', '\r', '\t', '\b', '\v', '\f', ' ':
			// not a valid delimiter
			i -= 2
			token(Word)
			i++
			goto SingleQuoteString
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
			'!', '"', '#', '$', '%', '&', '\'' /*(*/ /*)*/, '*', '+', '-', '.', '/',
			':', ';' /*<*/, '=' /*>*/, '?', '@',
			/*[*/ '\\' /*]*/, '^', '_', '`',
			/*{*/ '|' /*}*/, '~':
			// minor optimzation to avoid DecodeRuneInString
			charDelim = c
			goto DeliminatedStringCharacter
		default:
			r, w := utf8.DecodeRuneInString(s[i-1:])
			if w == 1 {
				charDelim = s[i-1]
				goto DeliminatedStringCharacter
			}
			i += w - 1
			runeDelim = r
			goto DeliminatedStringRune
		}
	}
	token(Word)
	goto Done

DeliminatedStringCharacter:
	for i < stop {
		c := s[i]
		i++
		if c == charDelim && i < stop && s[i] == '\'' {
			i++
			token(Literal)
			goto BaseState
		}
	}
	token(Literal)
	goto Done

DeliminatedStringRune:
	for i < stop {
		r, w := utf8.DecodeRuneInString(s[i:])
		i += w
		if r == runeDelim {
			token(Literal)
			goto BaseState
		}
	}
	token(Literal)
	goto Done

Dollar:
	// $1
	// $seq$ stuff $seq$
	// $$stuff$$
	firstDollarEnd = i
	if i < stop {
		c := s[i]
		if config.NoticeDollarQuotes {
			if c == '$' {
				e := strings.Index(s[i+1:], "$$")
				if e == -1 {
					i = firstDollarEnd
					// $
					token(Punctuation)
					goto BaseState
				}
				i += 3 + e
				token(Literal)
				goto BaseState
			}
			r, w := utf8.DecodeRuneInString(s[i:])
			if unicode.IsLetter(r) {
				i += w
				for i < stop {
					// nolint:govet
					c := s[i]
					r, w := utf8.DecodeRuneInString(s[i:])
					i++
					if c == '$' {
						endToken := s[tokenStart:i]
						e := strings.Index(s[i:], endToken)
						if e == -1 {
							i = firstDollarEnd
							// $
							token(Punctuation)
							goto BaseState
						}
						i += e + len(endToken)
						token(Literal)
						goto BaseState
					} else if unicode.IsLetter(r) {
						i += w - 1
						continue
					} else {
						i = firstDollarEnd
						// $
						token(Punctuation)
						goto BaseState
					}
				}
			}
		}
		if config.NoticeDollarNumber {
			switch c {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				i++
				for i < stop {
					c := s[i]
					i++
					switch c {
					case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
						continue
					}
					i--
					break
				}
				token(DollarNumber)
				goto BaseState
			}
		}
		// $
		token(Punctuation)
		goto BaseState
	}
	// $
	token(Punctuation)
	goto Done

Done:
	if stop < len(s) {
		if delimiter == "" {
			panic("empty delmiter")
		}
		if i+len(delimiter) < len(s) {
			if s[i:i+len(delimiter)] == delimiter {
				i += len(delimiter)
				token(Delimiter)
				goto DelimiterSearchStart
			}
		}

	}
	return tokens
}

// String returns exactly the original text unless the Tokens is a result of
// Strip() or CmdSplit(), or CmdSplitUnstripped().
func (ts Tokens) String() string {
	if len(ts) == 0 {
		return ""
	}
	strs := make([]string, len(ts))
	for i, t := range ts {
		strs[i] = t.Text
	}
	return strings.Join(strs, "")
}

// Strip removes leading/trailing whitespace and semicolons
// and strips all internal comments.  Internal whitespace
// is changed to a single space. Strip is not reversible if there
// is no command present (all whitespace and/or comment).
func (ts Tokens) Strip() Tokens {
	type CSpace int // CSpace are indexes into c.
	if len(ts) == 0 {
		return ts
	}
	i := 0
	// Initial comments/whitespace/etc are skipped
	for i < len(ts) {
		// nolint:exhaustive
		switch ts[i].Type {
		case Comment, Whitespace, Delimiter:
			i++
			continue
		}
		break
	}
	c := make(Tokens, 0, len(ts))
	// lastReal-1 is the index of the last item in c that is a real (not whitespace etc) token
	var lastReal CSpace
	// lastWhitespace is used to prevent multiple whitespaces between real items
	var lastWhitespace CSpace

	// captureSkip is called after appending to c but before updating lastReal
	// the item just appended may or may not end up stripped
	var lastCapture int
	captureSkip := func() {
		count := i - lastCapture
		if count > 0 {
			lastIndex := CSpace(len(c) - 1)
			token := c[lastIndex].Copy()
			token.Strip = ts[lastCapture : i+1] // includes self
			c[lastIndex] = token
		}
		lastCapture = i + 1
	}
	// lastKeptCapture tracks i as lastReal tracks c
	var lastKeptCapture int
	var nonStandardDelimiter bool

	for ; i < len(ts); i++ {
		// nolint:exhaustive
		switch ts[i].Type {
		case Comment:
			// skip it
		case Whitespace:
			// only append whitespace if there hasn't been a whitespace since lastReal
			// and the last captured token doesn't end with whitespace
			if lastWhitespace < lastReal {
				if !lastEndsWithWhitespace(c) {
					if ts[i].Text != " " {
						c = append(c, Token{
							Type:  Whitespace,
							Text:  " ",
							Strip: Tokens{ts[i]},
						})
					} else {
						c = append(c, ts[i])
					}
					captureSkip()
				}
				lastWhitespace = CSpace(len(c))
			}
		case Delimiter:
			c = append(c, ts[i])
			if nonStandardDelimiter {
				lastReal = CSpace(len(c))
				lastKeptCapture = i + 1
			}
			lastWhitespace = lastReal // no whitespace after a delimiter
			captureSkip()
		case DelimiterStatement:
			if len(c) > 0 && !strings.HasPrefix(ts[i].Text, "\n") {
				// Adjust prior whitespace
				lastIndex := len(c) - 1
				last := c[lastIndex]
				switch last.Type {
				case Whitespace:
					if last.Strip == nil {
						last = last.Copy()
						last.Strip = Tokens{c[lastIndex]}
					}
					last.Text = "\n"
					c[lastIndex] = last
				default:
					if !strings.HasSuffix(last.Text, "\n") {
						c = append(c, Token{
							Type:  Whitespace,
							Text:  "\n",
							Strip: make(Tokens, 0),
						})
					}
				}
			}
			nonStandardDelimiter = !delimiterIsSemicolon(ts[i].Text)
			lastWhitespace = CSpace(len(c))
			fallthrough
		default:
			c = append(c, ts[i])
			captureSkip()
			lastReal = CSpace(len(c))
			lastKeptCapture = i + 1
		}
	}
	if lastReal < CSpace(len(c)) {
		c = c[:lastReal]
	}
	if len(c) > 0 {
		count := i - lastKeptCapture
		if count > 0 {
			lastIndex := CSpace(len(c) - 1)
			token := c[lastIndex]
			if token.Strip == nil {
				token = token.Copy()
				token.Strip = ts[lastKeptCapture-1:]
			} else {
				token.Strip = copySlice(token.Strip)
				token.Strip = append(token.Strip, ts[lastKeptCapture:]...)
			}
			token.Split = ts[len(ts)-1].Split
			c[lastIndex] = token
			// c = append(c, Token{Type: Number}) // XXX
		}
	}
	return c
}

func lastEndsWithWhitespace(ts []Token) bool {
	if len(ts) == 0 {
		return false
	}
	lastText := ts[len(ts)-1].Text
	if lastText == "" {
		return false
	}
	switch lastText[len(lastText)-1] {
	case ' ', '\n', '\r', '\b', '\t':
		return true
	}
	return false
}

// Unstrip reverses a Strip
func (ts Tokens) Unstrip() Tokens {
	n := make([]Token, 0, len(ts))
	for _, token := range ts {
		if token.Strip != nil {
			n = append(n, token.Strip...)
		} else {
			n = append(n, token)
		}
	}
	return Tokens(n)
}

// CmdSplit breaks up the token array into multiple token arrays,
// one per command (splitting on ";" or on the delimiter if there
// is a delimiter set) and Strip()ing each of the returned Tokens.
//
// DELIMITER commands become an annotation on each command (in the first
// token) that has a special delimiter (rather than being a stand-alone command).
// That annotation will turn into bracketing each command with
// DELIMITER commands thus producing a result that is longer than the
// original but logically equivalent with each command self-contained.
func (ts Tokens) CmdSplit() TokensList {
	r := ts.CmdSplitUnstripped()
	stripped := make(TokensList, 0, len(r))
	for _, t := range r {
		s := t.Strip()
		if len(s) > 0 {
			stripped = append(stripped, s)
		}
	}
	return stripped
}

// CmdSplitUnstripped breaks up the token array into multiple token arrays,
// one per command (splitting on ";"). It does not Strip() the commands.
// Empty (just comments/whitespace) commands are eliminated and
// are not recoverable with Join().
//
// DELIMITER statements will be repeated on each command so that each
// statement becomes self-contained.
func (ts Tokens) CmdSplitUnstripped() TokensList {
	var r TokensList
	start := 0
	var delimiter string
	// These variables create a little state machine
	// that tracks if delimiter wrapping or unwrapping is required
	var needsWrap string
	var needsUnwrap bool
	var hasDelimiterStatement bool
	var hasContents bool
	for i, t := range ts {
		switch t.Type {
		case DelimiterStatement:
			if delimiter != "" && !hasContents && i-start > 0 {
				// flush accumulated whitespace, comment etc.
				r = append(r, wrapIfNeeded(false, needsWrap, false, ts[start:i+1], nil))
				start = i + 1
				needsWrap = ""
			}
			if delimiterIsSemicolon(t.Text) {
				needsUnwrap = false // we've just unwrapped
				delimiter = ""
			} else {
				delimiter = t.Text
				needsUnwrap = true
				hasDelimiterStatement = true
			}
		case Delimiter:
			tp := &ts[i]
			if needsUnwrap {
			Lookahead:
				for j := i + 1; j < len(ts); j++ {
					switch ts[j].Type {
					case Comment, Whitespace, Empty:
						// keep going
					case DelimiterStatement:
						if delimiterIsSemicolon(ts[j].Text) {
							i = j + 1
							needsUnwrap = false
							tp = nil
							delimiter = ""
						}
						break Lookahead
					}
				}
			}
			r = append(r, wrapIfNeeded(hasContents, needsWrap, needsUnwrap, ts[start:i], tp))
			start = i
			if tp != nil {
				start++
			}
			hasDelimiterStatement = false
			needsWrap = ""
			needsUnwrap = delimiter != "" // we'll need unwrap on the next statement
			hasContents = false
		case Whitespace, Comment, Empty:
			// nothing
		default:
			if delimiter != "" && !hasDelimiterStatement {
				needsWrap = delimiter
			}
			hasContents = true
		}
	}
	if start < len(ts) {
		r = append(r, wrapIfNeeded(hasContents, needsWrap, needsUnwrap, ts[start:], nil))
	}
	return r
}

var empty = Token{Type: Empty}

func wrapIfNeeded(hasContents bool, needsWrap string, needsUnwrap bool, ts []Token, t *Token) Tokens {
	lastIndex := len(ts) - 1
	if lastIndex == -1 {
		return Tokens{}
	}
	if needsWrap == "" && !needsUnwrap && t == nil {
		return Tokens(ts)
	}
	n := make([]Token, 0, len(ts)+2)
	if needsWrap != "" && hasContents {
		n = append(n, Token{
			Type:  DelimiterStatement,
			Text:  needsWrap,
			Split: &empty,
		})
	}
	if t != nil && !needsUnwrap {
		n = append(n, ts[:lastIndex]...)
		last := ts[lastIndex].Copy()
		last.Split = t
		n = append(n, last)
	} else {
		n = append(n, ts...)
	}
	if needsUnwrap && hasContents {
		if t != nil {
			n = append(n, *t)
		}
		n = append(n, Token{
			Type:  DelimiterStatement,
			Text:  "\nDELIMITER ;\n",
			Split: &empty,
		})
	}
	return Tokens(n)
}

const delimiterFrontLen = len("delimiter") + 1

func delimiterIsSemicolon(statement string) bool {
	if len(statement) < delimiterFrontLen {
		return false
	}
	return strings.TrimSpace(statement[delimiterFrontLen:]) == ";"
}

// Strings returns almost exactly the original text of each Tokens (except the semicolons)
// unless Strip() or CmdSplit() was called.
func (tl TokensList) Strings() []string {
	r := make([]string, 0, len(tl))
	for _, ts := range tl {
		s := ts.String()
		if s != "" {
			r = append(r, s)
		}
	}
	return r
}

// Join reverses Split: adding back delimiters between the token lists
//
// Join does not always recreate the original input. It tries to come
// close though. Use of DELIMITER will often create small differences.
func (tl TokensList) Join() Tokens {
	tl = list.FilterEmptySlices(tl)
	if len(tl) == 0 {
		return Tokens{}
	}
	var l int
	for _, tokens := range tl {
		l += len(tokens)
	}
	l += len(tl) - 1
	rejoined := make(Tokens, 0, l)
	indexLastReal := indexLastReal(tl)
	for i, tokens := range tl {
		for j, token := range tokens {
			if token.Type == Empty {
				continue
			}
			if token.Split != nil {
				if token.Split.Type == Empty && i < indexLastReal || (i == indexLastReal && j < len(tokens)-1) {
					continue
				}
				token = token.Copy()
				token.Split = nil
			}
			rejoined = append(rejoined, token)
		}
		last := tokens[len(tokens)-1]
		if last.Split != nil && last.Split.Type != Empty {
			split := *last.Split
			split.Split = nil
			rejoined = append(rejoined, split)
		}
	}
	return rejoined
}

func indexLastReal(tl TokensList) int {
	for i := len(tl) - 1; i >= 0; i-- {
		if !isWhitespaceOnly(tl[i]) {
			return i
		}
	}
	return -1
}

func isWhitespaceOnly(ts Tokens) bool {
	for _, token := range ts {
		switch token.Type {
		case Whitespace, Empty, Comment:
			// nah
		default:
			return false
		}
	}
	return true
}

func copySlice[E any](s []E) []E {
	if s == nil {
		return nil
	}
	n := make([]E, len(s))
	copy(n, s)
	return n
}
