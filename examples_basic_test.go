package shapes

import (
	"fmt"
)

// MakeArrow is a utility function
func ExampleMakeArrow() {
	matmul := MakeArrow(
		Abstract{Var('a'), Var('b')},
		Abstract{Var('b'), Var('c')},
		Abstract{Var('a'), Var('c')},
	)
	fmt.Printf("Correct MatMul: %v\n", matmul)

	wrong := Arrow{
		Arrow{Abstract{Var('a'), Var('b')},
			Abstract{Var('b'), Var('c')},
		},
		Abstract{Var('a'), Var('c')},
	}
	fmt.Printf("Wrong MatMul: %v\n", wrong)

	// it doesn't mean that you should use MakeArrow mindlessly.
	// Consider the higher order function Map: (a → a) → b → b
	// p.s equiv Go function signature:
	// 	func Map(f func(a int) int, b Tensor) Tensor
	Map := MakeArrow(
		Arrow{Var('a'), Var('a')}, // you can also use MakeArrow here
		Var('b'),
		Var('b'),
	)
	fmt.Printf("Correct Map: %v\n", Map)

	wrong = MakeArrow(
		Var('a'), Var('a'),
		Var('b'),
		Var('b'),
	)
	fmt.Printf("Wrong Map: %v\n", wrong)

	// Output:
	// Correct MatMul: (a, b) → (b, c) → (a, c)
	// Wrong MatMul: ((a, b) → (b, c)) → (a, c)
	// Correct Map: (a → a) → b → b
	// Wrong Map: a → a → b → b
}

// Gen is a generator for Abstracts
func ExampleGen() {
	a := Gen(2)
	fmt.Printf("Gen(2): %v\n", a)

	// Gen is not a stateful generator
	b := Gen(2)
	fmt.Printf("Gen(2): %v\n", b)

	// Gen handles a maximum of 50 characters(so far)
	c := Gen(50)
	fmt.Printf("Gen(50): %v\n", c)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Gen will panic if a `d` >= 51 is passed in.")
		}
	}()
	Gen(51)

	// Output:
	// Gen(2): (a, b)
	// Gen(2): (a, b)
	// Gen(50): (a, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p, q, r, s, t, u, v, w, x, y, z, α, β, γ, δ, ε, ζ, η, θ, ι, κ, λ, μ, ν, ξ, ο, π, ρ, ς, σ, τ, υ, φ, χ, ψ)
	// Gen will panic if a `d` >= 51 is passed in.

}

func Example_matMul() {
	matmul := Arrow{
		Abstract{Var('a'), Var('b')},
		Arrow{
			Abstract{Var('b'), Var('c')},
			Abstract{Var('a'), Var('c')},
		},
	}
	fmt.Printf("MatMul: %v\n", matmul)

	// Apply the first input to MatMul
	fst := Shape{2, 3}
	expr2, err := InferApp(matmul, fst)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Applying %v to MatMul:\n", fst)
	fmt.Printf("%v @ %v ↠ %v\n", matmul, fst, expr2)

	// Apply the second input
	snd := Shape{3, 4}
	expr3, err := InferApp(expr2, snd)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Applying %v to the result:\n", snd)
	fmt.Printf("%v @ %v ↠ %v\n", expr2, snd, expr3)

	// Bad example:
	bad2nd := Shape{4, 5}
	_, err = InferApp(expr2, bad2nd)
	fmt.Printf("What happens when you pass in a bad value (e.g. %v instead of %v):\n", bad2nd, snd)
	fmt.Printf("%v @ %v ↠ %v", expr2, bad2nd, err)

	// Output:
	// MatMul: (a, b) → (b, c) → (a, c)
	// Applying (2, 3) to MatMul:
	// (a, b) → (b, c) → (a, c) @ (2, 3) ↠ (3, c) → (2, c)
	// Applying (3, 4) to the result:
	// (3, c) → (2, c) @ (3, 4) ↠ (2, 4)
	// What happens when you pass in a bad value (e.g. (4, 5) instead of (3, 4)):
	// (3, c) → (2, c) @ (4, 5) ↠ Failed to solve [{(3, c) → (2, c) = (4, 5) → d}] | d: Unification Fail. 3 ~ 4 cannot proceed

}

