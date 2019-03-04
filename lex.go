package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type tokenType int

// Constants indicating the type of token stored in the token struct.
const (
	tokenError tokenType = iota
	tokenEOF
	tokenNumber
	tokenBool
	tokenIdentifier
	tokenLeftParen
	tokenRightParen
)

func (t tokenType) String() string {
	switch t {
	case tokenError:
		return "error"
	case tokenEOF:
		return "EOF"
	case tokenNumber:
		return "number"
	case tokenBool:
		return "bool"
	case tokenIdentifier:
		return "identifier"
	case tokenLeftParen:
		return "'('"
	case tokenRightParen:
		return "')'"
	default:
		// shouldn't be possible
		panic(fmt.Errorf("invalid token type: %d", t))
	}
}

// token represents a token returned from the lexer.
type token struct {
	typ tokenType
	val string
}

func (t token) String() string {
	if t.typ == tokenEOF {
		return "EOF"
	}
	return fmt.Sprintf("<%d(%v):'%v'>", t.typ, t.typ, t.val)
}

// errorTokenf formats according to a format specifier (see fmt) and returns the
// resulting string as an error token.
func errorTokenf(format string, args ...interface{}) token {
	return token{tokenError, fmt.Sprintf(format, args...)}
}

// lexer contains the lexer's execution state.
type lexer struct {
	input string
	pos   int // current position in input
	start int // start of current token in input
	width int
}

// newLexer creates a new lexer for the given input string.
func newLexer(input string) *lexer {
	return &lexer{
		input: input,
	}
}

var eof = rune(-1)

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	ch, width := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += width
	l.width = width
	return ch
}

// unnext steps back one rune. Only one rune is tracked, so this function can be
// called only once for each call of next(). More than one call will result in
// undefined behavior.
func (l *lexer) unnext() {
	l.pos -= l.width
}

// val returns a string containing all of the runes accumulated so far.
func (l *lexer) val() string {
	return l.input[l.start:l.pos]
}

// emit returns a token with the given type which contains all of the runes
// accumulated so far. It also sets the lexer's current position to the next token.
func (l *lexer) emit(typ tokenType) token {
	val := l.val()
	l.start = l.pos
	return token{typ, val}
}

// skipSpaces advances the lexer's current position to the first non-space rune.
func (l *lexer) skipSpaces() {
	for isSpace(l.next()) {
	}
	l.unnext()
	l.start = l.pos
}

// lexNumber scans a number and returns either a number token or an error token.
// In this language, a number is an arbitrary precision integer.
//
// Grammar:
//   digits  = digit, { digit } ;
//   number  = "-", digits | digits ;
//
// Precondition: The next character is either a minus sign or a digit.
func (l *lexer) lexNumber() token {
	ch := l.next()
	if ch == '-' {
		ch = l.next()
	}
	if !isDigit(ch) {
		if isBoundary(ch) {
			l.unnext()
		}
		return errorTokenf("bad number syntax: '%s'", l.val())
	}
	for {
		ch = l.next()
		if !isDigit(ch) {
			break
		}
	}
	if !isBoundary(ch) {
		return errorTokenf("bad number syntax: '%s'", l.val())
	}
	l.unnext()
	return l.emit(tokenNumber)
}

// lexIdentifier scans an identifier and returns either an identifier token, a
// bool token (special case of identifier), or an error token.
//
// Grammar:
//   ident = letter, { letter | digit }
//   bool  = "true" | "false" ;
//
// Precondition: The first character is a letter that has already been consumed.
func (l *lexer) lexIdentifier() token {
	var ch rune
	for {
		ch = l.next()
		if !isDigit(ch) && !isLetter(ch) {
			break
		}
	}
	if !isBoundary(ch) {
		return errorTokenf("bad identifier syntax: '%s'", l.val())
	}
	l.unnext()

	typ := tokenIdentifier
	switch l.val() {
	case "true", "false":
		typ = tokenBool
	}
	return l.emit(typ)
}

func (l *lexer) nextToken() token {
	l.skipSpaces()

	switch ch := l.next(); {
	case ch == '-' || isDigit(ch):
		l.unnext()
		return l.lexNumber()
	case isLetter(ch):
		return l.lexIdentifier()
	case ch == '(':
		return l.emit(tokenLeftParen)
	case ch == ')':
		return l.emit(tokenRightParen)
	case ch == eof:
		return l.emit(tokenEOF)
	default:
		return errorTokenf("illegal character: '%c'", ch)
	}
}

func isSpace(r rune) bool {
	return strings.ContainsRune(" \t\r\n", r)
}

func isLetter(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// isBoundary returns true if the given rune terminates a run of letters or
// digits. It's analogous to '\b' in regular expressions.
func isBoundary(r rune) bool {
	return isSpace(r) || r == ')' || r == eof
}
