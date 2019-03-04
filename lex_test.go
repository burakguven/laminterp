package main

import (
	"testing"
)

func mktok(typ tokenType, val string) token {
	return token{typ, val}
}

var (
	eofTok        = mktok(tokenEOF, "")
	leftParenTok  = mktok(tokenLeftParen, "(")
	rightParenTok = mktok(tokenRightParen, ")")
	appTok        = mktok(tokenIdentifier, "app")
	gtTok         = mktok(tokenIdentifier, "gt")
	ifTok         = mktok(tokenIdentifier, "if")
	lamTok        = mktok(tokenIdentifier, "lam")
	xTok          = mktok(tokenIdentifier, "x")
	trueTok       = mktok(tokenBool, "true")
	falseTok      = mktok(tokenBool, "false")
	oneTok        = mktok(tokenNumber, "1")
	threeTok      = mktok(tokenNumber, "3")
	fiveTok       = mktok(tokenNumber, "5")
	minusFiveTok  = mktok(tokenNumber, "-5")
	tenTok        = mktok(tokenNumber, "10")
)

type lexTest struct {
	name   string
	input  string
	tokens []token
}

var lexTests = []lexTest{
	{"empty", "", []token{eofTok}},
	{"spaces", " \t\r\n", []token{eofTok}},
	{"number", "3", []token{threeTok, eofTok}},
	{"negative number", "-5", []token{minusFiveTok, eofTok}},
	{"minus sign without number", "-", []token{errorTokenf("bad number syntax: '-'")}},
	{"naked bool", "true", []token{trueTok, eofTok}},
	{"bad number", "3/", []token{errorTokenf("bad number syntax: '3/'")}},
	{"lam", "lam x x", []token{lamTok, xTok, xTok, eofTok}},
	{"bool", "app lam x true false",
		[]token{appTok, lamTok, xTok, trueTok, falseTok, eofTok}},
	{"leading space", "   app lam x x 3",
		[]token{appTok, lamTok, xTok, xTok, threeTok, eofTok}},
	{"trailing space", "app lam x x 3    ",
		[]token{appTok, lamTok, xTok, xTok, threeTok, eofTok}},
	{"space in the middle", "app    lam x   x 3",
		[]token{appTok, lamTok, xTok, xTok, threeTok, eofTok}},
	{"bad identifier", "lam x' x",
		[]token{lamTok, errorTokenf("bad identifier syntax: 'x''")}},
	{"illegal character", "lam x x ]",
		[]token{lamTok, xTok, xTok, errorTokenf("illegal character: ']'")}},
	{"app", "app lam x x 3",
		[]token{appTok, lamTok, xTok, xTok, threeTok, eofTok}},
	{"example", "app app app if (app app gt 3 1) 10 5", []token{
		appTok, appTok, appTok, ifTok,
		leftParenTok, appTok, appTok, gtTok, threeTok, oneTok, rightParenTok,
		tenTok, fiveTok, eofTok}},
}

func collectTokens(input string) []token {
	var tokens []token
	l := newLexer(input)
	for {
		t := l.nextToken()
		tokens = append(tokens, t)
		if t.typ == tokenError || t.typ == tokenEOF {
			break
		}
	}
	return tokens
}

func tokensEqual(a, b []token) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].typ != b[i].typ || a[i].val != b[i].val {
			return false
		}
	}
	return true
}

func TestLexer(t *testing.T) {
	for _, lt := range lexTests {
		tokens := collectTokens(lt.input)
		if !tokensEqual(tokens, lt.tokens) {
			t.Errorf("[%s]\ninput: %q\nwant: %s\ngot: %s\n", lt.name, lt.input, lt.tokens, tokens)
		}
	}
}