// This examples shows how to describe a particular operation (addition), and how to use
// the infer functions to provide inference for the following types.
func Example_add() {
	// Consider the idea of adding two tensors - A and B - together.
	// If it's a matrix, then both A and B must have a same shape.
	// Thus we can use the following shape expression to describe addition:
	//	Add: a → a → a
	//
	// if A has shape `a`, then B also has to have shape `a`. The result is also shaped `a`.

	add := Arrow{
		Var('a'),
		Arrow{
			Var('a'),
			Var('a'),
		},
	}
	fmt.Printf("Add: %v\n", add)

	// pass in the first input
	fst := Shape{5, 2, 3, 1, 10}
	retExpr, err := InferApp(add, fst)
	if err != nil {
		fmt.Printf("Error %v\n", err)
	}
	fmt.Printf("Applying %v to Add:\n", fst)
	fmt.Printf("%v @ %v ↠ %v\n", add, fst, retExpr)

	// pass in the second input
	snd := Shape{5, 2, 3, 1, 10}
	retExpr2, err := InferApp(retExpr, snd)
	if err != nil {
		fmt.Printf("Error %v\n", err)
	}
	fmt.Printf("Applying %v to the result\n", snd)
	fmt.Printf("%v @ %v ↠ %v\n", retExpr, snd, retExpr2)

	// bad example:
	bad2nd := Shape{2, 3}
	_, err = InferApp(retExpr, bad2nd)

	fmt.Printf("Passing in a bad second input\n")
	fmt.Printf("%v @ %v ↠ %v", retExpr, bad2nd, err)

	// Output:
	// Add: a → a → a
	// Applying (5, 2, 3, 1, 10) to Add:
	// a → a → a @ (5, 2, 3, 1, 10) ↠ (5, 2, 3, 1, 10) → (5, 2, 3, 1, 10)
	// Applying (5, 2, 3, 1, 10) to the result
	// (5, 2, 3, 1, 10) → (5, 2, 3, 1, 10) @ (5, 2, 3, 1, 10) ↠ (5, 2, 3, 1, 10)
	// Passing in a bad second input
	// (5, 2, 3, 1, 10) → (5, 2, 3, 1, 10) @ (2, 3) ↠ Failed to solve [{(5, 2, 3, 1, 10) → (5, 2, 3, 1, 10) = (2, 3) → a}] | a: Unification Fail. (5, 2, 3, 1, 10) ~ (2, 3) cannot proceed as they do not contain the same amount of sub-expressions. (5, 2, 3, 1, 10) has 5 subexpressions while (2, 3) has 2 subexpressions

}

func Example_ravel() {
	ravel := Arrow{
		Var('a'),
		Abstract{UnaryOp{Prod, Var('a')}},
	}
	fmt.Printf("Ravel: %v\n", ravel)

	fst := Shape{2, 3, 4}
	retExpr, err := InferApp(ravel, fst)
	if err != nil {
		fmt.Printf("Error %v\n", err)
	}
	fmt.Printf("Applying %v to Ravel:\n", fst)
	fmt.Printf("%v @ %v ↠ %v", ravel, fst, retExpr)

	// Output:
	// Ravel: a → (Π a)
	// Applying (2, 3, 4) to Ravel:
	// a → (Π a) @ (2, 3, 4) ↠ (24)
}

