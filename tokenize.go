package sqltoken

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

//go:generate enumer -type=TokenType -json

type TokenType int

const (
	Comment            TokenType = iota // 0
	Whitespace                          // 1
	QuestionMark                        // 2, used in MySQL and SQLite substitution
	AtSign                              // 3, used in sqlserver substitution
	DollarNumber                        // 4, used in PostgreSQL and SQLite substitution
	ColonWord                           // 5, used in sqlx substitution
	Literal                             // 6, strings
	Identifier                          // 7, used in SQL Server for many things
	AtWord                              // 8, used in SQL Server and SQLite, subset of Identifier
	Number                              // 9
	Delimiter                           // 10, semicolon except for MySQL when DELIMITER is used
	Punctuation                         // 11
	Word                                // 12
	Other                               // 13, control characters and other non-printables
	DelimiterStatement                  // 14, DELIMITER command - MySQL only
	Empty                               // 15, marker used in Split for a token that should be eliminated in join
)

const Semicolon = Delimiter // Deprecated: for backwards compatibility only

type Token struct {
	DebugExtra // defined in debug_true.go and debug_false.go; go test -tags debugSQLToken -failfast
	Type       TokenType
	Text       string
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

type Tokens []Token

type TokensList []Tokens

// beginTransactionContinuer reports the word that may immediately follow BEGIN for
// BEGIN WORK / BEGIN TRANSACTION-style transaction starters (not a nested compound).
func beginTransactionContinuer(w string) bool {
	if len(w) == 0 {
		return false
	}
	switch w[0] {
	case 't', 'T':
		return strings.EqualFold(w, "transaction")
	case 'w', 'W':
		return strings.EqualFold(w, "work")
	default:
		return false
	}
}

func mysqlImmediatePreviousWordMatches(lastWord string, prevNonSpace Token, hasPrev bool) bool {
	return hasPrev && prevNonSpace.Type == Word && strings.EqualFold(prevNonSpace.Text, lastWord)
}

func mysqlPunctuationLikelyIdentifierContext(prevNonSpace Token, hasPrev bool) bool {
	if !hasPrev || prevNonSpace.Type != Punctuation {
		return false
	}
	switch prevNonSpace.Text {
	case ",", "(", ".", "=", ":=":
		return true
	default:
		return false
	}
}

func mysqlBeginLikelyIdentifier(lastWord string, prevNonSpace Token, hasPrev bool) bool {
	if mysqlPunctuationLikelyIdentifierContext(prevNonSpace, hasPrev) {
		return true
	}
	if !mysqlImmediatePreviousWordMatches(lastWord, prevNonSpace, hasPrev) {
		return false
	}
	switch strings.ToLower(lastWord) {
	// These are all reserved words
	case "all", "as", "by", "declare", "distinct", "fetch", "from", "having", "in", "inout",
		"into", "join", "on", "select", "set", "table", "update", "using", "values",
		"when", "where":
		return true
	default:
		return false
	}
}

func mysqlEndLikelyIdentifier(lastWord string, prevNonSpace Token, hasPrev bool) bool {
	if mysqlPunctuationLikelyIdentifierContext(prevNonSpace, hasPrev) {
		return true
	}
	if !mysqlImmediatePreviousWordMatches(lastWord, prevNonSpace, hasPrev) {
		return false
	}
	switch strings.ToLower(lastWord) {
	case "all", "as", "by", "declare", "distinct", "else", "fetch", "from", "having",
		"in", "inout", "into", "join", "on", "select", "set", "table", "update", "using",
		"values", "when", "where":
		return true
	default:
		return false
	}
}

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

	// BEGIN/END block tracking: lastWord is the previous Word in the current statement;
	// cleared on Delimiter. Used so BEGIN at statement start is a transaction, not a block.
	var lastWord string
	var beginEndDepth int

	// Inside a stored-program body, BEGIN may open a nested compound or be a lone "BEGIN;"
	// (transaction). Nesting depth stays pending until ';' or the next substantive token; see the
	// token() and tokenWord closures in this function.
	var pendingNestedBegin bool

	// True when pendingNestedBegin was set by BEGIN at statement start (lastWord=="" && depth 0).
	// Transaction starters are "BEGIN;" / BEGIN WORK / BEGIN TRANSACTION; compound blocks look like
	// "BEGIN NOT …" (MariaDB) or any other following token at top level (BEGIN then next statement).
	var pendingBeginFromStmtStart bool

	var lastWordIsCreateOrReplace bool

	// SingleStore only.
	// Tracks CREATE ... FUNCTION/PROCEDURE/TRIGGER/EVENT definitions so we can
	// preserve semicolons inside routine definitions in reserved-word mode.
	var inRoutineDefinition bool

	// SingleStore only.
	// true after the first routine-body BEGIN has been seen.
	var routineBodyStarted bool

	// SingleStore routine-root handling is only relevant in reserved-word mode.
	useRoutineDeclareHeuristic := config.BeginEndWordMode == BeginEndReserved

	token := func(t TokenType) {
		if debug {
			fmt.Printf("> %s: {%s}\n", t, s[tokenStart:i])
		}
		if i-tokenStart == 0 {
			return
		}
		// pendingNestedBegin is only set when NoticeBeginEndBlock is enabled (see tokenWord).
		if pendingNestedBegin {
			switch t {
			case Whitespace, Comment, Empty:
				// Still resolving after BEGIN: might be "BEGIN;" or "BEGIN ..." compound.
			case Delimiter:
				// tokenDelimiter resets pending before calling token(Delimiter).
			case Punctuation:
				if s[tokenStart:i] == ";" {
					// BEGIN; -- a transaction
					pendingNestedBegin = false
					pendingBeginFromStmtStart = false
				}
			case Word:
				switch {
				case pendingBeginFromStmtStart && beginEndDepth == 0:
					pendingNestedBegin = false
					pendingBeginFromStmtStart = false
				case strings.EqualFold(s[tokenStart:i], "begin"):
					// leave as is
				default:
					beginEndDepth++
					pendingNestedBegin = false
				}
			default:
				if pendingBeginFromStmtStart && beginEndDepth == 0 {
					pendingNestedBegin = false
					pendingBeginFromStmtStart = false
				} else {
					beginEndDepth++
					pendingNestedBegin = false
				}
			}
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
		if debug {
			var extra []string
			if inRoutineDefinition {
				extra = append(extra, "inRoutineDefinition")
			}
			if routineBodyStarted {
				extra = append(extra, "routineBodyStarted")
			}
			if lastWordIsCreateOrReplace {
				extra = append(extra, "lastWordIsCreateOrReplace")
			}
			if pendingNestedBegin {
				extra = append(extra, "pendingNestedBegin")
			}
			if pendingBeginFromStmtStart {
				extra = append(extra, "pendingBeginFromStmtStart")
			}
			if beginEndDepth > 0 {
				extra = append(extra, fmt.Sprintf("beginEndDepth=%d", beginEndDepth))
			}
			tokens[len(tokens)-1].SetDebug(strings.Join(extra, " "))
		}
	}

	prevTokenIsAtSign := func() bool {
		if len(tokens) == 0 {
			return false
		}
		t := tokens[len(tokens)-1]
		return t.Type == Punctuation && t.Text == "@"
	}

	lastNonSpaceToken := func() (Token, bool) {
		for j := len(tokens) - 1; j >= 0; j-- {
			switch tokens[j].Type {
			case Whitespace, Comment, Empty:
			default:
				return tokens[j], true
			}
		}
		return Token{}, false
	}

	// tokenWord handles Word tokens and tracks BEGIN/END blocks
	tokenWord := func() {
		var deferPendingBegin bool
		var beginDeferredFromStmtStart bool
		if config.NoticeBeginEndBlock {
			w := s[tokenStart:i]
			if useRoutineDeclareHeuristic && len(w) > 0 && beginEndDepth == 0 && lastWordIsCreateOrReplace {
				// Routine-root detection: CREATE/REPLACE + FUNCTION/PROCEDURE/TRIGGER/EVENT
				// starts a routine context before BEGIN appears.
				// First-byte dispatch keeps this check cheap for non-routine words.
				switch w[0] {
				case 'f', 'F':
					// note: "function" is a reserved word in SingleStore, but "@function" is allowed
					if strings.EqualFold(w, "function") {
						inRoutineDefinition = true
						routineBodyStarted = false
						beginEndDepth = 1
					}
				case 'p', 'P':
					// note: "procedure" is a reserved word in SingleStore, but "@procedure" is allowed
					if strings.EqualFold(w, "procedure") {
						inRoutineDefinition = true
						routineBodyStarted = false
						beginEndDepth = 1
					}
				case 't', 'T':
					// note: "trigger" is a reserved word in SingleStore, but "@trigger" is allowed
					if strings.EqualFold(w, "trigger") {
						inRoutineDefinition = true
						routineBodyStarted = false
						beginEndDepth = 1
					}
				case 'e', 'E':
					// note: "event" is a reserved word in SingleStore, but "@event" is allowed
					if strings.EqualFold(w, "event") {
						inRoutineDefinition = true
						routineBodyStarted = false
						beginEndDepth = 1
					}
				}
			}
			if pendingNestedBegin && len(w) > 0 {
				if beginTransactionContinuer(w) {
					pendingNestedBegin = false
					pendingBeginFromStmtStart = false
				} else if pendingBeginFromStmtStart && strings.EqualFold(w, "not") {
					// MariaDB: BEGIN [NOT ATOMIC] … END outside stored programs.
					beginEndDepth++
					pendingNestedBegin = false
					pendingBeginFromStmtStart = false
				} else if pendingBeginFromStmtStart {
					// BEGIN then next statement at script start — transaction + following body,
					// not a BEGIN/END block.
					pendingNestedBegin = false
					pendingBeginFromStmtStart = false
				} else {
					beginEndDepth++
					pendingNestedBegin = false
				}
			}
			if len(w) > 0 {
				// First-byte dispatch avoids ToLower on every word; EqualFold for keywords only.
				switch w[0] {
				case 'b', 'B':
					// note: "begin" is a reserved word in SingleStore, but "@begin" is allowed; "begin" is NOT reserved in MySQL
					if strings.EqualFold(w, "begin") {
						if useRoutineDeclareHeuristic && inRoutineDefinition {
							// SingleStore
							// First BEGIN is the routine body start; later BEGIN opens nested compounds.
							if !routineBodyStarted {
								routineBodyStarted = true
							} else {
								pendingNestedBegin = true
							}
							break
						}
						// Not block depth: @begin user var; SELECT begin ... (identifier).
						// Transaction vs compound: "BEGIN;" (incl. WORK/TRANSACTION) vs "BEGIN NOT …" /
						// "BEGIN stmt…". Deferred until ';' or the next keyword (see pending flags).
						if prevTokenIsAtSign() {
							break
						}
						if config.BeginEndWordMode == BeginEndContextual {
							prevNS, hasPrevNS := lastNonSpaceToken()
							// the previous word hints at using begin as an identifier
							// XA BEGIN xid (MySQL-style): second BEGIN is not a compound opener.
							// Keep this heuristic in contextual mode only.
							if mysqlBeginLikelyIdentifier(lastWord, prevNS, hasPrevNS) || strings.EqualFold(lastWord, "xa") {
								break
							}
						}
						deferPendingBegin = true
						beginDeferredFromStmtStart = lastWord == "" && beginEndDepth == 0
					}
				case 'c', 'C':
					// note: "case" is a reserved word in SingleStore; "case" is reserved in MySQL
					if strings.EqualFold(w, "case") {
						// CASE ... END - increment so the END doesn't close a BEGIN block
						// But not for END CASE (closing a CASE statement)
						if !strings.EqualFold(lastWord, "end") {
							beginEndDepth++
						}
					}
				case 'e', 'E':
					// note: "end" is a reserved word in SingleStore, but "@end" is allowed; "end" is NOT reserved in MySQL
					if strings.EqualFold(w, "end") {
						endAsIdent := prevTokenIsAtSign()
						if config.BeginEndWordMode == BeginEndContextual && !endAsIdent {
							prevNS, hasPrevNS := lastNonSpaceToken()
							endAsIdent = mysqlEndLikelyIdentifier(lastWord, prevNS, hasPrevNS)
						}
						if beginEndDepth > 0 && !endAsIdent {
							beginEndDepth--
							if useRoutineDeclareHeuristic && inRoutineDefinition && beginEndDepth == 0 {
								inRoutineDefinition = false
								routineBodyStarted = false
							}
						}
					}
				case 'i', 'I':
					// note: "if" is a reserved word in SingleStore, but "@if" is allowed; "if" is reserved in MySQL
					if strings.EqualFold(w, "if") {
						// END IF
						if strings.EqualFold(lastWord, "end") {
							beginEndDepth++
						}
					}
				case 'l', 'L':
					// note: "loop" is a reserved word in SingleStore, but "@loop" is allowed; "loop" is reserved in MySQL
					if strings.EqualFold(w, "loop") {
						// END LOOP
						if strings.EqualFold(lastWord, "end") {
							beginEndDepth++
						}
					}
				case 'r', 'R':
					// note: "repeat" is a reserved word in SingleStore, but "@repeat" is allowed; "repeat" is reserved in MySQL
					if strings.EqualFold(w, "repeat") {
						// END REPEAT
						if strings.EqualFold(lastWord, "end") {
							beginEndDepth++
						}
					}
				case 'w', 'W':
					// note: "while" is a reserved word in SingleStore, but "@while" is allowed; "while" is reserved in MySQL
					if strings.EqualFold(w, "while") {
						// END WHILE
						if strings.EqualFold(lastWord, "end") {
							beginEndDepth++
						}
					}
				}
			}
			lastWord = w
			if useRoutineDeclareHeuristic {
				lastWordIsCreateOrReplace = false
				if len(w) > 0 {
					switch w[0] {
					case 'c', 'C':
						// note: "create" is a reserved word in SingleStore, but "@create" is allowed; "create" is reserved in MySQL
						lastWordIsCreateOrReplace = strings.EqualFold(w, "create")
					case 'r', 'R':
						// note: "replace" is a reserved word in SingleStore, but "@replace" is allowed; "replace" is reserved in MySQL
						lastWordIsCreateOrReplace = strings.EqualFold(w, "replace")
					}
				}
			}
		}
		token(Word)
		if deferPendingBegin {
			pendingNestedBegin = true
			pendingBeginFromStmtStart = beginDeferredFromStmtStart
		}
	}

	// tokenDelimiter handles Delimiter tokens and resets block tracking
	tokenDelimiter := func() {
		if config.NoticeBeginEndBlock {
			beginEndDepth = 0
			lastWord = ""
			lastWordIsCreateOrReplace = false
			pendingNestedBegin = false
			pendingBeginFromStmtStart = false
			inRoutineDefinition = false
			routineBodyStarted = false
		}
		token(Delimiter)
	}

	// BaseState represents being between tokens. It is the initial state.
	//
	// Why is this written with Goto you might ask?  It's written
	// with goto because RE2 can't handle complex regex and PCRE
	// has external dependencies and thus isn't friendly for libraries.
	// So, it could have had a switch with a state variable, but that's
	// just a way to do goto that's lower performance.  Might as
	// well do goto the natural way.
BaseState:
	for i < stop {
		c := s[i]
		i++
		switch c {
		case '/':
			if i < stop && s[i] == '*' {
				goto CStyleComment
			}
			token(Punctuation)
		case '\'':
			if config.NoticeLiteralBackslashEscape {
				goto EscapedSingleQuoteString
			}
			goto SingleQuoteString
		case '"':
			if config.NoticeLiteralBackslashEscape {
				goto EscapedDoubleQuoteString
			}
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
			if config.NoticeDelimiterStatement && delimiter != "" {
				token(Punctuation)
			} else if config.NoticeBeginEndBlock && beginEndDepth > 0 {
				// Sub-statement boundary inside a block: do not reset beginEndDepth, but
				// clear lastWord so the next statement is not confused with END CASE / etc.
				// Lone "BEGIN;" clears pendingNestedBegin without adding nested compound depth.
				if pendingNestedBegin {
					pendingNestedBegin = false
					pendingBeginFromStmtStart = false
				}
				lastWord = ""
				lastWordIsCreateOrReplace = false
				token(Punctuation)
			} else {
				tokenDelimiter()
			}
		case '?':
			if config.NoticeQuestionNumber {
				goto QuestionMark
			}
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
					if config.NoticeDelimitedStrings && i < stop-2 && s[i+1] == '\'' {
						i += 2
						goto DeliminatedString
					}
				case '\'':
					i++
					goto SingleQuoteString
				}
			}
			goto Word
		case 'e', 'E':
			if config.NoticeEscapedStrings && i < stop-1 {
				if s[i] == '\'' {
					i++
					goto EscapedSingleQuoteString
				}
			}
			goto Word
		case 'q', 'Q':
			if config.NoticeDelimitedStrings && i < stop && s[i] == '\'' {
				i++
				goto DeliminatedString
			}
			goto Word
		case 'a' /*b*/, 'c', 'd' /*'e'*/, 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			/*n*/ 'o', 'p' /*q*/, 'r', 's', 't', 'u', 'v', 'w' /*x*/, 'y', 'z',
			'A' /*B*/, 'C', 'D' /*'E'*/, 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
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
		if c == '\'' {
			token(Literal)
			goto BaseState
		}
	}
	token(Literal)
	goto Done

EscapedSingleQuoteString:
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
		if c == '"' {
			token(Literal)
			goto BaseState
		}
	}
	token(Literal)
	goto Done

