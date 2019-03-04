package main

import (
	"fmt"
	"math/big"
)

type syntaxType int

// Constants representing the various syntax types. Used in parse errors to
// indicate the type of syntactical object being expected.
const (
	syntaxExpression syntaxType = iota
	syntaxRightParen
	syntaxNumber
	syntaxBool
	syntaxIdentifier
	syntaxEOF
)

func (t syntaxType) String() string {
	switch t {
	case syntaxExpression:
		return "expression"
	case syntaxRightParen:
		return "')'"
	case syntaxNumber:
		return "number"
	case syntaxBool:
		return "bool"
	case syntaxIdentifier:
		return "identifier"
	case syntaxEOF:
		return "EOF"
	default:
		// shouldn't be possible
		panic(fmt.Errorf("invalid syntax type: %d", t))
	}
}

// errorNodef formats according to a format specifier (see fmt) and returns the
// resulting string as an error node.
func errorNodef(format string, args ...interface{}) *node {
	return &node{nodeError, fmt.Errorf(format, args...)}
}

// expectError represents a specific kind of parse error where there's a
// mismatch between an expected grammatical object and the received token.
type expectError struct {
	want syntaxType
	got  tokenType
}

func newExpectError(want syntaxType, got tokenType) *node {
	return &node{nodeError, &expectError{want: want, got: got}}
}

func (e *expectError) Error() string {
	return fmt.Sprintf("expecting %s; got %s", e.want, e.got)
}

//go:generate stringer -type=nodeType
type nodeType int

// Constants indicating the type of the value stored in the node struct.
const (
	nodeError      nodeType = iota // node.val is set to an object which satisfies the error interface.
	nodeApp                        // node.val is set to an object of type appNode
	nodeLam                        // node.val is set to an object of type lamNode
	nodeIdentifier                 // node.val is set to a string which contains the name of the identifier
	nodeNumber                     // node.val is set to an object of type *big.Int
	nodeBool                       // node.val is set to a boolean value
)

// node represents a generic node in the parse tree.
type node struct {
	typ nodeType
	val interface{}
}

func (n *node) String() string {
	return fmt.Sprintf("<%d(%v):'%v'>", n.typ, n.typ, n.val)
}

// appNode represents a parsed app expression.
type appNode struct {
	fn, arg *node
}

// lamNode represents a parsed lambda function.
type lamNode struct {
	param string
	body  *node
}

// parser contains the parser's execution state.
type parser struct {
	lex *lexer
	buf *token // storage for unnext()
}

// newParser returns a new parser for the given input string.
func newParser(input string) *parser {
	return &parser{
		lex: newLexer(input),
	}
}

// next returns the next token from the lexer.
func (p *parser) next() token {
	if p.buf != nil {
		ret := *p.buf
		p.buf = nil
		return ret
	}
	return p.lex.nextToken()
}

// unnext saves the given token to be returned the next time next() is called.
// There's room for only one token, so the last token saved with unnext() needs
// to be retrieved with next() before another one can be stored. If this
// condition is not met, the function will panic.
func (p *parser) unnext(t token) {
	if p.buf != nil {
		panic(fmt.Errorf("internal parser error: multiple unnext"))
	}
	p.buf = &t
}

// parseIdentifier parses an identifier and returns either an identifier node or
// an error node.
func (p *parser) parseIdentifier() *node {
	tok := p.next()
	if tok.typ != tokenIdentifier {
		return newExpectError(syntaxIdentifier, tok.typ)
	}
	return &node{nodeIdentifier, tok.val}
}

// parseNumber parses a number and returns either a number node or an error
// node.
//
// Precondition: The next token from the lexer is a number token.
func (p *parser) parseNumber() *node {
	tok := p.next()
	n, ok := new(big.Int).SetString(tok.val, 10)
	if !ok {
		return errorNodef("bad number: '%s'", tok.val)
	}
	return &node{nodeNumber, n}
}

// parseBool parses a bool and returns either a bool node or an error node.
//
// Precondition: The next token from the lexer is a bool token.
func (p *parser) parseBool() *node {
	tok := p.next()
	var val bool
	switch tok.val {
	case "true":
		val = true
	case "false":
		val = false
	default:
		// shouldn't be possible since bools are validated by the lexer
		return errorNodef("bad bool: '%s'", tok.val)
	}
	return &node{nodeBool, val}
}

// parseApp parses a function application expression and returns either an app
// node or an error node.
//
// Grammar:
//   expr = "app", expr, expr
//
// Precondition: The 'app' token has been consumed and an expression is being
// expected.
func (p *parser) parseApp() *node {
	app := &appNode{}
	app.fn = p.parseExpression()
	if app.fn.typ == nodeError {
		return app.fn
	}
	app.arg = p.parseExpression()
	if app.arg.typ == nodeError {
		return app.arg
	}
	return &node{nodeApp, app}
}

// parseLam parses a lambda function expression and returns either a lam node or
// an error node.
//
// Grammar:
//   expr = "lam", ident, expr
//
// Precondition: The 'lam' token has been consumed and an identifier is being
// expected.
func (p *parser) parseLam() *node {
	lam := &lamNode{}
	param := p.parseIdentifier()
	if param.typ == nodeError {
		return param
	}
	lam.param = param.val.(string)
	lam.body = p.parseExpression()
	if lam.body.typ == nodeError {
		return lam.body
	}
	return &node{nodeLam, lam}
}

// parseExpression parses an expression and returns a node.
//
// Grammar:
//   expr = "(", expr, ")"
//   | "lam", ident, expr
//   | "app", expr, expr
//   | literal
//   | ident ;
func (p *parser) parseExpression() *node {
	switch tok := p.next(); {
	case tok.typ == tokenLeftParen:
		e := p.parseExpression()
		if e.typ == nodeError {
			return e
		}
		tok := p.next()
		if tok.typ != tokenRightParen {
			return newExpectError(syntaxRightParen, tok.typ)
		}
		return e
	case tok.typ == tokenRightParen:
		return newExpectError(syntaxExpression, tokenRightParen)
	case tok.typ == tokenIdentifier && tok.val == "lam":
		return p.parseLam()
	case tok.typ == tokenIdentifier && tok.val == "app":
		return p.parseApp()
	case tok.typ == tokenNumber:
		p.unnext(tok)
		return p.parseNumber()
	case tok.typ == tokenBool:
		p.unnext(tok)
		return p.parseBool()
	case tok.typ == tokenIdentifier:
		p.unnext(tok)
		return p.parseIdentifier()
	case tok.typ == tokenError:
		return &node{nodeError, fmt.Errorf("%s", tok.val)}
	case tok.typ == tokenEOF:
		return newExpectError(syntaxExpression, tokenEOF)
	default:
		return errorNodef("illegal token: %s", tok)
	}
}

// parse runs the parser and returns the root of the parse tree.
func (p *parser) parse() *node {
	root := p.parseExpression()
	if root.typ == nodeError {
		return root
	}

	// Make sure there aren't any trailing tokens
	if tok := p.next(); tok.typ != tokenEOF {
		return newExpectError(syntaxEOF, tok.typ)
	}
	return root
}

func parseString(s string) *node {
	return newParser(s).parse()
}

func isUnexpectedEOFError(n *node) bool {
	if n.typ != nodeError {
		return false
	}
	switch v := n.val.(type) {
	case *expectError:
		return v.got == tokenEOF
	default:
		return false
	}
}