func Example_transpose() {
	axes := Axes{0, 1, 3, 2}
	simple := Arrow{
		Var('a'),
		Arrow{
			axes,
			TransposeOf{
				axes,
				Var('a'),
			},
		},
	}
	fmt.Printf("Unconstrained Transpose: %v\n", simple)

	st := SubjectTo{
		Eq,
		UnaryOp{Dims, axes},
		UnaryOp{Dims, Var('a')},
	}
	transpose := Compound{
		Expr:      simple,
		SubjectTo: st,
	}
	fmt.Printf("Transpose: %v\n", transpose)

	fst := Shape{1, 2, 3, 4}
	retExpr, err := InferApp(transpose, fst)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Applying %v to %v:\n", fst, transpose)
	fmt.Printf("\t%v @ %v ↠ %v\n", transpose, fst, retExpr)
	snd := axes
	retExpr2, err := InferApp(retExpr, snd)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Applying %v to %v:\n", snd, retExpr)
	fmt.Printf("\t%v @ %v ↠ %v\n", retExpr, snd, retExpr2)

	// bad axes
	bad2nd := Axes{0, 2, 1, 3} // not the original axes {0,1,3,2}
	_, err = InferApp(retExpr, bad2nd)
	fmt.Printf("Bad Axes causes error: %v\n", err)

	// bad first input
	bad1st := Shape{2, 3, 4}
	_, err = InferApp(transpose, bad1st)
	fmt.Printf("Bad first input causes error: %v", err)

	// Output:
	// Unconstrained Transpose: a → X[0 1 3 2] → T X[0 1 3 2] a
	// Transpose: { a → X[0 1 3 2] → T X[0 1 3 2] a | (D X[0 1 3 2] = D a) }
	// Applying (1, 2, 3, 4) to { a → X[0 1 3 2] → T X[0 1 3 2] a | (D X[0 1 3 2] = D a) }:
	// 	{ a → X[0 1 3 2] → T X[0 1 3 2] a | (D X[0 1 3 2] = D a) } @ (1, 2, 3, 4) ↠ X[0 1 3 2] → T X[0 1 3 2] (1, 2, 3, 4)
	// Applying X[0 1 3 2] to X[0 1 3 2] → T X[0 1 3 2] (1, 2, 3, 4):
	// 	X[0 1 3 2] → T X[0 1 3 2] (1, 2, 3, 4) @ X[0 1 3 2] ↠ (1, 2, 4, 3)
	// Bad Axes causes error: Failed to solve [{X[0 1 3 2] → T X[0 1 3 2] (1, 2, 3, 4) = X[0 2 1 3] → a}] | a: Unification Fail. X[0 1 3 2] ~ X[0 2 1 3] cannot proceed
	// Bad first input causes error: SubjectTo (D X[0 1 3 2] = D (2, 3, 4)) resolved to false. Cannot continue
	//

}

func Example_index() {
	sizes := Sizes{0, 0, 1, 0}
	simple := Arrow{
		Var('a'),
		Arrow{
			Var('b'),
			Abstract{},
		},
	}
	fmt.Printf("Unconstrained Indexing: %v\n", simple)

	st := SubjectTo{
		And,
		SubjectTo{
			Eq,
			UnaryOp{Dims, Var('a')},
			UnaryOp{Dims, Var('b')},
		},
		SubjectTo{
			Lt,
			UnaryOp{ForAll, Var('b')},
			UnaryOp{ForAll, Var('a')},
		},
	}
	index := Compound{Expr: simple, SubjectTo: st}
	fmt.Printf("Indexing: %v\n", index)

	fst := Shape{1, 2, 3, 4}
	retExpr, err := InferApp(index, fst)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Applying %v to %v:\n", fst, index)
	fmt.Printf("\t%v @ %v ↠ %v\n", index, fst, retExpr)

	snd := sizes
	retExpr2, err := InferApp(retExpr, snd)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Applying %v to %v:\n", snd, retExpr)
	fmt.Printf("\t%v @ %v ↠ %v\n", retExpr, snd, retExpr2)

	// Output:
	// Unconstrained Indexing: a → b → ()
	// Indexing: { a → b → () | ((D a = D b) ∧ (∀ b < ∀ a)) }
	// Applying (1, 2, 3, 4) to { a → b → () | ((D a = D b) ∧ (∀ b < ∀ a)) }:
	// 	{ a → b → () | ((D a = D b) ∧ (∀ b < ∀ a)) } @ (1, 2, 3, 4) ↠ { b → () | ((D (1, 2, 3, 4) = D b) ∧ (∀ b < ∀ (1, 2, 3, 4))) }
	// Applying Sz[0 0 1 0] to { b → () | ((D (1, 2, 3, 4) = D b) ∧ (∀ b < ∀ (1, 2, 3, 4))) }:
	// 	{ b → () | ((D (1, 2, 3, 4) = D b) ∧ (∀ b < ∀ (1, 2, 3, 4))) } @ Sz[0 0 1 0] ↠ ()
}