EscapedDoubleQuoteString:
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
			tokenWord()
			goto BaseState
		case ' ', '\t':
			if config.NoticeDelimiterStatement && strings.EqualFold(s[tokenStart:i], "delimiter") && (tokenStart == 0 || s[tokenStart-1] == '\n') {
				goto DelimiterStatementStart
			}
			tokenWord()
			goto BaseState
		case '\n', '\r', '\b', '\v', '\f',
			'!', '"' /*#*/ /*$*/, '%', '&' /*'*/, '(', ')', '*', '+', '-', '.', '/',
			':', ';', '<', '=', '>', '?', /*@*/
			'[', '\\', ']', '^' /*_*/, '`',
			'{', '|', '}', '~':
			// minor optimization to avoid DecodeRuneInString
			tokenWord()
			goto BaseState
		case '\'':
			if config.NoticeCharsetLiteral && s[tokenStart] == '_' {
				i++
				if config.NoticeLiteralBackslashEscape {
					goto EscapedSingleQuoteString
				}
				goto SingleQuoteString
			}
			if config.NoticeNationalPrefix && (s[tokenStart] == 'n' || s[tokenStart] == 'N') && i-tokenStart == 1 {
				i++
				if config.NoticeLiteralBackslashEscape {
					goto EscapedSingleQuoteString
				}
				goto SingleQuoteString
			}
		}
		r, w := utf8.DecodeRuneInString(s[i:])
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			i += w
			continue
		}
		tokenWord()
		goto BaseState
	}
	tokenWord()
	goto Done

