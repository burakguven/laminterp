package main

import (
	"bufio"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"
)

func mknumobj(n int64) *object {
	return &object{objectNumber, big.NewInt(n)}
}

type evalTest struct {
	name  string
	input string
	val   *object
}

var (
	trueObj  = &object{objectBool, true}
	falseObj = &object{objectBool, false}
)

var evalTests = []evalTest{
	{"number", "3", mknumobj(3)},
	{"negative number", "-9", mknumobj(-9)},
	{"bool true", "true", trueObj},
	{"bool false", "false", falseObj},
	{"add", "app app add 1 3", mknumobj(4)},
	{"add negative", "app app add -7 3", mknumobj(-4)},
	{"add non-number first argument", "app app add false 1",
		errorObjectf("add: not a number: 'false'")},
	{"add non-number second argument", "app app add 1 true",
		errorObjectf("add: not a number: 'true'")},
	{"if true", "app app app if true 1 2", mknumobj(1)},
	{"if false", "app app app if false 1 2", mknumobj(2)},
	{"if with non-bool", "app app app if 1 2 3",
		errorObjectf("if: not a bool: '1'")},
	{"gt greater", "app app gt 2 1", trueObj},
	{"gt less", "app app gt 1 2", falseObj},
	{"gt equal", "app app gt 1 1", falseObj},
	{"gt negative", "app app gt -1 1", falseObj},
	{"gt with non-number first argument", "app app gt false 1",
		errorObjectf("gt: not a number: 'false'")},
	{"gt with non-number second argument", "app app gt 1 true",
		errorObjectf("gt: not a number: 'true'")},
	{"unknown identifier", "x", errorObjectf("unknown identifier: 'x'")},
	{"lam", "app lam x x 1", mknumobj(1)},
	{"app invalid function", "app true 1",
		errorObjectf("apply: invalid function: 'true'")},
	{"parent env reference",
		"app app lam x lam y x 1 2", mknumobj(1)},
	{"passing lam as argument",
		"app app (app (lam x lam y x) lam x x) lam y y 1", mknumobj(1)},
	{"app inside lam",
		"app lam x (app app add 1 x) 2", mknumobj(3)},
	{"example",
		"app app (app (lam f lam y lam x (app (app f y) x)) (lam x lam y x)) 3 4",
		mknumobj(3)},
}

func equalObject(a, b *object) bool {
	if a.typ != b.typ {
		return false
	}
	if a.typ == objectNumber {
		an := a.val.(*big.Int)
		bn := b.val.(*big.Int)
		return an.Cmp(bn) == 0
	}
	return a.val == b.val
}

func TestEval(t *testing.T) {
	for _, et := range evalTests {
		val := evalString(et.input)
		if !equalObject(val, et.val) {
			t.Errorf("[%s]: %s\nwant: %q\ngot: %q", et.name, et.input, et.val, val)
		}
	}
}

func readLines(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func zipStringSlices(slices ...[]string) [][]string {
	if len(slices) == 0 {
		return nil
	}
	var tuples [][]string
Outer:
	for i, el := range slices[0] {
		tuple := []string{el}
		for _, s := range slices[1:] {
			if i >= len(s) {
				break Outer
			}
			tuple = append(tuple, s[i])
		}
		tuples = append(tuples, tuple)
	}
	return tuples
}

func readZippedLines(filenames ...string) ([][]string, error) {
	if len(filenames) == 0 {
		return nil, fmt.Errorf("readZippedLines: no filenames")
	}
	var fileLines [][]string
	for _, f := range filenames {
		lines, err := readLines(f)
		if err != nil {
			return nil, err
		}
		fileLines = append(fileLines, lines)
		if len(fileLines[0]) != len(lines) {
			return nil, fmt.Errorf("readZippedLines: uneven files")
		}
	}
	return zipStringSlices(fileLines...), nil
}

func TestEvalFiles(t *testing.T) {
	inFiles, err := filepath.Glob("test_files/*.in")
	if err != nil {
		t.Fatal(err)
	}
	for _, inFile := range inFiles {
		outFile := inFile[:len(inFile)-len(".in")] + ".out"
		tuples, err := readZippedLines(inFile, outFile)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("testing %s", inFile)
		for _, tuple := range tuples {
			in, out := tuple[0], tuple[1]
			val := fmt.Sprint(evalString(in))
			if val != out {
				t.Errorf("%s\nwant: %q\ngot: %q", in, out, val)
			}
		}
	}
}
