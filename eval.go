package main

import (
	"fmt"
	"math/big"
)

type applyer interface {
	apply(*object) *object
}

type objectType int

// Constants indicating the type of the value stored in an object struct.
const (
	objectError  objectType = iota // object.val is set to an error string
	objectBool                     // object.val is set to a bool
	objectNumber                   // object.val is set to a *big.Int
	objectFunc                     // object.val is set to a funcObject
	objectLam                      // object.val is set to a *lamObject
)

// An object represents a generic object within the interpreter context.
type object struct {
	typ objectType
	val interface{}
}

func (v *object) String() string {
	switch v.typ {
	case objectError:
		return v.val.(string)
	case objectBool:
		if v.val.(bool) {
			return "true"
		}
		return "false"
	case objectNumber:
		return v.val.(*big.Int).String()
	case objectFunc:
		return fmt.Sprintf("<function %p>", v.val)
	case objectLam:
		lam := v.val.(*lamObject)
		return fmt.Sprintf("<lam %s %p>", lam.node.param, lam)
	default:
		// Shouldn't be possible
		panic(fmt.Errorf("invalid object type: %d", v.typ))
	}
}

// errorObjectf formats according to a format specifier (see fmt) and returns
// the resulting string as an error object.
func errorObjectf(format string, args ...interface{}) *object {
	return &object{objectError, fmt.Sprintf(format, args...)}
}

// A funcObject represents a built-in function within the interpreter context.
type funcObject func(*object) *object

var _ applyer = funcObject(nil)

// newFuncObject returns the given function wrapped into a function object.
func newFuncObject(fn func(*object) *object) *object {
	return &object{objectFunc, funcObject(fn)}
}

// apply calls f with v as an argument and returns the result.
func (f funcObject) apply(v *object) *object {
	return f(v)
}

// The builtin function add returns the sum of two numbers.
// Signature: number -> number -> number
var builtinAdd = newFuncObject(func(a *object) *object {
	if a.typ != objectNumber {
		return errorObjectf("add: not a number: '%s'", a)
	}
	return newFuncObject(func(b *object) *object {
		if b.typ != objectNumber {
			return errorObjectf("add: not a number: '%s'", b)
		}
		an := a.val.(*big.Int)
		bn := b.val.(*big.Int)
		return &object{objectNumber, new(big.Int).Add(an, bn)}
	})
})

// The builtin function if branches on a bool (the first argument).
// If the bool is true, the second argument is returned, otherwise the third.
// Signature: bool -> object -> object -> object
var builtinIf = newFuncObject(func(a *object) *object {
	if a.typ != objectBool {
		return errorObjectf("if: not a bool: '%s'", a)
	}
	return newFuncObject(func(b *object) *object {
		return newFuncObject(func(c *object) *object {
			if a.val.(bool) {
				return b
			}
			return c
		})
	})
})

// The builtin function gt compares two numbers and returns the result as a
// boolean which is true only if the first argument is greater than the second.
// Signature: number -> number -> bool
var builtinGt = newFuncObject(func(a *object) *object {
	if a.typ != objectNumber {
		return errorObjectf("gt: not a number: '%s'", a)
	}
	return newFuncObject(func(b *object) *object {
		if b.typ != objectNumber {
			return errorObjectf("gt: not a number: '%s'", b)
		}
		an := a.val.(*big.Int)
		bn := b.val.(*big.Int)
		return &object{objectBool, an.Cmp(bn) == 1}
	})
})

// An environment contains a list of symbols. It is used to resolve identifiers
// when evaluating a parse tree.
//
// Duplicates are allowed, with symbols added later taking precedence over
// earlier ones.
//
// This is an immutable data structure, so all operations return a new
// environment instead of mutating the current one.
type environment struct {
	parent *environment
	symbol string
	val    *object
}

// newEnvironment creates and returns an environment which contains the given
// symbol.
//
// The new environment can start out with one symbol by passing nil as the
// parent, or it can inherit symbols from another environment by passing the
// other environment as the parent parameter.
func newEnvironment(parent *environment, symbol string, val *object) *environment {
	return &environment{
		parent: parent,
		symbol: symbol,
		val:    val,
	}
}

// extend is a convenience function which returns the same thing as
// newEnvironment(e, symbol, val).
func (e *environment) extend(symbol string, val *object) *environment {
	return newEnvironment(e, symbol, val)
}

// lookup returns the value associated with a symbol. See the environment type
// definition for details on how duplicates are handled.
func (e *environment) lookup(symbol string) *object {
	for cur := e; cur != nil; cur = cur.parent {
		if symbol == cur.symbol {
			return cur.val
		}
	}
	return errorObjectf("unknown identifier: '%s'", symbol)
}

// A lamObject represents a lambda function within the interpreter context.
type lamObject struct {
	node *lamNode
	env  *environment
}

var _ applyer = &lamObject{}

func (v *lamObject) apply(arg *object) *object {
	return evalEnv(v.node.body, v.env.extend(v.node.param, arg))
}

// evalEnv evaluates a node within the context of a particular environment.
func evalEnv(n *node, env *environment) *object {
	switch n.typ {
	case nodeApp:
		app := n.val.(*appNode)
		fn := evalEnv(app.fn, env)
		if fn.typ == objectError {
			return fn
		}
		arg := evalEnv(app.arg, env)
		if arg.typ == objectError {
			return arg
		}
		if fnApplyer, ok := fn.val.(applyer); ok {
			return fnApplyer.apply(arg)
		}
		return errorObjectf("apply: invalid function: '%s'", fn)
	case nodeLam:
		return &object{objectLam, &lamObject{n.val.(*lamNode), env}}
	case nodeNumber:
		return &object{objectNumber, n.val}
	case nodeBool:
		return &object{objectBool, n.val}
	case nodeIdentifier:
		return env.lookup(n.val.(string))
	case nodeError:
		return errorObjectf("parse error: %s", n.val.(string))
	default:
		// Shouldn't be possible
		panic(fmt.Errorf("invalid node: %s", n.typ))
	}
}

// defaultEnvironment is an environment that contains the built-in functions.
// It's used as the default environment in some places, as noted.
var defaultEnvironment = newEnvironment(nil, "add", builtinAdd).
	extend("if", builtinIf).
	extend("gt", builtinGt)

// eval evaluates a node with the default environment.
func eval(n *node) *object {
	return evalEnv(n, defaultEnvironment)
}

// evalString parses and evaluates a string with the default environment.
func evalString(s string) *object {
	return eval(parseString(s))
}
