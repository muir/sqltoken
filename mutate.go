package sqltoken

import (
	"strings"

	"github.com/muir/list"
)

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
// Empty (just comments/whitespace) commands
// are eliminated and are not recoverable with Join().
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
// one per command (splitting on ";" or the current delimiter). It
// does not Strip() the commands.
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
	for i := 0; i < len(ts); i++ {
		t := ts[i]
		switch t.Type {
		case DelimiterStatement:
			if delimiter != "" && hasContents && i-start > 0 {
				// flush accumulated whitespace, comment etc.
				r = append(r, wrapIfNeeded(hasContents, needsWrap, false, ts[start:i+1], nil))
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
					default:
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
			needsUnwrap = false
			hasContents = false
		case Whitespace, Comment, Empty:
			// nothing
		default:
			if delimiter != "" {
				if !hasDelimiterStatement {
					needsWrap = delimiter
				}
				needsUnwrap = true
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
	if needsUnwrap {
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
				if token.Split.Type == Empty && (i < indexLastReal || (i == indexLastReal && j < len(tokens)-1)) {
					continue
				}
				token = token.Copy()
				token.Split = nil
			}
			if j == 0 && i != 0 && token.Type == DelimiterStatement && !strings.HasPrefix(token.Text, "\n") && len(rejoined) > 0 && !strings.HasSuffix(rejoined[len(rejoined)-1].Text, "\n") {
				rejoined = append(rejoined, Token{
					Type: Whitespace,
					Text: "\n",
				})
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