DelimiterStatementStart:
	// we may have to backtrack
	endDelimiterWord = i
	delimiterBuffer = ""

	for i < stop {
		c := s[i]
		switch c {
		case ' ', '\t', '\r', '\b', '\v', '\f':
			i++
		case '\n':
			// invalid delimiter command: backtrack
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
	tokenWord()
	goto BaseState

DelimiterContinues:
	for i < len(s) {
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
		for i < len(s) {
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
		for i < len(s) {
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
		for i < len(s) {
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
		for i < len(s) {
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
		for i <= end {
			if strings.HasPrefix(s[i:], delimiter) {
				stop = i
				i = restoreI
				goto BaseState
			}
			c := s[i]
			i++
			switch c {
			case 'E', 'e':
				if i < len(s) && s[i] == '\'' && !strings.HasPrefix(s[i:], delimiter) {
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
				if i < len(s) && s[i] == '*' && !strings.HasPrefix(s[i:], delimiter) {
					i++
					for ; i+1 < len(s); i++ {
						if s[i] == '*' && s[i+1] == '/' {
							i++
							break
						}
					}
				}
			case '-':
				if i < len(s) && s[i] == '-' && !strings.HasPrefix(s[i:], delimiter) {
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
			tokenWord()
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
	tokenWord()
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

QuestionMark:
	if i < stop {
		c := s[i]
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
			token(QuestionMark)
			goto BaseState
		}
		// ?
		token(QuestionMark)
		goto BaseState
	}
	// ?
	token(QuestionMark)
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
			// This should not happen
			stop = len(s)
			goto BaseState
		} else if i+len(delimiter) <= len(s) {
			if s[i:i+len(delimiter)] == delimiter {
				i += len(delimiter)
				tokenDelimiter()
				goto DelimiterSearchStart
			}
		}
	}
	return tokens
}
