package main

import (
	"math/big"
	"testing"
)

type parseTest struct {
	name  string
	input string
	root  *node
}

func mkapp(fn, arg *node) *node {
	return &node{nodeApp, &appNode{fn, arg}}
}

func mklam(param string, body *node) *node {
	return &node{nodeLam, &lamNode{param, body}}
}

func mkident(name string) *node {
	return &node{nodeIdentifier, name}
}

func mknum(n int64) *node {
	return &node{nodeNumber, big.NewInt(n)}
}

var (
	ifNode   = mkident("if")
	gtNode   = mkident("gt")
	addNode  = mkident("add")
	xNode    = mkident("x")
	yNode    = mkident("y")
	fNode    = mkident("f")
	trueNode = &node{nodeBool, true}
)

var parseTests = []parseTest{
	{"empty", "", errorNodef("expecting expression; got EOF")},
	{"number", "2", mknum(2)},
	{"negative number", "-7", mknum(-7)},
	{"bad number", "2s", errorNodef("bad number syntax: '2s'")},
	{"bool", "true", trueNode},
	{"ident", "x", xNode},
	{"paren", "(x)", xNode},
	{"multiple paren", "(((x)))", xNode},
	{"empty paren", "()", errorNodef("expecting expression; got ')'")},
	{"unclosed paren", "(1", errorNodef("expecting ')'; got EOF")},
	{"unopened paren", "1)", errorNodef("expecting EOF; got ')'")},
	{"paren grouping", "app (lam x x) 2",
		mkapp(mklam("x", xNode), mknum(2))},
	{"noparen", "app app add 1 3",
		mkapp(mkapp(addNode, mknum(1)), mknum(3))},
	{"lam", "lam x x", mklam("x", xNode)},
	{"lam illegal param", "lam 1 x", errorNodef("expecting identifier; got number")},
	{"lam illegal body", "lam x lam 1 y", errorNodef("expecting identifier; got number")},
	{"app", "app app gt 1 2", mkapp(mkapp(gtNode, mknum(1)), mknum(2))},
	{"app illegal fn", "app (lam 1 x) 2", errorNodef("expecting identifier; got number")},
	{"app illegal arg", "app (lam x x) (lam 1 x)", errorNodef("expecting identifier; got number")},
	{"example", "app app app if (app app gt 3 1) 10 5",
		mkapp(mkapp(
			mkapp(ifNode, mkapp(mkapp(gtNode, mknum(3)), mknum(1))),
			mknum(10)), mknum(5))},
	{"example 2", "app app (app (lam f lam y lam x (app (app f y) x)) (lam x lam y x)) 3 4",
		mkapp(mkapp(
			mkapp(mklam("f", mklam("y", mklam("x",
				mkapp(mkapp(fNode, yNode), xNode)))),
				mklam("x", mklam("y", xNode))),
			mknum(3)), mknum(4))},
}

func nodesEqual(a, b *node) bool {
	if a.typ != b.typ {
		return false
	}
	switch a.typ {
	case nodeNumber:
		av := a.val.(*big.Int)
		bv := b.val.(*big.Int)
		return av.Cmp(bv) == 0
	case nodeApp:
		av := a.val.(*appNode)
		bv := b.val.(*appNode)
		return nodesEqual(av.fn, bv.fn) && nodesEqual(av.arg, bv.arg)
	case nodeLam:
		av := a.val.(*lamNode)
		bv := b.val.(*lamNode)
		return av.param == bv.param && nodesEqual(av.body, bv.body)
	case nodeError:
		return a.val.(error).Error() == b.val.(error).Error()
	default:
		return a.val == b.val
	}
}

func TestParse(t *testing.T) {
	for _, pt := range parseTests {
		root := newParser(pt.input).parse()
		if !nodesEqual(root, pt.root) {
			t.Errorf("[%s]\ninput: %q\nwant: %v\ngot: %v\n", pt.name, pt.input, pt.root, root)
		}
	}
}