func Example_index_unidimensional() {
	sizes := Sizes{0}
	simple := Arrow{
		Var('a'),
		Arrow{
			Var('b'),
			Abstract{},
		},
	}
	fmt.Printf("Unconstrained Indexing: %v\n", simple)

	st := SubjectTo{
		And,
		SubjectTo{
			Eq,
			UnaryOp{Dims, Var('a')},
			UnaryOp{Dims, Var('b')},
		},
		SubjectTo{
			Lt,
			UnaryOp{ForAll, Var('b')},
			UnaryOp{ForAll, Var('a')},
		},
	}
	index := Compound{Expr: simple, SubjectTo: st}
	fmt.Printf("Indexing: %v\n", index)

	fst := Shape{5}
	retExpr, err := InferApp(index, fst)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Applying %v to %v:\n", fst, index)
	fmt.Printf("\t%v @ %v ↠ %v\n", index, fst, retExpr)

	snd := sizes
	retExpr2, err := InferApp(retExpr, snd)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Applying %v to %v:\n", snd, retExpr)
	fmt.Printf("\t%v @ %v ↠ %v\n", retExpr, snd, retExpr2)

	// Output:
	// Unconstrained Indexing: a → b → ()
	// Indexing: { a → b → () | ((D a = D b) ∧ (∀ b < ∀ a)) }
	// Applying (5) to { a → b → () | ((D a = D b) ∧ (∀ b < ∀ a)) }:
	// 	{ a → b → () | ((D a = D b) ∧ (∀ b < ∀ a)) } @ (5) ↠ { b → () | ((D (5) = D b) ∧ (∀ b < ∀ (5))) }
	// Applying Sz[0] to { b → () | ((D (5) = D b) ∧ (∀ b < ∀ (5))) }:
	// 	{ b → () | ((D (5) = D b) ∧ (∀ b < ∀ (5))) } @ Sz[0] ↠ ()

}

func Example_index_scalar() {
	sizes := Sizes{}
	simple := Arrow{
		Var('a'),
		Arrow{
			Var('b'),
			Abstract{},
		},
	}
	fmt.Printf("Unconstrained Indexing: %v\n", simple)

	st := SubjectTo{
		And,
		SubjectTo{
			Eq,
			UnaryOp{Dims, Var('a')},
			UnaryOp{Dims, Var('b')},
		},
		SubjectTo{
			Lt,
			UnaryOp{ForAll, Var('b')},
			UnaryOp{ForAll, Var('a')},
		},
	}
	index := Compound{Expr: simple, SubjectTo: st}
	fmt.Printf("Indexing: %v\n", index)

	fst := Shape{}
	retExpr, err := InferApp(index, fst)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Applying %v to %v:\n", fst, index)
	fmt.Printf("\t%v @ %v ↠ %v\n", index, fst, retExpr)

	snd := sizes
	retExpr2, err := InferApp(retExpr, snd)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Applying %v to %v:\n", snd, retExpr)
	fmt.Printf("\t%v @ %v ↠ %v\n", retExpr, snd, retExpr2)

	// Output:
	// Unconstrained Indexing: a → b → ()
	// Indexing: { a → b → () | ((D a = D b) ∧ (∀ b < ∀ a)) }
	// Applying () to { a → b → () | ((D a = D b) ∧ (∀ b < ∀ a)) }:
	// 	{ a → b → () | ((D a = D b) ∧ (∀ b < ∀ a)) } @ () ↠ { b → () | ((D () = D b) ∧ (∀ b < ∀ ())) }
	// Applying Sz[] to { b → () | ((D () = D b) ∧ (∀ b < ∀ ())) }:
	// 	{ b → () | ((D () = D b) ∧ (∀ b < ∀ ())) } @ Sz[] ↠ ()

}

func Example_slice() {
	sli := Range{0, 2, 1}
	simple := Arrow{
		Var('a'),
		Arrow{
			sli,
			SliceOf{
				sli,
				Var('a'),
			},
		},
	}
	slice := Compound{
		Expr: simple,
		SubjectTo: SubjectTo{
			OpType: Gte,
			A:      IndexOf{I: 0, A: Var('a')},
			B:      Size(2),
		},
	}

	fmt.Printf("slice: %v\n", slice)

	fst := Shape{2, 3, 4}
	retExpr, err := InferApp(slice, fst)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Applying %v to %v:\n", fst, slice)
	fmt.Printf("\t%v @ %v ↠ %v\n", slice, fst, retExpr)

	snd := sli
	retExpr2, err := InferApp(retExpr, snd)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Applying %v to %v:\n", snd, retExpr)
	fmt.Printf("\t%v @ %v ↠ %v\n", retExpr, snd, retExpr2)

	// Output:
	// slice: { a → [0:2] → a[0:2] | (a[0] ≥ 2) }
	// Applying (2, 3, 4) to { a → [0:2] → a[0:2] | (a[0] ≥ 2) }:
	// 	{ a → [0:2] → a[0:2] | (a[0] ≥ 2) } @ (2, 3, 4) ↠ [0:2] → (2, 3, 4)[0:2]
	// Applying [0:2] to [0:2] → (2, 3, 4)[0:2]:
	// 	[0:2] → (2, 3, 4)[0:2] @ [0:2] ↠ (2, 3, 4)[0:2]

}

