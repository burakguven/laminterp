laminterp
===========

Interpreter for a programming language based on [lambda calculus](https://en.wikipedia.org/wiki/Lambda_calculus).

## A Short Tour

This language is very simple. There are only a few main categories of syntax:

* Literals like `5` or `true`.
* Function application, which is written as `app <function> <argument>`.
* Lambda functions, which are written as `lam <binding> <body>`.

### Literals

There are only two types of literals: **booleans**, and **numbers** (arbitrary-precision integers). Booleans are represented as `true` and `false`. Numbers are represented in the usual way.

### Function Application

`app` is the notation used to call a function. The first argument is the function to call, and the second argument is the argument to the function. In mathematical notation, `app f x` would be represented as `f(x)`. So for example, the following expression computes the sum of two numbers:

```
app app add 1 2
```

Note that functions can take only one argument. Functions that effectively have multiple arguments can be created with [currying](https://en.wikipedia.org/wiki/Currying). So in the above example, `add` doesn't actually take two arguments. Instead, `app add 1` returns another function that adds `1` to its argument. Said another way, the example might be rewritten more explicitly as:

```
app (app add 1) 2
```

### Lambda Functions

The other main construct in this language is the lambda function. It looks like this:

```
lam x x
```

The first argument to lam is the function's parameter (also called the "binding"), and the second argument is the body of the function. The function given above is the identity function which returns whatever is passed to it. So for example, the value of the following expression is `7`:

```
app lam x x 7
```

### Built-in Functions

There are only a few built-in functions:

* `add`: adds two integers.
* `if`: branches on a bool. If the first argument (a `boolean`) is true, it returns the second argument, otherwise the third.
* `gt`: returns true if the first argument is greater than the second.

For example, the value of the following program is `4`.
```
app app app if (app app gt 1 2) 3 4
```

## An Example Program

The following program computes Fibonacci numbers:

```
   1   app
   2       app
   3           lam fix app
   4               lam rec
   5                   lam n app app app rec n 0 1
   6               app
   7                   fix
   8                   lam f lam n lam a lam b app
   9                       app app app
  10                           if
  11                           app app gt n 0
  12                           lam x app app app
  13                               f
  14                               app app add n -1
  15                               b
  16                               app app add a b
  17                           lam x a
  18                       false
  19           lam f app
  20               lam x app
  21                   f
  22                   lam y app
  23                       app x x
  24                       y
  25               lam x app
  26                   f
  27                   lam y app
  28                       app x x
  29                       y
  30       500
```

As a refresher, the Fibonacci series is defined recursively by the following set of equations:

```
F(0) = 0
F(1) = 1
F(n) = F(n-2) + F(n-1)
```

The program is a little more complex than it would be in other languages mainly due to two factors: 1) there are no looping constructs, and 2) there's no way to create definitions which makes it more challenging to create recursive functions. We can still use recursion however and get around the second limitation by using the [fixed-point combinator](https://en.wikipedia.org/wiki/Fixed-point_combinator) (also called the Y combinator).

You can see the fixed-point combinator (technically the [strict version](https://en.wikipedia.org/wiki/Fixed-point_combinator#Strict_fixed_point_combinator)) on lines 19-29, and the semi-recursive fibonacci function on lines 8-18. We apply that semi-recursive function to the fixed-point combinator in order to get an actually recursive function. The rest is just passing in the initial values (seen on lines 4-5). Finally, we call the resulting function with the value `500` on line 30.

There are more example programs in the `examples` folder.

## Installation

```
go get github.com/burakguven/laminterp
```

## Usage

After installation, there should be a binary named `laminterp` in the `$GOPATH/bin` directory. When run in a terminal without any arguments, this is an interactive shell that you might be familiar with from other interpreted programming languages. Here's an example session:

```
$ laminterp
>> app app add 1 2
3
>> app app gt 1 2
false
>> app app app if (app app gt 1 2) 3 4
4
```

When run with a file name, it runs the given program:

```
$ laminterp examples/fibonacci
139423224561697880139724382870407283950070256587697307264108962948325571622863290691557658876222521294125
```

## Acknowledgements

* The idea for this project came from one of the problems in [Vivint's 2017 coding competition](https://innovation.vivint.com/and-thats-a-wrap-vivint-s-game-of-codes-2017-deab198696aa). You can see the original specification for this particular problem [here](https://github.com/vivint/coding-competitions/blob/master/goc2017/laminterp/README.md). The test programs in the `test_files` directory also came from there. Many thanks to Vivint.
* The implementation of the lexer was inspired by the `text/template` package in the Go standard library.