func Example_reshape() {
	expr := Compound{
		Arrow{
			Var('a'),
			Arrow{
				Var('b'),
				Var('b'),
			},
		},
		SubjectTo{
			Eq,
			UnaryOp{Prod, Var('a')},
			UnaryOp{Prod, Var('b')},
		},
	}

	fmt.Printf("Reshape: %v\n", expr)

	fst := Shape{2, 3}
	snd := Shape{3, 2}

	retExpr, err := InferApp(expr, fst)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Applying %v to %v:\n", fst, expr)
	fmt.Printf("\t%v @ %v ↠ %v\n", expr, fst, retExpr)

	retExpr2, err := InferApp(retExpr, snd)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Applying %v to %v:\n", snd, retExpr)
	fmt.Printf("\t%v @ %v ↠ %v\n", retExpr, snd, retExpr2)

	bad := Shape{6, 2}
	_, err = InferApp(retExpr, bad)
	fmt.Printf("Applying a bad shape %v to %v:\n", bad, retExpr)
	fmt.Printf("\t%v\n", err)

	// Output:
	// Reshape: { a → b → b | (Π a = Π b) }
	// Applying (2, 3) to { a → b → b | (Π a = Π b) }:
	//	{ a → b → b | (Π a = Π b) } @ (2, 3) ↠ { b → b | (Π (2, 3) = Π b) }
	// Applying (3, 2) to { b → b | (Π (2, 3) = Π b) }:
	//	{ b → b | (Π (2, 3) = Π b) } @ (3, 2) ↠ (3, 2)
	// Applying a bad shape (6, 2) to { b → b | (Π (2, 3) = Π b) }:
	//	SubjectTo (Π (2, 3) = Π (6, 2)) resolved to false. Cannot continue

}

// The following shape expressions describe a columnwise summing of a matrix.
func Example_colwiseSumMatrix() {
	// Given A:
	// 	1 2 3
	//	4 5 6
	//
	// The columnwise sum is:
	// 	5 7 9
	//
	// The basic description can be explained as such:
	//	(r, c) → (1, c)
	//
	// Here a matrix is given as (r, c). After a columnwise sum, the result is 1 row of c columns.
	// However, to keep compatibility with Numpy, a colwise sum would look like this:
	// 	(r, c) → (c, )
	//
	// Lastly a generalized single-axis sum that would work across all tensors would be:
	// 	a → b | (D b = D a - 1)
	//
	// Here, it says that the sum is a function that takes a tensor of any shape called `a`, and returns a tensor with a different shape, called `b`.
	// The constraints however is that the dimensions of `b` must be the dimensions of `a` minus 1.

	basic := MakeArrow(
		Abstract{Var('r'), Var('c')},
		Abstract{Size(1), Var('c')},
	)
	fmt.Printf("Basic: %v\n", basic)

	compat := MakeArrow(
		Abstract{Var('r'), Var('c')},
		Abstract{Var('c')},
	)
	fmt.Printf("Numpy Compatible: %v\n", compat)

	general := Compound{
		Arrow{Var('a'), Var('b')},
		SubjectTo{
			Eq,
			UnaryOp{Dims, Var('b')},
			BinOp{
				Sub,
				UnaryOp{Dims, Var('a')},
				Size(1),
			},
		},
	}
	fmt.Printf("General: %v\n", general)

	// Output:
	// Basic: (r, c) → (1, c)
	// Numpy Compatible: (r, c) → (c)
	// General: { a → b | (D b = D a - 1) }
}

func Example_trace() {
	expr := Arrow{
		Abstract{Var('a'), Var('a')},
		Shape{},
	}
	fmt.Printf("Trace: %v\n", expr)

	// Output:
	// Trace: (a, a) → ()
}
